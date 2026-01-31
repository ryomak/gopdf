package gopdf

import (
	"fmt"
	"strings"

	"github.com/gomarkdown/markdown/ast"
	"github.com/ryomak/gopdf/internal/content/markdown"
	"github.com/ryomak/gopdf/internal/content/text"
)

// documentRenderer renders Markdown to a PDF document.
type documentRenderer struct {
	doc          *Document
	currentPage  *Page
	style        *markdown.Style
	currentY     float64
	pageSize     PageSize
	orientation  Orientation
	imageBasePath string
}

// newDocumentRenderer creates a new document renderer.
func newDocumentRenderer(pageSize PageSize, orientation Orientation, style *markdown.Style, imageBasePath string) *documentRenderer {
	if style == nil {
		style = markdown.DefaultDocumentStyle()
	}

	return &documentRenderer{
		style:         style,
		pageSize:      pageSize,
		orientation:   orientation,
		imageBasePath: imageBasePath,
	}
}

// render renders the Markdown AST to a PDF document.
func (r *documentRenderer) render(root ast.Node) (*Document, error) {
	r.doc = New()
	r.newPage()

	// Walk the AST and render nodes
	if err := r.walkNode(root); err != nil {
		return nil, err
	}

	return r.doc, nil
}

// newPage creates a new page and resets the Y position.
func (r *documentRenderer) newPage() {
	r.currentPage = r.doc.AddPage(r.pageSize, r.orientation)
	r.currentY = r.currentPage.Height() - r.style.MarginTop
}

// checkPageBreak checks if we need a new page and creates one if necessary.
func (r *documentRenderer) checkPageBreak(requiredHeight float64) {
	if r.currentY-requiredHeight < r.style.MarginBottom {
		r.newPage()
	}
}

// walkNode walks the AST recursively and renders nodes.
func (r *documentRenderer) walkNode(node ast.Node) error {
	// Process current node
	if err := r.renderNode(node); err != nil {
		return err
	}

	// Process children
	for _, child := range node.GetChildren() {
		if err := r.walkNode(child); err != nil {
			return err
		}
	}

	return nil
}

// renderNode renders a single AST node.
func (r *documentRenderer) renderNode(node ast.Node) error {
	switch n := node.(type) {
	case *ast.Heading:
		return r.renderHeading(n)
	case *ast.Paragraph:
		return r.renderParagraph(n)
	case *ast.Text:
		return r.renderText(n)
	case *ast.Softbreak, *ast.Hardbreak:
		// Line breaks are handled by the parent node
		return nil
	case *ast.Document:
		// Document node itself doesn't need rendering
		return nil
	default:
		// For now, skip unsupported node types
		// In the future, we'll add support for lists, code blocks, etc.
		return nil
	}
}

// renderHeading renders a heading node.
func (r *documentRenderer) renderHeading(heading *ast.Heading) error {
	// Get heading level (1-6)
	level := heading.Level

	// Determine font size based on level
	var fontSize float64
	switch level {
	case 1:
		fontSize = r.style.H1Size
	case 2:
		fontSize = r.style.H2Size
	case 3:
		fontSize = r.style.H3Size
	case 4:
		fontSize = r.style.H4Size
	case 5:
		fontSize = r.style.H5Size
	case 6:
		fontSize = r.style.H6Size
	default:
		fontSize = r.style.BodySize
	}

	// Check for page break
	r.checkPageBreak(fontSize + r.style.ParagraphSpacing)

	// Set font and color
	if err := r.currentPage.SetFont(FontHelveticaBold, fontSize); err != nil {
		return fmt.Errorf("failed to set font: %w", err)
	}
	r.currentPage.SetFillColor(convertColor(r.style.HeadingColor))

	// Extract text from children
	text := r.extractText(heading)

	// Draw the heading
	err := r.currentPage.DrawText(text, r.style.MarginLeft, r.currentY)
	if err != nil {
		return fmt.Errorf("failed to draw heading: %w", err)
	}

	// Move Y position down
	r.currentY -= fontSize + r.style.ParagraphSpacing

	return nil
}

