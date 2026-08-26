package render

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/adzpm/edgeemu-cli/internal/ds"
)

// numPad is the width of ". " after the index number.
const numPad = 2

// listLayout is the precomputed shape of the list output: which special
// fields are selected and how entries are indented.
type listLayout struct {
	hasName bool
	hasURL  bool
	fields  []romColumn
	numW    int
	indent  string
}

func newListLayout(cols []romColumn, total int) listLayout {
	l := listLayout{numW: len(strconv.Itoa(total))} //nolint:exhaustruct // remaining fields are filled below
	l.indent = strings.Repeat(" ", l.numW+numPad)

	for _, c := range cols {
		switch c.id {
		case colName:
			l.hasName = true
		case colURL:
			l.hasURL = true
		default:
			l.fields = append(l.fields, c)
		}
	}

	return l
}

// PrintROMs renders search results as a readable list: the name as a
// numbered heading, secondary fields on one line, and the URL alone on
// its own line. Nothing is ever truncated, and the URL never shares a
// line with other text, so terminals always keep it clickable.
func (p *Printer) PrintROMs(roms []ds.ROM, columnIDs []string) error {
	cols, err := selectColumns(columnIDs)
	if err != nil {
		return err
	}

	layout := newListLayout(cols, len(roms))

	for i, r := range roms {
		if err := p.printROMEntry(layout, i, r); err != nil {
			return err
		}
	}

	return nil
}

// printROMEntry renders one numbered list entry.
func (p *Printer) printROMEntry(l listLayout, i int, r ds.ROM) error {
	head := ""
	if l.hasName {
		head = r.Name
	}

	if _, err := fmt.Fprintf(p.w, "%0*d. %s\n", l.numW, i+1, head); err != nil {
		return err
	}

	if len(l.fields) > 0 {
		parts := make([]string, 0, len(l.fields))
		for _, c := range l.fields {
			parts = append(parts, fmt.Sprintf("%s: %v", c.id, c.value(r)))
		}

		if _, err := fmt.Fprintf(p.w, "%s%s\n", l.indent, strings.Join(parts, " · ")); err != nil {
			return err
		}
	}

	if l.hasURL {
		if _, err := fmt.Fprintf(p.w, "%s%s\n", l.indent, r.URL); err != nil {
			return err
		}
	}

	return nil
}

// PrintSystems renders the systems list as numbered, aligned lines:
//
//  01. Sega 32X    · id: sega-32x
func (p *Printer) PrintSystems(systems []ds.System) error {
	numW := len(strconv.Itoa(len(systems)))

	nameW := 0
	for _, s := range systems {
		if len(s.Name) > nameW {
			nameW = len(s.Name)
		}
	}

	for i, s := range systems {
		if _, err := fmt.Fprintf(p.w, "%0*d. %-*s · id: %s\n", numW, i+1, nameW, s.Name, s.ID); err != nil {
			return err
		}
	}

	return nil
}
