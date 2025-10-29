package gopdf

import (
	"math"
	"sort"
	"strings"

	"github.com/ryomak/gopdf/internal/content"
	"github.com/ryomak/gopdf/internal/core"
	"github.com/ryomak/gopdf/internal/utils"
)

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
}

// TextBlock はテキストの論理的なブロック
type TextBlock struct {
	Text       string        // テキスト内容
	Elements   []TextElement // 構成要素
	Rect Rectangle     // バウンディングボックス
	Font       string        // 主要フォント
	FontSize   float64       // 主要フォントサイズ
	Color      Color         // テキスト色
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

// Rectangle は矩形領域
type Rectangle struct {
	X      float64 // 左下X座標
	Y      float64 // 左下Y座標
	Width  float64 // 幅
	Height float64 // 高さ
}

// ExtractPageLayout はページの完全なレイアウト情報を抽出
func (r *PDFReader) ExtractPageLayout(pageNum int) (*PageLayout, error) {
	// ページを取得
	page, err := r.r.GetPage(pageNum)
	if err != nil {
		return nil, err
	}

	// ページサイズを取得
	width, height := r.getPageSize(page)

	// コンテンツストリームを取得
	contentsData, err := r.r.GetPageContents(page)
	if err != nil {
		return nil, err
	}

	// コンテンツストリームをパース
	parser := content.NewStreamParser(contentsData)
	operations, err := parser.ParseOperations()
	if err != nil {
		return nil, err
	}

	// テキスト要素を抽出
	textExtractor := content.NewTextExtractor(operations, r.r, page)
	textElements, err := textExtractor.Extract()
	if err != nil {
		return nil, err
	}

	// TextElementsをTextBlocksにグループ化
	textBlocks := r.groupTextElements(convertTextElements(textElements))

	// 画像を抽出（位置情報付き）
	imageExtractor := content.NewImageExtractor(r.r)
	imageBlocks, err := imageExtractor.ExtractImagesWithPosition(page, operations)
	if err != nil {
		return nil, err
	}

	return &PageLayout{
		PageNum:    pageNum,
		Width:      width,
		Height:     height,
		TextBlocks: textBlocks,
		Images:     convertImageBlocks(imageBlocks),
	}, nil
}

// ExtractAllLayouts は全ページのレイアウトを抽出
func (r *PDFReader) ExtractAllLayouts() (map[int]*PageLayout, error) {
	pageCount := r.PageCount()
	layouts := make(map[int]*PageLayout)

	for i := 0; i < pageCount; i++ {
		layout, err := r.ExtractPageLayout(i)
		if err != nil {
			return nil, err
		}
		layouts[i] = layout
	}

	return layouts, nil
}

// getPageSize はページのサイズを取得
func (r *PDFReader) getPageSize(page core.Dictionary) (width, height float64) {
	// デフォルトサイズ（A4）
	width, height = 595.0, 842.0

	// /MediaBoxを取得
	mediaBoxObj, ok := page[core.Name("MediaBox")]
	if !ok {
		return
	}

	mediaBox, ok := mediaBoxObj.(core.Array)
	if !ok || len(mediaBox) < 4 {
		return
	}

	// [llx lly urx ury]
	x1 := toFloat64(mediaBox[0])
	y1 := toFloat64(mediaBox[1])
	x2 := toFloat64(mediaBox[2])
	y2 := toFloat64(mediaBox[3])

	width = x2 - x1
	height = y2 - y1

	return
}

// convertTextElements は内部型から公開型に変換
func convertTextElements(internalElements []content.TextElement) []TextElement {
	return utils.Map(internalElements, func(elem content.TextElement) TextElement {
		return TextElement{
			Text:   elem.Text,
			X:      elem.X,
			Y:      elem.Y,
			Width:  estimateTextWidth(elem.Text, elem.Size, elem.Font),
			Height: elem.Size,
			Font:   elem.Font,
			Size:   elem.Size,
		}
	})
}

// convertImageBlocks は内部型から公開型に変換
func convertImageBlocks(internalBlocks []content.ImageBlock) []ImageBlock {
	return utils.Map(internalBlocks, func(block content.ImageBlock) ImageBlock {
		return ImageBlock{
			ImageInfo: ImageInfo{
				Name:        block.Name,
				Width:       block.Width,
				Height:      block.Height,
				ColorSpace:  block.ColorSpace,
				BitsPerComp: block.BitsPerComp,
				Filter:      block.Filter,
				Data:        block.Data,
				Format:      ImageFormat(block.Format),
			},
			X:            block.X,
			Y:            block.Y,
			PlacedWidth:  block.PlacedWidth,
			PlacedHeight: block.PlacedHeight,
		}
	})
}

// groupTextElements はTextElementsをTextBlocksにグループ化
func (r *PDFReader) groupTextElements(elements []TextElement) []TextBlock {
	if len(elements) == 0 {
		return nil
	}

	// 1. Y座標でソート（上から下）
	sorted := make([]TextElement, len(elements))
	copy(sorted, elements)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Y > sorted[j].Y
	})

	var blocks []TextBlock
	var currentBlock []TextElement
	threshold := 5.0 // ピクセル単位の閾値

	for i, elem := range sorted {
		if i == 0 {
			currentBlock = []TextElement{elem}
			continue
		}

		// 前の要素との距離を計算
		prev := sorted[i-1]
		xDist := math.Abs(elem.X - (prev.X + prev.Width))
		yDist := math.Abs(elem.Y - prev.Y)

		// 近接している場合は同じブロック
		if yDist < threshold && xDist < prev.Size*2 {
			currentBlock = append(currentBlock, elem)
		} else {
			// 新しいブロック
			if len(currentBlock) > 0 {
				blocks = append(blocks, createTextBlock(currentBlock))
			}
			currentBlock = []TextElement{elem}
		}
	}

	// 最後のブロック
	if len(currentBlock) > 0 {
		blocks = append(blocks, createTextBlock(currentBlock))
	}

	return blocks
}

