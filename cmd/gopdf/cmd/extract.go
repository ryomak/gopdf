package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ryomak/gopdf"
	"github.com/spf13/cobra"
)

var (
	extractPage     int
	extractOutput   string
	extractFormat   string
	extractPassword string
)

var extractCmd = &cobra.Command{
	Use:   "extract",
	Short: "Extract content from PDF",
	Long:  `Extract text or images from PDF files.`,
}

var extractTextCmd = &cobra.Command{
	Use:   "text <file.pdf>",
	Short: "Extract text from PDF",
	Long: `Extract text content from PDF files.

Output formats:
  - plain: Plain text (default)
  - blocks: Text blocks with position info
  - json: Full JSON with coordinates`,
	Args: cobra.ExactArgs(1),
	RunE: runExtractText,
}

var extractImagesCmd = &cobra.Command{
	Use:   "images <file.pdf>",
	Short: "Extract images from PDF",
	Long:  `Extract all images from PDF files to specified directory.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runExtractImages,
}

func init() {
	rootCmd.AddCommand(extractCmd)
	extractCmd.AddCommand(extractTextCmd)
	extractCmd.AddCommand(extractImagesCmd)

	extractTextCmd.Flags().IntVarP(&extractPage, "page", "p", 0, "page number to extract (1-based, 0 for all)")
	extractTextCmd.Flags().StringVarP(&extractOutput, "output", "o", "", "output file (default: stdout)")
	extractTextCmd.Flags().StringVarP(&extractFormat, "format", "f", "plain", "output format (plain|blocks|json)")
	extractTextCmd.Flags().StringVar(&extractPassword, "password", "", "password for encrypted PDF")

	extractImagesCmd.Flags().IntVarP(&extractPage, "page", "p", 0, "page number to extract (1-based, 0 for all)")
	extractImagesCmd.Flags().StringVarP(&extractOutput, "output", "o", "./extracted_images", "output directory")
	extractImagesCmd.Flags().StringVar(&extractPassword, "password", "", "password for encrypted PDF")
}

type TextBlockJSON struct {
	Text     string  `json:"text"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Width    float64 `json:"width"`
	Height   float64 `json:"height"`
	Font     string  `json:"font"`
	FontSize float64 `json:"font_size"`
}

type PageTextJSON struct {
	Page   int             `json:"page"`
	Blocks []TextBlockJSON `json:"blocks"`
}

func runExtractText(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	reader, err := gopdf.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open PDF: %w", err)
	}
	defer reader.Close()

	if reader.IsEncrypted() && extractPassword != "" {
		if err := reader.AuthenticateWithPassword(extractPassword); err != nil {
			return fmt.Errorf("failed to authenticate: %w", err)
		}
	}

	var output *os.File
	if extractOutput != "" {
		output, err = os.Create(extractOutput)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer output.Close()
	} else {
		output = os.Stdout
	}

	startPage := 0
	endPage := reader.PageCount()

	if extractPage > 0 {
		if extractPage > reader.PageCount() {
			return fmt.Errorf("page %d does not exist (total pages: %d)", extractPage, reader.PageCount())
		}
		startPage = extractPage - 1
		endPage = extractPage
	}

	switch extractFormat {
	case "plain":
		return extractTextPlain(reader, output, startPage, endPage)
	case "blocks":
		return extractTextBlocks(reader, output, startPage, endPage)
	case "json":
		return extractTextJSON(reader, output, startPage, endPage)
	default:
		return fmt.Errorf("unknown format: %s (use plain, blocks, or json)", extractFormat)
	}
}

func extractTextPlain(reader *gopdf.PDFReader, output *os.File, startPage, endPage int) error {
	for i := startPage; i < endPage; i++ {
		blocks, err := reader.ExtractPageTextBlocks(i)
		if err != nil {
			return fmt.Errorf("failed to extract text from page %d: %w", i+1, err)
		}

		for _, block := range blocks {
			fmt.Fprintln(output, block.Text)
		}

		if i < endPage-1 {
			fmt.Fprintln(output)
		}
	}
	return nil
}

