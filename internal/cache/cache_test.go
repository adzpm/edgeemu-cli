package cache

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adzpm/edgeemu-cli/internal/client"
	"github.com/adzpm/edgeemu-cli/internal/ds"
	"github.com/adzpm/edgeemu-cli/internal/fixtures"
)

func cachePath(t *testing.T) string {
	t.Helper()

	return filepath.Join(t.TempDir(), "systems.json")
}

func newTestClient(t *testing.T, hits *atomic.Int32) *client.Client {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.Write([]byte(fixtures.SystemsPage))
	}))
	t.Cleanup(srv.Close)

	return client.New(client.WithBaseURL(srv.URL))
}

func TestSystemsFetchesOnceThenServesFromCache(t *testing.T) {
	var hits atomic.Int32
	c := New(WithClient(newTestClient(t, &hits)), WithPath(cachePath(t)))

	first, err := c.Systems(context.Background(), false)
	require.NoError(t, err)
	require.Len(t, first, 3)
	require.EqualValues(t, 1, hits.Load())

	second, err := c.Systems(context.Background(), false)
	require.NoError(t, err)
	assert.EqualValues(t, 1, hits.Load(), "second call must be served from cache")
	assert.Equal(t, first, second)
}

func TestSystemsRefreshBypassesCache(t *testing.T) {
	var hits atomic.Int32
	c := New(WithClient(newTestClient(t, &hits)), WithPath(cachePath(t)))

	_, err := c.Systems(context.Background(), false)
	require.NoError(t, err)

	_, err = c.Systems(context.Background(), true)
	require.NoError(t, err)

	assert.EqualValues(t, 2, hits.Load(), "refresh must hit the network")
}

func TestLoadExpiry(t *testing.T) {
	p := cachePath(t)
	c := New(WithPath(p))

	// Write a cache stamped older than the TTL directly.
	stale := systemsCache{
		FetchedAt: time.Now().Add(-TTL - time.Hour),
		Systems:   []ds.System{{ID: "atari-2600", Name: "Atari 2600"}},
	}
	data, err := json.Marshal(stale)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(p, data, 0o644))

	assert.Nil(t, c.Load(TTL), "expired cache must not be returned")
	assert.Len(t, c.Load(0), 1, "zero maxAge accepts a cache of any age")
}

func TestLoadMissingAndCorrupt(t *testing.T) {
	p := cachePath(t)
	c := New(WithPath(p))

	assert.Nil(t, c.Load(0), "missing cache file must load as nil")

	require.NoError(t, os.WriteFile(p, []byte("{not json"), 0o644))
	assert.Nil(t, c.Load(0), "corrupt cache file must load as nil")
}

func TestCustomTTL(t *testing.T) {
	var hits atomic.Int32
	c := New(WithClient(newTestClient(t, &hits)), WithPath(cachePath(t)), WithTTL(time.Nanosecond))

	_, err := c.Systems(context.Background(), false)
	require.NoError(t, err)

	time.Sleep(time.Millisecond)

	_, err = c.Systems(context.Background(), false)
	require.NoError(t, err)

	assert.EqualValues(t, 2, hits.Load(), "expired TTL must refetch")
}

func TestSystemsFetchErrorIsReturned(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "down", http.StatusServiceUnavailable)
	}))
	t.Cleanup(srv.Close)

	c := New(WithClient(client.New(client.WithBaseURL(srv.URL))), WithPath(cachePath(t)))
	_, err := c.Systems(context.Background(), false)
	require.Error(t, err, "fetch failure with no cache must surface")
}

func TestSystemsWithoutClientErrors(t *testing.T) {
	c := New(WithPath(cachePath(t)))

	_, err := c.Systems(context.Background(), false)
	require.Error(t, err, "cache miss without a client must error, not panic")
}
