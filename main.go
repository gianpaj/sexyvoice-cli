package main

import (
	"fmt"
	"os"

	"github.com/gianpaj/sexyvoice-cli/cmd"
	"github.com/gianpaj/sexyvoice-cli/internal/output/styles"
)

var version = "dev"

func main() {
	rootCmd := cmd.NewRootCmd(version)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, styles.Error(err.Error()))
		os.Exit(1)
	}
}
