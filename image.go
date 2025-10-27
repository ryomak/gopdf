package gopdf

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"os"

	"github.com/ryomak/gopdf/internal/image/jpeg"
	"github.com/ryomak/gopdf/internal/image/png"
)

// Image represents an image that can be embedded in a PDF
type Image struct {
	Width            int
	Height           int
	Data             []byte
	ColorSpace       string
	BitsPerComponent int
	Filter           string  // "DCTDecode" for JPEG, "FlateDecode" for PNG
	SMask            *Image  // Soft mask (alpha channel) for transparency
}

// LoadJPEG loads a JPEG image from a reader
// It parses the JPEG header to extract image information and reads the entire image data
func LoadJPEG(r io.Reader) (*Image, error) {
	// Read all data into memory
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read image data: %w", err)
	}

	// Parse JPEG header from a new reader
	info, err := jpeg.DecodeInfo(io.NopCloser(newBytesReader(data)))
	if err != nil {
		return nil, fmt.Errorf("failed to decode JPEG info: %w", err)
	}

	return &Image{
		Width:            info.Width,
		Height:           info.Height,
		Data:             data,
		ColorSpace:       info.GetColorSpace(),
		BitsPerComponent: info.BitsPerComponent,
		Filter:           "DCTDecode",
	}, nil
}

// LoadJPEGFile loads a JPEG image from a file path
func LoadJPEGFile(path string) (*Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open image file: %w", err)
	}
	defer file.Close()

	return LoadJPEG(file)
}

// LoadPNG loads a PNG image from a reader
// It decodes the PNG and re-encodes pixel data with FlateDecode
func LoadPNG(r io.Reader) (*Image, error) {
	// Extract pixel data from PNG
	pixelData, err := png.ExtractPixelData(r)
	if err != nil {
		return nil, fmt.Errorf("failed to extract PNG pixel data: %w", err)
	}

	// Determine color space and compress data
	var colorSpace string
	var compressedData []byte
	var smask *Image

	if len(pixelData.RGBData) > 0 {
		// RGB or RGBA image
		colorSpace = "DeviceRGB"
		compressedData, err = compressWithZlib(pixelData.RGBData)
		if err != nil {
			return nil, fmt.Errorf("failed to compress RGB data: %w", err)
		}

		// Check for alpha channel
		if len(pixelData.AlphaData) > 0 {
			// Create SMask (soft mask) for alpha channel
			alphaCompressed, err := compressWithZlib(pixelData.AlphaData)
			if err != nil {
				return nil, fmt.Errorf("failed to compress alpha data: %w", err)
			}

			smask = &Image{
				Width:            pixelData.Width,
				Height:           pixelData.Height,
				Data:             alphaCompressed,
				ColorSpace:       "DeviceGray",
				BitsPerComponent: 8,
				Filter:           "FlateDecode",
			}
		}
	} else if len(pixelData.GrayData) > 0 {
		// Grayscale image
		colorSpace = "DeviceGray"
		compressedData, err = compressWithZlib(pixelData.GrayData)
		if err != nil {
			return nil, fmt.Errorf("failed to compress grayscale data: %w", err)
		}
	} else {
		return nil, fmt.Errorf("no pixel data extracted from PNG")
	}

	return &Image{
		Width:            pixelData.Width,
		Height:           pixelData.Height,
		Data:             compressedData,
		ColorSpace:       colorSpace,
		BitsPerComponent: 8,
		Filter:           "FlateDecode",
		SMask:            smask,
	}, nil
}

// LoadPNGFile loads a PNG image from a file path
func LoadPNGFile(path string) (*Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open image file: %w", err)
	}
	defer file.Close()

	return LoadPNG(file)
}

// compressWithZlib compresses data using Zlib/Deflate compression
func compressWithZlib(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)

	if _, err := w.Write(data); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// bytesReader wraps a byte slice to implement io.Reader
type bytesReader struct {
	data []byte
	pos  int
}

func newBytesReader(data []byte) *bytesReader {
	return &bytesReader{data: data}
}

func (r *bytesReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}
