package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adzpm/edgeemu-cli/client"
	"github.com/adzpm/edgeemu-cli/ds"
)

const searchPage = `<html><body><div class="grid">
  <div class="item">
    <details data-name="Sonic The Hedgehog (USA, Europe).zip">
      <summary>Sonic The Hedgehog (USA, Europe)</summary>
      <p><a href="/download/sega-genesis/Sonic.zip">download</a> (<span>377.87k, 588 DLs</span>)</p>
      <p>system: <span>Sega Mega Drive / Genesis</span></p>
      <p>unpacked size: <span>512.00k</span></p>
      <p>hash: <span>F9394E97</span></p>
    </details>
  </div>
  <div class="item">
    <details data-name="Sonic The Hedgehog 2 (World).zip">
      <summary>Sonic The Hedgehog 2 (World)</summary>
      <p><a href="/download/sega-genesis/Sonic2.zip">download</a> (<span>732.08k, 432 DLs</span>)</p>
      <p>system: <span>Sega Mega Drive / Genesis</span></p>
      <p>unpacked size: <span>1.00m</span></p>
      <p>hash: <span>24AB4C3A</span></p>
    </details>
  </div>
</div></body></html>`

const emptyPage = `<html><body><div class="grid"></div></body></html>`

const systemsPage = `<select name="system">
<option value="all" selected>Search all</option>
<option value="atari-2600">Atari 2600</option>
</select>`

// runCLI executes the root command against a stub server and returns
// what the command wrote to its writer.
func runCLI(t *testing.T, page string, args ...string) (string, error) {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(page))
	}))
	t.Cleanup(srv.Close)

	root := newRootCmd(client.New(client.WithBaseURL(srv.URL)))

	var buf bytes.Buffer
	root.Writer = &buf

	err := root.Run(context.Background(), append([]string{"edgeemu"}, args...))

	return buf.String(), err
}

func TestSearchJSON(t *testing.T) {
	out, err := runCLI(t, searchPage, "search", "sonic", "--json")
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	var roms []ds.ROM
	if err := json.Unmarshal([]byte(out), &roms); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, out)
	}
	if len(roms) != 2 {
		t.Fatalf("got %d roms, want 2", len(roms))
	}
	if roms[0].Name != "Sonic The Hedgehog (USA, Europe)" {
		t.Errorf("first rom = %+v", roms[0])
	}
}

func TestSearchJSONEmptyIsArray(t *testing.T) {
	out, err := runCLI(t, emptyPage, "search", "nothing-here", "--json")
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	if strings.TrimSpace(out) != "[]" {
		t.Errorf("empty result JSON = %q, want []", strings.TrimSpace(out))
	}
}

func TestSearchLimit(t *testing.T) {
	out, err := runCLI(t, searchPage, "search", "sonic", "--json", "-l", "1")
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	var roms []ds.ROM
	if err := json.Unmarshal([]byte(out), &roms); err != nil {
		t.Fatalf("bad JSON: %v", err)
	}
	if len(roms) != 1 {
		t.Errorf("got %d roms with -l 1, want 1", len(roms))
	}
}

func TestSearchNoResultsMessage(t *testing.T) {
	out, err := runCLI(t, emptyPage, "search", "nothing-here")
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	if !strings.Contains(out, "nothing found") {
		t.Errorf("output = %q, want 'nothing found'", out)
	}
}

func TestSearchEmptyQueryFails(t *testing.T) {
	if _, err := runCLI(t, emptyPage, "search"); err == nil {
		t.Fatal("want usage error for missing query, got nil")
	}
	if _, err := runCLI(t, emptyPage, "search", "   "); err == nil {
		t.Fatal("want usage error for blank query, got nil")
	}
}

func TestSearchUnknownColumnFails(t *testing.T) {
	if _, err := runCLI(t, searchPage, "search", "sonic", "-c", "bogus"); err == nil {
		t.Fatal("want error for unknown column, got nil")
	}
}

func TestSystemsCommand(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CACHE_HOME", filepath.Join(tmp, ".cache"))
	t.Setenv("LocalAppData", filepath.Join(tmp, "LocalAppData"))

	if _, err := runCLI(t, systemsPage, "systems"); err != nil {
		t.Fatalf("systems: %v", err)
	}
}
