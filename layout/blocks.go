package layout

// Color は色の表現
type Color struct {
	R, G, B float64
}

// TextBlock はテキストの論理的なブロック
type TextBlock struct {
	Text     string           // テキスト内容
	Elements []TextElement    // 構成要素
	Rect     Rectangle        // バウンディングボックス
	Font     string           // 主要フォント
	FontSize float64          // 主要フォントサイズ
	Color    Color            // テキスト色
}

// Bounds はブロックの境界矩形を返す（ContentBlockインターフェース実装）
func (tb TextBlock) Bounds() Rectangle {
	return tb.Rect
}

// Type はブロックの種類を返す（ContentBlockインターフェース実装）
func (tb TextBlock) Type() ContentBlockType {
	return ContentBlockTypeText
}

// Position はブロックの配置位置を返す（ContentBlockインターフェース実装）
func (tb TextBlock) Position() (x, y float64) {
	return tb.Rect.X, tb.Rect.Y
}

// ImageBlock は画像の配置情報
type ImageBlock struct {
	ImageInfo              // 画像データ（埋め込み）
	X            float64   // 配置X座標
	Y            float64   // 配置Y座標
	PlacedWidth  float64   // 表示幅
	PlacedHeight float64   // 表示高さ
}

// Bounds はブロックの境界矩形を返す（ContentBlockインターフェース実装）
func (ib ImageBlock) Bounds() Rectangle {
	return Rectangle{
		X:      ib.X,
		Y:      ib.Y,
		Width:  ib.PlacedWidth,
		Height: ib.PlacedHeight,
	}
}

// Type はブロックの種類を返す（ContentBlockインターフェース実装）
func (ib ImageBlock) Type() ContentBlockType {
	return ContentBlockTypeImage
}

// Position はブロックの配置位置を返す（ContentBlockインターフェース実装）
func (ib ImageBlock) Position() (x, y float64) {
	return ib.X, ib.Y
}
