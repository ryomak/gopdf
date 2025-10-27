package gopdf

import (
	"fmt"
	"io"
	"os"

	"github.com/ryomak/gopdf/internal/image/jpeg"
)

// Image represents an image that can be embedded in a PDF
type Image struct {
	Width            int
	Height           int
	Data             []byte
	ColorSpace       string
	BitsPerComponent int
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
