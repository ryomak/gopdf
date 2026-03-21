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

func TestCreateTextBlock(t *testing.T) {
	tests := []struct {
		name       string
		elements   []TextElement
		wantText   string
		wantFont   string
		wantBold   bool
		wantItalic bool
		wantRect   Rectangle
	}{
		{
			name: "single element",
			elements: []TextElement{
				{Text: "Hello", X: 10, Y: 100, Width: 30, Height: 12, Font: "Helvetica", Size: 12},
			},
			wantText: "Hello",
			wantFont: "Helvetica",
			wantRect: Rectangle{X: 10, Y: 100, Width: 30, Height: 12},
		},
		{
			name: "multiple elements close together no space",
			elements: []TextElement{
				{Text: "He", X: 10, Y: 100, Width: 12, Height: 12, Font: "Helvetica", Size: 12},
				{Text: "llo", X: 22, Y: 100, Width: 18, Height: 12, Font: "Helvetica", Size: 12},
			},
			wantText: "Hello",
			wantFont: "Helvetica",
			wantRect: Rectangle{X: 10, Y: 100, Width: 30, Height: 12},
		},
		{
			name: "multiple elements with space gap",
			elements: []TextElement{
				{Text: "Hello", X: 10, Y: 100, Width: 30, Height: 12, Font: "Helvetica", Size: 12},
				{Text: "World", X: 50, Y: 100, Width: 30, Height: 12, Font: "Helvetica", Size: 12},
			},
			wantText: "Hello World",
			wantFont: "Helvetica",
			wantRect: Rectangle{X: 10, Y: 100, Width: 70, Height: 12},
		},
		{
			name: "bold font detected",
			elements: []TextElement{
				{Text: "Bold", X: 10, Y: 100, Width: 30, Height: 14, Font: "Helvetica-Bold", Size: 14},
			},
			wantText:   "Bold",
			wantFont:   "Helvetica-Bold",
			wantBold:   true,
			wantItalic: false,
			wantRect:   Rectangle{X: 10, Y: 100, Width: 30, Height: 14},
		},
		{
			name: "italic font detected",
			elements: []TextElement{
				{Text: "Italic", X: 10, Y: 100, Width: 30, Height: 12, Font: "Times-Italic", Size: 12},
			},
			wantText:   "Italic",
			wantFont:   "Times-Italic",
			wantBold:   false,
			wantItalic: true,
			wantRect:   Rectangle{X: 10, Y: 100, Width: 30, Height: 12},
		},
		{
			name: "bounding box spans multiple elements at different positions",
			elements: []TextElement{
				{Text: "A", X: 10, Y: 200, Width: 10, Height: 12, Font: "Helvetica", Size: 12},
				{Text: "B", X: 50, Y: 180, Width: 10, Height: 14, Font: "Helvetica", Size: 12},
			},
			wantText: "A B",
			wantFont: "Helvetica",
			wantRect: Rectangle{X: 10, Y: 180, Width: 50, Height: 32},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CreateTextBlock(tt.elements)
			if got.Text != tt.wantText {
				t.Errorf("Text = %q, want %q", got.Text, tt.wantText)
			}
			if got.Font != tt.wantFont {
				t.Errorf("Font = %q, want %q", got.Font, tt.wantFont)
			}
			if got.IsBold != tt.wantBold {
				t.Errorf("IsBold = %v, want %v", got.IsBold, tt.wantBold)
			}
			if got.IsItalic != tt.wantItalic {
				t.Errorf("IsItalic = %v, want %v", got.IsItalic, tt.wantItalic)
			}
			if got.Rect != tt.wantRect {
				t.Errorf("Rect = %+v, want %+v", got.Rect, tt.wantRect)
			}
			if len(got.Elements) != len(tt.elements) {
				t.Errorf("Elements count = %d, want %d", len(got.Elements), len(tt.elements))
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
