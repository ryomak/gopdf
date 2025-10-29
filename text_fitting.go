package gopdf

import (
	"fmt"
	"strings"
)

// TextAlign はテキストの配置
type TextAlign int

const (
	AlignLeft TextAlign = iota
	AlignCenter
	AlignRight
)

// FitTextOptions はテキストフィッティングのオプション
type FitTextOptions struct {
	MaxFontSize float64   // 最大フォントサイズ
	MinFontSize float64   // 最小フォントサイズ
	LineSpacing float64   // 行間倍率（1.0 = フォントサイズと同じ）
	Padding     float64   // パディング
	AllowShrink bool      // 縮小を許可
	AllowGrow   bool      // 拡大を許可
	Alignment   TextAlign // テキスト配置
}

// DefaultFitTextOptions はデフォルトのフィッティングオプション
func DefaultFitTextOptions() FitTextOptions {
	return FitTextOptions{
		MaxFontSize: 24.0,
		MinFontSize: 6.0,
		LineSpacing: 1.2,
		Padding:     2.0,
		AllowShrink: true,
		AllowGrow:   false,
		Alignment:   AlignLeft,
	}
}

// FittedText はフィッティング結果
type FittedText struct {
	Lines      []string // 改行されたテキスト
	FontSize   float64  // 調整後のフォントサイズ
	LineHeight float64  // 行の高さ
}

// FitText は矩形領域内にテキストをフィッティング
func FitText(text string, bounds Rectangle, fontName string, opts FitTextOptions) (*FittedText, error) {
	// パディングを考慮
	availWidth := bounds.Width - opts.Padding*2
	availHeight := bounds.Height - opts.Padding*2

	if availWidth <= 0 || availHeight <= 0 {
		return nil, fmt.Errorf("bounds too small after padding")
	}

	// 2分探索でフォントサイズを決定
	minSize := opts.MinFontSize
	maxSize := opts.MaxFontSize
	var bestFit *FittedText

	// 最大20回の反復で収束させる
	for iteration := 0; iteration < 20 && maxSize-minSize > 0.1; iteration++ {
		midSize := (minSize + maxSize) / 2
		lineHeight := midSize * opts.LineSpacing

		// テキストを改行
		lines := wrapText(text, availWidth, fontName, midSize)
		totalHeight := float64(len(lines)) * lineHeight

		if totalHeight <= availHeight {
			// 収まる場合
			bestFit = &FittedText{
				Lines:      lines,
				FontSize:   midSize,
				LineHeight: lineHeight,
			}
			if opts.AllowGrow {
				minSize = midSize // もっと大きくできるか試す
			} else {
				break // 拡大しないので終了
			}
		} else {
			// 収まらない場合
			if opts.AllowShrink {
				maxSize = midSize // 小さくする
			} else {
				break // 縮小しないので終了
			}
		}
	}

	if bestFit == nil || bestFit.FontSize == 0 {
		return nil, fmt.Errorf("text does not fit in bounds")
	}

	return bestFit, nil
}

// FitTextInBlock はTextBlock内にテキストをフィッティング
func FitTextInBlock(text string, block TextBlock, fontName string, opts FitTextOptions) (*FittedText, error) {
	return FitText(text, block.Rect, fontName, opts)
}

// wrapText はテキストを指定幅で改行
func wrapText(text string, maxWidth float64, fontName string, fontSize float64) []string {
	// 空のテキストの場合
	if text == "" {
		return []string{""}
	}

	// 改行で分割
	paragraphs := strings.Split(text, "\n")
	var lines []string

	for _, paragraph := range paragraphs {
		// 段落が空の場合
		if paragraph == "" {
			lines = append(lines, "")
			continue
		}

		// 単語で分割
		words := strings.Fields(paragraph)
		var currentLine strings.Builder

		for _, word := range words {
			// 現在の行に単語を追加してみる
			testLine := currentLine.String()
			if testLine != "" {
				testLine += " "
			}
			testLine += word

			// テキスト幅を計算
			width := estimateTextWidth(testLine, fontSize, fontName)

			if width <= maxWidth {
				// 収まる場合
				if currentLine.Len() > 0 {
					currentLine.WriteString(" ")
				}
				currentLine.WriteString(word)
			} else {
				// 収まらない場合
				if currentLine.Len() > 0 {
					// 現在の行を確定
					lines = append(lines, currentLine.String())
					currentLine.Reset()
				}
				// 単語が1つでmaxWidthを超える場合は強制的に追加
				currentLine.WriteString(word)
			}
		}

		// 残りの行を追加
		if currentLine.Len() > 0 {
			lines = append(lines, currentLine.String())
		}
	}

	return lines
}

// EstimateLines はテキストが何行になるか推定
func EstimateLines(text string, maxWidth float64, fontName string, fontSize float64) int {
	lines := wrapText(text, maxWidth, fontName, fontSize)
	return len(lines)
}

// EstimateTotalHeight はテキストの総高さを推定
func EstimateTotalHeight(text string, maxWidth float64, fontName string, fontSize float64, lineSpacing float64) float64 {
	lineCount := EstimateLines(text, maxWidth, fontName, fontSize)
	lineHeight := fontSize * lineSpacing
	return float64(lineCount) * lineHeight
}
