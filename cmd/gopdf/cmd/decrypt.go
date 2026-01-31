package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ryomak/gopdf"
	"github.com/spf13/cobra"
)

var (
	decryptPassword string
)

var decryptCmd = &cobra.Command{
	Use:   "decrypt <input.pdf> <output.pdf>",
	Short: "Decrypt a PDF file",
	Long: `Decrypt a password-protected PDF file.

This command reads an encrypted PDF, authenticates with the provided password,
extracts its content, and creates a new unencrypted PDF.

Note: Complex layouts may not be perfectly preserved due to PDF reconstruction.`,
	Args: cobra.ExactArgs(2),
	RunE: runDecrypt,
}

func init() {
	rootCmd.AddCommand(decryptCmd)

	decryptCmd.Flags().StringVarP(&decryptPassword, "password", "p", "", "password to decrypt the PDF")
	_ = decryptCmd.MarkFlagRequired("password")
}

func runDecrypt(cmd *cobra.Command, args []string) error {
	inputPath := args[0]
	outputPath := args[1]
	log := NewLogger()

	log.Header(iconUnlock, "Decrypt PDF")
	log.Step("Opening %s", filepath.Base(inputPath))

	reader, err := gopdf.Open(inputPath)
	if err != nil {
		log.Error("Failed to open PDF: %v", err)
		return fmt.Errorf("failed to open input PDF: %w", err)
	}
	defer reader.Close()

	if !reader.IsEncrypted() {
		log.Warning("PDF is not encrypted")
		return fmt.Errorf("PDF is not encrypted")
	}

	log.Info("PDF is encrypted")
	log.Step("Authenticating...")

	if err := reader.AuthenticateWithPassword(decryptPassword); err != nil {
		log.Error("Authentication failed - incorrect password")
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	log.Success("Authentication successful")

	encInfo := reader.GetEncryptionInfo()
	if encInfo != nil {
		log.Section("🔐 Original Encryption")
		log.Table("Algorithm", fmt.Sprintf("V%d R%d", encInfo.V, encInfo.R))
		log.Table("Key Length", fmt.Sprintf("%d-bit", encInfo.Length))
		if encInfo.IsOwner {
			log.Table("Auth Level", "Owner (full access)")
		} else {
			log.Table("Auth Level", "User (restricted)")
		}
	}

	log.Info("Found %d pages", reader.PageCount())
	log.Step("Reconstructing PDF without encryption...")

	doc := gopdf.New()

	for i := 0; i < reader.PageCount(); i++ {
		log.Verbose("Processing page %d/%d", i+1, reader.PageCount())

		layout, err := reader.ExtractPageLayout(i)
		if err != nil {
			log.Error("Failed to extract page %d", i+1)
			return fmt.Errorf("failed to extract layout from page %d: %w", i+1, err)
		}

		_, err = gopdf.RenderLayout(doc, layout, gopdf.PDFTranslatorOptions{
			TargetFont: gopdf.FontHelvetica,
			KeepImages: true,
			KeepLayout: true,
			FittingOptions: gopdf.FitOptions{
				MaxFontSize: 72,
				MinFontSize: 6,
				AllowShrink: true,
				AllowGrow:   false,
			},
		})
		if err != nil {
			log.Error("Failed to render page %d", i+1)
			return fmt.Errorf("failed to render page %d: %w", i+1, err)
		}
	}

	log.Step("Writing decrypted PDF...")

	output, err := os.Create(outputPath)
	if err != nil {
		log.Error("Failed to create output file")
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer output.Close()

	if err := doc.WriteTo(output); err != nil {
		log.Error("Failed to write PDF")
		return fmt.Errorf("failed to write decrypted PDF: %w", err)
	}

	log.Divider()
	log.Success("Decrypted PDF created: %s", outputPath)
	log.Println()

	return nil
}
