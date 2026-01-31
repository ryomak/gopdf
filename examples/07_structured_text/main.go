// Example: 07_structured_text
// This example demonstrates how to extract text with position and style information.
package main

import (
	"fmt"
	"os"

	"github.com/ryomak/gopdf"
)

func main() {
	// まず、サンプルPDFを生成
	fmt.Println("Creating sample PDF with structured text...")
	createSamplePDF("sample.pdf")

	// PDFファイルを読み込む
	fmt.Println("\nReading PDF file...")
	reader, err := gopdf.Open("sample.pdf")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening PDF: %v\n", err)
		os.Exit(1)
	}
	defer reader.Close()

	// 位置情報付きテキスト要素を取得
	fmt.Println("\nExtracting text elements with position information:")
	elements, err := reader.ExtractPageTextElements(0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error extracting text elements: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d text elements\n\n", len(elements))

	// 各テキスト要素の詳細を表示
	fmt.Println("Text elements (original order):")
	for i, elem := range elements {
		fmt.Printf("%2d: '%s'\n", i+1, elem.Text)
		fmt.Printf("    Position: (%.1f, %.1f)\n", elem.X, elem.Y)
		fmt.Printf("    Size: %.1f x %.1f\n", elem.Width, elem.Height)
		fmt.Printf("    Font: %s, Size: %.1f\n\n", elem.Font, elem.Size)
	}

	// 読み順序でソート
	fmt.Println("Sorting text elements by reading order...")
	sorted := gopdf.SortTextElements(elements)

	fmt.Println("\nSorted text elements (reading order):")
	for i, elem := range sorted {
		fmt.Printf("%2d: '%s' at (%.1f, %.1f)\n", i+1, elem.Text, elem.X, elem.Y)
	}

	// 文字列に変換
	text := gopdf.TextElementsToString(sorted)
	fmt.Println("\nCombined text:")
	fmt.Println(text)

	// 全ページのテキスト要素を取得
	fmt.Println("\n--- Extracting from all pages ---")
	allElements, err := reader.ExtractAllTextElements()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error extracting all text elements: %v\n", err)
		os.Exit(1)
	}

	for pageNum, pageElements := range allElements {
		fmt.Printf("\nPage %d: %d text elements\n", pageNum+1, len(pageElements))
		sorted := gopdf.SortTextElements(pageElements)
		text := gopdf.TextElementsToString(sorted)
		fmt.Println(text)
	}

	fmt.Println("\nStructured text extraction completed successfully!")
}

// createSamplePDF は構造的テキスト抽出用のサンプルPDFを作成する
func createSamplePDF(filename string) {
	doc := gopdf.New()

	// ページ1: 複数のテキスト要素
	page1 := doc.AddPage(gopdf.PageSizeA4, gopdf.Portrait)

	// タイトル（大きいフォント、上部）
	page1.SetFont(gopdf.FontHelveticaBold, 24)
	page1.DrawText("Structured Text Extraction", 50, 800)

	// サブタイトル（中サイズ、タイトルの下）
	page1.SetFont(gopdf.FontHelvetica, 14)
	page1.DrawText("Demonstration of Position and Style Information", 50, 770)

	// セクション1（通常サイズ）
	page1.SetFont(gopdf.FontHelveticaBold, 12)
	page1.DrawText("Section 1: Basic Text", 50, 730)

	page1.SetFont(gopdf.FontHelvetica, 10)
	page1.DrawText("This is the first paragraph with normal size text.", 50, 710)
	page1.DrawText("It demonstrates how text position is tracked.", 50, 695)

	// セクション2（異なる位置）
	page1.SetFont(gopdf.FontHelveticaBold, 12)
	page1.DrawText("Section 2: Multiple Lines", 50, 660)

	page1.SetFont(gopdf.FontCourier, 10)
	page1.DrawText("Line 1 with Courier font", 50, 640)
	page1.DrawText("Line 2 with same font", 50, 625)
	page1.DrawText("Line 3 continues", 50, 610)

	// 右側にもテキスト（異なるX座標）
	page1.SetFont(gopdf.FontHelvetica, 10)
	page1.DrawText("Right", 300, 710)
	page1.DrawText("Side", 300, 695)
	page1.DrawText("Text", 300, 680)

	// ページ2: シンプルなテキスト
	page2 := doc.AddPage(gopdf.PageSizeA4, gopdf.Portrait)

	page2.SetFont(gopdf.FontHelveticaBold, 18)
	page2.DrawText("Page 2", 50, 800)

	page2.SetFont(gopdf.FontHelvetica, 12)
	page2.DrawText("This is the second page.", 50, 770)
	page2.DrawText("It has less content.", 50, 755)

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
