// Package completion provides dynamic shell completion for the CLI and
// an installer that wires it into the user's shell rc file.
package completion

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/adzpm/edgeemu-cli/internal/cache"
	"github.com/adzpm/edgeemu-cli/internal/render"
)

// fetchTimeout caps the network fetch during shell completion so a slow
// or unreachable site never hangs the user's TAB press.
const fetchTimeout = 3 * time.Second

// marker identifies the completion block in rc files, so reinstalls are
// detected even if the hook line itself changes between versions.
const marker = "# edgeemu shell completion"

// rc file permissions: private to the user.
const (
	rcDirPerm  = 0o750
	rcFilePerm = 0o600
)

// ErrUnsupportedShell is returned for shells without completion support.
var ErrUnsupportedShell = errors.New("unsupported shell")

// ErrShellNotDetected is returned when $SHELL gives no usable shell name.
var ErrShellNotDetected = errors.New(
	"cannot detect shell ($SHELL is not set), pass it explicitly: edgeemu install-completion <zsh|bash|fish>")

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
// -s/--system (served from cache for instant results), column IDs after
// -c/--columns, formats after -f/--format, and flag names otherwise.
//
// Flag suggestions are printed by hand instead of via
// cli.DefaultCompleteWithFlags: the default helper mis-detects the word
// being completed once a positional argument precedes the flag
// ("edgeemu search sonic --<TAB>") and suggests subcommands instead.
func (c *Completion) Search(ctx context.Context, cmd *cli.Command) {
	switch lastArg() {
	case "-s", "--system":
		c.completeSystems(ctx, cmd)
	case "-c", "--columns":
		completeValues(cmd, render.ColumnIDs())
	case "-f", "--format":
		completeValues(cmd, render.Formats())
	default:
		flagSuggestions(cmd)
	}
}

// Install appends the completion hook to the shell rc file. The shell is
// taken from the first argument, falling back to $SHELL.
func (c *Completion) Install(_ context.Context, cmd *cli.Command) error {
	shell := detectShell(cmd)
	if shell == "" {
		return ErrShellNotDetected
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	rcPath, line, err := rcLine(shell, home)
	if err != nil {
		return err
	}

	out := cmd.Root().Writer

	if data, err := os.ReadFile(rcPath); err == nil && strings.Contains(string(data), marker) {
		fmt.Fprintf(out, "completion for %s is already installed in %s\n", shell, rcPath)

		return nil
	}

	if err := appendHook(rcPath, line); err != nil {
		return err
	}

	fmt.Fprintf(out, "added to %s:\n\n    %s\n\nrestart your shell or run 'source %s' to enable it\n",
		rcPath, line, rcPath)

	return nil
}

// completeSystems prints system IDs, preferring the cache of any age;
// a cold cache is filled over the network within fetchTimeout.
func (c *Completion) completeSystems(ctx context.Context, cmd *cli.Command) {
	if c.cache == nil {
		return
	}

	systems := c.cache.Load(0)
	if systems == nil {
		fetchCtx, cancel := context.WithTimeout(ctx, fetchTimeout)
		defer cancel()

		var err error

		systems, err = c.cache.Systems(fetchCtx, false)
		if err != nil {
			return
		}
	}

	for _, s := range systems {
		fmt.Fprintln(cmd.Root().Writer, s.ID)
	}
}

// lastArg returns the word preceding --generate-shell-completion.
func lastArg() string {
	args := os.Args
	if len(args) > 0 && args[len(args)-1] == "--generate-shell-completion" {
		args = args[:len(args)-1]
	}

	if len(args) == 0 {
		return ""
	}

	return args[len(args)-1]
}

func completeValues(cmd *cli.Command, values []string) {
	for _, v := range values {
		fmt.Fprintln(cmd.Root().Writer, v)
	}
}

// flagSuggestions prints every visible flag of cmd as "--name:usage";
// the shell filters them by the prefix the user has typed.
func flagSuggestions(cmd *cli.Command) {
	for _, f := range cmd.VisibleFlags() {
		name := strings.TrimSpace(f.Names()[0])

		prefix := "--"
		if len(name) == 1 {
			prefix = "-"
		}

		usage := ""
		if df, ok := f.(cli.DocGenerationFlag); ok {
			usage = df.GetUsage()
		}

		fmt.Fprintf(cmd.Root().Writer, "%s%s:%s\n", prefix, name, usage)
	}
}

// detectShell picks the shell from the command argument or $SHELL.
func detectShell(cmd *cli.Command) string {
	if shell := cmd.Args().First(); shell != "" {
		return shell
	}

	if env := os.Getenv("SHELL"); env != "" {
		return filepath.Base(env)
	}

	return ""
}

// rcLine returns the shell rc file and the line enabling completion for it.
// The lines are guarded so a missing edgeemu binary (or, for zsh, an
// uninitialized completion system) skips the hook instead of breaking
// every shell startup.
func rcLine(shell, home string) (string, string, error) {
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
		return "", "", fmt.Errorf("%w: %q (supported: zsh, bash, fish)", ErrUnsupportedShell, shell)
	}
}

// appendHook writes the marker and hook line to the end of rcPath,
// creating the file and its directory when missing.
func appendHook(rcPath, line string) error {
	if err := os.MkdirAll(filepath.Dir(rcPath), rcDirPerm); err != nil {
		return err
	}

	f, err := os.OpenFile(rcPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, rcFilePerm)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(f, "\n%s\n%s\n", marker, line); err != nil {
		_ = f.Close()

		return err
	}

	return f.Close()
}
