package main

import (
	"context"
	"fmt"
	"os"

	"github.com/adzpm/edgeemu-cli/internal/app"
)

func main() {
	if err := app.New().Root().Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
