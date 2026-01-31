package gopdf

import (
	"github.com/ryomak/gopdf/internal/content/text"
	"github.com/ryomak/gopdf/internal/content/translate"
)

// text パッケージの再エクスポート
type (
	Align      = text.Align
	FitOptions = text.FitOptions
	FittedText = text.FittedText
)

const (
	AlignLeft   = text.AlignLeft
	AlignCenter = text.AlignCenter
	AlignRight  = text.AlignRight
)

// translate パッケージの再エクスポート
type (
	Translator    = translate.Translator
	TranslateFunc = translate.Func
)

// FitText は text.Fit のラッパー
func FitText(s string, bounds Rectangle, fontName string, opts FitOptions) (*FittedText, error) {
	return text.Fit(s, bounds, fontName, opts, text.DefaultWidthEstimator)
}

// DefaultFitOptions は text.DefaultFitOptions のラッパー
func DefaultFitOptions() FitOptions {
	return text.DefaultFitOptions()
}
