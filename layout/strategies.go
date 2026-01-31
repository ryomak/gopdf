package layout

import "fmt"

// adjustLayoutFlowDown は上から順に配置し、前のブロックとの間隔を保つ
func (pl *PageLayout) adjustLayoutFlowDown(opts LayoutAdjustmentOptions) error {
	blocks := pl.SortedContentBlocks()
	if len(blocks) == 0 {
		return nil
	}

	// ブロックとインデックスのマッピングを保持
	type blockInfo struct {
		blockType ContentBlockType
		index     int
	}

	blockIndexMap := make(map[interface{}]blockInfo)

	// TextBlocksのマッピング
	for i := range pl.TextBlocks {
		// ポインタではなく、テキスト内容で識別
		key := pl.TextBlocks[i].Text + fmt.Sprintf("_%f_%f", pl.TextBlocks[i].Rect.X, pl.TextBlocks[i].Rect.Width)
		blockIndexMap[key] = blockInfo{ContentBlockTypeText, i}
	}

	// Imagesのマッピング
	for i := range pl.Images {
		key := fmt.Sprintf("img_%f_%f_%f", pl.Images[i].X, pl.Images[i].PlacedWidth, pl.Images[i].PlacedHeight)
		blockIndexMap[key] = blockInfo{ContentBlockTypeImage, i}
	}

	getBlockInfo := func(block ContentBlock) (blockInfo, bool) {
		switch block.Type() {
		case ContentBlockTypeText:
			tb := block.(TextBlock)
			key := tb.Text + fmt.Sprintf("_%f_%f", tb.Rect.X, tb.Rect.Width)
			info, ok := blockIndexMap[key]
			return info, ok
		case ContentBlockTypeImage:
			ib := block.(ImageBlock)
			key := fmt.Sprintf("img_%f_%f_%f", ib.X, ib.PlacedWidth, ib.PlacedHeight)
			info, ok := blockIndexMap[key]
			return info, ok
		}
		return blockInfo{}, false
	}

	// 前のブロックの下端を追跡
	prevBottom := blocks[0].Bounds().Y

	for i := 1; i < len(blocks); i++ {
		currentBounds := blocks[i].Bounds()

		// 現在のブロックの理想的な上端位置（prevBottomの下、minSpacing分離す）
		idealTop := prevBottom - opts.MinSpacing

		// 現在のブロックの新しい下端位置
		newY := idealTop - currentBounds.Height

		// 現在の上端位置
		currentTop := currentBounds.Y + currentBounds.Height

		// 移動が必要かチェック（現在の上端が理想位置より上にある場合）
		if currentTop > idealTop {
			// ブロックを移動
			info, ok := getBlockInfo(blocks[i])
			if ok {
				switch info.blockType {
				case ContentBlockTypeText:
					pl.TextBlocks[info.index].Rect.Y = newY
				case ContentBlockTypeImage:
					pl.Images[info.index].Y = newY
				}
			}
			prevBottom = newY
		} else {
			// 移動不要、現在の位置を使用
			prevBottom = currentBounds.Y
		}
	}

	return nil
}

// adjustLayoutCompact はブロックを上に詰めて配置
func (pl *PageLayout) adjustLayoutCompact(opts LayoutAdjustmentOptions) error {
	blocks := pl.SortedContentBlocks()
	if len(blocks) == 0 {
		return nil
	}

	// ページトップから配置
	currentY := pl.Height - opts.PageMargin

	for _, block := range blocks {
		bounds := block.Bounds()
		newY := currentY - bounds.Height

		switch block.Type() {
		case ContentBlockTypeText:
			// 元のTextBlockを探して更新
			for i := range pl.TextBlocks {
				if pl.TextBlocks[i].Rect.X == bounds.X &&
					pl.TextBlocks[i].Rect.Y == bounds.Y &&
					pl.TextBlocks[i].Rect.Width == bounds.Width &&
					pl.TextBlocks[i].Rect.Height == bounds.Height {
					pl.TextBlocks[i].Rect.Y = newY
					break
				}
			}
		case ContentBlockTypeImage:
			for i := range pl.Images {
				if pl.Images[i].X == bounds.X &&
					pl.Images[i].Y == bounds.Y &&
					pl.Images[i].PlacedWidth == bounds.Width &&
					pl.Images[i].PlacedHeight == bounds.Height {
					pl.Images[i].Y = newY
					break
				}
			}
		}

		currentY = newY - opts.MinSpacing
	}

	return nil
}

// adjustLayoutEvenSpacing はブロックを均等間隔で配置
func (pl *PageLayout) adjustLayoutEvenSpacing(opts LayoutAdjustmentOptions) error {
	blocks := pl.SortedContentBlocks()
	if len(blocks) == 0 {
		return nil
	}

	// 全ブロックの高さの合計を計算
	totalHeight := float64(0)
	for _, block := range blocks {
		totalHeight += block.Bounds().Height
	}

	// 利用可能な空間
	availableSpace := pl.Height - 2*opts.PageMargin - totalHeight
	if availableSpace < 0 {
		// 空間が足りない場合はCompact戦略にフォールバック
		return pl.adjustLayoutCompact(opts)
	}

	// 均等な間隔を計算
	spacing := availableSpace / float64(len(blocks)-1)
	if spacing < opts.MinSpacing {
		spacing = opts.MinSpacing
	}

	// 配置
	currentY := pl.Height - opts.PageMargin

	for _, block := range blocks {
		bounds := block.Bounds()
		newY := currentY - bounds.Height

		switch block.Type() {
		case ContentBlockTypeText:
			for i := range pl.TextBlocks {
				if pl.TextBlocks[i].Rect.X == bounds.X &&
					pl.TextBlocks[i].Rect.Y == bounds.Y &&
					pl.TextBlocks[i].Rect.Width == bounds.Width &&
					pl.TextBlocks[i].Rect.Height == bounds.Height {
					pl.TextBlocks[i].Rect.Y = newY
					break
				}
			}
		case ContentBlockTypeImage:
			for i := range pl.Images {
				if pl.Images[i].X == bounds.X &&
					pl.Images[i].Y == bounds.Y &&
					pl.Images[i].PlacedWidth == bounds.Width &&
					pl.Images[i].PlacedHeight == bounds.Height {
					pl.Images[i].Y = newY
					break
				}
			}
		}

		currentY = newY - spacing
	}

	return nil
}

// adjustLayoutFitContent はブロックサイズを変えず、コンテンツをブロックに収める
// 注: この機能はgopdf.FitText関数に依存するため、layout/パッケージでは簡易実装のみ提供
// 完全な実装はgopdfパッケージ側で提供される
func (pl *PageLayout) adjustLayoutFitContent(opts LayoutAdjustmentOptions) error {
	// 基本的な実装はgopdfパッケージ側で提供
	// ここでは何もしない（または簡易的なフォントサイズ調整のみ）
	return nil
}
