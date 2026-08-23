package completion

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/urfave/cli/v3"

	"github.com/adzpm/edgeemu-cli/cache"
	"github.com/adzpm/edgeemu-cli/client"
	"github.com/adzpm/edgeemu-cli/table"
)

const systemsPage = `<select name="system">
<option value="all" selected>Search all</option>
<option value="atari-2600">Atari 2600</option>
<option value="sega-genesis">Sega Mega Drive / Genesis</option>
</select>`

func sandboxCacheDir(t *testing.T) {
	t.Helper()

	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CACHE_HOME", filepath.Join(tmp, ".cache"))
	t.Setenv("LocalAppData", filepath.Join(tmp, "LocalAppData"))
}

// runSearchComplete invokes the completion func as the shell script would,
// with lastArg as the word before --generate-shell-completion.
func runSearchComplete(t *testing.T, edge *client.Client, lastArg string) string {
	t.Helper()

	origArgs := os.Args
	os.Args = []string{"edgeemu", "search", lastArg, "--generate-shell-completion"}
	defer func() { os.Args = origArgs }()

	var buf bytes.Buffer
	cmd := &cli.Command{Name: "edgeemu", Writer: &buf}

	Search(edge)(context.Background(), cmd)

	return buf.String()
}

func TestSearchCompletesSystemsFromCacheWithoutNetwork(t *testing.T) {
	sandboxCacheDir(t)

	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.Write([]byte(systemsPage))
	}))
	t.Cleanup(srv.Close)

	edge := client.New(client.WithBaseURL(srv.URL))

	// Prefill the cache, as a prior command or completion would have.
	if _, err := cache.Systems(context.Background(), edge, false); err != nil {
		t.Fatalf("prefill cache: %v", err)
	}

	out := runSearchComplete(t, edge, "-s")

	if hits.Load() != 1 {
		t.Errorf("network hits = %d, want 1 (completion must be served from cache)", hits.Load())
	}
	for _, want := range []string{"atari-2600", "sega-genesis"} {
		if !strings.Contains(out, want) {
			t.Errorf("completion output missing %q:\n%s", want, out)
		}
	}
}

func TestSearchCompletesSystemsByFetchingWhenNoCache(t *testing.T) {
	sandboxCacheDir(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(systemsPage))
	}))
	t.Cleanup(srv.Close)

	out := runSearchComplete(t, client.New(client.WithBaseURL(srv.URL)), "--system")

	if !strings.Contains(out, "atari-2600") {
		t.Errorf("completion output missing systems fetched from network:\n%s", out)
	}
}

func TestSearchCompletesColumns(t *testing.T) {
	sandboxCacheDir(t)

	// The columns list is static: no client call should ever happen.
	out := runSearchComplete(t, client.New(client.WithBaseURL("http://127.0.0.1:0")), "-c")

	for _, want := range table.ColumnIDs() {
		if !strings.Contains(out, want) {
			t.Errorf("completion output missing column %q:\n%s", want, out)
		}
	}
}
