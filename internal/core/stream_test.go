package core

import (
	"bytes"
	"testing"
)

// TestStream はStream型の振る舞いをテストする
func TestStream(t *testing.T) {
	t.Run("empty stream", func(t *testing.T) {
		stream := &Stream{
			Dict: Dictionary{
				Name("Length"): Integer(0),
			},
			Data: []byte{},
		}
		if len(stream.Data) != 0 {
			t.Errorf("Empty stream data length = %d, want 0", len(stream.Data))
		}
		if stream.Dict[Name("Length")] != Integer(0) {
			t.Errorf("Stream Length = %v, want Integer(0)", stream.Dict[Name("Length")])
		}
	})

	t.Run("stream with data", func(t *testing.T) {
		data := []byte("Hello, World!")
		stream := &Stream{
			Dict: Dictionary{
				Name("Length"): Integer(len(data)),
			},
			Data: data,
		}
		if len(stream.Data) != 13 {
			t.Errorf("Stream data length = %d, want 13", len(stream.Data))
		}
		if !bytes.Equal(stream.Data, data) {
			t.Errorf("Stream data = %s, want %s", stream.Data, data)
		}
	})

	t.Run("content stream", func(t *testing.T) {
		// PDFコンテンツストリームの例
		content := `BT
/F1 12 Tf
100 700 Td
(Hello, World!) Tj
ET
`
		stream := &Stream{
			Dict: Dictionary{
				Name("Length"): Integer(len(content)),
			},
			Data: []byte(content),
		}
		if stream.Dict[Name("Length")] != Integer(len(content)) {
			t.Errorf("Content stream Length = %v, want Integer(%d)", stream.Dict[Name("Length")], len(content))
		}
		if string(stream.Data) != content {
			t.Errorf("Content stream data mismatch")
		}
	})

	t.Run("compressed stream", func(t *testing.T) {
		// 圧縮されたストリームの例（メタデータのみテスト）
		compressedData := []byte{0x78, 0x9c, 0x01, 0x02, 0x03}
		stream := &Stream{
			Dict: Dictionary{
				Name("Length"): Integer(len(compressedData)),
				Name("Filter"): Name("FlateDecode"),
			},
			Data: compressedData,
		}
		if stream.Dict[Name("Filter")] != Name("FlateDecode") {
			t.Errorf("Stream Filter = %v, want Name(FlateDecode)", stream.Dict[Name("Filter")])
		}
		if len(stream.Data) != 5 {
			t.Errorf("Compressed stream data length = %d, want 5", len(stream.Data))
		}
	})

	t.Run("image stream", func(t *testing.T) {
		// 画像ストリームの例（メタデータのみ）
		stream := &Stream{
			Dict: Dictionary{
				Name("Type"):             Name("XObject"),
				Name("Subtype"):          Name("Image"),
				Name("Width"):            Integer(100),
				Name("Height"):           Integer(100),
				Name("ColorSpace"):       Name("DeviceRGB"),
				Name("BitsPerComponent"): Integer(8),
				Name("Length"):           Integer(30000),
				Name("Filter"):           Name("DCTDecode"),
			},
			Data: make([]byte, 30000),
		}
		if stream.Dict[Name("Subtype")] != Name("Image") {
			t.Errorf("Image stream Subtype = %v, want Name(Image)", stream.Dict[Name("Subtype")])
		}
		if stream.Dict[Name("Width")] != Integer(100) {
			t.Errorf("Image Width = %v, want Integer(100)", stream.Dict[Name("Width")])
		}
	})
}

// TestStreamModification はStreamの変更をテストする
func TestStreamModification(t *testing.T) {
	stream := &Stream{
		Dict: Dictionary{
			Name("Length"): Integer(0),
		},
		Data: []byte{},
	}

	// データの追加
	newData := []byte("Test data")
	stream.Data = newData
	stream.Dict[Name("Length")] = Integer(len(newData))

	if len(stream.Data) != 9 {
		t.Errorf("Stream data length after modification = %d, want 9", len(stream.Data))
	}
	if stream.Dict[Name("Length")] != Integer(9) {
		t.Errorf("Stream Length after modification = %v, want Integer(9)", stream.Dict[Name("Length")])
	}
}