// renderParagraph renders a paragraph node with word wrapping.
func (r *documentRenderer) renderParagraph(para *ast.Paragraph) error {
	// Extract text from children
	paraText := r.extractText(para)

	if paraText == "" {
		return nil
	}

	// Calculate available width
	availableWidth := r.currentPage.Width() - r.style.MarginLeft - r.style.MarginRight

	// Wrap text to fit within the available width
	lines := text.Wrap(paraText, availableWidth, "Helvetica", r.style.BodySize, text.DefaultWidthEstimator)

	// Calculate total height needed
	lineHeight := r.style.BodySize * r.style.LineSpacing
	totalHeight := float64(len(lines))*lineHeight + r.style.ParagraphSpacing

	// Check for page break
	r.checkPageBreak(totalHeight)

	// Set font and color
	if err := r.currentPage.SetFont(FontHelvetica, r.style.BodySize); err != nil {
		return fmt.Errorf("failed to set font: %w", err)
	}
	r.currentPage.SetFillColor(convertColor(r.style.TextColor))

	// Draw each line
	for _, line := range lines {
		// Check if we need a new page for this line
		if r.currentY-lineHeight < r.style.MarginBottom {
			r.newPage()
			if err := r.currentPage.SetFont(FontHelvetica, r.style.BodySize); err != nil {
				return fmt.Errorf("failed to set font: %w", err)
			}
			r.currentPage.SetFillColor(convertColor(r.style.TextColor))
		}

		err := r.currentPage.DrawText(line, r.style.MarginLeft, r.currentY)
		if err != nil {
			return fmt.Errorf("failed to draw paragraph line: %w", err)
		}
		r.currentY -= lineHeight
	}

	// Add paragraph spacing
	r.currentY -= r.style.ParagraphSpacing

	return nil
}

// renderText renders a text node (usually handled by parent).
func (r *documentRenderer) renderText(text *ast.Text) error {
	// Text nodes are typically handled by their parent (paragraph, heading, etc.)
	// This is a no-op for now
	return nil
}

// extractText extracts all text content from a node and its children.
func (r *documentRenderer) extractText(node ast.Node) string {
	var text strings.Builder

	ast.WalkFunc(node, func(n ast.Node, entering bool) ast.WalkStatus {
		if !entering {
			return ast.GoToNext
		}

		switch t := n.(type) {
		case *ast.Text:
			text.Write(t.Literal)
		case *ast.Softbreak:
			text.WriteString(" ")
		case *ast.Hardbreak:
			text.WriteString("\n")
		}

		return ast.GoToNext
	})

	return text.String()
}

// convertColor converts internal markdown Color to gopdf Color.
func convertColor(c markdown.Color) Color {
	return Color{
		R: c.R,
		G: c.G,
		B: c.B,
	}
}

// slideRenderer renders Markdown to a presentation PDF (one slide per page).
type slideRenderer struct {
	doc           *Document
	currentPage   *Page
	style         *markdown.Style
	currentY      float64
	pageSize      PageSize
	orientation   Orientation
	imageBasePath string
}

// newSlideRenderer creates a new slide renderer.
func newSlideRenderer(pageSize PageSize, orientation Orientation, style *markdown.Style, imageBasePath string) *slideRenderer {
	if style == nil {
		style = markdown.DefaultSlideStyle()
	}

	return &slideRenderer{
		style:         style,
		pageSize:      pageSize,
		orientation:   orientation,
		imageBasePath: imageBasePath,
	}
}

// render renders the Markdown AST to a presentation PDF.
func (r *slideRenderer) render(root ast.Node) (*Document, error) {
	r.doc = New()

	// Split the AST into slides
	slides := r.splitIntoSlides(root)

	// Ensure at least one slide exists (create empty title slide if no content)
	if len(slides) == 0 {
		r.newPage()
		return r.doc, nil
	}

	// Render each slide
	for _, slide := range slides {
		if err := r.renderSlide(slide); err != nil {
			return nil, err
		}
	}

	return r.doc, nil
}

// slideContent represents content for a single slide.
type slideContent struct {
	title    string
	subtitle string
	items    []string
}

// splitIntoSlides splits the AST into slides based on horizontal rules or H1 headings.
func (r *slideRenderer) splitIntoSlides(root ast.Node) []slideContent {
	var slides []slideContent
	var currentSlide slideContent
	hasContent := false

	ast.WalkFunc(root, func(node ast.Node, entering bool) ast.WalkStatus {
		if !entering {
			return ast.GoToNext
		}

		switch n := node.(type) {
		case *ast.HorizontalRule:
			// Horizontal rule marks a slide break
			if hasContent {
				slides = append(slides, currentSlide)
				currentSlide = slideContent{}
				hasContent = false
			}
			return ast.SkipChildren

		case *ast.Heading:
			if n.Level == 1 {
				// H1 can also mark a new slide (if we already have content)
				if hasContent && currentSlide.title != "" {
					slides = append(slides, currentSlide)
					currentSlide = slideContent{}
				}
				currentSlide.title = r.extractTextFromNode(n)
				hasContent = true
			} else if n.Level == 2 {
				currentSlide.subtitle = r.extractTextFromNode(n)
				hasContent = true
			}
			return ast.SkipChildren

		case *ast.Paragraph:
			text := r.extractTextFromNode(n)
			if text != "" {
				currentSlide.items = append(currentSlide.items, text)
				hasContent = true
			}
			return ast.SkipChildren

		case *ast.List:
			// Extract list items
			for _, item := range n.GetChildren() {
				if listItem, ok := item.(*ast.ListItem); ok {
					text := r.extractTextFromNode(listItem)
					if text != "" {
						currentSlide.items = append(currentSlide.items, "• "+text)
						hasContent = true
					}
				}
			}
			return ast.SkipChildren

		case *ast.CodeBlock:
			// Add code block as a single item
			code := strings.TrimSpace(string(n.Literal))
			if code != "" {
				currentSlide.items = append(currentSlide.items, code)
				hasContent = true
			}
			return ast.SkipChildren
		}

		return ast.GoToNext
	})

	// Add the last slide if it has content
	if hasContent {
		slides = append(slides, currentSlide)
	}

	return slides
}

