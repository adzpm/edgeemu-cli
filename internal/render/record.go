package render

import (
	"bytes"
	"encoding/xml"

	"github.com/goccy/go-json"
	"gopkg.in/yaml.v3"

	"github.com/adzpm/edgeemu-cli/internal/ds"
)

// record is one search result reduced to the selected fields, keeping
// the canonical field order. It marshals to JSON, YAML and XML with the
// column IDs as keys, so every output format honors --columns.
type record []recordField

type recordField struct {
	id    string
	value any
}

// records builds filtered records from roms. The result is never nil,
// so an empty search encodes as an empty list.
func records(roms []ds.ROM, columnIDs []string) ([]record, error) {
	cols, err := selectColumns(columnIDs)
	if err != nil {
		return nil, err
	}

	recs := make([]record, 0, len(roms))

	for _, r := range roms {
		rec := make(record, 0, len(cols))
		for _, c := range cols {
			rec = append(rec, recordField{id: c.id, value: c.value(r)})
		}

		recs = append(recs, rec)
	}

	return recs, nil
}

func (r record) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer

	buf.WriteByte('{')

	for i, f := range r {
		if i > 0 {
			buf.WriteByte(',')
		}

		k, err := json.Marshal(f.id)
		if err != nil {
			return nil, err
		}

		v, err := json.Marshal(f.value)
		if err != nil {
			return nil, err
		}

		buf.Write(k)
		buf.WriteByte(':')
		buf.Write(v)
	}

	buf.WriteByte('}')

	return buf.Bytes(), nil
}

func (r record) MarshalYAML() (any, error) {
	node := &yaml.Node{Kind: yaml.MappingNode}

	for _, f := range r {
		var k, v yaml.Node

		err := k.Encode(f.id)
		if err != nil {
			return nil, err
		}

		err = v.Encode(f.value)
		if err != nil {
			return nil, err
		}

		node.Content = append(node.Content, &k, &v)
	}

	return node, nil
}

func (r record) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	err := e.EncodeToken(start)
	if err != nil {
		return err
	}

	for _, f := range r {
		err := e.EncodeElement(f.value, xml.StartElement{Name: xml.Name{Local: f.id}})
		if err != nil {
			return err
		}
	}

	return e.EncodeToken(start.End())
}
