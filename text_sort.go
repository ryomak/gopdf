package gopdf

import (
	"math"
	"strings"

	"github.com/ryomak/gopdf/internal/content/text"
)

// Note: math is used by TextElementsToString

// SortTextElements はテキスト要素を読み順序でソートする
// PDFの座標系（左下原点）を考慮し、上から下、左から右の順序にする
func SortTextElements(elements []TextElement) []TextElement {
	return text.SortByReadingOrder(elements)
}

// TextElementsToString はテキスト要素を文字列に変換する
// 読み順序でソートされていることを前提とする
func TextElementsToString(elements []TextElement) string {
	if len(elements) == 0 {
		return ""
	}

	var result strings.Builder
	prevY := elements[0].Y

	for i, elem := range elements {
		// Y座標が大きく変わったら改行（新しい行）
		if i > 0 && math.Abs(elem.Y-prevY) > elem.Size*0.5 {
			result.WriteString("\n")
		} else if i > 0 {
			// 同じ行内ではスペースで区切る
			result.WriteString(" ")
		}

		result.WriteString(elem.Text)
		prevY = elem.Y
	}

	return result.String()
}
