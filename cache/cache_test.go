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

	"github.com/adzpm/edgeemu-cli/client"
	"github.com/adzpm/edgeemu-cli/ds"
)

const systemsPage = `<select name="system">
<option value="all" selected>Search all</option>
<option value="atari-2600">Atari 2600</option>
<option value="sega-genesis">Sega Mega Drive / Genesis</option>
</select>`

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
		w.Write([]byte(systemsPage))
	}))
	t.Cleanup(srv.Close)

	return client.New(client.WithBaseURL(srv.URL))
}

func TestSystemsFetchesOnceThenServesFromCache(t *testing.T) {
	sandboxCacheDir(t)

	var hits atomic.Int32
	edge := newTestClient(t, &hits)

	first, err := Systems(context.Background(), edge, false)
	if err != nil {
		t.Fatalf("first Systems: %v", err)
	}
	if len(first) != 2 {
		t.Fatalf("got %d systems, want 2", len(first))
	}
	if hits.Load() != 1 {
		t.Fatalf("hits = %d, want 1", hits.Load())
	}

	second, err := Systems(context.Background(), edge, false)
	if err != nil {
		t.Fatalf("second Systems: %v", err)
	}
	if hits.Load() != 1 {
		t.Fatalf("hits = %d after second call, want 1 (must be served from cache)", hits.Load())
	}
	if len(second) != len(first) {
		t.Fatalf("cached result differs: %d vs %d systems", len(second), len(first))
	}
}

func TestSystemsRefreshBypassesCache(t *testing.T) {
	sandboxCacheDir(t)

	var hits atomic.Int32
	edge := newTestClient(t, &hits)

	if _, err := Systems(context.Background(), edge, false); err != nil {
		t.Fatalf("first Systems: %v", err)
	}
	if _, err := Systems(context.Background(), edge, true); err != nil {
		t.Fatalf("refresh Systems: %v", err)
	}

	if hits.Load() != 2 {
		t.Fatalf("hits = %d, want 2 (refresh must hit the network)", hits.Load())
	}
}

func TestLoadExpiry(t *testing.T) {
	sandboxCacheDir(t)

	// Write a cache stamped older than the TTL directly.
	p, err := path()
	if err != nil {
		t.Fatalf("path: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	stale := systemsCache{
		FetchedAt: time.Now().Add(-TTL - time.Hour),
		Systems:   []ds.System{{ID: "atari-2600", Name: "Atari 2600"}},
	}
	data, _ := json.Marshal(stale)
	if err := os.WriteFile(p, data, 0o644); err != nil {
		t.Fatalf("write cache: %v", err)
	}

	if got := Load(TTL); got != nil {
		t.Errorf("Load(TTL) = %d systems, want nil for expired cache", len(got))
	}
	if got := Load(0); len(got) != 1 {
		t.Errorf("Load(0) = %d systems, want 1 (zero maxAge accepts any age)", len(got))
	}
}

func TestLoadMissingAndCorrupt(t *testing.T) {
	sandboxCacheDir(t)

	if got := Load(0); got != nil {
		t.Errorf("Load with no cache file = %v, want nil", got)
	}

	p, err := path()
	if err != nil {
		t.Fatalf("path: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(p, []byte("{not json"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	if got := Load(0); got != nil {
		t.Errorf("Load with corrupt cache = %v, want nil", got)
	}
}

func TestSystemsFetchErrorIsReturned(t *testing.T) {
	sandboxCacheDir(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "down", http.StatusServiceUnavailable)
	}))
	t.Cleanup(srv.Close)

	edge := client.New(client.WithBaseURL(srv.URL))
	if _, err := Systems(context.Background(), edge, false); err == nil {
		t.Fatal("want error when fetch fails and no cache exists, got nil")
	}
}