// extractTextFromNode extracts all text content from a node.
func (r *slideRenderer) extractTextFromNode(node ast.Node) string {
	var text strings.Builder

	ast.WalkFunc(node, func(n ast.Node, entering bool) ast.WalkStatus {
		if !entering {
			return ast.GoToNext
		}

		switch t := n.(type) {
		case *ast.Text:
			text.Write(t.Literal)
		case *ast.Code:
			text.Write(t.Literal)
		case *ast.Softbreak:
			text.WriteString(" ")
		case *ast.Hardbreak:
			text.WriteString(" ")
		}

		return ast.GoToNext
	})

	return strings.TrimSpace(text.String())
}

// newPage creates a new slide page.
func (r *slideRenderer) newPage() {
	r.currentPage = r.doc.AddPage(r.pageSize, r.orientation)
	r.currentY = r.currentPage.Height() - r.style.MarginTop
}

// renderSlide renders a single slide.
func (r *slideRenderer) renderSlide(slide slideContent) error {
	r.newPage()

	pageWidth := r.currentPage.Width()
	contentWidth := pageWidth - r.style.MarginLeft - r.style.MarginRight

	// Render title (centered)
	if slide.title != "" {
		if err := r.currentPage.SetFont(FontHelveticaBold, r.style.H1Size); err != nil {
			return fmt.Errorf("failed to set title font: %w", err)
		}
		r.currentPage.SetFillColor(convertColor(r.style.HeadingColor))

		// Calculate centered X position (approximate)
		titleWidth := float64(len(slide.title)) * r.style.H1Size * 0.5
		titleX := r.style.MarginLeft + (contentWidth-titleWidth)/2
		if titleX < r.style.MarginLeft {
			titleX = r.style.MarginLeft
		}

		if err := r.currentPage.DrawText(slide.title, titleX, r.currentY); err != nil {
			return fmt.Errorf("failed to draw title: %w", err)
		}
		r.currentY -= r.style.H1Size + r.style.ParagraphSpacing
	}

	// Render subtitle (centered)
	if slide.subtitle != "" {
		if err := r.currentPage.SetFont(FontHelvetica, r.style.H2Size); err != nil {
			return fmt.Errorf("failed to set subtitle font: %w", err)
		}
		r.currentPage.SetFillColor(convertColor(r.style.TextColor))

		subtitleWidth := float64(len(slide.subtitle)) * r.style.H2Size * 0.5
		subtitleX := r.style.MarginLeft + (contentWidth-subtitleWidth)/2
		if subtitleX < r.style.MarginLeft {
			subtitleX = r.style.MarginLeft
		}

		if err := r.currentPage.DrawText(slide.subtitle, subtitleX, r.currentY); err != nil {
			return fmt.Errorf("failed to draw subtitle: %w", err)
		}
		r.currentY -= r.style.H2Size + r.style.ParagraphSpacing*2
	}

	// Render body items (left-aligned)
	if len(slide.items) > 0 {
		if err := r.currentPage.SetFont(FontHelvetica, r.style.BodySize); err != nil {
			return fmt.Errorf("failed to set body font: %w", err)
		}
		r.currentPage.SetFillColor(convertColor(r.style.TextColor))

		for _, item := range slide.items {
			// Check if we have room for this item
			if r.currentY-r.style.BodySize < r.style.MarginBottom {
				break // Don't overflow the slide
			}

			if err := r.currentPage.DrawText(item, r.style.MarginLeft, r.currentY); err != nil {
				return fmt.Errorf("failed to draw item: %w", err)
			}
			r.currentY -= r.style.BodySize*r.style.LineSpacing + r.style.ParagraphSpacing/2
		}
	}

	return nil
}
