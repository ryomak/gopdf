package content

import (
	"unicode/utf16"
	"unicode/utf8"

	"github.com/ryomak/gopdf/internal/core"
	"github.com/ryomak/gopdf/internal/reader"
	"github.com/ryomak/gopdf/internal/utils"
)

// TextElement はテキスト要素
type TextElement struct {
	Text string  // テキスト内容
	X    float64 // X座標
	Y    float64 // Y座標
	Font string  // フォント名
	Size float64 // フォントサイズ
}

// TextExtractor はテキストを抽出する
type TextExtractor struct {
	operations []Operation
	reader     *reader.Reader  // PDFリーダー (フォント情報取得用)
	page       core.Dictionary // ページリソース

	// フォント管理
	fontManager     *FontManager
	currentFontInfo *FontInfo

	// グラフィックス状態
	graphicsState      GraphicsState   // 現在のグラフィックス状態
	graphicsStateStack []GraphicsState // グラフィックス状態のスタック (q/Q用)
	pageLevelCTM       *Matrix         // ページレベルのCTM（最初のcmオペレータ）

	// テキスト状態
	textMatrix  [6]float64 // Current text matrix
	lineMatrix  [6]float64 // Current line matrix
	currentFont string
	fontSize    float64
	charSpacing float64
	wordSpacing float64
	leading     float64
}

// NewTextExtractor は新しいTextExtractorを作成する
func NewTextExtractor(operations []Operation, r *reader.Reader, page core.Dictionary) *TextExtractor {
	var fontManager *FontManager
	if r != nil {
		fontManager = NewFontManager(r)
	}

	return &TextExtractor{
		operations:     operations,
		reader:         r,
		page:           page,
		fontManager:    fontManager,
		graphicsState:  NewGraphicsState(), // CTMを単位行列で初期化
	}
}

