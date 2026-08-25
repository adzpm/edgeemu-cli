package app

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adzpm/edgeemu-cli/internal/client"
	"github.com/adzpm/edgeemu-cli/internal/fixtures"
	"github.com/adzpm/edgeemu-cli/internal/render"
)

// runDump executes the dump command against a stub site serving the
// systems page, the browse letters page and two letter buckets.
func runDump(t *testing.T, args ...string) (string, error) {
	t.Helper()
	sandboxCacheDir(t)

	mux := http.NewServeMux()
	mux.HandleFunc("/search.php", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(fixtures.SystemsPage))
	})
	mux.HandleFunc("/browse/sega-genesis", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(fixtures.BrowseLettersPage))
	})
	mux.HandleFunc("/browse/sega-genesis/q", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(fixtures.BrowseQPage))
	})
	mux.HandleFunc("/browse/sega-genesis/s", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(fixtures.BrowseSPage))
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	var buf bytes.Buffer

	root := New(
		WithClient(client.New(client.WithBaseURL(srv.URL))),
		WithPrinter(render.New(render.WithWriter(&buf))),
	).Root()
	root.Writer = &buf

	err := root.Run(context.Background(), append([]string{"edgeemu", "dump", "--delay", "0"}, args...))

	return buf.String(), err
}

func TestDumpSingleSystem(t *testing.T) {
	out, err := runDump(t, "-s", "sega-genesis", "-f", "json")
	require.NoError(t, err)

	var roms []map[string]any

	require.NoError(t, json.Unmarshal([]byte(out), &roms), "output is not valid JSON:\n%s", out)

	require.Len(t, roms, 3, "both letter buckets must be collected")
	assert.Equal(t, "QuackShot Starring Donald Duck (World) (En,Ja) (Rev A)", roms[0]["name"])
	assert.Equal(t, "Sonic & Knuckles (World)", roms[1]["name"])
	assert.Equal(t, "Sega Mega Drive / Genesis", roms[1]["system"],
		"the system name must be filled in from the systems list")
	assert.Equal(t, "4DCFD55C 0658F691", roms[1]["hash"])
}

func TestDumpHonorsColumns(t *testing.T) {
	out, err := runDump(t, "-s", "sega-genesis", "-f", "csv", "-c", "name,dls")
	require.NoError(t, err)

	rows, err := csv.NewReader(strings.NewReader(out)).ReadAll()
	require.NoError(t, err)

	require.Len(t, rows, 4, "header plus three ROMs")
	assert.Equal(t, []string{"name", "dls"}, rows[0])
	assert.Equal(t, []string{"QuackShot Starring Donald Duck (World) (En,Ja) (Rev A)", "56"}, rows[1],
		"a name containing commas must survive CSV quoting")
}

func TestDumpUnknownSystemFails(t *testing.T) {
	_, err := runDump(t, "-s", "nintendo-virtual-fridge")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nintendo-virtual-fridge")
}

func TestDumpValidatesFormatBeforeCrawling(t *testing.T) {
	_, err := runDump(t, "-s", "sega-genesis", "-f", "tsv")
	require.Error(t, err)
}

func TestDumpValidatesColumnsBeforeCrawling(t *testing.T) {
	_, err := runDump(t, "-s", "sega-genesis", "-c", "bogus")
	require.Error(t, err)
}
