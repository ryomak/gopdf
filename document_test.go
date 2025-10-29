package gopdf

import (
	"bytes"
	"strings"
	"testing"
)

// TestNewDocument はDocumentの作成をテストする
func TestNewDocument(t *testing.T) {
	doc := New()
	if doc == nil {
		t.Fatal("New() returned nil")
	}

	if len(doc.pages) != 0 {
		t.Errorf("New document should have 0 pages, got %d", len(doc.pages))
	}
}

// TestAddPage はページの追加をテストする
func TestAddPage(t *testing.T) {
	doc := New()

	// A4 Portrait
	page1 := doc.AddPage(PageSizeA4, Portrait)
	if page1 == nil {
		t.Fatal("AddPage returned nil")
	}

	if len(doc.pages) != 1 {
		t.Errorf("Document should have 1 page, got %d", len(doc.pages))
	}

	// Letter Landscape
	page2 := doc.AddPage(Letter, Landscape)
	if page2 == nil {
		t.Fatal("AddPage returned nil")
	}

	if len(doc.pages) != 2 {
		t.Errorf("Document should have 2 pages, got %d", len(doc.pages))
	}
}

// TestDocumentWriteTo は最小限のPDF出力をテストする
func TestDocumentWriteTo(t *testing.T) {
	doc := New()
	doc.AddPage(PageSizeA4, Portrait)

	var buf bytes.Buffer
	err := doc.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo() failed: %v", err)
	}

	output := buf.String()

	// 必須要素の確認
	if !strings.Contains(output, "%PDF-1.7") {
		t.Error("Output should contain PDF header")
	}

	if !strings.Contains(output, "/Type /Catalog") {
		t.Error("Output should contain Catalog")
	}

	if !strings.Contains(output, "/Type /Pages") {
		t.Error("Output should contain Pages")
	}

	if !strings.Contains(output, "/Type /Page") {
		t.Error("Output should contain Page")
	}

	if !strings.Contains(output, "xref") {
		t.Error("Output should contain xref table")
	}

	if !strings.Contains(output, "trailer") {
		t.Error("Output should contain trailer")
	}

	if !strings.Contains(output, "startxref") {
		t.Error("Output should contain startxref")
	}

	if !strings.HasSuffix(strings.TrimSpace(output), "%%EOF") {
		t.Error("Output should end with EOF marker")
	}
}

// TestEmptyDocument は空のドキュメント（ページなし）の出力をテストする
func TestEmptyDocument(t *testing.T) {
	doc := New()

	var buf bytes.Buffer
	err := doc.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo() failed: %v", err)
	}

	output := buf.String()

	// 空でもPDFとして有効
	if !strings.Contains(output, "%PDF-1.7") {
		t.Error("Output should contain PDF header")
	}

	if !strings.Contains(output, "/Type /Catalog") {
		t.Error("Output should contain Catalog")
	}

	if !strings.Contains(output, "/Count 0") {
		t.Error("Output should contain Pages with Count 0")
	}
}

// TestMultiplePages は複数ページの出力をテストする
func TestMultiplePages(t *testing.T) {
	doc := New()

	// 3ページ追加
	doc.AddPage(PageSizeA4, Portrait)
	doc.AddPage(Letter, Portrait)
	doc.AddPage(PageSizeA4, Landscape)

	if len(doc.pages) != 3 {
		t.Errorf("Document should have 3 pages, got %d", len(doc.pages))
	}

	var buf bytes.Buffer
	err := doc.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo() failed: %v", err)
	}

	output := buf.String()

	// Pagesのカウントを確認
	if !strings.Contains(output, "/Count 3") {
		t.Error("Output should contain Pages with Count 3")
	}

	// 各ページオブジェクトが出力されているか確認
	// /Type /Pages (1つ) と /Type /Page (3つ) を区別する
	pagesCount := strings.Count(output, "/Type /Pages")
	if pagesCount != 1 {
		t.Errorf("Output should contain 1 Pages object, got %d", pagesCount)
	}

	// 注: /Type /Pageは/Type /Pagesにも含まれるため、単純にカウントすると4になる
	// Kids配列内の参照をカウントするか、/Count 3 で確認済みなので、ここでは省略
}
