package image

import (
	"bytes"
	"compress/zlib"
	"io"
	"testing"
)

func TestCompressWithZlib(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "empty data",
			data:    []byte{},
			wantErr: false,
		},
		{
			name:    "small data",
			data:    []byte("Hello, World!"),
			wantErr: false,
		},
		{
			name:    "binary data",
			data:    []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD},
			wantErr: false,
		},
		{
			name:    "repetitive data",
			data:    bytes.Repeat([]byte("ABCD"), 1000),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compressed, err := CompressWithZlib(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("CompressWithZlib() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Verify we can decompress the data
			r, err := zlib.NewReader(bytes.NewReader(compressed))
			if err != nil {
				t.Errorf("Failed to create zlib reader: %v", err)
				return
			}
			defer r.Close()

			decompressed, err := io.ReadAll(r)
			if err != nil {
				t.Errorf("Failed to decompress: %v", err)
				return
			}

			if !bytes.Equal(decompressed, tt.data) {
				t.Errorf("Decompressed data doesn't match original")
			}
		})
	}
}

func TestCompressWithZlib_Compression(t *testing.T) {
	// Test that repetitive data is actually compressed
	data := bytes.Repeat([]byte("ABCDEFGH"), 10000)

	compressed, err := CompressWithZlib(data)
	if err != nil {
		t.Fatalf("CompressWithZlib() error = %v", err)
	}

	// Compressed data should be smaller than original
	if len(compressed) >= len(data) {
		t.Errorf("Compressed size (%d) should be less than original (%d)", len(compressed), len(data))
	}
}
