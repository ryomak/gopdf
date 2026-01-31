// Example: 01_empty_page
// This example demonstrates how to create a simple PDF with empty pages.
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
	doc.AddPage(gopdf.PageSizeA4, gopdf.Portrait)

	// Letterサイズの横向きページを追加
	doc.AddPage(gopdf.PageSizeLetter, gopdf.Landscape)

	// A4サイズの横向きページを追加
	doc.AddPage(gopdf.PageSizeA4, gopdf.Landscape)

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
	fmt.Printf("Total pages: %d\n", doc.PageCount())
}
