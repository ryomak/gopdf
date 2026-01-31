package png

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
)

// Info represents PNG image information
type Info struct {
	Width      int
	Height     int
	ColorModel color.Model
}

// PixelData represents extracted pixel data from PNG
type PixelData struct {
	Width     int
	Height    int
	RGBData   []byte // RGB data (3 bytes per pixel)
	GrayData  []byte // Grayscale data (1 byte per pixel)
	AlphaData []byte // Alpha channel (1 byte per pixel), if present
}

// DecodeInfo reads PNG image information from a reader
func DecodeInfo(r io.Reader) (*Info, error) {
	img, err := png.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("failed to decode PNG: %w", err)
	}

	bounds := img.Bounds()
	return &Info{
		Width:      bounds.Dx(),
		Height:     bounds.Dy(),
		ColorModel: img.ColorModel(),
	}, nil
}

// ExtractPixelData extracts pixel data from PNG image
// It separates RGB data and alpha channel for RGBA images
func ExtractPixelData(r io.Reader) (*PixelData, error) {
	img, err := png.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("failed to decode PNG: %w", err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	data := &PixelData{
		Width:  width,
		Height: height,
	}

	// Extract pixel data based on color model
	switch img.ColorModel() {
	case color.RGBAModel, color.NRGBA64Model:
		data.RGBData, data.AlphaData = extractRGBA(img, width, height)

	case color.GrayModel:
		data.GrayData = extractGray(img, width, height)

	case color.Gray16Model:
		data.GrayData = extractGray16(img, width, height)

	default:
		// Convert to RGBA for unsupported color models
		data.RGBData, data.AlphaData = extractRGBA(img, width, height)
	}

	return data, nil
}

// extractRGBA extracts RGB and alpha channel data separately
func extractRGBA(img image.Image, width, height int) ([]byte, []byte) {
	rgbData := make([]byte, width*height*3)
	alphaData := make([]byte, width*height)

	idx := 0
	alphaIdx := 0

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, a := img.At(x, y).RGBA()

			// Convert from 16-bit to 8-bit
			rgbData[idx] = uint8(r >> 8)
			rgbData[idx+1] = uint8(g >> 8)
			rgbData[idx+2] = uint8(b >> 8)
			idx += 3

			alphaData[alphaIdx] = uint8(a >> 8)
			alphaIdx++
		}
	}

	return rgbData, alphaData
}

// extractGray extracts grayscale data
func extractGray(img image.Image, width, height int) []byte {
	grayData := make([]byte, width*height)
	idx := 0

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := img.At(x, y)
			grayColor := color.GrayModel.Convert(c).(color.Gray)
			grayData[idx] = grayColor.Y
			idx++
		}
	}

	return grayData
}

// extractGray16 extracts 16-bit grayscale data and converts to 8-bit
func extractGray16(img image.Image, width, height int) []byte {
	grayData := make([]byte, width*height)
	idx := 0

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := img.At(x, y)
			grayColor := color.Gray16Model.Convert(c).(color.Gray16)
			// Convert 16-bit to 8-bit
			grayData[idx] = uint8(grayColor.Y >> 8)
			idx++
		}
	}

	return grayData
}
