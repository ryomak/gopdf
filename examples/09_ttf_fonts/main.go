// Example: 09_ttf_fonts
// This example demonstrates how to use TrueType fonts with gopdf,
// including Japanese text rendering.
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/ryomak/gopdf"
)

func main() {
	// Parse command line arguments
	fontPath := flag.String("font", "", "Path to TTF font file for English text")
	jpFontPath := flag.String("jp-font", "", "Path to Japanese TTF font file")
	flag.Parse()

	fmt.Println("TTF Font and Japanese Text Example")
	fmt.Println("===================================")
	fmt.Println()

	// Create a new PDF document
	doc := gopdf.New()

	// Add a page
	page := doc.AddPage(gopdf.PageSizeA4, gopdf.Portrait)

	// Title with standard font
	page.SetFont(gopdf.FontHelveticaBold, 24)
	page.DrawText("TTF Font Support Example", 50, 800)

	// Load a TTF font
	var ttfFont *gopdf.TTFFont
	var err error
	if *fontPath != "" {
		ttfFont, err = gopdf.LoadTTF(*fontPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading font from %s: %v\n", *fontPath, err)
			os.Exit(1)
		}
	} else {
		ttfFont, err = loadSystemFont()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading TTF font: %v\n", err)
			fmt.Println("Use -font flag to specify a font: go run main.go -font /path/to/font.ttf")
			os.Exit(1)
		}
	}

	fmt.Printf("Loaded TTF font: %s\n", ttfFont.Name())

	// Set the TTF font
	if err := page.SetTTFFont(ttfFont, 18); err != nil {
		fmt.Fprintf(os.Stderr, "Error setting TTF font: %v\n", err)
		os.Exit(1)
	}

	// Draw English text with TTF font
	page.DrawText("Hello, World!", 50, 750)

	// Draw text with special characters
	page.DrawText("Unicode: € £ ¥ © ® ™", 50, 720)

	// Try to draw Japanese text
	var japaneseFont *gopdf.TTFFont
	if *jpFontPath != "" {
		japaneseFont, err = gopdf.LoadTTF(*jpFontPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading Japanese font from %s: %v\n", *jpFontPath, err)
		}
	} else {
		japaneseFont, err = gopdf.LoadSystemJapaneseFont()
	}

	if err == nil {
		fmt.Printf("Loaded Japanese font: %s\n", japaneseFont.Name())

		if err := page.SetTTFFont(japaneseFont, 18); err != nil {
			fmt.Fprintf(os.Stderr, "Error setting Japanese font: %v\n", err)
		} else {
			// Draw Japanese text
			page.DrawText("こんにちは、世界！", 50, 680)
			page.DrawText("日本語のテキストです。", 50, 650)
		}
	} else {
		fmt.Println("Japanese font not found (use -jp-font flag to specify)")
		// Draw a note with the standard English font
		page.SetFont(gopdf.FontHelvetica, 14)
		page.DrawText("(Japanese font not available on this system)", 50, 680)
	}

	// Add more text at different sizes
	page.SetTTFFont(ttfFont, 12)
	page.DrawText("TTF fonts support Unicode characters", 50, 600)

	page.SetTTFFont(ttfFont, 16)
	page.DrawText("Different sizes are supported", 50, 570)

	page.SetTTFFont(ttfFont, 24)
	page.DrawText("Larger text", 50, 530)

	// Text width calculation
	width, err := ttfFont.TextWidth("Hello, World!", 18)
	if err == nil {
		fmt.Printf("Text width of 'Hello, World!' at 18pt: %.2f\n", width)
	}

	// Save to file
	filename := "ttf_fonts.pdf"
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

	fmt.Printf("\nPDF created successfully: %s\n", filename)
	fmt.Println("\nOpen the PDF to see TTF font rendering!")
}

// loadSystemFont loads a system font based on the operating system
func loadSystemFont() (*gopdf.TTFFont, error) {
	var fontPaths []string

	switch runtime.GOOS {
	case "darwin":
		// macOS system fonts
		fontPaths = []string{
			"/System/Library/Fonts/Helvetica.ttc",
			"/System/Library/Fonts/Supplemental/Arial.ttf",
			"/Library/Fonts/Arial.ttf",
		}
	case "linux":
		// Linux system fonts
		fontPaths = []string{
			"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
			"/usr/share/fonts/TTF/DejaVuSans.ttf",
			"/usr/share/fonts/liberation/LiberationSans-Regular.ttf",
		}
	case "windows":
		// Windows system fonts
		fontPaths = []string{
			"C:\\Windows\\Fonts\\arial.ttf",
			"C:\\Windows\\Fonts\\calibri.ttf",
		}
	}

	// Try each font path
	for _, path := range fontPaths {
		if _, err := os.Stat(path); err == nil {
			font, err := gopdf.LoadTTF(path)
			if err == nil {
				return font, nil
			}
		}
	}

	return nil, fmt.Errorf("no suitable system font found")
}
