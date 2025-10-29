package gopdf

// Color represents an RGB color in PDF (values from 0.0 to 1.0)
type Color struct {
	R, G, B float64
}

// NewRGB creates a Color from 8-bit RGB values (0-255)
func NewRGB(r, g, b uint8) Color {
	return Color{
		R: float64(r) / 255.0,
		G: float64(g) / 255.0,
		B: float64(b) / 255.0,
	}
}

// Predefined colors
var (
	ColorBlack = Color{R: 0, G: 0, B: 0}
	ColorWhite = Color{R: 1, G: 1, B: 1}
	ColorRed   = Color{R: 1, G: 0, B: 0}
	ColorGreen = Color{R: 0, G: 1, B: 0}
	ColorBlue  = Color{R: 0, G: 0, B: 1}
)

// LineCapStyle represents the line cap style
type LineCapStyle int

const (
	ButtCap   LineCapStyle = 0 // Butt cap (default)
	RoundCap  LineCapStyle = 1 // Round cap
	SquareCap LineCapStyle = 2 // Square cap
)

// LineJoinStyle represents the line join style
type LineJoinStyle int

const (
	MiterJoin LineJoinStyle = 0 // Miter join (default)
	RoundJoin LineJoinStyle = 1 // Round join
	BevelJoin LineJoinStyle = 2 // Bevel join
)
