package gopdf

import (
	"bytes"
	"os"
	"testing"
)

// TestOpen はファイルからPDF読み込みをテストする
func TestOpen(t *testing.T) {
	// WriterでPDFを生成
	doc := New()
	page := doc.AddPage(PageSizeA4, Portrait)
	_ = page.DrawText("Test", 100, 700) // エラー無視（テストの主目的ではない）

	// 一時ファイルに書き込み
	tmpfile, err := os.CreateTemp("", "test*.pdf")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if err := doc.WriteTo(tmpfile); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	// PDFを読み込む
	reader, err := Open(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to open PDF: %v", err)
	}
	defer reader.Close()

	// ページ数を確認
	if reader.PageCount() != 1 {
		t.Errorf("PageCount = %d, want 1", reader.PageCount())
	}
}

// TestOpenReader はReaderからPDF読み込みをテストする
func TestOpenReader(t *testing.T) {
	// WriterでPDFを生成
	doc := New()
	page := doc.AddPage(PageSizeA4, Portrait)
	_ = page.DrawText("Test", 100, 700) // エラー無視（テストの主目的ではない）

	// バッファに書き込み
	var buf bytes.Buffer
	if err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	// PDFを読み込む
	reader, err := OpenReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("Failed to open PDF: %v", err)
	}

	// ページ数を確認
	if reader.PageCount() != 1 {
		t.Errorf("PageCount = %d, want 1", reader.PageCount())
	}
}

// TestPDFReader_PageCount はPageCountメソッドをテストする
func TestPDFReader_PageCount(t *testing.T) {
	tests := []struct {
		name      string
		pageCount int
	}{
		{"1 page", 1},
		{"3 pages", 3},
		{"5 pages", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// PDFを生成
			doc := New()
			for i := 0; i < tt.pageCount; i++ {
				doc.AddPage(PageSizeA4, Portrait)
			}

			var buf bytes.Buffer
			if err := doc.WriteTo(&buf); err != nil {
				t.Fatalf("Failed to write PDF: %v", err)
			}

			// 読み込み
			reader, err := OpenReader(bytes.NewReader(buf.Bytes()))
			if err != nil {
				t.Fatalf("Failed to open PDF: %v", err)
			}

			if reader.PageCount() != tt.pageCount {
				t.Errorf("PageCount = %d, want %d", reader.PageCount(), tt.pageCount)
			}
		})
	}
}

// TestPDFReader_Info はInfoメソッドをテストする
func TestPDFReader_Info(t *testing.T) {
	// Writerで生成したPDFにはInfo辞書がないため、空のMetadataが返る
	doc := New()
	doc.AddPage(PageSizeA4, Portrait)

	var buf bytes.Buffer
	if err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("Failed to write PDF: %v", err)
	}

	reader, err := OpenReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("Failed to open PDF: %v", err)
	}

	info := reader.Info()

	// 空のメタデータを確認
	if info.Title != "" || info.Author != "" {
		t.Errorf("Expected empty metadata, got Title=%q, Author=%q", info.Title, info.Author)
	}
}
