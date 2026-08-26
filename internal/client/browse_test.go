package client

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adzpm/edgeemu-cli/internal/fixtures"
)

func TestBrowseLetters(t *testing.T) {
	var gotPath string

	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path

		w.Write([]byte(fixtures.BrowseLettersPage))
	})

	letters, err := c.BrowseLetters(context.Background(), "sega-genesis")
	require.NoError(t, err)

	assert.Equal(t, "/browse/sega-genesis", gotPath)
	assert.Equal(t, []string{"q", "s"}, letters, `the "-" placeholder must be skipped`)
}

func TestBrowseLettersEmptyPageIsError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`<html><body>layout changed</body></html>`))
	})

	_, err := c.BrowseLetters(context.Background(), "sega-genesis")
	require.Error(t, err)
}

func TestBrowse(t *testing.T) {
	var gotPath string

	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path

		w.Write([]byte(fixtures.BrowseSPage))
	})

	roms, err := c.Browse(context.Background(), "sega-genesis", "s")
	require.NoError(t, err)

	assert.Equal(t, "/browse/sega-genesis/s", gotPath)
	require.Len(t, roms, 2)

	first := roms[0]
	assert.Equal(t, "Sonic & Knuckles (World)", first.Name, "entities must be unescaped")
	assert.Equal(t, c.baseURL+"/download/demo-system/Game%20One%20%28Demo%29%20%26%20Co.zip", first.URL)
	assert.Equal(t, "1.36m", first.Size)
	assert.Equal(t, 341, first.Downloads)
	assert.Equal(t, "256.00k", first.UnpackedSize)
	assert.Equal(t, "4DCFD55C 0658F691", first.Hash, "hash after &nbsp; must parse")
	assert.Empty(t, first.System, "browse pages carry no system name")
}

func TestBrowseEmptyBucket(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`<html><body><div class="grid"></div></body></html>`))
	})

	roms, err := c.Browse(context.Background(), "sega-genesis", "x")
	require.NoError(t, err)
	assert.Empty(t, roms)
}

func TestBrowseHTTPError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	})

	_, err := c.Browse(context.Background(), "sega-genesis", "s")
	require.Error(t, err)
}
