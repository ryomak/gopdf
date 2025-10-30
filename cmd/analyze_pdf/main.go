package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ryomak/gopdf"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run cmd/analyze_pdf/main.go <pdf_file>")
		os.Exit(1)
	}

	inputPath := os.Args[1]

	reader, err := gopdf.Open(inputPath)
	if err != nil {
		log.Fatalf("Failed to open PDF: %v", err)
	}
	defer reader.Close()

	fmt.Printf("=== PDF Analysis: %s ===\n\n", inputPath)
	fmt.Printf("Page count: %d\n\n", reader.PageCount())

	// 最初のページを詳細分析
	pageNum := 0
	layout, err := reader.ExtractPageLayout(pageNum)
	if err != nil {
		log.Fatalf("Failed to extract layout: %v", err)
	}

	fmt.Printf("--- Page %d ---\n", pageNum+1)
	fmt.Printf("Page size: %.2f x %.2f\n\n", layout.Width, layout.Height)

	// 全てのテキストブロックを表示（ソート前）
	fmt.Println("=== All Text Blocks (original order) ===")
	for i, block := range layout.TextBlocks {
		fmt.Printf("\n--- ブロック %d ---\n", i+1)
		fmt.Printf("Text: %q\n", block.Text)
		fmt.Printf("Position: X=%.1f, Y=%.1f (bottom-left)\n", block.Rect.X, block.Rect.Y)
		fmt.Printf("Size: Width=%.1f, Height=%.1f\n", block.Rect.Width, block.Rect.Height)
		fmt.Printf("Top Y: %.1f (top edge)\n", block.Rect.Y+block.Rect.Height)
		fmt.Printf("Bottom Y: %.1f (bottom edge)\n", block.Rect.Y)
		fmt.Printf("Font: %s, Size: %.1f\n", block.Font, block.FontSize)

		// 視覚的な位置を判定
		relativePos := block.Rect.Y / layout.Height
		var visualPos string
		if relativePos > 0.66 {
			visualPos = "上部"
		} else if relativePos > 0.33 {
			visualPos = "中央"
		} else {
			visualPos = "下部"
		}
		fmt.Printf("視覚的位置: %s (Y座標/ページ高さ = %.2f%%)\n", visualPos, relativePos*100)

		// 最初の3要素の座標を表示
		if len(block.Elements) > 0 {
			fmt.Printf("Elements: %d個\n", len(block.Elements))
			for j := 0; j < len(block.Elements) && j < 3; j++ {
				elem := block.Elements[j]
				fmt.Printf("  Element %d: text=%q pos=(%.1f, %.1f) size=%.1f\n",
					j+1, elem.Text, elem.X, elem.Y, elem.Size)
			}
			if len(block.Elements) > 3 {
				fmt.Printf("  ... and %d more elements\n", len(block.Elements)-3)
			}
		}
	}

	// ContentBlocksでソートされた順序
	fmt.Println("\n\n=== Content Blocks (sorted, top-to-bottom) ===")
	contentBlocks, err := reader.ExtractPageContentBlocks(pageNum)
	if err != nil {
		log.Fatalf("Failed to extract content blocks: %v", err)
	}

	for i, cb := range contentBlocks {
		fmt.Printf("\n%d. ", i+1)
		switch cb.Type() {
		case gopdf.ContentBlockTypeText:
			if tb, ok := cb.(gopdf.TextBlock); ok {
				text := tb.Text
				if len(text) > 30 {
					text = text[:30] + "..."
				}
				fmt.Printf("TEXT: %q\n", text)
				fmt.Printf("   Y: %.1f (bottom) to %.1f (top)\n",
					tb.Rect.Y, tb.Rect.Y+tb.Rect.Height)
			}
		case gopdf.ContentBlockTypeImage:
			if ib, ok := cb.(gopdf.ImageBlock); ok {
				fmt.Printf("IMAGE: %dx%d\n", ib.Width, ib.Height)
				fmt.Printf("   Y: %.1f (bottom) to %.1f (top)\n",
					ib.Y, ib.Y+ib.PlacedHeight)
			}
		}
	}

	fmt.Println("\n\n=== Summary ===")
	fmt.Printf("Total text blocks: %d\n", len(layout.TextBlocks))
	fmt.Printf("Total images: %d\n", len(layout.Images))
	fmt.Printf("Sorted content blocks: %d\n", len(contentBlocks))
}
