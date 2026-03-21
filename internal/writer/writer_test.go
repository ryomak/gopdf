package writer

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ryomak/gopdf/internal/core"
	"github.com/ryomak/gopdf/internal/security"
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
	if err := w.WriteHeader(); err != nil {
		t.Fatalf("WriteHeader() failed: %v", err)
	}
	if _, err := w.AddObject(core.Dictionary{
		core.Name("Type"): core.Name("Catalog"),
	}); err != nil {
		t.Fatalf("AddObject() failed: %v", err)
	}

	// Trailer
	trailer := core.Dictionary{
		core.Name("Size"): core.Integer(2),
		core.Name("Root"): &core.Reference{ObjectNumber: 1, GenerationNumber: 0},
	}
	if err := w.WriteTrailer(trailer); err != nil {
		t.Fatalf("WriteTrailer() failed: %v", err)
	}

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

// TestSetEncryption はSetEncryptionメソッドをテストする
func TestSetEncryption(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)

	// Initially nil
	if w.encryption != nil {
		t.Error("encryption should be nil initially")
	}

	info, err := SetupEncryption("user", "owner", security.DefaultPermissions(), 40)
	if err != nil {
		t.Fatalf("SetupEncryption failed: %v", err)
	}

	w.SetEncryption(info)

	if w.encryption == nil {
		t.Error("encryption should not be nil after SetEncryption")
	}

	if w.encryption.KeyLength != 40 {
		t.Errorf("encryption.KeyLength = %d, want 40", w.encryption.KeyLength)
	}
}

// TestEncryptStream はencryptStreamメソッドをテストする
func TestEncryptStream(t *testing.T) {
	tests := []struct {
		name      string
		keyLength int
		data      []byte
	}{
		{
			"40-bit encryption",
			40,
			[]byte("Hello, World!"),
		},
		{
			"128-bit encryption",
			128,
			[]byte("Hello, World!"),
		},
		{
			"empty data",
			40,
			[]byte{},
		},
		{
			"binary data",
			128,
			[]byte{0x00, 0xFF, 0x0A, 0x0D, 0x42},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := SetupEncryption("user", "owner", security.DefaultPermissions(), tt.keyLength)
			if err != nil {
				t.Fatalf("SetupEncryption failed: %v", err)
			}

			var buf bytes.Buffer
			w := NewWriter(&buf)
			w.SetEncryption(info)

			origStream := &core.Stream{
				Dict: core.Dictionary{
					core.Name("Length"): core.Integer(len(tt.data)),
					core.Name("Type"):   core.Name("XObject"),
				},
				Data: tt.data,
			}

			result := w.encryptStream(origStream, 1, 0)

			// Result should be a new stream, not the same pointer
			if result == origStream {
				t.Error("encryptStream should return a new stream")
			}

			// Dict should be copied with updated Length
			if result.Dict[core.Name("Type")] == nil {
				t.Error("copied dict should preserve original keys")
			}

			lengthObj, ok := result.Dict[core.Name("Length")].(core.Integer)
			if !ok {
				t.Fatal("Length should be core.Integer")
			}
			if int(lengthObj) != len(result.Data) {
				t.Errorf("Length = %d, but data length = %d", int(lengthObj), len(result.Data))
			}

			// For non-empty data, encrypted data should differ from original
			if len(tt.data) > 0 && string(result.Data) == string(tt.data) {
				t.Error("encrypted data should differ from original data")
			}

			// Original stream should not be modified
			origLength, _ := origStream.Dict[core.Name("Length")].(core.Integer)
			if int(origLength) != len(tt.data) {
				t.Error("original stream should not be modified")
			}
		})
	}
}

