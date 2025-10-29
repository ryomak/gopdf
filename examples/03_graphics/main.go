// Example: 03_graphics
// This example demonstrates how to draw graphics (lines, rectangles, circles) in a PDF.
package main

import (
	"fmt"
	"os"

	"github.com/ryomak/gopdf"
)

func main() {
	// 新規PDFドキュメントを作成
	doc := gopdf.New()

	// A4サイズの縦向きページを追加
	page := doc.AddPage(gopdf.A4, gopdf.Portrait)

	// === タイトル ===
	page.SetFont(gopdf.HelveticaBold, 24)
	page.DrawText("Graphics Demo", 50, 800)

	// === 線の描画 ===
	page.SetFont(gopdf.Helvetica, 14)
	page.DrawText("Lines:", 50, 770)

	// 基本的な線
	page.SetLineWidth(1.0)
	page.SetStrokeColor(gopdf.Black)
	page.DrawLine(50, 750, 200, 750)

	// 太い赤い線
	page.SetLineWidth(3.0)
	page.SetStrokeColor(gopdf.Red)
	page.DrawLine(50, 740, 200, 740)

	// 青い線（異なるスタイル）
	page.SetLineWidth(2.0)
	page.SetStrokeColor(gopdf.Blue)
	page.SetLineCap(gopdf.RoundCap)
	page.DrawLine(50, 730, 200, 730)

	// === 矩形の描画 ===
	page.SetFont(gopdf.Helvetica, 14)
	page.SetStrokeColor(gopdf.Black)
	page.DrawText("Rectangles:", 50, 700)

	// 枠線のみの矩形
	page.SetLineWidth(1.5)
	page.SetStrokeColor(gopdf.Blue)
	page.DrawRectangle(50, 620, 100, 60)
	page.SetFont(gopdf.Helvetica, 10)
	page.DrawText("Stroke", 70, 645)

	// 塗りつぶしのみの矩形
	page.SetFillColor(gopdf.Color{R: 1.0, G: 1.0, B: 0.0}) // 黄色
	page.FillRectangle(170, 620, 100, 60)
	page.SetFont(gopdf.Helvetica, 10)
	page.DrawText("Fill", 200, 645)

	// 枠線＋塗りつぶしの矩形
	page.SetStrokeColor(gopdf.Black)
	page.SetLineWidth(2.0)
	page.SetFillColor(gopdf.Color{R: 0.8, G: 0.8, B: 0.8}) // 灰色
	page.DrawAndFillRectangle(290, 620, 100, 60)
	page.SetFont(gopdf.Helvetica, 10)
	page.DrawText("Both", 315, 645)

	// === 円の描画 ===
	page.SetFont(gopdf.Helvetica, 14)
	page.SetStrokeColor(gopdf.Black)
	page.DrawText("Circles:", 50, 590)

	// 枠線のみの円
	page.SetLineWidth(1.5)
	page.SetStrokeColor(gopdf.Red)
	page.DrawCircle(90, 520, 40)
	page.SetFont(gopdf.Helvetica, 10)
	page.DrawText("Stroke", 65, 460)

	// 塗りつぶしのみの円
	page.SetFillColor(gopdf.Color{R: 0.0, G: 1.0, B: 0.0}) // 緑
	page.FillCircle(220, 520, 40)
	page.SetFont(gopdf.Helvetica, 10)
	page.DrawText("Fill", 200, 460)

	// 枠線＋塗りつぶしの円
	page.SetStrokeColor(gopdf.Blue)
	page.SetLineWidth(2.5)
	page.SetFillColor(gopdf.Color{R: 1.0, G: 0.8, B: 0.8}) // 薄い赤
	page.DrawAndFillCircle(350, 520, 40)
	page.SetFont(gopdf.Helvetica, 10)
	page.DrawText("Both", 330, 460)

	// === 複雑な図形 ===
	page.SetFont(gopdf.Helvetica, 14)
	page.SetStrokeColor(gopdf.Black)
	page.DrawText("Complex Shapes:", 50, 430)

	// 家の形
	page.SetLineWidth(2.0)
	page.SetStrokeColor(gopdf.Black)
	page.SetFillColor(gopdf.Color{R: 0.9, G: 0.9, B: 0.7}) // 薄い黄色

	// 家の本体
	page.DrawAndFillRectangle(80, 280, 100, 80)

	// 屋根（三角形の近似として線を使用）
	page.SetFillColor(gopdf.Color{R: 0.8, G: 0.4, B: 0.2}) // 茶色
	page.DrawLine(80, 360, 130, 400)
	page.DrawLine(130, 400, 180, 360)

	// ドア
	page.SetFillColor(gopdf.Color{R: 0.6, G: 0.3, B: 0.1}) // 濃い茶色
	page.FillRectangle(110, 280, 30, 50)

	// 窓
	page.SetFillColor(gopdf.Color{R: 0.7, G: 0.9, B: 1.0}) // 薄い青
	page.FillRectangle(90, 340, 20, 15)
	page.FillRectangle(150, 340, 20, 15)

	// 太陽
	page.SetStrokeColor(gopdf.Color{R: 1.0, G: 0.8, B: 0.0}) // オレンジ
	page.SetFillColor(gopdf.Color{R: 1.0, G: 1.0, B: 0.0})   // 黄色
	page.SetLineWidth(1.5)
	page.DrawAndFillCircle(350, 380, 25)

	// === カラーグラデーション風 ===
	page.SetFont(gopdf.Helvetica, 14)
	page.SetStrokeColor(gopdf.Black)
	page.DrawText("Color Variations:", 50, 250)

	// 赤のグラデーション風
	for i := 0; i < 10; i++ {
		brightness := float64(i) / 9.0
		page.SetFillColor(gopdf.Color{R: brightness, G: 0, B: 0})
		page.FillRectangle(50+float64(i*20), 180, 20, 50)
	}

	// 緑のグラデーション風
	for i := 0; i < 10; i++ {
		brightness := float64(i) / 9.0
		page.SetFillColor(gopdf.Color{R: 0, G: brightness, B: 0})
		page.FillRectangle(50+float64(i*20), 120, 20, 50)
	}

	// 青のグラデーション風
	for i := 0; i < 10; i++ {
		brightness := float64(i) / 9.0
		page.SetFillColor(gopdf.Color{R: 0, G: 0, B: brightness})
		page.FillRectangle(50+float64(i*20), 60, 20, 50)
	}

	// === フッター ===
	page.SetFont(gopdf.HelveticaOblique, 10)
	page.SetStrokeColor(gopdf.Black)
	page.DrawText("Generated with gopdf - https://github.com/ryomak/gopdf", 50, 30)

	// ファイルに出力
	file, err := os.Create("output.pdf")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	if err := doc.WriteTo(file); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing PDF: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("PDF created successfully: output.pdf")
	fmt.Println("Open the file to see various graphics (lines, rectangles, circles)!")
}
