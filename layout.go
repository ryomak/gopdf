package gopdf

import (
	"fmt"
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
// BlockOverlap はブロックの重なり情報
type BlockOverlap struct {
	Block1 ContentBlock // 1つ目のブロック
	Block2 ContentBlock // 2つ目のブロック
	Area   float64      // 重なり面積
}

// MoveBlock はブロックを移動する
func (pl *PageLayout) MoveBlock(blockType ContentBlockType, index int, offsetX, offsetY float64) error {
	switch blockType {
	case ContentBlockTypeText:
		if index < 0 || index >= len(pl.TextBlocks) {
			return fmt.Errorf("text block index %d out of range [0, %d)", index, len(pl.TextBlocks))
		}
		pl.TextBlocks[index].Rect.X += offsetX
		pl.TextBlocks[index].Rect.Y += offsetY
	case ContentBlockTypeImage:
		if index < 0 || index >= len(pl.Images) {
			return fmt.Errorf("image block index %d out of range [0, %d)", index, len(pl.Images))
		}
		pl.Images[index].X += offsetX
		pl.Images[index].Y += offsetY
	default:
		return fmt.Errorf("unsupported block type: %s", blockType)
	}
	return nil
}

// ResizeBlock はブロックをリサイズする
func (pl *PageLayout) ResizeBlock(blockType ContentBlockType, index int, newWidth, newHeight float64) error {
	switch blockType {
	case ContentBlockTypeText:
		if index < 0 || index >= len(pl.TextBlocks) {
			return fmt.Errorf("text block index %d out of range [0, %d)", index, len(pl.TextBlocks))
		}
		pl.TextBlocks[index].Rect.Width = newWidth
		pl.TextBlocks[index].Rect.Height = newHeight
	case ContentBlockTypeImage:
		if index < 0 || index >= len(pl.Images) {
			return fmt.Errorf("image block index %d out of range [0, %d)", index, len(pl.Images))
		}
		pl.Images[index].PlacedWidth = newWidth
		pl.Images[index].PlacedHeight = newHeight
	default:
		return fmt.Errorf("unsupported block type: %s", blockType)
	}
	return nil
}

// DetectOverlaps は重なっているブロックを検出する
func (pl *PageLayout) DetectOverlaps() []BlockOverlap {
	var overlaps []BlockOverlap

	blocks := pl.SortedContentBlocks()

	for i := 0; i < len(blocks); i++ {
		for j := i + 1; j < len(blocks); j++ {
			area := calculateOverlapArea(blocks[i], blocks[j])
			if area > 0 {
				overlaps = append(overlaps, BlockOverlap{
					Block1: blocks[i],
					Block2: blocks[j],
					Area:   area,
				})
			}
		}
	}

	return overlaps
}

// calculateOverlapArea は2つのブロックの重なり面積を計算する
func calculateOverlapArea(block1, block2 ContentBlock) float64 {
	bounds1 := block1.Bounds()
	bounds2 := block2.Bounds()

	// 矩形の右上と左下の座標を計算
	// PDFは左下原点なので、Y座標が大きい方が上
	left1 := bounds1.X
	right1 := bounds1.X + bounds1.Width
	bottom1 := bounds1.Y
	top1 := bounds1.Y + bounds1.Height

	left2 := bounds2.X
	right2 := bounds2.X + bounds2.Width
	bottom2 := bounds2.Y
	top2 := bounds2.Y + bounds2.Height

	// 重なっている範囲を計算
	overlapLeft := math.Max(left1, left2)
	overlapRight := math.Min(right1, right2)
	overlapBottom := math.Max(bottom1, bottom2)
	overlapTop := math.Min(top1, top2)

	// 重なりの幅と高さ
	overlapWidth := overlapRight - overlapLeft
	overlapHeight := overlapTop - overlapBottom

	// 重なっていない場合は0を返す
	if overlapWidth <= 0 || overlapHeight <= 0 {
		return 0
	}

	return overlapWidth * overlapHeight
}

// SplitIntoPages はPageLayoutを複数ページに分割する
func (pl *PageLayout) SplitIntoPages(maxHeight, minSpacing, pageMargin float64) ([]*PageLayout, error) {
	var pages []*PageLayout

	currentPage := &PageLayout{
		Width:  pl.Width,
		Height: maxHeight,
	}
	currentY := maxHeight - pageMargin

	blocks := pl.SortedContentBlocks()

	for _, block := range blocks {
		bounds := block.Bounds()

		// 現在のページに収まらない場合
		if currentY-bounds.Height < pageMargin {
			// 現在のページにコンテンツがある場合のみ追加
			if len(currentPage.TextBlocks) > 0 || len(currentPage.Images) > 0 {
				pages = append(pages, currentPage)
			}

			// 新しいページを作成
			currentPage = &PageLayout{
				Width:  pl.Width,
				Height: maxHeight,
			}
			currentY = maxHeight - pageMargin
		}

		// ブロックを新しいY座標で追加
		newY := currentY - bounds.Height
		switch block.Type() {
		case ContentBlockTypeText:
			tb := block.(TextBlock)
			tb.Rect.Y = newY
			currentPage.TextBlocks = append(currentPage.TextBlocks, tb)
		case ContentBlockTypeImage:
			ib := block.(ImageBlock)
			ib.Y = newY
			currentPage.Images = append(currentPage.Images, ib)
		}

		currentY = newY - minSpacing
	}

	// 最後のページを追加（空でも追加）
	pages = append(pages, currentPage)

	return pages, nil
}
