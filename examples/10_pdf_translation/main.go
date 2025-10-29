// Example: 10_pdf_translation
// This example demonstrates how to extract page layout and translate PDF
// while preserving the original layout.
package main

import (
	"fmt"
	"os"

	"github.com/ryomak/gopdf"
)

func main() {
	// まず、英語のサンプルPDFを生成
	fmt.Println("Creating sample English PDF...")
	if err := createEnglishPDF("english.pdf"); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating English PDF: %v\n", err)
		os.Exit(1)
	}

	// レイアウト解析の例
	fmt.Println("\n--- Example 1: Layout Extraction ---")
	if err := demonstrateLayoutExtraction("english.pdf"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// 翻訳の例（簡易的な辞書翻訳）
	fmt.Println("\n--- Example 2: PDF Translation ---")
	if err := demonstrateTranslation("english.pdf", "japanese.pdf"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nAll examples completed successfully!")
}

// createEnglishPDF は英語のサンプルPDFを作成
func createEnglishPDF(filename string) error {
	doc := gopdf.New()

	// ページ1: タイトルと本文
	page1 := doc.AddPage(gopdf.PageSizeA4, gopdf.Portrait)

	// タイトル
	page1.SetFont(gopdf.FontHelveticaBold, 24)
	page1.DrawText("Technical Report", 50, 800)

	// サブタイトル
	page1.SetFont(gopdf.FontHelvetica, 14)
	page1.DrawText("Annual Performance Summary", 50, 770)

	// セクション1
	page1.SetFont(gopdf.FontHelveticaBold, 16)
	page1.DrawText("Introduction", 50, 720)

	page1.SetFont(gopdf.FontHelvetica, 12)
	page1.DrawText("This document provides a comprehensive", 50, 695)
	page1.DrawText("overview of the project performance", 50, 678)
	page1.DrawText("during the fiscal year.", 50, 661)

	// セクション2
	page1.SetFont(gopdf.FontHelveticaBold, 16)
	page1.DrawText("Key Findings", 50, 620)

	page1.SetFont(gopdf.FontHelvetica, 12)
	page1.DrawText("The analysis revealed significant", 50, 595)
	page1.DrawText("improvements in efficiency and", 50, 578)
	page1.DrawText("customer satisfaction metrics.", 50, 561)

	// セクション3
	page1.SetFont(gopdf.FontHelveticaBold, 16)
	page1.DrawText("Conclusion", 50, 520)

	page1.SetFont(gopdf.FontHelvetica, 12)
	page1.DrawText("We recommend continued investment", 50, 495)
	page1.DrawText("in infrastructure and training.", 50, 478)

	// ページ2: より複雑なレイアウト
	page2 := doc.AddPage(gopdf.PageSizeA4, gopdf.Portrait)

	page2.SetFont(gopdf.FontHelveticaBold, 20)
	page2.DrawText("Page 2 - Details", 50, 800)

	// 左カラム
	page2.SetFont(gopdf.FontHelveticaBold, 14)
	page2.DrawText("Left Column", 50, 750)

	page2.SetFont(gopdf.FontHelvetica, 11)
	page2.DrawText("First item", 50, 730)
	page2.DrawText("Second item", 50, 715)
	page2.DrawText("Third item", 50, 700)

	// 右カラム
	page2.SetFont(gopdf.FontHelveticaBold, 14)
	page2.DrawText("Right Column", 320, 750)

	page2.SetFont(gopdf.FontHelvetica, 11)
	page2.DrawText("Data point A", 320, 730)
	page2.DrawText("Data point B", 320, 715)
	page2.DrawText("Data point C", 320, 700)

	// ファイルに出力
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

// demonstrateLayoutExtraction はレイアウト抽出を実演
func demonstrateLayoutExtraction(filename string) error {
	reader, err := gopdf.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open PDF: %w", err)
	}
	defer reader.Close()

	// 全ページのレイアウトを抽出
	layouts, err := reader.ExtractAllLayouts()
	if err != nil {
		return fmt.Errorf("failed to extract layouts: %w", err)
	}

	// 各ページのレイアウト情報を表示
	for pageNum, layout := range layouts {
		fmt.Printf("\nPage %d:\n", pageNum+1)
		fmt.Printf("  Page size: %.1f x %.1f\n", layout.Width, layout.Height)
		fmt.Printf("  Text blocks: %d\n", len(layout.TextBlocks))
		fmt.Printf("  Images: %d\n", len(layout.Images))

		// テキストブロックの詳細
		for i, block := range layout.TextBlocks {
			fmt.Printf("\n  Text Block %d:\n", i+1)
			fmt.Printf("    Text: %s\n", block.Text)
			fmt.Printf("    Position: (%.1f, %.1f)\n", block.Bounds.X, block.Bounds.Y)
			fmt.Printf("    Size: %.1f x %.1f\n", block.Bounds.Width, block.Bounds.Height)
			fmt.Printf("    Font: %s, Size: %.1f\n", block.Font, block.FontSize)
		}

		// 画像の詳細
		for i, img := range layout.Images {
			fmt.Printf("\n  Image %d:\n", i+1)
			fmt.Printf("    Position: (%.1f, %.1f)\n", img.X, img.Y)
			fmt.Printf("    Placed size: %.1f x %.1f\n", img.PlacedWidth, img.PlacedHeight)
			fmt.Printf("    Original: %d x %d\n", img.Width, img.Height)
		}
	}

	return nil
}

// demonstrateTranslation は翻訳を実演
func demonstrateTranslation(inputPath string, outputPath string) error {
	fmt.Println("Translating PDF...")

	// デフォルト日本語フォント（Koruri）を使用
	// gopdf v2では日本語フォントが埋め込まれているため、手動でフォントをダウンロードする必要はありません
	jpFont, err := gopdf.DefaultJapaneseFont()
	useJapanese := false

	var targetFont interface{}
	var targetFontName string

	if err == nil {
		targetFont = jpFont
		targetFontName = "Koruri"
		useJapanese = true
		fmt.Println("  Using embedded Koruri font for Japanese text")
	} else {
		// フォント読み込みに失敗した場合は標準フォントを使用（英語のみ）
		targetFont = gopdf.FontHelvetica
		targetFontName = "Helvetica"
		fmt.Printf("  Warning: Failed to load Japanese font: %v\n", err)
		fmt.Println("  Using Helvetica (English only)")
	}

	// 翻訳辞書（useJapaneseがtrueの場合のみ日本語に翻訳）
	translationDict := map[string]string{
		"Technical Report":                      "Technical Report",
		"Annual Performance Summary":            "Annual Performance Summary",
		"Introduction":                          "Introduction",
		"This document provides a comprehensive": "This document provides a comprehensive",
		"overview of the project performance":   "overview of the project performance",
		"during the fiscal year.":               "during the fiscal year.",
		"Key Findings":                          "Key Findings",
		"The analysis revealed significant":     "The analysis revealed significant",
		"improvements in efficiency and":        "improvements in efficiency and",
		"customer satisfaction metrics.":        "customer satisfaction metrics.",
		"Conclusion":                            "Conclusion",
		"We recommend continued investment":     "We recommend continued investment",
		"in infrastructure and training.":       "in infrastructure and training.",
		"Page 2 - Details":                      "Page 2 - Details",
		"Left Column":                           "Left Column",
		"First item":                            "First item",
		"Second item":                           "Second item",
		"Third item":                            "Third item",
		"Right Column":                          "Right Column",
		"Data point A":                          "Data point A",
		"Data point B":                          "Data point B",
		"Data point C":                          "Data point C",
	}

	// 日本語フォントがある場合は日本語に翻訳
	if useJapanese {
		translationDict = map[string]string{
			"Technical Report":                      "技術レポート",
			"Annual Performance Summary":            "年次業績概要",
			"Introduction":                          "はじめに",
			"This document provides a comprehensive": "このドキュメントは包括的な",
			"overview of the project performance":   "プロジェクト業績の概要を",
			"during the fiscal year.":               "会計年度中に提供します。",
			"Key Findings":                          "主な発見",
			"The analysis revealed significant":     "分析により重要な",
			"improvements in efficiency and":        "効率性と",
			"customer satisfaction metrics.":        "顧客満足度の改善が明らかになりました。",
			"Conclusion":                            "結論",
			"We recommend continued investment":     "継続的な投資を",
			"in infrastructure and training.":       "インフラと研修に推奨します。",
			"Page 2 - Details":                      "ページ2 - 詳細",
			"Left Column":                           "左カラム",
			"First item":                            "最初の項目",
			"Second item":                           "2番目の項目",
			"Third item":                            "3番目の項目",
			"Right Column":                          "右カラム",
			"Data point A":                          "データポイントA",
			"Data point B":                          "データポイントB",
			"Data point C":                          "データポイントC",
		}
	}

	// Translatorインターフェースの実装
	translator := gopdf.TranslateFunc(func(text string) (string, error) {
		// 辞書に存在する場合は翻訳
		if translated, ok := translationDict[text]; ok {
			return translated, nil
		}
		// 存在しない場合はそのまま返す
		return text, nil
	})

	// 翻訳オプション
	opts := gopdf.PDFTranslatorOptions{
		Translator:     translator,
		TargetFont:     targetFont,
		TargetFontName: targetFontName,
		FittingOptions: gopdf.FitTextOptions{
			MaxFontSize: 24.0,
			MinFontSize: 8.0,
			LineSpacing: 1.2,
			Padding:     2.0,
			AllowShrink: true,
			AllowGrow:   false,
			Alignment:   gopdf.AlignLeft,
		},
		KeepImages: true,
		KeepLayout: true,
	}

	// PDF翻訳を実行
	err = gopdf.TranslatePDF(inputPath, outputPath, opts)
	if err != nil {
		return fmt.Errorf("translation failed: %w", err)
	}

	fmt.Printf("  Translated PDF saved: %s\n", outputPath)
	if useJapanese {
		fmt.Println("  ✓ Japanese translation completed")
	} else {
		fmt.Println("  ✓ Layout preserved (English text retained)")
	}
	return nil
}
