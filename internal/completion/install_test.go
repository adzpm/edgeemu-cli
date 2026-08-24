package completion

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func runInstall(t *testing.T, args ...string) error {
	t.Helper()

	root := &cli.Command{
		Name:     "edgeemu",
		Commands: []*cli.Command{{Name: "install-completion", Action: New().Install}},
	}

	return root.Run(context.Background(), append([]string{"edgeemu", "install-completion"}, args...))
}

func TestRcLine(t *testing.T) {
	for _, tc := range []struct {
		shell    string
		wantPath string
		wantIn   []string
	}{
		{"zsh", ".zshrc", []string{"command -v edgeemu", "command -v compdef", "edgeemu completion zsh"}},
		{"bash", ".bashrc", []string{"command -v edgeemu", "edgeemu completion bash"}},
		{"fish", filepath.Join(".config", "fish", "config.fish"), []string{"command -q edgeemu", "edgeemu completion fish"}},
	} {
		rcPath, line, err := rcLine(tc.shell, "/home/u")
		require.NoError(t, err, tc.shell)

		assert.True(t, strings.HasSuffix(rcPath, tc.wantPath), "rcLine(%s) path = %q, want suffix %q", tc.shell, rcPath, tc.wantPath)
		for _, want := range tc.wantIn {
			assert.Contains(t, line, want, tc.shell)
		}
	}

	_, _, err := rcLine("tcsh", "/home/u")
	require.Error(t, err)
}

func TestInstallIsIdempotent(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	require.NoError(t, runInstall(t, "zsh"))
	require.NoError(t, runInstall(t, "zsh"))

	data, err := os.ReadFile(filepath.Join(home, ".zshrc"))
	require.NoError(t, err)

	assert.Equal(t, 1, strings.Count(string(data), marker), "repeated install must not duplicate the hook")
	assert.Contains(t, string(data), "source <(edgeemu completion zsh)")
}

func TestInstallFishCreatesConfigDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	require.NoError(t, runInstall(t, "fish"))

	data, err := os.ReadFile(filepath.Join(home, ".config", "fish", "config.fish"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "edgeemu completion fish | source")
}

func TestInstallDetectsShellFromEnv(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("SHELL", "/usr/local/bin/bash")

	require.NoError(t, runInstall(t))

	assert.FileExists(t, filepath.Join(home, ".bashrc"))
}

func TestInstallNoShell(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("SHELL", "")

	require.Error(t, runInstall(t), "undetectable shell must be reported")
}

func TestInstallUnsupportedShell(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	require.Error(t, runInstall(t, "tcsh"))
}

func TestInstallFailsWhenConfigDirIsAFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	// fish needs ~/.config/fish/; a file in the way must surface an error.
	require.NoError(t, os.WriteFile(filepath.Join(home, ".config"), []byte("not a dir"), 0o644))

	require.Error(t, runInstall(t, "fish"))
}

func TestInstallFailsWhenRcPathIsADirectory(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	require.NoError(t, os.MkdirAll(filepath.Join(home, ".zshrc"), 0o755))

	require.Error(t, runInstall(t, "zsh"))
}