func extractTextBlocks(reader *gopdf.PDFReader, output *os.File, startPage, endPage int) error {
	for i := startPage; i < endPage; i++ {
		fmt.Fprintf(output, "=== Page %d ===\n", i+1)

		blocks, err := reader.ExtractPageTextBlocks(i)
		if err != nil {
			return fmt.Errorf("failed to extract text from page %d: %w", i+1, err)
		}

		for j, block := range blocks {
			fmt.Fprintf(output, "\n--- Block %d ---\n", j+1)
			fmt.Fprintf(output, "Position: (%.2f, %.2f)\n", block.Rect.X, block.Rect.Y)
			fmt.Fprintf(output, "Size: %.2f x %.2f\n", block.Rect.Width, block.Rect.Height)
			fmt.Fprintf(output, "Font: %s (%.1fpt)\n", block.Font, block.FontSize)
			fmt.Fprintf(output, "Text: %s\n", block.Text)
		}

		if i < endPage-1 {
			fmt.Fprintln(output)
		}
	}
	return nil
}

func extractTextJSON(reader *gopdf.PDFReader, output *os.File, startPage, endPage int) error {
	pages := make([]PageTextJSON, 0, endPage-startPage)

	for i := startPage; i < endPage; i++ {
		blocks, err := reader.ExtractPageTextBlocks(i)
		if err != nil {
			return fmt.Errorf("failed to extract text from page %d: %w", i+1, err)
		}

		pageJSON := PageTextJSON{
			Page:   i + 1,
			Blocks: make([]TextBlockJSON, 0, len(blocks)),
		}

		for _, block := range blocks {
			pageJSON.Blocks = append(pageJSON.Blocks, TextBlockJSON{
				Text:     block.Text,
				X:        block.Rect.X,
				Y:        block.Rect.Y,
				Width:    block.Rect.Width,
				Height:   block.Rect.Height,
				Font:     block.Font,
				FontSize: block.FontSize,
			})
		}

		pages = append(pages, pageJSON)
	}

	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	return encoder.Encode(pages)
}

func runExtractImages(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	reader, err := gopdf.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open PDF: %w", err)
	}
	defer reader.Close()

	if reader.IsEncrypted() && extractPassword != "" {
		if err := reader.AuthenticateWithPassword(extractPassword); err != nil {
			return fmt.Errorf("failed to authenticate: %w", err)
		}
	}

	if err := os.MkdirAll(extractOutput, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	startPage := 0
	endPage := reader.PageCount()

	if extractPage > 0 {
		if extractPage > reader.PageCount() {
			return fmt.Errorf("page %d does not exist (total pages: %d)", extractPage, reader.PageCount())
		}
		startPage = extractPage - 1
		endPage = extractPage
	}

	totalImages := 0

	for i := startPage; i < endPage; i++ {
		layout, err := reader.ExtractPageLayout(i)
		if err != nil {
			return fmt.Errorf("failed to extract layout from page %d: %w", i+1, err)
		}

		for j, img := range layout.Images {
			ext := strings.ToLower(string(img.Format))
			if ext == "" {
				ext = "bin"
			}

			fileName := fmt.Sprintf("page%d_image%d.%s", i+1, j+1, ext)
			filePath := filepath.Join(extractOutput, fileName)

			if err := os.WriteFile(filePath, img.Data, 0644); err != nil {
				return fmt.Errorf("failed to write image %s: %w", fileName, err)
			}

			if !quiet {
				fmt.Printf("Extracted: %s (%dx%d)\n", fileName, img.Width, img.Height)
			}
			totalImages++
		}
	}

	if !quiet {
		fmt.Printf("\nTotal images extracted: %d\n", totalImages)
	}

	return nil
}
