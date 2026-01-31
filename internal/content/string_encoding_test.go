package content

import (
	"testing"
)

func TestDecodePDFString_ASCII(t *testing.T) {
	data := []byte("Hello World")
	result := decodePDFString(data)
	expected := "Hello World"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestDecodePDFString_UTF16BE(t *testing.T) {
	// UTF-16BE BOM + "こんにちは" in UTF-16BE
	data := []byte{
		0xFE, 0xFF, // BOM
		0x30, 0x53, // こ
		0x30, 0x93, // ん
		0x30, 0x6B, // に
		0x30, 0x61, // ち
		0x30, 0x6F, // は
	}

	result := decodePDFString(data)
	expected := "こんにちは"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestDecodePDFString_UTF16LE(t *testing.T) {
	// UTF-16LE BOM + "テスト" in UTF-16LE
	data := []byte{
		0xFF, 0xFE, // BOM
		0xC6, 0x30, // テ
		0xB9, 0x30, // ス
		0xC8, 0x30, // ト
	}

	result := decodePDFString(data)
	expected := "テスト"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestDecodePDFString_UTF8(t *testing.T) {
	// Valid UTF-8
	data := []byte("こんにちは")
	result := decodePDFString(data)
	expected := "こんにちは"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestDecodePDFString_PDFDocEncoding_Latin1(t *testing.T) {
	// Latin-1範囲（0x00-0x7F, 0xA0-0xFF）
	data := []byte{0x48, 0x65, 0x6C, 0x6C, 0x6F, 0xA0, 0xE9} // "Hello é"
	result := decodePDFString(data)

	// 少なくともASCII部分は正しい
	if result[:5] != "Hello" {
		t.Errorf("ASCII part incorrect: got %q", result[:5])
	}
}

func TestDecodePDFString_PDFDocEncoding_SpecialChars(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected rune
	}{
		{"BULLET", []byte{0x80}, 0x2022},
		{"DAGGER", []byte{0x81}, 0x2020},
		{"ELLIPSIS", []byte{0x83}, 0x2026},
		{"EM DASH", []byte{0x84}, 0x2014},
		{"EN DASH", []byte{0x85}, 0x2013},
		{"TRADE MARK", []byte{0x92}, 0x2122},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := decodePDFString(tt.input)

			// 文字数をカウント（バイト数ではなく）
			runes := []rune(result)
			if len(runes) != 1 {
				t.Fatalf("Expected 1 character, got %d (bytes: %d, string: %q)", len(runes), len(result), result)
			}

			got := runes[0]
			if got != tt.expected {
				t.Errorf("Expected U+%04X, got U+%04X", tt.expected, got)
			}
		})
	}
}

func TestDecodePDFString_Empty(t *testing.T) {
	result := decodePDFString([]byte{})
	if result != "" {
		t.Errorf("Expected empty string, got %q", result)
	}
}

func TestDecodeUTF16BE(t *testing.T) {
	// "世界" in UTF-16BE
	data := []byte{
		0x4E, 0x16, // 世
		0x75, 0x4C, // 界
	}

	result := decodeUTF16BE(data)
	expected := "世界"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestDecodeUTF16BE_OddLength(t *testing.T) {
	// 奇数バイトはエラー
	data := []byte{0x00, 0x41, 0x00}
	result := decodeUTF16BE(data)

	if result != "" {
		t.Errorf("Expected empty string for odd length, got %q", result)
	}
}

func TestDecodeUTF16LE(t *testing.T) {
	// "ABC" in UTF-16LE
	data := []byte{
		0x41, 0x00, // A
		0x42, 0x00, // B
		0x43, 0x00, // C
	}

	result := decodeUTF16LE(data)
	expected := "ABC"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestDecodePDFDocEncoding_AllRanges(t *testing.T) {
	// 各範囲のテスト
	tests := []struct {
		name  string
		input byte
	}{
		{"ASCII lower", 0x41},        // 'A'
		{"ASCII upper", 0x7E},        // '~'
		{"Special range", 0x80},      // BULLET
		{"Special range end", 0x9F},  // REPLACEMENT
		{"Latin-1 start", 0xA0},      // NO-BREAK SPACE
		{"Latin-1 end", 0xFF},        // ÿ
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := decodePDFDocEncoding([]byte{tt.input})
			if result == "" {
				t.Errorf("Expected non-empty result for byte 0x%02X", tt.input)
			}
		})
	}
}
