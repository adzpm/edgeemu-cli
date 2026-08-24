// Command edgeemu is a CLI for searching ROMs on edgeemu.net.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/adzpm/edgeemu-cli/internal/app"
)

// version is stamped by the release build via
// -ldflags "-X main.version=v1.2.3".
var version = "dev"

func main() {
	root := app.New().Root()
	root.Version = version

	err := root.Run(context.Background(), os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
