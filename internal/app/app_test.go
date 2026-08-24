package app

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/adzpm/edgeemu-cli/internal/cache"
	"github.com/adzpm/edgeemu-cli/internal/client"
	"github.com/adzpm/edgeemu-cli/internal/completion"
	"github.com/adzpm/edgeemu-cli/internal/ds"
	"github.com/adzpm/edgeemu-cli/internal/fixtures"
	"github.com/adzpm/edgeemu-cli/internal/render"
)

// runCLI executes the root command against a stub server and returns
// everything the command printed.
func runCLI(t *testing.T, page string, args ...string) (string, error) {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(page))
	}))
	t.Cleanup(srv.Close)

	var buf bytes.Buffer
	root := New(
		WithClient(client.New(client.WithBaseURL(srv.URL))),
		WithPrinter(render.New(render.WithWriter(&buf))),
	).Root()
	root.Writer = &buf

	err := root.Run(context.Background(), append([]string{"edgeemu"}, args...))

	return buf.String(), err
}

func sandboxCacheDir(t *testing.T) {
	t.Helper()

	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CACHE_HOME", filepath.Join(tmp, ".cache"))
	t.Setenv("LocalAppData", filepath.Join(tmp, "LocalAppData"))
}

func TestNewAppliesAllOptions(t *testing.T) {
	edge := client.New()
	c := cache.New(cache.WithClient(edge))
	p := render.New()
	comp := completion.New(completion.WithCache(c))

	a := New(WithClient(edge), WithCache(c), WithPrinter(p), WithCompletion(comp))

	assert.Same(t, edge, a.edge)
	assert.Same(t, c, a.cache)
	assert.Same(t, p, a.printer)
	assert.Same(t, comp, a.comp)
}

func TestNewBuildsDefaults(t *testing.T) {
	a := New()

	assert.NotNil(t, a.edge)
	assert.NotNil(t, a.cache)
	assert.NotNil(t, a.printer)
	assert.NotNil(t, a.comp)
	assert.NotNil(t, a.Root())
}

func TestSearchDefaultsToListWithAllFields(t *testing.T) {
	out, err := runCLI(t, fixtures.SearchPage, "search", "sonic", "-s", "sega-genesis")
	require.NoError(t, err)

	assert.Contains(t, out, "1. Sonic & Knuckles (World)")
	assert.Contains(t, out, "2. Sonic The Hedgehog (USA, Europe)")
	// The full URL must be present, untruncated, on its own line.
	assert.Contains(t, out, "/download/sega-genesis/Sonic%20%26%20Knuckles%20%28World%29.zip")
	// Every field is shown by default, explicitly.
	assert.Contains(t, out, "system: Sega Mega Drive / Genesis")
	assert.Contains(t, out, "size: 1.36m")
	assert.Contains(t, out, "unpacked: 256.00k")
	assert.Contains(t, out, "dls: 341")
	assert.Contains(t, out, "hash: 4DCFD55C 0658F691")
}

func TestSearchColumnsNarrowTheList(t *testing.T) {
	out, err := runCLI(t, fixtures.SearchPage, "search", "sonic", "-c", "name,url")
	require.NoError(t, err)

	assert.Contains(t, out, "1. Sonic & Knuckles (World)")
	assert.NotContains(t, out, "system:")
	assert.NotContains(t, out, "size:")
}

func TestSearchJSON(t *testing.T) {
	out, err := runCLI(t, fixtures.SearchPage, "search", "sonic", "-f", "json")
	require.NoError(t, err)

	var roms []map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &roms), "output is not valid JSON:\n%s", out)

	require.Len(t, roms, 2)
	assert.Equal(t, "Sonic & Knuckles (World)", roms[0]["name"])
	assert.EqualValues(t, 341, roms[0]["dls"], "dls must stay numeric")
	assert.Equal(t, "256.00k", roms[0]["unpacked"])
}

func TestSearchJSONHonorsColumns(t *testing.T) {
	out, err := runCLI(t, fixtures.SearchPage, "search", "sonic", "-f", "json", "-c", "name,dls")
	require.NoError(t, err)

	var roms []map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &roms))

	require.Len(t, roms, 2)
	assert.Equal(t, map[string]any{"name": "Sonic & Knuckles (World)", "dls": float64(341)}, roms[0],
		"only the selected fields must be encoded")
}

func TestSearchYAML(t *testing.T) {
	out, err := runCLI(t, fixtures.SearchPage, "search", "sonic", "--format", "yaml")
	require.NoError(t, err)

	var roms []map[string]any
	require.NoError(t, yaml.Unmarshal([]byte(out), &roms), "output is not valid YAML:\n%s", out)

	require.Len(t, roms, 2)
	assert.Equal(t, "Sonic & Knuckles (World)", roms[0]["name"])
	assert.Equal(t, "256.00k", roms[0]["unpacked"])
}

func TestSearchYAMLHonorsColumns(t *testing.T) {
	out, err := runCLI(t, fixtures.SearchPage, "search", "sonic", "-f", "yaml", "-c", "name")
	require.NoError(t, err)

	var roms []map[string]any
	require.NoError(t, yaml.Unmarshal([]byte(out), &roms))

	require.Len(t, roms, 2)
	assert.Equal(t, map[string]any{"name": "Sonic & Knuckles (World)"}, roms[0])
}

