package gopdf

import (
	"testing"
)

func TestNewRubyText(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		ruby     string
		expected RubyText
	}{
		{
			name:     "Simple Japanese text",
			base:     "漢字",
			ruby:     "かんじ",
			expected: RubyText{Base: "漢字", Ruby: "かんじ"},
		},
		{
			name:     "Empty strings",
			base:     "",
			ruby:     "",
			expected: RubyText{Base: "", Ruby: ""},
		},
		{
			name:     "Multi-character base",
			base:     "東京都",
			ruby:     "とうきょうと",
			expected: RubyText{Base: "東京都", Ruby: "とうきょうと"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewRubyText(tt.base, tt.ruby)
			if result != tt.expected {
				t.Errorf("NewRubyText(%q, %q) = %+v, want %+v",
					tt.base, tt.ruby, result, tt.expected)
			}
		})
	}
}

func TestNewRubyTextPairs(t *testing.T) {
	tests := []struct {
		name     string
		pairs    []string
		expected []RubyText
	}{
		{
			name:  "Two pairs",
			pairs: []string{"漢字", "かんじ", "日本", "にほん"},
			expected: []RubyText{
				{Base: "漢字", Ruby: "かんじ"},
				{Base: "日本", Ruby: "にほん"},
			},
		},
		{
			name:  "Single pair",
			pairs: []string{"東京", "とうきょう"},
			expected: []RubyText{
				{Base: "東京", Ruby: "とうきょう"},
			},
		},
		{
			name:     "Empty input",
			pairs:    []string{},
			expected: []RubyText{},
		},
		{
			name:  "Odd number (last element ignored)",
			pairs: []string{"私", "わたし", "日本", "にほん", "住"},
			expected: []RubyText{
				{Base: "私", Ruby: "わたし"},
				{Base: "日本", Ruby: "にほん"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewRubyTextPairs(tt.pairs...)
			if len(result) != len(tt.expected) {
				t.Fatalf("NewRubyTextPairs() returned %d items, want %d",
					len(result), len(tt.expected))
			}
			for i, r := range result {
				if r != tt.expected[i] {
					t.Errorf("NewRubyTextPairs()[%d] = %+v, want %+v",
						i, r, tt.expected[i])
				}
			}
		})
	}
}

func TestDefaultRubyStyle(t *testing.T) {
	style := DefaultRubyStyle()

	if style.Alignment != RubyAlignCenter {
		t.Errorf("DefaultRubyStyle().Alignment = %v, want %v",
			style.Alignment, RubyAlignCenter)
	}
	if style.Offset != 1.0 {
		t.Errorf("DefaultRubyStyle().Offset = %f, want %f",
			style.Offset, 1.0)
	}
	if style.SizeRatio != 0.5 {
		t.Errorf("DefaultRubyStyle().SizeRatio = %f, want %f",
			style.SizeRatio, 0.5)
	}
	if style.CopyMode != RubyCopyBase {
		t.Errorf("DefaultRubyStyle().CopyMode = %v, want %v",
			style.CopyMode, RubyCopyBase)
	}
}

func TestRubyAlignment(t *testing.T) {
	// Test that constants have expected values
	tests := []struct {
		name      string
		alignment RubyAlignment
		expected  int
	}{
		{"Center", RubyAlignCenter, 0},
		{"Left", RubyAlignLeft, 1},
		{"Right", RubyAlignRight, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.alignment) != tt.expected {
				t.Errorf("%s alignment = %d, want %d",
					tt.name, int(tt.alignment), tt.expected)
			}
		})
	}
}

func TestRubyCopyMode(t *testing.T) {
	// Test that constants have expected values
	tests := []struct {
		name     string
		copyMode RubyCopyMode
		expected int
	}{
		{"Base", RubyCopyBase, 0},
		{"Ruby", RubyCopyRuby, 1},
		{"Both", RubyCopyBoth, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.copyMode) != tt.expected {
				t.Errorf("%s copy mode = %d, want %d",
					tt.name, int(tt.copyMode), tt.expected)
			}
		})
	}
}
