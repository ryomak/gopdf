package font

import (
	"sync"

	"github.com/ryomak/gopdf/internal/font/embedded"
)

var (
	defaultJPFont     *TTFFont
	defaultJPFontOnce sync.Once
	defaultJPFontErr  error
)

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
//	jpFont, err := font.DefaultJapaneseFont()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	page.SetTTFFont(jpFont, 16)
//	page.DrawText("こんにちは、世界！", 50, 800)
func DefaultJapaneseFont() (*TTFFont, error) {
	defaultJPFontOnce.Do(func() {
		defaultJPFont, defaultJPFontErr = LoadTTFFromBytes(embedded.KoruriRegular)
	})
	return defaultJPFont, defaultJPFontErr
}

// GetDefaultJapaneseFontLicense はKoruriフォントのライセンステキストを返す
//
// ドキュメントやアプリケーションにライセンス情報を表示する場合に使用できます。
func GetDefaultJapaneseFontLicense() string {
	return embedded.License
}
