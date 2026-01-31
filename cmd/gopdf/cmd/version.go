package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is set at build time via -ldflags
var Version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long:  `Print the version number of gopdf CLI.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("gopdf version %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
