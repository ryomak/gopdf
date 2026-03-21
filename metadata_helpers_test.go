package gopdf

import (
	"testing"
	"time"

	"github.com/ryomak/gopdf/internal/core"
)

func TestDecodeHexString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple ASCII hex",
			input:    "48656C6C6F",
			expected: "Hello",
		},
		{
			name:     "lowercase hex",
			input:    "48656c6c6f",
			expected: "Hello",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "odd length returns empty",
			input:    "4865C",
			expected: "",
		},
		{
			name:     "hex with spaces removed",
			input:    "48 65 6C 6C 6F",
			expected: "Hello",
		},
		{
			name:     "single byte",
			input:    "41",
			expected: "A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := decodeHexString(tt.input)
			if got != tt.expected {
				t.Errorf("decodeHexString(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestDecodeUTF16BE(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple BMP characters",
			input:    "00480065006C006C006F",
			expected: "Hello",
		},
		{
			name:     "Japanese characters",
			input:    "65E5672C8A9E",
			expected: "\u65E5\u672C\u8A9E", // 日本語
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "non-multiple-of-4 returns empty",
			input:    "004800",
			expected: "",
		},
		{
			name:     "surrogate pair (emoji U+1F600)",
			input:    "D83DDE00",
			expected: "\U0001F600",
		},
		{
			name:     "hex with spaces",
			input:    "00 48 00 69",
			expected: "Hi",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := decodeUTF16BE(tt.input)
			if got != tt.expected {
				t.Errorf("decodeUTF16BE(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestFormatPDFDate(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "zero time",
			input:    time.Time{},
			expected: "",
		},
		{
			name:     "UTC time",
			input:    time.Date(2025, 1, 29, 12, 30, 45, 0, time.UTC),
			expected: "D:20250129123045+00'00'",
		},
		{
			name:     "positive offset (JST +9:00)",
			input:    time.Date(2025, 6, 15, 9, 0, 0, 0, time.FixedZone("JST", 9*3600)),
			expected: "D:20250615090000+09'00'",
		},
		{
			name:     "negative offset (EST -5:00)",
			input:    time.Date(2025, 3, 1, 18, 45, 30, 0, time.FixedZone("EST", -5*3600)),
			expected: "D:20250301184530-05'00'",
		},
		{
			name:     "offset with minutes (IST +5:30)",
			input:    time.Date(2025, 12, 31, 23, 59, 59, 0, time.FixedZone("IST", 5*3600+30*60)),
			expected: "D:20251231235959+05'30'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatPDFDate(tt.input)
			if got != tt.expected {
				t.Errorf("formatPDFDate(%v) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestEncodeTextString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantType string // "paren" for (...), "hex" for <...>, "empty" for empty
	}{
		{
			name:     "empty string",
			input:    "",
			wantType: "empty",
		},
		{
			name:     "ASCII text uses parentheses",
			input:    "Hello World",
			wantType: "paren",
		},
		{
			name:     "Japanese text uses hex with BOM",
			input:    "\u65E5\u672C\u8A9E",
			wantType: "hex",
		},
		{
			name:     "mixed ASCII and non-ASCII uses hex",
			input:    "Hello \u4E16\u754C",
			wantType: "hex",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := encodeTextString(tt.input)
			s := string(got.(core.String))

			switch tt.wantType {
			case "empty":
				if s != "" {
					t.Errorf("expected empty string, got %q", s)
				}
			case "paren":
				if len(s) < 2 || s[0] != '(' || s[len(s)-1] != ')' {
					t.Errorf("expected parenthesized string, got %q", s)
				}
			case "hex":
				if len(s) < 2 || s[0] != '<' || s[len(s)-1] != '>' {
					t.Errorf("expected hex string, got %q", s)
				}
				// Check for BOM (FEFF)
				if len(s) > 5 && s[1:5] != "FEFF" {
					t.Errorf("expected UTF-16BE BOM, got %q", s[1:5])
				}
			}
		})
	}
}

func TestEncodeTextString_RoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"ASCII", "Hello World"},
		{"Japanese", "\u65E5\u672C\u8A9E"},
		{"Emoji surrogate pair", "\U0001F600"},
		{"Mixed", "Test \u30C6\u30B9\u30C8"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := encodeTextString(tt.input)
			decoded := decodeTextString(encoded)
			if decoded != tt.input {
				t.Errorf("round-trip failed: input=%q, encoded=%q, decoded=%q",
					tt.input, string(encoded.(core.String)), decoded)
			}
		})
	}
}

func TestParsePDFDate(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantErr   bool
		wantYear  int
		wantMonth time.Month
	}{
		{
			name:    "empty string",
			input:   "",
			wantErr: false,
		},
		{
			name:    "too short",
			input:   "D:20",
			wantErr: true,
		},
		{
			name:      "year only",
			input:     "D:2025",
			wantYear:  2025,
			wantMonth: time.January,
		},
		{
			name:      "full date with timezone",
			input:     "D:20250129123045+09'00'",
			wantYear:  2025,
			wantMonth: time.January,
		},
		{
			name:      "Z timezone",
			input:     "D:20250601120000Z",
			wantYear:  2025,
			wantMonth: time.June,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePDFDate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePDFDate(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if tt.input == "" {
				if !got.IsZero() {
					t.Errorf("expected zero time for empty input")
				}
				return
			}
			if got.Year() != tt.wantYear {
				t.Errorf("year = %d, want %d", got.Year(), tt.wantYear)
			}
			if got.Month() != tt.wantMonth {
				t.Errorf("month = %v, want %v", got.Month(), tt.wantMonth)
			}
		})
	}
}

