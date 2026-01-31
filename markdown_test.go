package gopdf

import (
	"bytes"
	"testing"
)

func TestNewMarkdownDocument_Document(t *testing.T) {
	markdown := `# Hello World

This is a test document.

## Section 1

Some content here.
`

	doc, err := NewMarkdownDocument(markdown, &MarkdownOptions{
		Mode:     MarkdownModeDocument,
		PageSize: PageSizeA4,
	})
	if err != nil {
		t.Fatalf("failed to create markdown document: %v", err)
	}

	// Write to buffer to verify PDF generation
	var buf bytes.Buffer
	if err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("failed to write document: %v", err)
	}

	// Verify PDF header
	if !bytes.HasPrefix(buf.Bytes(), []byte("%PDF-")) {
		t.Error("expected PDF header")
	}
}

func TestNewMarkdownDocument_Slide(t *testing.T) {
	markdown := `# Slide 1

This is the first slide.

---

# Slide 2

This is the second slide.

- Point 1
- Point 2
- Point 3

---

# Slide 3

Final slide with content.
`

	doc, err := NewMarkdownDocument(markdown, &MarkdownOptions{
		Mode:     MarkdownModeSlide,
		PageSize: PageSizePresentation16x9,
	})
	if err != nil {
		t.Fatalf("failed to create slide document: %v", err)
	}

	// Write to buffer to verify PDF generation
	var buf bytes.Buffer
	if err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("failed to write document: %v", err)
	}

	// Verify PDF header
	if !bytes.HasPrefix(buf.Bytes(), []byte("%PDF-")) {
		t.Error("expected PDF header")
	}
}

func TestNewMarkdownDocument_SlideEmpty(t *testing.T) {
	// Empty markdown should create at least one slide
	doc, err := NewMarkdownDocument("", &MarkdownOptions{
		Mode:     MarkdownModeSlide,
		PageSize: PageSizePresentation16x9,
	})
	if err != nil {
		t.Fatalf("failed to create empty slide document: %v", err)
	}

	var buf bytes.Buffer
	if err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("failed to write document: %v", err)
	}

	if !bytes.HasPrefix(buf.Bytes(), []byte("%PDF-")) {
		t.Error("expected PDF header")
	}
}

func TestNewMarkdownDocument_SlideWithSubtitle(t *testing.T) {
	markdown := `# Main Title

## Subtitle

Content here.
`

	doc, err := NewMarkdownDocument(markdown, &MarkdownOptions{
		Mode:     MarkdownModeSlide,
		PageSize: PageSizePresentation4x3,
	})
	if err != nil {
		t.Fatalf("failed to create slide with subtitle: %v", err)
	}

	var buf bytes.Buffer
	if err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("failed to write document: %v", err)
	}

	if !bytes.HasPrefix(buf.Bytes(), []byte("%PDF-")) {
		t.Error("expected PDF header")
	}
}

func TestNewMarkdownDocument_DefaultOptions(t *testing.T) {
	markdown := `# Test`

	// nil options should use defaults
	doc, err := NewMarkdownDocument(markdown, nil)
	if err != nil {
		t.Fatalf("failed with nil options: %v", err)
	}

	var buf bytes.Buffer
	if err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("failed to write document: %v", err)
	}

	if !bytes.HasPrefix(buf.Bytes(), []byte("%PDF-")) {
		t.Error("expected PDF header")
	}
}

func TestDefaultMarkdownStyle(t *testing.T) {
	style := DefaultMarkdownStyle()

	if style.H1Size != 36 {
		t.Errorf("expected H1Size 36, got %v", style.H1Size)
	}
	if style.BodySize != 12 {
		t.Errorf("expected BodySize 12, got %v", style.BodySize)
	}
}

func TestDefaultSlideStyle(t *testing.T) {
	style := DefaultSlideStyle()

	if style.H1Size != 48 {
		t.Errorf("expected H1Size 48, got %v", style.H1Size)
	}
	if style.BodySize != 18 {
		t.Errorf("expected BodySize 18, got %v", style.BodySize)
	}
}

func TestNewMarkdownDocument_UnknownMode(t *testing.T) {
	markdown := `# Test`

	_, err := NewMarkdownDocument(markdown, &MarkdownOptions{
		Mode: "invalid",
	})
	if err == nil {
		t.Error("expected error for unknown mode")
	}
}
