package reader

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/ryomak/gopdf/internal/core"
)

// createMinimalPDF は最小限のPDFを作成する
func createMinimalPDF() []byte {
	var buf bytes.Buffer

	// ヘッダー
	header := "%PDF-1.7\n\n"
	buf.WriteString(header)

	// オブジェクトの位置を記録
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

	// Object 4: Contents (with stream)
	offsets[4] = buf.Len()
	streamContent := "BT\n/F1 12 Tf\n100 700 Td\n(Hello, World!) Tj\nET\n"
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

	// xrefの開始位置を記録
	xrefStart := buf.Len()

	// xref table
	buf.WriteString("xref\n")
	buf.WriteString("0 6\n")
	buf.WriteString("0000000000 65535 f \n")
	for i := 1; i <= 5; i++ {
		buf.WriteString(fmt.Sprintf("%010d 00000 n \n", offsets[i]))
	}

	// trailer
	buf.WriteString("trailer\n")
	buf.WriteString("<< /Size 6 /Root 1 0 R >>\n")
	buf.WriteString("startxref\n")
	buf.WriteString(fmt.Sprintf("%d\n", xrefStart))
	buf.WriteString("%%EOF")

	return buf.Bytes()
}

// TestNewReader はReaderの基本的な作成をテストする
func TestNewReader(t *testing.T) {
	pdf := createMinimalPDF()
	reader, err := NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	// xrefエントリ数を確認
	if len(reader.xref) != 6 {
		t.Errorf("Expected 6 xref entries, got %d", len(reader.xref))
	}

	// trailerを確認
	if reader.trailer == nil {
		t.Fatal("Trailer is nil")
	}

	// /Sizeを確認
	sizeObj, ok := reader.trailer[core.Name("Size")]
	if !ok {
		t.Fatal("Trailer has no /Size")
	}
	size, ok := sizeObj.(core.Integer)
	if !ok || int(size) != 6 {
		t.Errorf("Trailer /Size = %v, want 6", sizeObj)
	}
}

// TestReader_FindStartXref はstartxref検索をテストする
func TestReader_FindStartXref(t *testing.T) {
	pdf := createMinimalPDF()
	reader := &Reader{r: bytes.NewReader(pdf)}

	offset, err := reader.findStartXref()
	if err != nil {
		t.Fatalf("Failed to find startxref: %v", err)
	}

	// オフセットが妥当な範囲にあることを確認
	if offset <= 0 || offset >= int64(len(pdf)) {
		t.Errorf("startxref offset = %d, should be between 0 and %d", offset, len(pdf))
	}
}

// TestReader_GetCatalog はCatalog取得をテストする
func TestReader_GetCatalog(t *testing.T) {
	pdf := createMinimalPDF()
	reader, err := NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	catalog, err := reader.GetCatalog()
	if err != nil {
		t.Fatalf("Failed to get catalog: %v", err)
	}

	// /Typeを確認
	typeObj, ok := catalog[core.Name("Type")]
	if !ok || typeObj != core.Name("Catalog") {
		t.Errorf("Catalog /Type = %v, want Catalog", typeObj)
	}

	// /Pagesを確認
	pagesObj, ok := catalog[core.Name("Pages")]
	if !ok {
		t.Error("Catalog has no /Pages")
	}
	if _, ok := pagesObj.(*core.Reference); !ok {
		t.Errorf("Catalog /Pages should be reference, got %T", pagesObj)
	}
}

// TestReader_GetPageCount はページ数取得をテストする
func TestReader_GetPageCount(t *testing.T) {
	pdf := createMinimalPDF()
	reader, err := NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	count, err := reader.GetPageCount()
	if err != nil {
		t.Fatalf("Failed to get page count: %v", err)
	}

	if count != 1 {
		t.Errorf("Page count = %d, want 1", count)
	}
}

