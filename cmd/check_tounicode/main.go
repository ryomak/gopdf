// Check ToUnicode CMap content in PDF fonts
package main

import (
	"fmt"
	"log"
	"os"

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

	r, err := reader.New(file)
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

	fmt.Printf("=== Checking ToUnicode CMaps ===\n\n")

	// Check each font
	for fontName, fontObj := range fonts {
		fmt.Printf("Font: %s\n", fontName)

		ref, ok := fontObj.(*core.Reference)
		if !ok {
			fmt.Println("  Not a reference\n")
			continue
		}

		fontDict, err := r.GetDictionary(ref)
		if err != nil {
			fmt.Printf("  Error: %v\n\n", err)
			continue
		}

		// Get BaseFont
		if baseFont, ok := fontDict[core.Name("BaseFont")].(core.Name); ok {
			fmt.Printf("  BaseFont: %s\n", baseFont)
		}

		// Get Subtype
		if subtype, ok := fontDict[core.Name("Subtype")].(core.Name); ok {
			fmt.Printf("  Subtype: %s\n", subtype)
		}

		// Check ToUnicode
		toUnicodeObj, hasToUnicode := fontDict[core.Name("ToUnicode")]
		if !hasToUnicode {
			fmt.Println("  ToUnicode: NOT FOUND ❌\n")
			continue
		}

		fmt.Println("  ToUnicode: FOUND ✓")

		// Get ToUnicode stream
		toUnicodeRef, ok := toUnicodeObj.(*core.Reference)
		if !ok {
			fmt.Println("  ToUnicode is not a reference\n")
			continue
		}

		toUnicodeStream, err := r.GetStream(toUnicodeRef)
		if err != nil {
			fmt.Printf("  Error getting ToUnicode stream: %v\n\n", err)
			continue
		}

		fmt.Printf("  ToUnicode stream size: %d bytes\n", len(toUnicodeStream.Data))

		// Show content
		content := string(toUnicodeStream.Data)

		// Show first 1000 chars
		if len(content) > 1000 {
			fmt.Printf("  Content preview (first 1000 chars):\n%s\n...\n\n", content[:1000])
		} else {
			fmt.Printf("  Full content:\n%s\n\n", content)
		}

		// Parse specific mappings if it contains beginbfchar or beginbfrange
		if containsMapping(content) {
			fmt.Println("  Contains character mappings ✓")
			parseMappings(content)
		} else {
			fmt.Println("  No character mappings found ❌")
		}

		fmt.Println()
	}
}

func containsMapping(content string) bool {
	return contains(content, "beginbfchar") || contains(content, "beginbfrange")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func parseMappings(content string) {
	// Look for beginbfchar sections
	lines := splitLines(content)
	inBfChar := false
	inBfRange := false
	mappingCount := 0

	for _, line := range lines {
		trimmed := trim(line)

		if trimmed == "beginbfchar" {
			inBfChar = true
			continue
		}
		if trimmed == "endbfchar" {
			inBfChar = false
			continue
		}
		if trimmed == "beginbfrange" {
			inBfRange = true
			continue
		}
		if trimmed == "endbfrange" {
			inBfRange = false
			continue
		}

		if (inBfChar || inBfRange) && len(trimmed) > 0 && trimmed[0] == '<' {
			if mappingCount < 10 {
				fmt.Printf("    Mapping: %s\n", trimmed)
			}
			mappingCount++
		}
	}

	if mappingCount > 10 {
		fmt.Printf("    ... and %d more mappings\n", mappingCount-10)
	}
	fmt.Printf("  Total mappings: %d\n", mappingCount)
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func trim(s string) string {
	start := 0
	end := len(s)

	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\r') {
		start++
	}

	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\r') {
		end--
	}

	return s[start:end]
}
