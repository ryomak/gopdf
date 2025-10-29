package main

import (
	"fmt"
	"os"

	"github.com/ryomak/gopdf/internal/content"
	"github.com/ryomak/gopdf/internal/core"
	"github.com/ryomak/gopdf/internal/reader"
)

func main() {
	pdfPath := os.ExpandEnv("$HOME/Downloads/Receipt-2021-3422.pdf")

	fmt.Printf("Opening PDF: %s\n", pdfPath)

	// Open PDF at low level
	file, err := os.Open(pdfPath)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	r, err := reader.NewReader(file)
	if err != nil {
		fmt.Printf("Error creating reader: %v\n", err)
		os.Exit(1)
	}

	// Get first page
	page, err := r.GetPage(0)
	if err != nil {
		fmt.Printf("Error getting page: %v\n", err)
		os.Exit(1)
	}

	// Get page resources
	resources, ok := page["Resources"]
	if !ok {
		fmt.Println("No Resources in page")
		os.Exit(1)
	}

	if ref, ok := resources.(*core.Reference); ok {
		resources, err = r.ResolveReference(ref)
		if err != nil {
			fmt.Printf("Error resolving Resources: %v\n", err)
			os.Exit(1)
		}
	}

	resDict, ok := resources.(core.Dictionary)
	if !ok {
		fmt.Println("Resources is not a Dictionary")
		os.Exit(1)
	}

	// Get Font resources
	fontResources, ok := resDict["Font"]
	if !ok {
		fmt.Println("No Font in Resources")
		os.Exit(1)
	}

	if ref, ok := fontResources.(*core.Reference); ok {
		fontResources, err = r.ResolveReference(ref)
		if err != nil {
			fmt.Printf("Error resolving Font resources: %v\n", err)
			os.Exit(1)
		}
	}

	fontDict, ok := fontResources.(core.Dictionary)
	if !ok {
		fmt.Println("Font is not a Dictionary")
		os.Exit(1)
	}

	fmt.Printf("\nFonts in PDF:\n")
	for fontName := range fontDict {
		fmt.Printf("\n=== Font: %s ===\n", fontName)

		fontObj := fontDict[fontName]
		if ref, ok := fontObj.(*core.Reference); ok {
			fontObj, err = r.ResolveReference(ref)
			if err != nil {
				fmt.Printf("  Error resolving font: %v\n", err)
				continue
			}
		}

		if font, ok := fontObj.(core.Dictionary); ok {
			fmt.Printf("  Type: %v\n", font["Type"])
			fmt.Printf("  Subtype: %v\n", font["Subtype"])
			fmt.Printf("  BaseFont: %v\n", font["BaseFont"])
			fmt.Printf("  Encoding: %v\n", font["Encoding"])

			// Check for ToUnicode
			if toUnicode, ok := font["ToUnicode"]; ok {
				fmt.Printf("  ToUnicode: EXISTS (reference: %v)\n", toUnicode)

				// Try to resolve and read ToUnicode
				if ref, ok := toUnicode.(*core.Reference); ok {
					toUnicode, err = r.ResolveReference(ref)
					if err != nil {
						fmt.Printf("  Error resolving ToUnicode: %v\n", err)
						continue
					}
				}

				if stream, ok := toUnicode.(*core.Stream); ok {
					fmt.Printf("  ToUnicode stream length (compressed): %d bytes\n", len(stream.Data))

					// Decode the stream (decompress if needed)
					data, err := r.DecodeStream(stream)
					if err != nil {
						fmt.Printf("  Error decoding stream: %v\n", err)
						continue
					}

					fmt.Printf("  ToUnicode stream length (decompressed): %d bytes\n", len(data))

					// Show first 500 bytes
					preview := data
					if len(preview) > 500 {
						preview = preview[:500]
					}
					fmt.Printf("  First 500 bytes (decompressed):\n%s\n", string(preview))

					// Try to parse it
					cmap, err := content.ParseToUnicodeCMap(data)
					if err != nil {
						fmt.Printf("  ERROR parsing ToUnicode: %v\n", err)
					} else {
						fmt.Printf("  ToUnicode parsed successfully!\n")
						fmt.Printf("  - CharMap entries: %d\n", cmap.GetCharMapSize())
						fmt.Printf("  - Range entries: %d\n", cmap.GetRangesSize())

						// Show a few sample mappings
						samples := cmap.GetSampleMappings(10)
						if len(samples) > 0 {
							fmt.Printf("  Sample charMap mappings:\n")
							for cid, unicode := range samples {
								fmt.Printf("    CID 0x%04X -> U+%04X (%c)\n", cid, unicode, unicode)
							}
						}
					}
				} else {
					fmt.Printf("  ToUnicode is not a Stream: %T\n", toUnicode)
				}
			} else {
				fmt.Printf("  ToUnicode: NONE\n")
			}
		}
	}

	// Now try extraction with FontManager
	fmt.Println("\n\n=== Testing FontManager ===")
	fontManager := content.NewFontManager(r)

	// Test F4 and F5
	for _, fontName := range []string{"F4", "F5"} {
		fmt.Printf("\n--- Font: %s ---\n", fontName)
		fontInfo, err := fontManager.GetFont(fontName, resDict)
		if err != nil {
			fmt.Printf("  Error: %v\n", err)
			continue
		}

		fmt.Printf("  Name: %s\n", fontInfo.Name)
		if fontInfo.ToUnicodeCMap != nil {
			fmt.Printf("  ToUnicode CMap: LOADED\n")
			fmt.Printf("  - CharMap entries: %d\n", fontInfo.ToUnicodeCMap.GetCharMapSize())
			fmt.Printf("  - Range entries: %d\n", fontInfo.ToUnicodeCMap.GetRangesSize())

			// Test some actual CIDs from the document
			// From the output we saw: "ڒ", "Ŷ", "Ǹ", "ʊ", "ɤ"
			// Let's see what CIDs these are
			testChars := []rune{'ڒ', 'Ŷ', 'Ǹ', 'ʊ', 'ɤ', 'ž', 'Å'}
			fmt.Printf("  Testing lookups for CIDs seen in output:\n")
			for _, ch := range testChars {
				cid := uint16(ch)
				if mapped, ok := fontInfo.ToUnicodeCMap.Lookup(cid); ok {
					fmt.Printf("    CID 0x%04X (%c) -> U+%04X (%c)\n", cid, ch, mapped, mapped)
				} else {
					fmt.Printf("    CID 0x%04X (%c) -> NOT FOUND\n", cid, ch)
				}
			}
		} else {
			fmt.Printf("  ToUnicode CMap: NONE\n")
		}
	}
}
