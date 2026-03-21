package content

import (
	"bytes"
	"testing"

	"github.com/ryomak/gopdf/internal/core"
	"github.com/ryomak/gopdf/internal/reader"
)

func TestNewImageExtractor(t *testing.T) {
	pdf := createMinimalPDF()
	r, err := reader.NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	ie := NewImageExtractor(r)
	if ie == nil {
		t.Fatal("NewImageExtractor returned nil")
	}
	if ie.reader != r {
		t.Error("ImageExtractor reader not set correctly")
	}
}

func TestImageExtractor_ExtractImages_NoXObject(t *testing.T) {
	pdf := createMinimalPDF()
	r, err := reader.NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	ie := NewImageExtractor(r)

	// Get page 1 from our minimal PDF
	page, err := r.GetPage(0)
	if err != nil {
		t.Fatalf("GetPage failed: %v", err)
	}

	// Our minimal PDF has no XObject, so ExtractImages should return nil
	images, err := ie.ExtractImages(page)
	if err != nil {
		t.Fatalf("ExtractImages failed: %v", err)
	}

	if len(images) != 0 {
		t.Errorf("Expected 0 images, got %d", len(images))
	}
}

func TestImageExtractor_ExtractImagesWithPosition_NoXObject(t *testing.T) {
	pdf := createMinimalPDF()
	r, err := reader.NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	ie := NewImageExtractor(r)

	page, err := r.GetPage(0)
	if err != nil {
		t.Fatalf("GetPage failed: %v", err)
	}

	operations := []Operation{
		{Operator: "q"},
		{Operator: "cm", Operands: []core.Object{
			core.Real(1), core.Real(0), core.Real(0), core.Real(1), core.Real(0), core.Real(0),
		}},
		{Operator: "Q"},
	}

	images, err := ie.ExtractImagesWithPosition(page, operations, nil)
	if err != nil {
		t.Fatalf("ExtractImagesWithPosition failed: %v", err)
	}

	if len(images) != 0 {
		t.Errorf("Expected 0 images, got %d", len(images))
	}
}
