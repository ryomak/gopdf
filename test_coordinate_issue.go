package gopdf

import (
	"fmt"
	"os"
)

// TestCoordinateIssue は座標反転問題を調査するためのテスト関数
func TestCoordinateIssue() error {
	// 1. サンプルPDFを作成（3つのテキストブロック：上、中、下）
	doc := New()
	page := doc.AddPage(PageSizeA4, Portrait)

	// フォント設定
	if err := page.SetFont(FontHelvetica, 12); err != nil {
		return err
	}

	// テキストを配置（Y座標が大きい = 上）
	if err := page.DrawText("TOP TEXT (Y=750)", 100, 750); err != nil {
		return err
	}
	if err := page.DrawText("MIDDLE TEXT (Y=400)", 100, 400); err != nil {
		return err
	}
	if err := page.DrawText("BOTTOM TEXT (Y=100)", 100, 100); err != nil {
		return err
	}

	// 一時ファイルに保存
	tmpFile := "test_coordinate_temp.pdf"
	file, err := os.Create(tmpFile)
	if err != nil {
		return err
	}
	if err := doc.WriteTo(file); err != nil {
		file.Close()
		return err
	}
	file.Close()

	// 2. PDFを読み込んでレイアウトを抽出
	reader, err := Open(tmpFile)
	if err != nil {
		return err
	}
	defer reader.Close()
	defer os.Remove(tmpFile)

	layout, err := reader.ExtractPageLayout(0)
	if err != nil {
		return err
	}

	// 3. ブロックの順序を確認
	fmt.Println("=== Extracted TextBlocks ===")
	for i, block := range layout.TextBlocks {
		fmt.Printf("Block %d: Text=%q, Rect.Y=%.2f, Height=%.2f, Top=%.2f\n",
			i, block.Text, block.Rect.Y, block.Rect.Height,
			block.Rect.Y+block.Rect.Height)
	}

	// 4. 新しいPDFを生成
	opts := PDFTranslatorOptions{
		TargetFont:     FontHelvetica,
		TargetFontName: "F1",
		FittingOptions: DefaultFitTextOptions(),
		KeepImages:     false,
		KeepLayout:     true,
	}

	outputDoc := New()
	_, err = RenderLayout(outputDoc, layout, opts)
	if err != nil {
		return err
	}

	// 出力
	outputFile := "test_coordinate_output.pdf"
	out, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer out.Close()

	if err := outputDoc.WriteTo(out); err != nil {
		return err
	}

	fmt.Printf("\nOutput saved to: %s\n", outputFile)
	fmt.Println("Please check if the order is correct (TOP -> MIDDLE -> BOTTOM)")

	return nil
}
