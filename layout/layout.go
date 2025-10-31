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
	PageNum    int          // ページ番号（0-indexed）
	Width      float64      // ページ幅
	Height     float64      // ページ高さ
	TextBlocks []TextBlock  // テキストブロック
	Images     []ImageBlock // 画像ブロック
	PageCTM    *Matrix      // ページレベルのCTM（座標系変換情報）
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

	// Y軸が反転しているかチェック（CTMのd成分が負の場合）
	yAxisFlipped := false
	if pl.PageCTM != nil && pl.PageCTM.D < 0 {
		yAxisFlipped = true
	}

	// Y座標でソート（上から下）
	sort.Slice(blocks, func(i, j int) bool {
		_, yi := blocks[i].Position()
		_, yj := blocks[j].Position()
		if yAxisFlipped {
			// Y軸が反転している場合：高いY値が視覚的に下なので、大きい方を先に
			return yi > yj
		}
		// 標準座標系：高いY値が視覚的に上なので、大きい方を先に
		return yi > yj
	})

	return blocks
}

// SortedContentBlocks はコンテンツブロックをソート順で返す
// ソート順: Y座標（上から下）、同じY座標ならX座標（左から右）
// 注: 抽出された座標は既に標準PDF座標系（左下原点、Y軸上向き）に変換済み
func (pl *PageLayout) SortedContentBlocks() []ContentBlock {
	blocks := pl.ContentBlocks()

	sort.Slice(blocks, func(i, j int) bool {
		boundsI := blocks[i].Bounds()
		boundsJ := blocks[j].Bounds()

		// 上端（Y+Height）で比較（上から下）
		// PDF座標系: Y値が大きいほど上にある
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
