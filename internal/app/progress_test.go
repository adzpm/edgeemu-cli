package app

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/adzpm/edgeemu-cli/internal/ds"
)

func TestBar(t *testing.T) {
	assert.Equal(t, "["+strings.Repeat("░", barWidth)+"]", bar(0, 27))
	assert.Equal(t, "["+strings.Repeat("█", barWidth)+"]", bar(27, 27))
	assert.Equal(t, "["+strings.Repeat("█", barWidth/2)+strings.Repeat("░", barWidth/2)+"]", bar(1, 2))
	assert.NotPanics(t, func() { bar(1, 0) }, "zero total must not divide by zero")
	assert.Equal(t, "["+strings.Repeat("█", barWidth)+"]", bar(5, 3), "overshoot must clamp to a full bar")
}

func TestNewProgressWidths(t *testing.T) {
	systems := make([]ds.System, 0, 12)
	for range 12 {
		systems = append(systems, ds.System{ID: "atari-2600", Name: "Atari 2600"})
	}

	p := newProgress(systems)

	assert.Equal(t, 2, p.numWidth, "index width must fit the total")
	assert.Equal(t, len("atari-2600"), p.idWidth, "ID column must fit the longest crawled system only")
}

func TestProgressPlainOutput(t *testing.T) {
	var buf bytes.Buffer

	p := &progress{w: &buf, tty: false, total: 12, current: 0, numWidth: 2, idWidth: 12}

	p.system(27)
	p.step("sega-genesis", "a", 100) // must be silent without a terminal
	p.finish("sega-genesis", 700)

	p.system(27)
	p.finish("atari-2600", 500)

	assert.Equal(t, "[01/12] sega-genesis: 700 entries\n[02/12] atari-2600: 500 entries\n", buf.String(),
		"indexes must carry leading zeros")
}

func TestProgressTTYOutput(t *testing.T) {
	var buf bytes.Buffer

	p := &progress{w: &buf, tty: true, total: 12, current: 8, numWidth: 2, idWidth: 12}

	p.system(2)
	p.step("sega-genesis", "b", 100)

	out := buf.String()
	assert.Contains(t, out, "\r[09/12] sega-genesis", "the index must be zero-padded to the total's width")
	assert.Contains(t, out, bar(1, 2)+"  50% · b · 100 entries", "the current letter and percent must be shown")

	p.finish("sega-genesis", 700)

	out = buf.String()
	assert.Contains(t, out, bar(2, 2)+" 100% · "+doneMark+" · 700 entries",
		"a finished system must show a full bar and a check mark instead of the letter")
	assert.True(t, strings.HasSuffix(out, "\n"), "the completed bar must stay as a finished line")
}
