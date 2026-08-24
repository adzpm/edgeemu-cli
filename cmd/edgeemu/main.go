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

	if err := root.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
