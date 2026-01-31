// Debug font encoding issues in PDFs
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ryomak/gopdf"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <pdf-file>")
		os.Exit(1)
	}

	pdfPath := os.Args[1]

	fmt.Printf("=== Font Encoding Debug: %s ===\n\n", pdfPath)

	reader, err := gopdf.Open(pdfPath)
	if err != nil {
		log.Fatalf("Failed to open PDF: %v", err)
	}
	defer reader.Close()

	fmt.Printf("Page count: %d\n\n", reader.PageCount())

	// Extract text and show which font each character uses
	fmt.Println("=== Text Extraction with Font Mapping ===")

	elements, err := reader.ExtractPageTextElements(0)
	if err != nil {
		log.Fatalf("Failed to extract text: %v", err)
	}

	// Group by font
	fontGroups := make(map[string][]string)
	for _, elem := range elements {
		fontGroups[elem.Font] = append(fontGroups[elem.Font], elem.Text)
	}

	for font, texts := range fontGroups {
		fmt.Printf("Font %s:\n", font)

		// Show unique characters
		charSet := make(map[rune]bool)
		for _, text := range texts {
			for _, r := range text {
				charSet[r] = true
			}
		}

		fmt.Printf("  Total elements: %d\n", len(texts))
		fmt.Printf("  Unique characters: %d\n", len(charSet))
		fmt.Print("  Sample characters: ")
		count := 0
		for r := range charSet {
			if count >= 20 {
				fmt.Print("...")
				break
			}
			if r >= 32 && r < 127 {
				fmt.Printf("%c ", r)
			} else {
				fmt.Printf("U+%04X ", r)
			}
			count++
		}
		fmt.Println()
		fmt.Println()
	}
}
