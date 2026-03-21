package reader

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"testing"

	"github.com/ryomak/gopdf/internal/core"
)

// flateNewWriter はzlib.NewWriterのヘルパー
func flateNewWriter(w io.Writer) (*zlib.Writer, error) {
	return zlib.NewWriterLevel(w, zlib.DefaultCompression)
}

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

// TestReader_ResolveReference は参照解決をテストする
func TestReader_ResolveReference(t *testing.T) {
	pdf := createMinimalPDF()
	reader, err := NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	ref := &core.Reference{ObjectNumber: 1, GenerationNumber: 0}
	obj, err := reader.ResolveReference(ref)
	if err != nil {
		t.Fatalf("Failed to resolve reference: %v", err)
	}

	dict, ok := obj.(core.Dictionary)
	if !ok {
		t.Fatalf("Expected Dictionary, got %T", obj)
	}

	if dict[core.Name("Type")] != core.Name("Catalog") {
		t.Errorf("Type = %v, want Catalog", dict[core.Name("Type")])
	}
}

// TestReader_GetPageResources はページリソース取得をテストする
func TestReader_GetPageResources(t *testing.T) {
	pdf := createMinimalPDF()
	reader, err := NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	page, err := reader.GetPage(0)
	if err != nil {
		t.Fatalf("Failed to get page: %v", err)
	}

	resources, err := reader.GetPageResources(page)
	if err != nil {
		t.Fatalf("Failed to get page resources: %v", err)
	}

	if resources == nil {
		t.Fatal("Resources is nil")
	}

	// /Fontが含まれることを確認
	fontObj, ok := resources[core.Name("Font")]
	if !ok {
		t.Error("Resources has no /Font")
	}
	_, ok = fontObj.(core.Dictionary)
	if !ok {
		t.Errorf("Font should be Dictionary, got %T", fontObj)
	}
}

// TestReader_GetPageResources_NoResources はリソースなしページをテストする
func TestReader_GetPageResources_NoResources(t *testing.T) {
	pdf := createMinimalPDF()
	reader, err := NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	// リソースなしのダミーページ辞書
	page := core.Dictionary{
		core.Name("Type"): core.Name("Page"),
	}

	resources, err := reader.GetPageResources(page)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if resources != nil {
		t.Errorf("Expected nil resources, got %v", resources)
	}
}

// TestReader_GetPageContents はページコンテンツ取得をテストする
func TestReader_GetPageContents(t *testing.T) {
	pdf := createMinimalPDF()
	reader, err := NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	page, err := reader.GetPage(0)
	if err != nil {
		t.Fatalf("Failed to get page: %v", err)
	}

	contents, err := reader.GetPageContents(page)
	if err != nil {
		t.Fatalf("Failed to get page contents: %v", err)
	}

	if len(contents) == 0 {
		t.Error("Contents is empty")
	}

	// コンテンツにBTが含まれることを確認
	if !bytes.Contains(contents, []byte("BT")) {
		t.Error("Contents should contain BT operator")
	}
}

// TestReader_GetPageContents_NoContents はコンテンツなしページをテストする
func TestReader_GetPageContents_NoContents(t *testing.T) {
	pdf := createMinimalPDF()
	reader, err := NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	// コンテンツなしのダミーページ辞書
	page := core.Dictionary{
		core.Name("Type"): core.Name("Page"),
	}

	contents, err := reader.GetPageContents(page)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(contents) != 0 {
		t.Errorf("Expected empty contents, got %d bytes", len(contents))
	}
}

// TestReader_DecodeStream はストリームデコードをテストする
func TestReader_DecodeStream(t *testing.T) {
	pdf := createMinimalPDF()
	reader, err := NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	// フィルターなしストリームのデコード
	stream := &core.Stream{
		Dict: core.Dictionary{
			core.Name("Length"): core.Integer(5),
		},
		Data: []byte("Hello"),
	}

	decoded, err := reader.DecodeStream(stream)
	if err != nil {
		t.Fatalf("Failed to decode stream: %v", err)
	}
	if string(decoded) != "Hello" {
		t.Errorf("Decoded = %q, want %q", string(decoded), "Hello")
	}
}

