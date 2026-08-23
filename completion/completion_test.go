package completion

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/adzpm/edgeemu-cli/client"
)

// TestSearchCompletionDoesNotHang ensures a TAB press never blocks on a
// slow or unreachable site: with no cache available, the network fetch
// must be abandoned after fetchTimeout.
func TestSearchCompletionDoesNotHang(t *testing.T) {
	t.Setenv("HOME", t.TempDir()) // empty cache dir

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

	edge := client.New(client.WithBaseURL(srv.URL))
	cmd := &cli.Command{Name: "edgeemu", Writer: io.Discard}

	start := time.Now()
	Search(edge)(context.Background(), cmd)

	if elapsed := time.Since(start); elapsed > fetchTimeout+2*time.Second {
		t.Fatalf("completion blocked for %v, want under %v", elapsed, fetchTimeout+2*time.Second)
	}
}
