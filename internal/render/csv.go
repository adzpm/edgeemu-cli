package render

import (
	"encoding/csv"
	"fmt"

	"github.com/adzpm/edgeemu-cli/internal/ds"
)

// CSVROMs renders search results as CSV: a header row with the selected
// field IDs followed by one row per ROM.
func (p *Printer) CSVROMs(roms []ds.ROM, columnIDs []string) error {
	cols, err := selectColumns(columnIDs)
	if err != nil {
		return err
	}

	w := csv.NewWriter(p.w)

	header := make([]string, 0, len(cols))
	for _, c := range cols {
		header = append(header, c.id)
	}

	if err := w.Write(header); err != nil {
		return err
	}

	for _, r := range roms {
		row := make([]string, 0, len(cols))
		for _, c := range cols {
			row = append(row, fmt.Sprint(c.value(r)))
		}

		if err := w.Write(row); err != nil {
			return err
		}
	}

	w.Flush()

	return w.Error()
}

// CSVSystems renders the systems list as CSV.
func (p *Printer) CSVSystems(systems []ds.System) error {
	w := csv.NewWriter(p.w)

	if err := w.Write([]string{"id", "name"}); err != nil {
		return err
	}

	for _, s := range systems {
		if err := w.Write([]string{s.ID, s.Name}); err != nil {
			return err
		}
	}

	w.Flush()

	return w.Error()
}
