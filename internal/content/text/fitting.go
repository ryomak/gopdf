package text

import (
	"fmt"
	"strings"

	"github.com/ryomak/gopdf/internal/content/layout"
)

// FitOptions はテキストフィッティングのオプション
type FitOptions struct {
	MaxFontSize float64 // 最大フォントサイズ
	MinFontSize float64 // 最小フォントサイズ
	LineSpacing float64 // 行間倍率（1.0 = フォントサイズと同じ）
	Padding     float64 // パディング
	AllowShrink bool    // 縮小を許可
	AllowGrow   bool    // 拡大を許可
	Alignment   Align   // テキスト配置
}

// DefaultFitOptions はデフォルトのフィッティングオプション
func DefaultFitOptions() FitOptions {
	return FitOptions{
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

// WidthEstimator はテキスト幅を推定する関数型
type WidthEstimator func(text string, fontSize float64, fontName string) float64

// DefaultWidthEstimator はデフォルトのテキスト幅推定関数
func DefaultWidthEstimator(text string, fontSize float64, _ string) float64 {
	// 簡易的な幅計算
	// 英数字の平均幅は fontSizeの約60%
	avgCharWidth := fontSize * 0.6
	return float64(len(text)) * avgCharWidth
}

// Fit は矩形領域内にテキストをフィッティング
func Fit(text string, bounds layout.Rectangle, fontName string, opts FitOptions, estimateWidth WidthEstimator) (*FittedText, error) {
	if estimateWidth == nil {
		estimateWidth = DefaultWidthEstimator
	}

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
		lines := Wrap(text, availWidth, fontName, midSize, estimateWidth)
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

// Wrap はテキストを指定幅で改行
func Wrap(text string, maxWidth float64, fontName string, fontSize float64, estimateWidth WidthEstimator) []string {
	if estimateWidth == nil {
		estimateWidth = DefaultWidthEstimator
	}

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
			width := estimateWidth(testLine, fontSize, fontName)

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
func EstimateLines(text string, maxWidth float64, fontName string, fontSize float64, estimateWidth WidthEstimator) int {
	lines := Wrap(text, maxWidth, fontName, fontSize, estimateWidth)
	return len(lines)
}

// EstimateTotalHeight はテキストの総高さを推定
func EstimateTotalHeight(text string, maxWidth float64, fontName string, fontSize float64, lineSpacing float64, estimateWidth WidthEstimator) float64 {
	lineCount := EstimateLines(text, maxWidth, fontName, fontSize, estimateWidth)
	lineHeight := fontSize * lineSpacing
	return float64(lineCount) * lineHeight
}
