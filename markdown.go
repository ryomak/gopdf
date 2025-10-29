package gopdf

import (
	"fmt"
	"os"

	"github.com/ryomak/gopdf/internal/markdown"
)

// MarkdownMode represents the rendering mode for Markdown conversion.
type MarkdownMode string

const (
	// MarkdownModeDocument renders Markdown as a document (multi-page).
	MarkdownModeDocument MarkdownMode = "document"

	// MarkdownModeSlide renders Markdown as presentation slides.
	MarkdownModeSlide MarkdownMode = "slide"
)

// MarkdownOptions contains options for Markdown conversion.
type MarkdownOptions struct {
	// Mode: "document" or "slide"
	Mode MarkdownMode

	// PageSize: Page size for the PDF (default: A4 for document, Presentation16x9 for slide)
	PageSize PageSize

	// Orientation: Page orientation (default: Portrait)
	Orientation Orientation

	// Style: Custom style settings (optional, uses default if nil)
	Style *MarkdownStyle

	// ImageBasePath: Base path for resolving relative image paths
	ImageBasePath string
}

// MarkdownStyle represents styling configuration for Markdown rendering.
type MarkdownStyle struct {
	// H1-H6 font sizes
	H1Size, H2Size, H3Size, H4Size, H5Size, H6Size float64

	// Body text font size
	BodySize float64

	// Code block font size
	CodeSize float64

	// Line spacing (multiplier)
	LineSpacing float64

	// Paragraph spacing (points)
	ParagraphSpacing float64

	// Margins
	MarginTop, MarginRight, MarginBottom, MarginLeft float64

	// Colors
	TextColor      Color
	HeadingColor   Color
	CodeBackground Color
	LinkColor      Color

	// Font path for TTF fonts (optional)
	FontPath string
}

// NewMarkdownDocument creates a PDF document from Markdown text.
func NewMarkdownDocument(markdownText string, opts *MarkdownOptions) (*Document, error) {
	if opts == nil {
		opts = &MarkdownOptions{
			Mode:        MarkdownModeDocument,
			PageSize:    PageSizeA4,
			Orientation: Portrait,
		}
	}

	// Set default page size based on mode
	if opts.PageSize.Width == 0 {
		if opts.Mode == MarkdownModeSlide {
			opts.PageSize = PageSizePresentation16x9
		} else {
			opts.PageSize = PageSizeA4
		}
	}

	// Set default orientation
	if opts.Orientation == 0 {
		opts.Orientation = Portrait
	}

	// Convert public style to internal style
	var style *markdown.Style
	if opts.Style != nil {
		style = convertToInternalStyle(opts.Style)
	} else {
		if opts.Mode == MarkdownModeSlide {
			style = markdown.DefaultSlideStyle()
		} else {
			style = markdown.DefaultDocumentStyle()
		}
	}

	// Parse Markdown
	parser := markdown.NewParser()
	ast := parser.ParseString(markdownText)

	// Render based on mode
	var doc *Document
	var err error

	switch opts.Mode {
	case MarkdownModeDocument:
		renderer := newDocumentRenderer(opts.PageSize, opts.Orientation, style, opts.ImageBasePath)
		doc, err = renderer.render(ast)
	case MarkdownModeSlide:
		// TODO: Implement slide renderer
		return nil, fmt.Errorf("slide mode not yet implemented")
	default:
		return nil, fmt.Errorf("unknown markdown mode: %s", opts.Mode)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to render markdown: %w", err)
	}

	return doc, nil
}

// NewMarkdownDocumentFromFile creates a PDF document from a Markdown file.
func NewMarkdownDocumentFromFile(filepath string, opts *MarkdownOptions) (*Document, error) {
	// Read the Markdown file
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read markdown file: %w", err)
	}

	// Set default image base path to the directory of the markdown file
	if opts != nil && opts.ImageBasePath == "" {
		opts.ImageBasePath = filepath[:len(filepath)-len(filepath[len(filepath)-1:])]
	}

	return NewMarkdownDocument(string(data), opts)
}

// DefaultMarkdownStyle returns the default style for document rendering.
func DefaultMarkdownStyle() *MarkdownStyle {
	return &MarkdownStyle{
		H1Size:           36,
		H2Size:           28,
		H3Size:           22,
		H4Size:           18,
		H5Size:           14,
		H6Size:           12,
		BodySize:         12,
		CodeSize:         10,
		LineSpacing:      1.2,
		ParagraphSpacing: 12,
		MarginTop:        72,
		MarginRight:      72,
		MarginBottom:     72,
		MarginLeft:       72,
		TextColor:        ColorBlack,
		HeadingColor:     ColorBlack,
		CodeBackground:   Color{R: 0.95, G: 0.95, B: 0.95},
		LinkColor:        ColorBlue,
	}
}

// DefaultSlideStyle returns the default style for slide rendering.
func DefaultSlideStyle() *MarkdownStyle {
	return &MarkdownStyle{
		H1Size:           48,
		H2Size:           36,
		H3Size:           28,
		H4Size:           24,
		H5Size:           20,
		H6Size:           18,
		BodySize:         18,
		CodeSize:         14,
		LineSpacing:      1.3,
		ParagraphSpacing: 18,
		MarginTop:        50,
		MarginRight:      50,
		MarginBottom:     50,
		MarginLeft:       50,
		TextColor:        ColorBlack,
		HeadingColor:     Color{R: 0.2, G: 0.2, B: 0.6},
		CodeBackground:   Color{R: 0.95, G: 0.95, B: 0.95},
		LinkColor:        ColorBlue,
	}
}

// convertToInternalStyle converts public MarkdownStyle to internal markdown.Style.
func convertToInternalStyle(s *MarkdownStyle) *markdown.Style {
	return &markdown.Style{
		H1Size:           s.H1Size,
		H2Size:           s.H2Size,
		H3Size:           s.H3Size,
		H4Size:           s.H4Size,
		H5Size:           s.H5Size,
		H6Size:           s.H6Size,
		BodySize:         s.BodySize,
		CodeSize:         s.CodeSize,
		LineSpacing:      s.LineSpacing,
		ParagraphSpacing: s.ParagraphSpacing,
		MarginTop:        s.MarginTop,
		MarginRight:      s.MarginRight,
		MarginBottom:     s.MarginBottom,
		MarginLeft:       s.MarginLeft,
		TextColor:        markdown.Color{R: s.TextColor.R, G: s.TextColor.G, B: s.TextColor.B},
		HeadingColor:     markdown.Color{R: s.HeadingColor.R, G: s.HeadingColor.G, B: s.HeadingColor.B},
		CodeBackground:   markdown.Color{R: s.CodeBackground.R, G: s.CodeBackground.G, B: s.CodeBackground.B},
		LinkColor:        markdown.Color{R: s.LinkColor.R, G: s.LinkColor.G, B: s.LinkColor.B},
		FontPath:         s.FontPath,
	}
}
