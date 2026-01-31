package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ryomak/gopdf"
	"github.com/spf13/cobra"
)

var (
	metadataJSON     bool
	metadataPassword string
)

var metadataCmd = &cobra.Command{
	Use:   "metadata",
	Short: "Manage PDF metadata",
	Long:  `Read or modify PDF document metadata.`,
}

var metadataGetCmd = &cobra.Command{
	Use:   "get <file.pdf>",
	Short: "Get PDF metadata",
	Long: `Display metadata from a PDF file including:
  - Title, Author, Subject, Keywords
  - Creator, Producer
  - Creation and modification dates`,
	Args: cobra.ExactArgs(1),
	RunE: runMetadataGet,
}

func init() {
	rootCmd.AddCommand(metadataCmd)
	metadataCmd.AddCommand(metadataGetCmd)

	metadataGetCmd.Flags().BoolVar(&metadataJSON, "json", false, "output in JSON format")
	metadataGetCmd.Flags().StringVarP(&metadataPassword, "password", "p", "", "password for encrypted PDF")
}

func runMetadataGet(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	reader, err := gopdf.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open PDF: %w", err)
	}
	defer reader.Close()

	if reader.IsEncrypted() && metadataPassword != "" {
		if err := reader.AuthenticateWithPassword(metadataPassword); err != nil {
			return fmt.Errorf("failed to authenticate: %w", err)
		}
	}

	meta := reader.Info()

	if metadataJSON {
		output := map[string]interface{}{
			"title":    meta.Title,
			"author":   meta.Author,
			"subject":  meta.Subject,
			"keywords": meta.Keywords,
			"creator":  meta.Creator,
			"producer": meta.Producer,
		}
		if !meta.CreationDate.IsZero() {
			output["creation_date"] = meta.CreationDate.Format(dateFormat)
		}
		if !meta.ModDate.IsZero() {
			output["mod_date"] = meta.ModDate.Format(dateFormat)
		}
		if len(meta.Custom) > 0 {
			output["custom"] = meta.Custom
		}

		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(output)
	}

	fmt.Println("PDF Metadata:")
	printIfNotEmpty("  Title", meta.Title)
	printIfNotEmpty("  Author", meta.Author)
	printIfNotEmpty("  Subject", meta.Subject)
	printIfNotEmpty("  Keywords", meta.Keywords)
	printIfNotEmpty("  Creator", meta.Creator)
	printIfNotEmpty("  Producer", meta.Producer)
	if !meta.CreationDate.IsZero() {
		fmt.Printf("  Created: %s\n", meta.CreationDate.Format(dateFormat))
	}
	if !meta.ModDate.IsZero() {
		fmt.Printf("  Modified: %s\n", meta.ModDate.Format(dateFormat))
	}

	if len(meta.Custom) > 0 {
		fmt.Println("\nCustom Fields:")
		for k, v := range meta.Custom {
			fmt.Printf("  %s: %s\n", k, v)
		}
	}

	return nil
}

func printIfNotEmpty(label, value string) {
	if value != "" {
		fmt.Printf("%s: %s\n", label, value)
	}
}
