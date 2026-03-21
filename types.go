package gopdf

import (
	"github.com/ryomak/gopdf/internal/content/text"
	"github.com/ryomak/gopdf/internal/content/translate"
)

// text パッケージの再エクスポート
type (
	Align         = text.Align
	FitOptions    = text.FitOptions
	FitOptionFunc = text.FitOptionFunc
	FittedText    = text.FittedText
)

// NewFitOptions creates FitOptions with the given functional options.
func NewFitOptions(opts ...FitOptionFunc) FitOptions {
	return text.NewFitOptions(opts...)
}

// FitOption functional options with WithFit prefix
var (
	WithFitMaxFontSize  = text.WithFitMaxFontSize
	WithFitMinFontSize  = text.WithFitMinFontSize
	WithFitLineSpacing  = text.WithFitLineSpacing
	WithFitPadding      = text.WithFitPadding
	WithFitAllowShrink  = text.WithFitAllowShrink
	WithFitAllowGrow    = text.WithFitAllowGrow
	WithFitAlignment    = text.WithFitAlignment
)

const (
	AlignLeft   = text.AlignLeft
	AlignCenter = text.AlignCenter
	AlignRight  = text.AlignRight
)

// translate パッケージの再エクスポート
type (
	Translator                = translate.Translator
	TranslateFunc             = translate.Func
	MyMemoryTranslator        = translate.MyMemoryTranslator
	LibreTranslateTranslator  = translate.LibreTranslateTranslator
)

// NewMyMemoryTranslator は無料のMyMemory翻訳APIを使用するTranslatorを作成
// sourceLang, targetLang は言語コード（例: "en", "ja"）
var NewMyMemoryTranslator = translate.NewMyMemoryTranslator

// NewLibreTranslateTranslator はLibreTranslate APIを使用するTranslatorを作成
// baseURL はLibreTranslateのエンドポイント（例: "http://localhost:5000"）
var NewLibreTranslateTranslator = translate.NewLibreTranslateTranslator

// FitText は text.Fit のラッパー
func FitText(s string, bounds Rectangle, fontName string, opts FitOptions) (*FittedText, error) {
	return text.Fit(s, bounds, fontName, opts, text.DefaultWidthEstimator)
}

// DefaultFitOptions は text.DefaultFitOptions のラッパー
func DefaultFitOptions() FitOptions {
	return text.DefaultFitOptions()
}
