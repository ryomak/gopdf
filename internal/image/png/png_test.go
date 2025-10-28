package png

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

// createTestPNG creates a simple test PNG image
func createTestPNG(width, height int, colorModel color.Model) ([]byte, image.Image) {
	var img image.Image

	switch colorModel {
	case color.RGBAModel:
		rgba := image.NewRGBA(image.Rect(0, 0, width, height))
		// Create a simple gradient
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				r := uint8(float64(x) / float64(width) * 255)
				g := uint8(float64(y) / float64(height) * 255)
				b := uint8(128)
				a := uint8(255)
				rgba.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: a})
			}
		}
		img = rgba

	case color.GrayModel:
		gray := image.NewGray(image.Rect(0, 0, width, height))
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				g := uint8(float64(x+y) / float64(width+height) * 255)
				gray.SetGray(x, y, color.Gray{Y: g})
			}
		}
		img = gray
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		panic(err) // テスト用のヘルパー関数なのでpanicで問題ない
	}
	return buf.Bytes(), img
}

// TestDecodeInfo はPNG画像情報のデコードをテストする
func TestDecodeInfo(t *testing.T) {
	tests := []struct {
		name           string
		width          int
		height         int
		colorModel     color.Model
		expectedWidth  int
		expectedHeight int
	}{
		{
			name:           "RGBA 100x100",
			width:          100,
			height:         100,
			colorModel:     color.RGBAModel,
			expectedWidth:  100,
			expectedHeight: 100,
		},
		{
			name:           "Grayscale 50x75",
			width:          50,
			height:         75,
			colorModel:     color.GrayModel,
			expectedWidth:  50,
			expectedHeight: 75,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pngData, _ := createTestPNG(tt.width, tt.height, tt.colorModel)
			reader := bytes.NewReader(pngData)

			info, err := DecodeInfo(reader)
			if err != nil {
				t.Fatalf("Failed to decode PNG info: %v", err)
			}

			if info.Width != tt.expectedWidth {
				t.Errorf("Width = %d, want %d", info.Width, tt.expectedWidth)
			}
			if info.Height != tt.expectedHeight {
				t.Errorf("Height = %d, want %d", info.Height, tt.expectedHeight)
			}
		})
	}
}

// TestExtractPixelData はPixel dataの抽出をテストする
func TestExtractPixelData(t *testing.T) {
	tests := []struct {
		name       string
		width      int
		height     int
		colorModel color.Model
	}{
		{"RGBA 10x10", 10, 10, color.RGBAModel},
		{"Gray 20x20", 20, 20, color.GrayModel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pngData, _ := createTestPNG(tt.width, tt.height, tt.colorModel)
			reader := bytes.NewReader(pngData)

			data, err := ExtractPixelData(reader)
			if err != nil {
				t.Fatalf("Failed to extract pixel data: %v", err)
			}

			if len(data.RGBData) == 0 && len(data.GrayData) == 0 {
				t.Error("Pixel data is empty")
			}

			// Check data size
			switch tt.colorModel {
			case color.RGBAModel:
				expectedSize := tt.width * tt.height * 3 // RGB without alpha
				if len(data.RGBData) != expectedSize {
					t.Errorf("RGB data size = %d, want %d", len(data.RGBData), expectedSize)
				}
				// Alpha channel should also be extracted
				expectedAlphaSize := tt.width * tt.height
				if len(data.AlphaData) != expectedAlphaSize {
					t.Errorf("Alpha data size = %d, want %d", len(data.AlphaData), expectedAlphaSize)
				}

			case color.GrayModel:
				expectedSize := tt.width * tt.height
				if len(data.GrayData) != expectedSize {
					t.Errorf("Gray data size = %d, want %d", len(data.GrayData), expectedSize)
				}
			}
		})
	}
}

// TestSeparateAlphaChannel はアルファチャンネル分離をテストする
func TestSeparateAlphaChannel(t *testing.T) {
	// Create RGBA image with varying alpha
	width, height := 10, 10
	rgba := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			alpha := uint8(float64(x) / float64(width) * 255)
			rgba.SetRGBA(x, y, color.RGBA{R: 255, G: 0, B: 0, A: alpha})
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, rgba); err != nil {
		t.Fatalf("Failed to encode PNG: %v", err)
	}

	reader := bytes.NewReader(buf.Bytes())
	data, err := ExtractPixelData(reader)
	if err != nil {
		t.Fatalf("Failed to extract pixel data: %v", err)
	}

	// Check that alpha channel is properly separated
	if len(data.AlphaData) != width*height {
		t.Errorf("Alpha data size = %d, want %d", len(data.AlphaData), width*height)
	}

	// Verify alpha values
	if data.AlphaData[0] != 0 { // First pixel should have alpha = 0
		t.Errorf("First alpha value = %d, want 0", data.AlphaData[0])
	}
	if data.AlphaData[width*height-1] < 200 { // Last pixel should have high alpha
		t.Errorf("Last alpha value = %d, want > 200", data.AlphaData[width*height-1])
	}
}
