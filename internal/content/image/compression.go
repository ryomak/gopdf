// Package image provides image processing utilities for PDF.
package image

import (
	"bytes"
	"compress/zlib"
)

// CompressWithZlib compresses data using Zlib/Deflate compression.
func CompressWithZlib(data []byte) ([]byte, error) {
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
