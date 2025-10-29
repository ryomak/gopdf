package gopdf

import (
	"os"
	"testing"

)

func TestPage_AddInvisibleText(t *testing.T) {
	doc := New()
	page := doc.AddPage(PageSizeA4, Portrait)

	// Set font
	if err := page.SetFont(FontHelvetica, 12); err != nil {
		t.Fatalf("SetFont failed: %v", err)
	}

	// Add invisible text
	err := page.AddInvisibleText("Hello World", 100, 700, 200, 20)
	if err != nil {
		t.Errorf("AddInvisibleText failed: %v", err)
	}
}

func TestPage_AddTextLayer(t *testing.T) {
	doc := New()
	page := doc.AddPage(PageSizeA4, Portrait)

	// Set font
	if err := page.SetFont(FontHelvetica, 12); err != nil {
		t.Fatalf("SetFont failed: %v", err)
	}

	// Create text layer
	layer := TextLayer{
		Words: []TextLayerWord{
			{Text: "Hello", Bounds: Rectangle{X: 100, Y: 700, Width: 50, Height: 12}},
			{Text: "World", Bounds: Rectangle{X: 160, Y: 700, Width: 50, Height: 12}},
		},
		RenderMode: TextRenderInvisible,
		Opacity:    0.0,
	}

	// Add text layer
	err := page.AddTextLayer(layer)
	if err != nil {
		t.Errorf("AddTextLayer failed: %v", err)
	}
}

func TestPage_AddTextLayer_EmptyWords(t *testing.T) {
	doc := New()
	page := doc.AddPage(PageSizeA4, Portrait)

	// Set font
	if err := page.SetFont(FontHelvetica, 12); err != nil {
		t.Fatalf("SetFont failed: %v", err)
	}

	// Create empty text layer
	layer := TextLayer{
		Words:      []TextLayerWord{},
		RenderMode: TextRenderInvisible,
		Opacity:    0.0,
	}

	// Should not fail
	err := page.AddTextLayer(layer)
	if err != nil {
		t.Errorf("AddTextLayer with empty words failed: %v", err)
	}
}

func TestPage_AddTextLayer_NoFont(t *testing.T) {
	doc := New()
	page := doc.AddPage(PageSizeA4, Portrait)

	// Don't set font - should use default

	layer := TextLayer{
		Words: []TextLayerWord{
			{Text: "Test", Bounds: Rectangle{X: 100, Y: 700, Width: 50, Height: 12}},
		},
		RenderMode: TextRenderInvisible,
		Opacity:    0.0,
	}

	// Should not fail (will use default font)
	err := page.AddTextLayer(layer)
	if err != nil {
		t.Errorf("AddTextLayer without font failed: %v", err)
	}
}

func TestPage_AddTextLayerWords(t *testing.T) {
	doc := New()
	page := doc.AddPage(PageSizeA4, Portrait)

	// Set font
	if err := page.SetFont(FontHelvetica, 12); err != nil {
		t.Fatalf("SetFont failed: %v", err)
	}

	// Create words
	words := []TextLayerWord{
		{Text: "First", Bounds: Rectangle{X: 100, Y: 700, Width: 40, Height: 12}},
		{Text: "Second", Bounds: Rectangle{X: 150, Y: 700, Width: 50, Height: 12}},
		{Text: "Third", Bounds: Rectangle{X: 210, Y: 700, Width: 40, Height: 12}},
	}

	// Add words
	err := page.AddTextLayerWords(words)
	if err != nil {
		t.Errorf("AddTextLayerWords failed: %v", err)
	}
}

