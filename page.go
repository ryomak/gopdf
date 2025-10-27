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
