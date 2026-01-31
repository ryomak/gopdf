package layout

import (
	"math"
	"sort"
)

// GroupElementsByLine groups text elements into lines based on Y coordinates and font.
// Elements on the same Y coordinate with the same font are grouped together.
func GroupElementsByLine(elements []TextElement) [][]TextElement {
	if len(elements) == 0 {
		return nil
	}

	// Sort by Y coordinate (descending), then by X coordinate (ascending)
	sorted := make([]TextElement, len(elements))
	copy(sorted, elements)
	sort.Slice(sorted, func(i, j int) bool {
		if math.Abs(sorted[i].Y-sorted[j].Y) < 1.0 {
			return sorted[i].X < sorted[j].X
		}
		return sorted[i].Y > sorted[j].Y
	})

	var lines [][]TextElement
	currentLine := []TextElement{sorted[0]}

	for i := 1; i < len(sorted); i++ {
		elem := sorted[i]
		prevElem := sorted[i-1]

		// Calculate Y difference
		yDiff := math.Abs(elem.Y - prevElem.Y)
		// Average font size
		avgSize := (elem.Size + prevElem.Size) / 2
		// Same line threshold: 50% of font size
		lineThreshold := avgSize * 0.5

		// Check if font or size changed
		fontChanged := elem.Font != prevElem.Font
		sizeChanged := math.Abs(elem.Size-prevElem.Size) > 0.5 // 0.5pt difference

		if yDiff < lineThreshold && !fontChanged && !sizeChanged {
			// Same line with same font/size
			currentLine = append(currentLine, elem)
		} else {
			// New line (different Y, font, or size)
			lines = append(lines, currentLine)
			currentLine = []TextElement{elem}
		}
	}

	// Add the last line
	lines = append(lines, currentLine)

	return lines
}

// ShouldMergeLines determines if two lines should be merged into the same block.
func ShouldMergeLines(prevLine, currLine []TextElement) bool {
	if len(prevLine) == 0 || len(currLine) == 0 {
		return false
	}

	// Don't merge if fonts are different
	if !HaveSameFont(prevLine, currLine) {
		return false
	}

	// Calculate line spacing (PDF origin is at bottom, so prevLine.minY > currLine.maxY)
	prevMinY := MinY(prevLine)
	currMaxY := MaxY(currLine)
	lineSpacing := prevMinY - currMaxY

	// Average font size
	avgSize := (AvgFontSize(prevLine) + AvgFontSize(currLine)) / 2

	// Line spacing threshold: font size * 1.5
	// Normal paragraph line spacing is 1.2-1.5, so beyond this is a new paragraph
	lineSpacingThreshold := avgSize * 1.5

	// Check X coordinate alignment
	prevLeftX := MinX(prevLine)
	currLeftX := MinX(currLine)
	xDiff := math.Abs(prevLeftX - currLeftX)

	// X difference threshold: 50 points
	xThreshold := 50.0

	// Conditions: line spacing within threshold AND X coordinates are close
	return lineSpacing <= lineSpacingThreshold && xDiff <= xThreshold
}

// HaveSameFont checks if two lines have the same font and size.
func HaveSameFont(line1, line2 []TextElement) bool {
	if len(line1) == 0 || len(line2) == 0 {
		return true
	}
	// Compare representative font and size (first element)
	fontMatch := line1[0].Font == line2[0].Font
	sizeMatch := math.Abs(line1[0].Size-line2[0].Size) <= 0.5
	return fontMatch && sizeMatch
}
