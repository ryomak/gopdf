package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/ryomak/gopdf"
	"github.com/spf13/cobra"
)

var (
	translateFont       string
	translateCommand    string
	translateUnit       string
	translateKeepImages bool
	translateDryRun     bool
)

var translateCmd = &cobra.Command{
	Use:   "translate <input.pdf> <output.pdf>",
	Short: "Translate PDF while preserving layout",
	Long: `Translate a PDF document while preserving its original layout.

The translation is performed using an external command that receives
text on stdin and outputs translated text on stdout.

Examples:
  # Using translate-shell (trans command)
  gopdf translate input.pdf output.pdf --font japanese.ttf --command "trans -b :ja"

  # Using custom script
  gopdf translate input.pdf output.pdf --font font.ttf --command "./my-translator.sh"

  # Dry run to see extractable text
  gopdf translate input.pdf output.pdf --dry-run

Translation units:
  - block: Translate entire text blocks (default)
  - line: Translate line by line
  - sentence: Translate sentence by sentence`,
	Args: cobra.ExactArgs(2),
	RunE: runTranslate,
}

func init() {
	rootCmd.AddCommand(translateCmd)

	translateCmd.Flags().StringVarP(&translateFont, "font", "f", "", "TTF font file for output (required for CJK)")
	translateCmd.Flags().StringVarP(&translateCommand, "command", "c", "", "translation command (receives text on stdin)")
	translateCmd.Flags().StringVarP(&translateUnit, "unit", "u", "block", "translation unit (block|line|sentence)")
	translateCmd.Flags().BoolVar(&translateKeepImages, "keep-images", true, "preserve images in output")
	translateCmd.Flags().BoolVar(&translateDryRun, "dry-run", false, "extract text without translating")
}

func runTranslate(cmd *cobra.Command, args []string) error {
	inputPath := args[0]
	outputPath := args[1]

	if translateDryRun {
		return runDryRun(inputPath)
	}

	if translateCommand == "" {
		return fmt.Errorf("--command is required (e.g., --command \"trans -b :ja\")")
	}

	var targetFont gopdf.Font
	if translateFont != "" {
		font, err := gopdf.LoadTTF(translateFont)
		if err != nil {
			return fmt.Errorf("failed to load font: %w", err)
		}
		targetFont = font
	} else {
		font, err := gopdf.LoadSystemJapaneseFont()
		if err != nil {
			targetFont = gopdf.FontHelvetica
			if !quiet {
				fmt.Println("Warning: No Japanese font specified, using Helvetica")
			}
		} else {
			targetFont = font
		}
	}

	var unit gopdf.TranslateUnit
	switch strings.ToLower(translateUnit) {
	case "block":
		unit = gopdf.TranslateUnitBlock
	case "line":
		unit = gopdf.TranslateUnitLine
	case "sentence":
		unit = gopdf.TranslateUnitSentence
	default:
		return fmt.Errorf("unknown translation unit: %s", translateUnit)
	}

	translator := gopdf.TranslateFunc(func(text string) (string, error) {
		return executeTranslateCommand(text, translateCommand)
	})

	opts := gopdf.PDFTranslatorOptions{
		Translator: translator,
		TargetFont: targetFont,
		FittingOptions: gopdf.FitOptions{
			MaxFontSize: 72,
			MinFontSize: 6,
			LineSpacing: 1.2,
			Padding:     2.0,
			AllowShrink: true,
			AllowGrow:   false,
			Alignment:   gopdf.AlignLeft,
		},
		KeepImages:    translateKeepImages,
		KeepLayout:    true,
		TranslateUnit: unit,
	}

	if !quiet {
		fmt.Printf("Translating %s...\n", inputPath)
	}

	err := gopdf.TranslatePDF(inputPath, outputPath, opts)
	if err != nil {
		return fmt.Errorf("translation failed: %w", err)
	}

	if !quiet {
		fmt.Printf("Translated PDF saved: %s\n", outputPath)
	}

	return nil
}

func executeTranslateCommand(text string, command string) (string, error) {
	if strings.TrimSpace(text) == "" {
		return text, nil
	}

	cmd := exec.Command("sh", "-c", command)
	cmd.Stdin = strings.NewReader(text)

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("command failed: %s", string(exitErr.Stderr))
		}
		return "", fmt.Errorf("command failed: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

func runDryRun(inputPath string) error {
	reader, err := gopdf.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open PDF: %w", err)
	}
	defer reader.Close()

	fmt.Printf("=== Extractable Text from %s ===\n\n", inputPath)

	for i := 0; i < reader.PageCount(); i++ {
		fmt.Printf("--- Page %d ---\n", i+1)

		blocks, err := reader.ExtractPageTextBlocks(i)
		if err != nil {
			return fmt.Errorf("failed to extract text from page %d: %w", i+1, err)
		}

		for j, block := range blocks {
			fmt.Printf("\nBlock %d:\n", j+1)
			fmt.Printf("  Font: %s (%.1fpt)\n", block.Font, block.FontSize)
			fmt.Printf("  Text: %s\n", block.Text)
		}
		fmt.Println()
	}

	return nil
}

