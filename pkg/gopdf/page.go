package gopdf

// Page represents a single page in a PDF document.
type Page struct {
	width  float64
	height float64
}

// Width returns the page width in points.
func (p *Page) Width() float64 {
	return p.width
}

// Height returns the page height in points.
func (p *Page) Height() float64 {
	return p.height
}
