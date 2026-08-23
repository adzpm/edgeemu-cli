package completion

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	"github.com/adzpm/edgeemu-cli/internal/cache"
	"github.com/adzpm/edgeemu-cli/internal/client"
	"github.com/adzpm/edgeemu-cli/internal/fixtures"
	"github.com/adzpm/edgeemu-cli/internal/table"
)

// sandboxCacheDir points the user cache directory into a temp dir on any OS.
func sandboxCacheDir(t *testing.T) {
	t.Helper()

	tmp := t.TempDir()
	t.Setenv("HOME", tmp)                                        // darwin
	t.Setenv("XDG_CACHE_HOME", filepath.Join(tmp, ".cache"))     // linux
	t.Setenv("LocalAppData", filepath.Join(tmp, "LocalAppData")) // windows
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
		w.Write([]byte(fixtures.SystemsPage))
	}))
	t.Cleanup(srv.Close)

	edge := client.New(client.WithBaseURL(srv.URL))

	// Prefill the cache, as a prior command or completion would have.
	_, err := cache.Systems(context.Background(), edge, false)
	require.NoError(t, err)

	out := runSearchComplete(t, edge, "-s")

	assert.EqualValues(t, 1, hits.Load(), "completion must be served from cache, not the network")
	assert.Contains(t, out, "atari-2600")
	assert.Contains(t, out, "sega-genesis")
}

func TestSearchCompletesSystemsByFetchingWhenNoCache(t *testing.T) {
	sandboxCacheDir(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(fixtures.SystemsPage))
	}))
	t.Cleanup(srv.Close)

	out := runSearchComplete(t, client.New(client.WithBaseURL(srv.URL)), "--system")

	assert.Contains(t, out, "atari-2600", "empty cache must fall back to a network fetch")
}

func TestSearchCompletesColumns(t *testing.T) {
	sandboxCacheDir(t)

	// The columns list is static: no client call should ever happen.
	out := runSearchComplete(t, client.New(client.WithBaseURL("http://127.0.0.1:0")), "-c")

	for _, want := range table.ColumnIDs() {
		assert.Contains(t, out, want)
	}
}
