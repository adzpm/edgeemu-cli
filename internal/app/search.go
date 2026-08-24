package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/adzpm/edgeemu-cli/internal/ds"
	"github.com/adzpm/edgeemu-cli/internal/render"
)

// Search handles the search command.
func (a *App) Search(ctx context.Context, cmd *cli.Command) error {
	query := strings.TrimSpace(strings.Join(cmd.Args().Slice(), " "))
	if query == "" {
		return ErrUsage
	}

	roms, err := a.edge.Search(ctx, query, cmd.String(systemFlag.Name))
	if err != nil {
		return err
	}

	if limit := cmd.Int(limitFlag.Name); limit > 0 && len(roms) > limit {
		roms = roms[:limit]
	}

	columns := cmd.StringSlice(columnsFlag.Name)
	if len(columns) == 0 {
		columns = render.ColumnIDs() // everything, explicitly
	}

	return a.printROMs(cmd.String(formatFlag.Name), roms, columns)
}

// printROMs renders search results in the requested format.
func (a *App) printROMs(format string, roms []ds.ROM, columns []string) error {
	switch format {
	case render.FormatJSON:
		return a.printer.JSONROMs(roms, columns)
	case render.FormatYAML:
		return a.printer.YAMLROMs(roms, columns)
	case render.FormatXML:
		return a.printer.XMLROMs(roms, columns)
	case render.FormatCSV:
		return a.printer.CSVROMs(roms, columns)
	case render.FormatList:
		if len(roms) == 0 {
			fmt.Fprintln(a.printer.Writer(), "nothing found")

			return nil
		}

		return a.printer.PrintROMs(roms, columns)
	default:
		return fmt.Errorf("%w %q (available: %s)", ErrUnknownFormat, format, strings.Join(render.Formats(), ", "))
	}
}
