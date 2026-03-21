package gopdf

import (
	"bytes"
	"strings"
	"testing"
)

// createTestPDFWithText creates a PDF with the given texts on a single page
func createTestPDFWithText(t *testing.T, texts []struct{ text string; x, y float64 }) *bytes.Buffer {
	t.Helper()
	doc := New()
	page := doc.AddPage(PageSizeA4, Portrait)

	if err := page.SetFont(FontHelvetica, 12); err != nil {
		t.Fatalf("SetFont failed: %v", err)
	}

	for _, item := range texts {
		if err := page.DrawText(item.text, item.x, item.y); err != nil {
			t.Fatalf("DrawText failed: %v", err)
		}
	}

	var buf bytes.Buffer
	if err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo failed: %v", err)
	}
	return &buf
}

func TestExtractPageText_SingleLine(t *testing.T) {
	buf := createTestPDFWithText(t, []struct{ text string; x, y float64 }{
		{"HelloWorld", 100, 700},
	})

	reader, err := OpenReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("OpenReader failed: %v", err)
	}
	defer reader.Close()

	text, err := reader.ExtractPageText(0)
	if err != nil {
		t.Fatalf("ExtractPageText failed: %v", err)
	}

	if !strings.Contains(text, "HelloWorld") {
		t.Errorf("expected text to contain 'HelloWorld', got %q", text)
	}
}

func TestExtractPageText_MultipleLines(t *testing.T) {
	buf := createTestPDFWithText(t, []struct{ text string; x, y float64 }{
		{"FirstLine", 100, 700},
		{"SecondLine", 100, 680},
		{"ThirdLine", 100, 660},
	})

	reader, err := OpenReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	text, err := reader.ExtractPageText(0)
	if err != nil {
		t.Fatalf("ExtractPageText failed: %v", err)
	}

	for _, expected := range []string{"FirstLine", "SecondLine", "ThirdLine"} {
		if !strings.Contains(text, expected) {
			t.Errorf("expected text to contain %q, got %q", expected, text)
		}
	}
}

func TestExtractPageText_InvalidPage(t *testing.T) {
	doc := New()
	doc.AddPage(PageSizeA4, Portrait)

	var buf bytes.Buffer
	if err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	reader, err := OpenReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	_, err = reader.ExtractPageText(99)
	if err == nil {
		t.Error("expected error for out-of-range page")
	}
}

func TestExtractText_AllPages(t *testing.T) {
	doc := New()

	// Page 1
	page1 := doc.AddPage(PageSizeA4, Portrait)
	if err := page1.SetFont(FontHelvetica, 12); err != nil {
		t.Fatal(err)
	}
	_ = page1.DrawText("PageOneContent", 100, 700)

	// Page 2
	page2 := doc.AddPage(PageSizeA4, Portrait)
	if err := page2.SetFont(FontHelvetica, 12); err != nil {
		t.Fatal(err)
	}
	_ = page2.DrawText("PageTwoContent", 100, 700)

	var buf bytes.Buffer
	if err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	reader, err := OpenReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	text, err := reader.ExtractText()
	if err != nil {
		t.Fatalf("ExtractText failed: %v", err)
	}

	if !strings.Contains(text, "PageOneContent") {
		t.Errorf("expected text to contain 'PageOneContent', got %q", text)
	}
	if !strings.Contains(text, "PageTwoContent") {
		t.Errorf("expected text to contain 'PageTwoContent', got %q", text)
	}
}

func TestExtractPageTextElements(t *testing.T) {
	buf := createTestPDFWithText(t, []struct{ text string; x, y float64 }{
		{"Element1", 100, 700},
		{"Element2", 100, 680},
	})

	reader, err := OpenReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	elements, err := reader.ExtractPageTextElements(0)
	if err != nil {
		t.Fatalf("ExtractPageTextElements failed: %v", err)
	}

	if len(elements) == 0 {
		t.Fatal("expected at least one text element")
	}

	// Verify elements have position data
	for _, elem := range elements {
		if elem.Text == "" {
			continue
		}
		if elem.Size <= 0 {
			t.Errorf("element %q has invalid size: %f", elem.Text, elem.Size)
		}
	}
}

