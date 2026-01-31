package utils

import (
	"strings"
	"unicode"
)

// CleanControlCharacters removes unprintable control characters from text
// while preserving common whitespace characters (tab, newline, carriage return)
func CleanControlCharacters(text string) string {
	if text == "" {
		return text
	}

	var cleaned strings.Builder
	cleaned.Grow(len(text))

	for _, r := range text {
		// 許可する文字:
		// - 通常の印刷可能文字 (>= 32)
		// - タブ (\t = 9)
		// - 改行 (\n = 10)
		// - キャリッジリターン (\r = 13)
		// - 非ブレーク空白 (U+00A0)
		// - その他のUnicode印刷可能文字

		if r >= 32 {
			// 通常の印刷可能文字
			cleaned.WriteRune(r)
		} else if r == '\t' || r == '\n' || r == '\r' {
			// 許可された空白文字
			cleaned.WriteRune(r)
		} else {
			// 制御文字は除去
			// 必要に応じて置換文字を使用することも可能
			// cleaned.WriteRune('\uFFFD') // Unicode replacement character
			continue // スキップ
		}
	}

	return cleaned.String()
}

// HasControlCharacters checks if text contains control characters
// (excluding common whitespace characters)
func HasControlCharacters(text string) bool {
	for _, r := range text {
		if r < 32 && r != '\t' && r != '\n' && r != '\r' {
			return true
		}
	}
	return false
}

// NormalizeWhitespace normalizes whitespace in text
// - Converts multiple spaces to single space
// - Trims leading and trailing whitespace
// - Preserves newlines
func NormalizeWhitespace(text string) string {
	lines := strings.Split(text, "\n")

	for i, line := range lines {
		// 各行の前後の空白を削除
		line = strings.TrimSpace(line)

		// 複数の連続する空白を1つに
		line = strings.Join(strings.Fields(line), " ")

		lines[i] = line
	}

	return strings.Join(lines, "\n")
}

// IsPrintable checks if a rune is printable
func IsPrintable(r rune) bool {
	if unicode.IsPrint(r) {
		return true
	}
	// Allow common whitespace
	return r == '\t' || r == '\n' || r == '\r'
}