// TestReader_DecodeStream_FlateDecode はzlib圧縮ストリームのデコードをテストする
func TestReader_DecodeStream_FlateDecode(t *testing.T) {
	pdf := createMinimalPDF()
	reader, err := NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	// zlib圧縮データを作成
	var compBuf bytes.Buffer
	w, err := flateNewWriter(&compBuf)
	if err != nil {
		t.Fatalf("Failed to create zlib writer: %v", err)
	}
	_, _ = w.Write([]byte("Hello, World!"))
	_ = w.Close()

	stream := &core.Stream{
		Dict: core.Dictionary{
			core.Name("Filter"): core.Name("FlateDecode"),
			core.Name("Length"): core.Integer(compBuf.Len()),
		},
		Data: compBuf.Bytes(),
	}

	decoded, err := reader.DecodeStream(stream)
	if err != nil {
		t.Fatalf("Failed to decode stream: %v", err)
	}
	if string(decoded) != "Hello, World!" {
		t.Errorf("Decoded = %q, want %q", string(decoded), "Hello, World!")
	}
}

// TestReader_DecodeStream_UnsupportedFilter はサポート外フィルターのテスト
func TestReader_DecodeStream_UnsupportedFilter(t *testing.T) {
	pdf := createMinimalPDF()
	reader, err := NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	stream := &core.Stream{
		Dict: core.Dictionary{
			core.Name("Filter"): core.Name("ASCIIHexDecode"),
			core.Name("Length"): core.Integer(5),
		},
		Data: []byte("Hello"),
	}

	// サポート外フィルターはそのまま返す
	decoded, err := reader.DecodeStream(stream)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if string(decoded) != "Hello" {
		t.Errorf("Decoded = %q, want %q", string(decoded), "Hello")
	}
}

// TestReader_DecodeStream_FilterArray はフィルター配列のテスト
func TestReader_DecodeStream_FilterArray(t *testing.T) {
	pdf := createMinimalPDF()
	reader, err := NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	// zlib圧縮データを作成
	var compBuf bytes.Buffer
	w, err := flateNewWriter(&compBuf)
	if err != nil {
		t.Fatalf("Failed to create zlib writer: %v", err)
	}
	_, _ = w.Write([]byte("Test Data"))
	_ = w.Close()

	stream := &core.Stream{
		Dict: core.Dictionary{
			core.Name("Filter"): core.Array{core.Name("FlateDecode")},
			core.Name("Length"): core.Integer(compBuf.Len()),
		},
		Data: compBuf.Bytes(),
	}

	decoded, err := reader.DecodeStream(stream)
	if err != nil {
		t.Fatalf("Failed to decode stream: %v", err)
	}
	if string(decoded) != "Test Data" {
		t.Errorf("Decoded = %q, want %q", string(decoded), "Test Data")
	}
}

// TestReader_IsEncrypted は暗号化チェックをテストする
func TestReader_IsEncrypted(t *testing.T) {
	pdf := createMinimalPDF()
	reader, err := NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	if reader.IsEncrypted() {
		t.Error("Expected not encrypted")
	}

	if reader.IsAuthenticated() {
		t.Error("Expected not authenticated")
	}

	if reader.GetEncryptionInfo() != nil {
		t.Error("Expected nil encryption info")
	}
}

// TestReader_AuthenticateWithPassword_NotEncrypted は非暗号化PDFでの認証エラーテスト
func TestReader_AuthenticateWithPassword_NotEncrypted(t *testing.T) {
	pdf := createMinimalPDF()
	reader, err := NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	err = reader.AuthenticateWithPassword("test")
	if err == nil {
		t.Error("Expected error for non-encrypted PDF, got nil")
	}
}

// TestReader_GetAllObjectNumbers はオブジェクト番号一覧取得をテストする
func TestReader_GetAllObjectNumbers(t *testing.T) {
	pdf := createMinimalPDF()
	reader, err := NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	nums := reader.GetAllObjectNumbers()
	// 5つの使用中オブジェクト（1-5）
	if len(nums) != 5 {
		t.Errorf("Expected 5 object numbers, got %d", len(nums))
	}

	// 各番号が1-5の範囲にあることを確認
	numSet := make(map[int]bool)
	for _, n := range nums {
		numSet[n] = true
	}
	for i := 1; i <= 5; i++ {
		if !numSet[i] {
			t.Errorf("Expected object number %d in list", i)
		}
	}
}

// TestReader_GetObjectGeneration はオブジェクト世代番号取得をテストする
func TestReader_GetObjectGeneration(t *testing.T) {
	pdf := createMinimalPDF()
	reader, err := NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	// 既存オブジェクトの世代番号
	gen := reader.GetObjectGeneration(1)
	if gen != 0 {
		t.Errorf("Generation = %d, want 0", gen)
	}

	// 存在しないオブジェクトの世代番号
	gen = reader.GetObjectGeneration(999)
	if gen != 0 {
		t.Errorf("Non-existent object generation = %d, want 0", gen)
	}
}

