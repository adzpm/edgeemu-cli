package app

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"golang.org/x/term"

	"github.com/adzpm/edgeemu-cli/internal/ds"
)

// barWidth is the character width of the progress bar.
const barWidth = 24

// doneMark replaces the current letter once a system is fully crawled.
const doneMark = "✓"

// progress reports per-system crawl progress on stderr. On a terminal
// the current system's line is redrawn in place and left completed; on
// plain output (CI, redirects) only one summary line per system is
// printed, so logs stay clean.
type progress struct {
	w        io.Writer
	tty      bool
	total    int // systems overall
	current  int // 1-based index of the system being crawled
	numWidth int // digits in total: [07/48] stays aligned
	idWidth  int // longest system ID: the bar column stays aligned
	letters  int // letter buckets of the current system
	done     int // finished buckets of the current system
}

func newProgress(systems []ds.System) *progress {
	idWidth := 0
	for _, s := range systems {
		if len(s.ID) > idWidth {
			idWidth = len(s.ID)
		}
	}

	return &progress{
		w:        os.Stderr,
		tty:      term.IsTerminal(int(os.Stderr.Fd())),
		total:    len(systems),
		current:  0,
		numWidth: len(strconv.Itoa(len(systems))),
		idWidth:  idWidth,
		letters:  0,
		done:     0,
	}
}

// bar renders a done/total gauge like [██████░░░░░░].
func bar(done, total int) string {
	if total <= 0 {
		total = 1
	}

	filled := min(done*barWidth/total, barWidth)

	return "[" + strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled) + "]"
}

// system starts the progress line of the next system with the given
// number of letter buckets.
func (p *progress) system(letters int) {
	p.current++
	p.letters = letters
	p.done = 0
}

// line renders one full progress line:
//
//	[09/48] colecovision [████████░░░░]  33% · h · 64 entries
func (p *progress) line(systemID, mark string, entries int) string {
	letters := max(p.letters, 1)
	percent := p.done * 100 / letters //nolint:mnd // percentage

	return fmt.Sprintf("[%0*d/%d] %-*s %s %3d%% · %s · %d entries",
		p.numWidth, p.current, p.total,
		p.idWidth, systemID,
		bar(p.done, letters), percent, mark, entries)
}

// step reports one finished letter bucket of the current system.
func (p *progress) step(systemID, letter string, entries int) {
	p.done++

	if !p.tty {
		return
	}

	fmt.Fprintf(p.w, "\r%s ", p.line(systemID, letter, entries))
}

// finish completes the current system's line. On a terminal the filled
// bar stays on screen with a check mark in place of the letter;
// otherwise a single summary line is printed.
func (p *progress) finish(systemID string, entries int) {
	if p.tty {
		p.done = p.letters
		fmt.Fprintf(p.w, "\r%s \n", p.line(systemID, doneMark, entries))

		return
	}

	fmt.Fprintf(p.w, "[%0*d/%d] %s: %d entries\n", p.numWidth, p.current, p.total, systemID, entries)
}
