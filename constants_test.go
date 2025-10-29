package gopdf

import (
	"math"
	"testing"
)

func TestPageSizes(t *testing.T) {
	tests := []struct {
		name           string
		pageSize       PageSize
		expectedWidth  float64
		expectedHeight float64
	}{
		{
			name:           "A4",
			pageSize:       PageSizeA4,
			expectedWidth:  595.0,
			expectedHeight: 842.0,
		},
		{
			name:           "Letter",
			pageSize:       PageSizeLetter,
			expectedWidth:  612.0,
			expectedHeight: 792.0,
		},
		{
			name:           "Legal",
			pageSize:       PageSizeLegal,
			expectedWidth:  612.0,
			expectedHeight: 1008.0,
		},
		{
			name:           "A3",
			pageSize:       PageSizeA3,
			expectedWidth:  842.0,
			expectedHeight: 1191.0,
		},
		{
			name:           "A5",
			pageSize:       PageSizeA5,
			expectedWidth:  420.0,
			expectedHeight: 595.0,
		},
		{
			name:           "Presentation 16:9",
			pageSize:       PageSizePresentation16x9,
			expectedWidth:  720.0,
			expectedHeight: 405.0,
		},
		{
			name:           "Presentation 4:3",
			pageSize:       PageSizePresentation4x3,
			expectedWidth:  720.0,
			expectedHeight: 540.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.pageSize.Width != tt.expectedWidth {
				t.Errorf("Width = %v, want %v", tt.pageSize.Width, tt.expectedWidth)
			}
			if tt.pageSize.Height != tt.expectedHeight {
				t.Errorf("Height = %v, want %v", tt.pageSize.Height, tt.expectedHeight)
			}
		})
	}
}

func TestPresentationAspectRatios(t *testing.T) {
	tests := []struct {
		name          string
		pageSize      PageSize
		expectedRatio float64
		tolerance     float64
	}{
		{
			name:          "16:9 Widescreen",
			pageSize:      PageSizePresentation16x9,
			expectedRatio: 16.0 / 9.0,
			tolerance:     0.01,
		},
		{
			name:          "4:3 Standard",
			pageSize:      PageSizePresentation4x3,
			expectedRatio: 4.0 / 3.0,
			tolerance:     0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualRatio := tt.pageSize.Width / tt.pageSize.Height
			diff := math.Abs(actualRatio - tt.expectedRatio)
			if diff > tt.tolerance {
				t.Errorf("Aspect ratio = %v, want %v (diff: %v)", actualRatio, tt.expectedRatio, diff)
			}
		})
	}
}

func TestOrientation(t *testing.T) {
	tests := []struct {
		name        string
		orientation Orientation
		pageSize    PageSize
		wantWidth   float64
		wantHeight  float64
	}{
		{
			name:        "Portrait A4",
			orientation: Portrait,
			pageSize:    PageSizeA4,
			wantWidth:   595.0,
			wantHeight:  842.0,
		},
		{
			name:        "Landscape A4",
			orientation: Landscape,
			pageSize:    PageSizeA4,
			wantWidth:   842.0,
			wantHeight:  595.0,
		},
		{
			name:        "Portrait Presentation 16:9",
			orientation: Portrait,
			pageSize:    PageSizePresentation16x9,
			wantWidth:   720.0,
			wantHeight:  405.0,
		},
		{
			name:        "Landscape Presentation 16:9",
			orientation: Landscape,
			pageSize:    PageSizePresentation16x9,
			wantWidth:   405.0,
			wantHeight:  720.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.orientation.Apply(tt.pageSize)
			if result.Width != tt.wantWidth {
				t.Errorf("Width = %v, want %v", result.Width, tt.wantWidth)
			}
			if result.Height != tt.wantHeight {
				t.Errorf("Height = %v, want %v", result.Height, tt.wantHeight)
			}
		})
	}
}
