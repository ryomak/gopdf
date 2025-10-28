package gopdf

import (
	"testing"
)

func TestFitText(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		bounds      Rectangle
		opts        FitTextOptions
		expectError bool
	}{
		{
			name: "Simple text fitting",
			text: "Hello World",
			bounds: Rectangle{
				X:      0,
				Y:      0,
				Width:  200,
				Height: 50,
			},
			opts: FitTextOptions{
				MaxFontSize: 20,
				MinFontSize: 8,
				LineSpacing: 1.2,
				Padding:     2,
				AllowShrink: true,
				AllowGrow:   false,
				Alignment:   AlignLeft,
			},
			expectError: false,
		},
		{
			name: "Long text requiring multiple lines",
			text: "This is a long text that will require multiple lines to fit within the specified bounds",
			bounds: Rectangle{
				X:      0,
				Y:      0,
				Width:  200,
				Height: 100,
			},
			opts: FitTextOptions{
				MaxFontSize: 16,
				MinFontSize: 8,
				LineSpacing: 1.3,
				Padding:     5,
				AllowShrink: true,
				AllowGrow:   false,
				Alignment:   AlignLeft,
			},
			expectError: false,
		},
		{
			name: "Text too large for bounds",
			text: "Very long text that absolutely cannot fit in a tiny box no matter how much we shrink it",
			bounds: Rectangle{
				X:      0,
				Y:      0,
				Width:  20,
				Height: 10,
			},
			opts: FitTextOptions{
				MaxFontSize: 12,
				MinFontSize: 6,
				LineSpacing: 1.2,
				Padding:     1,
				AllowShrink: true,
				AllowGrow:   false,
				Alignment:   AlignLeft,
			},
			expectError: true,
		},
		{
			name: "Empty text",
			text: "",
			bounds: Rectangle{
				X:      0,
				Y:      0,
				Width:  100,
				Height: 50,
			},
			opts: DefaultFitTextOptions(),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FitText(tt.text, tt.bounds, "Helvetica", tt.opts)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("Expected non-nil result")
				return
			}

			// フォントサイズが範囲内であることを検証
			if result.FontSize < tt.opts.MinFontSize || result.FontSize > tt.opts.MaxFontSize {
				t.Errorf("FontSize %.1f out of range [%.1f, %.1f]",
					result.FontSize, tt.opts.MinFontSize, tt.opts.MaxFontSize)
			}

			// 行が作成されていることを検証
			if tt.text != "" && len(result.Lines) == 0 {
				t.Error("Expected at least one line for non-empty text")
			}

			t.Logf("Result: FontSize=%.1f, LineHeight=%.1f, Lines=%d",
				result.FontSize, result.LineHeight, len(result.Lines))
			for i, line := range result.Lines {
				t.Logf("  Line %d: %q", i+1, line)
			}
		})
	}
}

func TestWrapText(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		maxWidth  float64
		fontSize  float64
		minLines  int
	}{
		{
			name:      "Short text",
			text:      "Hello",
			maxWidth:  200,
			fontSize:  12,
			minLines:  1,
		},
		{
			name:      "Text requiring wrapping",
			text:      "This is a longer text that should wrap",
			maxWidth:  100,
			fontSize:  12,
			minLines:  2,
		},
		{
			name:      "Empty text",
			text:      "",
			maxWidth:  100,
			fontSize:  12,
			minLines:  1,
		},
		{
			name:      "Text with newlines",
			text:      "Line 1\nLine 2\nLine 3",
			maxWidth:  200,
			fontSize:  12,
			minLines:  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := wrapText(tt.text, tt.maxWidth, "Helvetica", tt.fontSize)

			if len(lines) < tt.minLines {
				t.Errorf("Expected at least %d lines, got %d", tt.minLines, len(lines))
			}

			t.Logf("Wrapped into %d lines:", len(lines))
			for i, line := range lines {
				t.Logf("  %d: %q", i+1, line)
			}
		})
	}
}

func TestEstimateLines(t *testing.T) {
	text := "This is a test text that will be used to estimate line count"
	maxWidth := 150.0
	fontSize := 12.0

	lineCount := EstimateLines(text, maxWidth, "Helvetica", fontSize)

	if lineCount <= 0 {
		t.Errorf("EstimateLines returned %d, expected positive value", lineCount)
	}

	t.Logf("Estimated %d lines for text: %q", lineCount, text)
}

func TestEstimateTotalHeight(t *testing.T) {
	text := "Line 1\nLine 2\nLine 3"
	maxWidth := 200.0
	fontSize := 12.0
	lineSpacing := 1.5

	height := EstimateTotalHeight(text, maxWidth, "Helvetica", fontSize, lineSpacing)

	expectedMinHeight := fontSize * lineSpacing * 3 // 3行
	if height < expectedMinHeight {
		t.Errorf("Height %.1f is less than expected minimum %.1f", height, expectedMinHeight)
	}

	t.Logf("Estimated height: %.1f", height)
}

func TestDefaultFitTextOptions(t *testing.T) {
	opts := DefaultFitTextOptions()

	if opts.MaxFontSize <= opts.MinFontSize {
		t.Error("MaxFontSize should be greater than MinFontSize")
	}

	if opts.LineSpacing <= 0 {
		t.Error("LineSpacing should be positive")
	}

	if !opts.AllowShrink {
		t.Error("AllowShrink should be true by default")
	}

	t.Logf("Default options: MaxSize=%.1f, MinSize=%.1f, LineSpacing=%.1f",
		opts.MaxFontSize, opts.MinFontSize, opts.LineSpacing)
}

func TestFitTextInBlock(t *testing.T) {
	block := TextBlock{
		Text: "Original text",
		Bounds: Rectangle{
			X:      50,
			Y:      700,
			Width:  200,
			Height: 100,
		},
		Font:     "Helvetica",
		FontSize: 12,
	}

	newText := "New translated text that is longer than the original"
	opts := DefaultFitTextOptions()

	result, err := FitTextInBlock(newText, block, "Helvetica", opts)
	if err != nil {
		t.Fatalf("FitTextInBlock failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	t.Logf("Fitted text: FontSize=%.1f, Lines=%d", result.FontSize, len(result.Lines))
}

func TestTextAlign(t *testing.T) {
	alignments := []TextAlign{AlignLeft, AlignCenter, AlignRight}

	for _, align := range alignments {
		opts := DefaultFitTextOptions()
		opts.Alignment = align

		result, err := FitText("Test", Rectangle{Width: 100, Height: 50}, "Helvetica", opts)
		if err != nil {
			t.Errorf("FitText with alignment %d failed: %v", align, err)
			continue
		}

		if result == nil {
			t.Errorf("Expected non-nil result for alignment %d", align)
			continue
		}

		t.Logf("Alignment %d: FontSize=%.1f", align, result.FontSize)
	}
}