// Extract はテキストを抽出する
func (e *TextExtractor) Extract() ([]TextElement, error) {
	var elements []TextElement

	// 初期化
	e.resetTextState()

	for _, op := range e.operations {
		switch op.Operator {
		case "q": // Save graphics state
			e.graphicsStateStack = append(e.graphicsStateStack, e.graphicsState.Clone())

		case "Q": // Restore graphics state
			if len(e.graphicsStateStack) > 0 {
				e.graphicsState = e.graphicsStateStack[len(e.graphicsStateStack)-1]
				e.graphicsStateStack = e.graphicsStateStack[:len(e.graphicsStateStack)-1]
			}

		case "cm": // Modify current transformation matrix
			if len(op.Operands) >= 6 {
				a := getNumber(op.Operands[0])
				b := getNumber(op.Operands[1])
				c := getNumber(op.Operands[2])
				d := getNumber(op.Operands[3])
				e_ := getNumber(op.Operands[4])
				f := getNumber(op.Operands[5])

				// 新しい変換行列を現在のCTMに乗算
				newMatrix := Matrix{A: a, B: b, C: c, D: d, E: e_, F: f}
				e.graphicsState.CTM = e.graphicsState.CTM.Multiply(newMatrix)

				// ページレベルのCTM（最初のcm）を記録
				if e.pageLevelCTM == nil && len(e.graphicsStateStack) == 0 {
					e.pageLevelCTM = &newMatrix
				}
			}

		case "BT": // Begin text
			e.resetTextMatrices()

		case "ET": // End text
			// テキストオブジェクト終了

		case "Tf": // Set font
			if len(op.Operands) >= 2 {
				e.currentFont = getString(op.Operands[0])
				e.fontSize = getNumber(op.Operands[1])

				// フォント情報を取得
				if e.fontManager != nil && e.page != nil {
					// ページリソースからフォント情報を取得
					pageResources, ok := e.page["Resources"]
					if ok {
						// 間接参照を解決
						if ref, isRef := pageResources.(*core.Reference); isRef {
							var err error
							pageResources, err = e.reader.ResolveReference(ref)
							if err == nil {
								if resDict, ok := pageResources.(core.Dictionary); ok {
									fontInfo, err := e.fontManager.GetFont(e.currentFont, resDict)
									if err == nil {
										e.currentFontInfo = fontInfo
									}
								}
							}
						} else if resDict, ok := pageResources.(core.Dictionary); ok {
							fontInfo, err := e.fontManager.GetFont(e.currentFont, resDict)
							if err == nil {
								e.currentFontInfo = fontInfo
							}
						}
					}
				}
			}

		case "Td": // Move text position
			if len(op.Operands) >= 2 {
				tx := getNumber(op.Operands[0])
				ty := getNumber(op.Operands[1])
				e.moveText(tx, ty)
			}

		case "TD": // Move text position and set leading
			if len(op.Operands) >= 2 {
				tx := getNumber(op.Operands[0])
				ty := getNumber(op.Operands[1])
				e.leading = -ty
				e.moveText(tx, ty)
			}

		case "Tm": // Set text matrix
			if len(op.Operands) >= 6 {
				e.setTextMatrix(op.Operands)
			}

		case "T*": // Move to next line
			e.moveText(0, -e.leading)

		case "Tj": // Show text
			if len(op.Operands) >= 1 {
				text := e.getTextString(op.Operands[0])
				elem := e.createTextElement(text)
				elements = append(elements, elem)
			}

		case "TJ": // Show text with positioning
			if len(op.Operands) >= 1 {
				if array, ok := utils.ExtractAs[core.Array](op.Operands[0]); ok {
					for _, item := range array {
						if str, ok := utils.ExtractAs[core.String](item); ok {
							text := e.getTextString(core.String(str))
							elem := e.createTextElement(text)
							elements = append(elements, elem)
						}
						// 数値の場合は位置調整（今は無視）
					}
				}
			}

		case "'": // Move to next line and show text
			e.moveText(0, -e.leading)
			if len(op.Operands) >= 1 {
				text := e.getTextString(op.Operands[0])
				elem := e.createTextElement(text)
				elements = append(elements, elem)
			}

		case "\"": // Set word/char spacing, move to next line, show text
			if len(op.Operands) >= 3 {
				e.wordSpacing = getNumber(op.Operands[0])
				e.charSpacing = getNumber(op.Operands[1])
				e.moveText(0, -e.leading)
				text := e.getTextString(op.Operands[2])
				elem := e.createTextElement(text)
				elements = append(elements, elem)
			}

		case "Tc": // Set character spacing
			if len(op.Operands) >= 1 {
				e.charSpacing = getNumber(op.Operands[0])
			}

		case "Tw": // Set word spacing
			if len(op.Operands) >= 1 {
				e.wordSpacing = getNumber(op.Operands[0])
			}

		case "TL": // Set text leading
			if len(op.Operands) >= 1 {
				e.leading = getNumber(op.Operands[0])
			}
		}
	}

	return elements, nil
}

// resetTextState はテキスト状態をリセットする
func (e *TextExtractor) resetTextState() {
	e.currentFont = ""
	e.fontSize = 0
	e.charSpacing = 0
	e.wordSpacing = 0
	e.leading = 0
	e.resetTextMatrices()
}

// resetTextMatrices はテキストマトリックスをリセットする
func (e *TextExtractor) resetTextMatrices() {
	// 単位行列
	e.textMatrix = [6]float64{1, 0, 0, 1, 0, 0}
	e.lineMatrix = [6]float64{1, 0, 0, 1, 0, 0}
}

// moveText はテキスト位置を移動する
func (e *TextExtractor) moveText(tx, ty float64) {
	// Tlm = Tlm * [1 0 0 1 tx ty]
	e.lineMatrix[4] += tx
	e.lineMatrix[5] += ty

	// Tm = Tlm
	e.textMatrix = e.lineMatrix
}

// setTextMatrix はテキストマトリックスを設定する
func (e *TextExtractor) setTextMatrix(operands []core.Object) {
	e.textMatrix[0] = getNumber(operands[0])
	e.textMatrix[1] = getNumber(operands[1])
	e.textMatrix[2] = getNumber(operands[2])
	e.textMatrix[3] = getNumber(operands[3])
	e.textMatrix[4] = getNumber(operands[4])
	e.textMatrix[5] = getNumber(operands[5])

	// Tlm = Tm
	e.lineMatrix = e.textMatrix
}

