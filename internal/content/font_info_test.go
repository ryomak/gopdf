package content

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/ryomak/gopdf/internal/core"
	"github.com/ryomak/gopdf/internal/reader"
)

// createMinimalPDF creates a minimal valid PDF for testing
func createMinimalPDF() []byte {
	var buf bytes.Buffer

	header := "%PDF-1.7\n\n"
	buf.WriteString(header)

	offsets := make([]int, 6)

	// Object 1: Catalog
	offsets[1] = buf.Len()
	buf.WriteString("1 0 obj\n")
	buf.WriteString("<< /Type /Catalog /Pages 2 0 R >>\n")
	buf.WriteString("endobj\n\n")

	// Object 2: Pages
	offsets[2] = buf.Len()
	buf.WriteString("2 0 obj\n")
	buf.WriteString("<< /Type /Pages /Kids [3 0 R] /Count 1 >>\n")
	buf.WriteString("endobj\n\n")

	// Object 3: Page
	offsets[3] = buf.Len()
	buf.WriteString("3 0 obj\n")
	buf.WriteString("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >>\n")
	buf.WriteString("endobj\n\n")

	// Object 4: Contents
	offsets[4] = buf.Len()
	streamContent := "BT\n/F1 12 Tf\n100 700 Td\n(Hello) Tj\nET\n"
	buf.WriteString("4 0 obj\n")
	buf.WriteString(fmt.Sprintf("<< /Length %d >>\n", len(streamContent)))
	buf.WriteString("stream\n")
	buf.WriteString(streamContent)
	buf.WriteString("endstream\n")
	buf.WriteString("endobj\n\n")

	// Object 5: Font
	offsets[5] = buf.Len()
	buf.WriteString("5 0 obj\n")
	buf.WriteString("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\n")
	buf.WriteString("endobj\n\n")

	xrefStart := buf.Len()

	buf.WriteString("xref\n")
	buf.WriteString("0 6\n")
	buf.WriteString("0000000000 65535 f \n")
	for i := 1; i <= 5; i++ {
		buf.WriteString(fmt.Sprintf("%010d 00000 n \n", offsets[i]))
	}

	buf.WriteString("trailer\n")
	buf.WriteString("<< /Size 6 /Root 1 0 R >>\n")
	buf.WriteString("startxref\n")
	buf.WriteString(fmt.Sprintf("%d\n", xrefStart))
	buf.WriteString("%%EOF")

	return buf.Bytes()
}

func TestNewFontManager(t *testing.T) {
	pdf := createMinimalPDF()
	r, err := reader.NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	fm := NewFontManager(r)
	if fm == nil {
		t.Fatal("NewFontManager returned nil")
	}
	if fm.reader != r {
		t.Error("FontManager reader not set correctly")
	}
	if fm.fonts == nil {
		t.Error("FontManager fonts map not initialized")
	}
}

func TestFontManager_GetFont(t *testing.T) {
	pdf := createMinimalPDF()
	r, err := reader.NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	fm := NewFontManager(r)

	pageResources2 := core.Dictionary{
		core.Name("Font"): core.Dictionary{
			core.Name("F1"): core.Dictionary{
				core.Name("Type"):     core.Name("Font"),
				core.Name("Subtype"):  core.Name("Type1"),
				core.Name("BaseFont"): core.Name("Helvetica"),
			},
		},
	}

	info, err := fm.GetFont("F1", pageResources2)
	if err != nil {
		t.Fatalf("GetFont failed: %v", err)
	}

	if info == nil {
		t.Fatal("GetFont returned nil")
	}

	if info.BaseFont != "Helvetica" {
		t.Errorf("BaseFont = %q, want %q", info.BaseFont, "Helvetica")
	}

	// Second call should use cache
	info2, err := fm.GetFont("F1", pageResources2)
	if err != nil {
		t.Fatalf("GetFont (cached) failed: %v", err)
	}
	if info2 != info {
		t.Error("Expected cached FontInfo to be same pointer")
	}

	// Non-existent font - getFontDictionary will fail, but loadFontInfo returns info without error
	info3, err := fm.GetFont("F99", pageResources2)
	if err != nil {
		t.Fatalf("GetFont for missing font should not error: %v", err)
	}
	if info3.Name != "F99" {
		t.Errorf("Name = %q, want %q", info3.Name, "F99")
	}

	// Test with nil resources
	info4, err := fm.GetFont("F2", nil)
	if err != nil {
		t.Fatalf("GetFont with nil resources should not error: %v", err)
	}
	if info4.Name != "F2" {
		t.Errorf("Name = %q, want %q", info4.Name, "F2")
	}

	// Test with resources having Font as non-Dictionary
	badResources := core.Dictionary{
		core.Name("Font"): core.String("not a dict"),
	}
	info5, err := fm.GetFont("F3", badResources)
	if err != nil {
		t.Fatalf("GetFont with bad Font resources should not error: %v", err)
	}
	if info5.Name != "F3" {
		t.Errorf("Name = %q, want %q", info5.Name, "F3")
	}

	// Test with resources having no Font key
	noFontResources := core.Dictionary{
		core.Name("XObject"): core.Dictionary{},
	}
	info6, err := fm.GetFont("F4", noFontResources)
	if err != nil {
		t.Fatalf("GetFont with no Font key should not error: %v", err)
	}
	if info6.Name != "F4" {
		t.Errorf("Name = %q, want %q", info6.Name, "F4")
	}

	// Test font with non-Dictionary font object
	badFontResources := core.Dictionary{
		core.Name("Font"): core.Dictionary{
			core.Name("F5"): core.String("not a font dict"),
		},
	}
	info7, err := fm.GetFont("F5", badFontResources)
	if err != nil {
		t.Fatalf("GetFont with non-dict font should not error: %v", err)
	}
	if info7.Name != "F5" {
		t.Errorf("Name = %q, want %q", info7.Name, "F5")
	}
}

func TestFontManager_GetFont_WithToUnicode(t *testing.T) {
	pdf := createMinimalPDF()
	r, err := reader.NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	fm := NewFontManager(r)

	// Font dict with ToUnicode that's not a stream (will fail extraction gracefully)
	pageResources := core.Dictionary{
		core.Name("Font"): core.Dictionary{
			core.Name("F1"): core.Dictionary{
				core.Name("Type"):      core.Name("Font"),
				core.Name("BaseFont"):  core.Name("TestFont"),
				core.Name("ToUnicode"): core.String("not a stream"),
			},
		},
	}

	info, err := fm.GetFont("F1", pageResources)
	if err != nil {
		t.Fatalf("GetFont failed: %v", err)
	}

	// BaseFont should be set even though ToUnicode extraction failed
	if info.BaseFont != "TestFont" {
		t.Errorf("BaseFont = %q, want %q", info.BaseFont, "TestFont")
	}

	// ToUnicode should be nil since extraction failed
	if info.ToUnicodeCMap != nil {
		t.Error("ToUnicodeCMap should be nil when ToUnicode is not a stream")
	}
}

func TestNewTextExtractor_WithReader(t *testing.T) {
	pdf := createMinimalPDF()
	r, err := reader.NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	operations := []Operation{
		{Operator: "BT"},
		{Operator: "ET"},
	}

	extractor := NewTextExtractor(operations, r, nil)
	if extractor.fontManager == nil {
		t.Error("fontManager should be set when reader is provided")
	}
}
