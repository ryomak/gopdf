package main

import (
	"fmt"
	"os"

	"github.com/ryomak/gopdf"
)

func main() {
	pdfPath := os.ExpandEnv("$HOME/Downloads/Receipt-2021-3422.pdf")

	fmt.Printf("Opening PDF: %s\n", pdfPath)

	reader, err := gopdf.Open(pdfPath)
	if err != nil {
		fmt.Printf("Error opening PDF: %v\n", err)
		os.Exit(1)
	}
	defer reader.Close()

	fmt.Printf("Pages: %d\n\n", reader.PageCount())

	// Extract text from first page
	fmt.Println("=== Extracting text from page 1 ===")
	text, err := reader.ExtractPageText(0)
	if err != nil {
		fmt.Printf("Error extracting text: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Extracted text (first 1000 chars):\n%s\n\n", text[:min(1000, len(text))])

	// Extract text elements
	fmt.Println("=== Extracting text elements (first 30) ===")
	elements, err := reader.ExtractPageTextElements(0)
	if err != nil {
		fmt.Printf("Error extracting elements: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Total elements: %d\n\n", len(elements))
	for i, elem := range elements {
		if i >= 30 {
			break
		}
		fmt.Printf("[%d] Text: %q, Font: %s\n", i, elem.Text, elem.Font)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