// createTextElement はテキスト要素を作成する
func (e *TextExtractor) createTextElement(text string) TextElement {
	// テキストマトリックスから座標を取得
	x := e.textMatrix[4] // e
	y := e.textMatrix[5] // f

	// Note: CTMの扱いは複雑です。
	// 一部のPDFでは、Tmの座標が既に変換後の空間にあるため、
	// CTMを適用すると二重変換になってしまいます。
	// 現時点では、Tmの座標をそのまま使用します。
	// 将来的には、より正確なCTM処理が必要かもしれません。

	return TextElement{
		Text: text,
		X:    x,
		Y:    y,
		Font: e.currentFont,
		Size: e.fontSize,
	}
}

// getNumber はオブジェクトから数値を取得する
func getNumber(obj core.Object) float64 {
	switch v := obj.(type) {
	case core.Integer:
		return float64(v)
	case core.Real:
		return float64(v)
	default:
		return 0
	}
}

// getString はオブジェクトから文字列を取得する
// PDFの文字列エンコーディング(PDFDocEncoding, UTF-16BE)を考慮する
func getString(obj core.Object) string {
	switch v := obj.(type) {
	case core.String:
		return decodePDFString([]byte(v))
	case core.Name:
		return string(v)
	default:
		return ""
	}
}

// getTextString はテキスト表示用の文字列を取得する
// ToUnicode CMapがあればそれを使用し、なければ通常のエンコーディングを使用
func (e *TextExtractor) getTextString(obj core.Object) string {
	switch v := obj.(type) {
	case core.String:
		data := []byte(v)

		// ToUnicode CMapがあれば優先的に使用
		if e.currentFontInfo != nil && e.currentFontInfo.ToUnicodeCMap != nil {
			result := e.currentFontInfo.ToUnicodeCMap.LookupString(data)
			if result != "" {
				return result
			}
		}

		// ToUnicode がない、または失敗した場合は通常のデコード
		return decodePDFString(data)

	case core.Name:
		return string(v)
	default:
		return ""
	}
}

// decodePDFString はPDF文字列をデコードする
func decodePDFString(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	// UTF-16BE BOM (0xFE 0xFF) をチェック
	if len(data) >= 2 && data[0] == 0xFE && data[1] == 0xFF {
		return decodeUTF16BE(data[2:])
	}

	// UTF-16LE BOM (0xFF 0xFE) をチェック
	if len(data) >= 2 && data[0] == 0xFF && data[1] == 0xFE {
		return decodeUTF16LE(data[2:])
	}

	// UTF-8として有効かを先にチェック
	// UTF-8マルチバイト文字は0x80以上のバイトを含むが、これを優先する
	if utf8.Valid(data) {
		// ASCII範囲のみの場合はどちらでも同じ
		// マルチバイト文字がある場合はUTF-8として扱う
		hasMultiByte := false
		for i := 0; i < len(data); {
			r, size := utf8.DecodeRune(data[i:])
			if size > 1 {
				hasMultiByte = true
				break
			}
			if r == utf8.RuneError {
				break
			}
			i += size
		}

		// マルチバイト文字があればUTF-8
		if hasMultiByte {
			return string(data)
		}
	}

	// PDFDocEncodingの特殊範囲(0x80-0x9F)が単一バイトで含まれているか確認
	// これはUTF-8マルチバイトシーケンスの一部ではない場合のみ
	hasPDFDocSpecialChars := false
	for _, b := range data {
		if b >= 0x80 && b <= 0x9F {
			hasPDFDocSpecialChars = true
			break
		}
	}

	// 特殊文字があり、UTF-8マルチバイトではない場合
	if hasPDFDocSpecialChars {
		return decodePDFDocEncoding(data)
	}

	// ASCIIまたはLatin-1範囲として処理
	if utf8.Valid(data) {
		return string(data)
	}

	// PDFDocEncodingとして処理（基本的にはLatin-1）
	return decodePDFDocEncoding(data)
}

