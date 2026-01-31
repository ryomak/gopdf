package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

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
	log := NewLogger()

	log.Header(iconText, "PDF Metadata")
	log.Step("Opening %s", filepath.Base(filePath))

	reader, err := gopdf.Open(filePath)
	if err != nil {
		log.Error("Failed to open PDF: %v", err)
		return fmt.Errorf("failed to open PDF: %w", err)
	}
	defer reader.Close()

	if reader.IsEncrypted() && metadataPassword != "" {
		log.Step("Authenticating...")
		if err := reader.AuthenticateWithPassword(metadataPassword); err != nil {
			log.Error("Authentication failed")
			return fmt.Errorf("failed to authenticate: %w", err)
		}
		log.Success("Authenticated")
	}

	log.Step("Reading metadata...")

	meta := reader.Info()

	if metadataJSON {
		log.Println()
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

	log.Section("📝 Standard Metadata")
	printMetadataField(log, "Title", meta.Title)
	printMetadataField(log, "Author", meta.Author)
	printMetadataField(log, "Subject", meta.Subject)
	printMetadataField(log, "Keywords", meta.Keywords)
	printMetadataField(log, "Creator", meta.Creator)
	printMetadataField(log, "Producer", meta.Producer)
	if !meta.CreationDate.IsZero() {
		log.Table("Created", meta.CreationDate.Format(dateFormat))
	}
	if !meta.ModDate.IsZero() {
		log.Table("Modified", meta.ModDate.Format(dateFormat))
	}

	if len(meta.Custom) > 0 {
		log.Section("🔧 Custom Fields")
		for k, v := range meta.Custom {
			log.Table(k, v)
		}
	}

	log.Println()
	return nil
}

func printMetadataField(log *Logger, label, value string) {
	if value != "" {
		log.Table(label, value)
	}
}
