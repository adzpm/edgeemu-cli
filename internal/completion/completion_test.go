package completion

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v3"

	"github.com/adzpm/edgeemu-cli/internal/cache"
	"github.com/adzpm/edgeemu-cli/internal/client"
)

// TestSearchCompletionDoesNotHang ensures a TAB press never blocks on a
// slow or unreachable site: with no cache available, the network fetch
// must be abandoned after fetchTimeout.
func TestSearchCompletionDoesNotHang(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select { // never answers in time, but frees up once the client gives up
		case <-r.Context().Done():
		case <-time.After(30 * time.Second):
		}
	}))
	defer srv.Close()

	origArgs := os.Args
	os.Args = []string{"edgeemu", "search", "-s", "--generate-shell-completion"}
	defer func() { os.Args = origArgs }()

	c := cache.New(
		cache.WithClient(client.New(client.WithBaseURL(srv.URL))),
		cache.WithPath(cachePath(t)), // empty cache: forces the network path
	)
	cmd := &cli.Command{Name: "edgeemu", Writer: io.Discard}

	start := time.Now()
	New(WithCache(c)).Search(context.Background(), cmd)

	assert.Less(t, time.Since(start), fetchTimeout+2*time.Second, "completion blocked the shell")
}
