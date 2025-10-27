package writer

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ryomak/gopdf/internal/core"
)

// TestWriterHeader はPDFヘッダーの出力をテストする
func TestWriterHeader(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)

	err := w.WriteHeader()
	if err != nil {
		t.Fatalf("WriteHeader() failed: %v", err)
	}

	got := buf.String()
	want := "%PDF-1.7\n"
	if got != want {
		t.Errorf("WriteHeader() = %q, want %q", got, want)
	}
}

// TestWriterAddObject はオブジェクトの追加をテストする
func TestWriterAddObject(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)

	// ヘッダーを書く
	if err := w.WriteHeader(); err != nil {
		t.Fatalf("WriteHeader() failed: %v", err)
	}

	// オブジェクトを追加
	dict := core.Dictionary{
		core.Name("Type"): core.Name("Catalog"),
	}
	objNum, err := w.AddObject(dict)
	if err != nil {
		t.Fatalf("AddObject() failed: %v", err)
	}

	if objNum != 1 {
		t.Errorf("First object number = %d, want 1", objNum)
	}

	// 2つ目のオブジェクト
	dict2 := core.Dictionary{
		core.Name("Type"): core.Name("Pages"),
	}
	objNum2, err := w.AddObject(dict2)
	if err != nil {
		t.Fatalf("AddObject() failed: %v", err)
	}

	if objNum2 != 2 {
		t.Errorf("Second object number = %d, want 2", objNum2)
	}
}

// TestWriteSimplePDF は最小限のPDF生成をテストする
func TestWriteSimplePDF(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)

	// ヘッダー
	if err := w.WriteHeader(); err != nil {
		t.Fatalf("WriteHeader() failed: %v", err)
	}

	// Catalogオブジェクト
	catalogNum, err := w.AddObject(core.Dictionary{
		core.Name("Type"):  core.Name("Catalog"),
		core.Name("Pages"): &core.Reference{ObjectNumber: 2, GenerationNumber: 0},
	})
	if err != nil {
		t.Fatalf("AddObject(Catalog) failed: %v", err)
	}

	// Pagesオブジェクト
	_, err = w.AddObject(core.Dictionary{
		core.Name("Type"):  core.Name("Pages"),
		core.Name("Kids"):  core.Array{},
		core.Name("Count"): core.Integer(0),
	})
	if err != nil {
		t.Fatalf("AddObject(Pages) failed: %v", err)
	}

	// Trailer
	trailer := core.Dictionary{
		core.Name("Size"): core.Integer(3),
		core.Name("Root"): &core.Reference{ObjectNumber: catalogNum, GenerationNumber: 0},
	}

	if err := w.WriteTrailer(trailer); err != nil {
		t.Fatalf("WriteTrailer() failed: %v", err)
	}

	output := buf.String()

	// 必須要素の確認
	if !strings.Contains(output, "%PDF-1.7") {
		t.Error("Output should contain PDF header")
	}
	if !strings.Contains(output, "1 0 obj") {
		t.Error("Output should contain object 1")
	}
	if !strings.Contains(output, "2 0 obj") {
		t.Error("Output should contain object 2")
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
	if !strings.Contains(output, "%%EOF") {
		t.Error("Output should contain end-of-file marker")
	}
}

// TestXrefTableFormat はxrefテーブルのフォーマットをテストする
func TestXrefTableFormat(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)

	// ヘッダーとオブジェクトを書く
	w.WriteHeader()
	w.AddObject(core.Dictionary{
		core.Name("Type"): core.Name("Catalog"),
	})

	// Trailer
	trailer := core.Dictionary{
		core.Name("Size"): core.Integer(2),
		core.Name("Root"): &core.Reference{ObjectNumber: 1, GenerationNumber: 0},
	}
	w.WriteTrailer(trailer)

	output := buf.String()

	// xrefセクションを抽出
	xrefStart := strings.Index(output, "xref")
	if xrefStart == -1 {
		t.Fatal("xref not found")
	}

	xrefSection := output[xrefStart:]

	// xrefのフォーマットを確認
	if !strings.Contains(xrefSection, "0 2") {
		t.Error("xref should contain '0 2' (starting from 0, 2 entries)")
	}

	// 最初のエントリは常にfree
	if !strings.Contains(xrefSection, "0000000000 65535 f") {
		t.Error("xref should contain free entry '0000000000 65535 f'")
	}
}

// TestObjectOffsets はオブジェクトのオフセット計算をテストする
func TestObjectOffsets(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)

	w.WriteHeader()

	// オブジェクト1のオフセットを記録
	offset1 := buf.Len()
	w.AddObject(core.Dictionary{
		core.Name("Type"): core.Name("Catalog"),
	})

	// オブジェクト2のオフセットを記録
	offset2 := buf.Len()
	w.AddObject(core.Dictionary{
		core.Name("Type"): core.Name("Pages"),
	})

	// オフセットが正しく記録されているか確認
	if len(w.offsets) != 2 {
		t.Errorf("Expected 2 offsets, got %d", len(w.offsets))
	}

	if w.offsets[1] != int64(offset1) {
		t.Errorf("Offset for object 1 = %d, want %d", w.offsets[1], offset1)
	}

	if w.offsets[2] != int64(offset2) {
		t.Errorf("Offset for object 2 = %d, want %d", w.offsets[2], offset2)
	}
}
