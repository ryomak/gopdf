package gopdf

import (
	"sync"

	"github.com/ryomak/gopdf/internal/font"
)

var (
	defaultJPFont     *TTFFont
	defaultJPFontOnce sync.Once
	defaultJPFontErr  error
)

// TTFFont represents a TrueType Font for use in PDF documents
type TTFFont struct {
	internal    *font.TTFFont
	usedGlyphs  map[uint16]rune // glyphIndex → Unicode rune mapping
	glyphsMutex sync.Mutex       // Protect concurrent access to usedGlyphs
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

// TextWidth calculates the width of a text string at a given font size
func (f *TTFFont) TextWidth(text string, fontSize float64) (float64, error) {
	return f.internal.TextWidth(text, fontSize)
}

// DefaultJapaneseFont は埋め込まれた日本語フォント（Koruri）を返す
//
// 初回呼び出し時にフォントを読み込み、以降はキャッシュされた結果を返します。
// これにより、複数回呼び出してもパフォーマンスへの影響は最小限です。
//
// 使用フォント: Koruri (小瑠璃)
// ライセンス: Apache License 2.0
// 構成: M+ FONTS + Open Sans
//
// Example:
//
//	jpFont, err := gopdf.DefaultJapaneseFont()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	page.SetTTFFont(jpFont, 16)
//	page.DrawText("こんにちは、世界！", 50, 800)
func DefaultJapaneseFont() (*TTFFont, error) {
	defaultJPFontOnce.Do(func() {
		internalFont, err := font.DefaultJapaneseFont()
		if err != nil {
			defaultJPFontErr = err
			return
		}
		defaultJPFont = &TTFFont{
			internal:   internalFont,
			usedGlyphs: make(map[uint16]rune),
		}
	})
	return defaultJPFont, defaultJPFontErr
}

// GetDefaultJapaneseFontLicense はKoruriフォントのライセンステキストを返す
//
// ドキュメントやアプリケーションにライセンス情報を表示する場合に使用できます。
func GetDefaultJapaneseFontLicense() string {
	return font.GetDefaultJapaneseFontLicense()
}
