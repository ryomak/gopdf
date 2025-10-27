// Example: 06_read_pdf
// This example demonstrates how to read an existing PDF file and extract basic information.
package main

import (
	"fmt"
	"os"

	"github.com/ryomak/gopdf"
	"github.com/ryomak/gopdf/internal/font"
)

func main() {
	// まず、サンプルPDFを生成
	fmt.Println("Creating sample PDF...")
	createSamplePDF("sample.pdf")

	// PDFファイルを読み込む
	fmt.Println("\nReading PDF file...")
	reader, err := gopdf.Open("sample.pdf")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening PDF: %v\n", err)
		os.Exit(1)
	}
	defer reader.Close()

	// ページ数を取得
	pageCount := reader.PageCount()
	fmt.Printf("  Page count: %d\n", pageCount)

	// メタデータを取得
	info := reader.Info()
	fmt.Printf("  Title: %s\n", info.Title)
	fmt.Printf("  Author: %s\n", info.Author)
	fmt.Printf("  Subject: %s\n", info.Subject)
	fmt.Printf("  Keywords: %s\n", info.Keywords)
	fmt.Printf("  Creator: %s\n", info.Creator)
	fmt.Printf("  Producer: %s\n", info.Producer)

	fmt.Println("\nPDF reading completed successfully!")
	fmt.Println("\nNote: Text extraction is not yet implemented.")
	fmt.Println("Future versions will support:")
	fmt.Println("  - Text extraction from pages")
	fmt.Println("  - Image extraction")
	fmt.Println("  - More detailed page information")
}

// createSamplePDF は読み込み用のサンプルPDFを作成する
func createSamplePDF(filename string) {
	doc := gopdf.New()

	// ページ1
	page1 := doc.AddPage(gopdf.A4, gopdf.Portrait)
	page1.SetFont(font.HelveticaBold, 24)
	page1.DrawText("Sample PDF for Reading", 50, 800)

	page1.SetFont(font.Helvetica, 12)
	page1.DrawText("This PDF file is created for demonstrating the PDF reading functionality.", 50, 770)
	page1.DrawText("It contains multiple pages with text and graphics.", 50, 755)

	// グラフィックスを追加
	page1.SetStrokeColor(gopdf.NewRGB(0, 0, 255))
	page1.SetLineWidth(2)
	page1.DrawRectangle(50, 650, 500, 80)

	page1.SetFont(font.Courier, 10)
	page1.DrawText("Page 1 of 3", 50, 680)

	// ページ2
	page2 := doc.AddPage(gopdf.A4, gopdf.Portrait)
	page2.SetFont(font.HelveticaBold, 18)
	page2.DrawText("Page 2 - Graphics Demo", 50, 800)

	// 円を描画
	page2.SetStrokeColor(gopdf.NewRGB(255, 0, 0))
	page2.SetFillColor(gopdf.NewRGB(255, 200, 200))
	page2.SetLineWidth(3)
	page2.DrawAndFillCircle(150, 600, 50)

	page2.SetFont(font.Courier, 10)
	page2.DrawText("Page 2 of 3", 50, 680)

	// ページ3
	page3 := doc.AddPage(gopdf.A4, gopdf.Portrait)
	page3.SetFont(font.HelveticaBold, 18)
	page3.DrawText("Page 3 - Final Page", 50, 800)

	page3.SetFont(font.Helvetica, 12)
	page3.DrawText("This is the last page of the sample PDF.", 50, 770)
	page3.DrawText("Thank you for using gopdf!", 50, 755)

	page3.SetFont(font.Courier, 10)
	page3.DrawText("Page 3 of 3", 50, 680)

	// ファイルに出力
	file, err := os.Create(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	if err := doc.WriteTo(file); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing PDF: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("  Created: %s\n", filename)
}
