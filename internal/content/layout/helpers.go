// Package layout provides internal layout processing utilities.
package layout

import (
	"github.com/ryomak/gopdf/internal/core"
	"github.com/ryomak/gopdf/layout"
)

// MinY returns the minimum Y coordinate from a slice of text elements.
func MinY(elements []layout.TextElement) float64 {
	if len(elements) == 0 {
		return 0
	}
	min := elements[0].Y
	for _, e := range elements[1:] {
		if e.Y < min {
			min = e.Y
		}
	}
	return min
}

// MaxY returns the maximum Y coordinate (including height) from a slice of text elements.
func MaxY(elements []layout.TextElement) float64 {
	if len(elements) == 0 {
		return 0
	}
	max := elements[0].Y + elements[0].Height
	for _, e := range elements[1:] {
		if e.Y+e.Height > max {
			max = e.Y + e.Height
		}
	}
	return max
}

// MinX returns the minimum X coordinate from a slice of text elements.
func MinX(elements []layout.TextElement) float64 {
	if len(elements) == 0 {
		return 0
	}
	min := elements[0].X
	for _, e := range elements[1:] {
		if e.X < min {
			min = e.X
		}
	}
	return min
}

// AvgFontSize returns the average font size from a slice of text elements.
func AvgFontSize(elements []layout.TextElement) float64 {
	if len(elements) == 0 {
		return 0
	}
	sum := 0.0
	for _, e := range elements {
		sum += e.Size
	}
	return sum / float64(len(elements))
}

// ToFloat64 converts a core.Object to float64.
// Supports core.Integer and core.Real types.
func ToFloat64(obj core.Object) float64 {
	switch v := obj.(type) {
	case core.Integer:
		return float64(v)
	case core.Real:
		return float64(v)
	default:
		return 0
	}
}
