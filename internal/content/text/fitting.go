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

		// 日本語を含むかチェック
		if containsJapanese(paragraph) {
			// 日本語テキストは文字単位で改行（禁則処理付き）
			lines = append(lines, wrapJapanese(paragraph, maxWidth, fontName, fontSize, estimateWidth)...)
		} else {
			// 英語テキストは単語単位で改行
			lines = append(lines, wrapEnglish(paragraph, maxWidth, fontName, fontSize, estimateWidth)...)
		}
	}

	return lines
}

// wrapEnglish は英語テキストを単語単位で改行
func wrapEnglish(paragraph string, maxWidth float64, fontName string, fontSize float64, estimateWidth WidthEstimator) []string {
	var lines []string
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

	return lines
}

// wrapJapanese は日本語テキストを文字単位で改行（禁則処理付き）
func wrapJapanese(paragraph string, maxWidth float64, fontName string, fontSize float64, estimateWidth WidthEstimator) []string {
	var lines []string
	runes := []rune(paragraph)
	var currentLine strings.Builder

	for i := 0; i < len(runes); i++ {
		r := runes[i]

		// 現在の行に文字を追加してみる
		testLine := currentLine.String() + string(r)
		width := estimateWidth(testLine, fontSize, fontName)

		if width <= maxWidth {
			// 収まる場合
			currentLine.WriteRune(r)
		} else {
			// 収まらない場合、改行が必要
			if currentLine.Len() > 0 {
				// 禁則処理: 行頭禁則文字の場合は前の行に含める
				if isLineStartProhibited(r) {
					currentLine.WriteRune(r)
					lines = append(lines, currentLine.String())
					currentLine.Reset()
					continue
				}

				// 禁則処理: 次の文字が行頭禁則の場合、現在の文字も次の行に送る
				if i+1 < len(runes) && isLineStartProhibited(runes[i+1]) {
					// 現在の行の最後の文字を取得
					currentStr := currentLine.String()
					currentRunes := []rune(currentStr)
					if len(currentRunes) > 0 {
						// 最後の文字を次の行に送る
						lines = append(lines, string(currentRunes[:len(currentRunes)-1]))
						currentLine.Reset()
						currentLine.WriteRune(currentRunes[len(currentRunes)-1])
						currentLine.WriteRune(r)
						continue
					}
				}

				lines = append(lines, currentLine.String())
				currentLine.Reset()
			}
			currentLine.WriteRune(r)
		}
	}

	// 残りの行を追加
	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	return lines
}

// containsJapanese は文字列に日本語（ひらがな、カタカナ、漢字）が含まれるかチェック
func containsJapanese(s string) bool {
	for _, r := range s {
		if isJapanese(r) {
			return true
		}
	}
	return false
}

// isJapanese は文字が日本語かどうかチェック
func isJapanese(r rune) bool {
	// ひらがな: U+3040-U+309F
	// カタカナ: U+30A0-U+30FF
	// 漢字（CJK統合漢字）: U+4E00-U+9FFF
	// 全角記号: U+3000-U+303F
	return (r >= 0x3040 && r <= 0x309F) || // ひらがな
		(r >= 0x30A0 && r <= 0x30FF) || // カタカナ
		(r >= 0x4E00 && r <= 0x9FFF) || // 漢字
		(r >= 0x3000 && r <= 0x303F) // 全角記号
}

// isLineStartProhibited は行頭禁則文字かどうかチェック
func isLineStartProhibited(r rune) bool {
	// 行頭に置いてはいけない文字
	prohibited := "、。，．・：；？！゛゜´｀¨＾￣＿ヽヾゝゞ〃仝々〆〇ー―‐／＼～∥｜…‥'" +
		"）〕］｝〉》」』】°′″℃％‰" +
		"ぁぃぅぇぉっゃゅょゎゕゖ" + // 小書きひらがな
		"ァィゥェォッャュョヮヵヶ" // 小書きカタカナ
	return strings.ContainsRune(prohibited, r)
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
