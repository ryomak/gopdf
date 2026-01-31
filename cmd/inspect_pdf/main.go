package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ryomak/gopdf"
	"github.com/ryomak/gopdf/internal/reader"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run cmd/inspect_pdf/main.go <pdf_file>")
		os.Exit(1)
	}

	inputPath := os.Args[1]

	file, err := os.Open(inputPath)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	r, err := reader.NewReader(file)
	if err != nil {
		log.Fatalf("Failed to create reader: %v", err)
	}

	fmt.Printf("=== PDF Low-Level Inspection ===\n\n")

	// ページ情報を取得
	page, err := r.GetPage(0)
	if err != nil {
		log.Fatalf("Failed to get page: %v", err)
	}

	fmt.Println("--- Page Dictionary ---")
	for key, val := range page {
		fmt.Printf("%s: %v\n", key, val)
	}

	// MediaBox
	if mediaBox, ok := page["MediaBox"]; ok {
		fmt.Printf("\nMediaBox: %v\n", mediaBox)
	}

	// CropBox
	if cropBox, ok := page["CropBox"]; ok {
		fmt.Printf("CropBox: %v\n", cropBox)
	}

	// Rotate
	if rotate, ok := page["Rotate"]; ok {
		fmt.Printf("Rotate: %v\n", rotate)
	}

	// UserUnit
	if userUnit, ok := page["UserUnit"]; ok {
		fmt.Printf("UserUnit: %v\n", userUnit)
	}

	// Contents
	fmt.Println("\n--- Content Stream (first 2000 chars) ---")
	contents, err := r.GetPageContents(page)
	if err != nil {
		log.Fatalf("Failed to get contents: %v", err)
	}

	contentStr := string(contents)
	if len(contentStr) > 2000 {
		contentStr = contentStr[:2000] + "\n... (truncated)"
	}
	fmt.Println(contentStr)

	// gopdfでのページサイズ
	reader2, err := gopdf.Open(inputPath)
	if err != nil {
		log.Fatalf("Failed to open with gopdf: %v", err)
	}
	defer reader2.Close()

	layout, err := reader2.ExtractPageLayout(0)
	if err != nil {
		log.Fatalf("Failed to extract layout: %v", err)
	}

	fmt.Printf("\n--- gopdf Page Info ---\n")
	fmt.Printf("Width: %.2f\n", layout.Width)
	fmt.Printf("Height: %.2f\n", layout.Height)
}
