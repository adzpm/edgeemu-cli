package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/adzpm/edgeemu-cli/internal/ds"
	"github.com/adzpm/edgeemu-cli/internal/render"
)

// ErrUnknownSystem is returned for a -s value the site does not list.
var ErrUnknownSystem = errors.New("unknown system")

// Dump handles the dump command: it walks the per-letter /browse pages
// of the selected systems (which, unlike the search, are not capped at
// 100 entries) and renders the full index in the requested format.
func (a *App) Dump(ctx context.Context, cmd *cli.Command) error {
	format := cmd.String(formatFlag.Name)
	if !slices.Contains(render.Formats(), format) {
		return fmt.Errorf("%w %q (available: %s)", ErrUnknownFormat, format, strings.Join(render.Formats(), ", "))
	}

	columns := cmd.StringSlice(columnsFlag.Name)
	if len(columns) == 0 {
		columns = render.ColumnIDs() // everything, explicitly
	}

	// Validate the column selection before the crawl, not after it.
	if err := render.ValidateColumns(columns); err != nil {
		return err
	}

	systems, err := a.dumpSystems(ctx, cmd.String(systemFlag.Name))
	if err != nil {
		return err
	}

	roms, err := a.crawl(ctx, systems, cmd.Duration(delayFlag.Name))
	if err != nil {
		return err
	}

	return a.printROMs(format, roms, columns)
}

// dumpSystems resolves the -s flag into the list of systems to crawl.
func (a *App) dumpSystems(ctx context.Context, selector string) ([]ds.System, error) {
	systems, err := a.cache.Systems(ctx, false)
	if err != nil {
		return nil, err
	}

	if selector == "all" {
		return systems, nil
	}

	for _, s := range systems {
		if s.ID == selector {
			return []ds.System{s}, nil
		}
	}

	return nil, fmt.Errorf("%w %q (see 'edgeemu systems')", ErrUnknownSystem, selector)
}

// crawl walks every letter bucket of every system, pausing between
// requests to stay polite to the site. Progress goes to stderr so it
// never mixes with the dumped data on stdout.
func (a *App) crawl(ctx context.Context, systems []ds.System, delay time.Duration) ([]ds.ROM, error) {
	var all []ds.ROM

	for _, sys := range systems {
		letters, err := a.edge.BrowseLetters(ctx, sys.ID)
		if err != nil {
			return nil, err
		}

		count := 0

		for _, letter := range letters {
			roms, err := a.edge.Browse(ctx, sys.ID, letter)
			if err != nil {
				return nil, err
			}

			// The browse page does not repeat the system name.
			for i := range roms {
				roms[i].System = sys.Name
			}

			all = append(all, roms...)
			count += len(roms)

			fmt.Fprintf(os.Stderr, "\r%-35s %s (%d entries, %d total)   ", sys.ID, letter, count, len(all))

			if err := pause(ctx, delay); err != nil {
				return nil, err
			}
		}

		fmt.Fprintln(os.Stderr)
	}

	return all, nil
}

// pause sleeps for delay, aborting early when ctx is cancelled.
func pause(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(delay):
		return nil
	}
}
