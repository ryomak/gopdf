package gopdf

// StandardFont represents one of the 14 standard PDF fonts.
// These fonts are built into PDF viewers and don't need to be embedded.
type StandardFont string

// Sans-serif fonts
const (
	// FontHelvetica is a sans-serif font
	FontHelvetica StandardFont = "Helvetica"
	// FontHelveticaBold is a bold variant of Helvetica
	FontHelveticaBold StandardFont = "Helvetica-Bold"
	// FontHelveticaOblique is an oblique (italic) variant of Helvetica
	FontHelveticaOblique StandardFont = "Helvetica-Oblique"
	// FontHelveticaBoldOblique is a bold oblique variant of Helvetica
	FontHelveticaBoldOblique StandardFont = "Helvetica-BoldOblique"
)

// Serif fonts
const (
	// FontTimesRoman is a serif font (also known as Times New Roman)
	FontTimesRoman StandardFont = "Times-Roman"
	// FontTimesBold is a bold variant of Times Roman
	FontTimesBold StandardFont = "Times-Bold"
	// FontTimesItalic is an italic variant of Times Roman
	FontTimesItalic StandardFont = "Times-Italic"
	// FontTimesBoldItalic is a bold italic variant of Times Roman
	FontTimesBoldItalic StandardFont = "Times-BoldItalic"
)

// Monospace fonts
const (
	// FontCourier is a monospace font
	FontCourier StandardFont = "Courier"
	// FontCourierBold is a bold variant of Courier
	FontCourierBold StandardFont = "Courier-Bold"
	// FontCourierOblique is an oblique (italic) variant of Courier
	FontCourierOblique StandardFont = "Courier-Oblique"
	// FontCourierBoldOblique is a bold oblique variant of Courier
	FontCourierBoldOblique StandardFont = "Courier-BoldOblique"
)

// Symbol fonts
const (
	// FontSymbol is a symbol font containing mathematical and other symbols
	FontSymbol StandardFont = "Symbol"
	// FontZapfDingbats is a dingbat (ornamental) font
	FontZapfDingbats StandardFont = "ZapfDingbats"
)

// Name returns the PostScript name of the font.
func (f StandardFont) Name() string {
	return string(f)
}