// TestReader_GetTrailer はtrailer取得をテストする
func TestReader_GetTrailer(t *testing.T) {
	pdf := createMinimalPDF()
	reader, err := NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	trailer := reader.GetTrailer()
	if trailer == nil {
		t.Fatal("Trailer is nil")
	}

	// /Sizeを確認
	sizeObj, ok := trailer[core.Name("Size")]
	if !ok {
		t.Error("Trailer has no /Size")
	}
	size, ok := sizeObj.(core.Integer)
	if !ok || int(size) != 6 {
		t.Errorf("Trailer /Size = %v, want 6", sizeObj)
	}

	// /Rootを確認
	rootObj, ok := trailer[core.Name("Root")]
	if !ok {
		t.Error("Trailer has no /Root")
	}
	rootRef, ok := rootObj.(*core.Reference)
	if !ok {
		t.Errorf("Trailer /Root should be Reference, got %T", rootObj)
	} else if rootRef.ObjectNumber != 1 {
		t.Errorf("Root ObjectNumber = %d, want 1", rootRef.ObjectNumber)
	}
}

// TestReader_GetObject_NotFound は存在しないオブジェクトのエラーテスト
func TestReader_GetObject_NotFound(t *testing.T) {
	pdf := createMinimalPDF()
	reader, err := NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	_, err = reader.GetObject(999)
	if err == nil {
		t.Error("Expected error for non-existent object, got nil")
	}
}

// TestReader_GetObject_NotInUse は使用されていないオブジェクトのエラーテスト
func TestReader_GetObject_NotInUse(t *testing.T) {
	pdf := createMinimalPDF()
	reader, err := NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	// object 0 is free (not in use)
	_, err = reader.GetObject(0)
	if err == nil {
		t.Error("Expected error for free object, got nil")
	}
}

// TestReader_GetRawObjectWithGeneration はRawオブジェクト取得をテストする
func TestReader_GetRawObjectWithGeneration(t *testing.T) {
	pdf := createMinimalPDF()
	reader, err := NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	obj, gen, err := reader.GetRawObjectWithGeneration(1)
	if err != nil {
		t.Fatalf("Failed to get raw object: %v", err)
	}

	if gen != 0 {
		t.Errorf("Generation = %d, want 0", gen)
	}

	dict, ok := obj.(core.Dictionary)
	if !ok {
		t.Fatalf("Expected Dictionary, got %T", obj)
	}

	if dict[core.Name("Type")] != core.Name("Catalog") {
		t.Errorf("Type = %v, want Catalog", dict[core.Name("Type")])
	}
}

// TestReader_GetRawObjectWithGeneration_NotFound は存在しないオブジェクトのRaw取得エラーテスト
func TestReader_GetRawObjectWithGeneration_NotFound(t *testing.T) {
	pdf := createMinimalPDF()
	reader, err := NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	_, _, err = reader.GetRawObjectWithGeneration(999)
	if err == nil {
		t.Error("Expected error for non-existent object, got nil")
	}
}

// TestReader_GetRawObjectWithGeneration_NotInUse は使用されていないオブジェクトのRaw取得エラーテスト
func TestReader_GetRawObjectWithGeneration_NotInUse(t *testing.T) {
	pdf := createMinimalPDF()
	reader, err := NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	_, _, err = reader.GetRawObjectWithGeneration(0)
	if err == nil {
		t.Error("Expected error for free object, got nil")
	}
}

// TestReader_GetObject_StreamObject はストリームオブジェクトの取得をテストする
func TestReader_GetObject_StreamObject(t *testing.T) {
	pdf := createMinimalPDF()
	reader, err := NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	// Object 4 is the contents stream
	obj, err := reader.GetObject(4)
	if err != nil {
		t.Fatalf("Failed to get stream object: %v", err)
	}

	stream, ok := obj.(*core.Stream)
	if !ok {
		t.Fatalf("Expected Stream, got %T", obj)
	}

	if len(stream.Data) == 0 {
		t.Error("Stream data is empty")
	}

	// Dictに/Lengthが含まれること
	_, hasLength := stream.Dict[core.Name("Length")]
	if !hasLength {
		t.Error("Stream dict has no /Length")
	}
}