func TestSearchXML(t *testing.T) {
	out, err := runCLI(t, fixtures.SearchPage, "search", "sonic", "-f", "xml")
	require.NoError(t, err)

	var doc struct {
		ROMs []struct {
			Name string `xml:"name"`
			DLs  int    `xml:"dls"`
		} `xml:"rom"`
	}
	require.NoError(t, xml.Unmarshal([]byte(out), &doc), "output is not valid XML:\n%s", out)

	require.Len(t, doc.ROMs, 2)
	assert.Equal(t, "Sonic & Knuckles (World)", doc.ROMs[0].Name)
	assert.Equal(t, 341, doc.ROMs[0].DLs)
	assert.True(t, strings.HasPrefix(out, xml.Header), "XML output must start with the declaration")
}

func TestSearchXMLHonorsColumns(t *testing.T) {
	out, err := runCLI(t, fixtures.SearchPage, "search", "sonic", "-f", "xml", "-c", "name,url")
	require.NoError(t, err)

	assert.Contains(t, out, "<name>")
	assert.Contains(t, out, "<url>")
	assert.NotContains(t, out, "<hash>", "unselected fields must not be encoded")
	assert.NotContains(t, out, "<dls>")
}

func TestSystemsXML(t *testing.T) {
	sandboxCacheDir(t)

	out, err := runCLI(t, fixtures.SystemsPage, "systems", "-f", "xml")
	require.NoError(t, err)

	var doc struct {
		Systems []ds.System `xml:"system"`
	}
	require.NoError(t, xml.Unmarshal([]byte(out), &doc))
	assert.Len(t, doc.Systems, 3)
}

func TestSearchCSV(t *testing.T) {
	out, err := runCLI(t, fixtures.SearchPage, "search", "sonic", "-f", "csv")
	require.NoError(t, err)

	rows, err := csv.NewReader(strings.NewReader(out)).ReadAll()
	require.NoError(t, err, "output is not valid CSV:\n%s", out)

	require.Len(t, rows, 3, "header plus two ROMs")
	assert.Equal(t, []string{"name", "system", "size", "unpacked", "dls", "hash", "url"}, rows[0])
	assert.Equal(t, "Sonic & Knuckles (World)", rows[1][0])
	assert.Equal(t, "341", rows[1][4])
}

func TestSearchCSVHonorsColumns(t *testing.T) {
	out, err := runCLI(t, fixtures.SearchPage, "search", "sonic", "-f", "csv", "-c", "name,dls")
	require.NoError(t, err)

	rows, err := csv.NewReader(strings.NewReader(out)).ReadAll()
	require.NoError(t, err)

	require.Len(t, rows, 3)
	assert.Equal(t, []string{"name", "dls"}, rows[0])
	assert.Equal(t, []string{"Sonic & Knuckles (World)", "341"}, rows[1])
}

func TestSystemsCSV(t *testing.T) {
	sandboxCacheDir(t)

	out, err := runCLI(t, fixtures.SystemsPage, "systems", "-f", "csv")
	require.NoError(t, err)

	rows, err := csv.NewReader(strings.NewReader(out)).ReadAll()
	require.NoError(t, err)

	require.Len(t, rows, 4, "header plus three systems")
	assert.Equal(t, []string{"id", "name"}, rows[0])
	assert.Equal(t, []string{"atari-2600", "Atari 2600"}, rows[1])
}

func TestSearchJSONEmptyIsArray(t *testing.T) {
	out, err := runCLI(t, fixtures.EmptyPage, "search", "nothing-here", "-f", "json")
	require.NoError(t, err)

	assert.Equal(t, "[]", strings.TrimSpace(out), "empty result must encode as [], not null")
}

func TestSearchUnknownFormatFails(t *testing.T) {
	_, err := runCLI(t, fixtures.SearchPage, "search", "sonic", "-f", "tsv")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tsv")
}

func TestSearchLimit(t *testing.T) {
	out, err := runCLI(t, fixtures.SearchPage, "search", "sonic", "-f", "json", "-l", "1")
	require.NoError(t, err)

	var roms []map[string]any
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
	sandboxCacheDir(t)

	out, err := runCLI(t, fixtures.SystemsPage, "systems")
	require.NoError(t, err)

	assert.Contains(t, out, "atari-2600")
	assert.Contains(t, out, "Sega Mega Drive / Genesis")
}

func TestSystemsJSON(t *testing.T) {
	sandboxCacheDir(t)

	out, err := runCLI(t, fixtures.SystemsPage, "systems", "-f", "json")
	require.NoError(t, err)

	var systems []ds.System
	require.NoError(t, json.Unmarshal([]byte(out), &systems))
	assert.Len(t, systems, 3)
}

func TestSystemsYAML(t *testing.T) {
	sandboxCacheDir(t)

	out, err := runCLI(t, fixtures.SystemsPage, "systems", "-f", "yaml")
	require.NoError(t, err)

	var systems []ds.System
	require.NoError(t, yaml.Unmarshal([]byte(out), &systems))
	assert.Len(t, systems, 3)
	assert.Equal(t, "atari-2600", systems[0].ID)
}

func TestSystemsUnknownFormatFails(t *testing.T) {
	sandboxCacheDir(t)

	_, err := runCLI(t, fixtures.SystemsPage, "systems", "-f", "tsv")
	require.Error(t, err)
}
