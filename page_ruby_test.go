package gopdf

import (
	"os"
	"testing"

	"github.com/ryomak/gopdf/internal/font"
)

func TestPage_DrawRuby(t *testing.T) {
	// Create a test document
	doc := New()
	page := doc.AddPage(A4, Portrait)

	// Set a font
	if err := page.SetFont(font.Helvetica, 12); err != nil {
		t.Fatalf("SetFont failed: %v", err)
	}

	// Create ruby text
	rubyText := NewRubyText("Test", "test")
	style := DefaultRubyStyle()

	// Draw ruby
	width, err := page.DrawRuby(rubyText, 50, 700, style)
	if err != nil {
		t.Errorf("DrawRuby failed: %v", err)
	}

	if width <= 0 {
		t.Errorf("DrawRuby returned width = %f, want > 0", width)
	}
}

func TestPage_DrawRuby_NoFont(t *testing.T) {
	// Create a test document without setting font
	doc := New()
	page := doc.AddPage(A4, Portrait)

	rubyText := NewRubyText("Test", "test")
	style := DefaultRubyStyle()

	// Should fail because no font is set
	_, err := page.DrawRuby(rubyText, 50, 700, style)
	if err == nil {
		t.Error("DrawRuby should fail when no font is set")
	}
}

func TestPage_DrawRubyWithActualText(t *testing.T) {
	// Create a test document
	doc := New()
	page := doc.AddPage(A4, Portrait)

	// Set a font
	if err := page.SetFont(font.Helvetica, 12); err != nil {
		t.Fatalf("SetFont failed: %v", err)
	}

	// Create ruby text
	rubyText := NewRubyText("Test", "test")

	tests := []struct {
		name     string
		copyMode RubyCopyMode
	}{
		{"CopyBase", RubyCopyBase},
		{"CopyRuby", RubyCopyRuby},
		{"CopyBoth", RubyCopyBoth},
	}

	y := 700.0
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := DefaultRubyStyle()
			style.CopyMode = tt.copyMode

			width, err := page.DrawRubyWithActualText(rubyText, 50, y, style)
			if err != nil {
				t.Errorf("DrawRubyWithActualText failed: %v", err)
			}
			if width <= 0 {
				t.Errorf("DrawRubyWithActualText returned width = %f, want > 0", width)
			}
			y -= 50 // Move down for next test
		})
	}
}

func TestPage_DrawRubyTexts(t *testing.T) {
	// Create a test document
	doc := New()
	page := doc.AddPage(A4, Portrait)

	// Set a font
	if err := page.SetFont(font.Helvetica, 12); err != nil {
		t.Fatalf("SetFont failed: %v", err)
	}

	// Create multiple ruby texts
	texts := NewRubyTextPairs("Test1", "test1", "Test2", "test2", "Test3", "test3")
	style := DefaultRubyStyle()

	// Draw without ActualText
	totalWidth, err := page.DrawRubyTexts(texts, 50, 700, style, false)
	if err != nil {
		t.Errorf("DrawRubyTexts (no ActualText) failed: %v", err)
	}
	if totalWidth <= 0 {
		t.Errorf("DrawRubyTexts returned totalWidth = %f, want > 0", totalWidth)
	}

	// Draw with ActualText
	totalWidth2, err := page.DrawRubyTexts(texts, 50, 650, style, true)
	if err != nil {
		t.Errorf("DrawRubyTexts (with ActualText) failed: %v", err)
	}
	if totalWidth2 <= 0 {
		t.Errorf("DrawRubyTexts returned totalWidth = %f, want > 0", totalWidth2)
	}
}

func TestPage_DrawRuby_Alignments(t *testing.T) {
	// Create a test document
	doc := New()
	page := doc.AddPage(A4, Portrait)

	// Set a font
	if err := page.SetFont(font.Helvetica, 12); err != nil {
		t.Fatalf("SetFont failed: %v", err)
	}

	rubyText := NewRubyText("Test", "test")

	tests := []struct {
		name      string
		alignment RubyAlignment
	}{
		{"Center", RubyAlignCenter},
		{"Left", RubyAlignLeft},
		{"Right", RubyAlignRight},
	}

	y := 700.0
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := DefaultRubyStyle()
			style.Alignment = tt.alignment

			width, err := page.DrawRuby(rubyText, 50, y, style)
			if err != nil {
				t.Errorf("DrawRuby with %s alignment failed: %v", tt.name, err)
			}
			if width <= 0 {
				t.Errorf("DrawRuby returned width = %f, want > 0", width)
			}
			y -= 50 // Move down for next test
		})
	}
}

func TestPage_DrawRuby_SizeRatios(t *testing.T) {
	// Create a test document
	doc := New()
	page := doc.AddPage(A4, Portrait)

	// Set a font
	if err := page.SetFont(font.Helvetica, 12); err != nil {
		t.Fatalf("SetFont failed: %v", err)
	}

	rubyText := NewRubyText("Test", "test")

	tests := []struct {
		name      string
		sizeRatio float64
	}{
		{"30%", 0.3},
		{"50%", 0.5},
		{"70%", 0.7},
	}

	y := 700.0
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := DefaultRubyStyle()
			style.SizeRatio = tt.sizeRatio

			width, err := page.DrawRuby(rubyText, 50, y, style)
			if err != nil {
				t.Errorf("DrawRuby with size ratio %f failed: %v", tt.sizeRatio, err)
			}
			if width <= 0 {
				t.Errorf("DrawRuby returned width = %f, want > 0", width)
			}
			y -= 50 // Move down for next test
		})
	}
}

func TestPage_GetCurrentFontName(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*Page) error
		expected string
	}{
		{
			name: "Standard font",
			setup: func(p *Page) error {
				return p.SetFont(font.Helvetica, 12)
			},
			expected: "F1",
		},
		{
			name: "Bold font",
			setup: func(p *Page) error {
				return p.SetFont(font.HelveticaBold, 12)
			},
			expected: "F2",
		},
		{
			name: "No font set",
			setup: func(p *Page) error {
				return nil
			},
			expected: "F1", // Default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := New()
			page := doc.AddPage(A4, Portrait)

			if err := tt.setup(page); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			fontName := page.getCurrentFontName()
			if fontName != tt.expected {
				t.Errorf("getCurrentFontName() = %q, want %q", fontName, tt.expected)
			}
		})
	}
}

// Integration test: Create a PDF with ruby annotations
func TestPage_DrawRuby_Integration(t *testing.T) {
	// Skip if in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a test document
	doc := New()
	page := doc.AddPage(A4, Portrait)

	// Set font
	if err := page.SetFont(font.Helvetica, 16); err != nil {
		t.Fatalf("SetFont failed: %v", err)
	}

	// Draw title
	if err := page.DrawText("Ruby Annotation Test", 50, 800); err != nil {
		t.Fatalf("DrawText failed: %v", err)
	}

	// Draw ruby examples
	style := DefaultRubyStyle()
	rubyTexts := []RubyText{
		NewRubyText("Test1", "test1"),
		NewRubyText("Test2", "test2"),
		NewRubyText("Test3", "test3"),
	}

	y := 750.0
	for _, rt := range rubyTexts {
		_, err := page.DrawRuby(rt, 50, y, style)
		if err != nil {
			t.Errorf("DrawRuby failed: %v", err)
		}
		y -= 50
	}

	// Write to temp file
	tmpFile, err := os.CreateTemp("", "test_ruby_*.pdf")
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
