package gopdf

import (
	"bytes"
	"os"
	"runtime"
	"testing"
)

// getTestTTFPath returns a path to a system TTF font for testing
func getTestTTFPath() string {
	switch runtime.GOOS {
	case "darwin":
		paths := []string{
			"/System/Library/Fonts/Helvetica.ttc",
			"/System/Library/Fonts/Supplemental/Arial.ttf",
		}
		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	case "linux":
		paths := []string{
			"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
			"/usr/share/fonts/TTF/DejaVuSans.ttf",
		}
		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	case "windows":
		paths := []string{
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
	fontPath := getTestTTFPath()
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

	t.Logf("Loaded font: %s", font.Name())
}

func TestLoadTTFFromBytes(t *testing.T) {
	fontPath := getTestTTFPath()
	if fontPath == "" {
		t.Skip("No test font available on this system")
	}

	data, err := os.ReadFile(fontPath)
	if err != nil {
		t.Fatalf("Failed to read font file: %v", err)
	}

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

func TestTTFFont_TextWidth(t *testing.T) {
	fontPath := getTestTTFPath()
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
		{"World", 18.0},
		{"Test", 24.0},
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

func TestPage_SetTTFFont(t *testing.T) {
	fontPath := getTestTTFPath()
	if fontPath == "" {
		t.Skip("No test font available on this system")
	}

	doc := New()
	page := doc.AddPage(A4, Portrait)

	font, err := LoadTTF(fontPath)
	if err != nil {
		t.Fatalf("LoadTTF failed: %v", err)
	}

	err = page.SetTTFFont(font, 12.0)
	if err != nil {
		t.Fatalf("SetTTFFont failed: %v", err)
	}

	if page.currentTTFFont == nil {
		t.Error("currentTTFFont should be set")
	}

	if page.fontSize != 12.0 {
		t.Errorf("fontSize = %.1f, want 12.0", page.fontSize)
	}
}

func TestPage_DrawTextUTF8(t *testing.T) {
	fontPath := getTestTTFPath()
	if fontPath == "" {
		t.Skip("No test font available on this system")
	}

	doc := New()
	page := doc.AddPage(A4, Portrait)

	font, err := LoadTTF(fontPath)
	if err != nil {
		t.Fatalf("LoadTTF failed: %v", err)
	}

	err = page.SetTTFFont(font, 12.0)
	if err != nil {
		t.Fatalf("SetTTFFont failed: %v", err)
	}

	err = page.DrawTextUTF8("Hello, World!", 100, 700)
	if err != nil {
		t.Fatalf("DrawTextUTF8 failed: %v", err)
	}

	// Check that content was written
	content := page.content.String()
	if content == "" {
		t.Error("page content is empty")
	}

	// Check for expected PDF operators
	if !contains(content, "BT") {
		t.Error("content should contain BT (Begin Text)")
	}
	if !contains(content, "ET") {
		t.Error("content should contain ET (End Text)")
	}
	if !contains(content, "Tf") {
		t.Error("content should contain Tf (Set Font)")
	}
	if !contains(content, "Tj") {
		t.Error("content should contain Tj (Show Text)")
	}
}

func TestTTFFont_PDFGeneration(t *testing.T) {
	fontPath := getTestTTFPath()
	if fontPath == "" {
		t.Skip("No test font available on this system")
	}

	doc := New()
	page := doc.AddPage(A4, Portrait)

	font, err := LoadTTF(fontPath)
	if err != nil {
		t.Fatalf("LoadTTF failed: %v", err)
	}

	err = page.SetTTFFont(font, 18.0)
	if err != nil {
		t.Fatalf("SetTTFFont failed: %v", err)
	}

	err = page.DrawTextUTF8("Hello, World!", 100, 750)
	if err != nil {
		t.Fatalf("DrawTextUTF8 failed: %v", err)
	}

	err = page.DrawTextUTF8("Unicode: € £ ¥", 100, 720)
	if err != nil {
		t.Fatalf("DrawTextUTF8 failed: %v", err)
	}

	// Write to buffer
	var buf bytes.Buffer
	err = doc.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo failed: %v", err)
	}

	// Check PDF header
	output := buf.Bytes()
	if len(output) == 0 {
		t.Fatal("PDF output is empty")
	}

	if !bytes.HasPrefix(output, []byte("%PDF-1.7")) {
		t.Error("PDF should start with %PDF-1.7 header")
	}

	// Check for embedded font
	pdfStr := string(output)
	if !contains(pdfStr, "/Type /Font") {
		t.Error("PDF should contain font definition")
	}
	if !contains(pdfStr, "/Type0") {
		t.Error("PDF should contain Type0 font")
	}
	if !contains(pdfStr, "/CIDFontType2") {
		t.Error("PDF should contain CIDFontType2")
	}

	t.Logf("Generated PDF size: %d bytes", len(output))
}

func TestTTFFont_UnicodeText(t *testing.T) {
	fontPath := getTestTTFPath()
	if fontPath == "" {
		t.Skip("No test font available on this system")
	}

	doc := New()
	page := doc.AddPage(A4, Portrait)

	font, err := LoadTTF(fontPath)
	if err != nil {
		t.Fatalf("LoadTTF failed: %v", err)
	}

	err = page.SetTTFFont(font, 14.0)
	if err != nil {
		t.Fatalf("SetTTFFont failed: %v", err)
	}

	// Test various Unicode characters
	tests := []string{
		"Hello",
		"€ © ® ™",
		"Ñ ñ Ü ü",
		"Α Β Γ Δ", // Greek
	}

	y := 750.0
	for _, text := range tests {
		err = page.DrawTextUTF8(text, 100, y)
		if err != nil {
			t.Errorf("DrawTextUTF8(%q) failed: %v", text, err)
		}
		y -= 30
	}

	// Write to buffer
	var buf bytes.Buffer
	err = doc.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo failed: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("PDF output is empty")
	}
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}
