package font

import (
	"fmt"
	"os"

	"golang.org/x/image/font"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

// TTFFont represents a TrueType Font
type TTFFont struct {
	name     string
	data     []byte        // Original font file data
	font     *sfnt.Font    // Parsed font
	glyphMap map[rune]sfnt.GlyphIndex // rune → GlyphIndex mapping
}

// LoadTTF loads a TrueType font from a file path
func LoadTTF(path string) (*TTFFont, error) {
	// Read font file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read TTF file: %w", err)
	}

	return LoadTTFFromBytes(data)
}

// LoadTTFFromBytes loads a TrueType font from byte slice
// It handles both TTF (single font) and TTC (font collection) files
func LoadTTFFromBytes(data []byte) (*TTFFont, error) {
	// Try to parse as a single font first
	font, err := sfnt.Parse(data)
	if err != nil {
		// If it fails, try parsing as a collection and use the first font
		collection, collErr := sfnt.ParseCollection(data)
		if collErr != nil {
			return nil, fmt.Errorf("failed to parse TTF/TTC: %w (collection error: %v)", err, collErr)
		}

		if collection.NumFonts() == 0 {
			return nil, fmt.Errorf("font collection is empty")
		}

		// Use the first font in the collection
		font, err = collection.Font(0)
		if err != nil {
			return nil, fmt.Errorf("failed to get font from collection: %w", err)
		}
	}

	// Extract font name
	name, err := extractFontName(font)
	if err != nil {
		return nil, fmt.Errorf("failed to extract font name: %w", err)
	}

	// Build glyph map (rune → GlyphIndex)
	glyphMap := make(map[rune]sfnt.GlyphIndex)

	// Note: We'll build the glyphMap on-demand in GlyphWidth
	// to avoid scanning all possible Unicode characters upfront

	return &TTFFont{
		name:     name,
		data:     data,
		font:     font,
		glyphMap: glyphMap,
	}, nil
}

// Name returns the font name
func (f *TTFFont) Name() string {
	return f.name
}

// Data returns the original font file data
func (f *TTFFont) Data() []byte {
	return f.data
}

// Font returns the underlying sfnt.Font
func (f *TTFFont) Font() *sfnt.Font {
	return f.font
}

// GlyphWidth returns the width of a character in PDF user units
func (f *TTFFont) GlyphWidth(r rune, fontSize float64) (float64, error) {
	// Get or cache glyph index
	glyphIndex, ok := f.glyphMap[r]
	if !ok {
		// Look up glyph index for this rune
		buf := &sfnt.Buffer{}
		idx, err := f.font.GlyphIndex(buf, r)
		if err != nil {
			return 0, fmt.Errorf("failed to get glyph index for rune %c: %w", r, err)
		}
		glyphIndex = idx
		f.glyphMap[r] = idx
	}

	// Get glyph advance width at ppem=1000
	buf := &sfnt.Buffer{}
	advance, err := f.font.GlyphAdvance(buf, glyphIndex, fixed.I(1000), font.HintingNone)
	if err != nil {
		return 0, fmt.Errorf("failed to get glyph advance: %w", err)
	}

	// Convert advance from fixed.Int26_6 to float64
	// advance is in 26.6 fixed point format at ppem=1000
	widthInFontUnits := float64(advance) / 64.0 // Convert from fixed.Int26_6

	// Scale to actual font size
	// Since we got the advance at ppem=1000, scale it to fontSize
	widthInPoints := (widthInFontUnits * fontSize) / 1000.0

	return widthInPoints, nil
}

// GlyphWidthBatch returns widths for multiple characters
func (f *TTFFont) GlyphWidthBatch(runes []rune, fontSize float64) ([]float64, error) {
	widths := make([]float64, len(runes))
	for i, r := range runes {
		width, err := f.GlyphWidth(r, fontSize)
		if err != nil {
			return nil, err
		}
		widths[i] = width
	}
	return widths, nil
}

// extractFontName extracts the font name from the name table
func extractFontName(f *sfnt.Font) (string, error) {
	// Try to get PostScript name (Name ID 6)
	buf := &sfnt.Buffer{}
	name, err := f.Name(buf, sfnt.NameIDPostScript)
	if err == nil && name != "" {
		return name, nil
	}

	// Fallback to full font name (Name ID 4)
	name, err = f.Name(buf, sfnt.NameIDFull)
	if err == nil && name != "" {
		return name, nil
	}

	// Fallback to font family (Name ID 1)
	name, err = f.Name(buf, sfnt.NameIDFamily)
	if err == nil && name != "" {
		return name, nil
	}

	return "", fmt.Errorf("failed to extract font name")
}

// TextWidth calculates the total width of a text string
func (f *TTFFont) TextWidth(text string, fontSize float64) (float64, error) {
	totalWidth := 0.0
	for _, r := range text {
		width, err := f.GlyphWidth(r, fontSize)
		if err != nil {
			return 0, err
		}
		totalWidth += width
	}
	return totalWidth, nil
}
