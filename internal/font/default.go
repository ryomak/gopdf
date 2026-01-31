package font

import (
	"fmt"
	"os"
	"runtime"
)

// LoadSystemJapaneseFont loads a Japanese font from system fonts.
func LoadSystemJapaneseFont() (*TTFFont, error) {
	fontPaths := getSystemJapaneseFontPaths()
	for _, path := range fontPaths {
		if _, err := os.Stat(path); err == nil {
			ttfFont, err := LoadTTF(path)
			if err == nil {
				return ttfFont, nil
			}
		}
	}
	return nil, fmt.Errorf("no Japanese font found on system")
}

// getSystemJapaneseFontPaths returns OS-specific Japanese font paths
func getSystemJapaneseFontPaths() []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{
			"/System/Library/Fonts/ヒラギノ角ゴシック W3.ttc",
			"/System/Library/Fonts/Hiragino Sans GB.ttc",
			"/Library/Fonts/Arial Unicode.ttf",
			os.Getenv("HOME") + "/Library/Fonts/NotoSansJP-Regular.ttf",
		}
	case "linux":
		return []string{
			"/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc",
			"/usr/share/fonts/noto-cjk/NotoSansCJK-Regular.ttc",
			"/usr/share/fonts/truetype/takao-gothic/TakaoPGothic.ttf",
			"/usr/share/fonts/opentype/ipafont-gothic/ipagp.ttf",
			"/usr/share/fonts/google-noto-cjk/NotoSansCJK-Regular.ttc",
		}
	case "windows":
		return []string{
			"C:\\Windows\\Fonts\\YuGothM.ttc",
			"C:\\Windows\\Fonts\\meiryo.ttc",
			"C:\\Windows\\Fonts\\msgothic.ttc",
		}
	default:
		return nil
	}
}
