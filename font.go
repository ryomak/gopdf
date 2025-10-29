package gopdf

// StandardFont represents one of the 14 standard PDF fonts.
// These fonts are built into PDF viewers and don't need to be embedded.
type StandardFont string

// Sans-serif fonts
const (
	// Helvetica is a sans-serif font
	Helvetica StandardFont = "Helvetica"
	// HelveticaBold is a bold variant of Helvetica
	HelveticaBold StandardFont = "Helvetica-Bold"
	// HelveticaOblique is an oblique (italic) variant of Helvetica
	HelveticaOblique StandardFont = "Helvetica-Oblique"
	// HelveticaBoldOblique is a bold oblique variant of Helvetica
	HelveticaBoldOblique StandardFont = "Helvetica-BoldOblique"
)

// Serif fonts
const (
	// TimesRoman is a serif font (also known as Times New Roman)
	TimesRoman StandardFont = "Times-Roman"
	// TimesBold is a bold variant of Times Roman
	TimesBold StandardFont = "Times-Bold"
	// TimesItalic is an italic variant of Times Roman
	TimesItalic StandardFont = "Times-Italic"
	// TimesBoldItalic is a bold italic variant of Times Roman
	TimesBoldItalic StandardFont = "Times-BoldItalic"
)

// Monospace fonts
const (
	// Courier is a monospace font
	Courier StandardFont = "Courier"
	// CourierBold is a bold variant of Courier
	CourierBold StandardFont = "Courier-Bold"
	// CourierOblique is an oblique (italic) variant of Courier
	CourierOblique StandardFont = "Courier-Oblique"
	// CourierBoldOblique is a bold oblique variant of Courier
	CourierBoldOblique StandardFont = "Courier-BoldOblique"
)

// Symbol fonts
const (
	// Symbol is a symbol font containing mathematical and other symbols
	Symbol StandardFont = "Symbol"
	// ZapfDingbats is a dingbat (ornamental) font
	ZapfDingbats StandardFont = "ZapfDingbats"
)

// Name returns the PostScript name of the font.
func (f StandardFont) Name() string {
	return string(f)
}
