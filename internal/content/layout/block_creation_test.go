package layout

import (
	"testing"
)

func TestDetectFontStyle(t *testing.T) {
	tests := []struct {
		name       string
		fontName   string
		wantBold   bool
		wantItalic bool
	}{
		{
			name:       "plain font",
			fontName:   "Helvetica",
			wantBold:   false,
			wantItalic: false,
		},
		{
			name:       "bold font",
			fontName:   "Helvetica-Bold",
			wantBold:   true,
			wantItalic: false,
		},
		{
			name:       "italic font",
			fontName:   "Times-Italic",
			wantBold:   false,
			wantItalic: true,
		},
		{
			name:       "oblique font",
			fontName:   "Helvetica-Oblique",
			wantBold:   false,
			wantItalic: true,
		},
		{
			name:       "bold italic font",
			fontName:   "Times-BoldItalic",
			wantBold:   true,
			wantItalic: true,
		},
		{
			name:       "suffix BD",
			fontName:   "ArialBD",
			wantBold:   true,
			wantItalic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBold, gotItalic := DetectFontStyle(tt.fontName)
			if gotBold != tt.wantBold {
				t.Errorf("DetectFontStyle() bold = %v, want %v", gotBold, tt.wantBold)
			}
			if gotItalic != tt.wantItalic {
				t.Errorf("DetectFontStyle() italic = %v, want %v", gotItalic, tt.wantItalic)
			}
		})
	}
}

func TestCombineBlockText(t *testing.T) {
	tests := []struct {
		name  string
		lines [][]TextElement
		want  string
	}{
		{
			name:  "empty lines",
			lines: nil,
			want:  "",
		},
		{
			name: "single line single element",
			lines: [][]TextElement{
				{{Text: "Hello", X: 0, Size: 12}},
			},
			want: "Hello",
		},
		{
			name: "single line multiple elements close together",
			lines: [][]TextElement{
				{
					{Text: "Hello", X: 0, Width: 30, Size: 12},
					{Text: "World", X: 32, Size: 12}, // Gap < threshold
				},
			},
			want: "HelloWorld",
		},
		{
			name: "single line multiple elements with space",
			lines: [][]TextElement{
				{
					{Text: "Hello", X: 0, Width: 30, Size: 12},
					{Text: "World", X: 40, Size: 12}, // Gap > threshold (0.35 * 12 = 4.2)
				},
			},
			want: "Hello World",
		},
		{
			name: "multiple lines",
			lines: [][]TextElement{
				{{Text: "Line1", X: 0, Size: 12}},
				{{Text: "Line2", X: 0, Size: 12}},
			},
			want: "Line1\nLine2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CombineBlockText(tt.lines)
			if got != tt.want {
				t.Errorf("CombineBlockText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCreateTextBlockFromLines(t *testing.T) {
	tests := []struct {
		name         string
		lines        [][]TextElement
		wantText     string
		wantFont     string
		wantElements int
	}{
		{
			name:         "empty lines",
			lines:        nil,
			wantText:     "",
			wantFont:     "",
			wantElements: 0,
		},
		{
			name: "single line",
			lines: [][]TextElement{
				{
					{Text: "Hello", X: 0, Y: 100, Width: 30, Height: 12, Font: "Helvetica", Size: 12},
				},
			},
			wantText:     "Hello",
			wantFont:     "Helvetica",
			wantElements: 1,
		},
		{
			name: "multiple lines",
			lines: [][]TextElement{
				{
					{Text: "Line1", X: 0, Y: 100, Width: 30, Height: 12, Font: "Helvetica", Size: 12},
				},
				{
					{Text: "Line2", X: 0, Y: 85, Width: 30, Height: 12, Font: "Helvetica", Size: 12},
				},
			},
			wantText:     "Line1\nLine2",
			wantFont:     "Helvetica",
			wantElements: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CreateTextBlockFromLines(tt.lines)
			if got.Text != tt.wantText {
				t.Errorf("CreateTextBlockFromLines().Text = %q, want %q", got.Text, tt.wantText)
			}
			if got.Font != tt.wantFont {
				t.Errorf("CreateTextBlockFromLines().Font = %q, want %q", got.Font, tt.wantFont)
			}
			if len(got.Elements) != tt.wantElements {
				t.Errorf("CreateTextBlockFromLines() elements = %d, want %d", len(got.Elements), tt.wantElements)
			}
		})
	}
}
