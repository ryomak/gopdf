package jpeg

import (
	"fmt"
	"io"
)

// Info represents JPEG image information
type Info struct {
	Width            int
	Height           int
	ColorComponents  int // 1=Gray, 3=RGB, 4=CMYK
	BitsPerComponent int
}

// GetColorSpace returns the PDF color space name based on color components
func (i *Info) GetColorSpace() string {
	switch i.ColorComponents {
	case 1:
		return "DeviceGray"
	case 3:
		return "DeviceRGB"
	case 4:
		return "DeviceCMYK"
	default:
		return "DeviceRGB"
	}
}

// JPEG markers
const (
	markerSOI  = 0xD8 // Start of Image
	markerEOI  = 0xD9 // End of Image
	markerSOS  = 0xDA // Start of Scan
	markerSOF0 = 0xC0 // Start of Frame (Baseline DCT)
	markerSOF2 = 0xC2 // Start of Frame (Progressive DCT)
)

// DecodeInfo reads JPEG image information from a reader
// It extracts width, height, color components, and bits per component
func DecodeInfo(r io.Reader) (*Info, error) {
	// Read SOI marker
	marker, err := readMarker(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read SOI marker: %w", err)
	}
	if marker != markerSOI {
		return nil, fmt.Errorf("invalid JPEG: expected SOI marker (0xFF 0xD8), got 0xFF 0x%02X", marker)
	}

	// Scan for SOF marker
	for {
		marker, err := readMarker(r)
		if err != nil {
			return nil, fmt.Errorf("failed to read marker: %w", err)
		}

		// Check if this is a SOF marker
		if marker == markerSOF0 || marker == markerSOF2 {
			return decodeSOF(r)
		}

		// If it's EOI or SOS, we've gone too far without finding SOF
		if marker == markerEOI || marker == markerSOS {
			return nil, fmt.Errorf("no SOF marker found in JPEG")
		}

		// Skip this segment
		if err := skipSegment(r); err != nil {
			return nil, fmt.Errorf("failed to skip segment: %w", err)
		}
	}
}

// readMarker reads a JPEG marker (0xFF followed by marker byte)
func readMarker(r io.Reader) (byte, error) {
	buf := make([]byte, 2)
	if _, err := io.ReadFull(r, buf); err != nil {
		return 0, err
	}
	if buf[0] != 0xFF {
		return 0, fmt.Errorf("expected marker prefix 0xFF, got 0x%02X", buf[0])
	}
	return buf[1], nil
}

// skipSegment skips the current JPEG segment
func skipSegment(r io.Reader) error {
	// Read segment length (2 bytes, big-endian)
	buf := make([]byte, 2)
	if _, err := io.ReadFull(r, buf); err != nil {
		return err
	}
	length := int(buf[0])<<8 | int(buf[1])

	// Skip the rest of the segment (length includes the 2 bytes we just read)
	if length < 2 {
		return fmt.Errorf("invalid segment length: %d", length)
	}
	toSkip := length - 2

	// Read and discard
	_, err := io.CopyN(io.Discard, r, int64(toSkip))
	return err
}

// decodeSOF decodes a Start of Frame segment
func decodeSOF(r io.Reader) (*Info, error) {
	// Read segment length
	buf := make([]byte, 2)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}
	length := int(buf[0])<<8 | int(buf[1])

	// Read SOF data
	if length < 8 {
		return nil, fmt.Errorf("SOF segment too short: %d bytes", length)
	}

	data := make([]byte, length-2)
	if _, err := io.ReadFull(r, data); err != nil {
		return nil, err
	}

	// Parse SOF structure:
	// 0: bits per component
	// 1-2: height (big-endian)
	// 3-4: width (big-endian)
	// 5: number of components
	if len(data) < 6 {
		return nil, fmt.Errorf("SOF data too short")
	}

	info := &Info{
		BitsPerComponent: int(data[0]),
		Height:           int(data[1])<<8 | int(data[2]),
		Width:            int(data[3])<<8 | int(data[4]),
		ColorComponents:  int(data[5]),
	}

	return info, nil
}
