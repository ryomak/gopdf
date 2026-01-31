package layout

import (
	"fmt"
	"math"
)

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
