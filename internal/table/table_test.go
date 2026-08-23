package table

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adzpm/edgeemu-cli/internal/ds"
)

func TestFitWidthsToNoConstraintWhenFits(t *testing.T) {
	headers := []string{"A", "B"}
	rows := [][]string{{"aa", "bb"}}

	assert.Nil(t, fitWidthsTo(headers, rows, 200), "no constraints needed when the table already fits")
}

func TestFitWidthsToShrinksWidestFirst(t *testing.T) {
	headers := []string{"#", "Name", "URL"}
	rows := [][]string{
		{"1", "Short", strings.Repeat("u", 80)},
	}

	widths := fitWidthsTo(headers, rows, 40)
	require.NotNil(t, widths)

	// Narrow columns keep their natural width; only the widest shrinks.
	assert.Equal(t, 3, widths[0], `"1" plus cell padding`)
	assert.Equal(t, 7, widths[1], `"Short" plus cell padding`)

	total := len(headers) + 1 // vertical borders
	for i := range headers {
		total += widths[i]
	}
	assert.Equal(t, 40, total, "squeezed table must use exactly the terminal width")
}

func TestFitWidthsToRespectsMinimum(t *testing.T) {
	headers := []string{"A", "B", "C"}
	rows := [][]string{{strings.Repeat("a", 50), strings.Repeat("b", 50), strings.Repeat("c", 50)}}

	widths := fitWidthsTo(headers, rows, 10) // impossible to fit
	require.NotNil(t, widths)

	for i, w := range widths {
		assert.GreaterOrEqual(t, w, minColWidth, "column %d shrunk below the minimum", i)
	}
}

func TestFitWidthsToRaggedRowsDoNotPanic(t *testing.T) {
	headers := []string{"A"}
	rows := [][]string{{"a", "extra", "columns"}}

	assert.NotPanics(t, func() { fitWidthsTo(headers, rows, 10) })
}

func TestColumnIDsOrder(t *testing.T) {
	assert.Equal(t, []string{"name", "system", "size", "unpacked", "dls", "hash", "url"}, ColumnIDs())
}

func TestSelectColumns(t *testing.T) {
	// Requested out of order, with noise in case and spacing:
	// the result must follow canonical display order.
	cols, err := selectColumns([]string{" URL ", "Name", "size"})
	require.NoError(t, err)

	got := make([]string, len(cols))
	for i, c := range cols {
		got[i] = c.id
	}

	assert.Equal(t, []string{"name", "size", "url"}, got)
}

func TestSelectColumnsUnknown(t *testing.T) {
	_, err := selectColumns([]string{"name", "bogus"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bogus", "error must name the bad column")
}

// captureStdout runs fn while redirecting os.Stdout into a pipe and
// returns everything written. Render targets os.Stdout directly, and a
// pipe also disables terminal-width squeezing, keeping output stable.
func captureStdout(t *testing.T, fn func() error) string {
	t.Helper()

	orig := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
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

	require.NoError(t, fnErr)
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
		assert.Contains(t, out, want)
	}
	for _, banned := range []string{"https://example.com/1.zip", "Sega Mega Drive", "4DCFD55C"} {
		assert.NotContains(t, out, banned, "unselected column leaked into output")
	}
}

func TestPrintROMsUnknownColumn(t *testing.T) {
	err := PrintROMs([]ds.ROM{{Name: "x"}}, []string{"nope"})
	require.Error(t, err)
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
		assert.Contains(t, out, want)
	}
}
