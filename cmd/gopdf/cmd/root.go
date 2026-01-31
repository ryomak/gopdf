package cmd

import (
	"github.com/spf13/cobra"
)

var (
	verbose bool
	quiet   bool
)

var rootCmd = &cobra.Command{
	Use:   "gopdf",
	Short: "A CLI tool for PDF operations",
	Long: `gopdf is a command-line tool for PDF operations.

It provides various features for working with PDF files:
  - View PDF information (pages, size, metadata)
  - Extract text and images from PDF
  - Encrypt/decrypt PDF files
  - Convert Markdown to PDF
  - And more...

For more information, visit: https://github.com/ryomak/gopdf`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "quiet mode (errors only)")
}
