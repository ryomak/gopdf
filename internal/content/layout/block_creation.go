package layout

import (
	"math"
	"strings"

	"github.com/ryomak/gopdf/internal/utils"
	"github.com/ryomak/gopdf/layout"
)

// CreateTextBlockFromLines creates a TextBlock from a list of lines.
func CreateTextBlockFromLines(lines [][]layout.TextElement) layout.TextBlock {
	if len(lines) == 0 {
		return layout.TextBlock{}
	}

	// Collect all elements
	var allElements []layout.TextElement
	for _, line := range lines {
		allElements = append(allElements, line...)
	}

	// Calculate bounding box
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

	// Combine text (preserve line breaks)
	text := CombineBlockText(lines)

	// Detect font style
	isBold, isItalic := DetectFontStyle(allElements[0].Font)

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
		IsBold:   isBold,
		IsItalic: isItalic,
	}
}

// CreateTextBlock creates a TextBlock from text elements.
func CreateTextBlock(elements []layout.TextElement) layout.TextBlock {
	// Calculate bounding box
	minX, minY := elements[0].X, elements[0].Y
	maxX, maxY := elements[0].X+elements[0].Width, elements[0].Y+elements[0].Height

	var text strings.Builder
	var totalSize float64

	for i, elem := range elements {
		if i > 0 {
			// Calculate distance from previous element
			prevElem := elements[i-1]
			gap := elem.X - (prevElem.X + prevElem.Width)

			// Distance threshold: 35% of font size
			threshold := prevElem.Size * 0.35

			if gap > threshold {
				text.WriteString(" ")
			}
		}
		// Clean control characters before adding
		cleanText := utils.CleanControlCharacters(elem.Text)
		text.WriteString(cleanText)

		totalSize += elem.Size

		minX = math.Min(minX, elem.X)
		minY = math.Min(minY, elem.Y)
		maxX = math.Max(maxX, elem.X+elem.Width)
		maxY = math.Max(maxY, elem.Y+elem.Height)
	}

	avgSize := totalSize / float64(len(elements))

	// Detect font style
	isBold, isItalic := DetectFontStyle(elements[0].Font)

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
		Color:    layout.Color{R: 0, G: 0, B: 0}, // Default black
		IsBold:   isBold,
		IsItalic: isItalic,
	}
}

// CombineBlockText combines text in a block (preserving line breaks).
func CombineBlockText(lines [][]layout.TextElement) string {
	var result strings.Builder

	for i, line := range lines {
		if i > 0 {
			result.WriteString("\n") // Line break between lines
		}

		// Combine text within line (considering element distance)
		for j, elem := range line {
			if j > 0 {
				// Calculate distance from previous element
				prevElem := line[j-1]
				gap := elem.X - (prevElem.X + prevElem.Width)

				// Distance threshold: 35% of font size
				// Space between words while considering kerning
				threshold := prevElem.Size * 0.35

				if gap > threshold {
					result.WriteString(" ")
				}
			}
			// Clean control characters before adding
			cleanText := utils.CleanControlCharacters(elem.Text)
			result.WriteString(cleanText)
		}
	}

	return result.String()
}

// DetectFontStyle detects bold/italic from font name.
func DetectFontStyle(fontName string) (isBold, isItalic bool) {
	upper := strings.ToUpper(fontName)
	isBold = strings.Contains(upper, "BOLD") ||
		strings.Contains(upper, "-B") ||
		strings.HasSuffix(upper, "BD")
	isItalic = strings.Contains(upper, "ITALIC") ||
		strings.Contains(upper, "OBLIQUE") ||
		strings.Contains(upper, "-I") ||
		strings.Contains(upper, "-O")
	return
}
