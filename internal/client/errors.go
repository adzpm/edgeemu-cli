package client

import "errors"

// ErrRequestFailed is returned when edgeemu.net answers with a non-200 status.
var ErrRequestFailed = errors.New("request failed")

// ErrNoSystems is returned when the systems dropdown cannot be parsed,
// which most likely means the page layout has changed.
var ErrNoSystems = errors.New("no systems parsed: page layout may have changed")

// ErrNoEntries is returned when a page lists ROM entries but none of
// them could be parsed, which most likely means the page layout has
// changed. Without it such a page would pass for an empty result.
var ErrNoEntries = errors.New("no entries parsed: page layout may have changed")
