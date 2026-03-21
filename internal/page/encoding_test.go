package page

import (
	"testing"
)

func TestEscapeString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no special characters",
			input: "Hello World",
			want:  "Hello World",
		},
		{
			name:  "backslash",
			input: "path\\file",
			want:  "path\\\\file",
		},
		{
			name:  "open parenthesis",
			input: "hello(world",
			want:  "hello\\(world",
		},
		{
			name:  "close parenthesis",
			input: "hello)world",
			want:  "hello\\)world",
		},
		{
			name:  "all special characters",
			input: "(test\\example)",
			want:  "\\(test\\\\example\\)",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EscapeString(tt.input)
			if got != tt.want {
				t.Errorf("EscapeString(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestTextToHexString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "ASCII characters",
			input: "AB",
			want:  "00410042",
		},
		{
			name:  "Japanese characters",
			input: "あ",
			want:  "3042",
		},
		{
			name:  "mixed characters",
			input: "Aあ",
			want:  "00413042",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "non-BMP character (emoji)",
			input: "\U0001F600", // 😀 U+1F600
			want:  "D83DDE00",  // UTF-16 surrogate pair: D83D DE00
		},
		{
			name:  "non-BMP with BMP mixed",
			input: "A\U0001F600B",
			want:  "0041D83DDE000042",
		},
		{
			name:  "musical symbol (non-BMP)",
			input: "\U0001D11E", // 𝄞 U+1D11E
			want:  "D834DD1E",  // UTF-16 surrogate pair: D834 DD1E
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TextToHexString(tt.input)
			if got != tt.want {
				t.Errorf("TextToHexString(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// mockGlyphIndexer implements GlyphIndexer for testing
type mockGlyphIndexer struct {
	mapping map[rune]int
}

func (m *mockGlyphIndexer) GetGlyphIndex(r rune) (int, error) {
	if idx, ok := m.mapping[r]; ok {
		return idx, nil
	}
	return int(r), nil // Default to rune value
}

// mockGlyphRecorder implements GlyphRecorder for testing
type mockGlyphRecorder struct {
	recorded map[uint16]rune
}

func (m *mockGlyphRecorder) RecordGlyph(glyphIndex uint16, r rune) {
	if m.recorded == nil {
		m.recorded = make(map[uint16]rune)
	}
	m.recorded[glyphIndex] = r
}

func TestTextToGlyphIndices(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		mapping map[rune]int
		want    string
		wantErr bool
	}{
		{
			name:    "ASCII with default mapping",
			text:    "AB",
			mapping: nil,
			want:    "00410042",
			wantErr: false,
		},
		{
			name:    "custom glyph mapping",
			text:    "AB",
			mapping: map[rune]int{'A': 100, 'B': 200},
			want:    "006400C8",
			wantErr: false,
		},
		{
			name:    "empty string",
			text:    "",
			mapping: nil,
			want:    "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			indexer := &mockGlyphIndexer{mapping: tt.mapping}
			recorder := &mockGlyphRecorder{}

			got, err := TextToGlyphIndices(tt.text, indexer, recorder)
			if (err != nil) != tt.wantErr {
				t.Errorf("TextToGlyphIndices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("TextToGlyphIndices() = %q, want %q", got, tt.want)
			}
		})
	}
}
