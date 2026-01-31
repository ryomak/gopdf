package main

import (
	"os"

	"github.com/ryomak/gopdf/cmd/gopdf/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
