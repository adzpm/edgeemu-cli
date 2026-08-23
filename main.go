package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/adzpm/edgeemu-cli/client"
)

var (
	systemFlag = &cli.StringFlag{
		Name:    "system",
		Aliases: []string{"s"},
		Value:   "all",
		Usage:   "system to search in (see 'edgeemu systems')",
	}

	jsonFlag = &cli.BoolFlag{
		Name:  "json",
		Usage: "output results as JSON",
	}

	limitFlag = &cli.IntFlag{
		Name:    "limit",
		Aliases: []string{"l"},
		Usage:   "max results to show (0 = all)",
	}

	columnsFlag = &cli.StringSliceFlag{
		Name:    "columns",
		Aliases: []string{"c"},
		Usage:   "columns to show, comma-separated (" + strings.Join(columnIDs(), ", ") + ")",
	}
)

func main() {
	edge := client.New()

	cmd := &cli.Command{
		Name:  "edgeemu",
		Usage: "search ROMs on edgeemu.net",
		Commands: []*cli.Command{
			{
				Name:      "search",
				Aliases:   []string{"s"},
				Usage:     "search for ROMs",
				ArgsUsage: "<query>",
				Flags: []cli.Flag{
					systemFlag,
					jsonFlag,
					limitFlag,
					columnsFlag,
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					query := strings.Join(cmd.Args().Slice(), " ")
					if query == "" {
						return fmt.Errorf("usage: edgeemu search <query>")
					}

					roms, err := edge.Search(ctx, query, cmd.String(systemFlag.Name))
					if err != nil {
						return err
					}

					if limit := cmd.Int(limitFlag.Name); limit > 0 && len(roms) > limit {
						roms = roms[:limit]
					}

					if cmd.Bool(jsonFlag.Name) {
						enc := json.NewEncoder(os.Stdout)
						enc.SetIndent("", "  ")
						return enc.Encode(roms)
					}

					if len(roms) == 0 {
						fmt.Println("nothing found")
						return nil
					}

					columns := cmd.StringSlice(columnsFlag.Name)
					if len(columns) == 0 {
						columns = []string{"name", "url"}
						if cmd.String(systemFlag.Name) == "all" {
							columns = append(columns, "system")
						}
					}

					return printROMs(roms, columns)
				},
			},
			{
				Name:  "systems",
				Usage: "list available systems",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					systems, err := edge.Systems(ctx)
					if err != nil {
						return err
					}

					return printSystems(systems)
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
