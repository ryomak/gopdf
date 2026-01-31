package gopdf

import (
	"fmt"
	"os"
	"runtime"
	"sync"

	"github.com/ryomak/gopdf/internal/font"
)

// TTFFont represents a TrueType Font for use in PDF documents
type TTFFont struct {
	internal    *font.TTFFont
	usedGlyphs  map[uint16]rune // glyphIndex → Unicode rune mapping
	glyphsMutex sync.Mutex      // Protect concurrent access to usedGlyphs
}

// LoadTTF loads a TrueType font from a file path
func LoadTTF(path string) (*TTFFont, error) {
	internalFont, err := font.LoadTTF(path)
	if err != nil {
		return nil, err
	}

	return &TTFFont{
		internal:   internalFont,
		usedGlyphs: make(map[uint16]rune),
	}, nil
}

// LoadTTFFromBytes loads a TrueType font from a byte slice
func LoadTTFFromBytes(data []byte) (*TTFFont, error) {
	internalFont, err := font.LoadTTFFromBytes(data)
	if err != nil {
		return nil, err
	}

	return &TTFFont{
		internal:   internalFont,
		usedGlyphs: make(map[uint16]rune),
	}, nil
}

// Name returns the font name
func (f *TTFFont) Name() string {
	return f.internal.Name()
}

// isFont implements Font interface
func (f *TTFFont) isFont() {}

// TextWidth calculates the width of a text string at a given font size
func (f *TTFFont) TextWidth(text string, fontSize float64) (float64, error) {
	return f.internal.TextWidth(text, fontSize)
}

// LoadSystemJapaneseFont loads a Japanese font from system fonts.
//
// Search order by OS:
//   - macOS: Hiragino Sans
//   - Linux: Noto Sans CJK, Takao Gothic, IPA Gothic
//   - Windows: Yu Gothic, Meiryo, MS Gothic
//
// Returns an error if no Japanese font is found on the system.
//
// Example:
//
//	jpFont, err := gopdf.LoadSystemJapaneseFont()
//	if err != nil {
//	    // Fall back to a bundled font or skip Japanese text
//	    log.Printf("Japanese font not found: %v", err)
//	}
//	page.SetFont(jpFont, 16)
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