// createTextBlock はTextElementsからTextBlockを作成
func createTextBlock(elements []TextElement) TextBlock {
	// バウンディングボックスを計算
	minX, minY := elements[0].X, elements[0].Y
	maxX, maxY := elements[0].X+elements[0].Width, elements[0].Y+elements[0].Height

	var text strings.Builder
	var totalSize float64

	for i, elem := range elements {
		if i > 0 {
			text.WriteString(" ")
		}
		text.WriteString(elem.Text)

		totalSize += elem.Size

		minX = math.Min(minX, elem.X)
		minY = math.Min(minY, elem.Y)
		maxX = math.Max(maxX, elem.X+elem.Width)
		maxY = math.Max(maxY, elem.Y+elem.Height)
	}

	avgSize := totalSize / float64(len(elements))

	return TextBlock{
		Text:     text.String(),
		Elements: elements,
		Rect: Rectangle{
			X:      minX,
			Y:      minY,
			Width:  maxX - minX,
			Height: maxY - minY,
		},
		Font:     elements[0].Font,
		FontSize: avgSize,
		Color:    Color{R: 0, G: 0, B: 0}, // デフォルト黒
	}
}

func toFloat64(obj core.Object) float64 {
	switch v := obj.(type) {
	case core.Integer:
		return float64(v)
	case core.Real:
		return float64(v)
	default:
		return 0
	}
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
	sort.Slice(blocks, func(i, j int) bool {
		_, yi := blocks[i].Position()
		_, yj := blocks[j].Position()
		return yi > yj
	})

	return blocks
}

// SortedContentBlocks はコンテンツブロックをソート順で返す
// ソート順: Y座標（上から下）、同じY座標ならX座標（左から右）
func (pl *PageLayout) SortedContentBlocks() []ContentBlock {
	blocks := pl.ContentBlocks()

	sort.Slice(blocks, func(i, j int) bool {
		xi, yi := blocks[i].Position()
		xj, yj := blocks[j].Position()

		// Y座標で比較（上から下）
		if math.Abs(yi-yj) > 1.0 {
			return yi > yj
		}

		// X座標で比較（左から右）
		return xi < xj
	})

	return blocks
}
