package content

import (
	"github.com/ryomak/gopdf/internal/core"
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
func NewTextExtractor(operations []Operation) *TextExtractor {
	return &TextExtractor{
		operations: operations,
	}
}

// Extract はテキストを抽出する
func (e *TextExtractor) Extract() ([]TextElement, error) {
	var elements []TextElement

	// 初期化
	e.resetTextState()

	for _, op := range e.operations {
		switch op.Operator {
		case "BT": // Begin text
			e.resetTextMatrices()

		case "ET": // End text
			// テキストオブジェクト終了

		case "Tf": // Set font
			if len(op.Operands) >= 2 {
				e.currentFont = getString(op.Operands[0])
				e.fontSize = getNumber(op.Operands[1])
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
				text := getString(op.Operands[0])
				elem := e.createTextElement(text)
				elements = append(elements, elem)
			}

		case "TJ": // Show text with positioning
			if len(op.Operands) >= 1 {
				if array, ok := op.Operands[0].(core.Array); ok {
					for _, item := range array {
						if str, ok := item.(core.String); ok {
							elem := e.createTextElement(string(str))
							elements = append(elements, elem)
						}
						// 数値の場合は位置調整（今は無視）
					}
				}
			}

		case "'": // Move to next line and show text
			e.moveText(0, -e.leading)
			if len(op.Operands) >= 1 {
				text := getString(op.Operands[0])
				elem := e.createTextElement(text)
				elements = append(elements, elem)
			}

		case "\"": // Set word/char spacing, move to next line, show text
			if len(op.Operands) >= 3 {
				e.wordSpacing = getNumber(op.Operands[0])
				e.charSpacing = getNumber(op.Operands[1])
				e.moveText(0, -e.leading)
				text := getString(op.Operands[2])
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
	return TextElement{
		Text: text,
		X:    e.textMatrix[4], // e
		Y:    e.textMatrix[5], // f
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
func getString(obj core.Object) string {
	switch v := obj.(type) {
	case core.String:
		return string(v)
	case core.Name:
		return string(v)
	default:
		return ""
	}
}
