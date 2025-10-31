// Debug font encoding issues in PDFs
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ryomak/gopdf"
	"github.com/ryomak/gopdf/internal/content"
	"github.com/ryomak/gopdf/internal/core"
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

	// Access internal reader
	internalReader := getInternalReader(reader)
	if internalReader == nil {
		log.Fatal("Failed to access internal reader")
	}

	// Get first page
	page, err := internalReader.GetPage(0)
	if err != nil {
		log.Fatalf("Failed to get page: %v", err)
	}

	// Get resources
	resources, err := internalReader.GetPageResources(page)
	if err != nil {
		log.Fatalf("Failed to get resources: %v", err)
	}

	// Get fonts
	fontsObj, ok := resources[core.Name("Font")]
	if !ok {
		fmt.Println("No fonts found")
		return
	}

	fonts, ok := fontsObj.(core.Dictionary)
	if !ok {
		fmt.Println("Invalid fonts dictionary")
		return
	}

	fmt.Printf("Found %d fonts\n\n", len(fonts))

	// Analyze each font
	for fontName, fontObj := range fonts {
		fmt.Printf("=== Font: %s ===\n", fontName)

		ref, ok := fontObj.(*core.Reference)
		if !ok {
			fmt.Println("  Not a reference, skipping")
			continue
		}

		fontDict, err := internalReader.GetDictionary(ref)
		if err != nil {
			fmt.Printf("  Error getting font dict: %v\n", err)
			continue
		}

		// Basic font info
		if subtype, ok := fontDict[core.Name("Subtype")].(core.Name); ok {
			fmt.Printf("  Subtype: %s\n", subtype)
		}

		if baseFont, ok := fontDict[core.Name("BaseFont")].(core.Name); ok {
			fmt.Printf("  BaseFont: %s\n", baseFont)
		}

		// Check for Encoding
		if encodingObj, ok := fontDict[core.Name("Encoding")]; ok {
			fmt.Printf("  Encoding: %v\n", encodingObj)

			// If it's a reference, get the actual encoding
			if encRef, ok := encodingObj.(*core.Reference); ok {
				encDict, err := internalReader.GetDictionary(encRef)
				if err == nil {
					fmt.Printf("    Encoding Dict: %v\n", encDict)
				}
			}
		} else {
			fmt.Println("  Encoding: NOT FOUND")
		}

		// Check for ToUnicode CMap
		if toUnicodeObj, ok := fontDict[core.Name("ToUnicode")]; ok {
			fmt.Println("  ToUnicode: PRESENT")

			// Try to get the ToUnicode stream
			if toUnicodeRef, ok := toUnicodeObj.(*core.Reference); ok {
				toUnicodeStream, err := internalReader.GetStream(toUnicodeRef)
				if err == nil {
					fmt.Printf("    ToUnicode stream length: %d bytes\n", len(toUnicodeStream.Data))

					// Show first 500 characters of the CMap
					preview := string(toUnicodeStream.Data)
					if len(preview) > 500 {
						preview = preview[:500] + "..."
					}
					fmt.Printf("    ToUnicode preview:\n%s\n", preview)
				}
			}
		} else {
			fmt.Println("  ToUnicode: NOT FOUND ⚠️")
		}

		// Check for DescendantFonts (CIDFont)
		if descendantObj, ok := fontDict[core.Name("DescendantFonts")]; ok {
			fmt.Println("  DescendantFonts: PRESENT (CIDFont)")

			if descArray, ok := descendantObj.(core.Array); ok && len(descArray) > 0 {
				if descRef, ok := descArray[0].(*core.Reference); ok {
					descDict, err := internalReader.GetDictionary(descRef)
					if err == nil {
						if cidInfo, ok := descDict[core.Name("CIDSystemInfo")]; ok {
							fmt.Printf("    CIDSystemInfo: %v\n", cidInfo)
						}
					}
				}
			}
		}

		fmt.Println()
	}

	// Now extract text and show which font each character uses
	fmt.Println("\n=== Text Extraction with Font Mapping ===\n")

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
		fmt.Println("\n")
	}
}

// Helper to access internal reader
func getInternalReader(r *gopdf.PDFReader) *content.Reader {
	// This is a workaround to access internal reader
	// In real implementation, we would add a public method
	return nil // This won't work, but shows the intent
}

// We need to modify the code to actually access the internal reader
// Let's create a simpler version that uses existing APIs
