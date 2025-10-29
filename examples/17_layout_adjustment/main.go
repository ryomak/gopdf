package main

import (
	"fmt"
	"log"

	"github.com/ryomak/gopdf"
)

func main() {
	// 既存のPDFを読み込む例
	// reader, err := gopdf.Open("input.pdf")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// layout, err := reader.ExtractPageLayout(0)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// または、新規でPageLayoutを作成
	layout := &gopdf.PageLayout{
		Width:  595,  // A4幅
		Height: 842,  // A4高さ
		TextBlocks: []gopdf.TextBlock{
			{
				Text:     "タイトル",
				Font:     "Helvetica-Bold",
				FontSize: 24,
				Rect:     gopdf.Rectangle{X: 50, Y: 750, Width: 495, Height: 30},
			},
			{
				Text:     "本文の最初の段落です。この文章は少し長めに書かれています。",
				Font:     "Helvetica",
				FontSize: 12,
				Rect:     gopdf.Rectangle{X: 50, Y: 700, Width: 495, Height: 40},
			},
			{
				Text:     "2番目の段落です。",
				Font:     "Helvetica",
				FontSize: 12,
				Rect:     gopdf.Rectangle{X: 50, Y: 650, Width: 495, Height: 20},
			},
		},
		Images: []gopdf.ImageBlock{
			{
				X:            50,
				Y:            500,
				PlacedWidth:  200,
				PlacedHeight: 150,
			},
		},
	}

	fmt.Println("=== 1. ブロックの移動 (MoveBlock) ===")
	// TextBlock[1]を右に20px、下に30px移動
	err := layout.MoveBlock(gopdf.ContentBlockTypeText, 1, 20, -30)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("TextBlock[1]移動後: X=%.1f, Y=%.1f\n",
		layout.TextBlocks[1].Rect.X, layout.TextBlocks[1].Rect.Y)

	// ImageBlock[0]を右に50px移動
	err = layout.MoveBlock(gopdf.ContentBlockTypeImage, 0, 50, 0)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("ImageBlock[0]移動後: X=%.1f, Y=%.1f\n\n",
		layout.Images[0].X, layout.Images[0].Y)

	fmt.Println("=== 2. ブロックのリサイズ (ResizeBlock) ===")
	// TextBlock[0]の幅を変更
	err = layout.ResizeBlock(gopdf.ContentBlockTypeText, 0, 400, 35)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("TextBlock[0]リサイズ後: Width=%.1f, Height=%.1f\n",
		layout.TextBlocks[0].Rect.Width, layout.TextBlocks[0].Rect.Height)

	// ImageBlock[0]のサイズを変更
	err = layout.ResizeBlock(gopdf.ContentBlockTypeImage, 0, 250, 180)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("ImageBlock[0]リサイズ後: Width=%.1f, Height=%.1f\n\n",
		layout.Images[0].PlacedWidth, layout.Images[0].PlacedHeight)

	fmt.Println("=== 3. 重なり検出 (DetectOverlaps) ===")
	// わざと重なりを作る
	testLayout := &gopdf.PageLayout{
		Width:  595,
		Height: 842,
		TextBlocks: []gopdf.TextBlock{
			{
				Text: "Block A",
				Rect: gopdf.Rectangle{X: 100, Y: 100, Width: 200, Height: 50},
			},
			{
				Text: "Block B (重なる)",
				Rect: gopdf.Rectangle{X: 150, Y: 120, Width: 200, Height: 50},
			},
			{
				Text: "Block C (重ならない)",
				Rect: gopdf.Rectangle{X: 100, Y: 200, Width: 200, Height: 50},
			},
		},
	}

	overlaps := testLayout.DetectOverlaps()
	if len(overlaps) > 0 {
		fmt.Printf("重なりが%d件見つかりました:\n", len(overlaps))
		for i, overlap := range overlaps {
			fmt.Printf("  %d. 重なり面積: %.1f\n", i+1, overlap.Area)
			fmt.Printf("     Block1: %v\n", overlap.Block1.(gopdf.TextBlock).Text)
			fmt.Printf("     Block2: %v\n", overlap.Block2.(gopdf.TextBlock).Text)
		}
	} else {
		fmt.Println("重なりは検出されませんでした")
	}
	fmt.Println()

	fmt.Println("=== 4. ページ分割 (SplitIntoPages) ===")
	// 高さの大きいブロックを作成
	tallLayout := &gopdf.PageLayout{
		Width:  595,
		Height: 842,
		TextBlocks: []gopdf.TextBlock{
			{
				Text: "Block 1 (高さ300)",
				Rect: gopdf.Rectangle{X: 50, Y: 800, Width: 495, Height: 300},
			},
			{
				Text: "Block 2 (高さ300)",
				Rect: gopdf.Rectangle{X: 50, Y: 450, Width: 495, Height: 300},
			},
			{
				Text: "Block 3 (高さ300)",
				Rect: gopdf.Rectangle{X: 50, Y: 100, Width: 495, Height: 300},
			},
		},
	}

	// ページ高さ500、最小スペーシング10、マージン20で分割
	pages, err := tallLayout.SplitIntoPages(500, 10, 20)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%dページに分割されました:\n", len(pages))
	for i, page := range pages {
		fmt.Printf("  Page %d: TextBlocks=%d, Images=%d\n",
			i+1, len(page.TextBlocks), len(page.Images))
		for j, tb := range page.TextBlocks {
			fmt.Printf("    TextBlock[%d]: Y=%.1f, Height=%.1f\n",
				j, tb.Rect.Y, tb.Rect.Height)
		}
	}
	fmt.Println()

	fmt.Println("=== 5. 自動調整 (AdjustLayout with StrategyFlowDown) ===")
	// 翻訳や編集で間隔が狭くなった場合の自動調整
	autoLayout := &gopdf.PageLayout{
		Width:  595,
		Height: 842,
		TextBlocks: []gopdf.TextBlock{
			{
				Text: "タイトル",
				Rect: gopdf.Rectangle{X: 50, Y: 680, Width: 400, Height: 20},
			},
			{
				Text: "本文1: これは最初の段落です。",
				Rect: gopdf.Rectangle{X: 50, Y: 665, Width: 400, Height: 20}, // 間隔5px
			},
			{
				Text: "本文2: これは2番目の段落です。",
				Rect: gopdf.Rectangle{X: 50, Y: 650, Width: 400, Height: 20}, // 間隔5px
			},
		},
	}

	fmt.Println("自動調整前:")
	for i, tb := range autoLayout.TextBlocks {
		fmt.Printf("  Block %d: Y=%.1f (間隔が狭い)\n", i, tb.Rect.Y)
	}

	// StrategyFlowDown で自動調整（後続ブロックを自動的にずらす）
	opts := gopdf.LayoutAdjustmentOptions{
		Strategy:   gopdf.StrategyFlowDown,
		MinSpacing: 10, // 最小間隔10px
	}
	err = autoLayout.AdjustLayout(opts)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n自動調整後 (最小間隔10pxを確保):")
	for i, tb := range autoLayout.TextBlocks {
		fmt.Printf("  Block %d: Y=%.1f\n", i, tb.Rect.Y)
	}

	// 間隔を確認
	for i := 0; i < len(autoLayout.TextBlocks)-1; i++ {
		bottom := autoLayout.TextBlocks[i].Rect.Y
		top := autoLayout.TextBlocks[i+1].Rect.Y + autoLayout.TextBlocks[i+1].Rect.Height
		spacing := bottom - top
		fmt.Printf("  Block %d ↔ %d の間隔: %.1f px\n", i, i+1, spacing)
	}

	fmt.Println("\n=== 6. 実用例: ResizeBlock + AdjustLayout ===")
	// テキストを編集してサイズが変わった場合
	editLayout := &gopdf.PageLayout{
		Width:  595,
		Height: 842,
		TextBlocks: []gopdf.TextBlock{
			{
				Text: "Short",
				Rect: gopdf.Rectangle{X: 50, Y: 680, Width: 200, Height: 20},
			},
			{
				Text: "Next",
				Rect: gopdf.Rectangle{X: 50, Y: 650, Width: 200, Height: 20},
			},
			{
				Text: "Third",
				Rect: gopdf.Rectangle{X: 50, Y: 620, Width: 200, Height: 20},
			},
		},
	}

	fmt.Println("編集前:")
	fmt.Printf("  Block 0: Height=%.1f\n", editLayout.TextBlocks[0].Rect.Height)

	// テキストを編集して高さを変更
	editLayout.TextBlocks[0].Text = "This is now a much longer text that requires more space"
	editLayout.ResizeBlock(gopdf.ContentBlockTypeText, 0, 400, 60)

	fmt.Println("\nResizeBlock後:")
	fmt.Printf("  Block 0: Height=%.1f (20→60に変更)\n", editLayout.TextBlocks[0].Rect.Height)

	// 自動調整で後続ブロックをずらす
	err = editLayout.AdjustLayout(gopdf.LayoutAdjustmentOptions{
		Strategy:   gopdf.StrategyFlowDown,
		MinSpacing: 10,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nAdjustLayout後 (後続ブロックが自動調整):")
	for i, tb := range editLayout.TextBlocks {
		fmt.Printf("  Block %d: Y=%.1f\n", i, tb.Rect.Y)
	}

	// 重なりがないことを確認
	overlaps = editLayout.DetectOverlaps()
	if len(overlaps) == 0 {
		fmt.Println("✓ 重なりなし！")
	} else {
		fmt.Printf("✗ 重なりが%d件あります\n", len(overlaps))
	}

	fmt.Println("\n=== 7. StrategyFitContent: ブロックサイズを変えずに内容を縮小 ===")
	// ブロックサイズは固定で、内容をブロックに収める
	fitLayout := &gopdf.PageLayout{
		Width:  595,
		Height: 842,
		TextBlocks: []gopdf.TextBlock{
			{
				Text:     "This is a very long text that will not fit in the small block at the current font size",
				Font:     "Helvetica",
				FontSize: 20, // 大きすぎるフォントサイズ
				Rect:     gopdf.Rectangle{X: 50, Y: 700, Width: 150, Height: 60}, // 小さいブロック
			},
			{
				Text:     "Normal text that fits",
				Font:     "Helvetica",
				FontSize: 12,
				Rect:     gopdf.Rectangle{X: 50, Y: 630, Width: 300, Height: 30},
			},
			{
				Text:     "Another text block",
				Font:     "Helvetica",
				FontSize: 14,
				Rect:     gopdf.Rectangle{X: 50, Y: 590, Width: 200, Height: 25},
			},
		},
	}

	fmt.Println("調整前:")
	for i, tb := range fitLayout.TextBlocks {
		fmt.Printf("  Block %d: FontSize=%.1f, Rect=%.0fx%.0f\n", i, tb.FontSize, tb.Rect.Width, tb.Rect.Height)
	}
	fmt.Printf("  Block 0 のテキスト: \"%s...\"\n", fitLayout.TextBlocks[0].Text[:40])

	// StrategyFitContent で調整（ブロックサイズは変えず、フォントサイズを調整）
	err = fitLayout.AdjustLayout(gopdf.LayoutAdjustmentOptions{
		Strategy: gopdf.StrategyFitContent,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n調整後（ブロックサイズは変わらず、フォントサイズのみ調整）:")
	for i, tb := range fitLayout.TextBlocks {
		fmt.Printf("  Block %d: FontSize=%.1f, Rect=%.0fx%.0f\n", i, tb.FontSize, tb.Rect.Width, tb.Rect.Height)
		if i == 0 {
			fmt.Printf("    → フォントサイズが %.1f から %.1f に縮小されました\n", 20.0, tb.FontSize)
		}
	}
	fmt.Println("\n  ✓ Block 0: 大きすぎたテキストがブロックに収まるように縮小")
	fmt.Println("  ✓ Block 1, 2: すでに収まっているのでフォントサイズは変更なし")
	fmt.Println("  ✓ すべてのブロックの位置とサイズは変わらない")
}