func TestDecodeTextString(t *testing.T) {
	tests := []struct {
		name     string
		input    core.Object
		expected string
	}{
		{
			name:     "non-string object",
			input:    core.Integer(42),
			expected: "",
		},
		{
			name:     "parenthesized string",
			input:    core.String("(Hello World)"),
			expected: "Hello World",
		},
		{
			name:     "escaped parentheses",
			input:    core.String("(Hello \\(World\\))"),
			expected: "Hello (World)",
		},
		{
			name:     "hex string without BOM",
			input:    core.String("<48656C6C6F>"),
			expected: "Hello",
		},
		{
			name:     "UTF-16BE with BOM",
			input:    core.String("<FEFF004800650065>"),
			expected: "Hee",
		},
		{
			name:     "plain string (no delimiters)",
			input:    core.String("plain"),
			expected: "plain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := decodeTextString(tt.input)
			if got != tt.expected {
				t.Errorf("decodeTextString(%v) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestHexChar(t *testing.T) {
	tests := []struct {
		input    byte
		expected byte
	}{
		{0x00, '0'},
		{0x09, '9'},
		{0x0A, 'A'},
		{0x0F, 'F'},
		{0xFF, 'F'}, // masked to 0x0F
	}

	for _, tt := range tests {
		got := hexChar(tt.input)
		if got != tt.expected {
			t.Errorf("hexChar(0x%02X) = %c, want %c", tt.input, got, tt.expected)
		}
	}
}

func TestIsASCII(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"empty", "", true},
		{"ASCII only", "Hello World 123!", true},
		{"Japanese", "\u65E5\u672C", false},
		{"mixed", "Hello \u4E16\u754C", false},
		{"max ASCII", string([]byte{127}), true},
		{"above ASCII", string([]byte{128}), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isASCII(tt.input)
			if got != tt.expected {
				t.Errorf("isASCII(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestUnescapeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"no escapes", "hello", "hello"},
		{"escaped parens", "hello \\(world\\)", "hello (world)"},
		{"escaped backslash", "path\\\\to\\\\file", "path\\to\\file"},
		{"mixed", "\\(a\\\\b\\)", "(a\\b)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := unescapeString(tt.input)
			if got != tt.expected {
				t.Errorf("unescapeString(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestCreateInfoDict(t *testing.T) {
	tests := []struct {
		name     string
		metadata *Metadata
		wantNil  bool
		wantKeys []string
	}{
		{
			name:     "nil metadata",
			metadata: nil,
			wantNil:  true,
		},
		{
			name: "title only",
			metadata: &Metadata{
				Title: "Test",
			},
			wantKeys: []string{"Title", "Producer", "CreationDate"},
		},
		{
			name: "all fields",
			metadata: &Metadata{
				Title:    "Test",
				Author:   "Author",
				Subject:  "Subject",
				Keywords: "key1, key2",
				Creator:  "Creator",
				Producer: "Producer",
				ModDate:  time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			wantKeys: []string{"Title", "Author", "Subject", "Keywords", "Creator", "Producer", "CreationDate", "ModDate"},
		},
		{
			name: "custom fields",
			metadata: &Metadata{
				Custom: map[string]string{
					"Department": "Engineering",
				},
			},
			wantKeys: []string{"Producer", "CreationDate", "Department"},
		},
		{
			name: "empty custom field values are skipped",
			metadata: &Metadata{
				Custom: map[string]string{
					"Valid":   "value",
					"":        "empty key",
					"EmptyVal": "",
				},
			},
			wantKeys: []string{"Producer", "CreationDate", "Valid"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := createInfoDict(tt.metadata)
			if tt.wantNil {
				if got != nil {
					t.Errorf("expected nil, got %v", got)
				}
				return
			}
			for _, key := range tt.wantKeys {
				if _, ok := got[core.Name(key)]; !ok {
					t.Errorf("missing key %q in info dict", key)
				}
			}
		})
	}
}

func TestParseInfoDict(t *testing.T) {
	t.Run("standard fields", func(t *testing.T) {
		dict := core.Dictionary{
			core.Name("Title"):  core.String("(Test Title)"),
			core.Name("Author"): core.String("(Test Author)"),
		}
		m := parseInfoDict(dict)
		if m.Title != "Test Title" {
			t.Errorf("Title = %q, want %q", m.Title, "Test Title")
		}
		if m.Author != "Test Author" {
			t.Errorf("Author = %q, want %q", m.Author, "Test Author")
		}
	})

	t.Run("date fields", func(t *testing.T) {
		dict := core.Dictionary{
			core.Name("CreationDate"): core.String("(D:20250129120000+00'00')"),
		}
		m := parseInfoDict(dict)
		if m.CreationDate.IsZero() {
			t.Error("CreationDate should not be zero")
		}
		if m.CreationDate.Year() != 2025 {
			t.Errorf("year = %d, want 2025", m.CreationDate.Year())
		}
	})

	t.Run("custom fields", func(t *testing.T) {
		dict := core.Dictionary{
			core.Name("CustomKey"): core.String("(CustomValue)"),
			core.Name("Trapped"):   core.String("(False)"), // should be skipped
		}
		m := parseInfoDict(dict)
		if v, ok := m.Custom["CustomKey"]; !ok || v != "CustomValue" {
			t.Errorf("Custom[CustomKey] = %q, want %q", v, "CustomValue")
		}
		if _, ok := m.Custom["Trapped"]; ok {
			t.Error("Trapped should be skipped from custom fields")
		}
	})
}
