# レイアウト自動調整機能 設計書

## 1. 概要

PageLayoutのコンテンツブロック（TextBlock、ImageBlock）を自動的に調整する汎用的な機能を提供する。

### 1.1. 目的

- **翻訳だけでなく**、テキスト編集、フォント変更、コンテンツ追加など、あらゆる場面でのレイアウト調整に対応
- はみ出し、重なり、空白の問題を自動的に解決
- ユーザーが細かい位置調整をしなくても良いようにする

### 1.2. ユースケース

1. **PDF翻訳**: 翻訳後のテキストが長くなってはみ出す
2. **テキスト編集**: 既存PDFのテキストを編集してレイアウトが崩れる
3. **フォント変更**: フォントを変更してテキストサイズが変わる
4. **コンテンツ追加**: 新しいブロックを追加して既存ブロックと重なる
5. **マージン調整**: 全体のマージンを変更してレイアウトを再配置

## 2. 設計

### 2.1. 基本概念

#### レイアウト調整の3つのレベル

1. **ブロック内調整**: 単一ブロック内でテキストを調整（FitText）
2. **ブロック間調整**: 複数ブロック間の位置関係を調整（AutoAdjustLayout）
3. **ページ分割**: ページに収まらない場合の複数ページ分割（SplitIntoPages）

### 2.2. データ構造

```go
package gopdf

// LayoutAdjustmentOptions はレイアウト自動調整のオプション
type LayoutAdjustmentOptions struct {
	// ブロック内調整
	EnableTextFitting bool           // テキストフィッティングを有効化
	FittingOptions    FitTextOptions // フィッティングオプション

	// ブロック間調整
	EnableAutoPosition bool    // 自動位置調整を有効化
	MinSpacing         float64 // ブロック間の最小間隔
	PageMargin         float64 // ページ端からのマージン

	// ページ分割
	EnablePageSplit bool    // 複数ページ分割を有効化
	MaxPageHeight   float64 // ページの最大高さ

	// 配置戦略
	Strategy LayoutStrategy // レイアウト戦略
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

	// StrategyFlowDown は上から下に流し込む
	StrategyFlowDown LayoutStrategy = "flow_down"
)

// DefaultLayoutAdjustmentOptions はデフォルトのオプション
func DefaultLayoutAdjustmentOptions() LayoutAdjustmentOptions {
	return LayoutAdjustmentOptions{
		EnableTextFitting:  true,
		FittingOptions:     DefaultFitTextOptions(),
		EnableAutoPosition: true,
		MinSpacing:         10.0,
		PageMargin:         20.0,
		EnablePageSplit:    false,
		MaxPageHeight:      842.0, // A4
		Strategy:           StrategyCompact,
	}
}
```

### 2.3. 公開API

