package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// searchPage mimics the real edgeemu.net search results markup: one item
// with HTML entities, one plain, both wrapped in the surrounding page.
const searchPage = `<!DOCTYPE html>
<html><body>
<h1>search results for sonic</h1>
<div class="grid">
  <div class="item">
    <details data-name="Sonic &amp; Knuckles (World).zip">
      <summary>Sonic &amp; Knuckles (World)</summary>
      <p><a href="/download/sega-genesis/Sonic%20%26%20Knuckles%20%28World%29.zip">download</a> (<span>1.36m, 341 DLs</span>)</p>
      <p>system: <span>Sega Mega Drive / Genesis</span></p>
      <p>unpacked size: <span>256.00k</span></p>
      <p>hash: <span>4DCFD55C 0658F691</span></p>
    </details>
  </div>
  <div class="item">
    <details data-name="Sonic The Hedgehog (USA, Europe).zip">
      <summary>Sonic The Hedgehog (USA, Europe)</summary>
      <p><a href="/download/sega-genesis/Sonic%20The%20Hedgehog%20%28USA%2C%20Europe%29.zip">download</a> (<span>377.87k, 588 DLs</span>)</p>
      <p>system: <span>Sega Mega Drive / Genesis</span></p>
      <p>unpacked size: <span>512.00k</span></p>
      <p>hash: <span></span></p>
    </details>
  </div>
</div>
</body></html>`

const systemsPage = `<!DOCTYPE html>
<html><body>
<select name="system" class="dropdown">
  <option value="all" selected>Search all, or select from the list</option>
  <option value="atari-2600">Atari 2600</option>
  <option value="sega-genesis">Sega Mega Drive / Genesis</option>
  <option value="microsoft-msx">Microsoft MSX / MSX2</option>
</select>
</body></html>`

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
		w.Write([]byte(searchPage))
	})

	roms, err := c.Search(context.Background(), "sonic & co", "sega-genesis")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("method = %s, want POST", gotMethod)
	}
	if gotQuery != "sonic & co" {
		t.Errorf("search form value = %q, want %q", gotQuery, "sonic & co")
	}
	if gotSystem != "sega-genesis" {
		t.Errorf("system form value = %q, want %q", gotSystem, "sega-genesis")
	}

	if len(roms) != 2 {
		t.Fatalf("got %d roms, want 2", len(roms))
	}

	first := roms[0]
	if first.Name != "Sonic & Knuckles (World)" {
		t.Errorf("Name = %q, want unescaped ampersand", first.Name)
	}
	if want := c.baseURL + "/download/sega-genesis/Sonic%20%26%20Knuckles%20%28World%29.zip"; first.URL != want {
		t.Errorf("URL = %q, want %q", first.URL, want)
	}
	if first.Size != "1.36m" {
		t.Errorf("Size = %q, want 1.36m", first.Size)
	}
	if first.Downloads != 341 {
		t.Errorf("Downloads = %d, want 341", first.Downloads)
	}
	if first.System != "Sega Mega Drive / Genesis" {
		t.Errorf("System = %q", first.System)
	}
	if first.UnpackedSize != "256.00k" {
		t.Errorf("UnpackedSize = %q, want 256.00k", first.UnpackedSize)
	}
	if first.Hash != "4DCFD55C 0658F691" {
		t.Errorf("Hash = %q", first.Hash)
	}

	if roms[1].Hash != "" {
		t.Errorf("empty hash parsed as %q, want empty", roms[1].Hash)
	}
}

func TestSearchNoResults(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html><body><h1>search results for zzz</h1><div class="grid"></div></body></html>`))
	})

	roms, err := c.Search(context.Background(), "zzz", "all")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(roms) != 0 {
		t.Fatalf("got %d roms, want 0", len(roms))
	}
}

func TestSearchHTTPError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	})

	if _, err := c.Search(context.Background(), "sonic", "all"); err == nil {
		t.Fatal("want error on HTTP 500, got nil")
	}
}

func TestSearchContextCancelled(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	if _, err := c.Search(ctx, "sonic", "all"); err == nil {
		t.Fatal("want error on cancelled context, got nil")
	}
}

func TestSystems(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		w.Write([]byte(systemsPage))
	})

	systems, err := c.Systems(context.Background())
	if err != nil {
		t.Fatalf("Systems: %v", err)
	}

	if len(systems) != 3 {
		t.Fatalf("got %d systems, want 3 (the 'all' option must be skipped)", len(systems))
	}
	if systems[0].ID != "atari-2600" || systems[0].Name != "Atari 2600" {
		t.Errorf("first system = %+v", systems[0])
	}
	if systems[1].ID != "sega-genesis" {
		t.Errorf("second system = %+v", systems[1])
	}
}

func TestSystemsEmptyPageIsError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html><body>layout changed</body></html>`))
	})

	if _, err := c.Systems(context.Background()); err == nil {
		t.Fatal("want error when no systems parsed, got nil")
	}
}

func TestSystemsHTTPError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusBadGateway)
	})

	if _, err := c.Systems(context.Background()); err == nil {
		t.Fatal("want error on HTTP 502, got nil")
	}
}

func TestWithBaseURLTrimsTrailingSlash(t *testing.T) {
	c := New(WithBaseURL("https://example.com///"))
	if c.baseURL != "https://example.com" {
		t.Errorf("baseURL = %q, want trailing slashes trimmed", c.baseURL)
	}
}

func TestWithHTTPClientNilIgnored(t *testing.T) {
	c := New(WithHTTPClient(nil))
	if c.http == nil {
		t.Fatal("nil http client applied, default lost")
	}
}
