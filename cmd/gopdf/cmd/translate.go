package cmd

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ryomak/gopdf"
	"github.com/spf13/cobra"
)

var (
	translateFont       string
	translateBoldFont   string
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
	translateCmd.Flags().StringVar(&translateBoldFont, "bold-font", "", "TTF font file for bold text (optional)")
	translateCmd.Flags().StringVarP(&translateCommand, "command", "c", "", "translation command (receives text on stdin)")
	translateCmd.Flags().StringVarP(&translateUnit, "unit", "u", "block", "translation unit (block|line|sentence)")
	translateCmd.Flags().BoolVar(&translateKeepImages, "keep-images", true, "preserve images in output")
	translateCmd.Flags().BoolVar(&translateDryRun, "dry-run", false, "extract text without translating")
}

func runTranslate(cmd *cobra.Command, args []string) error {
	inputPath := args[0]
	outputPath := args[1]
	log := NewLogger()

	if translateDryRun {
		return runDryRun(inputPath, log)
	}

	log.Header("🌐", "Translate PDF")

	if translateCommand == "" {
		log.Error("Translation command is required")
		return fmt.Errorf("--command is required (e.g., --command \"trans -b :ja\")")
	}

	log.Step("Loading font...")

	var targetFont gopdf.Font
	if translateFont != "" {
		font, err := gopdf.LoadTTF(translateFont)
		if err != nil {
			log.Error("Failed to load font: %v", err)
			return fmt.Errorf("failed to load font: %w", err)
		}
		targetFont = font
		log.Info("Font: %s", filepath.Base(translateFont))
	} else {
		font, err := gopdf.LoadSystemJapaneseFont()
		if err != nil {
			targetFont = gopdf.FontHelvetica
			log.Warning("No Japanese font found, using Helvetica")
		} else {
			targetFont = font
			log.Info("Font: System Japanese font")
		}
	}

	// Boldフォントの読み込み（オプション）
	var targetBoldFont gopdf.Font
	if translateBoldFont != "" {
		font, err := gopdf.LoadTTF(translateBoldFont)
		if err != nil {
			log.Error("Failed to load bold font: %v", err)
			return fmt.Errorf("failed to load bold font: %w", err)
		}
		targetBoldFont = font
		log.Info("Bold Font: %s", filepath.Base(translateBoldFont))
	}

	var unit gopdf.TranslateUnit
	switch strings.ToLower(translateUnit) {
	case "block":
		unit = gopdf.TranslateUnitBlock
		log.Info("Unit: Block")
	case "line":
		unit = gopdf.TranslateUnitLine
		log.Info("Unit: Line")
	case "sentence":
		unit = gopdf.TranslateUnitSentence
		log.Info("Unit: Sentence")
	default:
		log.Error("Unknown translation unit: %s", translateUnit)
		return fmt.Errorf("unknown translation unit: %s", translateUnit)
	}

	log.Info("Command: %s", translateCommand)

	translator := gopdf.TranslateFunc(func(text string) (string, error) {
		return executeTranslateCommand(text, translateCommand)
	})

	opts := gopdf.PDFTranslatorOptions{
		Translator:     translator,
		TargetFont:     targetFont,
		TargetBoldFont: targetBoldFont,
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

	log.Step("Opening %s", filepath.Base(inputPath))
	log.Step("Translating PDF (this may take a while)...")

	err := gopdf.TranslatePDF(inputPath, outputPath, opts)
	if err != nil {
		log.Error("Translation failed: %v", err)
		return fmt.Errorf("translation failed: %w", err)
	}

	log.Divider()
	log.Success("Translated PDF saved: %s", outputPath)
	log.Println()

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

func runDryRun(inputPath string, log *Logger) error {
	log.Header("🔍", "Dry Run - Extract Text")
	log.Step("Opening %s", filepath.Base(inputPath))

	reader, err := gopdf.Open(inputPath)
	if err != nil {
		log.Error("Failed to open PDF: %v", err)
		return fmt.Errorf("failed to open PDF: %w", err)
	}
	defer reader.Close()

	log.Info("Found %d pages", reader.PageCount())
	log.Divider()

	for i := 0; i < reader.PageCount(); i++ {
		log.Section(fmt.Sprintf("📄 Page %d", i+1))

		blocks, err := reader.ExtractPageTextBlocks(i)
		if err != nil {
			log.Error("Failed to extract text from page %d", i+1)
			return fmt.Errorf("failed to extract text from page %d: %w", i+1, err)
		}

		for j, block := range blocks {
			log.Print("\n")
			log.DetailHighlight(fmt.Sprintf("Block %d", j+1), fmt.Sprintf("%s (%.1fpt)", block.Font, block.FontSize))
			log.Print("  %s\n", block.Text)
		}
	}

	log.Divider()
	log.Success("Dry run complete - no files were modified")

	return nil
}
