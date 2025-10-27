package gopdf

import (
	"bytes"
	"fmt"

	"github.com/ryomak/gopdf/internal/font"
)

// Page represents a single page in a PDF document.
type Page struct {
	width       float64
	height      float64
	content     bytes.Buffer
	currentFont *font.StandardFont
	fontSize    float64
	fonts       map[string]font.StandardFont // fontKey -> font
}

// Width returns the page width in points.
func (p *Page) Width() float64 {
	return p.width
}

// Height returns the page height in points.
func (p *Page) Height() float64 {
	return p.height
}

// SetFont sets the current font and size for subsequent text operations.
func (p *Page) SetFont(f font.StandardFont, size float64) error {
	p.currentFont = &f
	p.fontSize = size

	// フォントをページのフォントリストに追加
	if p.fonts == nil {
		p.fonts = make(map[string]font.StandardFont)
	}
	fontKey := p.getFontKey(f)
	p.fonts[fontKey] = f

	return nil
}

// DrawText draws text at the specified position.
// The position (x, y) is in PDF units (points), where (0, 0) is the bottom-left corner.
func (p *Page) DrawText(text string, x, y float64) error {
	if p.currentFont == nil {
		return fmt.Errorf("no font set; call SetFont before DrawText")
	}

	// PDFコンテンツストリームの生成
	// BT = Begin Text
	// /F1 12 Tf = Set font F1 at size 12
	// 100 700 Td = Move to position (100, 700)
	// (Hello, World!) Tj = Show text
	// ET = End Text

	fontKey := p.getFontKey(*p.currentFont)

	fmt.Fprintf(&p.content, "BT\n")
	fmt.Fprintf(&p.content, "/%s %.2f Tf\n", fontKey, p.fontSize)
	fmt.Fprintf(&p.content, "%.2f %.2f Td\n", x, y)
	fmt.Fprintf(&p.content, "(%s) Tj\n", p.escapeString(text))
	fmt.Fprintf(&p.content, "ET\n")

	return nil
}

