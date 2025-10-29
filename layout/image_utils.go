package layout

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"os"
)

// SaveImage は画像をファイルに保存する
func (img *ImageInfo) SaveImage(filename string) error {
	// ファイルを作成
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// データを書き込み
	_, err = file.Write(img.Data)
	if err != nil {
		return fmt.Errorf("failed to write image data: %w", err)
	}

	return nil
}

// ToImage は画像をimage.Imageに変換する
func (img *ImageInfo) ToImage() (image.Image, error) {
	switch img.Format {
	case ImageFormatJPEG:
		return jpeg.Decode(bytes.NewReader(img.Data))
	case ImageFormatPNG:
		// FlateDecode画像をデコード
		return decodeFlateImage(img)
	default:
		return nil, fmt.Errorf("unsupported image format: %s", img.Format)
	}
}

// decodeFlateImage はFlateDecode圧縮された画像データをimage.Imageに変換する
func decodeFlateImage(img *ImageInfo) (image.Image, error) {
	// Zlibで展開
	r, err := zlib.NewReader(bytes.NewReader(img.Data))
	if err != nil {
		return nil, fmt.Errorf("failed to create zlib reader: %w", err)
	}
	defer r.Close()

	rawData, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress image data: %w", err)
	}

	// 色空間に応じて画像を構築
	switch img.ColorSpace {
	case "DeviceRGB", "/DeviceRGB":
		return decodeRGBImage(rawData, img.Width, img.Height, img.BitsPerComp)
	case "DeviceGray", "/DeviceGray":
		return decodeGrayImage(rawData, img.Width, img.Height, img.BitsPerComp)
	case "DeviceCMYK", "/DeviceCMYK":
		return decodeCMYKImage(rawData, img.Width, img.Height, img.BitsPerComp)
	default:
		return nil, fmt.Errorf("unsupported color space: %s", img.ColorSpace)
	}
}

// decodeRGBImage はRGBピクセルデータからimage.Imageを構築する
func decodeRGBImage(data []byte, width, height, bitsPerComp int) (image.Image, error) {
	if bitsPerComp != 8 {
		return nil, fmt.Errorf("unsupported bits per component for RGB: %d", bitsPerComp)
	}

	expectedSize := width * height * 3
	if len(data) < expectedSize {
		return nil, fmt.Errorf("insufficient RGB data: got %d bytes, expected %d", len(data), expectedSize)
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			offset := (y*width + x) * 3
			img.Set(x, y, color.RGBA{
				R: data[offset],
				G: data[offset+1],
				B: data[offset+2],
				A: 255,
			})
		}
	}

	return img, nil
}

// decodeGrayImage はグレースケールピクセルデータからimage.Imageを構築する
func decodeGrayImage(data []byte, width, height, bitsPerComp int) (image.Image, error) {
	if bitsPerComp != 8 {
		return nil, fmt.Errorf("unsupported bits per component for Gray: %d", bitsPerComp)
	}

	expectedSize := width * height
	if len(data) < expectedSize {
		return nil, fmt.Errorf("insufficient Gray data: got %d bytes, expected %d", len(data), expectedSize)
	}

	img := image.NewGray(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			offset := y*width + x
			img.Set(x, y, color.Gray{Y: data[offset]})
		}
	}

	return img, nil
}

// decodeCMYKImage はCMYKピクセルデータからimage.Imageを構築する
func decodeCMYKImage(data []byte, width, height, bitsPerComp int) (image.Image, error) {
	if bitsPerComp != 8 {
		return nil, fmt.Errorf("unsupported bits per component for CMYK: %d", bitsPerComp)
	}

	expectedSize := width * height * 4
	if len(data) < expectedSize {
		return nil, fmt.Errorf("insufficient CMYK data: got %d bytes, expected %d", len(data), expectedSize)
	}

	// CMYKからRGBAに変換
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			offset := (y*width + x) * 4
			c := float64(data[offset]) / 255.0
			m := float64(data[offset+1]) / 255.0
			yy := float64(data[offset+2]) / 255.0
			k := float64(data[offset+3]) / 255.0

			// CMYK to RGB conversion
			r := 255 * (1 - c) * (1 - k)
			g := 255 * (1 - m) * (1 - k)
			b := 255 * (1 - yy) * (1 - k)

			img.Set(x, y, color.RGBA{
				R: uint8(r),
				G: uint8(g),
				B: uint8(b),
				A: 255,
			})
		}
	}

	return img, nil
}
