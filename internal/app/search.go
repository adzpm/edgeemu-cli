package app

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/adzpm/edgeemu-cli/internal/ds"
)

// Search handles the search command.
func (a *App) Search(ctx context.Context, cmd *cli.Command) error {
	query := strings.TrimSpace(strings.Join(cmd.Args().Slice(), " "))
	if query == "" {
		return fmt.Errorf("usage: edgeemu search <query>")
	}

	roms, err := a.edge.Search(ctx, query, cmd.String(systemFlag.Name))
	if err != nil {
		return err
	}

	if limit := cmd.Int(limitFlag.Name); limit > 0 && len(roms) > limit {
		roms = roms[:limit]
	}

	if cmd.Bool(jsonFlag.Name) {
		if roms == nil {
			roms = []ds.ROM{} // encode as [], not null
		}

		enc := json.NewEncoder(cmd.Root().Writer)
		enc.SetIndent("", "  ")
		return enc.Encode(roms)
	}

	if len(roms) == 0 {
		fmt.Fprintln(cmd.Root().Writer, "nothing found")
		return nil
	}

	columns := cmd.StringSlice(columnsFlag.Name)
	if len(columns) == 0 {
		columns = []string{"name", "url"}
		if cmd.String(systemFlag.Name) == "all" {
			columns = append(columns, "system")
		}
	}

	return a.table.PrintROMs(roms, columns)
}
