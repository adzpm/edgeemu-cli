package app

import "errors"

// ErrUsage is returned when a command is invoked without a query.
var ErrUsage = errors.New("usage: edgeemu search <query>")

// ErrUnknownFormat is returned for a -f value outside render.Formats.
var ErrUnknownFormat = errors.New("unknown format")
