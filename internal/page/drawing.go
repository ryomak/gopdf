// Package page provides internal page drawing and encoding utilities.
package page

import (
	"bytes"
	"fmt"
)

// DrawTextInternal draws text at the specified position using PDF operators.
// This function writes to the content buffer using the given font key and encoded text.
// If useBrackets is true, the text is wrapped in parentheses; otherwise in angle brackets (hex).
func DrawTextInternal(
	content *bytes.Buffer,
	x, y float64,
	fontKey string,
	fontSize float64,
	encodedText string,
	useBrackets bool,
) {
	fmt.Fprintf(content, "BT\n")
	fmt.Fprintf(content, "/%s %.2f Tf\n", fontKey, fontSize)
	fmt.Fprintf(content, "%.2f %.2f Td\n", x, y)

	if useBrackets {
		fmt.Fprintf(content, "(%s) Tj\n", encodedText)
	} else {
		fmt.Fprintf(content, "<%s> Tj\n", encodedText)
	}

	fmt.Fprintf(content, "ET\n")
}

// DrawCirclePath draws a circle path using 4 Bézier curves.
// κ = 4 * (√2 - 1) / 3 ≈ 0.5522847498
func DrawCirclePath(content *bytes.Buffer, centerX, centerY, radius float64) {
	// Magic constant for circle approximation using Bézier curves
	const kappa = 0.5522847498

	// Calculate control point offset
	offset := radius * kappa

	// Calculate key points on the circle
	x0 := centerX + radius // Right
	y0 := centerY
	x1 := centerX          // Left
	y1 := centerY
	x2 := centerX          // Center X
	y2 := centerY + radius // Top
	x3 := centerX          // Center X
	y3 := centerY - radius // Bottom

	// Start at the right point (3 o'clock position)
	fmt.Fprintf(content, "%.2f %.2f m\n", x0, y0)

	// Draw 4 Bézier curves to approximate a circle
	// Curve 1: Right to Top (3 o'clock to 12 o'clock)
	fmt.Fprintf(content, "%.2f %.2f %.2f %.2f %.2f %.2f c\n",
		x0, y0+offset, // Control point 1
		x2+offset, y2, // Control point 2
		x2, y2) // End point

	// Curve 2: Top to Left (12 o'clock to 9 o'clock)
	fmt.Fprintf(content, "%.2f %.2f %.2f %.2f %.2f %.2f c\n",
		x2-offset, y2, // Control point 1
		x1, y1+offset, // Control point 2
		x1, y1) // End point

	// Curve 3: Left to Bottom (9 o'clock to 6 o'clock)
	fmt.Fprintf(content, "%.2f %.2f %.2f %.2f %.2f %.2f c\n",
		x1, y1-offset, // Control point 1
		x3-offset, y3, // Control point 2
		x3, y3) // End point

	// Curve 4: Bottom to Right (6 o'clock to 3 o'clock)
	fmt.Fprintf(content, "%.2f %.2f %.2f %.2f %.2f %.2f c\n",
		x3+offset, y3, // Control point 1
		x0, y0-offset, // Control point 2
		x0, y0) // End point
}
