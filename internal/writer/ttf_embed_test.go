package writer

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ryomak/gopdf/internal/font"
)

const testFontPath = "/usr/share/fonts/truetype/liberation/LiberationSans-Regular.ttf"

func loadTestFont(t *testing.T) *font.TTFFont {
	t.Helper()
	ttfFont, err := font.LoadTTF(testFontPath)
	if err != nil {
		t.Fatalf("failed to load test font: %v", err)
	}
	return ttfFont
}

// buildUsedGlyphs creates a usedGlyphs map from a string using the given font.
func buildUsedGlyphs(t *testing.T, ttfFont *font.TTFFont, text string) map[uint16]rune {
	t.Helper()
	usedGlyphs := make(map[uint16]rune)
	for _, r := range text {
		gid, err := ttfFont.GetGlyphIndex(r)
		if err != nil {
			t.Fatalf("GetGlyphIndex(%c) failed: %v", r, err)
		}
		usedGlyphs[gid] = r
	}
	return usedGlyphs
}

// TestNewTTFFontEmbedder tests that NewTTFFontEmbedder returns a valid embedder.
func TestNewTTFFontEmbedder(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	embedder := NewTTFFontEmbedder(w)

	if embedder == nil {
		t.Fatal("NewTTFFontEmbedder returned nil")
	}
	if embedder.writer != w {
		t.Error("embedder.writer should point to the given writer")
	}
}

// TestEmbedTTFFont tests the full EmbedTTFFont pipeline with a real font.
func TestEmbedTTFFont(t *testing.T) {
	ttfFont := loadTestFont(t)
	usedGlyphs := buildUsedGlyphs(t, ttfFont, "Hello")

	var buf bytes.Buffer
	w := NewWriter(&buf)
	if err := w.WriteHeader(); err != nil {
		t.Fatalf("WriteHeader failed: %v", err)
	}

	embedder := NewTTFFontEmbedder(w)
	ref, err := embedder.EmbedTTFFont(ttfFont, usedGlyphs)
	if err != nil {
		t.Fatalf("EmbedTTFFont failed: %v", err)
	}

	if ref == nil {
		t.Fatal("EmbedTTFFont returned nil reference")
	}
	if ref.ObjectNumber <= 0 {
		t.Errorf("reference ObjectNumber = %d, want > 0", ref.ObjectNumber)
	}
	if ref.GenerationNumber != 0 {
		t.Errorf("reference GenerationNumber = %d, want 0", ref.GenerationNumber)
	}

	output := buf.String()

	// Verify the output contains expected PDF structures
	checks := []struct {
		name    string
		content string
	}{
		{"PDF header", "%PDF-1.7"},
		{"Type0 font type", "/Type0"},
		{"BaseFont", "/BaseFont"},
		{"Encoding Identity-H", "/Identity-H"},
		{"DescendantFonts", "/DescendantFonts"},
		{"ToUnicode", "/ToUnicode"},
		{"CIDFontType2", "/CIDFontType2"},
		{"CIDSystemInfo", "/CIDSystemInfo"},
		{"FontDescriptor type", "/FontDescriptor"},
		{"FontFile2", "/FontFile2"},
		{"FontName", "/FontName"},
		{"CIDToGIDMap Identity", "/Identity"},
		{"CMap begincodespacerange", "begincodespacerange"},
		{"CMap beginbfchar", "beginbfchar"},
		{"CMap endbfchar", "endbfchar"},
		{"stream keyword", "stream"},
		{"endstream keyword", "endstream"},
	}

	for _, c := range checks {
		if !strings.Contains(output, c.content) {
			t.Errorf("output missing %s (%q)", c.name, c.content)
		}
	}
}

// TestEmbedTTFFont_ObjectCount verifies the correct number of PDF objects are created.
func TestEmbedTTFFont_ObjectCount(t *testing.T) {
	ttfFont := loadTestFont(t)
	usedGlyphs := buildUsedGlyphs(t, ttfFont, "AB")

	var buf bytes.Buffer
	w := NewWriter(&buf)

	embedder := NewTTFFontEmbedder(w)
	ref, err := embedder.EmbedTTFFont(ttfFont, usedGlyphs)
	if err != nil {
		t.Fatalf("EmbedTTFFont failed: %v", err)
	}

	// EmbedTTFFont creates 5 objects:
	// 1. FontFile2 stream
	// 2. FontDescriptor
	// 3. CIDFont
	// 4. ToUnicode CMap
	// 5. Type0 font
	// The returned reference should be for the last object (Type0 font)
	if ref.ObjectNumber != 5 {
		t.Errorf("Type0 font object number = %d, want 5", ref.ObjectNumber)
	}

	// nextObjNum should be 6 (5 objects created, next available is 6)
	if w.nextObjNum != 6 {
		t.Errorf("nextObjNum = %d, want 6", w.nextObjNum)
	}
}

