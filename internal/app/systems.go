package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/adzpm/edgeemu-cli/internal/render"
)

// Systems handles the systems command.
func (a *App) Systems(ctx context.Context, cmd *cli.Command) error {
	systems, err := a.cache.Systems(ctx, cmd.Bool(refreshFlag.Name))
	if err != nil {
		return err
	}

	switch format := cmd.String(formatFlag.Name); format {
	case render.FormatJSON:
		return a.printer.JSON(systems)
	case render.FormatYAML:
		return a.printer.YAML(systems)
	case render.FormatXML:
		return a.printer.XMLSystems(systems)
	case render.FormatCSV:
		return a.printer.CSVSystems(systems)
	case render.FormatList:
		return a.printer.PrintSystems(systems)
	default:
		return fmt.Errorf("%w %q (available: %s)", ErrUnknownFormat, format, strings.Join(render.Formats(), ", "))
	}
}
