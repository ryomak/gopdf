package translate

// ValidateImagePosition validates image coordinates and returns fallback positions if abnormal.
func ValidateImagePosition(x, y, width, height, pageWidth, pageHeight float64) (newX, newY float64) {
	const maxOffset = 10000.0

	// Detect abnormal values
	if x < -maxOffset || x > pageWidth+maxOffset ||
		y < -maxOffset || y > pageHeight+maxOffset {
		// Fallback: place at center-top of page
		newX = (pageWidth - width) / 2
		if newX < 0 {
			newX = 0
		}
		newY = pageHeight - height - 50
		if newY < 0 {
			newY = 0
		}
		return newX, newY
	}

	return x, y
}
