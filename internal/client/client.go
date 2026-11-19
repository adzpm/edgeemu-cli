// Package client implements the HTTP client for edgeemu.net.
package client

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const defaultBaseURL = "https://edgeemu.net"

// defaultTimeout bounds every request of the default HTTP client.
const defaultTimeout = 5 * time.Minute

// maxResponseSize caps how much of a response body is read; search pages
// are a few hundred KB at most, so 20 MiB means something is very wrong.
const maxResponseSize = 20 << 20

// entryMarker opens every ROM entry on the search and browse pages.
const entryMarker = "<details data-name="

// checkParsed guards against a silent layout change: a page that lists
// ROM entries must yield at least one parsed entry.
func checkParsed(body []byte, parsed int, page string) error {
	if listed := bytes.Count(body, []byte(entryMarker)); parsed == 0 && listed > 0 {
		return fmt.Errorf("%w (0 of %d on %s)", ErrNoEntries, listed, page)
	}

	return nil
}

// Client talks to edgeemu.net.
type Client struct {
	baseURL string
	http    *http.Client
}

// Option customizes a Client.
type Option func(*Client)

// WithBaseURL overrides the edgeemu.net base URL.
func WithBaseURL(u string) Option {
	return func(c *Client) { c.baseURL = strings.TrimRight(u, "/") }
}

// WithHTTPClient overrides the underlying HTTP client. A nil client is ignored.
func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) {
		if h != nil {
			c.http = h
		}
	}
}

// New creates a Client with sane defaults, applying the given options.
func New(opts ...Option) *Client {
	c := &Client{
		baseURL: defaultBaseURL,
		http:    &http.Client{Timeout: defaultTimeout},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}