// TestAddObjectWithEncryption はAddObjectが暗号化付きで動作するかテストする
func TestAddObjectWithEncryption(t *testing.T) {
	info, err := SetupEncryption("user", "owner", security.DefaultPermissions(), 40)
	if err != nil {
		t.Fatalf("SetupEncryption failed: %v", err)
	}

	var buf bytes.Buffer
	w := NewWriter(&buf)
	w.SetEncryption(info)

	if err := w.WriteHeader(); err != nil {
		t.Fatalf("WriteHeader failed: %v", err)
	}

	// Add a stream object - should be encrypted
	stream := &core.Stream{
		Dict: core.Dictionary{
			core.Name("Length"): core.Integer(5),
		},
		Data: []byte("hello"),
	}

	objNum, err := w.AddObject(stream)
	if err != nil {
		t.Fatalf("AddObject(stream) failed: %v", err)
	}

	if objNum != 1 {
		t.Errorf("object number = %d, want 1", objNum)
	}

	output := buf.String()
	// The original plaintext should NOT appear in the output (it's encrypted)
	// Extract content between "stream\n" and "\nendstream"
	streamIdx := strings.Index(output, "stream\n")
	endstreamIdx := strings.Index(output, "\nendstream")
	if streamIdx != -1 && endstreamIdx != -1 {
		dataSection := output[streamIdx+len("stream\n") : endstreamIdx]
		if dataSection == "hello" {
			t.Error("stream data should be encrypted, but plaintext found")
		}
	}

	// Non-stream objects should not be affected
	dict := core.Dictionary{
		core.Name("Type"): core.Name("Catalog"),
	}
	objNum2, err := w.AddObject(dict)
	if err != nil {
		t.Fatalf("AddObject(dict) failed: %v", err)
	}
	if objNum2 != 2 {
		t.Errorf("object number = %d, want 2", objNum2)
	}
}

// TestWriteTrailerWithEncryption は暗号化ありのトレーラー出力をテストする
func TestWriteTrailerWithEncryption(t *testing.T) {
	tests := []struct {
		name      string
		keyLength int
	}{
		{"40-bit encryption", 40},
		{"128-bit encryption", 128},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := SetupEncryption("user", "owner", security.DefaultPermissions(), tt.keyLength)
			if err != nil {
				t.Fatalf("SetupEncryption failed: %v", err)
			}

			var buf bytes.Buffer
			w := NewWriter(&buf)
			w.SetEncryption(info)

			if err := w.WriteHeader(); err != nil {
				t.Fatalf("WriteHeader failed: %v", err)
			}

			// Add a basic object
			_, err = w.AddObject(core.Dictionary{
				core.Name("Type"): core.Name("Catalog"),
			})
			if err != nil {
				t.Fatalf("AddObject failed: %v", err)
			}

			trailer := core.Dictionary{
				core.Name("Root"): &core.Reference{ObjectNumber: 1, GenerationNumber: 0},
			}

			if err := w.WriteTrailer(trailer); err != nil {
				t.Fatalf("WriteTrailer failed: %v", err)
			}

			output := buf.String()

			// Trailer should contain Encrypt reference
			if !strings.Contains(output, "/Encrypt") {
				t.Error("trailer should contain /Encrypt key")
			}

			// Trailer should contain file ID
			if !strings.Contains(output, "/ID") {
				t.Error("trailer should contain /ID key")
			}

			// Should contain xref, trailer, startxref, %%EOF
			if !strings.Contains(output, "xref") {
				t.Error("output should contain xref")
			}
			if !strings.Contains(output, "trailer") {
				t.Error("output should contain trailer")
			}
			if !strings.Contains(output, "startxref") {
				t.Error("output should contain startxref")
			}
			if !strings.Contains(output, "%%EOF") {
				t.Error("output should contain EOF marker")
			}

			// Encrypt dictionary should be added as an object
			// With 1 catalog obj + 1 encrypt obj = nextObjNum should be 3
			if !strings.Contains(output, "/Size 3") {
				t.Error("Size should be 3 (2 objects + 1 free entry)")
			}

			// Encrypt dict should contain Standard filter
			if !strings.Contains(output, "/Filter") {
				t.Error("Encrypt dictionary should contain /Filter")
			}
		})
	}
}

// TestObjectOffsets はオブジェクトのオフセット計算をテストする
func TestObjectOffsets(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)

	if err := w.WriteHeader(); err != nil {
		t.Fatalf("WriteHeader() failed: %v", err)
	}

	// オブジェクト1のオフセットを記録
	offset1 := buf.Len()
	if _, err := w.AddObject(core.Dictionary{
		core.Name("Type"): core.Name("Catalog"),
	}); err != nil {
		t.Fatalf("AddObject() failed: %v", err)
	}

	// オブジェクト2のオフセットを記録
	offset2 := buf.Len()
	if _, err := w.AddObject(core.Dictionary{
		core.Name("Type"): core.Name("Pages"),
	}); err != nil {
		t.Fatalf("AddObject() failed: %v", err)
	}

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
