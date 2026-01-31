package layout

// YRange represents a Y coordinate range (PDF origin is at bottom).
type YRange struct {
	Min float64 // Bottom edge
	Max float64 // Top edge
}

// GetImageYRanges returns a list of Y coordinate ranges from image blocks.
func GetImageYRanges(images []ImageBlock) []YRange {
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

// HasImageBetween checks if there's an image between two lines.
func HasImageBetween(prevLine, currLine []TextElement, imageRanges []YRange) bool {
	if len(prevLine) == 0 || len(currLine) == 0 || len(imageRanges) == 0 {
		return false
	}

	// Bottom edge of previous line (min Y)
	prevMinY := MinY(prevLine)

	// Top edge of current line (max Y)
	currMaxY := MaxY(currLine)

	// Y coordinate range between two lines
	// PDF origin is at bottom, so prevMinY > currMaxY
	if prevMinY <= currMaxY {
		return false
	}

	betweenRange := YRange{
		Min: currMaxY,
		Max: prevMinY,
	}

	// Check if any image is in this range
	for _, imgRange := range imageRanges {
		if OverlapsYRange(betweenRange, imgRange) {
			return true
		}
	}

	return false
}

// OverlapsYRange checks if two Y coordinate ranges overlap.
func OverlapsYRange(range1, range2 YRange) bool {
	// No overlap if range1.Max < range2.Min or range2.Max < range1.Min
	return !(range1.Max < range2.Min || range2.Max < range1.Min)
}
