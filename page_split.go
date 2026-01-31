package gopdf

// SplitOptions はページ分割のオプション
type SplitOptions struct {
	MinSpacing float64 // ブロック間の最小間隔（デフォルト: 10.0）
	PageMargin float64 // ページ端からのマージン（デフォルト: 50.0）
}

// DefaultSplitOptions はデフォルトのページ分割オプションを返す
func DefaultSplitOptions() SplitOptions {
	return SplitOptions{
		MinSpacing: 10.0,
		PageMargin: 50.0,
	}
}

// SplitContentBlocksIntoPages はコンテンツブロックをページに分割する
// 既存の layout.PageLayout.SplitIntoPages を使いやすくラップ
// 設計書: docs/cross_page_block_merging_design.md
func SplitContentBlocksIntoPages(
	blocks []ContentBlock,
	pageSize PageSize,
	options SplitOptions,
) ([]*PageLayout, error) {
	// デフォルト値の設定
	if options.MinSpacing == 0 {
		options.MinSpacing = 10.0
	}
	if options.PageMargin == 0 {
		options.PageMargin = 50.0
	}

	// PageLayoutを作成
	pageLayout := &PageLayout{
		Width:  pageSize.Width,
		Height: pageSize.Height,
	}

	// ブロックをTextBlocksとImagesに分類
	for _, block := range blocks {
		switch block.Type() {
		case ContentBlockTypeText:
			pageLayout.TextBlocks = append(pageLayout.TextBlocks, block.(TextBlock))
		case ContentBlockTypeImage:
			pageLayout.Images = append(pageLayout.Images, block.(ImageBlock))
		}
	}

	// 既存の SplitIntoPages を使用
	return pageLayout.SplitIntoPages(pageSize.Height, options.MinSpacing, options.PageMargin)
}

// SplitContentBlocksIntoPagesWithDefaults はデフォルトオプションでページ分割する
func SplitContentBlocksIntoPagesWithDefaults(
	blocks []ContentBlock,
	pageSize PageSize,
) ([]*PageLayout, error) {
	return SplitContentBlocksIntoPages(blocks, pageSize, DefaultSplitOptions())
}
