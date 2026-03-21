package page

import (
	"fmt"
	"strings"
)

// EscapeString escapes special characters in PDF strings.
// It escapes backslashes and parentheses.
func EscapeString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "(", "\\(")
	s = strings.ReplaceAll(s, ")", "\\)")
	return s
}

// TextToHexString converts UTF-8 text to a UTF-16BE hex string for PDF.
// Characters outside the BMP (U+10000 and above) are encoded as surrogate pairs.
func TextToHexString(text string) string {
	result := ""

	for _, r := range text {
		if r <= 0xFFFF {
			result += fmt.Sprintf("%04X", r)
		} else {
			// UTF-16 surrogate pair for non-BMP characters
			r -= 0x10000
			high := 0xD800 + (r>>10)&0x3FF
			low := 0xDC00 + r&0x3FF
			result += fmt.Sprintf("%04X%04X", high, low)
		}
	}

	return result
}

// GlyphIndexer is an interface for getting glyph indices from a TTF font.
type GlyphIndexer interface {
	GetGlyphIndex(r rune) (int, error)
}

// GlyphRecorder records glyph usage for ToUnicode CMap generation.
type GlyphRecorder interface {
	RecordGlyph(glyphIndex uint16, r rune)
}

// TextToGlyphIndices converts UTF-8 text to glyph indices for TTF fonts.
// This ensures proper rendering by using actual glyph IDs from the font.
func TextToGlyphIndices(text string, indexer GlyphIndexer, recorder GlyphRecorder) (string, error) {
	var result string

	for _, r := range text {
		// Get the glyph index for this character
		glyphIndex, err := indexer.GetGlyphIndex(r)
		if err != nil {
			return "", fmt.Errorf("failed to get glyph index for character %c (U+%04X): %w", r, r, err)
		}

		// Record glyph usage for ToUnicode CMap generation
		if recorder != nil {
			recorder.RecordGlyph(uint16(glyphIndex), r)
		}

		// Convert glyph index to 4-digit hex string
		result += fmt.Sprintf("%04X", glyphIndex)
	}

	return result, nil
}

