// Example: 11_ruby_annotation
// This example demonstrates how to add ruby (furigana) annotations to text.
package main

import (
	"fmt"
	"os"

	"github.com/ryomak/gopdf"
)

func main() {
	// Check for Japanese font
	ttfFontPath := "NotoSansJP-Regular.ttf"
	useJapanese := false

	var jpFont *gopdf.TTFFont
	if _, err := os.Stat(ttfFontPath); err == nil {
		// Load TTF font
		var err error
		jpFont, err = gopdf.LoadTTF(ttfFontPath)
		if err == nil {
			useJapanese = true
			fmt.Println("Using Japanese TTF font for ruby examples")
		} else {
			fmt.Printf("Warning: Failed to load TTF font: %v\n", err)
		}
	}

	if !useJapanese {
		fmt.Println("Warning: TTF font not found. Skipping Japanese examples")
		fmt.Println("To display Japanese ruby, download NotoSansJP-Regular.ttf from:")
		fmt.Println("https://fonts.google.com/noto/specimen/Noto+Sans+JP")
		fmt.Println("(Use the static font from the static/ folder)")
		return
	}

	// Example 1: Basic ruby annotation
	fmt.Println("\n--- Example 1: Basic Ruby Annotation ---")
	if err := createBasicRubyExample(jpFont, "ruby_basic.pdf"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Example 2: Different alignments
	fmt.Println("\n--- Example 2: Ruby Alignment Options ---")
	if err := createAlignmentExample(jpFont, "ruby_alignment.pdf"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Example 3: ActualText and copy modes
	fmt.Println("\n--- Example 3: ActualText Copy Modes ---")
	if err := createActualTextExample(jpFont, "ruby_actualtext.pdf"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Example 4: Multiple ruby texts
	fmt.Println("\n--- Example 4: Multiple Ruby Texts ---")
	if err := createMultipleRubyExample(jpFont, "ruby_multiple.pdf"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nAll examples completed successfully!")
}

// createBasicRubyExample creates a PDF with basic ruby annotation
func createBasicRubyExample(jpFont *gopdf.TTFFont, filename string) error {
	doc := gopdf.New()
	page := doc.AddPage(gopdf.A4, gopdf.Portrait)

	// Set Japanese font
	page.SetTTFFont(jpFont, 20)

	// Title
	page.DrawTextUTF8("ルビ（ふりがな）の基本例", 50, 800)

	// Basic ruby examples with default style
	style := gopdf.DefaultRubyStyle()
	page.SetTTFFont(jpFont, 16)

	y := 750.0
	rubyTexts := []gopdf.RubyText{
		gopdf.NewRubyText("漢字", "かんじ"),
		gopdf.NewRubyText("日本語", "にほんご"),
		gopdf.NewRubyText("東京", "とうきょう"),
		gopdf.NewRubyText("富士山", "ふじさん"),
	}

	for _, rt := range rubyTexts {
		width, err := page.DrawRuby(rt, 50, y, style)
		if err != nil {
			return err
		}
		y -= 50
		fmt.Printf("  Drew ruby: %s(%s), width: %.1f\n", rt.Base, rt.Ruby, width)
	}

	// Save file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if err := doc.WriteTo(file); err != nil {
		return fmt.Errorf("failed to write PDF: %w", err)
	}

	fmt.Printf("  Created: %s\n", filename)
	return nil
}

// createAlignmentExample creates a PDF demonstrating different ruby alignments
func createAlignmentExample(jpFont *gopdf.TTFFont, filename string) error {
	doc := gopdf.New()
	page := doc.AddPage(gopdf.A4, gopdf.Portrait)

	// Set Japanese font
	page.SetTTFFont(jpFont, 20)
	page.DrawTextUTF8("ルビの配置オプション", 50, 800)

	// Create styles with different alignments
	page.SetTTFFont(jpFont, 16)
	y := 750.0

	// Center alignment (default)
	style := gopdf.DefaultRubyStyle()
	style.Alignment = gopdf.RubyAlignCenter
	page.DrawTextUTF8("中央揃え:", 50, y)
	page.DrawRuby(gopdf.NewRubyText("東京", "とうきょう"), 200, y, style)
	y -= 50

	// Left alignment
	style.Alignment = gopdf.RubyAlignLeft
	page.DrawTextUTF8("左揃え:", 50, y)
	page.DrawRuby(gopdf.NewRubyText("東京", "とうきょう"), 200, y, style)
	y -= 50

	// Right alignment
	style.Alignment = gopdf.RubyAlignRight
	page.DrawTextUTF8("右揃え:", 50, y)
	page.DrawRuby(gopdf.NewRubyText("東京", "とうきょう"), 200, y, style)
	y -= 50

	// Different size ratios
	y -= 20
	page.DrawTextUTF8("サイズ比率:", 50, y)
	y -= 40

	// 30% size
	style = gopdf.DefaultRubyStyle()
	style.SizeRatio = 0.3
	page.DrawTextUTF8("30%:", 50, y)
	page.DrawRuby(gopdf.NewRubyText("漢字", "かんじ"), 200, y, style)
	y -= 50

	// 50% size (default)
	style.SizeRatio = 0.5
	page.DrawTextUTF8("50%:", 50, y)
	page.DrawRuby(gopdf.NewRubyText("漢字", "かんじ"), 200, y, style)
	y -= 50

	// 70% size
	style.SizeRatio = 0.7
	page.DrawTextUTF8("70%:", 50, y)
	page.DrawRuby(gopdf.NewRubyText("漢字", "かんじ"), 200, y, style)

	// Save file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if err := doc.WriteTo(file); err != nil {
		return fmt.Errorf("failed to write PDF: %w", err)
	}

	fmt.Printf("  Created: %s\n", filename)
	return nil
}

// createActualTextExample creates a PDF demonstrating ActualText copy modes
func createActualTextExample(jpFont *gopdf.TTFFont, filename string) error {
	doc := gopdf.New()
	page := doc.AddPage(gopdf.A4, gopdf.Portrait)

	// Set Japanese font
	page.SetTTFFont(jpFont, 20)
	page.DrawTextUTF8("ActualTextコピーモード", 50, 800)

	page.SetTTFFont(jpFont, 12)
	page.DrawTextUTF8("※PDFからテキストをコピーすると動作が確認できます", 50, 780)

	page.SetTTFFont(jpFont, 16)
	y := 730.0
	rubyText := gopdf.NewRubyText("東京", "とうきょう")

	// Copy base only (default)
	style := gopdf.DefaultRubyStyle()
	style.CopyMode = gopdf.RubyCopyBase
	page.DrawTextUTF8("親文字のみコピー:", 50, y)
	page.DrawRubyWithActualText(rubyText, 300, y, style)
	page.SetTTFFont(jpFont, 12)
	page.DrawTextUTF8("→ コピー: 東京", 450, y)
	y -= 50

	// Copy ruby only
	page.SetTTFFont(jpFont, 16)
	style.CopyMode = gopdf.RubyCopyRuby
	page.DrawTextUTF8("ルビのみコピー:", 50, y)
	page.DrawRubyWithActualText(rubyText, 300, y, style)
	page.SetTTFFont(jpFont, 12)
	page.DrawTextUTF8("→ コピー: とうきょう", 450, y)
	y -= 50

	// Copy both
	page.SetTTFFont(jpFont, 16)
	style.CopyMode = gopdf.RubyCopyBoth
	page.DrawTextUTF8("両方コピー:", 50, y)
	page.DrawRubyWithActualText(rubyText, 300, y, style)
	page.SetTTFFont(jpFont, 12)
	page.DrawTextUTF8("→ コピー: 東京(とうきょう)", 450, y)

	// Save file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if err := doc.WriteTo(file); err != nil {
		return fmt.Errorf("failed to write PDF: %w", err)
	}

	fmt.Printf("  Created: %s\n", filename)
	fmt.Println("  Try copying text from the PDF to see ActualText in action!")
	return nil
}

// createMultipleRubyExample creates a PDF with multiple sequential ruby texts
func createMultipleRubyExample(jpFont *gopdf.TTFFont, filename string) error {
	doc := gopdf.New()
	page := doc.AddPage(gopdf.A4, gopdf.Portrait)

	// Set Japanese font
	page.SetTTFFont(jpFont, 20)
	page.DrawTextUTF8("複数のルビテキスト", 50, 800)

	page.SetTTFFont(jpFont, 16)
	y := 750.0

	// Using helper function to create multiple ruby texts
	texts := gopdf.NewRubyTextPairs(
		"私", "わたし",
		"日本", "にほん",
		"住", "す",
	)

	style := gopdf.DefaultRubyStyle()

	// Draw with ActualText support
	page.DrawTextUTF8("文章例:", 50, y)
	totalWidth, err := page.DrawRubyTexts(texts, 150, y, style, true)
	if err != nil {
		return err
	}

	// Continue with non-ruby text
	page.DrawTextUTF8("んでいます。", 150+totalWidth, y)
	fmt.Printf("  Drew multiple ruby texts, total width: %.1f\n", totalWidth)

	y -= 50

	// Another example
	texts2 := gopdf.NewRubyTextPairs(
		"今日", "きょう",
		"天気", "てんき",
		"良", "よ",
	)

	page.DrawTextUTF8("文章例:", 50, y)
	totalWidth2, err := page.DrawRubyTexts(texts2, 150, y, style, true)
	if err != nil {
		return err
	}
	page.DrawTextUTF8("いです。", 150+totalWidth2, y)

	// Save file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if err := doc.WriteTo(file); err != nil {
		return fmt.Errorf("failed to write PDF: %w", err)
	}

	fmt.Printf("  Created: %s\n", filename)
	return nil
}
