package cache

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/adzpm/edgeemu-cli/internal/client"
	"github.com/adzpm/edgeemu-cli/internal/ds"
)

// TTL is how long a cached systems list is considered fresh by default.
const TTL = 24 * time.Hour

// Cache stores the systems list on disk between runs.
type Cache struct {
	edge *client.Client
	ttl  time.Duration
	path string
}

// Option customizes a Cache.
type Option func(*Cache)

// WithClient sets the client used to fetch systems on a cache miss.
func WithClient(edge *client.Client) Option {
	return func(c *Cache) { c.edge = edge }
}

// WithTTL overrides how long a cached list is considered fresh.
func WithTTL(ttl time.Duration) Option {
	return func(c *Cache) { c.ttl = ttl }
}

// WithPath overrides the cache file location.
func WithPath(path string) Option {
	return func(c *Cache) { c.path = path }
}

// New creates a Cache with sane defaults, applying the given options.
// Without WithPath the cache lives in the user cache directory; without
// WithClient a cache miss returns an error instead of fetching.
func New(opts ...Option) *Cache {
	c := &Cache{ttl: TTL}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

type systemsCache struct {
	FetchedAt time.Time   `json:"fetched_at"`
	Systems   []ds.System `json:"systems"`
}

func (c *Cache) filePath() (string, error) {
	if c.path != "" {
		return c.path, nil
	}

	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "edgeemu", "systems.json"), nil
}

// Load returns the cached systems list, or nil if there is no usable
// cache. A maxAge of 0 accepts a cache of any age.
func (c *Cache) Load(maxAge time.Duration) []ds.System {
	p, err := c.filePath()
	if err != nil {
		return nil
	}

	data, err := os.ReadFile(p)
	if err != nil {
		return nil
	}

	var sc systemsCache
	if json.Unmarshal(data, &sc) != nil {
		return nil
	}

	if maxAge > 0 && time.Since(sc.FetchedAt) > maxAge {
		return nil
	}

	return sc.Systems
}

func (c *Cache) store(systems []ds.System) {
	p, err := c.filePath()
	if err != nil {
		return
	}

	if os.MkdirAll(filepath.Dir(p), 0o755) != nil {
		return
	}

	data, err := json.Marshal(systemsCache{FetchedAt: time.Now(), Systems: systems})
	if err != nil {
		return
	}

	// Write via a temp file and rename so concurrent readers (e.g. shell
	// completion) never observe a partially written cache.
	tmp, err := os.CreateTemp(filepath.Dir(p), ".systems-*")
	if err != nil {
		return
	}

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
		return
	}

	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmp.Name())
		return
	}

	if os.Chmod(tmp.Name(), 0o644) != nil || os.Rename(tmp.Name(), p) != nil {
		_ = os.Remove(tmp.Name())
	}
}

// Systems returns the systems list, preferring a fresh cache over the
// network. With refresh, the cache is bypassed and rewritten.
func (c *Cache) Systems(ctx context.Context, refresh bool) ([]ds.System, error) {
	if !refresh {
		if systems := c.Load(c.ttl); systems != nil {
			return systems, nil
		}
	}

	if c.edge == nil {
		return nil, errors.New("cache: no client configured to fetch systems")
	}

	systems, err := c.edge.Systems(ctx)
	if err != nil {
		return nil, err
	}

	c.store(systems)

	return systems, nil
}
