package layout

import (
	"testing"
)

func TestGroupElementsByLine(t *testing.T) {
	tests := []struct {
		name         string
		elements     []TextElement
		wantNumLines int
	}{
		{
			name:         "empty slice",
			elements:     nil,
			wantNumLines: 0,
		},
		{
			name: "single element",
			elements: []TextElement{
				{X: 0, Y: 100, Size: 12, Font: "Helvetica"},
			},
			wantNumLines: 1,
		},
		{
			name: "same line elements",
			elements: []TextElement{
				{X: 0, Y: 100, Size: 12, Font: "Helvetica"},
				{X: 50, Y: 100, Size: 12, Font: "Helvetica"},
				{X: 100, Y: 100, Size: 12, Font: "Helvetica"},
			},
			wantNumLines: 1,
		},
		{
			name: "different lines by Y",
			elements: []TextElement{
				{X: 0, Y: 100, Size: 12, Font: "Helvetica"},
				{X: 0, Y: 80, Size: 12, Font: "Helvetica"},
			},
			wantNumLines: 2,
		},
		{
			name: "different lines by font",
			elements: []TextElement{
				{X: 0, Y: 100, Size: 12, Font: "Helvetica"},
				{X: 50, Y: 100, Size: 12, Font: "Times-Roman"},
			},
			wantNumLines: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GroupElementsByLine(tt.elements)
			if len(got) != tt.wantNumLines {
				t.Errorf("GroupElementsByLine() returned %d lines, want %d", len(got), tt.wantNumLines)
			}
		})
	}
}

func TestShouldMergeLines(t *testing.T) {
	tests := []struct {
		name     string
		prevLine []TextElement
		currLine []TextElement
		want     bool
	}{
		{
			name:     "empty prev line",
			prevLine: nil,
			currLine: []TextElement{{Y: 100, Size: 12, Font: "Helvetica"}},
			want:     false,
		},
		{
			name:     "empty curr line",
			prevLine: []TextElement{{Y: 100, Size: 12, Font: "Helvetica"}},
			currLine: nil,
			want:     false,
		},
		{
			name:     "same font and close spacing",
			prevLine: []TextElement{{X: 0, Y: 100, Height: 12, Size: 12, Font: "Helvetica"}},
			currLine: []TextElement{{X: 0, Y: 85, Height: 12, Size: 12, Font: "Helvetica"}},
			want:     true,
		},
		{
			name:     "different font",
			prevLine: []TextElement{{X: 0, Y: 100, Size: 12, Font: "Helvetica"}},
			currLine: []TextElement{{X: 0, Y: 85, Size: 12, Font: "Times-Roman"}},
			want:     false,
		},
		{
			name:     "large line spacing",
			prevLine: []TextElement{{X: 0, Y: 100, Height: 12, Size: 12, Font: "Helvetica"}},
			currLine: []TextElement{{X: 0, Y: 50, Height: 12, Size: 12, Font: "Helvetica"}},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldMergeLines(tt.prevLine, tt.currLine)
			if got != tt.want {
				t.Errorf("ShouldMergeLines() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHaveSameFont(t *testing.T) {
	tests := []struct {
		name  string
		line1 []TextElement
		line2 []TextElement
		want  bool
	}{
		{
			name:  "both empty",
			line1: nil,
			line2: nil,
			want:  true,
		},
		{
			name:  "one empty",
			line1: []TextElement{{Font: "Helvetica", Size: 12}},
			line2: nil,
			want:  true,
		},
		{
			name:  "same font and size",
			line1: []TextElement{{Font: "Helvetica", Size: 12}},
			line2: []TextElement{{Font: "Helvetica", Size: 12}},
			want:  true,
		},
		{
			name:  "different font",
			line1: []TextElement{{Font: "Helvetica", Size: 12}},
			line2: []TextElement{{Font: "Times-Roman", Size: 12}},
			want:  false,
		},
		{
			name:  "different size",
			line1: []TextElement{{Font: "Helvetica", Size: 12}},
			line2: []TextElement{{Font: "Helvetica", Size: 14}},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HaveSameFont(tt.line1, tt.line2)
			if got != tt.want {
				t.Errorf("HaveSameFont() = %v, want %v", got, tt.want)
			}
		})
	}
}
