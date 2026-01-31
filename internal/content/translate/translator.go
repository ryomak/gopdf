// Package translate はPDF翻訳機能を提供します
package translate

// Translator はテキスト翻訳のインターフェース
type Translator interface {
	// Translate はテキストを翻訳する
	Translate(text string) (string, error)
}

// Func は関数型Translator
type Func func(string) (string, error)

// Translate はFuncの実装
func (f Func) Translate(text string) (string, error) {
	return f(text)
}
