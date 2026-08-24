package render

import (
	"errors"
	"fmt"
	"strings"

	"github.com/adzpm/edgeemu-cli/internal/ds"
)

// Selectable field IDs; also the keys/headers of structured outputs.
const (
	colName     = "name"
	colSystem   = "system"
	colSize     = "size"
	colUnpacked = "unpacked"
	colDLs      = "dls"
	colHash     = "hash"
	colURL      = "url"
)

// ErrUnknownColumn is returned for a -c value outside ColumnIDs.
var ErrUnknownColumn = errors.New("unknown column")

// romColumn is a selectable field of the search results output.
type romColumn struct {
	id    string
	value func(ds.ROM) any
}

// romColumns lists all selectable fields in their display order.
var romColumns = []romColumn{
	{colName, func(r ds.ROM) any { return r.Name }},
	{colSystem, func(r ds.ROM) any { return r.System }},
	{colSize, func(r ds.ROM) any { return r.Size }},
	{colUnpacked, func(r ds.ROM) any { return r.UnpackedSize }},
	{colDLs, func(r ds.ROM) any { return r.Downloads }},
	{colHash, func(r ds.ROM) any { return r.Hash }},
	{colURL, func(r ds.ROM) any { return r.URL }},
}

// ColumnIDs returns the IDs of all selectable fields in display order.
func ColumnIDs() []string {
	ids := make([]string, 0, len(romColumns))
	for _, c := range romColumns {
		ids = append(ids, c.id)
	}

	return ids
}

// selectColumns resolves requested field IDs, keeping the canonical
// display order regardless of the order requested.
func selectColumns(ids []string) ([]romColumn, error) {
	want := make(map[string]bool, len(ids))

	for _, id := range ids {
		id = strings.ToLower(strings.TrimSpace(id))
		found := false

		for _, c := range romColumns {
			if c.id == id {
				found = true

				break
			}
		}

		if !found {
			return nil, fmt.Errorf("%w %q (available: %s)", ErrUnknownColumn, id, strings.Join(ColumnIDs(), ", "))
		}

		want[id] = true
	}

	var cols []romColumn

	for _, c := range romColumns {
		if want[c.id] {
			cols = append(cols, c)
		}
	}

	return cols, nil
}
