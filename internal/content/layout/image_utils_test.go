package layout

import (
	"bytes"
	"compress/zlib"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"testing"
)

func TestDecodeRGBImage(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		width       int
		height      int
		bitsPerComp int
		wantErr     bool
		validate    func(t *testing.T, img image.Image)
	}{
		{
			name:        "valid 2x2 RGB image",
			data:        []byte{255, 0, 0, 0, 255, 0, 0, 0, 255, 255, 255, 0},
			width:       2,
			height:      2,
			bitsPerComp: 8,
			wantErr:     false,
			validate: func(t *testing.T, img image.Image) {
				t.Helper()
				bounds := img.Bounds()
				if bounds.Dx() != 2 || bounds.Dy() != 2 {
					t.Errorf("bounds = %v, want 2x2", bounds)
				}
				// Check pixel (0,0) = red
				r, g, b, a := img.At(0, 0).RGBA()
				if r>>8 != 255 || g>>8 != 0 || b>>8 != 0 || a>>8 != 255 {
					t.Errorf("pixel (0,0) = (%d,%d,%d,%d), want (255,0,0,255)", r>>8, g>>8, b>>8, a>>8)
				}
			},
		},
		{
			name:        "valid 1x1 RGB image",
			data:        []byte{128, 64, 32},
			width:       1,
			height:      1,
			bitsPerComp: 8,
			wantErr:     false,
			validate: func(t *testing.T, img image.Image) {
				t.Helper()
				r, g, b, _ := img.At(0, 0).RGBA()
				if r>>8 != 128 || g>>8 != 64 || b>>8 != 32 {
					t.Errorf("pixel = (%d,%d,%d), want (128,64,32)", r>>8, g>>8, b>>8)
				}
			},
		},
		{
			name:        "unsupported bits per component",
			data:        []byte{0, 0, 0},
			width:       1,
			height:      1,
			bitsPerComp: 16,
			wantErr:     true,
		},
		{
			name:        "insufficient data",
			data:        []byte{255, 0},
			width:       1,
			height:      1,
			bitsPerComp: 8,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decodeRGBImage(tt.data, tt.width, tt.height, tt.bitsPerComp)
			if (err != nil) != tt.wantErr {
				t.Errorf("decodeRGBImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.validate != nil && err == nil {
				tt.validate(t, got)
			}
		})
	}
}

func TestDecodeGrayImage(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		width       int
		height      int
		bitsPerComp int
		wantErr     bool
		validate    func(t *testing.T, img image.Image)
	}{
		{
			name:        "valid 2x2 gray image",
			data:        []byte{0, 128, 255, 64},
			width:       2,
			height:      2,
			bitsPerComp: 8,
			wantErr:     false,
			validate: func(t *testing.T, img image.Image) {
				t.Helper()
				bounds := img.Bounds()
				if bounds.Dx() != 2 || bounds.Dy() != 2 {
					t.Errorf("bounds = %v, want 2x2", bounds)
				}
				// Check pixel (0,0) = 0 (black)
				r, _, _, _ := img.At(0, 0).RGBA()
				if r>>8 != 0 {
					t.Errorf("pixel (0,0) gray = %d, want 0", r>>8)
				}
				// Check pixel (1,0) = 128
				r, _, _, _ = img.At(1, 0).RGBA()
				if r>>8 != 128 {
					t.Errorf("pixel (1,0) gray = %d, want 128", r>>8)
				}
			},
		},
		{
			name:        "valid 1x1 gray image",
			data:        []byte{200},
			width:       1,
			height:      1,
			bitsPerComp: 8,
			wantErr:     false,
		},
		{
			name:        "unsupported bits per component",
			data:        []byte{0},
			width:       1,
			height:      1,
			bitsPerComp: 4,
			wantErr:     true,
		},
		{
			name:        "insufficient data",
			data:        []byte{},
			width:       1,
			height:      1,
			bitsPerComp: 8,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decodeGrayImage(tt.data, tt.width, tt.height, tt.bitsPerComp)
			if (err != nil) != tt.wantErr {
				t.Errorf("decodeGrayImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.validate != nil && err == nil {
				tt.validate(t, got)
			}
		})
	}
}

