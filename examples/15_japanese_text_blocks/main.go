// Example: 15_japanese_text_blocks
// This example tests Japanese and English text extraction with TextBlocks.
package main

import (
	"fmt"
	"os"

	"github.com/ryomak/gopdf"
)

func main() {
	fmt.Println("=== Creating PDF with Japanese and English text ===")

	// PDFを作成
	doc := gopdf.New()
	page := doc.AddPage(gopdf.PageSizeA4, gopdf.Portrait)

	// TTFフォントをロード（日本語対応）
	// システムにインストールされているフォントを探す
	fontPaths := []string{
		"/System/Library/Fonts/ヒラギノ角ゴシック W3.ttc",
		"/System/Library/Fonts/Hiragino Sans GB.ttc",
		"/usr/share/fonts/truetype/takao-gothic/TakaoPGothic.ttf",
		"/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc",
		os.Getenv("HOME") + "/Library/Fonts/NotoSansJP-Regular.ttf",
	}

	var font *gopdf.TTFFont
	var err error
	for _, path := range fontPaths {
		if _, statErr := os.Stat(path); statErr == nil {
			font, err = gopdf.LoadTTF(path)
			if err == nil {
				fmt.Printf("Using font: %s\n", path)
				break
			}
		}
	}

	if font == nil {
		fmt.Println("Warning: No Japanese font found, using standard font")
		// 標準フォントを使用（日本語は正しく表示されない）
		page.SetFont(gopdf.FontHelvetica, 12)
		page.DrawText("English: Hello World", 50, 800)
		page.DrawText("Japanese: Cannot display correctly", 50, 780)
	} else {
		// TTFフォントを使用
		page.SetTTFFont(font, 12)

		// 英語と日本語を混在させたテキスト
		page.DrawTextUTF8("English: Hello World", 50, 800)
		page.DrawTextUTF8("Japanese: こんにちは世界", 50, 780)
		page.DrawTextUTF8("Mixed: Hello こんにちは World 世界", 50, 760)

		// 段落のようなテキスト
		page.DrawTextUTF8("第1段落：これは日本語のテキストです。", 50, 720)
		page.DrawTextUTF8("This is English text in paragraph 1.", 50, 700)

		page.DrawTextUTF8("第2段落：PDFからテキストを抽出します。", 50, 660)
		page.DrawTextUTF8("Paragraph 2: Extract text from PDF.", 50, 640)

		// 箇条書き風
		page.DrawTextUTF8("リスト項目：", 50, 600)
		page.DrawTextUTF8("• 項目1", 70, 580)
		page.DrawTextUTF8("• Item 2", 70, 560)
		page.DrawTextUTF8("• アイテム3", 70, 540)
	}

	// PDFをファイルに保存
	filename := "japanese_text_test.pdf"
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	if err := doc.WriteTo(file); err != nil {
		fmt.Printf("Error writing PDF: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created: %s\n\n", filename)

	// PDFを読み込んで確認
	fmt.Println("=== Reading PDF and extracting text ===")
	reader, err := gopdf.Open(filename)
	if err != nil {
		fmt.Printf("Error opening PDF: %v\n", err)
		os.Exit(1)
	}
	defer reader.Close()

	// TextElementsを抽出（細かい単位）
	fmt.Println("\n--- TextElements (raw) ---")
	elements, err := reader.ExtractPageTextElements(0)
	if err != nil {
		fmt.Printf("Error extracting text elements: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Total elements: %d\n", len(elements))
	for i, elem := range elements {
		if i < 10 { // 最初の10個だけ表示
			fmt.Printf("[%d] Text: %q\n", i, elem.Text)
			fmt.Printf("     Position: (%.1f, %.1f), Size: %.1f\n", elem.X, elem.Y, elem.Size)
		}
	}
	if len(elements) > 10 {
		fmt.Printf("... and %d more elements\n", len(elements)-10)
	}

	// TextBlocksを抽出（グルーピング済み）
	fmt.Println("\n--- TextBlocks (grouped) ---")
	blocks, err := reader.ExtractPageTextBlocks(0)
	if err != nil {
		fmt.Printf("Error extracting text blocks: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Total blocks: %d\n\n", len(blocks))
	for i, block := range blocks {
		fmt.Printf("=== Block %d ===\n", i+1)
		fmt.Printf("Elements: %d\n", len(block.Elements))
		fmt.Printf("Bounds: (%.1f, %.1f) - %.1fx%.1f\n",
			block.Rect.X, block.Rect.Y,
			block.Rect.Width, block.Rect.Height)
		fmt.Printf("Text:\n%s\n\n", block.Text)

		// 文字化けチェック
		if containsGarbled(block.Text) {
			fmt.Printf("⚠️  WARNING: Block %d may contain garbled text!\n\n", i+1)
		}
	}

	// 全体のテキストを結合して表示
	fmt.Println("\n--- Combined Text from all blocks ---")
	for i, block := range blocks {
		if i > 0 {
			fmt.Println()
		}
		fmt.Print(block.Text)
	}
	fmt.Println()

	fmt.Println("\n✓ Test completed!")
}

// containsGarbled checks if text contains garbled characters
func containsGarbled(text string) bool {
	// 一般的な文字化けパターンをチェック
	garbledPatterns := []string{
		"Ã",    // UTF-8の二重エンコード
		"â€",   // よくある文字化けパターン
		"ï¿½", // REPLACEMENT CHARACTER
	}

	for _, pattern := range garbledPatterns {
		for i := 0; i < len(text); i++ {
			if i+len(pattern) <= len(text) && text[i:i+len(pattern)] == pattern {
				return true
			}
		}
	}

	return false
}
