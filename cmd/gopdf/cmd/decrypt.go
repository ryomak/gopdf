package cmd

import (
	"fmt"
	"os"

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

	reader, err := gopdf.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input PDF: %w", err)
	}
	defer reader.Close()

	if !reader.IsEncrypted() {
		return fmt.Errorf("PDF is not encrypted")
	}

	if err := reader.AuthenticateWithPassword(decryptPassword); err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	if !quiet {
		fmt.Println("Authentication successful")
	}

	doc := gopdf.New()

	for i := 0; i < reader.PageCount(); i++ {
		layout, err := reader.ExtractPageLayout(i)
		if err != nil {
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
			return fmt.Errorf("failed to render page %d: %w", i+1, err)
		}
	}

	output, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer output.Close()

	if err := doc.WriteTo(output); err != nil {
		return fmt.Errorf("failed to write decrypted PDF: %w", err)
	}

	if !quiet {
		fmt.Printf("Decrypted PDF created: %s\n", outputPath)
	}

	return nil
}
