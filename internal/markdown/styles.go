package markdown

// Color represents an RGB color (values from 0.0 to 1.0)
type Color struct {
	R, G, B float64
}

// Predefined colors
var (
	ColorBlack = Color{R: 0, G: 0, B: 0}
	ColorWhite = Color{R: 1, G: 1, B: 1}
	ColorRed   = Color{R: 1, G: 0, B: 0}
	ColorGreen = Color{R: 0, G: 1, B: 0}
	ColorBlue  = Color{R: 0, G: 0, B: 1}
)

// Style represents the styling configuration for Markdown rendering.
type Style struct {
	// Heading font sizes (H1-H6)
	H1Size, H2Size, H3Size, H4Size, H5Size, H6Size float64

	// Body text font size
	BodySize float64

	// Code block font size
	CodeSize float64

	// Line spacing (multiplier, e.g., 1.5 for 1.5x spacing)
	LineSpacing float64

	// Paragraph spacing (points)
	ParagraphSpacing float64

	// Margins
	MarginTop, MarginRight, MarginBottom, MarginLeft float64

	// Colors
	TextColor      Color
	HeadingColor   Color
	CodeBackground Color
	LinkColor      Color

	// Font path for TTF fonts (optional)
	FontPath string
}

// DefaultDocumentStyle returns the default style for document rendering.
func DefaultDocumentStyle() *Style {
	return &Style{
		H1Size:           36,
		H2Size:           28,
		H3Size:           22,
		H4Size:           18,
		H5Size:           14,
		H6Size:           12,
		BodySize:         12,
		CodeSize:         10,
		LineSpacing:      1.2,
		ParagraphSpacing: 12,
		MarginTop:        72,
		MarginRight:      72,
		MarginBottom:     72,
		MarginLeft:       72,
		TextColor:        ColorBlack,
		HeadingColor:     ColorBlack,
		CodeBackground:   Color{R: 0.95, G: 0.95, B: 0.95},
		LinkColor:        ColorBlue,
	}
}

// DefaultSlideStyle returns the default style for slide rendering.
func DefaultSlideStyle() *Style {
	return &Style{
		H1Size:           48,
		H2Size:           36,
		H3Size:           28,
		H4Size:           24,
		H5Size:           20,
		H6Size:           18,
		BodySize:         18,
		CodeSize:         14,
		LineSpacing:      1.3,
		ParagraphSpacing: 18,
		MarginTop:        50,
		MarginRight:      50,
		MarginBottom:     50,
		MarginLeft:       50,
		TextColor:        ColorBlack,
		HeadingColor:     Color{R: 0.2, G: 0.2, B: 0.6},
		CodeBackground:   Color{R: 0.95, G: 0.95, B: 0.95},
		LinkColor:        ColorBlue,
	}
}
