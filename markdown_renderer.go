package gopdf

import (
	"fmt"
	"strings"

	"github.com/gomarkdown/markdown/ast"
	"github.com/ryomak/gopdf/internal/markdown"
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

// renderParagraph renders a paragraph node.
func (r *documentRenderer) renderParagraph(para *ast.Paragraph) error {
	// Extract text from children
	text := r.extractText(para)

	if text == "" {
		return nil
	}

	// Check for page break
	estimatedHeight := r.style.BodySize * r.style.LineSpacing * 3 // Estimate 3 lines
	r.checkPageBreak(estimatedHeight)

	// Set font and color
	if err := r.currentPage.SetFont(FontHelvetica, r.style.BodySize); err != nil {
		return fmt.Errorf("failed to set font: %w", err)
	}
	r.currentPage.SetFillColor(convertColor(r.style.TextColor))

	// For now, draw as a single line
	// TODO: Implement word wrapping for long paragraphs
	err := r.currentPage.DrawText(text, r.style.MarginLeft, r.currentY)
	if err != nil {
		return fmt.Errorf("failed to draw paragraph: %w", err)
	}

	// Move Y position down
	r.currentY -= r.style.BodySize * r.style.LineSpacing + r.style.ParagraphSpacing

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
