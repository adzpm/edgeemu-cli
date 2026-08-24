package render

import (
	"encoding/xml"
	"fmt"

	"github.com/adzpm/edgeemu-cli/internal/ds"
)

// XML list wrappers: encoding/xml cannot marshal a bare slice into a
// valid document, so each list gets a root element here.
type xmlROMs struct {
	XMLName xml.Name `xml:"roms"`
	ROMs    []ds.ROM `xml:"rom"`
}

type xmlSystems struct {
	XMLName xml.Name    `xml:"systems"`
	Systems []ds.System `xml:"system"`
}

// XMLROMs renders search results as an XML document rooted at <roms>.
func (p *Printer) XMLROMs(roms []ds.ROM) error {
	return p.xml(xmlROMs{ROMs: roms})
}

// XMLSystems renders the systems list as an XML document rooted at <systems>.
func (p *Printer) XMLSystems(systems []ds.System) error {
	return p.xml(xmlSystems{Systems: systems})
}

func (p *Printer) xml(v any) error {
	out, err := xml.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(p.w, "%s%s\n", xml.Header, out)
	return err
}
