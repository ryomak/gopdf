package cmd

import (
	"fmt"
	"os"
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

	var mode gopdf.MarkdownMode
	switch strings.ToLower(markdownMode) {
	case "document", "doc":
		mode = gopdf.MarkdownModeDocument
	case "slide", "slides", "presentation":
		mode = gopdf.MarkdownModeSlide
	default:
		return fmt.Errorf("unknown mode: %s (use document or slide)", markdownMode)
	}

	pageSize, err := parsePageSize(markdownPageSize)
	if err != nil {
		return err
	}

	var orientation gopdf.Orientation
	switch strings.ToLower(markdownOrientation) {
	case "portrait", "p":
		orientation = gopdf.Portrait
	case "landscape", "l":
		orientation = gopdf.Landscape
	default:
		return fmt.Errorf("unknown orientation: %s (use portrait or landscape)", markdownOrientation)
	}

	opts := &gopdf.MarkdownOptions{
		Mode:        mode,
		PageSize:    pageSize,
		Orientation: orientation,
	}

	doc, err := gopdf.NewMarkdownDocumentFromFile(inputPath, opts)
	if err != nil {
		return fmt.Errorf("failed to convert markdown: %w", err)
	}

	if markdownUserPwd != "" {
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
			return fmt.Errorf("failed to set encryption: %w", err)
		}
	}

	output, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer output.Close()

	if err := doc.WriteTo(output); err != nil {
		return fmt.Errorf("failed to write PDF: %w", err)
	}

	if !quiet {
		fmt.Printf("PDF created: %s\n", outputPath)
		fmt.Printf("  Mode: %s\n", markdownMode)
		fmt.Printf("  Page size: %s\n", markdownPageSize)
		fmt.Printf("  Orientation: %s\n", markdownOrientation)
		if markdownUserPwd != "" {
			fmt.Println("  Encrypted: yes")
		}
	}

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
