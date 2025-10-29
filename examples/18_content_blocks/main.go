// Example: 18_content_blocks
// This example demonstrates extracting unified content blocks (text + images)
package main

import (
	"fmt"
	"os"

	"github.com/ryomak/gopdf"
)

func main() {
	fmt.Println("=== Creating PDF with text and images ===")

	// PDFを作成
	doc := gopdf.New()
	page := doc.AddPage(gopdf.PageSizeA4, gopdf.Portrait)

	page.SetFont(gopdf.FontHelvetica, 12)

	// テキスト1
	page.DrawText("This is the first text block", 50, 800)
	page.DrawText("with multiple lines", 50, 785)

	// テキスト2
	page.DrawText("This is the second text block", 50, 750)

	// 画像を配置（テキストの間）
	// 実際の画像を使わずに、矩形で代用
	page.SetFillColor(gopdf.NewRGB(200, 200, 200))
	page.FillRectangle(50, 700, 100, 40) // Y=700-740に画像（矩形）

	// テキスト3（画像の下）
	page.DrawText("This is text after the image", 50, 680)
	page.DrawText("also with multiple lines", 50, 665)

	// テキスト4
	page.DrawText("Final text block", 50, 630)

	// PDFをファイルに保存
	filename := "content_blocks_test.pdf"
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
	fmt.Println("=== Reading PDF and extracting content blocks ===")
	reader, err := gopdf.Open(filename)
	if err != nil {
		fmt.Printf("Error opening PDF: %v\n", err)
		os.Exit(1)
	}
	defer reader.Close()

	// 従来のテキストブロックのみ抽出
	fmt.Println("\n--- TextBlocks only (old API) ---")
	textBlocks, err := reader.ExtractPageTextBlocks(0)
	if err != nil {
		fmt.Printf("Error extracting text blocks: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Total text blocks: %d\n\n", len(textBlocks))
	for i, block := range textBlocks {
		fmt.Printf("TextBlock %d:\n", i+1)
		fmt.Printf("  Position: (%.1f, %.1f)\n", block.Rect.X, block.Rect.Y)
		fmt.Printf("  Text: %s\n\n", block.Text)
	}

	// 新しいAPI: テキストと画像の統合コンテンツブロック
	fmt.Println("\n--- ContentBlocks (text + images, new API) ---")
	contentBlocks, err := reader.ExtractPageContentBlocks(0)
	if err != nil {
		fmt.Printf("Error extracting content blocks: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Total content blocks: %d\n\n", len(contentBlocks))
	for i, block := range contentBlocks {
		bounds := block.Bounds()
		fmt.Printf("Block %d [%s]:\n", i+1, block.Type())
		fmt.Printf("  Position: (%.1f, %.1f)\n", bounds.X, bounds.Y)
		fmt.Printf("  Size: %.1fx%.1f\n", bounds.Width, bounds.Height)

		switch block.Type() {
		case gopdf.ContentBlockTypeText:
			tb := block.(gopdf.TextBlock)
			fmt.Printf("  Text: %s\n", tb.Text)
		case gopdf.ContentBlockTypeImage:
			ib := block.(gopdf.ImageBlock)
			fmt.Printf("  Image: %dx%d\n", ib.Width, ib.Height)
		}
		fmt.Println()
	}

	fmt.Println("✓ Test completed!")
}
