package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"golang.org/x/term"

	"github.com/adzpm/edgeemu-cli/ds"
)

// minColWidth is the narrowest a column may be squeezed to,
// including cell padding: enough for a few characters plus an ellipsis.
const minColWidth = 6

// fitWidths returns per-column widths so the table fits the terminal,
// shrinking the widest columns first. Returns nil when stdout is not a
// terminal or the table already fits, meaning no constraints are needed.
func fitWidths(headers []string, rows [][]string) tw.Mapper[int, int] {
	termW, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || termW <= 0 {
		return nil
	}

	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = runewidth.StringWidth(h) + 2
	}
	for _, row := range rows {
		for i, c := range row {
			if w := runewidth.StringWidth(c) + 2; w > widths[i] {
				widths[i] = w
			}
		}
	}

	total := len(headers) + 1 // vertical borders
	for _, w := range widths {
		total += w
	}
	if total <= termW {
		return nil
	}

	for total > termW {
		widest := 0
		for i, w := range widths {
			if w > widths[widest] {
				widest = i
			}
		}
		if widths[widest] <= minColWidth {
			break
		}
		widths[widest]--
		total--
	}

	m := tw.Mapper[int, int]{}
	for i, w := range widths {
		m[i] = w
	}
	return m
}

func renderTable(headers []string, rows [][]string) error {
	opts := []tablewriter.Option{tablewriter.WithHeaderAutoFormat(tw.Off)}
	if widths := fitWidths(headers, rows); widths != nil {
		opts = append(opts,
			tablewriter.WithColumnWidths(widths),
			tablewriter.WithHeaderAutoWrap(tw.WrapTruncate),
			tablewriter.WithRowAutoWrap(tw.WrapTruncate),
		)
	}

	table := tablewriter.NewTable(os.Stdout, opts...)

	hs := make([]any, len(headers))
	for i, h := range headers {
		hs[i] = h
	}
	table.Header(hs...)

	for _, row := range rows {
		if err := table.Append(row); err != nil {
			return err
		}
	}

	return table.Render()
}

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

func columnIDs() []string {
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
			return nil, fmt.Errorf("unknown column %q (available: %s)", id, strings.Join(columnIDs(), ", "))
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

func printROMs(roms []ds.ROM, columnIDs []string) error {
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

	return renderTable(headers, rows)
}

func printSystems(systems []ds.System) error {
	rows := make([][]string, 0, len(systems))
	for _, s := range systems {
		rows = append(rows, []string{s.ID, s.Name})
	}
	return renderTable([]string{"ID", "Name"}, rows)
}