func TestExtractAllTextElements(t *testing.T) {
	doc := New()

	page1 := doc.AddPage(PageSizeA4, Portrait)
	_ = page1.SetFont(FontHelvetica, 12)
	_ = page1.DrawText("P1Text", 100, 700)

	page2 := doc.AddPage(PageSizeA4, Portrait)
	_ = page2.SetFont(FontHelvetica, 12)
	_ = page2.DrawText("P2Text", 100, 700)

	var buf bytes.Buffer
	if err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	reader, err := OpenReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	allElements, err := reader.ExtractAllTextElements()
	if err != nil {
		t.Fatalf("ExtractAllTextElements failed: %v", err)
	}

	if len(allElements) != 2 {
		t.Errorf("expected elements for 2 pages, got %d", len(allElements))
	}
}

func TestExtractPageTextBlocks(t *testing.T) {
	buf := createTestPDFWithText(t, []struct{ text string; x, y float64 }{
		{"Block1Line1", 100, 700},
		{"Block1Line2", 100, 685},
	})

	reader, err := OpenReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	blocks, err := reader.ExtractPageTextBlocks(0)
	if err != nil {
		t.Fatalf("ExtractPageTextBlocks failed: %v", err)
	}

	if len(blocks) == 0 {
		t.Error("expected at least one text block")
	}
}

func TestExtractAllTextBlocks(t *testing.T) {
	doc := New()
	page := doc.AddPage(PageSizeA4, Portrait)
	_ = page.SetFont(FontHelvetica, 12)
	_ = page.DrawText("Text", 100, 700)

	var buf bytes.Buffer
	if err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	reader, err := OpenReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	allBlocks, err := reader.ExtractAllTextBlocks()
	if err != nil {
		t.Fatalf("ExtractAllTextBlocks failed: %v", err)
	}

	if len(allBlocks) == 0 {
		t.Error("expected at least one page of blocks")
	}
}

func TestExtractPageContentBlocks(t *testing.T) {
	doc := New()
	page := doc.AddPage(PageSizeA4, Portrait)
	_ = page.SetFont(FontHelvetica, 12)
	_ = page.DrawText("Content", 100, 700)

	var buf bytes.Buffer
	if err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	reader, err := OpenReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	blocks, err := reader.ExtractPageContentBlocks(0)
	if err != nil {
		t.Fatalf("ExtractPageContentBlocks failed: %v", err)
	}

	if len(blocks) == 0 {
		t.Error("expected at least one content block")
	}

	for _, block := range blocks {
		if block.Type() != ContentBlockTypeText && block.Type() != ContentBlockTypeImage {
			t.Errorf("unexpected block type: %v", block.Type())
		}
	}
}

func TestExtractAllContentBlocks(t *testing.T) {
	doc := New()
	page := doc.AddPage(PageSizeA4, Portrait)
	_ = page.SetFont(FontHelvetica, 12)
	_ = page.DrawText("AllContent", 100, 700)

	var buf bytes.Buffer
	if err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	reader, err := OpenReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	allBlocks, err := reader.ExtractAllContentBlocks()
	if err != nil {
		t.Fatalf("ExtractAllContentBlocks failed: %v", err)
	}

	if len(allBlocks) == 0 {
		t.Error("expected at least one page")
	}
}

func TestExtractAllContentBlocksFlattened(t *testing.T) {
	tests := []struct {
		name              string
		mergeAcrossPages  bool
	}{
		{"without merging", false},
		{"with merging", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := New()

			page1 := doc.AddPage(PageSizeA4, Portrait)
			_ = page1.SetFont(FontHelvetica, 12)
			_ = page1.DrawText("Page1", 100, 700)

			page2 := doc.AddPage(PageSizeA4, Portrait)
			_ = page2.SetFont(FontHelvetica, 12)
			_ = page2.DrawText("Page2", 100, 700)

			var buf bytes.Buffer
			if err := doc.WriteTo(&buf); err != nil {
				t.Fatal(err)
			}

			reader, err := OpenReader(bytes.NewReader(buf.Bytes()))
			if err != nil {
				t.Fatal(err)
			}
			defer reader.Close()

			blocks, err := reader.ExtractAllContentBlocksFlattened(tt.mergeAcrossPages)
			if err != nil {
				t.Fatalf("ExtractAllContentBlocksFlattened failed: %v", err)
			}

			if len(blocks) == 0 {
				t.Error("expected at least one block")
			}
		})
	}
}

