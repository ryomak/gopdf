package utils

import (
	"strings"
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
		if r >= 32 || r == '\t' || r == '\n' || r == '\r' {
			cleaned.WriteRune(r)
		}
	}

	return cleaned.String()
}