// decodeUTF16BE はUTF-16BEをデコードする
func decodeUTF16BE(data []byte) string {
	if len(data)%2 != 0 {
		return ""
	}

	u16s := make([]uint16, 0, len(data)/2)
	for i := 0; i < len(data); i += 2 {
		u16s = append(u16s, uint16(data[i])<<8|uint16(data[i+1]))
	}

	return string(utf16.Decode(u16s))
}

// decodeUTF16LE はUTF-16LEをデコードする
func decodeUTF16LE(data []byte) string {
	if len(data)%2 != 0 {
		return ""
	}

	u16s := make([]uint16, 0, len(data)/2)
	for i := 0; i < len(data); i += 2 {
		u16s = append(u16s, uint16(data[i+1])<<8|uint16(data[i]))
	}

	return string(utf16.Decode(u16s))
}

// decodePDFDocEncoding はPDFDocEncodingをデコードする
// PDFDocEncodingは基本的にLatin-1と同じだが、0x80-0x9Fの範囲で特殊文字を定義
func decodePDFDocEncoding(data []byte) string {
	// PDFDocEncodingの0x80-0x9F範囲の特殊文字マッピング
	pdfDocEncodingTable := map[byte]rune{
		0x80: 0x2022, // BULLET
		0x81: 0x2020, // DAGGER
		0x82: 0x2021, // DOUBLE DAGGER
		0x83: 0x2026, // HORIZONTAL ELLIPSIS
		0x84: 0x2014, // EM DASH
		0x85: 0x2013, // EN DASH
		0x86: 0x0192, // LATIN SMALL LETTER F WITH HOOK
		0x87: 0x2044, // FRACTION SLASH
		0x88: 0x2039, // SINGLE LEFT-POINTING ANGLE QUOTATION MARK
		0x89: 0x203A, // SINGLE RIGHT-POINTING ANGLE QUOTATION MARK
		0x8A: 0x2212, // MINUS SIGN
		0x8B: 0x2030, // PER MILLE SIGN
		0x8C: 0x201E, // DOUBLE LOW-9 QUOTATION MARK
		0x8D: 0x201C, // LEFT DOUBLE QUOTATION MARK
		0x8E: 0x201D, // RIGHT DOUBLE QUOTATION MARK
		0x8F: 0x2018, // LEFT SINGLE QUOTATION MARK
		0x90: 0x2019, // RIGHT SINGLE QUOTATION MARK
		0x91: 0x201A, // SINGLE LOW-9 QUOTATION MARK
		0x92: 0x2122, // TRADE MARK SIGN
		0x93: 0xFB01, // LATIN SMALL LIGATURE FI
		0x94: 0xFB02, // LATIN SMALL LIGATURE FL
		0x95: 0x0141, // LATIN CAPITAL LETTER L WITH STROKE
		0x96: 0x0152, // LATIN CAPITAL LIGATURE OE
		0x97: 0x0160, // LATIN CAPITAL LETTER S WITH CARON
		0x98: 0x0178, // LATIN CAPITAL LETTER Y WITH DIAERESIS
		0x99: 0x017D, // LATIN CAPITAL LETTER Z WITH CARON
		0x9A: 0x0131, // LATIN SMALL LETTER DOTLESS I
		0x9B: 0x0142, // LATIN SMALL LETTER L WITH STROKE
		0x9C: 0x0153, // LATIN SMALL LIGATURE OE
		0x9D: 0x0161, // LATIN SMALL LETTER S WITH CARON
		0x9E: 0x017E, // LATIN SMALL LETTER Z WITH CARON
		0x9F: 0xFFFD, // REPLACEMENT CHARACTER
	}

	runes := make([]rune, 0, len(data))
	for _, b := range data {
		if b >= 0x80 && b <= 0x9F {
			if r, ok := pdfDocEncodingTable[b]; ok {
				runes = append(runes, r)
			} else {
				runes = append(runes, rune(b))
			}
		} else {
			runes = append(runes, rune(b))
		}
	}

	return string(runes)
}

// GetPageLevelCTM はページレベルのCTMを返す
func (e *TextExtractor) GetPageLevelCTM() *Matrix {
	return e.pageLevelCTM
}
