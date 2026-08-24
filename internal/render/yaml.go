package render

import (
	"gopkg.in/yaml.v3"
)

// YAML renders v as YAML.
func (p *Printer) YAML(v any) error {
	enc := yaml.NewEncoder(p.w)

	err := enc.Encode(v)
	if err != nil {
		_ = enc.Close()

		return err
	}

	return enc.Close()
}
