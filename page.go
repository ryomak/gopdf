package gopdf

import (
	"bytes"
	"fmt"

	"github.com/ryomak/gopdf/internal/font"
)

// Page represents a single page in a PDF document.
type Page struct {
	width          float64
	height         float64
	content        bytes.Buffer
	currentFont    *font.StandardFont
	currentTTFFont *TTFFont
	fontSize       float64
	fonts          map[string]font.StandardFont // fontKey -> font
	ttfFonts       map[string]*TTFFont          // fontKey -> TTF font
	images         []*Image                     // images used in this page
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
func (p *Page) SetFont(f StandardFont, size float64) error {
	// 公開APIの型を内部実装の型に変換
	internalFont := font.StandardFont(f)

	p.currentFont = &internalFont
	p.fontSize = size

	// フォントをページのフォントリストに追加
	if p.fonts == nil {
		p.fonts = make(map[string]font.StandardFont)
	}
	fontKey := p.getFontKey(internalFont)
	p.fonts[fontKey] = internalFont

	return nil
}

// drawTextInternal は DrawText と DrawTextUTF8 の共通ロジック
// このメソッドは内部実装用であり、外部から直接呼び出すべきではない
func (p *Page) drawTextInternal(
	x, y float64,
	fontKey string,
	encodedText string,
	useBrackets bool,
) {
	fmt.Fprintf(&p.content, "BT\n")
	// Set text color to black (RGB: 0, 0, 0)
	fmt.Fprintf(&p.content, "0 0 0 rg\n")
	fmt.Fprintf(&p.content, "/%s %.2f Tf\n", fontKey, p.fontSize)
	fmt.Fprintf(&p.content, "%.2f %.2f Td\n", x, y)

	if useBrackets {
		fmt.Fprintf(&p.content, "(%s) Tj\n", encodedText)
	} else {
		fmt.Fprintf(&p.content, "<%s> Tj\n", encodedText)
	}

	fmt.Fprintf(&p.content, "ET\n")
}

// DrawText draws text at the specified position.
// The position (x, y) is in PDF units (points), where (0, 0) is the bottom-left corner.
func (p *Page) DrawText(text string, x, y float64) error {
	// Support both standard fonts and TTF fonts
	if p.currentTTFFont != nil {
		// Use TTF font (supports Unicode)
		fontKey := p.getTTFFontKey(p.currentTTFFont)
		encodedText, err := p.textToGlyphIndices(text, p.currentTTFFont)
		if err != nil {
			return fmt.Errorf("failed to convert text to glyph indices: %w", err)
		}
		p.drawTextInternal(x, y, fontKey, encodedText, false)
		return nil
	}

	if p.currentFont != nil {
		// Use standard font (ASCII/Latin-1 only)
		fontKey := p.getFontKey(*p.currentFont)
		encodedText := p.escapeString(text)
		p.drawTextInternal(x, y, fontKey, encodedText, true)
		return nil
	}

	return fmt.Errorf("no font set; call SetFont or SetTTFFont before DrawText")
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

// DrawImage draws an image at the specified position with the specified size.
// The image is transformed using a CTM (Current Transformation Matrix).
func (p *Page) DrawImage(img *Image, x, y, width, height float64) error {
	if img == nil {
		return fmt.Errorf("image cannot be nil")
	}

	// Add image to the page's image list
	p.images = append(p.images, img)

	// Get image resource name (Im1, Im2, etc.)
	imageKey := fmt.Sprintf("Im%d", len(p.images))

	// Write PDF operators to content stream
	// q: Save graphics state
	// a b c d e f cm: Transformation matrix
	// /Name Do: Draw XObject
	// Q: Restore graphics state
	fmt.Fprintf(&p.content, "q\n")
	fmt.Fprintf(&p.content, "%.2f %.2f %.2f %.2f %.2f %.2f cm\n", width, 0.0, 0.0, height, x, y)
	fmt.Fprintf(&p.content, "/%s Do\n", imageKey)
	fmt.Fprintf(&p.content, "Q\n")

	return nil
}

// SetTTFFont sets the current TTF font and size for subsequent text operations.
func (p *Page) SetTTFFont(f *TTFFont, size float64) error {
	if f == nil {
		return fmt.Errorf("TTF font cannot be nil")
	}

	p.currentTTFFont = f
	p.currentFont = nil // Clear standard font
	p.fontSize = size

	// Add font to the page's TTF font list
	if p.ttfFonts == nil {
		p.ttfFonts = make(map[string]*TTFFont)
	}

	// Check if this font is already registered
	for _, existingFont := range p.ttfFonts {
		if existingFont == f {
			// Font already registered, no need to add again
			return nil
		}
	}

	// Generate new key for this font
	fontKey := p.getTTFFontKey(f)
	p.ttfFonts[fontKey] = f

	return nil
}

// DrawTextUTF8 draws UTF-8 encoded text at the specified position using the current TTF font.
// This method supports Unicode characters including Japanese, Chinese, Korean, etc.
//
// Deprecated: Use DrawText instead. DrawText now automatically handles both standard and TTF fonts.
func (p *Page) DrawTextUTF8(text string, x, y float64) error {
	return p.DrawText(text, x, y)
}

// getTTFFontKey returns the font resource name for a TTF font.
func (p *Page) getTTFFontKey(f *TTFFont) string {
	// Check if this font is already registered and return its key
	for key, existingFont := range p.ttfFonts {
		if existingFont == f {
			return key
		}
	}

	// Generate a unique key based on current font count
	// Use F15+ to avoid conflicts with standard fonts (F1-F14)
	if p.ttfFonts == nil {
		return "F15"
	}

	// Count existing TTF fonts to determine the key
	return fmt.Sprintf("F%d", 15+len(p.ttfFonts))
}

// textToHexString converts UTF-8 text to hex string for PDF
// For Type0 fonts, we use UTF-16BE encoding
func (p *Page) textToHexString(text string) string {
	result := ""

	for _, r := range text {
		// Convert rune to UTF-16BE (simplified: only BMP characters)
		if r <= 0xFFFF {
			result += fmt.Sprintf("%04X", r)
		} else {
			// For characters outside BMP, use surrogate pairs
			// This is a simplified implementation
			result += fmt.Sprintf("%04X", r)
		}
	}

	return result
}

// textToGlyphIndices converts UTF-8 text to glyph indices for TTF fonts
// This ensures proper rendering by using actual glyph IDs from the font
func (p *Page) textToGlyphIndices(text string, ttfFont *TTFFont) (string, error) {
	var result string

	for _, r := range text {
		// Get the glyph index for this character
		glyphIndex, err := ttfFont.internal.GetGlyphIndex(r)
		if err != nil {
			return "", fmt.Errorf("failed to get glyph index for character %c (U+%04X): %w", r, r, err)
		}

		// Record glyph usage for ToUnicode CMap generation
		ttfFont.glyphsMutex.Lock()
		ttfFont.usedGlyphs[uint16(glyphIndex)] = r
		ttfFont.glyphsMutex.Unlock()

		// Convert glyph index to 4-digit hex string
		result += fmt.Sprintf("%04X", glyphIndex)
	}

	return result, nil
}

// DrawRuby draws ruby (furigana) text above base text
// Returns the width of the drawn text (maximum of base and ruby width)
func (p *Page) DrawRuby(rubyText RubyText, x, y float64, style RubyStyle) (float64, error) {
	// 現在のフォントとサイズを取得
	if p.currentFont == nil && p.currentTTFFont == nil {
		return 0, fmt.Errorf("no font set; call SetFont or SetTTFFont before DrawRuby")
	}

	baseFontSize := p.fontSize
	rubyFontSize := baseFontSize * style.SizeRatio

	// フォント名を取得（幅計算用）
	fontName := p.getCurrentFontName()

	// 親文字とルビの幅を計算
	baseWidth := estimateTextWidth(rubyText.Base, baseFontSize, fontName)
	rubyWidth := estimateTextWidth(rubyText.Ruby, rubyFontSize, fontName)

	// 最大幅を取得
	maxWidth := baseWidth
	if rubyWidth > maxWidth {
		maxWidth = rubyWidth
	}

	// ルビのX座標を計算（アラインメントに応じて）
	var rubyX float64
	switch style.Alignment {
	case RubyAlignCenter:
		rubyX = x + (baseWidth-rubyWidth)/2
	case RubyAlignLeft:
		rubyX = x
	case RubyAlignRight:
		rubyX = x + baseWidth - rubyWidth
	default:
		rubyX = x + (baseWidth-rubyWidth)/2 // デフォルトは中央揃え
	}

	// ルビのY座標を計算（親文字の上に配置）
	rubyY := y + baseFontSize + style.Offset

	// ルビテキストを描画
	originalFontSize := p.fontSize
	if p.currentTTFFont != nil {
		if err := p.SetTTFFont(p.currentTTFFont, rubyFontSize); err != nil {
			return 0, err
		}
		if err := p.DrawTextUTF8(rubyText.Ruby, rubyX, rubyY); err != nil {
			return 0, err
		}
	} else {
		if err := p.SetFont(StandardFont(p.currentFont.Name()), rubyFontSize); err != nil {
			return 0, err
		}
		if err := p.DrawText(rubyText.Ruby, rubyX, rubyY); err != nil {
			return 0, err
		}
	}

	// フォントサイズを元に戻す
	if p.currentTTFFont != nil {
		if err := p.SetTTFFont(p.currentTTFFont, originalFontSize); err != nil {
			return 0, err
		}
	} else {
		if err := p.SetFont(StandardFont(p.currentFont.Name()), originalFontSize); err != nil {
			return 0, err
		}
	}

	// 親文字を描画
	if p.currentTTFFont != nil {
		if err := p.DrawTextUTF8(rubyText.Base, x, y); err != nil {
			return 0, err
		}
	} else {
		if err := p.DrawText(rubyText.Base, x, y); err != nil {
			return 0, err
		}
	}

	return maxWidth, nil
}

// DrawRubyWithActualText draws ruby text with ActualText support for proper copy behavior
// ActualText allows controlling what text is copied when users copy the PDF content
func (p *Page) DrawRubyWithActualText(rubyText RubyText, x, y float64, style RubyStyle) (float64, error) {
	// 現在のフォントとサイズを取得
	if p.currentFont == nil && p.currentTTFFont == nil {
		return 0, fmt.Errorf("no font set; call SetFont or SetTTFFont before DrawRubyWithActualText")
	}

	// ActualTextの内容を決定
	var actualText string
	switch style.CopyMode {
	case RubyCopyBase:
		actualText = rubyText.Base
	case RubyCopyRuby:
		actualText = rubyText.Ruby
	case RubyCopyBoth:
		actualText = fmt.Sprintf("%s(%s)", rubyText.Base, rubyText.Ruby)
	default:
		actualText = rubyText.Base
	}

	// Begin marked content with ActualText
	fmt.Fprintf(&p.content, "/Span <</ActualText (%s)>> BDC\n", p.escapeString(actualText))

	// ルビを描画
	width, err := p.DrawRuby(rubyText, x, y, style)
	if err != nil {
		return 0, err
	}

	// End marked content
	fmt.Fprintf(&p.content, "EMC\n")

	return width, nil
}

// DrawRubyTexts draws multiple ruby texts in sequence
// Returns the total width of all drawn texts
func (p *Page) DrawRubyTexts(texts []RubyText, x, y float64, style RubyStyle, useActualText bool) (float64, error) {
	currentX := x
	totalWidth := 0.0

	for _, text := range texts {
		var width float64
		var err error

		if useActualText {
			width, err = p.DrawRubyWithActualText(text, currentX, y, style)
		} else {
			width, err = p.DrawRuby(text, currentX, y, style)
		}

		if err != nil {
			return totalWidth, err
		}

		currentX += width
		totalWidth += width
	}

	return totalWidth, nil
}

// getCurrentFontName returns the current font name for width estimation
func (p *Page) getCurrentFontName() string {
	if p.currentTTFFont != nil {
		return p.getTTFFontKey(p.currentTTFFont)
	}
	if p.currentFont != nil {
		return p.getFontKey(*p.currentFont)
	}
	return "F1" // デフォルト
}

// AddTextLayer はページにテキストレイヤーを追加する
// テキストは通常透明にして、画像の上に配置される（コピー・検索可能）
func (p *Page) AddTextLayer(layer TextLayer) error {
	if len(layer.Words) == 0 {
		return nil // 単語がない場合は何もしない
	}

	// フォントが設定されていない場合はデフォルトフォントを使用
	if p.currentFont == nil && p.currentTTFFont == nil {
		if err := p.SetFont(FontHelvetica, 12); err != nil {
			return fmt.Errorf("failed to set default font: %w", err)
		}
	}

	// Graphics state for opacity
	if layer.Opacity < 1.0 {
		fmt.Fprintf(&p.content, "q\n") // Save graphics state
		fmt.Fprintf(&p.content, "/GS1 gs\n")
	}

	// 各単語を描画
	for _, word := range layer.Words {
		if word.Text == "" {
			continue
		}

		// フォントサイズを単語の高さに合わせる
		fontSize := word.Bounds.Height
		if fontSize <= 0 {
			fontSize = 12 // デフォルトサイズ
		}

		// テキストを描画
		fmt.Fprintf(&p.content, "BT\n") // Begin Text

		// フォントとサイズを設定
		if p.currentTTFFont != nil {
			fontKey := p.getTTFFontKey(p.currentTTFFont)
			fmt.Fprintf(&p.content, "/%s %.2f Tf\n", fontKey, fontSize)
		} else if p.currentFont != nil {
			fontKey := p.getFontKey(*p.currentFont)
			fmt.Fprintf(&p.content, "/%s %.2f Tf\n", fontKey, fontSize)
		}

		// テキストレンダリングモードを設定
		fmt.Fprintf(&p.content, "%d Tr\n", layer.RenderMode)

		// 位置を設定
		fmt.Fprintf(&p.content, "%.2f %.2f Td\n", word.Bounds.X, word.Bounds.Y)

		// テキストを描画
		if p.currentTTFFont != nil {
			hexString := p.textToHexString(word.Text)
			fmt.Fprintf(&p.content, "<%s> Tj\n", hexString)
		} else {
			fmt.Fprintf(&p.content, "(%s) Tj\n", p.escapeString(word.Text))
		}

		fmt.Fprintf(&p.content, "ET\n") // End Text
	}

	// Restore graphics state
	if layer.Opacity < 1.0 {
		fmt.Fprintf(&p.content, "Q\n")
	}

	return nil
}

// AddTextLayerWords は個別の単語を追加する（簡易版）
func (p *Page) AddTextLayerWords(words []TextLayerWord) error {
	layer := NewTextLayer(words)
	return p.AddTextLayer(layer)
}

// AddInvisibleText は指定位置に透明テキストを追加
// 画像の特定箇所をコピー・検索可能にする簡易メソッド
func (p *Page) AddInvisibleText(text string, x, y, width, height float64) error {
	word := TextLayerWord{
		Text: text,
		Bounds: Rectangle{
			X:      x,
			Y:      y,
			Width:  width,
			Height: height,
		},
	}

	layer := TextLayer{
		Words:      []TextLayerWord{word},
		RenderMode: TextRenderInvisible,
		Opacity:    0.0,
	}

	return p.AddTextLayer(layer)
}
