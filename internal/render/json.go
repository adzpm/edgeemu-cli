package render

import (
	"fmt"

	"github.com/goccy/go-json"
)

// JSON renders v as indented JSON via goccy/go-json, the fastest pure-Go
// JSON codec (bytedance/sonic is faster still, but as of v1.15 it does
// not support go1.27 and silently falls back to encoding/json with a
// warning on stderr).
func (p *Printer) JSON(v any) error {
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(p.w, "%s\n", out)

	return err
}
