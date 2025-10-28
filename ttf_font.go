package gopdf

import (
	"github.com/ryomak/gopdf/internal/font"
)

// TTFFont represents a TrueType Font for use in PDF documents
type TTFFont struct {
	internal *font.TTFFont
}

// LoadTTF loads a TrueType font from a file path
func LoadTTF(path string) (*TTFFont, error) {
	internalFont, err := font.LoadTTF(path)
	if err != nil {
		return nil, err
	}

	return &TTFFont{
		internal: internalFont,
	}, nil
}

// LoadTTFFromBytes loads a TrueType font from a byte slice
func LoadTTFFromBytes(data []byte) (*TTFFont, error) {
	internalFont, err := font.LoadTTFFromBytes(data)
	if err != nil {
		return nil, err
	}

	return &TTFFont{
		internal: internalFont,
	}, nil
}

// Name returns the font name
func (f *TTFFont) Name() string {
	return f.internal.Name()
}

// TextWidth calculates the width of a text string at a given font size
func (f *TTFFont) TextWidth(text string, fontSize float64) (float64, error) {
	return f.internal.TextWidth(text, fontSize)
}