// getFontKey returns the font resource name (e.g., "F1", "F2") for a given font.
func (p *Page) getFontKey(f font.StandardFont) string {
	// 簡易的な実装: フォント名のハッシュ値を使用
	// 実際には、ドキュメント全体でユニークなキーを管理する必要がある
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

// escapeString escapes special characters in PDF strings.
func (p *Page) escapeString(s string) string {
	// TODO: 完全なエスケープ処理の実装
	// 現在は基本的な文字のみ対応
	result := s
	result = replaceAll(result, "\\", "\\\\")
	result = replaceAll(result, "(", "\\(")
	result = replaceAll(result, ")", "\\)")
	return result
}

// replaceAll is a helper function to replace all occurrences of old with new.
func replaceAll(s, old, new string) string {
	result := ""
	for i := 0; i < len(s); i++ {
		if i <= len(s)-len(old) && s[i:i+len(old)] == old {
			result += new
			i += len(old) - 1
		} else {
			result += string(s[i])
		}
	}
	return result
}

// SetLineWidth sets the line width for subsequent drawing operations.
func (p *Page) SetLineWidth(width float64) {
	fmt.Fprintf(&p.content, "%.2f w\n", width)
}

// SetStrokeColor sets the stroke color for subsequent drawing operations.
func (p *Page) SetStrokeColor(c Color) {
	fmt.Fprintf(&p.content, "%.2f %.2f %.2f RG\n", c.R, c.G, c.B)
}

// SetFillColor sets the fill color for subsequent drawing operations.
func (p *Page) SetFillColor(c Color) {
	fmt.Fprintf(&p.content, "%.2f %.2f %.2f rg\n", c.R, c.G, c.B)
}

// SetLineCap sets the line cap style for subsequent drawing operations.
func (p *Page) SetLineCap(cap LineCapStyle) {
	fmt.Fprintf(&p.content, "%d J\n", cap)
}

// SetLineJoin sets the line join style for subsequent drawing operations.
func (p *Page) SetLineJoin(join LineJoinStyle) {
	fmt.Fprintf(&p.content, "%d j\n", join)
}

// DrawLine draws a line from (x1, y1) to (x2, y2).
func (p *Page) DrawLine(x1, y1, x2, y2 float64) {
	fmt.Fprintf(&p.content, "%.2f %.2f m\n", x1, y1)
	fmt.Fprintf(&p.content, "%.2f %.2f l\n", x2, y2)
	fmt.Fprintf(&p.content, "S\n")
}

// DrawRectangle draws a rectangle outline at (x, y) with the specified width and height.
func (p *Page) DrawRectangle(x, y, width, height float64) {
	fmt.Fprintf(&p.content, "%.2f %.2f %.2f %.2f re\n", x, y, width, height)
	fmt.Fprintf(&p.content, "S\n")
}

// FillRectangle draws a filled rectangle at (x, y) with the specified width and height.
func (p *Page) FillRectangle(x, y, width, height float64) {
	fmt.Fprintf(&p.content, "%.2f %.2f %.2f %.2f re\n", x, y, width, height)
	fmt.Fprintf(&p.content, "f\n")
}

// DrawAndFillRectangle draws a filled rectangle with an outline at (x, y) with the specified width and height.
func (p *Page) DrawAndFillRectangle(x, y, width, height float64) {
	fmt.Fprintf(&p.content, "%.2f %.2f %.2f %.2f re\n", x, y, width, height)
	fmt.Fprintf(&p.content, "B\n")
}

// drawCirclePath draws a circle path using 4 Bézier curves.
// κ = 4 * (√2 - 1) / 3 ≈ 0.5522847498
func (p *Page) drawCirclePath(centerX, centerY, radius float64) {
	// Magic constant for circle approximation using Bézier curves
	const kappa = 0.5522847498

	// Calculate control point offset
	offset := radius * kappa

	// Calculate key points on the circle
	x0 := centerX + radius // Right
	y0 := centerY
	x1 := centerX          // Left
	y1 := centerY
	x2 := centerX          // Center X
	y2 := centerY + radius // Top
	x3 := centerX          // Center X
	y3 := centerY - radius // Bottom

	// Start at the right point (3 o'clock position)
	fmt.Fprintf(&p.content, "%.2f %.2f m\n", x0, y0)

	// Draw 4 Bézier curves to approximate a circle
	// Curve 1: Right to Top (3 o'clock to 12 o'clock)
	fmt.Fprintf(&p.content, "%.2f %.2f %.2f %.2f %.2f %.2f c\n",
		x0, y0+offset,        // Control point 1
		x2+offset, y2,        // Control point 2
		x2, y2)               // End point

	// Curve 2: Top to Left (12 o'clock to 9 o'clock)
	fmt.Fprintf(&p.content, "%.2f %.2f %.2f %.2f %.2f %.2f c\n",
		x2-offset, y2,        // Control point 1
		x1, y1+offset,        // Control point 2
		x1, y1)               // End point

	// Curve 3: Left to Bottom (9 o'clock to 6 o'clock)
	fmt.Fprintf(&p.content, "%.2f %.2f %.2f %.2f %.2f %.2f c\n",
		x1, y1-offset,        // Control point 1
		x3-offset, y3,        // Control point 2
		x3, y3)               // End point

	// Curve 4: Bottom to Right (6 o'clock to 3 o'clock)
	fmt.Fprintf(&p.content, "%.2f %.2f %.2f %.2f %.2f %.2f c\n",
		x3+offset, y3,        // Control point 1
		x0, y0-offset,        // Control point 2
		x0, y0)               // End point
}

// DrawCircle draws a circle outline with the specified center and radius.
func (p *Page) DrawCircle(centerX, centerY, radius float64) {
	p.drawCirclePath(centerX, centerY, radius)
	fmt.Fprintf(&p.content, "S\n")
}

// FillCircle draws a filled circle with the specified center and radius.
func (p *Page) FillCircle(centerX, centerY, radius float64) {
	p.drawCirclePath(centerX, centerY, radius)
	fmt.Fprintf(&p.content, "f\n")
}

// DrawAndFillCircle draws a filled circle with an outline with the specified center and radius.
func (p *Page) DrawAndFillCircle(centerX, centerY, radius float64) {
	p.drawCirclePath(centerX, centerY, radius)
	fmt.Fprintf(&p.content, "B\n")
}
