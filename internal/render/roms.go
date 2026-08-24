package render

import (
	"github.com/adzpm/edgeemu-cli/internal/ds"
)

// JSONROMs renders search results as JSON with only the selected fields.
func (p *Printer) JSONROMs(roms []ds.ROM, columnIDs []string) error {
	recs, err := records(roms, columnIDs)
	if err != nil {
		return err
	}

	return p.JSON(recs)
}

// YAMLROMs renders search results as YAML with only the selected fields.
func (p *Printer) YAMLROMs(roms []ds.ROM, columnIDs []string) error {
	recs, err := records(roms, columnIDs)
	if err != nil {
		return err
	}

	return p.YAML(recs)
}

// XMLROMs renders search results as an XML document rooted at <roms>,
// with only the selected fields.
func (p *Printer) XMLROMs(roms []ds.ROM, columnIDs []string) error {
	recs, err := records(roms, columnIDs)
	if err != nil {
		return err
	}

	return p.xml(xmlROMs{ROMs: recs})
}
