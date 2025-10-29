package gopdf

import (
	"math"
	"sort"
	"strings"

	"github.com/ryomak/gopdf/internal/content"
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

	// 画像を抽出（位置情報付き）
	imageExtractor := content.NewImageExtractor(r.r)
	imageBlocks, err := imageExtractor.ExtractImagesWithPosition(page, operations)
	if err != nil {
		return nil, err
	}

	convertedImageBlocks := convertImageBlocks(imageBlocks)

	// TextElementsをTextBlocksにグループ化（画像を考慮）
	textBlocks := r.groupTextElementsWithImages(
		convertTextElements(textElements),
		convertedImageBlocks,
	)

	return &PageLayout{
		PageNum:    pageNum,
		Width:      width,
		Height:     height,
		TextBlocks: textBlocks,
		Images:     convertedImageBlocks,
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
		}
	})
}

// YRange はY座標の範囲（PDFは下が原点）
type YRange struct {
	Min float64 // 下端
	Max float64 // 上端
}

// groupTextElements はTextElementsをTextBlocksにグループ化
// 設計書: docs/text_block_grouping_design.md
func (r *PDFReader) groupTextElements(elements []layout.TextElement) []layout.TextBlock {
	return r.groupTextElementsWithImages(elements, nil)
}

// groupTextElementsWithImages は画像の位置を考慮してTextElementsをグループ化
// 設計書: docs/unified_content_grouping_design.md
func (r *PDFReader) groupTextElementsWithImages(
	elements []layout.TextElement,
	images []layout.ImageBlock,
) []layout.TextBlock {
	if len(elements) == 0 {
		return nil
	}

	// 1. 行単位でグルーピング
	lines := groupElementsByLine(elements)
	if len(lines) == 0 {
		return nil
	}

	// 2. 画像のY座標範囲を取得
	imageRanges := getImageYRanges(images)

	// 3. ブロック単位でグルーピング（画像を考慮）
	var blocks []layout.TextBlock
	currentBlock := [][]layout.TextElement{lines[0]}

	for i := 1; i < len(lines); i++ {
		prevLine := lines[i-1]
		currLine := lines[i]

		// 前の行と現在の行の間に画像があるかチェック
		hasImage := len(currentBlock) > 0 &&
			hasImageBetween(currentBlock[len(currentBlock)-1], currLine, imageRanges)

		if hasImage {
			// 画像が挟まっているのでブロックを分割
			blocks = append(blocks, createTextBlockFromLines(currentBlock))
			currentBlock = [][]layout.TextElement{currLine}
		} else if shouldMergeLines(prevLine, currLine) {
			// 通常の判定で同じブロックとする
			currentBlock = append(currentBlock, currLine)
		} else {
			// 行間が広いので新しいブロック
			blocks = append(blocks, createTextBlockFromLines(currentBlock))
			currentBlock = [][]layout.TextElement{currLine}
		}
	}

	// 最後のブロックを追加
	if len(currentBlock) > 0 {
		blocks = append(blocks, createTextBlockFromLines(currentBlock))
	}

	return blocks
}

// groupElementsByLine は要素を行単位でグルーピング
func groupElementsByLine(elements []layout.TextElement) [][]layout.TextElement {
	if len(elements) == 0 {
		return nil
	}

	// Y座標でソート（上から下）
	sorted := make([]layout.TextElement, len(elements))
	copy(sorted, elements)
	sort.Slice(sorted, func(i, j int) bool {
		// Y座標が同じ場合はX座標でソート（左から右）
		if math.Abs(sorted[i].Y-sorted[j].Y) < 1.0 {
			return sorted[i].X < sorted[j].X
		}
		return sorted[i].Y > sorted[j].Y
	})

	var lines [][]layout.TextElement
	currentLine := []layout.TextElement{sorted[0]}

	for i := 1; i < len(sorted); i++ {
		elem := sorted[i]
		prevElem := sorted[i-1]

		// Y座標の差を計算
		yDiff := math.Abs(elem.Y - prevElem.Y)
		// フォントサイズの平均
		avgSize := (elem.Size + prevElem.Size) / 2
		// 同じ行の閾値: フォントサイズの50%
		lineThreshold := avgSize * 0.5

		if yDiff < lineThreshold {
			// 同じ行
			currentLine = append(currentLine, elem)
		} else {
			// 新しい行
			lines = append(lines, currentLine)
			currentLine = []layout.TextElement{elem}
		}
	}

	// 最後の行を追加
	lines = append(lines, currentLine)

	return lines
}

