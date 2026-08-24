package app

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

// Systems handles the systems command.
func (a *App) Systems(ctx context.Context, cmd *cli.Command) error {
	systems, err := a.cache.Systems(ctx, cmd.Bool(refreshFlag.Name))
	if err != nil {
		return err
	}

	switch format := cmd.String(formatFlag.Name); format {
	case "json":
		return a.printer.JSON(systems)
	case "yaml":
		return a.printer.YAML(systems)
	case "xml":
		return a.printer.XMLSystems(systems)
	case "list":
		return a.printer.PrintSystems(systems)
	default:
		return fmt.Errorf("unknown format %q (available: list, json, yaml, xml)", format)
	}
}
