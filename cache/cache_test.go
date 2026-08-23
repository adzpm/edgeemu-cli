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

	"github.com/adzpm/edgeemu-cli/client"
	"github.com/adzpm/edgeemu-cli/ds"
	"github.com/adzpm/edgeemu-cli/internal/fixtures"
)

// sandboxCacheDir points the user cache directory into a temp dir on any OS.
func sandboxCacheDir(t *testing.T) {
	t.Helper()

	tmp := t.TempDir()
	t.Setenv("HOME", tmp)                                        // darwin
	t.Setenv("XDG_CACHE_HOME", filepath.Join(tmp, ".cache"))     // linux
	t.Setenv("LocalAppData", filepath.Join(tmp, "LocalAppData")) // windows
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
	sandboxCacheDir(t)

	var hits atomic.Int32
	edge := newTestClient(t, &hits)

	first, err := Systems(context.Background(), edge, false)
	require.NoError(t, err)
	require.Len(t, first, 3)
	require.EqualValues(t, 1, hits.Load())

	second, err := Systems(context.Background(), edge, false)
	require.NoError(t, err)
	assert.EqualValues(t, 1, hits.Load(), "second call must be served from cache")
	assert.Equal(t, first, second)
}

func TestSystemsRefreshBypassesCache(t *testing.T) {
	sandboxCacheDir(t)

	var hits atomic.Int32
	edge := newTestClient(t, &hits)

	_, err := Systems(context.Background(), edge, false)
	require.NoError(t, err)

	_, err = Systems(context.Background(), edge, true)
	require.NoError(t, err)

	assert.EqualValues(t, 2, hits.Load(), "refresh must hit the network")
}

func TestLoadExpiry(t *testing.T) {
	sandboxCacheDir(t)

	// Write a cache stamped older than the TTL directly.
	p, err := path()
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o755))

	stale := systemsCache{
		FetchedAt: time.Now().Add(-TTL - time.Hour),
		Systems:   []ds.System{{ID: "atari-2600", Name: "Atari 2600"}},
	}
	data, err := json.Marshal(stale)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(p, data, 0o644))

	assert.Nil(t, Load(TTL), "expired cache must not be returned")
	assert.Len(t, Load(0), 1, "zero maxAge accepts a cache of any age")
}

func TestLoadMissingAndCorrupt(t *testing.T) {
	sandboxCacheDir(t)

	assert.Nil(t, Load(0), "missing cache file must load as nil")

	p, err := path()
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o755))
	require.NoError(t, os.WriteFile(p, []byte("{not json"), 0o644))

	assert.Nil(t, Load(0), "corrupt cache file must load as nil")
}

func TestSystemsFetchErrorIsReturned(t *testing.T) {
	sandboxCacheDir(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "down", http.StatusServiceUnavailable)
	}))
	t.Cleanup(srv.Close)

	edge := client.New(client.WithBaseURL(srv.URL))
	_, err := Systems(context.Background(), edge, false)
	require.Error(t, err, "fetch failure with no cache must surface")
}
