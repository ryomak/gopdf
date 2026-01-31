package layout

import (
	"github.com/ryomak/gopdf/layout"
)

// GroupTextElements groups TextElements into TextBlocks.
// Design doc: docs/text_block_grouping_design.md
func GroupTextElements(elements []layout.TextElement) []layout.TextBlock {
	return GroupTextElementsWithImages(elements, nil)
}

// GroupTextElementsWithImages groups TextElements into TextBlocks, considering image positions.
// Design doc: docs/unified_content_grouping_design.md
func GroupTextElementsWithImages(
	elements []layout.TextElement,
	images []layout.ImageBlock,
) []layout.TextBlock {
	if len(elements) == 0 {
		return nil
	}

	// 1. Group by line
	lines := GroupElementsByLine(elements)
	if len(lines) == 0 {
		return nil
	}

	// 2. Get Y coordinate ranges of images
	imageRanges := GetImageYRanges(images)

	// 3. Group into blocks (considering images)
	var blocks []layout.TextBlock
	currentBlock := [][]layout.TextElement{lines[0]}

	for i := 1; i < len(lines); i++ {
		prevLine := lines[i-1]
		currLine := lines[i]

		// Check if there's an image between previous and current line
		hasImage := len(currentBlock) > 0 &&
			HasImageBetween(currentBlock[len(currentBlock)-1], currLine, imageRanges)

		if hasImage {
			// Image is between lines, split the block
			blocks = append(blocks, CreateTextBlockFromLines(currentBlock))
			currentBlock = [][]layout.TextElement{currLine}
		} else if ShouldMergeLines(prevLine, currLine) {
			// Normal judgment: same block
			currentBlock = append(currentBlock, currLine)
		} else {
			// Line spacing is large, new block
			blocks = append(blocks, CreateTextBlockFromLines(currentBlock))
			currentBlock = [][]layout.TextElement{currLine}
		}
	}

	// Add the last block
	if len(currentBlock) > 0 {
		blocks = append(blocks, CreateTextBlockFromLines(currentBlock))
	}

	return blocks
}
