package translate

import (
	"strings"
	"testing"
)

// mockTranslator is a test translator that transforms text.
type mockTranslator struct {
	transform func(string) string
}

func (m *mockTranslator) Translate(text string) (string, error) {
	return m.transform(text), nil
}

func TestTranslateText(t *testing.T) {
	upperTranslator := &mockTranslator{
		transform: func(s string) string { return strings.ToUpper(s) },
	}

	tests := []struct {
		name       string
		text       string
		unit       TranslateUnit
		translator Translator
		want       string
	}{
		{
			name:       "block translation",
			text:       "hello world",
			unit:       TranslateUnitBlock,
			translator: upperTranslator,
			want:       "HELLO WORLD",
		},
		{
			name:       "line translation",
			text:       "line1\nline2\nline3",
			unit:       TranslateUnitLine,
			translator: upperTranslator,
			want:       "LINE1\nLINE2\nLINE3",
		},
		{
			name:       "sentence translation",
			text:       "Hello. World!",
			unit:       TranslateUnitSentence,
			translator: upperTranslator,
			want:       "HELLO. WORLD!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TranslateText(tt.text, tt.translator, tt.unit)
			if err != nil {
				t.Errorf("TranslateText() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("TranslateText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTranslateByDelimiter(t *testing.T) {
	upperTranslator := &mockTranslator{
		transform: func(s string) string { return strings.ToUpper(s) },
	}

	tests := []struct {
		name       string
		text       string
		splitDelim string
		joinDelim  string
		want       string
	}{
		{
			name:       "newline delimiter",
			text:       "a\nb\nc",
			splitDelim: "\n",
			joinDelim:  "\n",
			want:       "A\nB\nC",
		},
		{
			name:       "comma delimiter with different join",
			text:       "a,b,c",
			splitDelim: ",",
			joinDelim:  " - ",
			want:       "A - B - C",
		},
		{
			name:       "empty parts",
			text:       "a\n\nb",
			splitDelim: "\n",
			joinDelim:  "\n",
			want:       "A\n\nB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TranslateByDelimiter(tt.text, upperTranslator, tt.splitDelim, tt.joinDelim)
			if err != nil {
				t.Errorf("TranslateByDelimiter() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("TranslateByDelimiter() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsASCIIOnly(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{
			name: "ASCII only",
			s:    "Hello World 123!",
			want: true,
		},
		{
			name: "contains Japanese",
			s:    "Hello こんにちは",
			want: false,
		},
		{
			name: "contains emoji",
			s:    "Hello 😀",
			want: false,
		},
		{
			name: "empty string",
			s:    "",
			want: true,
		},
		{
			name: "all special ASCII",
			s:    "!@#$%^&*()",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsASCIIOnly(tt.s)
			if got != tt.want {
				t.Errorf("IsASCIIOnly() = %v, want %v", got, tt.want)
			}
		})
	}
}
