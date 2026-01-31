package text

import (
	"testing"

	"github.com/ryomak/gopdf/internal/content/layout"
)

func TestWrap(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		maxWidth  float64
		fontSize  float64
		wantLines int
	}{
		{
			name:      "empty text",
			text:      "",
			maxWidth:  100,
			fontSize:  10,
			wantLines: 1,
		},
		{
			name:      "single word",
			text:      "Hello",
			maxWidth:  100,
			fontSize:  10,
			wantLines: 1,
		},
		{
			name:      "multiple words",
			text:      "Hello World",
			maxWidth:  100,
			fontSize:  10,
			wantLines: 1,
		},
		{
			name:      "wrap needed",
			text:      "Hello World This is a long text",
			maxWidth:  50,
			fontSize:  10,
			wantLines: 5, // Each word wraps to its own line at 50px width with 10pt font
		},
		{
			name:      "with newlines",
			text:      "Line1\nLine2\nLine3",
			maxWidth:  100,
			fontSize:  10,
			wantLines: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := Wrap(tt.text, tt.maxWidth, "Helvetica", tt.fontSize, DefaultWidthEstimator)
			if len(lines) != tt.wantLines {
				t.Errorf("Wrap() = %d lines, want %d lines", len(lines), tt.wantLines)
			}
		})
	}
}

func TestFit(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		bounds  layout.Rectangle
		opts    FitOptions
		wantErr bool
	}{
		{
			name: "fits without change",
			text: "Hello",
			bounds: layout.Rectangle{
				X: 0, Y: 0, Width: 100, Height: 50,
			},
			opts:    DefaultFitOptions(),
			wantErr: false,
		},
		{
			name: "needs shrink",
			text: "This is a long text that needs wrapping",
			bounds: layout.Rectangle{
				X: 0, Y: 0, Width: 100, Height: 80,
			},
			opts:    DefaultFitOptions(),
			wantErr: false,
		},
		{
			name: "bounds too small",
			text: "Hello",
			bounds: layout.Rectangle{
				X: 0, Y: 0, Width: 1, Height: 1,
			},
			opts:    DefaultFitOptions(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Fit(tt.text, tt.bounds, "Helvetica", tt.opts, DefaultWidthEstimator)
			if (err != nil) != tt.wantErr {
				t.Errorf("Fit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAlign(t *testing.T) {
	tests := []struct {
		name  string
		align Align
		want  int
	}{
		{"AlignLeft", AlignLeft, 0},
		{"AlignCenter", AlignCenter, 1},
		{"AlignRight", AlignRight, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.align) != tt.want {
				t.Errorf("Align = %d, want %d", tt.align, tt.want)
			}
		})
	}
}
