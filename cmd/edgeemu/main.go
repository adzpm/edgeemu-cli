// Command edgeemu is a CLI for searching ROMs on edgeemu.net.
package main

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/adzpm/edgeemu-cli/internal/app"
)

// devVersion is the placeholder of a build nobody stamped.
const devVersion = "dev"

// version is stamped by the release build via
// -ldflags "-X main.version=v1.2.3".
var version = devVersion

func main() {
	root := app.New().Root()
	root.Version = resolveVersion(version, debug.ReadBuildInfo)

	err := root.Run(context.Background(), os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// resolveVersion prefers the stamped version and falls back to the
// module version recorded by the Go toolchain, so that binaries from
// "go install ...@v1.2.3" report their real version instead of "dev".
func resolveVersion(stamped string, readBuildInfo func() (*debug.BuildInfo, bool)) string {
	if stamped != devVersion {
		return stamped
	}

	info, ok := readBuildInfo()
	if !ok || info.Main.Version == "" || info.Main.Version == "(devel)" {
		return stamped
	}

	return info.Main.Version
}