func TestEstimateTextWidth_TableDriven(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		fontSize float64
		font     string
	}{
		{"empty text", "", 12, "Helvetica"},
		{"hello", "Hello", 12, "Helvetica"},
		{"large font", "Test", 24, "Times-Roman"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			width := estimateTextWidth(tt.text, tt.fontSize, tt.font)
			expectedWidth := float64(len(tt.text)) * tt.fontSize * 0.6
			if width != expectedWidth {
				t.Errorf("estimateTextWidth(%q, %f, %q) = %f, want %f",
					tt.text, tt.fontSize, tt.font, width, expectedWidth)
			}
		})
	}
}

func TestExtractPageText_EmptyPage(t *testing.T) {
	doc := New()
	doc.AddPage(PageSizeA4, Portrait) // Empty page, no text

	var buf bytes.Buffer
	if err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	reader, err := OpenReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	text, err := reader.ExtractPageText(0)
	if err != nil {
		t.Fatalf("ExtractPageText on empty page failed: %v", err)
	}

	if text != "" {
		t.Errorf("expected empty text for empty page, got %q", text)
	}
}

func TestExtractPageLayout_Roundtrip(t *testing.T) {
	doc := New()
	page := doc.AddPage(PageSizeA4, Portrait)
	_ = page.SetFont(FontHelvetica, 14)
	_ = page.DrawText("LayoutTest", 50, 600)

	var buf bytes.Buffer
	if err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	reader, err := OpenReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	layout, err := reader.ExtractPageLayout(0)
	if err != nil {
		t.Fatalf("ExtractPageLayout failed: %v", err)
	}

	if layout == nil {
		t.Fatal("layout should not be nil")
	}
	if layout.Width <= 0 || layout.Height <= 0 {
		t.Errorf("invalid page size: %f x %f", layout.Width, layout.Height)
	}
}

func TestExtractAllLayouts_MultiSize(t *testing.T) {
	doc := New()

	page1 := doc.AddPage(PageSizeA4, Portrait)
	_ = page1.SetFont(FontHelvetica, 12)
	_ = page1.DrawText("Layout1", 100, 700)

	page2 := doc.AddPage(PageSizeLetter, Portrait)
	_ = page2.SetFont(FontHelvetica, 12)
	_ = page2.DrawText("Layout2", 100, 700)

	var buf bytes.Buffer
	if err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	reader, err := OpenReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	layouts, err := reader.ExtractAllLayouts()
	if err != nil {
		t.Fatalf("ExtractAllLayouts failed: %v", err)
	}

	if len(layouts) != 2 {
		t.Errorf("expected 2 layouts, got %d", len(layouts))
	}
}

func TestPDFReader_IsEncrypted(t *testing.T) {
	// Unencrypted PDF
	doc := New()
	doc.AddPage(PageSizeA4, Portrait)

	var buf bytes.Buffer
	if err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	reader, err := OpenReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	if reader.IsEncrypted() {
		t.Error("unencrypted PDF should not be marked as encrypted")
	}
}

func TestExtractAllImages_NoImages(t *testing.T) {
	doc := New()
	page := doc.AddPage(PageSizeA4, Portrait)
	_ = page.SetFont(FontHelvetica, 12)
	_ = page.DrawText("TextOnly", 100, 700)

	var buf bytes.Buffer
	if err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	reader, err := OpenReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	images, err := reader.ExtractAllImages()
	if err != nil {
		t.Fatalf("ExtractAllImages failed: %v", err)
	}

	// No images in text-only PDF
	if len(images) != 0 {
		t.Errorf("expected 0 pages with images, got %d", len(images))
	}
}

func TestPDFReader_GetEncryptionInfo_Unencrypted(t *testing.T) {
	doc := New()
	doc.AddPage(PageSizeA4, Portrait)

	var buf bytes.Buffer
	if err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	reader, err := OpenReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	info := reader.GetEncryptionInfo()
	if info != nil {
		t.Error("expected nil encryption info for unencrypted PDF")
	}
}
