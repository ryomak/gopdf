// Example: 12_ocr_text_layer
// This example demonstrates how to add a text layer to image-based PDFs
// making them searchable and copyable (like OCR results).
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"os"

	"github.com/ryomak/gopdf"
)

func main() {
	// Example 1: Simple invisible text
	fmt.Println("--- Example 1: Simple Invisible Text ---")
	if err := createSimpleInvisibleText("simple_invisible.pdf"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Example 2: Simulate OCR with text layer
	fmt.Println("\n--- Example 2: Simulated OCR Text Layer ---")
	if err := createOCRTextLayer("ocr_searchable.pdf"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Example 3: Multiple words with coordinates
	fmt.Println("\n--- Example 3: Multiple Words with Coordinates ---")
	if err := createMultipleWordsExample("multiple_words.pdf"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nAll examples completed successfully!")
	fmt.Println("\nNote: Open the PDFs and try to:")
	fmt.Println("  1. Select and copy text (the image contains invisible text)")
	fmt.Println("  2. Search for text (Cmd+F / Ctrl+F)")
}

// createSimpleInvisibleText creates a PDF with an image and simple invisible text
func createSimpleInvisibleText(filename string) error {
	doc := gopdf.New()
	page := doc.AddPage(gopdf.PageSizeA4, gopdf.Portrait)

	// Create a simple image with text rendered on it
	img := createSampleImage("Hello, World!", "This is a sample image")
	tempFile := "temp_image.jpg"
	if err := saveImageAsJPEG(img, tempFile); err != nil {
		return fmt.Errorf("failed to save temp image: %w", err)
	}
	defer os.Remove(tempFile)

	// Load and draw the image
	pdfImage, err := gopdf.LoadJPEGFile(tempFile)
	if err != nil {
		return fmt.Errorf("failed to load image: %w", err)
	}
	if err := page.DrawImage(pdfImage, 0, 0, gopdf.PageSizeA4.Width, gopdf.PageSizeA4.Height); err != nil {
		return fmt.Errorf("failed to draw image: %w", err)
	}

	// Add invisible text at specific positions (matching the image content)
	// This makes the text copyable and searchable
	page.AddInvisibleText("Hello, World!", 50, 750, 200, 20)
	page.AddInvisibleText("This is a sample image", 50, 700, 250, 20)

	// Save
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if err := doc.WriteTo(file); err != nil {
		return fmt.Errorf("failed to write PDF: %w", err)
	}

	fmt.Printf("  Created: %s\n", filename)
	fmt.Println("  Try selecting and copying text from the image!")
	return nil
}

// createOCRTextLayer simulates OCR results and creates a searchable PDF
func createOCRTextLayer(filename string) error {
	doc := gopdf.New()
	page := doc.AddPage(gopdf.PageSizeA4, gopdf.Portrait)

	// Create a sample image
	img := createSampleImage(
		"Document Title",
		"First paragraph of text.",
		"Second paragraph of text.",
	)
	tempFile := "temp_ocr_image.jpg"
	if err := saveImageAsJPEG(img, tempFile); err != nil {
		return fmt.Errorf("failed to save temp image: %w", err)
	}
	defer os.Remove(tempFile)

	// Load and draw the image
	pdfImage, err := gopdf.LoadJPEGFile(tempFile)
	if err != nil {
		return fmt.Errorf("failed to load image: %w", err)
	}
	if err := page.DrawImage(pdfImage, 0, 0, gopdf.PageSizeA4.Width, gopdf.PageSizeA4.Height); err != nil {
		return fmt.Errorf("failed to draw image: %w", err)
	}

	// Simulate OCR results (in a real application, this would come from an OCR API)
	ocrResult := gopdf.OCRResult{
		Text: "Document Title First paragraph of text. Second paragraph of text.",
		Words: []gopdf.OCRWord{
			// Title (pixels from top-left)
			{Text: "Document", Confidence: 0.99, Bounds: gopdf.Rectangle{X: 100, Y: 50, Width: 150, Height: 30}},
			{Text: "Title", Confidence: 0.99, Bounds: gopdf.Rectangle{X: 260, Y: 50, Width: 100, Height: 30}},
			// First paragraph
			{Text: "First", Confidence: 0.98, Bounds: gopdf.Rectangle{X: 100, Y: 150, Width: 80, Height: 20}},
			{Text: "paragraph", Confidence: 0.98, Bounds: gopdf.Rectangle{X: 190, Y: 150, Width: 140, Height: 20}},
			{Text: "of", Confidence: 0.99, Bounds: gopdf.Rectangle{X: 340, Y: 150, Width: 30, Height: 20}},
			{Text: "text.", Confidence: 0.98, Bounds: gopdf.Rectangle{X: 380, Y: 150, Width: 70, Height: 20}},
			// Second paragraph
			{Text: "Second", Confidence: 0.97, Bounds: gopdf.Rectangle{X: 100, Y: 200, Width: 100, Height: 20}},
			{Text: "paragraph", Confidence: 0.98, Bounds: gopdf.Rectangle{X: 210, Y: 200, Width: 140, Height: 20}},
			{Text: "of", Confidence: 0.99, Bounds: gopdf.Rectangle{X: 360, Y: 200, Width: 30, Height: 20}},
			{Text: "text.", Confidence: 0.98, Bounds: gopdf.Rectangle{X: 400, Y: 200, Width: 70, Height: 20}},
		},
	}

	// Convert pixel coordinates to PDF coordinates
	// Image size: 800x600 pixels, PDF page: A4 (595x842 points)
	imageWidth := 800
	imageHeight := 600
	textLayer := ocrResult.ToTextLayer(imageWidth, imageHeight, gopdf.PageSizeA4.Width, gopdf.PageSizeA4.Height)

	// Add text layer to page
	if err := page.AddTextLayer(textLayer); err != nil {
		return fmt.Errorf("failed to add text layer: %w", err)
	}

	// Save
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if err := doc.WriteTo(file); err != nil {
		return fmt.Errorf("failed to write PDF: %w", err)
	}

	fmt.Printf("  Created: %s\n", filename)
	fmt.Println("  OCR simulation: Text layer added from OCRResult")
	fmt.Println("  Try searching for 'paragraph' in the PDF!")
	return nil
}

// createMultipleWordsExample demonstrates manual word positioning
func createMultipleWordsExample(filename string) error {
	doc := gopdf.New()
	page := doc.AddPage(gopdf.PageSizeA4, gopdf.Portrait)

	// Create image
	img := createSampleImage(
		"Line 1: The quick brown fox",
		"Line 2: jumps over the lazy dog",
	)
	tempFile := "temp_multi_image.jpg"
	if err := saveImageAsJPEG(img, tempFile); err != nil {
		return fmt.Errorf("failed to save temp image: %w", err)
	}
	defer os.Remove(tempFile)

	// Draw image
	pdfImage, err := gopdf.LoadJPEGFile(tempFile)
	if err != nil {
		return fmt.Errorf("failed to load image: %w", err)
	}
	if err := page.DrawImage(pdfImage, 0, 0, gopdf.PageSizeA4.Width, gopdf.PageSizeA4.Height); err != nil {
		return fmt.Errorf("failed to draw image: %w", err)
	}

	// Manually create word boundaries (in a real app, these come from OCR)
	words := []gopdf.TextLayerWord{
		// Line 1
		{Text: "Line", Bounds: gopdf.Rectangle{X: 50, Y: 750, Width: 40, Height: 12}},
		{Text: "1:", Bounds: gopdf.Rectangle{X: 95, Y: 750, Width: 15, Height: 12}},
		{Text: "The", Bounds: gopdf.Rectangle{X: 115, Y: 750, Width: 35, Height: 12}},
		{Text: "quick", Bounds: gopdf.Rectangle{X: 155, Y: 750, Width: 45, Height: 12}},
		{Text: "brown", Bounds: gopdf.Rectangle{X: 205, Y: 750, Width: 50, Height: 12}},
		{Text: "fox", Bounds: gopdf.Rectangle{X: 260, Y: 750, Width: 30, Height: 12}},
		// Line 2
		{Text: "Line", Bounds: gopdf.Rectangle{X: 50, Y: 700, Width: 40, Height: 12}},
		{Text: "2:", Bounds: gopdf.Rectangle{X: 95, Y: 700, Width: 15, Height: 12}},
		{Text: "jumps", Bounds: gopdf.Rectangle{X: 115, Y: 700, Width: 50, Height: 12}},
		{Text: "over", Bounds: gopdf.Rectangle{X: 170, Y: 700, Width: 40, Height: 12}},
		{Text: "the", Bounds: gopdf.Rectangle{X: 215, Y: 700, Width: 30, Height: 12}},
		{Text: "lazy", Bounds: gopdf.Rectangle{X: 250, Y: 700, Width: 40, Height: 12}},
		{Text: "dog", Bounds: gopdf.Rectangle{X: 295, Y: 700, Width: 35, Height: 12}},
	}

	// Add text layer
	if err := page.AddTextLayerWords(words); err != nil {
		return fmt.Errorf("failed to add text layer: %w", err)
	}

	// Save
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if err := doc.WriteTo(file); err != nil {
		return fmt.Errorf("failed to write PDF: %w", err)
	}

	fmt.Printf("  Created: %s\n", filename)
	fmt.Println("  Multiple words with precise positioning")
	fmt.Println("  Try searching for individual words like 'fox' or 'lazy'!")
	return nil
}

// createSampleImage creates a simple image with text
func createSampleImage(lines ...string) image.Image {
	width := 800
	height := 600

	// Create image
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill background with light gray
	bg := color.RGBA{240, 240, 240, 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{bg}, image.Point{}, draw.Src)

	// Draw border
	drawRect(img, 10, 10, width-20, height-20, color.RGBA{200, 200, 200, 255})

	// Note: In a real application, you would render actual text here
	// For simplicity, we're just creating a placeholder image
	// The actual text will be in the invisible text layer

	return img
}

// drawRect draws a rectangle outline
func drawRect(img *image.RGBA, x, y, w, h int, c color.Color) {
	// Top
	for i := x; i < x+w; i++ {
		img.Set(i, y, c)
	}
	// Bottom
	for i := x; i < x+w; i++ {
		img.Set(i, y+h, c)
	}
	// Left
	for i := y; i < y+h; i++ {
		img.Set(x, i, c)
	}
	// Right
	for i := y; i < y+h; i++ {
		img.Set(x+w, i, c)
	}
}

// saveImageAsJPEG saves an image as JPEG
func saveImageAsJPEG(img image.Image, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return jpeg.Encode(file, img, &jpeg.Options{Quality: 90})
}
