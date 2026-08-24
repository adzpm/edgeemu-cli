package render

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adzpm/edgeemu-cli/internal/ds"
)

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

func TestPrintROMs(t *testing.T) {
	roms := []ds.ROM{
		{
			Name: "Sonic & Knuckles (World)", System: "Sega Mega Drive / Genesis",
			URL: "https://example.com/very/long/url/1.zip", Size: "1.36m", Downloads: 341, Hash: "4DCFD55C",
		},
		{
			Name: "Sonic The Hedgehog 2 (World)", System: "Sega Mega Drive / Genesis",
			URL: "https://example.com/2.zip", Size: "732.08k", Downloads: 432, Hash: "24AB4C3A",
		},
	}

	var buf bytes.Buffer

	require.NoError(t, New(WithWriter(&buf)).PrintROMs(roms, []string{"name", "system", "size", "url"}))
	out := buf.String()

	assert.Contains(t, out, "1. Sonic & Knuckles (World)")
	assert.Contains(t, out, "2. Sonic The Hedgehog 2 (World)")
	assert.Contains(t, out, "system: Sega Mega Drive / Genesis · size: 1.36m")
	assert.NotContains(t, out, "hash:", "unselected fields must not appear")

	// The URL must sit alone on its own line, untouched.
	var urlLines int

	for line := range strings.SplitSeq(out, "\n") {
		if strings.Contains(line, "https://") {
			urlLines++

			assert.Equal(t, "https", strings.SplitN(strings.TrimSpace(line), ":", 2)[0])
			assert.NotContains(t, line, "·", "URL line must not carry other fields")
		}
	}

	assert.Equal(t, 2, urlLines)
}

func TestPrintROMsMinimalColumns(t *testing.T) {
	roms := []ds.ROM{{Name: "Sonic", URL: "https://example.com/1.zip", Size: "1m"}}

	var buf bytes.Buffer

	require.NoError(t, New(WithWriter(&buf)).PrintROMs(roms, []string{"name", "url"}))
	out := buf.String()

	assert.Contains(t, out, "1. Sonic")
	assert.Contains(t, out, "https://example.com/1.zip")
	assert.NotContains(t, out, "size:", "unselected fields must not appear")
}

func TestPrintROMsUnknownColumn(t *testing.T) {
	var buf bytes.Buffer

	err := New(WithWriter(&buf)).PrintROMs([]ds.ROM{{Name: "x"}}, []string{"nope"})
	require.Error(t, err)
}

func TestPrintSystemsAligned(t *testing.T) {
	systems := []ds.System{
		{ID: "atari-2600", Name: "Atari 2600"},
		{ID: "sega-genesis", Name: "Sega Mega Drive / Genesis"},
	}

	var buf bytes.Buffer

	require.NoError(t, New(WithWriter(&buf)).PrintSystems(systems))

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	require.Len(t, lines, 2)

	assert.Equal(t, "atari-2600    Atari 2600", lines[0])
	assert.Equal(t, "sega-genesis  Sega Mega Drive / Genesis", lines[1])
}
