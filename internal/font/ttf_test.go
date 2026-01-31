package font

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// getTestFontPath returns a path to a TTF font for testing
// It tries common system font locations based on the OS
func getTestFontPath() string {
	switch runtime.GOOS {
	case "darwin":
		// macOS system fonts
		paths := []string{
			"/System/Library/Fonts/Helvetica.ttc",
			"/System/Library/Fonts/Supplemental/Arial.ttf",
			"/Library/Fonts/Arial.ttf",
		}
		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	case "linux":
		// Linux system fonts
		paths := []string{
			"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
			"/usr/share/fonts/TTF/DejaVuSans.ttf",
			"/usr/share/fonts/liberation/LiberationSans-Regular.ttf",
		}
		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	case "windows":
		// Windows system fonts
		paths := []string{
			filepath.Join(os.Getenv("WINDIR"), "Fonts", "arial.ttf"),
			"C:\\Windows\\Fonts\\arial.ttf",
		}
		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	}
	return ""
}

func TestLoadTTF(t *testing.T) {
	fontPath := getTestFontPath()
	if fontPath == "" {
		t.Skip("No test font available on this system")
	}

	font, err := LoadTTF(fontPath)
	if err != nil {
		t.Fatalf("LoadTTF failed: %v", err)
	}

	if font == nil {
		t.Fatal("font is nil")
	}

	if font.Name() == "" {
		t.Error("font name is empty")
	}

	if len(font.Data()) == 0 {
		t.Error("font data is empty")
	}

	if font.Font() == nil {
		t.Error("sfnt.Font is nil")
	}
}

func TestLoadTTFFromBytes(t *testing.T) {
	fontPath := getTestFontPath()
	if fontPath == "" {
		t.Skip("No test font available on this system")
	}

	// Read font file
	data, err := os.ReadFile(fontPath)
	if err != nil {
		t.Fatalf("Failed to read font file: %v", err)
	}

	// Load from bytes
	font, err := LoadTTFFromBytes(data)
	if err != nil {
		t.Fatalf("LoadTTFFromBytes failed: %v", err)
	}

	if font == nil {
		t.Fatal("font is nil")
	}

	if font.Name() == "" {
		t.Error("font name is empty")
	}
}

