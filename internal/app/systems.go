package app

import (
	"context"

	"github.com/urfave/cli/v3"
)

// Systems handles the systems command.
func (a *App) Systems(ctx context.Context, cmd *cli.Command) error {
	systems, err := a.cache.Systems(ctx, cmd.Bool(refreshFlag.Name))
	if err != nil {
		return err
	}

	return a.table.PrintSystems(systems)
}