// shouldMergeLines は2つの行を同じブロックにマージするべきか判定
func shouldMergeLines(prevLine, currLine []layout.TextElement) bool {
	if len(prevLine) == 0 || len(currLine) == 0 {
		return false
	}

	// 行間を計算（PDFは下が原点なので、prevLineの下端とcurrLineの上端の差）
	prevMinY := minY(prevLine)
	currMaxY := maxY(currLine)
	lineSpacing := prevMinY - currMaxY

	// フォントサイズの平均
	avgSize := (avgFontSize(prevLine) + avgFontSize(currLine)) / 2

	// 行間の閾値: フォントサイズ * 1.5
	// 通常の段落内の行間は1.2-1.5程度なので、これを超えたら段落が変わったと判定
	lineSpacingThreshold := avgSize * 1.5

	// X座標の範囲をチェック（段落の左端が揃っているか）
	prevLeftX := minX(prevLine)
	currLeftX := minX(currLine)
	xDiff := math.Abs(prevLeftX - currLeftX)

	// X座標の差の閾値: 50ポイント
	// インデントなどで多少ずれていても同じ段落と判定
	xThreshold := 50.0

	// 条件: 行間が閾値以内 AND X座標が近い
	return lineSpacing <= lineSpacingThreshold && xDiff <= xThreshold
}

// createTextBlockFromLines は行のリストからTextBlockを作成
func createTextBlockFromLines(lines [][]layout.TextElement) layout.TextBlock {
	if len(lines) == 0 {
		return layout.TextBlock{}
	}

	// 全要素を収集
	var allElements []layout.TextElement
	for _, line := range lines {
		allElements = append(allElements, line...)
	}

	// バウンディングボックスを計算
	minX, minY := allElements[0].X, allElements[0].Y
	maxX, maxY := allElements[0].X+allElements[0].Width, allElements[0].Y+allElements[0].Height

	var totalSize float64
	for _, elem := range allElements {
		totalSize += elem.Size
		minX = math.Min(minX, elem.X)
		minY = math.Min(minY, elem.Y)
		maxX = math.Max(maxX, elem.X+elem.Width)
		maxY = math.Max(maxY, elem.Y+elem.Height)
	}

	avgSize := totalSize / float64(len(allElements))

	// テキストを結合（行間に改行を入れる）
	text := combineBlockText(lines)

	return layout.TextBlock{
		Text:     text,
		Elements: allElements,
		Rect: layout.Rectangle{
			X:      minX,
			Y:      minY,
			Width:  maxX - minX,
			Height: maxY - minY,
		},
		Font:     allElements[0].Font,
		FontSize: avgSize,
		Color:    layout.Color{R: 0, G: 0, B: 0},
	}
}

// combineBlockText はブロック内のテキストを結合（行間に改行を保持）
func combineBlockText(lines [][]layout.TextElement) string {
	var result strings.Builder

	for i, line := range lines {
		if i > 0 {
			result.WriteString("\n") // 行間は改行
		}

		// 行内のテキストを結合（要素間の距離を考慮）
		for j, elem := range line {
			if j > 0 {
				// 前の要素との距離を計算
				prevElem := line[j-1]
				gap := elem.X - (prevElem.X + prevElem.Width)

				// 距離の閾値: フォントサイズの35%
				// これより大きい場合はスペースを入れる
				// 文字間のカーニングを考慮しつつ、単語間は分離
				threshold := prevElem.Size * 0.35

				if gap > threshold {
					result.WriteString(" ")
				}
			}
			result.WriteString(elem.Text)
		}
	}

	return result.String()
}

