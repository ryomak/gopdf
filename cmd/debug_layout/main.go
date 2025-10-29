package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ryomak/gopdf"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run cmd/debug_layout/main.go <pdf_file>")
		os.Exit(1)
	}

	inputPath := os.Args[1]

	// PDFを開く
	reader, err := gopdf.Open(inputPath)
	if err != nil {
		log.Fatalf("Failed to open PDF: %v", err)
	}
	defer reader.Close()

	fmt.Printf("=== PDF Analysis: %s ===\n", inputPath)
	fmt.Printf("Page count: %d\n\n", reader.PageCount())

	// 各ページを分析
	for i := 0; i < reader.PageCount(); i++ {
		fmt.Printf("--- Page %d ---\n", i+1)

		layout, err := reader.ExtractPageLayout(i)
		if err != nil {
			log.Printf("Failed to extract layout from page %d: %v", i, err)
			continue
		}

		fmt.Printf("Page size: %.2f x %.2f\n", layout.Width, layout.Height)
		fmt.Printf("Text blocks: %d\n", len(layout.TextBlocks))
		fmt.Printf("Images: %d\n\n", len(layout.Images))

		// テキストブロックの詳細
		for j, block := range layout.TextBlocks {
			// テキストを最初の30文字だけ表示
			text := block.Text
			if len(text) > 30 {
				text = text[:30] + "..."
			}

			fmt.Printf("  Block %d:\n", j)
			fmt.Printf("    Text: %q\n", text)
			fmt.Printf("    Position: X=%.2f, Y=%.2f (bottom-left)\n", block.Rect.X, block.Rect.Y)
			fmt.Printf("    Size: Width=%.2f, Height=%.2f\n", block.Rect.Width, block.Rect.Height)
			fmt.Printf("    Top Y: %.2f\n", block.Rect.Y+block.Rect.Height)
			fmt.Printf("    Bottom Y: %.2f\n", block.Rect.Y)
			fmt.Printf("    Font: %s, Size: %.2f\n\n", block.Font, block.FontSize)
		}

		// 画像の詳細
		for j, img := range layout.Images {
			fmt.Printf("  Image %d:\n", j)
			fmt.Printf("    Position: X=%.2f, Y=%.2f (bottom-left)\n", img.X, img.Y)
			fmt.Printf("    Size: %.2f x %.2f\n", img.PlacedWidth, img.PlacedHeight)
			fmt.Printf("    Top Y: %.2f\n", img.Y+img.PlacedHeight)
			fmt.Printf("    Bottom Y: %.2f\n", img.Y)
			fmt.Printf("    Format: %s\n\n", img.Format)
		}

		// ContentBlocksで統合された順序を確認
		fmt.Println("  Content blocks (sorted order):")
		contentBlocks, err := reader.ExtractPageContentBlocks(i)
		if err == nil {
			for k, cb := range contentBlocks {
				switch cb.Type() {
				case gopdf.ContentBlockTypeText:
					if tb, ok := cb.(gopdf.TextBlock); ok {
						text := tb.Text
						if len(text) > 20 {
							text = text[:20] + "..."
						}
						fmt.Printf("    %d. TEXT: %q at Y=%.2f (top: %.2f)\n",
							k, text, tb.Rect.Y,
							tb.Rect.Y+tb.Rect.Height)
					}
				case gopdf.ContentBlockTypeImage:
					if ib, ok := cb.(gopdf.ImageBlock); ok {
						fmt.Printf("    %d. IMAGE at Y=%.2f (top: %.2f)\n",
							k, ib.Y,
							ib.Y+ib.PlacedHeight)
					}
				}
			}
		}

		fmt.Println()
	}
}
