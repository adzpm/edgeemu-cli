package cache

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/adzpm/edgeemu-cli/client"
	"github.com/adzpm/edgeemu-cli/ds"
)

// TTL is how long a cached systems list is considered fresh.
const TTL = 24 * time.Hour

type systemsCache struct {
	FetchedAt time.Time   `json:"fetched_at"`
	Systems   []ds.System `json:"systems"`
}

func path() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "edgeemu", "systems.json"), nil
}

// Load returns the cached systems list, or nil if there is no usable
// cache. A maxAge of 0 accepts a cache of any age.
func Load(maxAge time.Duration) []ds.System {
	p, err := path()
	if err != nil {
		return nil
	}

	data, err := os.ReadFile(p)
	if err != nil {
		return nil
	}

	var c systemsCache
	if json.Unmarshal(data, &c) != nil {
		return nil
	}

	if maxAge > 0 && time.Since(c.FetchedAt) > maxAge {
		return nil
	}

	return c.Systems
}

func store(systems []ds.System) {
	p, err := path()
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

	_ = os.WriteFile(p, data, 0o644)
}

// Systems returns the systems list, preferring a fresh cache over the
// network. With refresh, the cache is bypassed and rewritten.
func Systems(ctx context.Context, edge *client.Client, refresh bool) ([]ds.System, error) {
	if !refresh {
		if systems := Load(TTL); systems != nil {
			return systems, nil
		}
	}

	systems, err := edge.Systems(ctx)
	if err != nil {
		return nil, err
	}

	store(systems)

	return systems, nil
}
