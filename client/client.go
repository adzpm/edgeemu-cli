package client

import (
	"net/http"
	"time"
)

const defaultBaseURL = "https://edgeemu.net"

// Client talks to edgeemu.net.
type Client struct {
	baseURL string
	http    *http.Client
}

// Option customizes a Client.
type Option func(*Client)

// WithBaseURL overrides the edgeemu.net base URL.
func WithBaseURL(u string) Option {
	return func(c *Client) { c.baseURL = u }
}

// WithHTTPClient overrides the underlying HTTP client.
func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) { c.http = h }
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
