package completion

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/urfave/cli/v3"
)

func runInstall(t *testing.T, args ...string) error {
	t.Helper()

	root := &cli.Command{
		Name:     "edgeemu",
		Commands: []*cli.Command{{Name: "install-completion", Action: Install}},
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
		if err != nil {
			t.Fatalf("rcLine(%s): %v", tc.shell, err)
		}
		if !strings.HasSuffix(rcPath, tc.wantPath) {
			t.Errorf("rcLine(%s) path = %q, want suffix %q", tc.shell, rcPath, tc.wantPath)
		}
		for _, want := range tc.wantIn {
			if !strings.Contains(line, want) {
				t.Errorf("rcLine(%s) line %q missing %q", tc.shell, line, want)
			}
		}
	}

	if _, _, err := rcLine("tcsh", "/home/u"); err == nil {
		t.Error("rcLine(tcsh): want error, got nil")
	}
}

func TestInstallIsIdempotent(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if err := runInstall(t, "zsh"); err != nil {
		t.Fatalf("first install: %v", err)
	}
	if err := runInstall(t, "zsh"); err != nil {
		t.Fatalf("second install: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(home, ".zshrc"))
	if err != nil {
		t.Fatalf("read .zshrc: %v", err)
	}

	if got := strings.Count(string(data), marker); got != 1 {
		t.Errorf("marker appears %d times, want 1:\n%s", got, data)
	}
	if !strings.Contains(string(data), "source <(edgeemu completion zsh)") {
		t.Errorf(".zshrc missing completion hook:\n%s", data)
	}
}

func TestInstallFishCreatesConfigDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if err := runInstall(t, "fish"); err != nil {
		t.Fatalf("install fish: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(home, ".config", "fish", "config.fish"))
	if err != nil {
		t.Fatalf("read config.fish: %v", err)
	}
	if !strings.Contains(string(data), "edgeemu completion fish | source") {
		t.Errorf("config.fish missing hook:\n%s", data)
	}
}

func TestInstallDetectsShellFromEnv(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("SHELL", "/usr/local/bin/bash")

	if err := runInstall(t); err != nil {
		t.Fatalf("install: %v", err)
	}

	if _, err := os.Stat(filepath.Join(home, ".bashrc")); err != nil {
		t.Errorf(".bashrc not created: %v", err)
	}
}

func TestInstallNoShell(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("SHELL", "")

	if err := runInstall(t); err == nil {
		t.Fatal("want error when shell cannot be detected, got nil")
	}
}

func TestInstallUnsupportedShell(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	if err := runInstall(t, "tcsh"); err == nil {
		t.Fatal("want error for unsupported shell, got nil")
	}
}
