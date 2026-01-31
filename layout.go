package gopdf

import (
	"github.com/ryomak/gopdf/internal/content"
	contentlayout "github.com/ryomak/gopdf/internal/content/layout"
	"github.com/ryomak/gopdf/internal/content/text"
	"github.com/ryomak/gopdf/internal/core"
	"github.com/ryomak/gopdf/internal/utils"
	"github.com/ryomak/gopdf/layout"
)

// 型エイリアス（後方互換性のため、ユーザーは layout パッケージを直接使うことを推奨）
type (
	ContentBlock            = layout.ContentBlock
	ContentBlockType        = layout.ContentBlockType
	PageLayout              = layout.PageLayout
	TextBlock               = layout.TextBlock
	ImageBlock              = layout.ImageBlock
	Rectangle               = layout.Rectangle
	BlockOverlap            = layout.BlockOverlap
	LayoutStrategy          = layout.LayoutStrategy
	LayoutAdjustmentOptions = layout.LayoutAdjustmentOptions
)

// 定数エイリアス
const (
	ContentBlockTypeText  = layout.ContentBlockTypeText
	ContentBlockTypeImage = layout.ContentBlockTypeImage

	StrategyPreservePosition = layout.StrategyPreservePosition
	StrategyCompact          = layout.StrategyCompact
	StrategyEvenSpacing      = layout.StrategyEvenSpacing
	StrategyFlowDown         = layout.StrategyFlowDown
	StrategyFitContent       = layout.StrategyFitContent
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

	// ページレベルのCTMを取得
	var pageCTM *layout.Matrix
	var pageLevelCTM *content.Matrix
	if ctm := textExtractor.GetPageLevelCTM(); ctm != nil {
		pageCTM = &layout.Matrix{
			A: ctm.A,
			B: ctm.B,
			C: ctm.C,
			D: ctm.D,
			E: ctm.E,
			F: ctm.F,
		}
		pageLevelCTM = ctm
	}

	// 画像を抽出（位置情報付き、ページレベルCTMを渡す）
	imageExtractor := content.NewImageExtractor(r.r)
	imageBlocks, err := imageExtractor.ExtractImagesWithPosition(page, operations, pageLevelCTM)
	if err != nil {
		return nil, err
	}

	convertedImageBlocks := convertImageBlocks(imageBlocks)

	// TextElementsをTextBlocksにグループ化（画像を考慮）
	textBlocks := contentlayout.GroupTextElementsWithImages(
		convertTextElements(textElements),
		convertedImageBlocks,
	)

	// Y軸が反転している場合、座標を標準座標系に変換
	if pageCTM != nil && pageCTM.D < 0 {
		// TextBlocksの座標を変換
		for i := range textBlocks {
			// TextBlockのRect座標を変換
			textBlocks[i].Rect.Y = height - textBlocks[i].Rect.Y - textBlocks[i].Rect.Height

			// 各TextElementの座標も変換
			for j := range textBlocks[i].Elements {
				textBlocks[i].Elements[j].Y = height - textBlocks[i].Elements[j].Y
			}
		}

		// ImageBlocksの座標も変換
		for i := range convertedImageBlocks {
			convertedImageBlocks[i].Y = height - convertedImageBlocks[i].Y - convertedImageBlocks[i].PlacedHeight
		}
	}

	return &PageLayout{
		PageNum:    pageNum,
		Width:      width,
		Height:     height,
		TextBlocks: textBlocks,
		Images:     convertedImageBlocks,
		PageCTM:    pageCTM,
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
func convertTextElements(internalElements []content.TextElement) []layout.TextElement {
	return utils.Map(internalElements, func(elem content.TextElement) layout.TextElement {
		return layout.TextElement{
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
			ImageInfo: layout.ImageInfo{
				Name:        block.Name,
				Width:       block.Width,
				Height:      block.Height,
				ColorSpace:  block.ColorSpace,
				BitsPerComp: block.BitsPerComp,
				Filter:      block.Filter,
				Data:        block.Data,
				Format:      layout.ImageFormat(block.Format),
			},
			X:            block.X,
			Y:            block.Y,
			PlacedWidth:  block.PlacedWidth,
			PlacedHeight: block.PlacedHeight,
			Transform: layout.Matrix{
				A: block.Transform.A,
				B: block.Transform.B,
				C: block.Transform.C,
				D: block.Transform.D,
				E: block.Transform.E,
				F: block.Transform.F,
			},
		}
	})
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

// AdjustLayout はPageLayoutを自動調整する（gopdf固有の実装）
// layout.PageLayout.AdjustLayout() をオーバーライドして、FitText等の機能を使う
func AdjustLayout(pl *PageLayout, opts LayoutAdjustmentOptions) error {
	// StrategyFitContent以外は layout パッケージの実装を使用
	if opts.Strategy != StrategyFitContent {
		return pl.AdjustLayout(opts)
	}

	// StrategyFitContent はgopdf固有の実装
	return adjustLayoutFitContent(pl, opts)
}

// adjustLayoutFitContent はブロックサイズを変えず、コンテンツをブロックに収める
func adjustLayoutFitContent(pl *PageLayout, opts LayoutAdjustmentOptions) error {
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
		// text.Wrapで改行してから行数をカウント
		wrapped := text.Wrap(block.Text, block.Rect.Width, fontName, block.FontSize, text.DefaultWidthEstimator)
		lineHeight := block.FontSize * 1.2
		currentHeight := float64(len(wrapped)) * lineHeight

		// 収まる場合はフォントサイズを変更しない
		if currentHeight <= block.Rect.Height {
			continue
		}

		// 収まらない場合のみフィット
		result, err := text.Fit(
			block.Text,
			block.Rect,
			fontName,
			text.FitOptions{
				MaxFontSize: block.FontSize, // 現在のフォントサイズを最大とする
				MinFontSize: 6.0,
				LineSpacing: 1.2,
				Padding:     0,
				AllowShrink: true,
				AllowGrow:   false, // 拡大は許可しない
			},
			text.DefaultWidthEstimator,
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

// flattenContentBlocks はページ境界を保持したままブロックをフラット化
// ページを跨いだ統合は行わない
func flattenContentBlocks(pageBlocks map[int][]layout.ContentBlock) []layout.ContentBlock {
	return contentlayout.FlattenContentBlocks(pageBlocks)
}

// mergeContentBlocksAcrossPages はページを跨いでコンテンツブロックを統合
// 設計書: docs/cross_page_block_merging_design.md
func mergeContentBlocksAcrossPages(pageBlocks map[int][]layout.ContentBlock) []layout.ContentBlock {
	return contentlayout.MergeContentBlocksAcrossPages(pageBlocks)
}
