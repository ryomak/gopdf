package layout

import (
	"math"
	"sort"

	"github.com/ryomak/gopdf/layout"
)

// FlattenContentBlocks flattens page blocks while preserving page boundaries.
// Does not merge blocks across pages.
func FlattenContentBlocks(pageBlocks map[int][]layout.ContentBlock) []layout.ContentBlock {
	if len(pageBlocks) == 0 {
		return nil
	}

	// Sort by page number
	pageNums := make([]int, 0, len(pageBlocks))
	for pageNum := range pageBlocks {
		pageNums = append(pageNums, pageNum)
	}
	sort.Ints(pageNums)

	// Simply concatenate
	var allBlocks []layout.ContentBlock
	for _, pageNum := range pageNums {
		allBlocks = append(allBlocks, pageBlocks[pageNum]...)
	}

	return allBlocks
}

// MergeContentBlocksAcrossPages merges content blocks across pages.
// Design doc: docs/cross_page_block_merging_design.md
func MergeContentBlocksAcrossPages(pageBlocks map[int][]layout.ContentBlock) []layout.ContentBlock {
	if len(pageBlocks) == 0 {
		return nil
	}

	// 1. Sort by page number and create merged list
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

	// 2. Merge consecutive text blocks
	var merged []layout.ContentBlock
	var currentTextBlock *layout.TextBlock

	for _, block := range allBlocks {
		switch block.Type() {
		case layout.ContentBlockTypeText:
			tb := block.(layout.TextBlock)

			if currentTextBlock == nil {
				// Start new text block
				currentTextBlock = &tb
			} else if CanMergeTextBlocks(*currentTextBlock, tb) {
				// Can merge with previous block
				currentTextBlock.Text += "\n" + tb.Text
				currentTextBlock.Elements = append(currentTextBlock.Elements, tb.Elements...)
				// Expand bounds
				UpdateTextBlockBounds(currentTextBlock, tb)
			} else {
				// Cannot merge, finalize previous block
				merged = append(merged, *currentTextBlock)
				currentTextBlock = &tb
			}

		case layout.ContentBlockTypeImage:
			// Finalize text block when image appears
			if currentTextBlock != nil {
				merged = append(merged, *currentTextBlock)
				currentTextBlock = nil
			}
			merged = append(merged, block)
		}
	}

	// Add the last text block
	if currentTextBlock != nil {
		merged = append(merged, *currentTextBlock)
	}

	return merged
}

// CanMergeTextBlocks determines if two text blocks can be merged.
func CanMergeTextBlocks(block1, block2 layout.TextBlock) bool {
	// Same font name
	if block1.Font != block2.Font {
		return false
	}

	// Similar font size (±1 point tolerance)
	sizeDiff := math.Abs(block1.FontSize - block2.FontSize)
	if sizeDiff > 1.0 {
		return false
	}

	// Same color
	if block1.Color != block2.Color {
		return false
	}

	return true
}

// UpdateTextBlockBounds expands the bounds of target to include source.
func UpdateTextBlockBounds(target *layout.TextBlock, source layout.TextBlock) {
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
