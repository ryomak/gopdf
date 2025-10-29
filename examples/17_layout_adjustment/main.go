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

	fmt.Println("=== 5. 実用例: 翻訳後の自動調整 ===")
	// 翻訳でテキストが長くなった場合のシミュレーション
	docLayout := &gopdf.PageLayout{
		Width:  595,
		Height: 842,
		TextBlocks: []gopdf.TextBlock{
			{
				Text: "Original short text",
				Rect: gopdf.Rectangle{X: 50, Y: 700, Width: 200, Height: 20},
			},
			{
				Text: "Next block",
				Rect: gopdf.Rectangle{X: 50, Y: 670, Width: 200, Height: 20},
			},
		},
	}

	fmt.Println("翻訳前:")
	fmt.Printf("  Block 0: Y=%.1f, Height=%.1f\n", docLayout.TextBlocks[0].Rect.Y, docLayout.TextBlocks[0].Rect.Height)
	fmt.Printf("  Block 1: Y=%.1f, Height=%.1f\n", docLayout.TextBlocks[1].Rect.Y, docLayout.TextBlocks[1].Rect.Height)

	// 翻訳で高さが増えたとする
	docLayout.TextBlocks[0].Text = "これは翻訳された長いテキストで、元のテキストよりもかなり長くなっています。"
	docLayout.ResizeBlock(gopdf.ContentBlockTypeText, 0, 400, 60) // 高さを20→60に

	// 重なりをチェック
	overlaps = docLayout.DetectOverlaps()
	if len(overlaps) > 0 {
		fmt.Println("\n翻訳後、重なりが検出されました！")
		// Block 1を下に移動して重なりを解消
		oldHeight := float64(20)
		newHeight := float64(60)
		offset := newHeight - oldHeight
		docLayout.MoveBlock(gopdf.ContentBlockTypeText, 1, 0, -offset-10) // 10pxのマージンも追加

		fmt.Println("Block 1を自動調整しました")
	}

	fmt.Println("\n調整後:")
	fmt.Printf("  Block 0: Y=%.1f, Height=%.1f\n", docLayout.TextBlocks[0].Rect.Y, docLayout.TextBlocks[0].Rect.Height)
	fmt.Printf("  Block 1: Y=%.1f, Height=%.1f\n", docLayout.TextBlocks[1].Rect.Y, docLayout.TextBlocks[1].Rect.Height)

	// 重なりが解消されたか確認
	overlaps = docLayout.DetectOverlaps()
	if len(overlaps) == 0 {
		fmt.Println("✓ 重なりが解消されました！")
	}
}