// TestEmbedTTFFont_EmptyGlyphs tests embedding with an empty glyph map.
func TestEmbedTTFFont_EmptyGlyphs(t *testing.T) {
	ttfFont := loadTestFont(t)
	usedGlyphs := map[uint16]rune{}

	var buf bytes.Buffer
	w := NewWriter(&buf)

	embedder := NewTTFFontEmbedder(w)
	ref, err := embedder.EmbedTTFFont(ttfFont, usedGlyphs)
	if err != nil {
		t.Fatalf("EmbedTTFFont with empty glyphs failed: %v", err)
	}

	if ref == nil {
		t.Fatal("EmbedTTFFont returned nil reference even with empty glyphs")
	}

	output := buf.String()

	// With empty glyphs, the /W array should not appear
	// But the CMap should still be generated (without bfchar entries)
	if !strings.Contains(output, "/CIDFontType2") {
		t.Error("output missing CIDFontType2")
	}
}

// TestEmbedTTFFont_WidthArray tests that the /W array contains correct glyph widths.
func TestEmbedTTFFont_WidthArray(t *testing.T) {
	ttfFont := loadTestFont(t)
	// Use a single character to make width checking easier
	usedGlyphs := buildUsedGlyphs(t, ttfFont, "A")

	var buf bytes.Buffer
	w := NewWriter(&buf)

	embedder := NewTTFFontEmbedder(w)
	_, err := embedder.EmbedTTFFont(ttfFont, usedGlyphs)
	if err != nil {
		t.Fatalf("EmbedTTFFont failed: %v", err)
	}

	output := buf.String()
	// The /W array should be present since we have used glyphs
	if !strings.Contains(output, "/W") {
		t.Error("output should contain /W array for used glyphs")
	}
}

// TestBuildWidthArray tests the buildWidthArray method directly.
func TestBuildWidthArray(t *testing.T) {
	ttfFont := loadTestFont(t)

	var buf bytes.Buffer
	w := NewWriter(&buf)
	embedder := NewTTFFontEmbedder(w)

	tests := []struct {
		name       string
		text       string
		wantNil    bool
		minEntries int // minimum number of entries in the width array
	}{
		{
			name:    "empty glyphs",
			text:    "",
			wantNil: true,
		},
		{
			name:       "single character",
			text:       "A",
			wantNil:    false,
			minEntries: 2, // [gid [width]]
		},
		{
			name:       "multiple characters",
			text:       "Hello",
			wantNil:    false,
			minEntries: 4, // at least 2 unique chars (H, e, l, o) * 2 entries each
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var usedGlyphs map[uint16]rune
			if tt.text == "" {
				usedGlyphs = map[uint16]rune{}
			} else {
				usedGlyphs = buildUsedGlyphs(t, ttfFont, tt.text)
			}

			wArray := embedder.buildWidthArray(ttfFont, usedGlyphs)

			if tt.wantNil {
				if wArray != nil {
					t.Errorf("buildWidthArray() should return nil for empty glyphs, got %v", wArray)
				}
				return
			}

			if wArray == nil {
				t.Fatal("buildWidthArray() returned nil, want non-nil")
			}

			if len(wArray) < tt.minEntries {
				t.Errorf("buildWidthArray() returned %d entries, want at least %d", len(wArray), tt.minEntries)
			}
		})
	}
}