func TestPage_AddTextLayer_DifferentRenderModes(t *testing.T) {
	doc := New()
	page := doc.AddPage(PageSizeA4, Portrait)

	// Set font
	if err := page.SetFont(FontHelvetica, 12); err != nil {
		t.Fatalf("SetFont failed: %v", err)
	}

	tests := []struct {
		name       string
		renderMode TextRenderMode
	}{
		{"Normal", TextRenderNormal},
		{"Stroke", TextRenderStroke},
		{"FillStroke", TextRenderFillStroke},
		{"Invisible", TextRenderInvisible},
	}

	y := 700.0
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layer := TextLayer{
				Words: []TextLayerWord{
					{Text: "Test", Bounds: Rectangle{X: 100, Y: y, Width: 50, Height: 12}},
				},
				RenderMode: tt.renderMode,
				Opacity:    1.0,
			}

			err := page.AddTextLayer(layer)
			if err != nil {
				t.Errorf("AddTextLayer with %s mode failed: %v", tt.name, err)
			}
			y -= 30
		})
	}
}

func TestPage_AddTextLayer_WithOpacity(t *testing.T) {
	doc := New()
	page := doc.AddPage(PageSizeA4, Portrait)

	// Set font
	if err := page.SetFont(FontHelvetica, 12); err != nil {
		t.Fatalf("SetFont failed: %v", err)
	}

	tests := []struct {
		name    string
		opacity float64
	}{
		{"Transparent", 0.0},
		{"Half", 0.5},
		{"Opaque", 1.0},
	}

	y := 700.0
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layer := TextLayer{
				Words: []TextLayerWord{
					{Text: "Test", Bounds: Rectangle{X: 100, Y: y, Width: 50, Height: 12}},
				},
				RenderMode: TextRenderNormal,
				Opacity:    tt.opacity,
			}

			err := page.AddTextLayer(layer)
			if err != nil {
				t.Errorf("AddTextLayer with opacity %f failed: %v", tt.opacity, err)
			}
			y -= 30
		})
	}
}

// Integration test: Create a PDF with text layer
func TestPage_AddTextLayer_Integration(t *testing.T) {
	// Skip if in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create document
	doc := New()
	page := doc.AddPage(PageSizeA4, Portrait)

	// Set font
	if err := page.SetFont(FontHelvetica, 12); err != nil {
		t.Fatalf("SetFont failed: %v", err)
	}

	// Add title (visible)
	if err := page.DrawText("Text Layer Integration Test", 50, 800); err != nil {
		t.Fatalf("DrawText failed: %v", err)
	}

	// Add invisible text layer (simulating OCR result)
	words := []TextLayerWord{
		{Text: "This", Bounds: Rectangle{X: 50, Y: 700, Width: 30, Height: 12}},
		{Text: "text", Bounds: Rectangle{X: 85, Y: 700, Width: 30, Height: 12}},
		{Text: "is", Bounds: Rectangle{X: 120, Y: 700, Width: 15, Height: 12}},
		{Text: "invisible", Bounds: Rectangle{X: 140, Y: 700, Width: 60, Height: 12}},
		{Text: "but", Bounds: Rectangle{X: 205, Y: 700, Width: 25, Height: 12}},
		{Text: "searchable", Bounds: Rectangle{X: 235, Y: 700, Width: 80, Height: 12}},
	}

	layer := NewTextLayer(words)
	if err := page.AddTextLayer(layer); err != nil {
		t.Fatalf("AddTextLayer failed: %v", err)
	}

	// Add another line using AddInvisibleText
	if err := page.AddInvisibleText("Second line of invisible text", 50, 650, 250, 12); err != nil {
		t.Fatalf("AddInvisibleText failed: %v", err)
	}

	// Write to temp file
	tmpFile, err := os.CreateTemp("", "test_text_layer_*.pdf")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if err := doc.WriteTo(tmpFile); err != nil {
		t.Fatalf("WriteTo failed: %v", err)
	}

	// Check file size
	stat, err := tmpFile.Stat()
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	if stat.Size() == 0 {
		t.Error("Generated PDF is empty")
	}

	t.Logf("Created test PDF: %s (size: %d bytes)", tmpFile.Name(), stat.Size())
}
