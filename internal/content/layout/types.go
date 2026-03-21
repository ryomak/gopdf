// Package layout provides internal layout processing utilities.
package layout

import "sort"

// ContentBlock はページ内のコンテンツブロックを表す統一インターフェース
type ContentBlock interface {
	// Bounds はブロックの境界矩形を返す
	Bounds() Rectangle

	// Type はブロックの種類を返す
	Type() ContentBlockType

	// Position はブロックの配置位置を返す（左下座標）
	Position() (x, y float64)

	// WithY は新しいY座標でブロックのコピーを返す
	WithY(y float64) ContentBlock

	// AddToLayout はブロックをPageLayoutの適切なスライスに追加する
	AddToLayout(pl *PageLayout)
}

// ContentBlockType はコンテンツブロックの種類
type ContentBlockType string

const (
	// ContentBlockTypeText はテキストブロック
	ContentBlockTypeText ContentBlockType = "text"
	// ContentBlockTypeImage は画像ブロック
	ContentBlockTypeImage ContentBlockType = "image"
)

// PageLayout はページの完全なレイアウト情報
type PageLayout struct {
	PageNum            int          // ページ番号（0-indexed）
	Width              float64      // ページ幅
	Height             float64      // ページ高さ
	TextBlocks         []TextBlock  // テキストブロック
	Images             []ImageBlock // 画像ブロック
	PageCTM            *Matrix      // ページレベルのCTM（座標系変換情報）
	GraphicsOperations []byte       // 元のコンテンツストリームからのグラフィックス操作（テキスト以外）
}

// Rectangle は矩形領域
type Rectangle struct {
	X      float64 // 左下X座標
	Y      float64 // 左下Y座標
	Width  float64 // 幅
	Height float64 // 高さ
}

// ContentBlocks はページ内のすべてのコンテンツブロックをY座標順で返す
func (pl *PageLayout) ContentBlocks() []ContentBlock {
	var blocks []ContentBlock

	// TextBlocksを追加
	for _, tb := range pl.TextBlocks {
		blocks = append(blocks, tb)
	}

	// ImageBlocksを追加
	for _, ib := range pl.Images {
		blocks = append(blocks, ib)
	}

	// Y座標でソート（上から下）
	// 注: 座標は既に標準座標系に変換済み（Y値が大きいほど上）
	sort.Slice(blocks, func(i, j int) bool {
		_, yi := blocks[i].Position()
		_, yj := blocks[j].Position()
		return yi > yj // 大きい方を先に（上から下）
	})

	return blocks
}

// SortedContentBlocks はコンテンツブロックをソート順で返す
// ソート順: Y座標（上から下）、同じY座標ならX座標（左から右）
// 注: 座標は既に標準PDF座標系（左下原点、Y軸上向き）に変換済み
func (pl *PageLayout) SortedContentBlocks() []ContentBlock {
	blocks := pl.ContentBlocks()

	sort.Slice(blocks, func(i, j int) bool {
		boundsI := blocks[i].Bounds()
		boundsJ := blocks[j].Bounds()

		// 上端（Y+Height）で比較（上から下）
		// 座標は標準PDF座標系: Y値が大きいほど上にある
		// 読む順序: 上から下なので、Y値が大きい方を先に
		topI := boundsI.Y + boundsI.Height
		topJ := boundsJ.Y + boundsJ.Height

		const epsilon = 1.0
		if topI-topJ > epsilon || topJ-topI > epsilon {
			return topI > topJ // 上端が高い方（Y値が大きい方）を先に
		}

		// X座標で比較（左から右）
		return boundsI.X < boundsJ.X
	})

	return blocks
}

// BlockOverlap はブロックの重なり情報
type BlockOverlap struct {
	Block1 ContentBlock // 1つ目のブロック
	Block2 ContentBlock // 2つ目のブロック
	Area   float64      // 重なり面積
}

// LayoutStrategy はレイアウト調整の戦略
type LayoutStrategy string