// TestReader_isEncryptObject は暗号化オブジェクト判定をテストする
func TestReader_isEncryptObject(t *testing.T) {
	pdf := createMinimalPDF()
	reader, err := NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	// 暗号化なしの場合は常にfalse
	if reader.isEncryptObject(1) {
		t.Error("Expected false for non-encrypted PDF")
	}
}

// TestReader_GetImageXObject_NotImage は画像でないXObjectのエラーテスト
func TestReader_GetImageXObject_NotImage(t *testing.T) {
	pdf := createMinimalPDF()
	reader, err := NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	// Object 5 is a Font, not a Stream
	ref := &core.Reference{ObjectNumber: 5, GenerationNumber: 0}
	_, err = reader.GetImageXObject(ref)
	if err == nil {
		t.Error("Expected error for non-stream object, got nil")
	}
}

// createMinimalPDFWithImageXObject は画像XObjectを含むPDFを作成する
func createMinimalPDFWithImageXObject() []byte {
	var buf bytes.Buffer

	header := "%PDF-1.7\n\n"
	buf.WriteString(header)

	offsets := make([]int, 7)

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

	// Object 3: Page with XObject resource
	offsets[3] = buf.Len()
	buf.WriteString("3 0 obj\n")
	buf.WriteString("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << /XObject << /Im1 6 0 R >> >> >>\n")
	buf.WriteString("endobj\n\n")

	// Object 4: Contents
	streamContent := "BT /F1 12 Tf 100 700 Td (Hello) Tj ET\n"
	offsets[4] = buf.Len()
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

	// Object 6: Image XObject
	imgData := "\xff\xd8\xff\xe0" // Fake JPEG header
	offsets[6] = buf.Len()
	buf.WriteString("6 0 obj\n")
	buf.WriteString(fmt.Sprintf("<< /Type /XObject /Subtype /Image /Width 100 /Height 100 /ColorSpace /DeviceRGB /BitsPerComponent 8 /Filter /DCTDecode /Length %d >>\n", len(imgData)))
	buf.WriteString("stream\n")
	buf.Write([]byte(imgData))
	buf.WriteString("\nendstream\n")
	buf.WriteString("endobj\n\n")

	xrefStart := buf.Len()

	buf.WriteString("xref\n")
	buf.WriteString("0 7\n")
	buf.WriteString("0000000000 65535 f \n")
	for i := 1; i <= 6; i++ {
		buf.WriteString(fmt.Sprintf("%010d 00000 n \n", offsets[i]))
	}

	buf.WriteString("trailer\n")
	buf.WriteString("<< /Size 7 /Root 1 0 R >>\n")
	buf.WriteString("startxref\n")
	buf.WriteString(fmt.Sprintf("%d\n", xrefStart))
	buf.WriteString("%%EOF")

	return buf.Bytes()
}

// TestReader_GetImageXObject はImageXObject取得をテストする
func TestReader_GetImageXObject(t *testing.T) {
	pdf := createMinimalPDFWithImageXObject()
	reader, err := NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	ref := &core.Reference{ObjectNumber: 6, GenerationNumber: 0}
	img, err := reader.GetImageXObject(ref)
	if err != nil {
		t.Fatalf("Failed to get image xobject: %v", err)
	}

	if img.Width != 100 {
		t.Errorf("Width = %d, want 100", img.Width)
	}
	if img.Height != 100 {
		t.Errorf("Height = %d, want 100", img.Height)
	}
	if img.ColorSpace != "DeviceRGB" {
		t.Errorf("ColorSpace = %s, want DeviceRGB", img.ColorSpace)
	}
	if img.BitsPerComponent != 8 {
		t.Errorf("BitsPerComponent = %d, want 8", img.BitsPerComponent)
	}
	if img.Filter != "DCTDecode" {
		t.Errorf("Filter = %s, want DCTDecode", img.Filter)
	}
}

// TestReader_GetPageContents_InvalidType はコンテンツが不正な型の場合のエラーテスト
func TestReader_GetPageContents_InvalidType(t *testing.T) {
	pdf := createMinimalPDF()
	reader, err := NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	// コンテンツが整数の場合（不正）
	page := core.Dictionary{
		core.Name("Type"):     core.Name("Page"),
		core.Name("Contents"): core.Integer(42),
	}

	_, err = reader.GetPageContents(page)
	if err == nil {
		t.Error("Expected error for invalid contents type, got nil")
	}
}

