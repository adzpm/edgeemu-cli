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
		return fmt.Errorf("usage: edgeemu search <query>")
	}

	format := cmd.String(formatFlag.Name)

	roms, err := a.edge.Search(ctx, query, cmd.String(systemFlag.Name))
	if err != nil {
		return err
	}

	if limit := cmd.Int(limitFlag.Name); limit > 0 && len(roms) > limit {
		roms = roms[:limit]
	}

	switch format {
	case "json", "yaml", "xml":
		if roms == nil {
			roms = []ds.ROM{} // encode as an empty list, not null
		}

		switch format {
		case "json":
			return a.printer.JSON(roms)
		case "yaml":
			return a.printer.YAML(roms)
		default:
			return a.printer.XMLROMs(roms)
		}
	case "list":
		if len(roms) == 0 {
			fmt.Fprintln(a.printer.Writer(), "nothing found")
			return nil
		}

		columns := cmd.StringSlice(columnsFlag.Name)
		if len(columns) == 0 {
			columns = render.ColumnIDs() // everything, explicitly
		}

		return a.printer.PrintROMs(roms, columns)
	default:
		return fmt.Errorf("unknown format %q (available: list, json, yaml, xml)", format)
	}
}