func TestDecodeCMYKImage(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		width       int
		height      int
		bitsPerComp int
		wantErr     bool
		validate    func(t *testing.T, img image.Image)
	}{
		{
			name:        "valid 1x1 CMYK image pure white (0,0,0,0)",
			data:        []byte{0, 0, 0, 0},
			width:       1,
			height:      1,
			bitsPerComp: 8,
			wantErr:     false,
			validate: func(t *testing.T, img image.Image) {
				t.Helper()
				r, g, b, _ := img.At(0, 0).RGBA()
				// CMYK(0,0,0,0) -> RGB(255,255,255)
				if r>>8 != 255 || g>>8 != 255 || b>>8 != 255 {
					t.Errorf("pixel = (%d,%d,%d), want (255,255,255)", r>>8, g>>8, b>>8)
				}
			},
		},
		{
			name:        "valid 1x1 CMYK image pure black (0,0,0,255)",
			data:        []byte{0, 0, 0, 255},
			width:       1,
			height:      1,
			bitsPerComp: 8,
			wantErr:     false,
			validate: func(t *testing.T, img image.Image) {
				t.Helper()
				r, g, b, _ := img.At(0, 0).RGBA()
				// CMYK(0,0,0,1) -> RGB(0,0,0)
				if r>>8 != 0 || g>>8 != 0 || b>>8 != 0 {
					t.Errorf("pixel = (%d,%d,%d), want (0,0,0)", r>>8, g>>8, b>>8)
				}
			},
		},
		{
			name:        "valid 2x1 CMYK image",
			data:        []byte{0, 0, 0, 0, 255, 0, 0, 0},
			width:       2,
			height:      1,
			bitsPerComp: 8,
			wantErr:     false,
			validate: func(t *testing.T, img image.Image) {
				t.Helper()
				bounds := img.Bounds()
				if bounds.Dx() != 2 || bounds.Dy() != 1 {
					t.Errorf("bounds = %v, want 2x1", bounds)
				}
			},
		},
		{
			name:        "unsupported bits per component",
			data:        []byte{0, 0, 0, 0},
			width:       1,
			height:      1,
			bitsPerComp: 16,
			wantErr:     true,
		},
		{
			name:        "insufficient data",
			data:        []byte{0, 0, 0},
			width:       1,
			height:      1,
			bitsPerComp: 8,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decodeCMYKImage(tt.data, tt.width, tt.height, tt.bitsPerComp)
			if (err != nil) != tt.wantErr {
				t.Errorf("decodeCMYKImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.validate != nil && err == nil {
				tt.validate(t, got)
			}
		})
	}
}

