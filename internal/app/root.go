package app

import (
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/adzpm/edgeemu-cli/internal/cache"
	"github.com/adzpm/edgeemu-cli/internal/client"
	"github.com/adzpm/edgeemu-cli/internal/completion"
	"github.com/adzpm/edgeemu-cli/internal/render"
)

var (
	systemFlag = &cli.StringFlag{
		Name:    "system",
		Aliases: []string{"s"},
		Value:   "all",
		Usage:   "system to search in (see 'edgeemu systems')",
	}

	formatFlag = &cli.StringFlag{
		Name:    "format",
		Aliases: []string{"f"},
		Value:   "list",
		Usage:   "output format: list, json, yaml, xml or csv",
	}

	limitFlag = &cli.IntFlag{
		Name:    "limit",
		Aliases: []string{"l"},
		Usage:   "max results to show (0 = all)",
	}

	columnsFlag = &cli.StringSliceFlag{
		Name:    "columns",
		Aliases: []string{"c"},
		Usage:   "columns to show, comma-separated (" + strings.Join(render.ColumnIDs(), ", ") + ")",
	}

	refreshFlag = &cli.BoolFlag{
		Name:    "refresh",
		Aliases: []string{"r"},
		Usage:   "bypass the cache and refetch",
	}
)

// App wires the edgeemu.net client into the CLI command actions.
type App struct {
	edge    *client.Client
	cache   *cache.Cache
	printer *render.Printer
	comp    *completion.Completion
}

// Option customizes an App.
type Option func(*App)

// WithClient overrides the edgeemu.net client.
func WithClient(edge *client.Client) Option {
	return func(a *App) { a.edge = edge }
}

// WithCache overrides the systems cache.
func WithCache(c *cache.Cache) Option {
	return func(a *App) { a.cache = c }
}

// WithPrinter overrides the output printer.
func WithPrinter(p *render.Printer) Option {
	return func(a *App) { a.printer = p }
}

// WithCompletion overrides the shell completion provider.
func WithCompletion(c *completion.Completion) Option {
	return func(a *App) { a.comp = c }
}

// New creates the application with sane defaults, applying the given
// options. Dependencies not overridden are built around the client.
func New(opts ...Option) *App {
	a := &App{}

	for _, opt := range opts {
		opt(a)
	}

	if a.edge == nil {
		a.edge = client.New()
	}
	if a.cache == nil {
		a.cache = cache.New(cache.WithClient(a.edge))
	}
	if a.printer == nil {
		a.printer = render.New()
	}
	if a.comp == nil {
		a.comp = completion.New(completion.WithCache(a.cache))
	}

	return a
}

// Root builds the root CLI command tree.
func (a *App) Root() *cli.Command {
	return &cli.Command{
		Name:                  "edgeemu",
		Usage:                 "search ROMs on edgeemu.net",
		EnableShellCompletion: true,
		Commands: []*cli.Command{
			{
				Name:      "search",
				Aliases:   []string{"s"},
				Usage:     "search for ROMs",
				ArgsUsage: "<query>",
				Flags: []cli.Flag{
					systemFlag,
					formatFlag,
					limitFlag,
					columnsFlag,
				},
				ShellComplete: a.comp.Search,
				Action:        a.Search,
			},
			{
				Name:   "systems",
				Usage:  "list available systems",
				Flags:  []cli.Flag{formatFlag, refreshFlag},
				Action: a.Systems,
			},
			{
				Name:      "install-completion",
				Usage:     "install shell completion (zsh, bash, or fish)",
				ArgsUsage: "[shell]",
				Action:    a.comp.Install,
			},
		},
	}
}
