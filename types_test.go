package gopdf

import (
	"testing"
)

func TestFitText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		bounds   Rectangle
		fontName string
		opts     FitOptions
		wantErr  bool
	}{
		{
			name:     "simple text fits",
			text:     "Hello",
			bounds:   Rectangle{X: 0, Y: 0, Width: 200, Height: 100},
			fontName: "Helvetica",
			opts:     DefaultFitOptions(),
			wantErr:  false,
		},
		{
			name:     "empty text",
			text:     "",
			bounds:   Rectangle{X: 0, Y: 0, Width: 200, Height: 100},
			fontName: "Helvetica",
			opts:     DefaultFitOptions(),
			wantErr:  false, // empty text returns empty result without error
		},
		{
			name:     "long text with small bounds",
			text:     "This is a longer text that needs to be fitted into a small area",
			bounds:   Rectangle{X: 0, Y: 0, Width: 100, Height: 50},
			fontName: "Helvetica",
			opts:     DefaultFitOptions(),
			wantErr:  false,
		},
		{
			name:     "custom fit options with shrink",
			text:     "Test text",
			bounds:   Rectangle{X: 10, Y: 20, Width: 150, Height: 80},
			fontName: "Helvetica",
			opts: NewFitOptions(
				WithFitMaxFontSize(24),
				WithFitMinFontSize(6),
				WithFitAllowShrink(true),
				WithFitLineSpacing(1.5),
			),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FitText(tt.text, tt.bounds, tt.fontName, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("FitText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if result == nil {
				t.Fatal("FitText() returned nil result without error")
			}
			if result.FontSize <= 0 {
				t.Errorf("FontSize = %f, want > 0", result.FontSize)
			}
			if len(result.Lines) == 0 {
				t.Error("expected at least one line")
			}
		})
	}
}

func TestDefaultFitOptions(t *testing.T) {
	opts := DefaultFitOptions()

	// Verify we get reasonable default values
	if opts.MaxFontSize <= 0 {
		t.Errorf("MaxFontSize = %f, want > 0", opts.MaxFontSize)
	}
	if opts.MinFontSize <= 0 {
		t.Errorf("MinFontSize = %f, want > 0", opts.MinFontSize)
	}
	if opts.LineSpacing <= 0 {
		t.Errorf("LineSpacing = %f, want > 0", opts.LineSpacing)
	}
}

func TestNewFitOptions(t *testing.T) {
	tests := []struct {
		name          string
		opts          []FitOptionFunc
		checkMaxFont  float64
		checkMinFont  float64
		checkAlign    Align
	}{
		{
			name: "custom max font",
			opts: []FitOptionFunc{
				WithFitMaxFontSize(24),
			},
			checkMaxFont: 24,
		},
		{
			name: "custom min font",
			opts: []FitOptionFunc{
				WithFitMinFontSize(8),
			},
			checkMinFont: 8,
		},
		{
			name: "center alignment",
			opts: []FitOptionFunc{
				WithFitAlignment(AlignCenter),
			},
			checkAlign: AlignCenter,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := NewFitOptions(tt.opts...)
			if tt.checkMaxFont > 0 && opts.MaxFontSize != tt.checkMaxFont {
				t.Errorf("MaxFontSize = %f, want %f", opts.MaxFontSize, tt.checkMaxFont)
			}
			if tt.checkMinFont > 0 && opts.MinFontSize != tt.checkMinFont {
				t.Errorf("MinFontSize = %f, want %f", opts.MinFontSize, tt.checkMinFont)
			}
			if tt.checkAlign != 0 && opts.Alignment != tt.checkAlign {
				t.Errorf("Alignment = %v, want %v", opts.Alignment, tt.checkAlign)
			}
		})
	}
}

func TestAlignConstants(t *testing.T) {
	// Verify alignment constants are distinct
	if AlignLeft == AlignCenter {
		t.Error("AlignLeft should not equal AlignCenter")
	}
	if AlignLeft == AlignRight {
		t.Error("AlignLeft should not equal AlignRight")
	}
	if AlignCenter == AlignRight {
		t.Error("AlignCenter should not equal AlignRight")
	}
}