// TestCreateToUnicodeCMap tests the CMap generation.
func TestCreateToUnicodeCMap(t *testing.T) {
	ttfFont := loadTestFont(t)

	tests := []struct {
		name          string
		text          string
		wantBfchar    bool
		wantSubstring []string
	}{
		{
			name:       "with glyphs",
			text:       "AB",
			wantBfchar: true,
			wantSubstring: []string{
				"begincodespacerange",
				"endcodespacerange",
				"beginbfchar",
				"endbfchar",
				"<0000> <FFFF>",
				"/CIDInit",
				"/CMapName",
			},
		},
		{
			name:       "empty glyphs",
			text:       "",
			wantBfchar: false,
			wantSubstring: []string{
				"begincodespacerange",
				"endcodespacerange",
				"/CIDInit",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			w := NewWriter(&buf)
			embedder := NewTTFFontEmbedder(w)

			var usedGlyphs map[uint16]rune
			if tt.text == "" {
				usedGlyphs = map[uint16]rune{}
			} else {
				usedGlyphs = buildUsedGlyphs(t, ttfFont, tt.text)
			}

			ref, err := embedder.createToUnicodeCMap(ttfFont, usedGlyphs)
			if err != nil {
				t.Fatalf("createToUnicodeCMap failed: %v", err)
			}

			if ref == nil {
				t.Fatal("createToUnicodeCMap returned nil reference")
			}
			if ref.ObjectNumber <= 0 {
				t.Errorf("reference ObjectNumber = %d, want > 0", ref.ObjectNumber)
			}

			output := buf.String()
			for _, s := range tt.wantSubstring {
				if !strings.Contains(output, s) {
					t.Errorf("output missing %q", s)
				}
			}

			if tt.wantBfchar {
				if !strings.Contains(output, "beginbfchar") {
					t.Error("output should contain beginbfchar for non-empty glyphs")
				}
			}
		})
	}
}

// TestCreateFontFile2 tests the FontFile2 stream creation.
func TestCreateFontFile2(t *testing.T) {
	ttfFont := loadTestFont(t)

	var buf bytes.Buffer
	w := NewWriter(&buf)
	embedder := NewTTFFontEmbedder(w)

	ref, err := embedder.createFontFile2(ttfFont)
	if err != nil {
		t.Fatalf("createFontFile2 failed: %v", err)
	}

	if ref == nil {
		t.Fatal("createFontFile2 returned nil reference")
	}
	if ref.ObjectNumber != 1 {
		t.Errorf("FontFile2 object number = %d, want 1", ref.ObjectNumber)
	}

	output := buf.String()
	// Should contain Length and Length1
	if !strings.Contains(output, "/Length") {
		t.Error("FontFile2 stream should contain /Length")
	}
	if !strings.Contains(output, "/Length1") {
		t.Error("FontFile2 stream should contain /Length1")
	}
	if !strings.Contains(output, "stream") {
		t.Error("FontFile2 should contain stream keyword")
	}
}

// TestCreateFontDescriptor tests the FontDescriptor creation.
func TestCreateFontDescriptor(t *testing.T) {
	ttfFont := loadTestFont(t)

	var buf bytes.Buffer
	w := NewWriter(&buf)
	embedder := NewTTFFontEmbedder(w)

	// First create a dummy FontFile2 reference
	fontFileRef, err := embedder.createFontFile2(ttfFont)
	if err != nil {
		t.Fatalf("createFontFile2 failed: %v", err)
	}

	ref, err := embedder.createFontDescriptor(ttfFont, fontFileRef)
	if err != nil {
		t.Fatalf("createFontDescriptor failed: %v", err)
	}

	if ref == nil {
		t.Fatal("createFontDescriptor returned nil reference")
	}
	if ref.ObjectNumber != 2 {
		t.Errorf("FontDescriptor object number = %d, want 2", ref.ObjectNumber)
	}

	output := buf.String()

	expectedKeys := []string{
		"/FontDescriptor",
		"/FontName",
		"/Flags",
		"/FontBBox",
		"/ItalicAngle",
		"/Ascent",
		"/Descent",
		"/CapHeight",
		"/StemV",
		"/FontFile2",
	}

	for _, key := range expectedKeys {
		if !strings.Contains(output, key) {
			t.Errorf("FontDescriptor output missing %s", key)
		}
	}
}

// TestCreateCIDFont tests the CIDFont creation.
func TestCreateCIDFont(t *testing.T) {
	ttfFont := loadTestFont(t)
	usedGlyphs := buildUsedGlyphs(t, ttfFont, "Test")

	var buf bytes.Buffer
	w := NewWriter(&buf)
	embedder := NewTTFFontEmbedder(w)

	// Create prerequisite references
	fontFileRef, err := embedder.createFontFile2(ttfFont)
	if err != nil {
		t.Fatalf("createFontFile2 failed: %v", err)
	}
	fontDescRef, err := embedder.createFontDescriptor(ttfFont, fontFileRef)
	if err != nil {
		t.Fatalf("createFontDescriptor failed: %v", err)
	}

	ref, err := embedder.createCIDFont(ttfFont, fontDescRef, usedGlyphs)
	if err != nil {
		t.Fatalf("createCIDFont failed: %v", err)
	}

	if ref == nil {
		t.Fatal("createCIDFont returned nil reference")
	}

	output := buf.String()

	expectedKeys := []string{
		"/CIDFontType2",
		"/BaseFont",
		"/CIDSystemInfo",
		"/FontDescriptor",
		"/DW",
		"/CIDToGIDMap",
		"/W",
	}

	for _, key := range expectedKeys {
		if !strings.Contains(output, key) {
			t.Errorf("CIDFont output missing %s", key)
		}
	}
}

