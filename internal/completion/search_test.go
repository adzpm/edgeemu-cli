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
	"github.com/adzpm/edgeemu-cli/internal/render"
)

func cachePath(t *testing.T) string {
	t.Helper()

	return filepath.Join(t.TempDir(), "systems.json")
}

// runSearchComplete invokes the completion method as the shell script
// would, with lastArg as the word before --generate-shell-completion.
func runSearchComplete(t *testing.T, comp *Completion, lastArg string) string {
	t.Helper()

	origArgs := os.Args
	os.Args = []string{"edgeemu", "search", lastArg, "--generate-shell-completion"}
	defer func() { os.Args = origArgs }()

	var buf bytes.Buffer
	cmd := &cli.Command{Name: "edgeemu", Writer: &buf}

	comp.Search(context.Background(), cmd)

	return buf.String()
}

func TestSearchCompletesSystemsFromCacheWithoutNetwork(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.Write([]byte(fixtures.SystemsPage))
	}))
	t.Cleanup(srv.Close)

	c := cache.New(
		cache.WithClient(client.New(client.WithBaseURL(srv.URL))),
		cache.WithPath(cachePath(t)),
	)

	// Prefill the cache, as a prior command or completion would have.
	_, err := c.Systems(context.Background(), false)
	require.NoError(t, err)

	out := runSearchComplete(t, New(WithCache(c)), "-s")

	assert.EqualValues(t, 1, hits.Load(), "completion must be served from cache, not the network")
	assert.Contains(t, out, "atari-2600")
	assert.Contains(t, out, "sega-genesis")
}

func TestSearchCompletesSystemsByFetchingWhenNoCache(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(fixtures.SystemsPage))
	}))
	t.Cleanup(srv.Close)

	c := cache.New(
		cache.WithClient(client.New(client.WithBaseURL(srv.URL))),
		cache.WithPath(cachePath(t)),
	)

	out := runSearchComplete(t, New(WithCache(c)), "--system")

	assert.Contains(t, out, "atari-2600", "empty cache must fall back to a network fetch")
}

func TestSearchCompletesColumns(t *testing.T) {
	// The columns list is static: no cache or network is involved.
	out := runSearchComplete(t, New(), "-c")

	for _, want := range render.ColumnIDs() {
		assert.Contains(t, out, want)
	}
}

func TestSearchSystemsWithoutCacheIsSilent(t *testing.T) {
	out := runSearchComplete(t, New(), "-s")

	assert.Empty(t, out, "no cache configured must print nothing, not panic")
}

// runSearchCompleteArgs is runSearchComplete with full control over the
// words preceding --generate-shell-completion.
func runSearchCompleteArgs(t *testing.T, comp *Completion, words ...string) string {
	t.Helper()

	origArgs := os.Args
	os.Args = append(append([]string{"edgeemu", "search"}, words...), "--generate-shell-completion")
	defer func() { os.Args = origArgs }()

	var buf bytes.Buffer
	cmd := &cli.Command{
		Name:   "edgeemu",
		Writer: &buf,
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "system", Aliases: []string{"s"}, Usage: "system to search in"},
			&cli.StringSliceFlag{Name: "columns", Aliases: []string{"c"}, Usage: "columns to show"},
		},
	}

	comp.Search(context.Background(), cmd)

	return buf.String()
}

func TestSearchSuggestsFlagsAfterPositionalArg(t *testing.T) {
	// Regression: "edgeemu search sonic --<TAB>" must offer flags, not
	// subcommands (the default cli helper breaks on the positional arg).
	for _, words := range [][]string{
		{"sonic", "--"},
		{"sonic", "-"},
		{"sonic"},
		{"--"},
	} {
		out := runSearchCompleteArgs(t, New(), words...)

		assert.Contains(t, out, "--system:", "words=%v", words)
		assert.Contains(t, out, "--columns:", "words=%v", words)
	}
}
