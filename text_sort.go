package gopdf

import (
	"math"
	"sort"
	"strings"
)

// SortTextElements はテキスト要素を読み順序でソートする
// PDFの座標系（左下原点）を考慮し、上から下、左から右の順序にする
func SortTextElements(elements []TextElement) []TextElement {
	if len(elements) == 0 {
		return elements
	}

	// 1. Y座標でグループ化（行の検出）
	lines := groupByLine(elements)

	// 2. 各行内でX座標でソート（左から右）
	for _, line := range lines {
		sort.Slice(line, func(i, j int) bool {
			return line[i].X < line[j].X
		})
	}

	// 3. 行をY座標の降順でソート（上から下）
	// PDFは左下原点なので、Y座標が大きい方が上
	sort.Slice(lines, func(i, j int) bool {
		return lines[i][0].Y > lines[j][0].Y
	})

	// フラット化
	result := make([]TextElement, 0, len(elements))
	for _, line := range lines {
		result = append(result, line...)
	}

	return result
}

// groupByLine は同じ行のテキスト要素をグループ化する
func groupByLine(elements []TextElement) [][]TextElement {
	if len(elements) == 0 {
		return nil
	}

	// Y座標でソート（降順）
	sorted := make([]TextElement, len(elements))
	copy(sorted, elements)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Y > sorted[j].Y
	})

	var lines [][]TextElement
	currentLine := []TextElement{sorted[0]}
	currentY := sorted[0].Y
	threshold := sorted[0].Size * 0.5 // Y座標の差の閾値

	for i := 1; i < len(sorted); i++ {
		elem := sorted[i]
		// Y座標の差が閾値以内なら同じ行
		if math.Abs(elem.Y-currentY) <= threshold {
			currentLine = append(currentLine, elem)
		} else {
			lines = append(lines, currentLine)
			currentLine = []TextElement{elem}
			currentY = elem.Y
			threshold = elem.Size * 0.5
		}
	}
	lines = append(lines, currentLine)

	return lines
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
