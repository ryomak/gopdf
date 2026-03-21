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
	translateFont           string
	translateBoldFont       string
	translateCommand        string
	translateUnit           string
	translateKeepImages     bool
	translateDryRun         bool
	translateMyMemory       string
	translateLibreTranslate string
	translateSourceLang     string
	translateTargetLang     string
	translateEmail          string
)

var translateCmd = &cobra.Command{
	Use:   "translate <input.pdf> <output.pdf>",
	Short: "Translate PDF while preserving layout",
	Long: `Translate a PDF document while preserving its original layout.

The translation can be performed using:
  1. An external command (--command)
  2. MyMemory free API (--mymemory <target-lang>)
  3. LibreTranslate API (--libretranslate <endpoint>)

Examples:
  # Using translate-shell (trans command)
  gopdf translate input.pdf output.pdf --font japanese.ttf --command "trans -b :ja"

  # Using MyMemory free translation API (no API key needed)
  gopdf translate input.pdf output.pdf --font japanese.ttf --mymemory ja
  gopdf translate input.pdf output.pdf --font japanese.ttf --mymemory ja --source-lang en

  # Using LibreTranslate (self-hosted)
  gopdf translate input.pdf output.pdf --font japanese.ttf --libretranslate http://localhost:5000 --source-lang en --target-lang ja

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
	translateCmd.Flags().StringVar(&translateMyMemory, "mymemory", "", "use MyMemory free API with target language (e.g., ja, zh, ko)")
	translateCmd.Flags().StringVar(&translateLibreTranslate, "libretranslate", "", "use LibreTranslate API at given endpoint URL")
	translateCmd.Flags().StringVar(&translateSourceLang, "source-lang", "en", "source language code (default: en)")
	translateCmd.Flags().StringVar(&translateTargetLang, "target-lang", "", "target language code for LibreTranslate (e.g., ja, zh, ko)")
	translateCmd.Flags().StringVar(&translateEmail, "email", "", "email for MyMemory API (increases daily limit to 50000 chars)")
}

func runTranslate(cmd *cobra.Command, args []string) error {
	inputPath := args[0]
	outputPath := args[1]
	log := NewLogger()

	if translateDryRun {
		return runDryRun(inputPath, log)
	}

	log.Header("🌐", "Translate PDF")

	translator, err := buildTranslator(log)
	if err != nil {
		return err
	}

	targetFont, targetBoldFont, err := loadTranslateFonts(log)
	if err != nil {
		return err
	}

	unit, err := parseTranslateUnit(log)
	if err != nil {
		return err
	}

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

	if err := gopdf.TranslatePDF(inputPath, outputPath, opts); err != nil {
		log.Error("Translation failed: %v", err)
		return fmt.Errorf("translation failed: %w", err)
	}

	log.Divider()
	log.Success("Translated PDF saved: %s", outputPath)
	log.Println()

	return nil
}

// buildTranslator は翻訳エンジンを構築する
func buildTranslator(log *Logger) (gopdf.Translator, error) {
	translatorCount := 0
	if translateCommand != "" {
		translatorCount++
	}
	if translateMyMemory != "" {
		translatorCount++
	}
	if translateLibreTranslate != "" {
		translatorCount++
	}

	if translatorCount == 0 {
		log.Error("No translation method specified")
		return nil, fmt.Errorf("specify one of: --command, --mymemory, or --libretranslate")
	}
	if translatorCount > 1 {
		log.Error("Multiple translation methods specified")
		return nil, fmt.Errorf("specify only one of: --command, --mymemory, or --libretranslate")
	}

	switch {
	case translateCommand != "":
		log.Info("Method: External command")
		log.Info("Command: %s", translateCommand)
		return gopdf.TranslateFunc(func(text string) (string, error) {
			return executeTranslateCommand(text, translateCommand)
		}), nil

	case translateMyMemory != "":
		log.Info("Method: MyMemory free API")
		log.Info("Source: %s -> Target: %s", translateSourceLang, translateMyMemory)
		t := gopdf.NewMyMemoryTranslator(translateSourceLang, translateMyMemory)
		if translateEmail != "" {
			t.Email = translateEmail
			log.Info("Email: %s (50000 chars/day limit)", translateEmail)
		} else {
			log.Info("Limit: 5000 chars/day (use --email to increase)")
		}
		return t, nil

	case translateLibreTranslate != "":
		if translateTargetLang == "" {
			return nil, fmt.Errorf("--target-lang is required with --libretranslate (e.g., --target-lang ja)")
		}
		log.Info("Method: LibreTranslate API")
		log.Info("Endpoint: %s", translateLibreTranslate)
		log.Info("Source: %s -> Target: %s", translateSourceLang, translateTargetLang)
		return gopdf.NewLibreTranslateTranslator(translateLibreTranslate, translateSourceLang, translateTargetLang), nil

	default:
		return nil, fmt.Errorf("no translation method specified")
	}
}

// loadTranslateFonts はフォントを読み込む
func loadTranslateFonts(log *Logger) (gopdf.Font, gopdf.Font, error) {
	log.Step("Loading font...")

	var targetFont gopdf.Font
	if translateFont != "" {
		f, err := gopdf.LoadTTF(translateFont)
		if err != nil {
			log.Error("Failed to load font: %v", err)
			return nil, nil, fmt.Errorf("failed to load font: %w", err)
		}
		targetFont = f
		log.Info("Font: %s", filepath.Base(translateFont))
	} else {
		f, err := gopdf.LoadSystemJapaneseFont()
		if err != nil {
			targetFont = gopdf.FontHelvetica
			log.Warning("No Japanese font found, using Helvetica")
		} else {
			targetFont = f
			log.Info("Font: System Japanese font")
		}
	}

	var targetBoldFont gopdf.Font
	if translateBoldFont != "" {
		f, err := gopdf.LoadTTF(translateBoldFont)
		if err != nil {
			log.Error("Failed to load bold font: %v", err)
			return nil, nil, fmt.Errorf("failed to load bold font: %w", err)
		}
		targetBoldFont = f
		log.Info("Bold Font: %s", filepath.Base(translateBoldFont))
	}

	return targetFont, targetBoldFont, nil
}

// parseTranslateUnit は翻訳単位をパースする
func parseTranslateUnit(log *Logger) (gopdf.TranslateUnit, error) {
	switch strings.ToLower(translateUnit) {
	case "block":
		log.Info("Unit: Block")
		return gopdf.TranslateUnitBlock, nil
	case "line":
		log.Info("Unit: Line")
		return gopdf.TranslateUnitLine, nil
	case "sentence":
		log.Info("Unit: Sentence")
		return gopdf.TranslateUnitSentence, nil
	default:
		log.Error("Unknown translation unit: %s", translateUnit)
		return 0, fmt.Errorf("unknown translation unit: %s", translateUnit)
	}
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