```go
package gopdf

// AdjustLayout はPageLayoutを自動調整する
func (pl *PageLayout) AdjustLayout(opts LayoutAdjustmentOptions) error {
	// 1. ブロック内調整（テキストフィッティング）
	if opts.EnableTextFitting {
		if err := pl.fitAllTextBlocks(opts.FittingOptions); err != nil {
			return err
		}
	}

	// 2. ブロック間調整（位置調整）
	if opts.EnableAutoPosition {
		if err := pl.adjustBlockPositions(opts); err != nil {
			return err
		}
	}

	return nil
}

// SplitIntoPages はPageLayoutを複数ページに分割
func (pl *PageLayout) SplitIntoPages(opts LayoutAdjustmentOptions) ([]*PageLayout, error) {
	if !opts.EnablePageSplit {
		return []*PageLayout{pl}, nil
	}

	// ページ分割ロジック
	return pl.splitByHeight(opts.MaxPageHeight, opts.MinSpacing, opts.PageMargin)
}

// MoveBlock はブロックを移動（手動調整用）
func (pl *PageLayout) MoveBlock(blockType ContentBlockType, index int, offsetX, offsetY float64) error {
	switch blockType {
	case ContentBlockTypeText:
		if index >= 0 && index < len(pl.TextBlocks) {
			pl.TextBlocks[index].Rect.X += offsetX
			pl.TextBlocks[index].Rect.Y += offsetY
		}
	case ContentBlockTypeImage:
		if index >= 0 && index < len(pl.Images) {
			pl.Images[index].X += offsetX
			pl.Images[index].Y += offsetY
		}
	}
	return nil
}

// ResizeBlock はブロックをリサイズ（手動調整用）
func (pl *PageLayout) ResizeBlock(blockType ContentBlockType, index int, newWidth, newHeight float64) error {
	switch blockType {
	case ContentBlockTypeText:
		if index >= 0 && index < len(pl.TextBlocks) {
			pl.TextBlocks[index].Rect.Width = newWidth
			pl.TextBlocks[index].Rect.Height = newHeight
		}
	case ContentBlockTypeImage:
		if index >= 0 && index < len(pl.Images) {
			pl.Images[index].PlacedWidth = newWidth
			pl.Images[index].PlacedHeight = newHeight
		}
	}
	return nil
}

// DetectOverlaps は重なっているブロックを検出
func (pl *PageLayout) DetectOverlaps() []BlockOverlap {
	var overlaps []BlockOverlap

	blocks := pl.SortedContentBlocks()
	for i := 0; i < len(blocks); i++ {
		for j := i + 1; j < len(blocks); j++ {
			if blocksOverlap(blocks[i], blocks[j]) {
				overlaps = append(overlaps, BlockOverlap{
					Block1: blocks[i],
					Block2: blocks[j],
					Area:   calculateOverlapArea(blocks[i], blocks[j]),
				})
			}
		}
	}

	return overlaps
}

// BlockOverlap はブロックの重なり情報
type BlockOverlap struct {
	Block1 ContentBlock
	Block2 ContentBlock
	Area   float64 // 重なり面積
}
```

### 2.4. 内部実装

#### fitAllTextBlocks: ブロック内調整

```go
// fitAllTextBlocks はすべてのTextBlockをフィッティング
func (pl *PageLayout) fitAllTextBlocks(opts FitTextOptions) error {
	for i := range pl.TextBlocks {
		block := &pl.TextBlocks[i]

		// テキストをフィッティング
		fitted, err := FitText(block.Text, block.Rect, block.Font, opts)
		if err != nil {
			// フィッティングできない場合は警告して続行
			continue
		}

		// フォントサイズを更新
		block.FontSize = fitted.FontSize

		// 複数行の場合、高さを調整
		if len(fitted.Lines) > 1 {
			newHeight := float64(len(fitted.Lines)) * fitted.LineHeight
			// 高さが増えた場合のみ更新（縮小はしない）
			if newHeight > block.Rect.Height {
				block.Rect.Height = newHeight
			}
		}
	}

	return nil
}
```

#### adjustBlockPositions: ブロック間調整

