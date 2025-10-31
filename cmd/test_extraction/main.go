// Test extraction accuracy for gopdf
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

	fmt.Printf("=== Testing PDF Extraction: %s ===\n\n", pdfPath)

	reader, err := gopdf.Open(pdfPath)
	if err != nil {
		log.Fatalf("Failed to open PDF: %v", err)
	}
	defer reader.Close()

	pageCount := reader.PageCount()
	fmt.Printf("Total pages: %d\n\n", pageCount)

	// Test each page
	for pageNum := 0; pageNum < pageCount; pageNum++ {
		fmt.Printf("\n=== Page %d ===\n", pageNum+1)

		// 1. Extract raw text elements
		fmt.Println("\n--- Raw Text Elements ---")
		elements, err := reader.ExtractPageTextElements(pageNum)
		if err != nil {
			log.Printf("Error extracting text elements: %v", err)
			continue
		}

		fmt.Printf("Total elements: %d\n", len(elements))
		if len(elements) > 0 {
			fmt.Println("First 10 elements:")
			for i, elem := range elements {
				if i >= 10 {
					break
				}
				fmt.Printf("  [%d] Text=%q, X=%.2f, Y=%.2f, Size=%.2f, Font=%s\n",
					i, elem.Text, elem.X, elem.Y, elem.Size, elem.Font)
			}
		}

		// 2. Extract text blocks
		fmt.Println("\n--- Text Blocks ---")
		blocks, err := reader.ExtractPageTextBlocks(pageNum)
		if err != nil {
			log.Printf("Error extracting text blocks: %v", err)
			continue
		}

		fmt.Printf("Total blocks: %d\n", len(blocks))
		for i, block := range blocks {
			fmt.Printf("\nBlock %d:\n", i)
			fmt.Printf("  Position: (%.2f, %.2f)\n", block.Rect.X, block.Rect.Y)
			fmt.Printf("  Size: %.2f x %.2f\n", block.Rect.Width, block.Rect.Height)
			fmt.Printf("  Font: %s, Size: %.2f\n", block.Font, block.FontSize)
			fmt.Printf("  Elements: %d\n", len(block.Elements))
			fmt.Printf("  Text: %q\n", block.Text)
		}

		// 3. Extract layout (text + images)
		fmt.Println("\n--- Page Layout ---")
		layout, err := reader.ExtractPageLayout(pageNum)
		if err != nil {
			log.Printf("Error extracting layout: %v", err)
			continue
		}

		fmt.Printf("Page size: %.2f x %.2f\n", layout.Width, layout.Height)
		fmt.Printf("Text blocks: %d\n", len(layout.TextBlocks))
		fmt.Printf("Images: %d\n", len(layout.Images))

		if layout.PageCTM != nil {
			fmt.Printf("PageCTM: [%.2f %.2f %.2f %.2f %.2f %.2f]\n",
				layout.PageCTM.A, layout.PageCTM.B, layout.PageCTM.C,
				layout.PageCTM.D, layout.PageCTM.E, layout.PageCTM.F)
		}

		// 4. Extract content blocks (sorted)
		fmt.Println("\n--- Sorted Content Blocks ---")
		contentBlocks := layout.SortedContentBlocks()
		fmt.Printf("Total content blocks: %d\n", len(contentBlocks))

		for i, block := range contentBlocks {
			bounds := block.Bounds()
			fmt.Printf("\nBlock %d [%s]:\n", i, block.Type())
			fmt.Printf("  Position: (%.2f, %.2f)\n", bounds.X, bounds.Y)
			fmt.Printf("  Size: %.2f x %.2f\n", bounds.Width, bounds.Height)

			switch block.Type() {
			case gopdf.ContentBlockTypeText:
				tb := block.(gopdf.TextBlock)
				fmt.Printf("  Font: %s, Size: %.2f\n", tb.Font, tb.FontSize)
				fmt.Printf("  Text: %q\n", tb.Text)
			case gopdf.ContentBlockTypeImage:
				ib := block.(gopdf.ImageBlock)
				fmt.Printf("  Image: %dx%d (%s)\n", ib.Width, ib.Height, ib.Format)
				fmt.Printf("  Transform: [%.2f %.2f %.2f %.2f %.2f %.2f]\n",
					ib.Transform.A, ib.Transform.B, ib.Transform.C,
					ib.Transform.D, ib.Transform.E, ib.Transform.F)
			}
		}

		// 5. Check for coordinate system issues
		fmt.Println("\n--- Coordinate System Check ---")
		if len(blocks) > 1 {
			// Check if blocks are sorted correctly (top to bottom)
			for i := 0; i < len(blocks)-1; i++ {
				curr := blocks[i]
				next := blocks[i+1]

				// In PDF coordinate system (origin at bottom-left):
				// Higher Y values should appear first (top of page)
				if curr.Rect.Y < next.Rect.Y {
					fmt.Printf("⚠️  WARNING: Block %d (Y=%.2f) appears before Block %d (Y=%.2f)\n",
						i, curr.Rect.Y, i+1, next.Rect.Y)
					fmt.Printf("   This suggests incorrect sorting or coordinate transformation\n")
				}
			}
		}

		// 6. Check for character encoding issues
		fmt.Println("\n--- Character Encoding Check ---")
		hasControlChars := false
		hasUnusualChars := false

		for _, elem := range elements {
			for _, r := range elem.Text {
				if r < 32 && r != '\n' && r != '\r' && r != '\t' {
					hasControlChars = true
				}
				if r > 127 && r < 160 {
					hasUnusualChars = true
				}
			}
		}

		if hasControlChars {
			fmt.Println("⚠️  WARNING: Control characters detected in text")
		}
		if hasUnusualChars {
			fmt.Println("⚠️  WARNING: Unusual Unicode characters detected (possible encoding issue)")
		}
		if !hasControlChars && !hasUnusualChars {
			fmt.Println("✓ No obvious character encoding issues")
		}
	}

	fmt.Println("\n\n=== Extraction Test Complete ===")
}
