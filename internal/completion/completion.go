package completion

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/adzpm/edgeemu-cli/internal/cache"
	"github.com/adzpm/edgeemu-cli/internal/table"
)

// fetchTimeout caps the network fetch during shell completion so a slow
// or unreachable site never hangs the user's TAB press.
const fetchTimeout = 3 * time.Second

// marker identifies the completion block in rc files, so reinstalls are
// detected even if the hook line itself changes between versions.
const marker = "# edgeemu shell completion"

// Completion provides shell completion values and the installer.
type Completion struct {
	cache *cache.Cache
}

// Option customizes a Completion.
type Option func(*Completion)

// WithCache sets the systems cache used to complete -s/--system values.
func WithCache(c *cache.Cache) Option {
	return func(cp *Completion) { cp.cache = c }
}

// New creates a Completion with sane defaults, applying the given options.
// Without WithCache, system ID completion is unavailable (columns and
// default flag completion still work).
func New(opts ...Option) *Completion {
	c := &Completion{}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Search completes values for the search command: system IDs after
// -s/--system (served from cache for instant results) and column IDs
// after -c/--columns, falling back to the default flag completion.
func (c *Completion) Search(ctx context.Context, cmd *cli.Command) {
	args := os.Args
	if len(args) > 0 && args[len(args)-1] == "--generate-shell-completion" {
		args = args[:len(args)-1]
	}

	last := ""
	if len(args) > 0 {
		last = args[len(args)-1]
	}

	switch last {
	case "-s", "--system":
		if c.cache == nil {
			return
		}

		systems := c.cache.Load(0)
		if systems == nil {
			fetchCtx, cancel := context.WithTimeout(ctx, fetchTimeout)
			defer cancel()

			var err error
			if systems, err = c.cache.Systems(fetchCtx, false); err != nil {
				return
			}
		}

		for _, s := range systems {
			fmt.Fprintln(cmd.Root().Writer, s.ID)
		}
	case "-c", "--columns":
		for _, id := range table.ColumnIDs() {
			fmt.Fprintln(cmd.Root().Writer, id)
		}
	default:
		cli.DefaultCompleteWithFlags(ctx, cmd)
	}
}

// rcLine returns the shell rc file and the line enabling completion for it.
// The lines are guarded so a missing edgeemu binary (or, for zsh, an
// uninitialized completion system) skips the hook instead of breaking
// every shell startup.
func rcLine(shell, home string) (rcPath, line string, err error) {
	switch shell {
	case "zsh":
		return filepath.Join(home, ".zshrc"),
			"command -v edgeemu >/dev/null && command -v compdef >/dev/null && source <(edgeemu completion zsh)", nil
	case "bash":
		return filepath.Join(home, ".bashrc"),
			"command -v edgeemu >/dev/null && source <(edgeemu completion bash)", nil
	case "fish":
		return filepath.Join(home, ".config", "fish", "config.fish"),
			"command -q edgeemu; and edgeemu completion fish | source", nil
	default:
		return "", "", fmt.Errorf("unsupported shell %q (supported: zsh, bash, fish)", shell)
	}
}

// Install appends the completion hook to the shell rc file. The shell is
// taken from the first argument, falling back to $SHELL.
func (c *Completion) Install(ctx context.Context, cmd *cli.Command) error {
	shell := cmd.Args().First()
	if shell == "" {
		if env := os.Getenv("SHELL"); env != "" {
			shell = filepath.Base(env)
		}
	}
	if shell == "" {
		return fmt.Errorf("cannot detect shell ($SHELL is not set), pass it explicitly: edgeemu install-completion <zsh|bash|fish>")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	rcPath, line, err := rcLine(shell, home)
	if err != nil {
		return err
	}

	if data, err := os.ReadFile(rcPath); err == nil && strings.Contains(string(data), marker) {
		fmt.Printf("completion for %s is already installed in %s\n", shell, rcPath)
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(rcPath), 0o755); err != nil {
		return err
	}

	f, err := os.OpenFile(rcPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}

	defer f.Close()

	if _, err := fmt.Fprintf(f, "\n%s\n%s\n", marker, line); err != nil {
		return err
	}

	fmt.Printf("added to %s:\n\n    %s\n\nrestart your shell or run 'source %s' to enable it\n", rcPath, line, rcPath)

	return nil
}
