package render

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/adzpm/edgeemu-cli/internal/ds"
)

// PrintROMs renders search results as a readable list: the name as a
// numbered heading, secondary fields on one line, and the URL alone on
// its own line. Nothing is ever truncated, and the URL never shares a
// line with other text, so terminals always keep it clickable.
func (p *Printer) PrintROMs(roms []ds.ROM, columnIDs []string) error {
	cols, err := selectColumns(columnIDs)
	if err != nil {
		return err
	}

	var hasName, hasURL bool
	var fields []romColumn

	for _, c := range cols {
		switch c.id {
		case "name":
			hasName = true
		case "url":
			hasURL = true
		default:
			fields = append(fields, c)
		}
	}

	numW := len(strconv.Itoa(len(roms)))
	indent := strings.Repeat(" ", numW+2)

	for i, r := range roms {
		if i > 0 {
			if _, err := fmt.Fprintln(p.w); err != nil {
				return err
			}
		}

		head := ""
		if hasName {
			head = r.Name
		}
		if _, err := fmt.Fprintf(p.w, "%*d. %s\n", numW, i+1, head); err != nil {
			return err
		}

		if len(fields) > 0 {
			parts := make([]string, 0, len(fields))
			for _, c := range fields {
				parts = append(parts, c.id+": "+c.value(r))
			}
			if _, err := fmt.Fprintf(p.w, "%s%s\n", indent, strings.Join(parts, " · ")); err != nil {
				return err
			}
		}

		if hasURL {
			if _, err := fmt.Fprintf(p.w, "%s%s\n", indent, r.URL); err != nil {
				return err
			}
		}
	}

	return nil
}

// PrintSystems renders the systems list as aligned "id  name" lines.
func (p *Printer) PrintSystems(systems []ds.System) error {
	idW := 0
	for _, s := range systems {
		if len(s.ID) > idW {
			idW = len(s.ID)
		}
	}

	for _, s := range systems {
		if _, err := fmt.Fprintf(p.w, "%-*s  %s\n", idW, s.ID, s.Name); err != nil {
			return err
		}
	}

	return nil
}