```go
// adjustBlockPositions はブロックの位置を自動調整
func (pl *PageLayout) adjustBlockPositions(opts LayoutAdjustmentOptions) error {
	switch opts.Strategy {
	case StrategyCompact:
		return pl.adjustCompact(opts)
	case StrategyEvenSpacing:
		return pl.adjustEvenSpacing(opts)
	case StrategyFlowDown:
		return pl.adjustFlowDown(opts)
	case StrategyPreservePosition:
		return pl.adjustPreservePosition(opts)
	default:
		return pl.adjustCompact(opts)
	}
}

// adjustCompact は上に詰めて配置
func (pl *PageLayout) adjustCompact(opts LayoutAdjustmentOptions) error {
	blocks := pl.SortedContentBlocks()

	currentY := pl.Height - opts.PageMargin

	for _, block := range blocks {
		bounds := block.GetBounds()

		// 新しいY座標を計算
		newY := currentY - bounds.Height

		// ブロックを移動
		switch block.GetType() {
	case ContentBlockTypeText:
			tb := block.(TextBlock)
			offsetY := newY - tb.Rect.Y
			pl.moveTextBlockByValue(tb, 0, offsetY)
		case ContentBlockTypeImage:
			ib := block.(ImageBlock)
			offsetY := newY - ib.Y
			pl.moveImageBlockByValue(ib, 0, offsetY)
		}

		// 次のブロックの位置を更新
		currentY = newY - opts.MinSpacing
	}

	return nil
}

// adjustFlowDown は上から下に流し込む（重なりを防ぐ）
func (pl *PageLayout) adjustFlowDown(opts LayoutAdjustmentOptions) error {
	blocks := pl.SortedContentBlocks()
	if len(blocks) == 0 {
		return nil
	}

	currentY := pl.Height - opts.PageMargin

	for _, block := range blocks {
		bounds := block.GetBounds()
		_, blockY := block.GetPosition()

		// 前のブロックと重なる場合のみ調整
		if blockY+bounds.Height > currentY {
			newY := currentY - bounds.Height

			// ブロックを移動
			offsetY := newY - blockY
			pl.moveBlockByInterface(block, 0, offsetY)

			currentY = newY - opts.MinSpacing
		} else {
			// 重ならない場合は元の位置を保持
			currentY = blockY - opts.MinSpacing
		}
	}

	return nil
}
```

#### splitByHeight: ページ分割

```go
// splitByHeight は高さでページ分割
func (pl *PageLayout) splitByHeight(maxHeight, minSpacing, pageMargin float64) ([]*PageLayout, error) {
	var pages []*PageLayout

	currentPage := &PageLayout{
		Width:  pl.Width,
		Height: maxHeight,
	}
	currentY := maxHeight - pageMargin

	blocks := pl.SortedContentBlocks()

	for _, block := range blocks {
		bounds := block.GetBounds()

		// 現在のページに収まらない場合
		if currentY-bounds.Height < pageMargin {
			// 現在のページを確定
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

		// ブロックを追加
		newY := currentY - bounds.Height
		switch block.GetType() {
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

	// 最後のページを追加
	if len(currentPage.TextBlocks) > 0 || len(currentPage.Images) > 0 {
		pages = append(pages, currentPage)
	}

	return pages, nil
}
```

## 3. 使用例

### 3.1. 基本: 自動調整

```go
// PDFを読み込み
reader, _ := gopdf.Open("input.pdf")
layout, _ := reader.ExtractPageLayout(0)

// テキストを編集（翻訳、修正など）
for i := range layout.TextBlocks {
    layout.TextBlocks[i].Text = editText(layout.TextBlocks[i].Text)
}

// 自動調整
opts := gopdf.DefaultLayoutAdjustmentOptions()
opts.Strategy = gopdf.StrategyFlowDown // 重なりを防ぐ
layout.AdjustLayout(opts)

// 新しいPDFとして出力
doc := gopdf.New()
page := doc.AddPage(gopdf.CustomSize(layout.Width, layout.Height), gopdf.Portrait)
page.RenderLayout(layout)
doc.WriteTo(outputFile)
```

### 3.2. 高度: カスタム調整 + ページ分割

```go
reader, _ := gopdf.Open("input.pdf")
layout, _ := reader.ExtractPageLayout(0)

// 1. テキスト編集
for i := range layout.TextBlocks {
    layout.TextBlocks[i].Text = translate(layout.TextBlocks[i].Text)
}

// 2. 重なりをチェック
overlaps := layout.DetectOverlaps()
if len(overlaps) > 0 {
    fmt.Printf("Warning: %d overlaps detected\n", len(overlaps))
}

// 3. レイアウト自動調整
opts := gopdf.DefaultLayoutAdjustmentOptions()
opts.Strategy = gopdf.StrategyFlowDown
opts.EnablePageSplit = true
opts.MaxPageHeight = 842.0 // A4

// 単一ページで調整を試みる
layout.AdjustLayout(opts)

// ページに収まらない場合は分割
pages, _ := layout.SplitIntoPages(opts)

// 4. 各ページをレンダリング
doc := gopdf.New()
for _, pageLayout := range pages {
    page := doc.AddPage(gopdf.A4, gopdf.Portrait)
    page.RenderLayout(pageLayout)
}
doc.WriteTo(outputFile)
```

