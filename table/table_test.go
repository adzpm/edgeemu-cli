package table

import (
	"os"
	"strings"
	"testing"

	"github.com/adzpm/edgeemu-cli/ds"
)

func TestFitWidthsToNoConstraintWhenFits(t *testing.T) {
	headers := []string{"A", "B"}
	rows := [][]string{{"aa", "bb"}}

	if got := fitWidthsTo(headers, rows, 200); got != nil {
		t.Errorf("fitWidthsTo = %v, want nil when the table already fits", got)
	}
}

func TestFitWidthsToShrinksWidestFirst(t *testing.T) {
	headers := []string{"#", "Name", "URL"}
	rows := [][]string{
		{"1", "Short", strings.Repeat("u", 80)},
	}

	widths := fitWidthsTo(headers, rows, 40)
	if widths == nil {
		t.Fatal("fitWidthsTo = nil, want constrained widths")
	}

	// Narrow columns keep their natural width; only the widest shrinks.
	if widths[0] != 3 { // "1" + padding
		t.Errorf("col 0 width = %d, want 3", widths[0])
	}
	if widths[1] != 7 { // "Short" + padding
		t.Errorf("col 1 width = %d, want 7", widths[1])
	}

	total := len(headers) + 1
	for i := range headers {
		total += widths[i]
	}
	if total != 40 {
		t.Errorf("total width = %d, want exactly 40", total)
	}
}

func TestFitWidthsToRespectsMinimum(t *testing.T) {
	headers := []string{"A", "B", "C"}
	rows := [][]string{{strings.Repeat("a", 50), strings.Repeat("b", 50), strings.Repeat("c", 50)}}

	widths := fitWidthsTo(headers, rows, 10) // impossible to fit
	if widths == nil {
		t.Fatal("fitWidthsTo = nil, want constrained widths")
	}

	for i, w := range widths {
		if w < minColWidth {
			t.Errorf("col %d width = %d, below minimum %d", i, w, minColWidth)
		}
	}
}

func TestFitWidthsToRaggedRowsDoNotPanic(t *testing.T) {
	headers := []string{"A"}
	rows := [][]string{{"a", "extra", "columns"}}

	// Must not panic on rows longer than headers.
	fitWidthsTo(headers, rows, 10)
}

func TestColumnIDsOrder(t *testing.T) {
	want := []string{"name", "system", "size", "unpacked", "dls", "hash", "url"}
	got := ColumnIDs()

	if len(got) != len(want) {
		t.Fatalf("ColumnIDs() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ColumnIDs() = %v, want %v", got, want)
		}
	}
}

func TestSelectColumns(t *testing.T) {
	// Requested out of order, with noise in case and spacing:
	// the result must follow canonical display order.
	cols, err := selectColumns([]string{" URL ", "Name", "size"})
	if err != nil {
		t.Fatalf("selectColumns: %v", err)
	}

	got := make([]string, len(cols))
	for i, c := range cols {
		got[i] = c.id
	}

	want := []string{"name", "size", "url"}
	if len(got) != len(want) {
		t.Fatalf("selected %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("selected %v, want %v", got, want)
		}
	}
}

func TestSelectColumnsUnknown(t *testing.T) {
	_, err := selectColumns([]string{"name", "bogus"})
	if err == nil {
		t.Fatal("want error for unknown column, got nil")
	}
	if !strings.Contains(err.Error(), "bogus") {
		t.Errorf("error %q does not name the bad column", err)
	}
}

// captureStdout runs fn while redirecting os.Stdout into a pipe and
// returns everything written. Render targets os.Stdout directly, and a
// pipe also disables terminal-width squeezing, keeping output stable.
func captureStdout(t *testing.T, fn func() error) string {
	t.Helper()

	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w

	done := make(chan string)
	go func() {
		buf := make([]byte, 0, 4096)
		tmp := make([]byte, 4096)
		for {
			n, err := r.Read(tmp)
			buf = append(buf, tmp[:n]...)
			if err != nil {
				break
			}
		}
		done <- string(buf)
	}()

	fnErr := fn()

	w.Close()
	os.Stdout = orig
	out := <-done

	if fnErr != nil {
		t.Fatalf("render: %v", fnErr)
	}
	return out
}

func TestPrintROMs(t *testing.T) {
	roms := []ds.ROM{
		{Name: "Sonic & Knuckles (World)", System: "Sega Mega Drive / Genesis", URL: "https://example.com/1.zip", Size: "1.36m", Downloads: 341, Hash: "4DCFD55C"},
		{Name: "Sonic The Hedgehog 2 (World)", System: "Sega Mega Drive / Genesis", URL: "https://example.com/2.zip", Size: "732.08k", Downloads: 432, Hash: "24AB4C3A"},
	}

	out := captureStdout(t, func() error {
		return PrintROMs(roms, []string{"name", "dls"})
	})

	for _, want := range []string{"Name", "DLs", "Sonic & Knuckles (World)", "341", "432"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
	for _, banned := range []string{"https://example.com/1.zip", "Sega Mega Drive", "4DCFD55C"} {
		if strings.Contains(out, banned) {
			t.Errorf("output contains %q from an unselected column:\n%s", banned, out)
		}
	}
}

func TestPrintROMsUnknownColumn(t *testing.T) {
	err := PrintROMs([]ds.ROM{{Name: "x"}}, []string{"nope"})
	if err == nil {
		t.Fatal("want error for unknown column, got nil")
	}
}

func TestPrintSystems(t *testing.T) {
	systems := []ds.System{
		{ID: "atari-2600", Name: "Atari 2600"},
		{ID: "sega-genesis", Name: "Sega Mega Drive / Genesis"},
	}

	out := captureStdout(t, func() error {
		return PrintSystems(systems)
	})

	for _, want := range []string{"ID", "atari-2600", "Sega Mega Drive / Genesis"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}
