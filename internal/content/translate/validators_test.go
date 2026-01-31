package translate

import (
	"testing"
)

func TestValidateImagePosition(t *testing.T) {
	tests := []struct {
		name       string
		x, y       float64
		width      float64
		height     float64
		pageWidth  float64
		pageHeight float64
		wantX      float64
		wantY      float64
	}{
		{
			name:       "valid position",
			x:          100,
			y:          200,
			width:      50,
			height:     50,
			pageWidth:  595,
			pageHeight: 842,
			wantX:      100,
			wantY:      200,
		},
		{
			name:       "x too negative",
			x:          -20000,
			y:          200,
			width:      100,
			height:     100,
			pageWidth:  595,
			pageHeight: 842,
			wantX:      247.5, // (595 - 100) / 2
			wantY:      692,   // 842 - 100 - 50
		},
		{
			name:       "x too positive",
			x:          20000,
			y:          200,
			width:      100,
			height:     100,
			pageWidth:  595,
			pageHeight: 842,
			wantX:      247.5,
			wantY:      692,
		},
		{
			name:       "y too negative",
			x:          100,
			y:          -20000,
			width:      100,
			height:     100,
			pageWidth:  595,
			pageHeight: 842,
			wantX:      247.5,
			wantY:      692,
		},
		{
			name:       "y too positive",
			x:          100,
			y:          20000,
			width:      100,
			height:     100,
			pageWidth:  595,
			pageHeight: 842,
			wantX:      247.5,
			wantY:      692,
		},
		{
			name:       "large image fallback clamps",
			x:          -20000,
			y:          200,
			width:      1000,
			height:     1000,
			pageWidth:  595,
			pageHeight: 842,
			wantX:      0,   // (595 - 1000) / 2 = -202.5, clamped to 0
			wantY:      0,   // 842 - 1000 - 50 = -208, clamped to 0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotX, gotY := ValidateImagePosition(tt.x, tt.y, tt.width, tt.height, tt.pageWidth, tt.pageHeight)
			if gotX != tt.wantX {
				t.Errorf("ValidateImagePosition() gotX = %v, want %v", gotX, tt.wantX)
			}
			if gotY != tt.wantY {
				t.Errorf("ValidateImagePosition() gotY = %v, want %v", gotY, tt.wantY)
			}
		})
	}
}