// TestReader_GetPage はページ取得をテストする
func TestReader_GetPage(t *testing.T) {
	pdf := createMinimalPDF()
	reader, err := NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	page, err := reader.GetPage(0)
	if err != nil {
		t.Fatalf("Failed to get page 0: %v", err)
	}

	// /Typeを確認
	typeObj, ok := page[core.Name("Type")]
	if !ok || typeObj != core.Name("Page") {
		t.Errorf("Page /Type = %v, want Page", typeObj)
	}

	// /MediaBoxを確認
	mediaBoxObj, ok := page[core.Name("MediaBox")]
	if !ok {
		t.Fatal("Page has no /MediaBox")
	}
	mediaBox, ok := mediaBoxObj.(core.Array)
	if !ok {
		t.Fatalf("Page /MediaBox should be array, got %T", mediaBoxObj)
	}
	if len(mediaBox) != 4 {
		t.Errorf("MediaBox length = %d, want 4", len(mediaBox))
	}
}

// TestReader_GetObject はオブジェクト取得をテストする
func TestReader_GetObject(t *testing.T) {
	pdf := createMinimalPDF()
	reader, err := NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	tests := []struct {
		name   string
		objNum int
	}{
		{"Catalog", 1},
		{"Pages", 2},
		{"Page", 3},
		// Contents (4) はStreamなので今は除外
		{"Font", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj, err := reader.GetObject(tt.objNum)
			if err != nil {
				t.Fatalf("Failed to get object %d: %v", tt.objNum, err)
			}
			if obj == nil {
				t.Errorf("Object %d is nil", tt.objNum)
			}

			// 2回目の取得はキャッシュから（エラーなく取得できればOK）
			_, err = reader.GetObject(tt.objNum)
			if err != nil {
				t.Fatalf("Failed to get cached object %d: %v", tt.objNum, err)
			}
		})
	}
}

// TestReader_GetInfo はInfo辞書取得をテストする
func TestReader_GetInfo(t *testing.T) {
	// Infoを持つPDFを作成
	var buf bytes.Buffer

	content := `%PDF-1.7

1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj

2 0 obj
<< /Type /Pages /Kids [] /Count 0 >>
endobj

3 0 obj
<< /Title (Test Document) /Author (Test Author) >>
endobj

`
	buf.WriteString(content)
	xrefStart := buf.Len()

	xref := `xref
0 4
0000000000 65535 f
0000000010 00000 n
0000000060 00000 n
0000000112 00000 n
`
	buf.WriteString(xref)

	trailer := `trailer
<< /Size 4 /Root 1 0 R /Info 3 0 R >>
startxref
`
	buf.WriteString(trailer)
	buf.WriteString(fmt.Sprintf("%d\n", xrefStart))
	buf.WriteString("%%EOF")

	pdf := buf.Bytes()

	reader, err := NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	info, err := reader.GetInfo()
	if err != nil {
		t.Fatalf("Failed to get info: %v", err)
	}

	// /Titleを確認
	titleObj, ok := info[core.Name("Title")]
	if !ok {
		t.Error("Info has no /Title")
	} else {
		title, ok := titleObj.(core.String)
		if !ok || string(title) != "Test Document" {
			t.Errorf("Info /Title = %v, want 'Test Document'", titleObj)
		}
	}

	// /Authorを確認
	authorObj, ok := info[core.Name("Author")]
	if !ok {
		t.Error("Info has no /Author")
	} else {
		author, ok := authorObj.(core.String)
		if !ok || string(author) != "Test Author" {
			t.Errorf("Info /Author = %v, want 'Test Author'", authorObj)
		}
	}
}

// TestReader_GetInfo_NoInfo はInfoがない場合をテストする
func TestReader_GetInfo_NoInfo(t *testing.T) {
	pdf := createMinimalPDF()
	reader, err := NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	info, err := reader.GetInfo()
	if err != nil {
		t.Fatalf("Failed to get info: %v", err)
	}

	// Infoがない場合は空の辞書が返る
	if len(info) != 0 {
		t.Errorf("Info should be empty, got %d entries", len(info))
	}
}

// TestReader_GetPage_OutOfRange は範囲外のページ取得をテストする
func TestReader_GetPage_OutOfRange(t *testing.T) {
	pdf := createMinimalPDF()
	reader, err := NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	_, err = reader.GetPage(1)
	if err == nil {
		t.Error("Expected error for out of range page, but got none")
	}

	_, err = reader.GetPage(-1)
	if err == nil {
		t.Error("Expected error for negative page number, but got none")
	}
}