const (
	// StrategyPreservePosition は元の位置をできるだけ保持
	StrategyPreservePosition LayoutStrategy = "preserve_position"

	// StrategyCompact は上に詰めて配置
	StrategyCompact LayoutStrategy = "compact"

	// StrategyEvenSpacing は均等間隔で配置
	StrategyEvenSpacing LayoutStrategy = "even_spacing"

	// StrategyFlowDown は上から下に流し込む（後続ブロックを自動調整）
	StrategyFlowDown LayoutStrategy = "flow_down"

	// StrategyFitContent はブロックサイズを変えず、コンテンツをブロックに収める
	StrategyFitContent LayoutStrategy = "fit_content"
)

// LayoutAdjustmentOptions はレイアウト自動調整のオプション
type LayoutAdjustmentOptions struct {
	// 配置戦略
	Strategy LayoutStrategy

	// ブロック間の最小間隔
	MinSpacing float64

	// ページ端からのマージン
	PageMargin float64
}

// DefaultLayoutAdjustmentOptions はデフォルトのオプション
func DefaultLayoutAdjustmentOptions() LayoutAdjustmentOptions {
	return LayoutAdjustmentOptions{
		Strategy:   StrategyCompact,
		MinSpacing: 10.0,
		PageMargin: 20.0,
	}
}

// TextElement はテキスト要素（循環参照を避けるため独自に定義）
type TextElement struct {
	Text   string
	X      float64
	Y      float64
	Width  float64
	Height float64
	Font   string
	Size   float64
}

// GetX returns the X coordinate (implements SortableTextElement interface)
func (e TextElement) GetX() float64 {
	return e.X
}

// GetY returns the Y coordinate (implements SortableTextElement interface)
func (e TextElement) GetY() float64 {
	return e.Y
}

// GetSize returns the font size (implements SortableTextElement interface)
func (e TextElement) GetSize() float64 {
	return e.Size
}

// ImageFormat は画像フォーマット
type ImageFormat string

const (
	// ImageFormatJPEG はJPEG形式
	ImageFormatJPEG ImageFormat = "jpeg"
	// ImageFormatPNG はPNG形式
	ImageFormatPNG ImageFormat = "png"
	// ImageFormatUnknown は不明な形式
	ImageFormatUnknown ImageFormat = "unknown"
)

// ImageInfo は画像情報
type ImageInfo struct {
	Name        string
	Width       int
	Height      int
	ColorSpace  string
	BitsPerComp int
	Filter      string
	Data        []byte
	Format      ImageFormat
}

// Color は色の表現
type Color struct {
	R, G, B float64
}

// TextBlock はテキストの論理的なブロック
type TextBlock struct {
	Text     string        // テキスト内容
	Elements []TextElement // 構成要素
	Rect     Rectangle     // バウンディングボックス
	Font     string        // 主要フォント
	FontSize float64       // 主要フォントサイズ
	Color    Color         // テキスト色
	IsBold   bool          // 太字フラグ
	IsItalic bool          // 斜体フラグ
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

// WithY は新しいY座標でTextBlockのコピーを返す（ContentBlockインターフェース実装）
func (tb TextBlock) WithY(y float64) ContentBlock {
	newTB := tb
	newTB.Rect.Y = y
	return newTB
}

// AddToLayout はTextBlockをPageLayoutのTextBlocksに追加する（ContentBlockインターフェース実装）
func (tb TextBlock) AddToLayout(pl *PageLayout) {
	pl.TextBlocks = append(pl.TextBlocks, tb)
}

// Matrix は変換行列（CTM: Current Transformation Matrix）
type Matrix struct {
	A, B, C, D, E, F float64 // [a b c d e f]
}

// ImageBlock は画像の配置情報
type ImageBlock struct {
	ImageInfo              // 画像データ（埋め込み）
	X            float64   // 配置X座標
	Y            float64   // 配置Y座標
	PlacedWidth  float64   // 表示幅
	PlacedHeight float64   // 表示高さ
	Transform    Matrix    // 変換行列（CTM）
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

// WithY は新しいY座標でImageBlockのコピーを返す（ContentBlockインターフェース実装）
func (ib ImageBlock) WithY(y float64) ContentBlock {
	newIB := ib
	newIB.Y = y
	return newIB
}

// AddToLayout はImageBlockをPageLayoutのImagesに追加する（ContentBlockインターフェース実装）
func (ib ImageBlock) AddToLayout(pl *PageLayout) {
	pl.Images = append(pl.Images, ib)
}
