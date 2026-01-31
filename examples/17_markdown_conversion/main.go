package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ryomak/gopdf"
)

func main() {
	fmt.Println("Converting Markdown to PDF...")

	// Example 1: Convert from Markdown file with default settings
	if err := convertMarkdownFile(); err != nil {
		log.Fatalf("Failed to convert markdown file: %v", err)
	}
	fmt.Println("✓ Created output_from_file.pdf")

	// Example 2: Convert from Markdown string
	if err := convertMarkdownString(); err != nil {
		log.Fatalf("Failed to convert markdown string: %v", err)
	}
	fmt.Println("✓ Created output_from_string.pdf")

	// Example 3: Convert with custom style
	if err := convertWithCustomStyle(); err != nil {
		log.Fatalf("Failed to convert with custom style: %v", err)
	}
	fmt.Println("✓ Created output_custom_style.pdf")
}

func convertMarkdownFile() error {
	// Convert Markdown file to PDF with default settings
	doc, err := gopdf.NewMarkdownDocumentFromFile("sample.md", &gopdf.MarkdownOptions{
		Mode:        gopdf.MarkdownModeDocument,
		PageSize:    gopdf.PageSizeA4,
		Orientation: gopdf.Portrait,
	})
	if err != nil {
		return err
	}

	f, err := os.Create("output_from_file.pdf")
	if err != nil {
		return err
	}
	defer f.Close()

	return doc.WriteTo(f)
}

func convertMarkdownString() error {
	markdown := `# Hello from Markdown

This PDF was generated from a Markdown string!

## Features

- Simple API
- Pure Go implementation
- CommonMark support

## Conclusion

Converting Markdown to PDF is easy with gopdf.
`

	doc, err := gopdf.NewMarkdownDocument(markdown, &gopdf.MarkdownOptions{
		Mode:        gopdf.MarkdownModeDocument,
		PageSize:    gopdf.PageSizeA4,
		Orientation: gopdf.Portrait,
	})
	if err != nil {
		return err
	}

	f, err := os.Create("output_from_string.pdf")
	if err != nil {
		return err
	}
	defer f.Close()

	return doc.WriteTo(f)
}

func convertWithCustomStyle() error {
	markdown := `# Custom Styled Document

This document uses custom styling.

## Beautiful Headers

Headers can be styled with custom colors and sizes.

### Subheading

The body text also uses custom styling for better readability.

## Conclusion

Custom styles make your documents unique!
`

	// Create custom style
	style := gopdf.DefaultMarkdownStyle()
	style.H1Size = 42
	style.H2Size = 32
	style.H3Size = 24
	style.BodySize = 14
	style.HeadingColor = gopdf.Color{R: 0.1, G: 0.3, B: 0.7}
	style.TextColor = gopdf.Color{R: 0.2, G: 0.2, B: 0.2}
	style.LineSpacing = 1.5
	style.ParagraphSpacing = 18

	doc, err := gopdf.NewMarkdownDocument(markdown, &gopdf.MarkdownOptions{
		Mode:        gopdf.MarkdownModeDocument,
		PageSize:    gopdf.PageSizeA4,
		Orientation: gopdf.Portrait,
		Style:       style,
	})
	if err != nil {
		return err
	}

	f, err := os.Create("output_custom_style.pdf")
	if err != nil {
		return err
	}
	defer f.Close()

	return doc.WriteTo(f)
}
