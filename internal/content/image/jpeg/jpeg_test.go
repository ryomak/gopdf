package jpeg

import (
	"bytes"
	"testing"
)

// TestDecodeJPEGInfo はJPEG画像情報のデコードをテストする
func TestDecodeJPEGInfo(t *testing.T) {
	tests := []struct {
		name           string
		jpegData       []byte
		expectedWidth  int
		expectedHeight int
		expectedColors int
		expectedBits   int
		expectError    bool
	}{
		{
			name: "RGB JPEG 640x480",
			// 最小限のJPEGヘッダー: SOI + SOF0 + 画像情報
			jpegData: []byte{
				0xFF, 0xD8, // SOI
				0xFF, 0xC0, // SOF0 (Baseline DCT)
				0x00, 0x11, // Length: 17 bytes
				0x08,       // Bits per component: 8
				0x01, 0xE0, // Height: 480
				0x02, 0x80, // Width: 640
				0x03, // Number of components: 3 (RGB)
				// Component 1 (Y)
				0x01, 0x22, 0x00,
				// Component 2 (Cb)
				0x02, 0x11, 0x01,
				// Component 3 (Cr)
				0x03, 0x11, 0x01,
				0xFF, 0xD9, // EOI
			},
			expectedWidth:  640,
			expectedHeight: 480,
			expectedColors: 3,
			expectedBits:   8,
			expectError:    false,
		},
		{
			name: "Grayscale JPEG 100x100",
			jpegData: []byte{
				0xFF, 0xD8, // SOI
				0xFF, 0xC0, // SOF0
				0x00, 0x0B, // Length: 11 bytes
				0x08,       // Bits per component: 8
				0x00, 0x64, // Height: 100
				0x00, 0x64, // Width: 100
				0x01, // Number of components: 1 (Grayscale)
				// Component 1
				0x01, 0x11, 0x00,
				0xFF, 0xD9, // EOI
			},
			expectedWidth:  100,
			expectedHeight: 100,
			expectedColors: 1,
			expectedBits:   8,
			expectError:    false,
		},
		{
			name: "Invalid JPEG - missing SOI",
			jpegData: []byte{
				0xFF, 0xC0, // SOF0 (no SOI)
				0x00, 0x11,
			},
			expectError: true,
		},
		{
			name: "Invalid JPEG - no SOF",
			jpegData: []byte{
				0xFF, 0xD8, // SOI
				0xFF, 0xD9, // EOI (no SOF)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader(tt.jpegData)
			info, err := DecodeInfo(reader)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if info.Width != tt.expectedWidth {
				t.Errorf("Width = %d, want %d", info.Width, tt.expectedWidth)
			}
			if info.Height != tt.expectedHeight {
				t.Errorf("Height = %d, want %d", info.Height, tt.expectedHeight)
			}
			if info.ColorComponents != tt.expectedColors {
				t.Errorf("ColorComponents = %d, want %d", info.ColorComponents, tt.expectedColors)
			}
			if info.BitsPerComponent != tt.expectedBits {
				t.Errorf("BitsPerComponent = %d, want %d", info.BitsPerComponent, tt.expectedBits)
			}
		})
	}
}

// TestGetColorSpace はカラースペースの判定をテストする
func TestGetColorSpace(t *testing.T) {
	tests := []struct {
		name           string
		colorComponents int
		expected       string
	}{
		{"Grayscale", 1, "DeviceGray"},
		{"RGB", 3, "DeviceRGB"},
		{"CMYK", 4, "DeviceCMYK"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &Info{ColorComponents: tt.colorComponents}
			colorSpace := info.GetColorSpace()
			if colorSpace != tt.expected {
				t.Errorf("GetColorSpace() = %s, want %s", colorSpace, tt.expected)
			}
		})
	}
}
