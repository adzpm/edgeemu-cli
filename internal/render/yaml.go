package render

import (
	"gopkg.in/yaml.v3"
)

// YAML renders v as YAML.
func (p *Printer) YAML(v any) error {
	enc := yaml.NewEncoder(p.w)
	if err := enc.Encode(v); err != nil {
		_ = enc.Close()
		return err
	}

	return enc.Close()
}
