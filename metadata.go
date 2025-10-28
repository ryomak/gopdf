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

// parsePDFDate parses a PDF date string to time.Time
// Format: D:YYYYMMDDHHmmSSOHH'mm'
// Example: D:20250129123045+09'00'
func parsePDFDate(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}

	// Remove "D:" prefix if present
	if strings.HasPrefix(s, "D:") {
		s = s[2:]
	}

	// Minimum length check (at least YYYY)
	if len(s) < 4 {
		return time.Time{}, fmt.Errorf("invalid PDF date format: too short")
	}

	// Parse components
	year := 0
	month := 1
	day := 1
	hour := 0
	minute := 0
	second := 0
	offsetHours := 0
	offsetMinutes := 0

	// Year (required)
	fmt.Sscanf(s[0:4], "%d", &year)

	// Month (optional)
	if len(s) >= 6 {
		fmt.Sscanf(s[4:6], "%d", &month)
	}

	// Day (optional)
	if len(s) >= 8 {
		fmt.Sscanf(s[6:8], "%d", &day)
	}

	// Hour (optional)
	if len(s) >= 10 {
		fmt.Sscanf(s[8:10], "%d", &hour)
	}

	// Minute (optional)
	if len(s) >= 12 {
		fmt.Sscanf(s[10:12], "%d", &minute)
	}

	// Second (optional)
	if len(s) >= 14 {
		fmt.Sscanf(s[12:14], "%d", &second)
	}

	// Timezone offset (optional)
	loc := time.UTC
	if len(s) > 14 {
		tzPart := s[14:]
		if tzPart == "Z" {
			loc = time.UTC
		} else if len(tzPart) >= 3 {
			sign := tzPart[0]
			// Remove apostrophes from offset
			tzPart = strings.ReplaceAll(tzPart[1:], "'", "")

			parts := strings.Split(tzPart, "+")
			if len(parts) < 2 {
				parts = strings.Split(tzPart, "-")
			}

			if len(tzPart) >= 2 {
				fmt.Sscanf(tzPart[0:2], "%d", &offsetHours)
			}
			if len(tzPart) >= 4 {
				fmt.Sscanf(tzPart[2:4], "%d", &offsetMinutes)
			}

			offsetSeconds := offsetHours*3600 + offsetMinutes*60
			if sign == '-' {
				offsetSeconds = -offsetSeconds
			}
			loc = time.FixedZone("PDF", offsetSeconds)
		}
	}

	return time.Date(year, time.Month(month), day, hour, minute, second, 0, loc), nil
}

// unescapeString unescapes PDF string special characters
// Unescapes: \(, \), \\
func unescapeString(s string) string {
	s = strings.ReplaceAll(s, "\\(", "(")
	s = strings.ReplaceAll(s, "\\)", ")")
	s = strings.ReplaceAll(s, "\\\\", "\\")
	return s
}

// decodeTextString decodes a PDF text string
// Handles both PDFDocEncoding (ASCII with escapes) and UTF-16BE (with BOM)
func decodeTextString(obj core.Object) string {
	str, ok := obj.(core.String)
	if !ok {
		return ""
	}

	s := string(str)

	// Remove parentheses if present (literal string)
	if len(s) >= 2 && s[0] == '(' && s[len(s)-1] == ')' {
		s = s[1 : len(s)-1]
		// Unescape
		s = unescapeString(s)
		return s
	}

	// Remove angle brackets if present (hex string)
	if len(s) >= 2 && s[0] == '<' && s[len(s)-1] == '>' {
		s = s[1 : len(s)-1]

		// Check for UTF-16BE BOM (FEFF)
		if strings.HasPrefix(s, "FEFF") || strings.HasPrefix(s, "feff") {
			// UTF-16BE encoded
			return decodeUTF16BE(s[4:]) // Skip BOM
		}

		// Regular hex string
		return decodeHexString(s)
	}

	return s
}

// decodeHexString decodes a hex string to UTF-8
func decodeHexString(hexStr string) string {
	// Remove spaces
	hexStr = strings.ReplaceAll(hexStr, " ", "")

	if len(hexStr)%2 != 0 {
		return ""
	}

	bytes := make([]byte, len(hexStr)/2)
	for i := 0; i < len(hexStr); i += 2 {
		var b byte
		fmt.Sscanf(hexStr[i:i+2], "%02x", &b)
		bytes[i/2] = b
	}

	return string(bytes)
}

// decodeUTF16BE decodes a UTF-16BE hex string to UTF-8
func decodeUTF16BE(hexStr string) string {
	// Remove spaces
	hexStr = strings.ReplaceAll(hexStr, " ", "")

	if len(hexStr)%4 != 0 {
		return ""
	}

	runes := make([]rune, 0)

	for i := 0; i < len(hexStr); i += 4 {
		var code uint16
		fmt.Sscanf(hexStr[i:i+4], "%04x", &code)

		// Check for surrogate pair
		if code >= 0xD800 && code <= 0xDBFF {
			// High surrogate
			if i+8 <= len(hexStr) {
				var low uint16
				fmt.Sscanf(hexStr[i+4:i+8], "%04x", &low)
				if low >= 0xDC00 && low <= 0xDFFF {
					// Valid surrogate pair
					r := 0x10000 + (rune(code&0x3FF)<<10) + rune(low&0x3FF)
					runes = append(runes, r)
					i += 4 // Skip the low surrogate in next iteration
					continue
				}
			}
		}

		runes = append(runes, rune(code))
	}

	return string(runes)
}

// parseInfoDict parses a PDF Info dictionary into a Metadata struct
func parseInfoDict(dict core.Dictionary) Metadata {
	metadata := Metadata{
		Custom: make(map[string]string),
	}

	// Standard fields to check
	standardFields := map[string]*string{
		"Title":    &metadata.Title,
		"Author":   &metadata.Author,
		"Subject":  &metadata.Subject,
		"Keywords": &metadata.Keywords,
		"Creator":  &metadata.Creator,
		"Producer": &metadata.Producer,
	}

	// Parse standard text fields
	for key, field := range standardFields {
		if obj, ok := dict[core.Name(key)]; ok {
			*field = decodeTextString(obj)
		}
	}

	// Parse date fields
	if obj, ok := dict[core.Name("CreationDate")]; ok {
		dateStr := decodeTextString(obj)
		if t, err := parsePDFDate(dateStr); err == nil {
			metadata.CreationDate = t
		}
	}

	if obj, ok := dict[core.Name("ModDate")]; ok {
		dateStr := decodeTextString(obj)
		if t, err := parsePDFDate(dateStr); err == nil {
			metadata.ModDate = t
		}
	}

	// Parse custom fields (any field not in standard list)
	for key, obj := range dict {
		keyStr := string(key)
		// Skip standard fields
		if _, isStandard := standardFields[keyStr]; isStandard {
			continue
		}
		if keyStr == "CreationDate" || keyStr == "ModDate" || keyStr == "Trapped" {
			continue
		}

		// Add to custom fields
		value := decodeTextString(obj)
		if value != "" {
			metadata.Custom[keyStr] = value
		}
	}

	return metadata
}
