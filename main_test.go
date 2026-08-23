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
	require.NoError(t, err)

	var roms []ds.ROM
	require.NoError(t, json.Unmarshal([]byte(out), &roms), "output is not valid JSON:\n%s", out)

	require.Len(t, roms, 2)
	assert.Equal(t, "Sonic The Hedgehog (USA, Europe)", roms[0].Name)
	assert.Equal(t, 588, roms[0].Downloads)
}

func TestSearchJSONEmptyIsArray(t *testing.T) {
	out, err := runCLI(t, emptyPage, "search", "nothing-here", "--json")
	require.NoError(t, err)

	assert.Equal(t, "[]", strings.TrimSpace(out), "empty result must encode as [], not null")
}

func TestSearchLimit(t *testing.T) {
	out, err := runCLI(t, searchPage, "search", "sonic", "--json", "-l", "1")
	require.NoError(t, err)

	var roms []ds.ROM
	require.NoError(t, json.Unmarshal([]byte(out), &roms))
	assert.Len(t, roms, 1)
}

func TestSearchNoResultsMessage(t *testing.T) {
	out, err := runCLI(t, emptyPage, "search", "nothing-here")
	require.NoError(t, err)

	assert.Contains(t, out, "nothing found")
}

func TestSearchEmptyQueryFails(t *testing.T) {
	_, err := runCLI(t, emptyPage, "search")
	require.Error(t, err, "missing query must fail")

	_, err = runCLI(t, emptyPage, "search", "   ")
	require.Error(t, err, "blank query must fail")
}

func TestSearchUnknownColumnFails(t *testing.T) {
	_, err := runCLI(t, searchPage, "search", "sonic", "-c", "bogus")
	require.Error(t, err)
}

func TestSystemsCommand(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CACHE_HOME", filepath.Join(tmp, ".cache"))
	t.Setenv("LocalAppData", filepath.Join(tmp, "LocalAppData"))

	_, err := runCLI(t, systemsPage, "systems")
	require.NoError(t, err)
}
