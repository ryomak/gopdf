// Package font provides font management functionality for PDF documents.
package font

import (
	"fmt"
)

// StandardFont represents one of the 14 standard PDF fonts.
// These fonts are built into PDF viewers and don't need to be embedded.
type StandardFont string

// Standard Type1 fonts defined in PDF specification
const (
	// Sans-serif fonts
	Helvetica           StandardFont = "Helvetica"
	HelveticaBold       StandardFont = "Helvetica-Bold"
	HelveticaOblique    StandardFont = "Helvetica-Oblique"
	HelveticaBoldOblique StandardFont = "Helvetica-BoldOblique"

	// Serif fonts
	TimesRoman      StandardFont = "Times-Roman"
	TimesBold       StandardFont = "Times-Bold"
	TimesItalic     StandardFont = "Times-Italic"
	TimesBoldItalic StandardFont = "Times-BoldItalic"

	// Monospace fonts
	Courier           StandardFont = "Courier"
	CourierBold       StandardFont = "Courier-Bold"
	CourierOblique    StandardFont = "Courier-Oblique"
	CourierBoldOblique StandardFont = "Courier-BoldOblique"

	// Symbol fonts
	Symbol       StandardFont = "Symbol"
	ZapfDingbats StandardFont = "ZapfDingbats"
)

// Name returns the PostScript name of the font.
func (f StandardFont) Name() string {
	return string(f)
}

// Type returns the font type (always "Type1" for standard fonts).
func (f StandardFont) Type() string {
	return "Type1"
}

// Encoding returns the encoding used by the font.
func (f StandardFont) Encoding() string {
	// Symbol and ZapfDingbats use their own encoding
	if f == Symbol || f == ZapfDingbats {
		return "BuiltIn"
	}
	// Other fonts use WinAnsiEncoding (similar to ISO-8859-1)
	return "WinAnsiEncoding"
}

// IsStandard returns true since this is a standard font.
func (f StandardFont) IsStandard() bool {
	return true
}

// standardFonts is a map of font names to StandardFont values.
var standardFonts = map[string]StandardFont{
	"Helvetica":            Helvetica,
	"Helvetica-Bold":       HelveticaBold,
	"Helvetica-Oblique":    HelveticaOblique,
	"Helvetica-BoldOblique": HelveticaBoldOblique,
	"Times-Roman":          TimesRoman,
	"Times-Bold":           TimesBold,
	"Times-Italic":         TimesItalic,
	"Times-BoldItalic":     TimesBoldItalic,
	"Courier":              Courier,
	"Courier-Bold":         CourierBold,
	"Courier-Oblique":      CourierOblique,
	"Courier-BoldOblique":  CourierBoldOblique,
	"Symbol":               Symbol,
	"ZapfDingbats":         ZapfDingbats,
}

// GetStandardFont returns a StandardFont by name.
// Returns an error if the font name is not a standard font.
func GetStandardFont(name string) (StandardFont, error) {
	font, ok := standardFonts[name]
	if !ok {
		return "", fmt.Errorf("not a standard font: %s", name)
	}
	return font, nil
}

// IsStandardFontName checks if a given name is a standard font.
func IsStandardFontName(name string) bool {
	_, ok := standardFonts[name]
	return ok
}
