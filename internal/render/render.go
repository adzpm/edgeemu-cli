// Package render prints search results and system lists in all
// supported output formats.
package render

import (
	"io"
	"os"
)

// Printer renders search results and system lists as plain text.
type Printer struct {
	w io.Writer
}

// Option customizes a Printer.
type Option func(*Printer)

// WithWriter overrides the output destination (os.Stdout by default).
func WithWriter(w io.Writer) Option {
	return func(p *Printer) { p.w = w }
}

// New creates a Printer with sane defaults, applying the given options.
func New(opts ...Option) *Printer {
	p := &Printer{w: os.Stdout}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// Writer exposes the printer's output destination for one-off messages.
func (p *Printer) Writer() io.Writer {
	return p.w
}