// ヘルパー関数
func minY(elements []layout.TextElement) float64 {
	if len(elements) == 0 {
		return 0
	}
	min := elements[0].Y
	for _, e := range elements[1:] {
		if e.Y < min {
			min = e.Y
		}
	}
	return min
}

func maxY(elements []layout.TextElement) float64 {
	if len(elements) == 0 {
		return 0
	}
	max := elements[0].Y + elements[0].Height
	for _, e := range elements[1:] {
		if e.Y+e.Height > max {
			max = e.Y + e.Height
		}
	}
	return max
}

func minX(elements []layout.TextElement) float64 {
	if len(elements) == 0 {
		return 0
	}
	min := elements[0].X
	for _, e := range elements[1:] {
		if e.X < min {
			min = e.X
		}
	}
	return min
}

func avgFontSize(elements []layout.TextElement) float64 {
	if len(elements) == 0 {
		return 0
	}
	sum := 0.0
	for _, e := range elements {
		sum += e.Size
	}
	return sum / float64(len(elements))
}

// createTextBlock はTextElementsからTextBlockを作成
func createTextBlock(elements []layout.TextElement) layout.TextBlock {
	// バウンディングボックスを計算
	minX, minY := elements[0].X, elements[0].Y
	maxX, maxY := elements[0].X+elements[0].Width, elements[0].Y+elements[0].Height

	var text strings.Builder
	var totalSize float64

	for i, elem := range elements {
		if i > 0 {
			// 前の要素との距離を計算
			prevElem := elements[i-1]
			gap := elem.X - (prevElem.X + prevElem.Width)

			// 距離の閾値: フォントサイズの35%
			// 文字間のカーニングを考慮しつつ、単語間は分離
			threshold := prevElem.Size * 0.35

			if gap > threshold {
				text.WriteString(" ")
			}
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
		Rect: layout.Rectangle{
			X:      minX,
			Y:      minY,
			Width:  maxX - minX,
			Height: maxY - minY,
		},
		Font:     elements[0].Font,
		FontSize: avgSize,
		Color:    layout.Color{R: 0, G: 0, B: 0}, // デフォルト黒
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

// getImageYRanges は画像のY座標範囲のリストを取得
func getImageYRanges(images []layout.ImageBlock) []YRange {
	if len(images) == 0 {
		return nil
	}

	ranges := make([]YRange, len(images))
	for i, img := range images {
		ranges[i] = YRange{
			Min: img.Y,
			Max: img.Y + img.PlacedHeight,
		}
	}
	return ranges
}

// hasImageBetween は2つの行の間に画像があるかチェック
func hasImageBetween(prevLine, currLine []layout.TextElement, imageRanges []YRange) bool {
	if len(prevLine) == 0 || len(currLine) == 0 || len(imageRanges) == 0 {
		return false
	}

	// 前の行の下端（最小Y）
	prevMinY := minY(prevLine)

	// 現在の行の上端（最大Y）
	currMaxY := maxY(currLine)

	// 2つの行の間のY座標範囲
	// PDFは下が原点なので、prevMinY > currMaxY のはず
	if prevMinY <= currMaxY {
		return false
	}

	betweenRange := YRange{
		Min: currMaxY,
		Max: prevMinY,
	}

	// この範囲に画像があるかチェック
	for _, imgRange := range imageRanges {
		if overlapsYRange(betweenRange, imgRange) {
			return true
		}
	}

	return false
}

// overlapsYRange は2つのY座標範囲が重なっているかチェック
func overlapsYRange(range1, range2 YRange) bool {
	// range1.Max < range2.Min または range2.Max < range1.Min なら重なっていない
	return !(range1.Max < range2.Min || range2.Max < range1.Min)
}

// flattenContentBlocks はページ境界を保持したままブロックをフラット化
// ページを跨いだ統合は行わない
func flattenContentBlocks(pageBlocks map[int][]layout.ContentBlock) []layout.ContentBlock {
	if len(pageBlocks) == 0 {
		return nil
	}

	// ページ順にソート
	pageNums := make([]int, 0, len(pageBlocks))
	for pageNum := range pageBlocks {
		pageNums = append(pageNums, pageNum)
	}
	sort.Ints(pageNums)

	// 単純に結合
	var allBlocks []layout.ContentBlock
	for _, pageNum := range pageNums {
		allBlocks = append(allBlocks, pageBlocks[pageNum]...)
	}

	return allBlocks
}

// mergeContentBlocksAcrossPages はページを跨いでコンテンツブロックを統合
// 設計書: docs/cross_page_block_merging_design.md
func mergeContentBlocksAcrossPages(pageBlocks map[int][]layout.ContentBlock) []layout.ContentBlock {
	if len(pageBlocks) == 0 {
		return nil
	}

	// 1. ページ順にソートして統合リストを作成
	var allBlocks []layout.ContentBlock
	pageNums := make([]int, 0, len(pageBlocks))
	for pageNum := range pageBlocks {
		pageNums = append(pageNums, pageNum)
	}
	sort.Ints(pageNums)

	for _, pageNum := range pageNums {
		allBlocks = append(allBlocks, pageBlocks[pageNum]...)
	}

	if len(allBlocks) == 0 {
		return nil
	}

	// 2. 連続するテキストブロックを統合
	var merged []layout.ContentBlock
	var currentTextBlock *layout.TextBlock

	for _, block := range allBlocks {
		switch block.Type() {
		case layout.ContentBlockTypeText:
			tb := block.(layout.TextBlock)

			if currentTextBlock == nil {
				// 新しいテキストブロック開始
				currentTextBlock = &tb
			} else if canMergeTextBlocks(*currentTextBlock, tb) {
				// 前のブロックと統合可能
				currentTextBlock.Text += "\n" + tb.Text
				currentTextBlock.Elements = append(currentTextBlock.Elements, tb.Elements...)
				// 境界を拡張
				updateTextBlockBounds(currentTextBlock, tb)
			} else {
				// 統合できないので前のブロックを確定
				merged = append(merged, *currentTextBlock)
				currentTextBlock = &tb
			}

		case layout.ContentBlockTypeImage:
			// 画像が来たらテキストブロックを確定
			if currentTextBlock != nil {
				merged = append(merged, *currentTextBlock)
				currentTextBlock = nil
			}
			merged = append(merged, block)
		}
	}

	// 最後のテキストブロックを追加
	if currentTextBlock != nil {
		merged = append(merged, *currentTextBlock)
	}

	return merged
}

// canMergeTextBlocks は2つのテキストブロックが統合可能か判定
func canMergeTextBlocks(block1, block2 layout.TextBlock) bool {
	// フォント名が同じ
	if block1.Font != block2.Font {
		return false
	}

	// フォントサイズが近い（±1ポイントの差は許容）
	sizeDiff := math.Abs(block1.FontSize - block2.FontSize)
	if sizeDiff > 1.0 {
		return false
	}

	// 色が同じ
	if block1.Color != block2.Color {
		return false
	}

	return true
}

// updateTextBlockBounds はテキストブロックの境界を拡張
func updateTextBlockBounds(target *layout.TextBlock, source layout.TextBlock) {
	minX := math.Min(target.Rect.X, source.Rect.X)
	minY := math.Min(target.Rect.Y, source.Rect.Y)

	maxX := math.Max(
		target.Rect.X+target.Rect.Width,
		source.Rect.X+source.Rect.Width,
	)
	maxY := math.Max(
		target.Rect.Y+target.Rect.Height,
		source.Rect.Y+source.Rect.Height,
	)

	target.Rect = layout.Rectangle{
		X:      minX,
		Y:      minY,
		Width:  maxX - minX,
		Height: maxY - minY,
	}
}
