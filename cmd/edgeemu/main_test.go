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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adzpm/edgeemu-cli/internal/client"
	"github.com/adzpm/edgeemu-cli/internal/ds"
	"github.com/adzpm/edgeemu-cli/internal/fixtures"
)

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
	out, err := runCLI(t, fixtures.SearchPage, "search", "sonic", "--json")
	require.NoError(t, err)

	var roms []ds.ROM
	require.NoError(t, json.Unmarshal([]byte(out), &roms), "output is not valid JSON:\n%s", out)

	require.Len(t, roms, 2)
	assert.Equal(t, "Sonic & Knuckles (World)", roms[0].Name)
	assert.Equal(t, 341, roms[0].Downloads)
}

func TestSearchJSONEmptyIsArray(t *testing.T) {
	out, err := runCLI(t, fixtures.EmptyPage, "search", "nothing-here", "--json")
	require.NoError(t, err)

	assert.Equal(t, "[]", strings.TrimSpace(out), "empty result must encode as [], not null")
}

func TestSearchLimit(t *testing.T) {
	out, err := runCLI(t, fixtures.SearchPage, "search", "sonic", "--json", "-l", "1")
	require.NoError(t, err)

	var roms []ds.ROM
	require.NoError(t, json.Unmarshal([]byte(out), &roms))
	assert.Len(t, roms, 1)
}

func TestSearchNoResultsMessage(t *testing.T) {
	out, err := runCLI(t, fixtures.EmptyPage, "search", "nothing-here")
	require.NoError(t, err)

	assert.Contains(t, out, "nothing found")
}

func TestSearchEmptyQueryFails(t *testing.T) {
	_, err := runCLI(t, fixtures.EmptyPage, "search")
	require.Error(t, err, "missing query must fail")

	_, err = runCLI(t, fixtures.EmptyPage, "search", "   ")
	require.Error(t, err, "blank query must fail")
}

func TestSearchUnknownColumnFails(t *testing.T) {
	_, err := runCLI(t, fixtures.SearchPage, "search", "sonic", "-c", "bogus")
	require.Error(t, err)
}

func TestSystemsCommand(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CACHE_HOME", filepath.Join(tmp, ".cache"))
	t.Setenv("LocalAppData", filepath.Join(tmp, "LocalAppData"))

	_, err := runCLI(t, fixtures.SystemsPage, "systems")
	require.NoError(t, err)
}