// TestReader_GetPageResources_Reference はリソースが参照の場合のテスト
func TestReader_GetPageResources_Reference(t *testing.T) {
	pdf := createMinimalPDFWithImageXObject()
	reader, err := NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	page, err := reader.GetPage(0)
	if err != nil {
		t.Fatalf("Failed to get page: %v", err)
	}

	resources, err := reader.GetPageResources(page)
	if err != nil {
		t.Fatalf("Failed to get page resources: %v", err)
	}

	if resources == nil {
		t.Fatal("Resources is nil")
	}

	// /XObjectが含まれることを確認
	_, ok := resources[core.Name("XObject")]
	if !ok {
		t.Error("Resources has no /XObject")
	}
}

// TestReader_GetPageContents_WithContentsReference はContentsが参照の場合のテスト
func TestReader_GetPageContents_WithContentsReference(t *testing.T) {
	pdf := createMinimalPDF()
	reader, err := NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	// Contentsを参照として持つページを作成
	page := core.Dictionary{
		core.Name("Type"):     core.Name("Page"),
		core.Name("Contents"): &core.Reference{ObjectNumber: 4, GenerationNumber: 0},
	}

	contents, err := reader.GetPageContents(page)
	if err != nil {
		t.Fatalf("Failed to get page contents: %v", err)
	}
	if len(contents) == 0 {
		t.Error("Contents is empty")
	}
	if !bytes.Contains(contents, []byte("BT")) {
		t.Error("Contents should contain BT operator")
	}
}

// TestReader_GetPageContents_WithContentsArray はContentsが配列参照の場合のテスト
func TestReader_GetPageContents_WithContentsArray(t *testing.T) {
	pdf := createMinimalPDF()
	reader, err := NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	// Contentsを配列として持つページ（同じストリームを2回参照）
	page := core.Dictionary{
		core.Name("Type"): core.Name("Page"),
		core.Name("Contents"): core.Array{
			&core.Reference{ObjectNumber: 4, GenerationNumber: 0},
		},
	}

	contents, err := reader.GetPageContents(page)
	if err != nil {
		t.Fatalf("Failed to get page contents: %v", err)
	}
	if len(contents) == 0 {
		t.Error("Contents is empty")
	}
}

// TestReader_GetPageContents_WithInlineStream はContentsがインラインStreamの場合のテスト
func TestReader_GetPageContents_WithInlineStream(t *testing.T) {
	pdf := createMinimalPDF()
	reader, err := NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	streamData := []byte("BT /F1 12 Tf (Test) Tj ET")
	page := core.Dictionary{
		core.Name("Type"): core.Name("Page"),
		core.Name("Contents"): &core.Stream{
			Dict: core.Dictionary{
				core.Name("Length"): core.Integer(len(streamData)),
			},
			Data: streamData,
		},
	}

	contents, err := reader.GetPageContents(page)
	if err != nil {
		t.Fatalf("Failed to get page contents: %v", err)
	}
	if string(contents) != string(streamData) {
		t.Errorf("Contents = %q, want %q", string(contents), string(streamData))
	}
}

// TestReader_decryptObject は各型のオブジェクト復号化をテストする
func TestReader_decryptObject(t *testing.T) {
	pdf := createMinimalPDF()
	reader, err := NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	// 暗号化情報をモック（Authenticated=false）
	reader.encryption = &EncryptionInfo{
		Authenticated: false,
	}

	// String型: 未認証なのでそのまま返る
	strObj := core.String("test")
	result := reader.decryptObject(strObj, 1, 0)
	if s, ok := result.(core.String); !ok || string(s) != "test" {
		t.Errorf("Unauthenticated string decrypt failed: %v", result)
	}

	// Dictionary型: 未認証なのでそのまま返る
	dictObj := core.Dictionary{
		core.Name("Key"): core.String("value"),
	}
	result = reader.decryptObject(dictObj, 1, 0)
	if _, ok := result.(core.Dictionary); !ok {
		t.Errorf("Expected Dictionary, got %T", result)
	}

	// Array型: 未認証なのでそのまま返る
	arrObj := core.Array{core.String("item1"), core.Integer(42)}
	result = reader.decryptObject(arrObj, 1, 0)
	if arr, ok := result.(core.Array); !ok || len(arr) != 2 {
		t.Errorf("Expected Array of length 2, got %v", result)
	}

	// Stream型: 未認証なのでそのまま返る
	streamObj := &core.Stream{
		Dict: core.Dictionary{
			core.Name("Length"): core.Integer(5),
			core.Name("Filter"): core.Name("FlateDecode"),
			core.Name("Key"):    core.String("val"),
		},
		Data: []byte("hello"),
	}
	result = reader.decryptObject(streamObj, 1, 0)
	if _, ok := result.(*core.Stream); !ok {
		t.Errorf("Expected Stream, got %T", result)
	}

	// Integer型: 復号化不要
	intObj := core.Integer(42)
	result = reader.decryptObject(intObj, 1, 0)
	if result != intObj {
		t.Errorf("Integer should be unchanged, got %v", result)
	}

	// Boolean型: 復号化不要
	boolObj := core.Boolean(true)
	result = reader.decryptObject(boolObj, 1, 0)
	if result != boolObj {
		t.Errorf("Boolean should be unchanged, got %v", result)
	}

	// nil型
	result = reader.decryptObject(nil, 1, 0)
	if result != nil {
		t.Errorf("nil should remain nil, got %v", result)
	}
}

