package gopdf

import (
	"math"
	"strings"
	"unicode"

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
	var prevText string

	for i, elem := range elements {
		// Y座標が大きく変わったら改行（新しい行）
		if i > 0 && math.Abs(elem.Y-prevY) > elem.Size*0.5 {
			result.WriteString("\n")
		} else if i > 0 {
			// 同じ行内でCJK文字が関係する場合はスペースを入れない
			if !needsSpaceBetween(prevText, elem.Text) {
				// スペース不要（CJK文字が関係する場合）
			} else {
				// 英語同士などの場合はスペースで区切る
				result.WriteString(" ")
			}
		}

		result.WriteString(elem.Text)
		prevText = elem.Text
		prevY = elem.Y
	}

	return result.String()
}

// needsSpaceBetween は2つのテキスト間にスペースが必要かを判定する
// 前のテキストの最後の文字か次のテキストの最初の文字がCJK文字の場合はスペース不要
func needsSpaceBetween(prev, next string) bool {
	if prev == "" || next == "" {
		return false
	}

	// 前のテキストの最後の文字を取得
	prevRunes := []rune(prev)
	lastRune := prevRunes[len(prevRunes)-1]

	// 次のテキストの最初の文字を取得
	nextRunes := []rune(next)
	firstRune := nextRunes[0]

	// どちらかがCJK文字ならスペース不要
	if isCJK(lastRune) || isCJK(firstRune) {
		return false
	}

	return true
}

// isCJK は文字がCJK（中国語、日本語、韓国語）文字かを判定する
func isCJK(r rune) bool {
	return unicode.Is(unicode.Han, r) || // 漢字
		unicode.Is(unicode.Hiragana, r) || // ひらがな
		unicode.Is(unicode.Katakana, r) || // カタカナ
		unicode.Is(unicode.Hangul, r) || // ハングル
		(r >= 0x3000 && r <= 0x303F) || // CJK句読点
		(r >= 0xFF00 && r <= 0xFFEF) // 全角英数字・記号
}