func TestLoadTTF_InvalidPath(t *testing.T) {
	_, err := LoadTTF("/nonexistent/path/font.ttf")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestLoadTTFFromBytes_InvalidData(t *testing.T) {
	invalidData := []byte("this is not a valid TTF file")

	_, err := LoadTTFFromBytes(invalidData)
	if err == nil {
		t.Error("Expected error for invalid TTF data, got nil")
	}
}

func TestTTFFont_Name(t *testing.T) {
	fontPath := getTestFontPath()
	if fontPath == "" {
		t.Skip("No test font available on this system")
	}

	font, err := LoadTTF(fontPath)
	if err != nil {
		t.Fatalf("LoadTTF failed: %v", err)
	}

	name := font.Name()
	if name == "" {
		t.Error("Font name should not be empty")
	}

	t.Logf("Font name: %s", name)
}

func TestTTFFont_GlyphWidth(t *testing.T) {
	fontPath := getTestFontPath()
	if fontPath == "" {
		t.Skip("No test font available on this system")
	}

	font, err := LoadTTF(fontPath)
	if err != nil {
		t.Fatalf("LoadTTF failed: %v", err)
	}

	tests := []struct {
		char     rune
		fontSize float64
	}{
		{'A', 12.0},
		{'B', 12.0},
		{'a', 12.0},
		{'1', 12.0},
		{'W', 24.0},
		{'i', 24.0},
	}

	for _, tt := range tests {
		width, err := font.GlyphWidth(tt.char, tt.fontSize)
		if err != nil {
			t.Errorf("GlyphWidth(%c, %.1f) error: %v", tt.char, tt.fontSize, err)
			continue
		}

		if width <= 0 {
			t.Errorf("GlyphWidth(%c, %.1f) = %.2f, expected positive value", tt.char, tt.fontSize, width)
		}

		t.Logf("GlyphWidth(%c, %.1f) = %.2f", tt.char, tt.fontSize, width)
	}
}

func TestTTFFont_GlyphWidth_Caching(t *testing.T) {
	fontPath := getTestFontPath()
	if fontPath == "" {
		t.Skip("No test font available on this system")
	}

	font, err := LoadTTF(fontPath)
	if err != nil {
		t.Fatalf("LoadTTF failed: %v", err)
	}

	char := 'A'
	fontSize := 12.0

	// First call - should cache the glyph
	width1, err1 := font.GlyphWidth(char, fontSize)
	if err1 != nil {
		t.Fatalf("First GlyphWidth call failed: %v", err1)
	}

	// Second call - should use cached glyph
	width2, err2 := font.GlyphWidth(char, fontSize)
	if err2 != nil {
		t.Fatalf("Second GlyphWidth call failed: %v", err2)
	}

	if width1 != width2 {
		t.Errorf("GlyphWidth not consistent: first=%.2f, second=%.2f", width1, width2)
	}
}

func TestTTFFont_TextWidth(t *testing.T) {
	fontPath := getTestFontPath()
	if fontPath == "" {
		t.Skip("No test font available on this system")
	}

	font, err := LoadTTF(fontPath)
	if err != nil {
		t.Fatalf("LoadTTF failed: %v", err)
	}

	tests := []struct {
		text     string
		fontSize float64
	}{
		{"Hello", 12.0},
		{"World", 12.0},
		{"Test", 24.0},
		{"ABC", 18.0},
	}

	for _, tt := range tests {
		width, err := font.TextWidth(tt.text, tt.fontSize)
		if err != nil {
			t.Errorf("TextWidth(%q, %.1f) error: %v", tt.text, tt.fontSize, err)
			continue
		}

		if width <= 0 {
			t.Errorf("TextWidth(%q, %.1f) = %.2f, expected positive value", tt.text, tt.fontSize, width)
		}

		t.Logf("TextWidth(%q, %.1f) = %.2f", tt.text, tt.fontSize, width)
	}
}

func TestTTFFont_GlyphWidthBatch(t *testing.T) {
	fontPath := getTestFontPath()
	if fontPath == "" {
		t.Skip("No test font available on this system")
	}

	font, err := LoadTTF(fontPath)
	if err != nil {
		t.Fatalf("LoadTTF failed: %v", err)
	}

	runes := []rune{'H', 'e', 'l', 'l', 'o'}
	fontSize := 12.0

	widths, err := font.GlyphWidthBatch(runes, fontSize)
	if err != nil {
		t.Fatalf("GlyphWidthBatch failed: %v", err)
	}

	if len(widths) != len(runes) {
		t.Fatalf("Expected %d widths, got %d", len(runes), len(widths))
	}

	for i, width := range widths {
		if width <= 0 {
			t.Errorf("Width for rune %c is not positive: %.2f", runes[i], width)
		}
		t.Logf("Width(%c) = %.2f", runes[i], width)
	}
}

func TestTTFFont_UnicodeCharacters(t *testing.T) {
	fontPath := getTestFontPath()
	if fontPath == "" {
		t.Skip("No test font available on this system")
	}

	font, err := LoadTTF(fontPath)
	if err != nil {
		t.Fatalf("LoadTTF failed: %v", err)
	}

	// Test various Unicode characters
	tests := []struct {
		name     string
		char     rune
		fontSize float64
	}{
		{"Latin", 'A', 12.0},
		{"Number", '1', 12.0},
		{"Symbol", '@', 12.0},
		{"Euro", '€', 12.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			width, err := font.GlyphWidth(tt.char, tt.fontSize)
			if err != nil {
				// Some fonts may not have all characters
				t.Logf("GlyphWidth(%c) not available: %v", tt.char, err)
				return
			}

			if width <= 0 {
				t.Errorf("GlyphWidth(%c) = %.2f, expected positive value", tt.char, width)
			}

			t.Logf("GlyphWidth(%c, %.1f) = %.2f", tt.char, tt.fontSize, width)
		})
	}
}