// TestReader_isEncryptObject_WithEncryption は暗号化PDFでの暗号化オブジェクト判定をテストする
func TestReader_isEncryptObject_WithEncryption(t *testing.T) {
	pdf := createMinimalPDF()
	reader, err := NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	// 暗号化情報をセットし、trailerにEncrypt参照を追加
	reader.encryption = &EncryptionInfo{}
	reader.trailer[core.Name("Encrypt")] = &core.Reference{ObjectNumber: 10, GenerationNumber: 0}

	if !reader.isEncryptObject(10) {
		t.Error("Expected true for encrypt object")
	}
	if reader.isEncryptObject(1) {
		t.Error("Expected false for non-encrypt object")
	}
}

// TestReader_GetPageResources_WithReference はResourcesが参照の場合のテスト
func TestReader_GetPageResources_WithReference(t *testing.T) {
	// ResourcesがIndirectReferenceのPDFを作成する
	var buf bytes.Buffer
	header := "%PDF-1.7\n\n"
	buf.WriteString(header)

	offsets := make([]int, 5)

	// Object 1: Catalog
	offsets[1] = buf.Len()
	buf.WriteString("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n\n")

	// Object 2: Pages
	offsets[2] = buf.Len()
	buf.WriteString("2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n\n")

	// Object 3: Page with Resources as reference
	offsets[3] = buf.Len()
	buf.WriteString("3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources 4 0 R >>\nendobj\n\n")

	// Object 4: Resources dictionary
	offsets[4] = buf.Len()
	buf.WriteString("4 0 obj\n<< /Font << /F1 << /Type /Font /Subtype /Type1 /BaseFont /Helvetica >> >> >>\nendobj\n\n")

	xrefStart := buf.Len()
	buf.WriteString("xref\n0 5\n")
	buf.WriteString("0000000000 65535 f \n")
	for i := 1; i <= 4; i++ {
		buf.WriteString(fmt.Sprintf("%010d 00000 n \n", offsets[i]))
	}
	buf.WriteString("trailer\n<< /Size 5 /Root 1 0 R >>\nstartxref\n")
	buf.WriteString(fmt.Sprintf("%d\n", xrefStart))
	buf.WriteString("%%EOF")

	reader, err := NewReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	page, err := reader.GetPage(0)
	if err != nil {
		t.Fatalf("Failed to get page: %v", err)
	}

	resources, err := reader.GetPageResources(page)
	if err != nil {
		t.Fatalf("Failed to get resources: %v", err)
	}

	if resources == nil {
		t.Fatal("Resources is nil")
	}

	_, hasFonts := resources[core.Name("Font")]
	if !hasFonts {
		t.Error("Resources should have /Font")
	}
}

// TestReader_NewReader_InvalidPDF は不正なPDFの読み込みエラーテスト
func TestReader_NewReader_InvalidPDF(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{
			name:  "Empty input",
			input: []byte(""),
		},
		{
			name:  "No startxref",
			input: []byte("%PDF-1.7\n%%EOF"),
		},
		{
			name:  "Invalid xref offset",
			input: []byte("%PDF-1.7\nstartxref\nabc\n%%EOF"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewReader(bytes.NewReader(tt.input))
			if err == nil {
				t.Error("Expected error for invalid PDF, got nil")
			}
		})
	}
}

// TestReader_EncryptionInfo_DecryptStream_Unauthenticated は未認証時のDecryptStreamテスト
func TestReader_EncryptionInfo_DecryptStream_Unauthenticated(t *testing.T) {
	ei := &EncryptionInfo{
		Authenticated: false,
	}

	data := []byte("test data")
	result := ei.DecryptStream(data, 1, 0)
	if !bytes.Equal(result, data) {
		t.Errorf("Unauthenticated should return data as-is")
	}
}

