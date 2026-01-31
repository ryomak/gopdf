package layout

import (
	"testing"
)

func TestGetImageYRanges(t *testing.T) {
	tests := []struct {
		name   string
		images []ImageBlock
		want   []YRange
	}{
		{
			name:   "empty images",
			images: nil,
			want:   nil,
		},
		{
			name: "single image",
			images: []ImageBlock{
				{Y: 100, PlacedHeight: 50},
			},
			want: []YRange{
				{Min: 100, Max: 150},
			},
		},
		{
			name: "multiple images",
			images: []ImageBlock{
				{Y: 100, PlacedHeight: 50},
				{Y: 200, PlacedHeight: 30},
			},
			want: []YRange{
				{Min: 100, Max: 150},
				{Min: 200, Max: 230},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetImageYRanges(tt.images)
			if len(got) != len(tt.want) {
				t.Errorf("GetImageYRanges() length = %d, want %d", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("GetImageYRanges()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestOverlapsYRange(t *testing.T) {
	tests := []struct {
		name   string
		range1 YRange
		range2 YRange
		want   bool
	}{
		{
			name:   "overlapping ranges",
			range1: YRange{Min: 0, Max: 100},
			range2: YRange{Min: 50, Max: 150},
			want:   true,
		},
		{
			name:   "non-overlapping ranges",
			range1: YRange{Min: 0, Max: 50},
			range2: YRange{Min: 100, Max: 150},
			want:   false,
		},
		{
			name:   "touching ranges",
			range1: YRange{Min: 0, Max: 50},
			range2: YRange{Min: 50, Max: 100},
			want:   true,
		},
		{
			name:   "contained range",
			range1: YRange{Min: 0, Max: 100},
			range2: YRange{Min: 25, Max: 75},
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := OverlapsYRange(tt.range1, tt.range2)
			if got != tt.want {
				t.Errorf("OverlapsYRange() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasImageBetween(t *testing.T) {
	tests := []struct {
		name        string
		prevLine    []TextElement
		currLine    []TextElement
		imageRanges []YRange
		want        bool
	}{
		{
			name:        "empty prev line",
			prevLine:    nil,
			currLine:    []TextElement{{Y: 50, Height: 12}},
			imageRanges: []YRange{{Min: 60, Max: 80}},
			want:        false,
		},
		{
			name:        "empty curr line",
			prevLine:    []TextElement{{Y: 100, Height: 12}},
			currLine:    nil,
			imageRanges: []YRange{{Min: 60, Max: 80}},
			want:        false,
		},
		{
			name:        "no images",
			prevLine:    []TextElement{{Y: 100, Height: 12}},
			currLine:    []TextElement{{Y: 50, Height: 12}},
			imageRanges: nil,
			want:        false,
		},
		{
			name:        "image between lines",
			prevLine:    []TextElement{{Y: 100, Height: 12}},
			currLine:    []TextElement{{Y: 50, Height: 12}},
			imageRanges: []YRange{{Min: 65, Max: 95}}, // Between 62 and 100
			want:        true,
		},
		{
			name:        "image not between lines",
			prevLine:    []TextElement{{Y: 100, Height: 12}},
			currLine:    []TextElement{{Y: 80, Height: 12}},
			imageRanges: []YRange{{Min: 10, Max: 20}},
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasImageBetween(tt.prevLine, tt.currLine, tt.imageRanges)
			if got != tt.want {
				t.Errorf("HasImageBetween() = %v, want %v", got, tt.want)
			}
		})
	}
}
