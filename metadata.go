package gopdf

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/ryomak/gopdf/internal/core"
)

// Metadata represents PDF document metadata (Info dictionary)
// as defined in PDF 1.7 specification.
type Metadata struct {
	// Standard fields
	Title        string    // The document's title
	Author       string    // The name of the person who created the document
	Subject      string    // The subject of the document
	Keywords     string    // Keywords associated with the document
	Creator      string    // The application that created the original document
	Producer     string    // The application that converted the document to PDF
	CreationDate time.Time // The date and time the document was created
	ModDate      time.Time // The date and time the document was last modified

	// Custom fields (key-value pairs)
	// Any additional metadata fields not covered by standard fields
	Custom map[string]string
}

// SetMetadata sets the document metadata.
func (d *Document) SetMetadata(metadata Metadata) {
	d.metadata = &metadata
}

// GetMetadata returns the document metadata.
// Returns nil if no metadata is set.
func (d *Document) GetMetadata() *Metadata {
	return d.metadata
}

// formatPDFDate formats a time.Time to PDF date string.
// Format: D:YYYYMMDDHHmmSSOHH'mm'
// Example: D:20250129123045+09'00'
func formatPDFDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	// Get timezone offset
	_, offset := t.Zone()
	offsetHours := offset / 3600
	offsetMinutes := (offset % 3600) / 60

	// Format offset sign
	offsetSign := "+"
	if offsetHours < 0 {
		offsetSign = "-"
		offsetHours = -offsetHours
		offsetMinutes = -offsetMinutes
	}

	return fmt.Sprintf("D:%04d%02d%02d%02d%02d%02d%s%02d'%02d'",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second(),
		offsetSign, offsetHours, offsetMinutes)
}

// escapeString escapes special characters in PDF strings.
// Escapes: (, ), \
func escapeString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "(", "\\(")
	s = strings.ReplaceAll(s, ")", "\\)")
	return s
}

// isASCII checks if a string contains only ASCII characters
func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > 127 {
			return false
		}
	}
	return true
}

// encodeTextString encodes a string for PDF text string.
// Uses UTF-16BE with BOM for non-ASCII characters.
// Uses PDFDocEncoding (escaped) for ASCII characters.
func encodeTextString(s string) core.Object {
	if s == "" {
		return core.String("")
	}

	// If all ASCII, use simple string with escaping
	if isASCII(s) {
		return core.String("(" + escapeString(s) + ")")
	}

	// For non-ASCII, use UTF-16BE with BOM
	// BOM: FEFF
	runes := []rune(s)
	hexBytes := make([]byte, 0, (len(runes)+1)*4) // BOM + runes

	// Add BOM
	hexBytes = append(hexBytes, 'F', 'E', 'F', 'F')

	// Convert each rune to UTF-16BE
	for _, r := range runes {
		if r <= 0xFFFF {
			// BMP (Basic Multilingual Plane)
			hexBytes = append(hexBytes,
				hexChar(byte(r>>12)),
				hexChar(byte(r>>8)),
				hexChar(byte(r>>4)),
				hexChar(byte(r)),
			)
		} else {
			// Surrogate pair for characters outside BMP
			r -= 0x10000
			high := 0xD800 + (r >> 10)
			low := 0xDC00 + (r & 0x3FF)

			hexBytes = append(hexBytes,
				hexChar(byte(high>>12)),
				hexChar(byte(high>>8)),
				hexChar(byte(high>>4)),
				hexChar(byte(high)),
				hexChar(byte(low>>12)),
				hexChar(byte(low>>8)),
				hexChar(byte(low>>4)),
				hexChar(byte(low)),
			)
		}
	}

	return core.String("<" + string(hexBytes) + ">")
}

// hexChar converts a 4-bit value to a hex character
func hexChar(b byte) byte {
	b &= 0x0F
	if b < 10 {
		return '0' + b
	}
	return 'A' + (b - 10)
}

// createInfoDict creates a PDF Info dictionary from Metadata
func createInfoDict(metadata *Metadata) core.Dictionary {
	if metadata == nil {
		return nil
	}

	dict := core.Dictionary{}

	// Add standard fields
	if metadata.Title != "" {
		dict[core.Name("Title")] = encodeTextString(metadata.Title)
	}

	if metadata.Author != "" {
		dict[core.Name("Author")] = encodeTextString(metadata.Author)
	}

	if metadata.Subject != "" {
		dict[core.Name("Subject")] = encodeTextString(metadata.Subject)
	}

	if metadata.Keywords != "" {
		dict[core.Name("Keywords")] = encodeTextString(metadata.Keywords)
	}

	if metadata.Creator != "" {
		dict[core.Name("Creator")] = encodeTextString(metadata.Creator)
	}

	// Producer: use provided value or default to "gopdf"
	producer := metadata.Producer
	if producer == "" {
		producer = "gopdf"
	}
	dict[core.Name("Producer")] = encodeTextString(producer)

	// CreationDate: use provided value or current time
	creationDate := metadata.CreationDate
	if creationDate.IsZero() {
		creationDate = time.Now()
	}
	if dateStr := formatPDFDate(creationDate); dateStr != "" {
		dict[core.Name("CreationDate")] = core.String("(" + dateStr + ")")
	}

	// ModDate: only add if set
	if !metadata.ModDate.IsZero() {
		if dateStr := formatPDFDate(metadata.ModDate); dateStr != "" {
			dict[core.Name("ModDate")] = core.String("(" + dateStr + ")")
		}
	}

	// Add custom fields
	for key, value := range metadata.Custom {
		if key != "" && value != "" {
			dict[core.Name(key)] = encodeTextString(value)
		}
	}

	return dict
}

// needsUTF16 checks if a string contains non-ASCII characters
// that require UTF-16 encoding
func needsUTF16(s string) bool {
	return !utf8.ValidString(s) || !isASCII(s)
}
