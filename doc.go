// Package gopdf provides a high-level API for creating and manipulating PDF documents.
//
// # Overview
//
// gopdf is a pure Go library for creating, reading, and manipulating PDF documents.
// It supports a wide range of features including:
//
//   - Creating PDF documents with text, images, and graphics
//   - Reading and parsing existing PDF documents
//   - Text extraction and layout analysis
//   - TrueType font embedding with Japanese support
//   - Image embedding (JPEG, PNG)
//   - Ruby text (furigana) annotations
//   - Markdown to PDF conversion
//   - PDF translation with layout preservation
//   - PDF encryption and security
//
// # Basic Usage
//
// Creating a simple PDF document:
//
//	doc := gopdf.New()
//	page := doc.NewPage(gopdf.PageSizeA4)
//	page.SetFont(gopdf.FontHelvetica, 12)
//	page.DrawText("Hello, World!", 100, 700)
//	doc.SaveToFile("output.pdf")
//
// Opening and reading an existing PDF:
//
//	reader, err := gopdf.Open("input.pdf")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	text, err := reader.ExtractText(0) // Extract text from first page
//
// # Page Sizes
//
// The package provides standard page sizes:
//
//	PageSizeA4     - 210mm x 297mm
//	PageSizeLetter - 8.5in x 11in
//	PageSizeLegal  - 8.5in x 14in
//	PageSizeA3     - 297mm x 420mm
//	PageSizeA5     - 148mm x 210mm
//
// # Fonts
//
// gopdf supports the 14 standard PDF fonts and TrueType font embedding:
//
//	page.SetFont(gopdf.FontHelvetica, 12)     // Standard font
//	ttf, _ := gopdf.LoadTTF("custom.ttf")
//	page.SetFont(ttf, 12)                  // TrueType font
//
// # Japanese Support
//
// For Japanese text, load a system font or specify a font path:
//
//	// Option 1: Load from system fonts (macOS: Hiragino, Linux: Noto, Windows: Yu Gothic)
//	jpFont, err := gopdf.LoadSystemJapaneseFont()
//	if err != nil {
//	    log.Fatal("No Japanese font found on system")
//	}
//	page.SetFont(jpFont, 12)
//	page.DrawText("日本語テキスト", 100, 700)
//
//	// Option 2: Load from specific path
//	jpFont, err := gopdf.LoadTTF("/path/to/NotoSansJP-Regular.ttf")
package gopdf