// TestReader_EncryptionInfo_DecryptString_Unauthenticated は未認証時のDecryptStringテスト
func TestReader_EncryptionInfo_DecryptString_Unauthenticated(t *testing.T) {
	ei := &EncryptionInfo{
		Authenticated: false,
	}

	data := []byte("test string")
	result := ei.DecryptString(data, 1, 0)
	if result != "test string" {
		t.Errorf("Unauthenticated should return string as-is, got %q", result)
	}
}

// TestParseEncryptDict はEncrypt辞書のパースをテストする
func TestParseEncryptDict(t *testing.T) {
	t.Run("Valid encrypt dict", func(t *testing.T) {
		dict := core.Dictionary{
			core.Name("Filter"): core.Name("Standard"),
			core.Name("V"):      core.Integer(1),
			core.Name("R"):      core.Integer(2),
			core.Name("O"):      core.String("owner_password_hash_32_bytes_pad!"),
			core.Name("U"):      core.String("user_password_hash_32_bytes_pad!!"),
			core.Name("P"):      core.Integer(-44),
			core.Name("Length"): core.Integer(40),
		}

		info, err := parseEncryptDict(dict, []byte("file-id"))
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if info.Filter != "Standard" {
			t.Errorf("Filter = %s, want Standard", info.Filter)
		}
		if info.V != 1 {
			t.Errorf("V = %d, want 1", info.V)
		}
		if info.R != 2 {
			t.Errorf("R = %d, want 2", info.R)
		}
		if info.Length != 40 {
			t.Errorf("Length = %d, want 40", info.Length)
		}
		if info.KeyLengthBytes != 5 {
			t.Errorf("KeyLengthBytes = %d, want 5", info.KeyLengthBytes)
		}
	})

	t.Run("Default length", func(t *testing.T) {
		dict := core.Dictionary{
			core.Name("Filter"): core.Name("Standard"),
			core.Name("V"):      core.Integer(1),
			core.Name("R"):      core.Integer(2),
			core.Name("O"):      core.String("owner_password_hash_32_bytes_pad!"),
			core.Name("U"):      core.String("user_password_hash_32_bytes_pad!!"),
			core.Name("P"):      core.Integer(-44),
		}

		info, err := parseEncryptDict(dict, []byte("file-id"))
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if info.Length != 40 {
			t.Errorf("Default length = %d, want 40", info.Length)
		}
	})

	t.Run("Missing Filter", func(t *testing.T) {
		dict := core.Dictionary{
			core.Name("V"): core.Integer(1),
		}
		_, err := parseEncryptDict(dict, []byte("file-id"))
		if err == nil {
			t.Error("Expected error for missing Filter")
		}
	})

	t.Run("Unsupported Filter", func(t *testing.T) {
		dict := core.Dictionary{
			core.Name("Filter"): core.Name("Custom"),
		}
		_, err := parseEncryptDict(dict, []byte("file-id"))
		if err == nil {
			t.Error("Expected error for unsupported Filter")
		}
	})

	t.Run("Missing V", func(t *testing.T) {
		dict := core.Dictionary{
			core.Name("Filter"): core.Name("Standard"),
		}
		_, err := parseEncryptDict(dict, []byte("file-id"))
		if err == nil {
			t.Error("Expected error for missing V")
		}
	})

	t.Run("Missing R", func(t *testing.T) {
		dict := core.Dictionary{
			core.Name("Filter"): core.Name("Standard"),
			core.Name("V"):      core.Integer(1),
		}
		_, err := parseEncryptDict(dict, []byte("file-id"))
		if err == nil {
			t.Error("Expected error for missing R")
		}
	})

	t.Run("Missing O", func(t *testing.T) {
		dict := core.Dictionary{
			core.Name("Filter"): core.Name("Standard"),
			core.Name("V"):      core.Integer(1),
			core.Name("R"):      core.Integer(2),
		}
		_, err := parseEncryptDict(dict, []byte("file-id"))
		if err == nil {
			t.Error("Expected error for missing O")
		}
	})

	t.Run("Missing U", func(t *testing.T) {
		dict := core.Dictionary{
			core.Name("Filter"): core.Name("Standard"),
			core.Name("V"):      core.Integer(1),
			core.Name("R"):      core.Integer(2),
			core.Name("O"):      core.String("owner_hash"),
		}
		_, err := parseEncryptDict(dict, []byte("file-id"))
		if err == nil {
			t.Error("Expected error for missing U")
		}
	})

	t.Run("Missing P", func(t *testing.T) {
		dict := core.Dictionary{
			core.Name("Filter"): core.Name("Standard"),
			core.Name("V"):      core.Integer(1),
			core.Name("R"):      core.Integer(2),
			core.Name("O"):      core.String("owner_hash"),
			core.Name("U"):      core.String("user_hash"),
		}
		_, err := parseEncryptDict(dict, []byte("file-id"))
		if err == nil {
			t.Error("Expected error for missing P")
		}
	})
}