### 3.3. 手動調整と自動調整の組み合わせ

```go
layout, _ := reader.ExtractPageLayout(0)

// テキスト編集
layout.TextBlocks[0].Text = "Very long translated text..."

// 手動で特定のブロックを調整
layout.ResizeBlock(gopdf.ContentBlockTypeText, 0, 400, 100) // 幅と高さを拡大
layout.MoveBlock(gopdf.ContentBlockTypeText, 1, 0, -50)     // 2番目のブロックを下げる

// 残りを自動調整
opts := gopdf.DefaultLayoutAdjustmentOptions()
opts.Strategy = gopdf.StrategyFlowDown
layout.AdjustLayout(opts)

// レンダリング
page.RenderLayout(layout)
```

### 3.4. 翻訳専用の便利関数（ラッパー）

```go
// TranslateAndAdjust は翻訳とレイアウト調整を一度に行う
func TranslateAndAdjust(
    inputPath string,
    outputPath string,
    translator Translator,
    opts LayoutAdjustmentOptions,
) error {
    reader, _ := Open(inputPath)
    defer reader.Close()

    doc := New()

    for i := 0; i < reader.PageCount(); i++ {
        layout, _ := reader.ExtractPageLayout(i)

        // 翻訳
        for j := range layout.TextBlocks {
            translated, _ := translator.Translate(layout.TextBlocks[j].Text)
            layout.TextBlocks[j].Text = translated
        }

        // レイアウト調整
        layout.AdjustLayout(opts)

        // ページ分割が必要な場合
        pages, _ := layout.SplitIntoPages(opts)
        for _, pageLayout := range pages {
            page := doc.AddPage(A4, Portrait)
            page.RenderLayout(pageLayout)
        }
    }

    file, _ := os.Create(outputPath)
    defer file.Close()
    return doc.WriteTo(file)
}
```

## 4. テスト計画

### 4.1. ユニットテスト

```go
// TestAdjustLayout_Compact はCompact戦略のテスト
func TestAdjustLayout_Compact(t *testing.T) {
    layout := &PageLayout{
        Height: 842,
        TextBlocks: []TextBlock{
            {Rect: Rectangle{Y: 700, Height: 50}},
            {Rect: Rectangle{Y: 500, Height: 50}},
        },
    }

    opts := DefaultLayoutAdjustmentOptions()
    opts.Strategy = StrategyCompact
    layout.AdjustLayout(opts)

    // 上から詰めて配置されていることを確認
    assert.True(t, layout.TextBlocks[0].Rect.Y > layout.TextBlocks[1].Rect.Y)
    // 間隔が確保されていることを確認
    gap := layout.TextBlocks[0].Rect.Y - (layout.TextBlocks[1].Rect.Y + layout.TextBlocks[1].Rect.Height)
    assert.Equal(t, opts.MinSpacing, gap)
}

// TestDetectOverlaps は重なり検出のテスト
func TestDetectOverlaps(t *testing.T) {
    layout := &PageLayout{
        TextBlocks: []TextBlock{
            {Rect: Rectangle{X: 100, Y: 100, Width: 200, Height: 50}},
            {Rect: Rectangle{X: 150, Y: 120, Width: 200, Height: 50}}, // 重なる
        },
    }

    overlaps := layout.DetectOverlaps()
    assert.Equal(t, 1, len(overlaps))
}
```

## 5. 参考資料

- [docs/content_block_abstraction_design.md](./content_block_abstraction_design.md)
- [docs/text_block_grouping_design.md](./text_block_grouping_design.md)
- [text_fitting.go](../text_fitting.go)
- [translator.go](../translator.go)