func TestDecodeFlateImage(t *testing.T) {
	// Helper to compress data with zlib
	zlibCompress := func(data []byte) []byte {
		var buf bytes.Buffer
		w := zlib.NewWriter(&buf)
		_, _ = w.Write(data)
		_ = w.Close()
		return buf.Bytes()
	}

	tests := []struct {
		name    string
		img     *ImageInfo
		wantErr bool
	}{
		{
			name: "valid RGB flate image",
			img: &ImageInfo{
				Width:       2,
				Height:      2,
				BitsPerComp: 8,
				ColorSpace:  "DeviceRGB",
				Data:        zlibCompress([]byte{255, 0, 0, 0, 255, 0, 0, 0, 255, 255, 255, 0}),
			},
			wantErr: false,
		},
		{
			name: "valid Gray flate image",
			img: &ImageInfo{
				Width:       2,
				Height:      2,
				BitsPerComp: 8,
				ColorSpace:  "DeviceGray",
				Data:        zlibCompress([]byte{0, 128, 255, 64}),
			},
			wantErr: false,
		},
		{
			name: "valid CMYK flate image",
			img: &ImageInfo{
				Width:       1,
				Height:      1,
				BitsPerComp: 8,
				ColorSpace:  "DeviceCMYK",
				Data:        zlibCompress([]byte{0, 0, 0, 0}),
			},
			wantErr: false,
		},
		{
			name: "slash prefix color space DeviceRGB",
			img: &ImageInfo{
				Width:       1,
				Height:      1,
				BitsPerComp: 8,
				ColorSpace:  "/DeviceRGB",
				Data:        zlibCompress([]byte{128, 64, 32}),
			},
			wantErr: false,
		},
		{
			name: "slash prefix color space DeviceGray",
			img: &ImageInfo{
				Width:       1,
				Height:      1,
				BitsPerComp: 8,
				ColorSpace:  "/DeviceGray",
				Data:        zlibCompress([]byte{128}),
			},
			wantErr: false,
		},
		{
			name: "slash prefix color space DeviceCMYK",
			img: &ImageInfo{
				Width:       1,
				Height:      1,
				BitsPerComp: 8,
				ColorSpace:  "/DeviceCMYK",
				Data:        zlibCompress([]byte{0, 0, 0, 0}),
			},
			wantErr: false,
		},
		{
			name: "unsupported color space",
			img: &ImageInfo{
				Width:       1,
				Height:      1,
				BitsPerComp: 8,
				ColorSpace:  "ICCBased",
				Data:        zlibCompress([]byte{0}),
			},
			wantErr: true,
		},
		{
			name: "invalid zlib data",
			img: &ImageInfo{
				Width:       1,
				Height:      1,
				BitsPerComp: 8,
				ColorSpace:  "DeviceRGB",
				Data:        []byte{0xFF, 0xFE, 0xFD}, // not valid zlib
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decodeFlateImage(tt.img)
			if (err != nil) != tt.wantErr {
				t.Errorf("decodeFlateImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && got == nil {
				t.Error("decodeFlateImage() returned nil image without error")
			}
		})
	}
}

func TestToImage(t *testing.T) {
	// Create a small JPEG image in memory
	makeJPEGData := func() []byte {
		img := image.NewRGBA(image.Rect(0, 0, 2, 2))
		img.Set(0, 0, color.RGBA{R: 255, A: 255})
		img.Set(1, 0, color.RGBA{G: 255, A: 255})
		img.Set(0, 1, color.RGBA{B: 255, A: 255})
		img.Set(1, 1, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		var buf bytes.Buffer
		_ = jpeg.Encode(&buf, img, nil)
		return buf.Bytes()
	}

	// Create zlib-compressed RGB data for PNG format
	zlibCompress := func(data []byte) []byte {
		var buf bytes.Buffer
		w := zlib.NewWriter(&buf)
		_, _ = w.Write(data)
		_ = w.Close()
		return buf.Bytes()
	}

	tests := []struct {
		name    string
		img     *ImageInfo
		wantErr bool
	}{
		{
			name: "JPEG format",
			img: &ImageInfo{
				Format: ImageFormatJPEG,
				Data:   makeJPEGData(),
				Width:  2,
				Height: 2,
			},
			wantErr: false,
		},
		{
			name: "PNG format (FlateDecode)",
			img: &ImageInfo{
				Format:      ImageFormatPNG,
				Data:        zlibCompress([]byte{255, 0, 0, 0, 255, 0, 0, 0, 255, 255, 255, 0}),
				Width:       2,
				Height:      2,
				BitsPerComp: 8,
				ColorSpace:  "DeviceRGB",
			},
			wantErr: false,
		},
		{
			name: "unknown format returns error",
			img: &ImageInfo{
				Format: ImageFormatUnknown,
				Data:   []byte{0, 0, 0},
			},
			wantErr: true,
		},
		{
			name: "invalid JPEG data",
			img: &ImageInfo{
				Format: ImageFormatJPEG,
				Data:   []byte{0xFF, 0xD8, 0xFF, 0x00}, // truncated JPEG
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.img.ToImage()
			if (err != nil) != tt.wantErr {
				t.Errorf("ToImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && got == nil {
				t.Error("ToImage() returned nil image without error")
			}
		})
	}
}

func TestSaveImage(t *testing.T) {
	// Create a temp file path
	tmpFile := t.TempDir() + "/test_image.bin"

	tests := []struct {
		name     string
		img      *ImageInfo
		filename string
		wantErr  bool
	}{
		{
			name: "save valid image data",
			img: &ImageInfo{
				Data: []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
			},
			filename: tmpFile,
			wantErr:  false,
		},
		{
			name: "save to invalid path",
			img: &ImageInfo{
				Data: []byte{1, 2, 3},
			},
			filename: "/nonexistent/dir/file.bin",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.img.SaveImage(tt.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("SaveImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				// Verify file was created and has correct content
				data, readErr := os.ReadFile(tt.filename)
				if readErr != nil {
					t.Fatalf("failed to read saved file: %v", readErr)
				}
				if !bytes.Equal(data, tt.img.Data) {
					t.Errorf("saved data = %v, want %v", data, tt.img.Data)
				}
			}
		})
	}
}
