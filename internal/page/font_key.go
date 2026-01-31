package page

import (
	"fmt"

	"github.com/ryomak/gopdf/internal/font"
)

// GetFontKey returns the font resource name (e.g., "F1", "F2") for a given standard font.
func GetFontKey(f font.StandardFont) string {
	switch f {
	case font.Helvetica:
		return "F1"
	case font.HelveticaBold:
		return "F2"
	case font.HelveticaOblique:
		return "F3"
	case font.HelveticaBoldOblique:
		return "F4"
	case font.TimesRoman:
		return "F5"
	case font.TimesBold:
		return "F6"
	case font.TimesItalic:
		return "F7"
	case font.TimesBoldItalic:
		return "F8"
	case font.Courier:
		return "F9"
	case font.CourierBold:
		return "F10"
	case font.CourierOblique:
		return "F11"
	case font.CourierBoldOblique:
		return "F12"
	case font.Symbol:
		return "F13"
	case font.ZapfDingbats:
		return "F14"
	default:
		return "F1"
	}
}

// TTFFontKeyGetter is an interface for getting TTF font keys.
type TTFFontKeyGetter interface {
	// GetRegisteredKey returns the key for an already registered font, or empty string if not found.
	GetRegisteredKey(fontPtr interface{}) string
	// GetFontCount returns the number of TTF fonts registered.
	GetFontCount() int
}

// GetTTFFontKey returns the font resource name for a TTF font.
// It checks if the font is already registered and returns its key,
// or generates a new unique key based on current font count.
func GetTTFFontKey(fontPtr interface{}, getter TTFFontKeyGetter) string {
	// Check if this font is already registered and return its key
	if key := getter.GetRegisteredKey(fontPtr); key != "" {
		return key
	}

	// Generate a unique key based on current font count
	// Use F15+ to avoid conflicts with standard fonts (F1-F14)
	count := getter.GetFontCount()
	if count == 0 {
		return "F15"
	}

	return fmt.Sprintf("F%d", 15+count)
}
