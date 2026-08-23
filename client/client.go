package client

import (
	"net/http"
	"strings"
	"time"
)

const defaultBaseURL = "https://edgeemu.net"

// maxResponseSize caps how much of a response body is read; search pages
// are a few hundred KB at most, so 20 MiB means something is very wrong.
const maxResponseSize = 20 << 20

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
		http:    &http.Client{Timeout: 5 * time.Minute},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}
