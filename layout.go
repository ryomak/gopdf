package gopdf

import (
	"math"
	"sort"
	"strings"

	"github.com/ryomak/gopdf/internal/content"
	"github.com/ryomak/gopdf/internal/core"
	"github.com/ryomak/gopdf/internal/model"
	"github.com/ryomak/gopdf/internal/utils"
	"github.com/ryomak/gopdf/layout"
)

// 型エイリアス（後方互換性のため、ユーザーは layout パッケージを直接使うことを推奨）
type (
	// layout パッケージの型
	ContentBlock            = layout.ContentBlock
	ContentBlockType        = layout.ContentBlockType
	PageLayout              = layout.PageLayout
	TextBlock               = layout.TextBlock
	ImageBlock              = layout.ImageBlock
	BlockOverlap            = layout.BlockOverlap
	LayoutStrategy          = layout.LayoutStrategy
	LayoutAdjustmentOptions = layout.LayoutAdjustmentOptions

	// internal/model パッケージの型（ユーザーが直接使える）
	Rectangle   = model.Rectangle
	TextElement = model.TextElement
	ImageInfo   = model.ImageInfo
	ImageFormat = model.ImageFormat
)

// 定数エイリアス
const (
	// layout パッケージの定数
	ContentBlockTypeText  = layout.ContentBlockTypeText
	ContentBlockTypeImage = layout.ContentBlockTypeImage

	StrategyPreservePosition = layout.StrategyPreservePosition
	StrategyCompact          = layout.StrategyCompact
	StrategyEvenSpacing      = layout.StrategyEvenSpacing
	StrategyFlowDown         = layout.StrategyFlowDown
	StrategyFitContent       = layout.StrategyFitContent

	// internal/model パッケージの定数
	ImageFormatJPEG    = model.ImageFormatJPEG
	ImageFormatPNG     = model.ImageFormatPNG
	ImageFormatUnknown = model.ImageFormatUnknown
)

// DefaultLayoutAdjustmentOptions はデフォルトのレイアウト調整オプションを返す
func DefaultLayoutAdjustmentOptions() LayoutAdjustmentOptions {
	return layout.DefaultLayoutAdjustmentOptions()
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
		l, err := r.ExtractPageLayout(i)
		if err != nil {
			return nil, err
		}
		layouts[i] = l
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
func convertTextElements(internalElements []content.TextElement) []model.TextElement {
	return utils.Map(internalElements, func(elem content.TextElement) model.TextElement {
		return model.TextElement{
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
func convertImageBlocks(internalBlocks []content.ImageBlock) []layout.ImageBlock {
	return utils.Map(internalBlocks, func(block content.ImageBlock) layout.ImageBlock {
		return layout.ImageBlock{
			ImageInfo: model.ImageInfo{
				Name:        block.Name,
				Width:       block.Width,
				Height:      block.Height,
				ColorSpace:  block.ColorSpace,
				BitsPerComp: block.BitsPerComp,
				Filter:      block.Filter,
				Data:        block.Data,
				Format:      model.ImageFormat(block.Format),
			},
			X:            block.X,
			Y:            block.Y,
			PlacedWidth:  block.PlacedWidth,
			PlacedHeight: block.PlacedHeight,
		}
	})
}

// groupTextElements はTextElementsをTextBlocksにグループ化
func (r *PDFReader) groupTextElements(elements []model.TextElement) []layout.TextBlock {
	if len(elements) == 0 {
		return nil
	}

	// 1. Y座標でソート（上から下）
	sorted := make([]model.TextElement, len(elements))
	copy(sorted, elements)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Y > sorted[j].Y
	})

	var blocks []layout.TextBlock
	var currentBlock []model.TextElement
	threshold := 5.0 // ピクセル単位の閾値

	for i, elem := range sorted {
		if i == 0 {
			currentBlock = []model.TextElement{elem}
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
			currentBlock = []model.TextElement{elem}
		}
	}

	// 最後のブロック
	if len(currentBlock) > 0 {
		blocks = append(blocks, createTextBlock(currentBlock))
	}

	return blocks
}

// createTextBlock はTextElementsからTextBlockを作成
func createTextBlock(elements []model.TextElement) layout.TextBlock {
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

	return layout.TextBlock{
		Text:     text.String(),
		Elements: elements,
		Rect: model.Rectangle{
			X:      minX,
			Y:      minY,
			Width:  maxX - minX,
			Height: maxY - minY,
		},
		Font:     elements[0].Font,
		FontSize: avgSize,
		Color:    model.Color{R: 0, G: 0, B: 0}, // デフォルト黒
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

// adjustLayoutFitContentWithFitting はFitTextを使った高度なフィッティング
// layout.PageLayout.AdjustLayout() で StrategyFitContent を使うと簡易版が実行される
// こちらの関数を直接呼ぶと、より正確なフィッティングが可能
func adjustLayoutFitContentWithFitting(pl *PageLayout, opts LayoutAdjustmentOptions) error {
	// TextBlocksを調整
	for i := range pl.TextBlocks {
		block := &pl.TextBlocks[i]

		// 空のテキストはスキップ
		if block.Text == "" {
			continue
		}

		// フォント名を取得（設定されていない場合はHelveticaを使用）
		fontName := block.Font
		if fontName == "" {
			fontName = "Helvetica"
		}

		// 現在のフォントサイズで収まるかチェック
		// wrapTextで改行してから行数をカウント
		wrapped := wrapText(block.Text, block.Rect.Width, fontName, block.FontSize)
		lineHeight := block.FontSize * 1.2
		currentHeight := float64(len(wrapped)) * lineHeight

		// 収まる場合はフォントサイズを変更しない
		if currentHeight <= block.Rect.Height {
			continue
		}

		// 収まらない場合のみフィット
		result, err := FitText(
			block.Text,
			block.Rect,
			fontName,
			FitTextOptions{
				MaxFontSize: block.FontSize, // 現在のフォントサイズを最大とする
				MinFontSize: 6.0,
				LineSpacing: 1.2,
				Padding:     0,
				AllowShrink: true,
				AllowGrow:   false, // 拡大は許可しない
			},
		)

		// エラーが発生した場合は元のフォントサイズを維持
		if err != nil {
			continue
		}

		// フォントサイズを更新（元より小さい場合のみ）
		if result.FontSize < block.FontSize {
			block.FontSize = result.FontSize
		}
	}

	// ImageBlocksを調整（ブロックサイズがないので、最大サイズを制限する場合のみ）
	// 画像は元のサイズを保持するが、必要に応じてアスペクト比を維持しながら縮小
	// ここでは特に制限がないので、画像サイズはそのまま
	// 必要であれば、LayoutAdjustmentOptionsにMaxImageWidth/Heightを追加して制御可能

	return nil
}
