package text

import (
	"math"
	"sort"
)

// SortableTextElement is an interface for text elements that can be sorted.
type SortableTextElement interface {
	GetX() float64
	GetY() float64
	GetSize() float64
}

// GroupByLine groups text elements by Y coordinate (detecting lines).
func GroupByLine[T SortableTextElement](elements []T) [][]T {
	if len(elements) == 0 {
		return nil
	}

	// Sort by Y coordinate (descending)
	sorted := make([]T, len(elements))
	copy(sorted, elements)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].GetY() > sorted[j].GetY()
	})

	var lines [][]T
	currentLine := []T{sorted[0]}
	currentY := sorted[0].GetY()
	threshold := sorted[0].GetSize() * 0.5 // Y difference threshold

	for i := 1; i < len(sorted); i++ {
		elem := sorted[i]
		// Elements within threshold are on the same line
		if math.Abs(elem.GetY()-currentY) <= threshold {
			currentLine = append(currentLine, elem)
		} else {
			lines = append(lines, currentLine)
			currentLine = []T{elem}
			currentY = elem.GetY()
			threshold = elem.GetSize() * 0.5
		}
	}
	lines = append(lines, currentLine)

	return lines
}

// SortByReadingOrder sorts elements in reading order (top to bottom, left to right).
func SortByReadingOrder[T SortableTextElement](elements []T) []T {
	if len(elements) == 0 {
		return elements
	}

	// 1. Group by line
	lines := GroupByLine(elements)

	// 2. Sort each line by X coordinate (left to right)
	for _, line := range lines {
		sort.Slice(line, func(i, j int) bool {
			return line[i].GetX() < line[j].GetX()
		})
	}

	// 3. Sort lines by Y coordinate (descending = top to bottom in PDF)
	sort.Slice(lines, func(i, j int) bool {
		return lines[i][0].GetY() > lines[j][0].GetY()
	})

	// Flatten
	result := make([]T, 0, len(elements))
	for _, line := range lines {
		result = append(result, line...)
	}

	return result
}
