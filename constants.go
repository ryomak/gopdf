// Package gopdf provides a high-level API for creating and manipulating PDF documents.
package gopdf

// PageSize represents standard PDF page sizes in points (1 point = 1/72 inch).
type PageSize struct {
	Width  float64
	Height float64
}

// Standard page sizes
var (
	// PageSizeA4 size: 210mm x 297mm
	PageSizeA4 = PageSize{Width: 595.0, Height: 842.0}

	// PageSizeLetter size: 8.5in x 11in
	PageSizeLetter = PageSize{Width: 612.0, Height: 792.0}

	// PageSizeLegal size: 8.5in x 14in
	PageSizeLegal = PageSize{Width: 612.0, Height: 1008.0}

	// PageSizeA3 size: 297mm x 420mm
	PageSizeA3 = PageSize{Width: 842.0, Height: 1191.0}

	// PageSizeA5 size: 148mm x 210mm
	PageSizeA5 = PageSize{Width: 420.0, Height: 595.0}
)

// Orientation represents page orientation.
type Orientation int

const (
	// Portrait orientation (vertical)
	Portrait Orientation = iota
	// Landscape orientation (horizontal)
	Landscape
)

// Apply applies the orientation to a page size.
func (o Orientation) Apply(size PageSize) PageSize {
	if o == Landscape {
		return PageSize{Width: size.Height, Height: size.Width}
	}
	return size
}
