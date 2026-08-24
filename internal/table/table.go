package table

import (
	"io"
	"os"

	"github.com/mattn/go-runewidth"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"golang.org/x/term"
)

// minColWidth is the narrowest a column may be squeezed to,
// including cell padding: enough for a few characters plus an ellipsis.
const minColWidth = 6

// Printer renders bordered tables.
type Printer struct {
	w io.Writer
}

// Option customizes a Printer.
type Option func(*Printer)

// WithWriter overrides the output destination (os.Stdout by default).
func WithWriter(w io.Writer) Option {
	return func(p *Printer) { p.w = w }
}

// New creates a Printer with sane defaults, applying the given options.
func New(opts ...Option) *Printer {
	p := &Printer{w: os.Stdout}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// termWidth returns the terminal width of the printer's output, or 0
// when the output is not a terminal.
func (p *Printer) termWidth() int {
	f, ok := p.w.(*os.File)
	if !ok {
		return 0
	}

	w, _, err := term.GetSize(int(f.Fd()))
	if err != nil || w <= 0 {
		return 0
	}

	return w
}

// fitWidths returns per-column widths so the table fits the terminal,
// shrinking the widest columns first. Returns nil when the output is not
// a terminal or the table already fits, meaning no constraints are needed.
func (p *Printer) fitWidths(headers []string, rows [][]string) tw.Mapper[int, int] {
	termW := p.termWidth()
	if termW <= 0 {
		return nil
	}

	return fitWidthsTo(headers, rows, termW)
}

// fitWidthsTo squeezes column widths into termW; see fitWidths.
func fitWidthsTo(headers []string, rows [][]string, termW int) tw.Mapper[int, int] {
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = runewidth.StringWidth(h) + 2
	}
	for _, row := range rows {
		for i, c := range row {
			if i >= len(widths) {
				break
			}
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

// Render prints rows under headers as a bordered table,
// squeezed to the terminal width when necessary.
func (p *Printer) Render(headers []string, rows [][]string) error {
	opts := []tablewriter.Option{tablewriter.WithHeaderAutoFormat(tw.Off)}
	if widths := p.fitWidths(headers, rows); widths != nil {
		opts = append(opts,
			tablewriter.WithColumnWidths(widths),
			tablewriter.WithHeaderAutoWrap(tw.WrapTruncate),
			tablewriter.WithRowAutoWrap(tw.WrapTruncate),
		)
	}

	table := tablewriter.NewTable(p.w, opts...)

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
