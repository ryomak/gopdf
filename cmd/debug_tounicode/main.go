// Debug ToUnicode CMap for specific fonts
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ryomak/gopdf"
	"github.com/ryomak/gopdf/internal/content"
	"github.com/ryomak/gopdf/internal/core"
	"github.com/ryomak/gopdf/internal/reader"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <pdf-file>")
		os.Exit(1)
	}

	pdfPath := os.Args[1]

	file, err := os.Open(pdfPath)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	r, err := reader.NewReader(file)
	if err != nil {
		log.Fatalf("Failed to create reader: %v", err)
	}

	// Get first page
	page, err := r.GetPage(0)
	if err != nil {
		log.Fatalf("Failed to get page: %v", err)
	}

	// Get resources
	resources, err := r.GetPageResources(page)
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

	fmt.Printf("=== ToUnicode CMap Analysis ===\n\n")

	// Focus on F4 font (the one with Japanese text)
	checkFont(r, fonts, "F4")
	checkFont(r, fonts, "F9")
	checkFont(r, fonts, "F12")
	checkFont(r, fonts, "F14")
	checkFont(r, fonts, "F16")
	checkFont(r, fonts, "F18")
	checkFont(r, fonts, "F19")

	// Now test actual text extraction
	fmt.Println("\n=== Actual Text Extraction Test ===")
	testExtraction(pdfPath)
}

func checkFont(r *reader.Reader, fonts core.Dictionary, fontName string) {
	fmt.Printf("--- Font: %s ---\n", fontName)

	fontObj, ok := fonts[core.Name(fontName)]
	if !ok {
		fmt.Println("  Not found")
		return
	}

	ref, ok := fontObj.(*core.Reference)
	if !ok {
		fmt.Println("  Not a reference")
		return
	}

	fontObj, err := r.ResolveReference(ref)
	if err != nil {
		fmt.Printf("  Error: %v\n\n", err)
		return
	}
	fontDict, ok := fontObj.(core.Dictionary)
	if !ok {
		fmt.Println("  Not a dictionary")
		return
	}

	// Get BaseFont
	if baseFont, ok := fontDict[core.Name("BaseFont")].(core.Name); ok {
		fmt.Printf("  BaseFont: %s\n", baseFont)
	}

	// Get Subtype
	if subtype, ok := fontDict[core.Name("Subtype")].(core.Name); ok {
		fmt.Printf("  Subtype: %s\n", subtype)
	}

	// Get ToUnicode
	toUnicodeObj, ok := fontDict[core.Name("ToUnicode")]
	if !ok {
		fmt.Println("  ToUnicode: NOT FOUND")
		return
	}

	toUnicodeRef, ok := toUnicodeObj.(*core.Reference)
	if !ok {
		fmt.Println("  ToUnicode is not a reference")
		return
	}

	toUnicodeObj2, err := r.ResolveReference(toUnicodeRef)
	if err != nil {
		fmt.Printf("  Error getting ToUnicode: %v\n\n", err)
		return
	}
	toUnicodeStream, ok := toUnicodeObj2.(*core.Stream)
	if !ok {
		fmt.Println("  ToUnicode is not a stream")
		return
	}

	// Decode stream
	data, err := r.DecodeStream(toUnicodeStream)
	if err != nil {
		fmt.Printf("  Error decoding ToUnicode: %v\n\n", err)
		return
	}

	// Parse CMap
	cmap, err := content.ParseToUnicodeCMap(data)
	if err != nil {
		fmt.Printf("  Error parsing ToUnicode CMap: %v\n\n", err)
		return
	}

	fmt.Printf("  ToUnicode CMap parsed ✓\n")
	fmt.Printf("  Char mappings: %d\n", cmap.GetCharMapSize())
	fmt.Printf("  Range mappings: %d\n", cmap.GetRangesSize())

	// Show sample mappings
	samples := cmap.GetSampleMappings(10)
	if len(samples) > 0 {
		fmt.Println("  Sample mappings:")
		for cid, unicode := range samples {
			fmt.Printf("    CID 0x%04X → U+%04X (%c)\n", cid, unicode, unicode)
		}
	}

	// Test specific CIDs that appear in the problematic text
	// From content stream: <0692> appears in the text
	testCIDs := []uint16{0x0692, 0x00F3, 0x0034, 0x0039, 0x0031, 0x0030}
	fmt.Println("  Testing specific CIDs:")
	for _, cid := range testCIDs {
		if r, ok := cmap.Lookup(cid); ok {
			fmt.Printf("    CID 0x%04X → U+%04X (%c)\n", cid, r, r)
		} else {
			fmt.Printf("    CID 0x%04X → NOT MAPPED\n", cid)
		}
	}

	fmt.Println()
}

func testExtraction(pdfPath string) {
	reader, err := gopdf.Open(pdfPath)
	if err != nil {
		log.Fatalf("Failed to open PDF: %v", err)
	}
	defer reader.Close()

	elements, err := reader.ExtractPageTextElements(0)
	if err != nil {
		log.Fatalf("Failed to extract text: %v", err)
	}

	// Look for elements with problematic text
	fmt.Println("Elements with unusual characters:")
	for i, elem := range elements {
		if elem.Font != "F4" && elem.Font != "F9" && elem.Font != "F12" &&
		   elem.Font != "F14" && elem.Font != "F16" && elem.Font != "F18" && elem.Font != "F19" {
			continue
		}

		hasUnusual := false
		for _, r := range elem.Text {
			// Check for Latin extended characters that might be garbled Japanese
			if (r >= 0x00C0 && r <= 0x00FF) || r == 0x2030 {
				hasUnusual = true
				break
			}
		}

		if hasUnusual && i < 50 {
			fmt.Printf("  [%d] Font=%s, Text=%q, Hex=", i, elem.Font, elem.Text)
			for _, r := range elem.Text {
				fmt.Printf("U+%04X ", r)
			}
			fmt.Println()
		}
	}
}
