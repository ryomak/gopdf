package main

import (
	"fmt"
	"log"

	"github.com/ryomak/gopdf"
)

func main() {
	// ContentBlockインターフェースの使用例
	fmt.Println("=== ContentBlock Interface Example ===\n")

	// 1. TextBlockとImageBlockを作成
	textBlock := gopdf.TextBlock{
		Text: "Hello, World!",
		Rect: gopdf.Rectangle{
			X:      100,
			Y:      700,
			Width:  200,
			Height: 50,
		},
		Font:     "Helvetica",
		FontSize: 12,
	}

	imageBlock := gopdf.ImageBlock{
		X:            100,
		Y:            500,
		PlacedWidth:  300,
		PlacedHeight: 200,
	}

	// 2. ContentBlockインターフェースとして扱う
	var blocks []gopdf.ContentBlock
	blocks = append(blocks, textBlock)
	blocks = append(blocks, imageBlock)

	// 3. 統一的に処理
	fmt.Println("All Content Blocks:")
	for i, block := range blocks {
		x, y := block.GetPosition()
		bounds := block.GetBounds()
		fmt.Printf("Block %d:\n", i+1)
		fmt.Printf("  Type: %s\n", block.GetType())
		fmt.Printf("  Position: (%.1f, %.1f)\n", x, y)
		fmt.Printf("  Size: %.1f x %.1f\n", bounds.Width, bounds.Height)
		fmt.Println()
	}

	// 4. PageLayoutの使用例
	layout := &gopdf.PageLayout{
		PageNum: 0,
		Width:   595,
		Height:  842,
		TextBlocks: []gopdf.TextBlock{
			{
				Text: "First Block",
				Rect: gopdf.Rectangle{X: 100, Y: 700, Width: 200, Height: 50},
			},
			{
				Text: "Second Block",
				Rect: gopdf.Rectangle{X: 100, Y: 600, Width: 200, Height: 50},
			},
		},
		Images: []gopdf.ImageBlock{
			{
				X:            100,
				Y:            500,
				PlacedWidth:  300,
				PlacedHeight: 200,
			},
		},
	}

	// 5. ContentBlocksを取得してソート
	sortedBlocks := layout.SortedContentBlocks()
	fmt.Println("Sorted Content Blocks (top to bottom, left to right):")
	for i, block := range sortedBlocks {
		x, y := block.GetPosition()
		fmt.Printf("%d. Type=%s Position=(%.0f, %.0f)\n", i+1, block.GetType(), x, y)
	}

	fmt.Println("\n=== DefaultPDFTranslatorOptions Example ===\n")

	// 6. DefaultPDFTranslatorOptionsの使用例
	opts := gopdf.DefaultPDFTranslatorOptions(
		gopdf.FontHelvetica,
		"Helvetica",
	)

	fmt.Printf("Default Translator Options:\n")
	fmt.Printf("  TargetFontName: %s\n", opts.TargetFontName)
	fmt.Printf("  KeepImages: %v\n", opts.KeepImages)
	fmt.Printf("  KeepLayout: %v\n", opts.KeepLayout)
	fmt.Printf("  FittingOptions.MaxFontSize: %.1f\n", opts.FittingOptions.MaxFontSize)
	fmt.Printf("  FittingOptions.MinFontSize: %.1f\n", opts.FittingOptions.MinFontSize)
	fmt.Printf("  FittingOptions.AllowShrink: %v\n", opts.FittingOptions.AllowShrink)

	// 7. オプションをカスタマイズ
	opts.FittingOptions.MinFontSize = 8.0
	opts.FittingOptions.MaxFontSize = 20.0
	fmt.Printf("\nCustomized MinFontSize: %.1f\n", opts.FittingOptions.MinFontSize)

	// 実際の翻訳を行う場合:
	// opts.Translator = gopdf.TranslateFunc(func(text string) (string, error) {
	//     // 翻訳ロジック
	//     return "Translated: " + text, nil
	// })
	// err := gopdf.TranslatePDF("input.pdf", "output.pdf", opts)

	log.Println("\n✓ All examples completed successfully!")
}
