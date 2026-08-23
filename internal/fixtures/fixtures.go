// Package fixtures holds HTML pages mirroring the real edgeemu.net markup,
// shared by tests across packages. If the site layout changes, refresh
// these files (e.g. curl the live pages) and the parser expectations.
package fixtures

import _ "embed"

// SearchPage is a search results page with two items: one containing
// HTML entities and a multi-word hash, one with an empty hash.
//
//go:embed search_results.html
var SearchPage string

// EmptyPage is a search results page with no items.
//
//go:embed empty_results.html
var EmptyPage string

// SystemsPage is a page with the system dropdown: the "all" option
// (which parsers must skip) plus three systems.
//
//go:embed systems.html
var SystemsPage string
