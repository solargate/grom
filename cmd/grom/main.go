package main

import (
	"os"

	"github.com/solargate/grom/cmd/grom/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