// TestReader_detectEncryption_NoEncrypt は暗号化なしのdetectEncryptionテスト
func TestReader_detectEncryption_NoEncrypt(t *testing.T) {
	pdf := createMinimalPDF()
	reader, err := NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	// detectEncryption is already called during NewReader
	// and should not set encryption
	if reader.encryption != nil {
		t.Error("Expected nil encryption for non-encrypted PDF")
	}
}

// TestReader_detectEncryption_InvalidEncryptType は不正なEncryptエントリ型のテスト
func TestReader_detectEncryption_InvalidEncryptType(t *testing.T) {
	pdf := createMinimalPDF()
	reader, err := NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	// 不正な型のEncryptエントリを設定
	reader.trailer[core.Name("Encrypt")] = core.Integer(42)
	err = reader.detectEncryption()
	if err == nil {
		t.Error("Expected error for invalid Encrypt type")
	}
}

// TestReader_detectEncryption_DirectDict はEncryptが直接辞書の場合のテスト
func TestReader_detectEncryption_DirectDict(t *testing.T) {
	pdf := createMinimalPDF()
	reader, err := NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	// 直接辞書のEncryptエントリ（IDなしなのでエラーになるはず）
	reader.trailer[core.Name("Encrypt")] = core.Dictionary{
		core.Name("Filter"): core.Name("Standard"),
		core.Name("V"):      core.Integer(1),
	}
	err = reader.detectEncryption()
	if err == nil {
		t.Error("Expected error for missing File ID")
	}
}

// TestReader_detectEncryption_WithIDAndDirectDict はEncryptが直接辞書でIDありの場合のテスト
func TestReader_detectEncryption_WithIDAndDirectDict(t *testing.T) {
	pdf := createMinimalPDF()
	reader, err := NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	// IDを設定
	reader.trailer[core.Name("ID")] = core.Array{
		core.String("1234567890123456"),
		core.String("1234567890123456"),
	}

	// 直接辞書のEncryptエントリ（Filterが不正なのでエラー）
	reader.trailer[core.Name("Encrypt")] = core.Dictionary{
		core.Name("Filter"): core.Name("Custom"),
	}
	err = reader.detectEncryption()
	if err == nil {
		t.Error("Expected error for unsupported filter")
	}
}

// TestReader_detectEncryption_WithReference はEncryptが参照の場合のテスト
func TestReader_detectEncryption_WithReference(t *testing.T) {
	pdf := createMinimalPDF()
	reader, err := NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	// IDを設定
	reader.trailer[core.Name("ID")] = core.Array{
		core.String("1234567890123456"),
		core.String("1234567890123456"),
	}

	// Encryptを存在しないオブジェクトへの参照に設定
	reader.trailer[core.Name("Encrypt")] = &core.Reference{ObjectNumber: 999, GenerationNumber: 0}
	err = reader.detectEncryption()
	if err == nil {
		t.Error("Expected error for non-existent encrypt reference")
	}
}

// TestReader_decodeStream_FilterNameOnly はフィルター名のみのdecodeStreamテスト
func TestReader_decodeStream_FilterNameOnly(t *testing.T) {
	pdf := createMinimalPDF()
	reader, err := NewReader(bytes.NewReader(pdf))
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}

	// zlib圧縮データを作成
	var compBuf bytes.Buffer
	w, err := flateNewWriter(&compBuf)
	if err != nil {
		t.Fatalf("Failed to create zlib writer: %v", err)
	}
	_, _ = w.Write([]byte("Test"))
	_ = w.Close()

	// 不正なzlibデータでエラーを発生させる
	stream := &core.Stream{
		Dict: core.Dictionary{
			core.Name("Filter"): core.Name("FlateDecode"),
		},
		Data: []byte("not-valid-zlib-data"),
	}

	_, err = reader.decodeStream(stream)
	if err == nil {
		t.Error("Expected error for invalid zlib data")
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
