package table

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/adzpm/edgeemu-cli/internal/ds"
)

// romColumn is a selectable column of the search results table.
type romColumn struct {
	id     string
	header string
	value  func(ds.ROM) string
}

// romColumns lists all selectable columns in their display order.
var romColumns = []romColumn{
	{"name", "Name", func(r ds.ROM) string { return r.Name }},
	{"system", "System", func(r ds.ROM) string { return r.System }},
	{"size", "Size", func(r ds.ROM) string { return r.Size }},
	{"unpacked", "Unpacked", func(r ds.ROM) string { return r.UnpackedSize }},
	{"dls", "DLs", func(r ds.ROM) string { return strconv.Itoa(r.Downloads) }},
	{"hash", "Hash", func(r ds.ROM) string { return r.Hash }},
	{"url", "URL", func(r ds.ROM) string { return r.URL }},
}

// ColumnIDs returns the IDs of all selectable columns in display order.
func ColumnIDs() []string {
	ids := make([]string, 0, len(romColumns))
	for _, c := range romColumns {
		ids = append(ids, c.id)
	}

	return ids
}

// selectColumns resolves requested column IDs into columns, keeping
// the canonical display order regardless of the order requested.
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
			return nil, fmt.Errorf("unknown column %q (available: %s)", id, strings.Join(ColumnIDs(), ", "))
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

// PrintROMs renders search results with the requested columns.
func (p *Printer) PrintROMs(roms []ds.ROM, columnIDs []string) error {
	cols, err := selectColumns(columnIDs)
	if err != nil {
		return err
	}

	headers := make([]string, 0, len(cols)+1)
	headers = append(headers, "#")
	for _, c := range cols {
		headers = append(headers, c.header)
	}

	rows := make([][]string, 0, len(roms))
	for i, r := range roms {
		row := make([]string, 0, len(cols)+1)
		row = append(row, strconv.Itoa(i+1))
		for _, c := range cols {
			row = append(row, c.value(r))
		}

		rows = append(rows, row)
	}

	return p.Render(headers, rows)
}

// PrintSystems renders the systems list.
func (p *Printer) PrintSystems(systems []ds.System) error {
	rows := make([][]string, 0, len(systems))
	for _, s := range systems {
		rows = append(rows, []string{s.ID, s.Name})
	}

	return p.Render([]string{"ID", "Name"}, rows)
}
