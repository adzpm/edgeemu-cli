package client

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adzpm/edgeemu-cli/internal/fixtures"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	return New(WithBaseURL(srv.URL))
}

func TestSearch(t *testing.T) {
	var gotMethod, gotQuery, gotSystem string

	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotQuery = r.FormValue("search")
		gotSystem = r.FormValue("system")

		w.Write([]byte(fixtures.SearchPage))
	})

	roms, err := c.Search(context.Background(), "sonic & co", "sega-genesis")
	require.NoError(t, err)

	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "sonic & co", gotQuery)
	assert.Equal(t, "sega-genesis", gotSystem)

	require.Len(t, roms, 2)

	first := roms[0]
	assert.Equal(t, "Sonic & Knuckles (World)", first.Name, "entities must be unescaped")
	assert.Equal(t, c.baseURL+"/download/sega-genesis/Sonic%20%26%20Knuckles%20%28World%29.zip", first.URL)
	assert.Equal(t, "1.36m", first.Size)
	assert.Equal(t, 341, first.Downloads)
	assert.Equal(t, "Sega Mega Drive / Genesis", first.System)
	assert.Equal(t, "256.00k", first.UnpackedSize)
	assert.Equal(t, "4DCFD55C 0658F691", first.Hash)

	assert.Empty(t, roms[1].Hash, "empty hash must parse as empty string")
}

func TestSearchNoResults(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(fixtures.EmptyPage))
	})

	roms, err := c.Search(context.Background(), "zzz", "all")
	require.NoError(t, err)
	assert.Empty(t, roms)
}

func TestSearchHTTPError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	})

	_, err := c.Search(context.Background(), "sonic", "all")
	require.Error(t, err)
}

func TestSearchContextCancelled(t *testing.T) {
	c := newTestClient(t, func(_ http.ResponseWriter, r *http.Request) {
		// Drain the body first: the server only notices a dropped client
		// (and cancels r.Context()) once the request body is consumed.
		io.Copy(io.Discard, r.Body)
		<-r.Context().Done()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := c.Search(ctx, "sonic", "all")
	require.Error(t, err)
}

func TestSystems(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		w.Write([]byte(fixtures.SystemsPage))
	})

	systems, err := c.Systems(context.Background())
	require.NoError(t, err)

	require.Len(t, systems, 3, "the 'all' option must be skipped")
	assert.Equal(t, "atari-2600", systems[0].ID)
	assert.Equal(t, "Atari 2600", systems[0].Name)
	assert.Equal(t, "sega-genesis", systems[1].ID)
	assert.Equal(t, "Microsoft MSX / MSX2", systems[2].Name)
}

func TestSystemsEmptyPageIsError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`<html><body>layout changed</body></html>`))
	})

	_, err := c.Systems(context.Background())
	require.Error(t, err, "zero parsed systems must be reported, not cached")
}

func TestSystemsHTTPError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusBadGateway)
	})

	_, err := c.Systems(context.Background())
	require.Error(t, err)
}

func TestWithBaseURLTrimsTrailingSlash(t *testing.T) {
	c := New(WithBaseURL("https://example.com///"))
	assert.Equal(t, "https://example.com", c.baseURL)
}

func TestWithHTTPClientNilIgnored(t *testing.T) {
	c := New(WithHTTPClient(nil))
	assert.NotNil(t, c.http, "nil http client must not replace the default")
}

func TestWithHTTPClientOverrides(t *testing.T) {
	custom := &http.Client{Timeout: time.Second}
	c := New(WithHTTPClient(custom))
	assert.Same(t, custom, c.http)
}
