package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ryomak/gopdf"
	"github.com/spf13/cobra"
)

var (
	markdownMode        string
	markdownPageSize    string
	markdownOrientation string
	markdownFont        string
	markdownUserPwd     string
	markdownOwnerPwd    string
)

var markdownCmd = &cobra.Command{
	Use:   "markdown <input.md> <output.pdf>",
	Short: "Convert Markdown to PDF",
	Long: `Convert a Markdown file to PDF document or presentation slides.

Modes:
  - document: Multi-page document (default)
  - slide: Presentation slides (separated by --- or # headings)

Page sizes:
  - A4, Letter, Legal, A3, A5
  - 16:9, 4:3 (for slides)`,
	Args: cobra.ExactArgs(2),
	RunE: runMarkdown,
}

func init() {
	rootCmd.AddCommand(markdownCmd)

	markdownCmd.Flags().StringVarP(&markdownMode, "mode", "m", "document", "conversion mode (document|slide)")
	markdownCmd.Flags().StringVar(&markdownPageSize, "page-size", "A4", "page size (A4|Letter|Legal|A3|A5|16:9|4:3)")
	markdownCmd.Flags().StringVar(&markdownOrientation, "orientation", "portrait", "page orientation (portrait|landscape)")
	markdownCmd.Flags().StringVar(&markdownFont, "font", "", "TTF font file for Japanese/CJK text")
	markdownCmd.Flags().StringVar(&markdownUserPwd, "user-password", "", "encrypt with user password")
	markdownCmd.Flags().StringVar(&markdownOwnerPwd, "owner-password", "", "encrypt with owner password")
}

func runMarkdown(cmd *cobra.Command, args []string) error {
	inputPath := args[0]
	outputPath := args[1]
	log := NewLogger()

	log.Header(iconText, "Markdown to PDF")
	log.Step("Reading %s", filepath.Base(inputPath))

	var mode gopdf.MarkdownMode
	switch strings.ToLower(markdownMode) {
	case "document", "doc":
		mode = gopdf.MarkdownModeDocument
		log.Info("Mode: Document")
	case "slide", "slides", "presentation":
		mode = gopdf.MarkdownModeSlide
		log.Info("Mode: Presentation Slides")
	default:
		log.Error("Unknown mode: %s", markdownMode)
		return fmt.Errorf("unknown mode: %s (use document or slide)", markdownMode)
	}

	pageSize, err := parsePageSize(markdownPageSize)
	if err != nil {
		log.Error("Unknown page size: %s", markdownPageSize)
		return err
	}
	log.Info("Page size: %s", markdownPageSize)

	var orientation gopdf.Orientation
	switch strings.ToLower(markdownOrientation) {
	case "portrait", "p":
		orientation = gopdf.Portrait
	case "landscape", "l":
		orientation = gopdf.Landscape
		log.Info("Orientation: Landscape")
	default:
		log.Error("Unknown orientation: %s", markdownOrientation)
		return fmt.Errorf("unknown orientation: %s (use portrait or landscape)", markdownOrientation)
	}

	opts := &gopdf.MarkdownOptions{
		Mode:        mode,
		PageSize:    pageSize,
		Orientation: orientation,
	}

	log.Step("Converting Markdown to PDF...")

	doc, err := gopdf.NewMarkdownDocumentFromFile(inputPath, opts)
	if err != nil {
		log.Error("Conversion failed: %v", err)
		return fmt.Errorf("failed to convert markdown: %w", err)
	}

	if markdownUserPwd != "" {
		log.Step("Applying encryption...")
		ownerPwd := markdownOwnerPwd
		if ownerPwd == "" {
			ownerPwd = markdownUserPwd
		}
		err = doc.SetEncryption(gopdf.EncryptionOptions{
			UserPassword:  markdownUserPwd,
			OwnerPassword: ownerPwd,
			Permissions:   gopdf.DefaultPermissions(),
			KeyLength:     128,
		})
		if err != nil {
			log.Error("Failed to set encryption")
			return fmt.Errorf("failed to set encryption: %w", err)
		}
		log.Success("Encryption applied")
	}

	log.Step("Writing PDF...")

	output, err := os.Create(outputPath)
	if err != nil {
		log.Error("Failed to create output file")
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer output.Close()

	if err := doc.WriteTo(output); err != nil {
		log.Error("Failed to write PDF")
		return fmt.Errorf("failed to write PDF: %w", err)
	}

	log.Divider()
	log.Success("PDF created: %s", outputPath)
	log.Section("📋 Details")
	log.Table("Input", filepath.Base(inputPath))
	log.Table("Output", filepath.Base(outputPath))
	log.Table("Mode", markdownMode)
	log.Table("Page Size", markdownPageSize)
	log.Table("Orientation", markdownOrientation)
	if markdownUserPwd != "" {
		log.Table("Encrypted", "Yes")
	}
	log.Println()

	return nil
}

func parsePageSize(size string) (gopdf.PageSize, error) {
	switch strings.ToUpper(size) {
	case "A4":
		return gopdf.PageSizeA4, nil
	case "A3":
		return gopdf.PageSizeA3, nil
	case "A5":
		return gopdf.PageSizeA5, nil
	case "LETTER":
		return gopdf.PageSizeLetter, nil
	case "LEGAL":
		return gopdf.PageSizeLegal, nil
	case "16:9":
		return gopdf.PageSizePresentation16x9, nil
	case "4:3":
		return gopdf.PageSizePresentation4x3, nil
	default:
		return gopdf.PageSize{}, fmt.Errorf("unknown page size: %s", size)
	}
}