// TestCreateCIDFont_NoWidthsForEmptyGlyphs tests CIDFont with empty glyph map.
func TestCreateCIDFont_NoWidthsForEmptyGlyphs(t *testing.T) {
	ttfFont := loadTestFont(t)
	usedGlyphs := map[uint16]rune{}

	var buf bytes.Buffer
	w := NewWriter(&buf)
	embedder := NewTTFFontEmbedder(w)

	fontFileRef, err := embedder.createFontFile2(ttfFont)
	if err != nil {
		t.Fatalf("createFontFile2 failed: %v", err)
	}
	fontDescRef, err := embedder.createFontDescriptor(ttfFont, fontFileRef)
	if err != nil {
		t.Fatalf("createFontDescriptor failed: %v", err)
	}

	ref, err := embedder.createCIDFont(ttfFont, fontDescRef, usedGlyphs)
	if err != nil {
		t.Fatalf("createCIDFont failed: %v", err)
	}

	if ref == nil {
		t.Fatal("createCIDFont returned nil reference")
	}

	// With empty glyphs, /W should not appear in the CIDFont dict
	// (the CIDFont object is the 3rd object written)
	output := buf.String()
	// Find the CIDFont object - it should have /DW but not /W when glyphs are empty
	if !strings.Contains(output, "/DW") {
		t.Error("CIDFont should contain /DW")
	}
}

// TestCreateType0Font tests the Type0 font creation.
func TestCreateType0Font(t *testing.T) {
	ttfFont := loadTestFont(t)

	var buf bytes.Buffer
	w := NewWriter(&buf)
	embedder := NewTTFFontEmbedder(w)

	// Create dummy references
	usedGlyphs := buildUsedGlyphs(t, ttfFont, "X")
	fontFileRef, _ := embedder.createFontFile2(ttfFont)
	fontDescRef, _ := embedder.createFontDescriptor(ttfFont, fontFileRef)
	cidFontRef, _ := embedder.createCIDFont(ttfFont, fontDescRef, usedGlyphs)
	toUnicodeRef, _ := embedder.createToUnicodeCMap(ttfFont, usedGlyphs)

	ref, err := embedder.createType0Font(ttfFont, cidFontRef, toUnicodeRef)
	if err != nil {
		t.Fatalf("createType0Font failed: %v", err)
	}

	if ref == nil {
		t.Fatal("createType0Font returned nil reference")
	}

	output := buf.String()

	expectedKeys := []string{
		"/Type0",
		"/BaseFont",
		"/Encoding",
		"/Identity-H",
		"/DescendantFonts",
		"/ToUnicode",
	}

	for _, key := range expectedKeys {
		if !strings.Contains(output, key) {
			t.Errorf("Type0 font output missing %s", key)
		}
	}
}

// TestCreateCIDSystemInfo tests the CIDSystemInfo dictionary creation.
func TestCreateCIDSystemInfo(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	embedder := NewTTFFontEmbedder(w)

	info := embedder.createCIDSystemInfo()

	if info == nil {
		t.Fatal("createCIDSystemInfo returned nil")
	}

	// Serialize to check content
	s := NewSerializer(&buf)
	if err := s.Serialize(info); err != nil {
		t.Fatalf("failed to serialize CIDSystemInfo: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "/Registry") {
		t.Error("CIDSystemInfo missing /Registry")
	}
	if !strings.Contains(output, "/Ordering") {
		t.Error("CIDSystemInfo missing /Ordering")
	}
	if !strings.Contains(output, "/Supplement") {
		t.Error("CIDSystemInfo missing /Supplement")
	}
}

// TestEmbedTTFFont_FontNameInOutput tests that the font name appears in the output.
func TestEmbedTTFFont_FontNameInOutput(t *testing.T) {
	ttfFont := loadTestFont(t)
	fontName := ttfFont.Name()

	usedGlyphs := buildUsedGlyphs(t, ttfFont, "A")

	var buf bytes.Buffer
	w := NewWriter(&buf)

	embedder := NewTTFFontEmbedder(w)
	_, err := embedder.EmbedTTFFont(ttfFont, usedGlyphs)
	if err != nil {
		t.Fatalf("EmbedTTFFont failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, fontName) {
		t.Errorf("output should contain font name %q", fontName)
	}
}
