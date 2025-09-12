package main

import (
	"os"

	"github.com/euforicio/spec-kit/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
