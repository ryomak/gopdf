package model

// TextElement は単一のテキスト要素を表す
type TextElement struct {
	Text   string  // テキスト内容
	X      float64 // X座標
	Y      float64 // Y座標
	Width  float64 // 幅
	Height float64 // 高さ
	Font   string  // フォント名
	Size   float64 // フォントサイズ
}
