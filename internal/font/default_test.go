package font

import (
	"runtime"
	"testing"
)

func TestLoadSystemJapaneseFont(t *testing.T) {
	font, err := LoadSystemJapaneseFont()

	// On CI or systems without Japanese fonts, this may fail
	if err != nil {
		t.Skipf("Skipping test: %v (no system Japanese font available)", err)
	}

	if font == nil {
		t.Fatal("LoadSystemJapaneseFont() returned nil")
	}

	// Check font name
	name := font.Name()
	if name == "" {
		t.Error("Font name is empty")
	}
	t.Logf("Found system Japanese font: %s", name)
}

func TestLoadSystemJapaneseFont_TextWidth(t *testing.T) {
	font, err := LoadSystemJapaneseFont()
	if err != nil {
		t.Skipf("Skipping test: %v (no system Japanese font available)", err)
	}

	tests := []struct {
		name     string
		text     string
		fontSize float64
	}{
		{"English text", "Hello, World!", 12.0},
		{"Japanese text", "こんにちは、世界！", 12.0},
		{"Mixed text", "Hello, 世界！", 12.0},
		{"Empty string", "", 12.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			width, err := font.TextWidth(tt.text, tt.fontSize)
			if err != nil {
				t.Errorf("TextWidth() error = %v", err)
				return
			}
			if tt.text == "" && width != 0 {
				t.Errorf("TextWidth() for empty string = %v, want 0", width)
			}
			if tt.text != "" && width <= 0 {
				t.Errorf("TextWidth() = %v, want > 0", width)
			}
			t.Logf("Text '%s' width at %.1fpt: %.2f", tt.text, tt.fontSize, width)
		})
	}
}

func TestGetSystemJapaneseFontPaths(t *testing.T) {
	paths := getSystemJapaneseFontPaths()

	switch runtime.GOOS {
	case "darwin", "linux", "windows":
		if len(paths) == 0 {
			t.Errorf("Expected font paths for %s, got none", runtime.GOOS)
		}
		t.Logf("Font search paths for %s:", runtime.GOOS)
		for _, p := range paths {
			t.Logf("  - %s", p)
		}
	default:
		if paths != nil {
			t.Errorf("Expected nil paths for unsupported OS %s", runtime.GOOS)
		}
	}
}

// Benchmark: System font loading
func BenchmarkLoadSystemJapaneseFont(b *testing.B) {
	// First load to check if font is available
	_, err := LoadSystemJapaneseFont()
	if err != nil {
		b.Skipf("Skipping benchmark: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Note: This will load from disk each time
		_, _ = LoadSystemJapaneseFont()
	}
}
