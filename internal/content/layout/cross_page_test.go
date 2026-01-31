package layout

import (
	"testing"

	"github.com/ryomak/gopdf/layout"
)

func TestFlattenContentBlocks(t *testing.T) {
	tb1 := layout.TextBlock{Text: "Page1Block1"}
	tb2 := layout.TextBlock{Text: "Page1Block2"}
	tb3 := layout.TextBlock{Text: "Page2Block1"}

	tests := []struct {
		name       string
		pageBlocks map[int][]layout.ContentBlock
		wantLen    int
	}{
		{
			name:       "empty map",
			pageBlocks: nil,
			wantLen:    0,
		},
		{
			name: "single page",
			pageBlocks: map[int][]layout.ContentBlock{
				0: {tb1, tb2},
			},
			wantLen: 2,
		},
		{
			name: "multiple pages",
			pageBlocks: map[int][]layout.ContentBlock{
				0: {tb1, tb2},
				1: {tb3},
			},
			wantLen: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FlattenContentBlocks(tt.pageBlocks)
			if len(got) != tt.wantLen {
				t.Errorf("FlattenContentBlocks() length = %d, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestCanMergeTextBlocks(t *testing.T) {
	tests := []struct {
		name   string
		block1 layout.TextBlock
		block2 layout.TextBlock
		want   bool
	}{
		{
			name: "same font and size",
			block1: layout.TextBlock{
				Font:     "Helvetica",
				FontSize: 12,
				Color:    layout.Color{R: 0, G: 0, B: 0},
			},
			block2: layout.TextBlock{
				Font:     "Helvetica",
				FontSize: 12,
				Color:    layout.Color{R: 0, G: 0, B: 0},
			},
			want: true,
		},
		{
			name: "different font",
			block1: layout.TextBlock{
				Font:     "Helvetica",
				FontSize: 12,
				Color:    layout.Color{R: 0, G: 0, B: 0},
			},
			block2: layout.TextBlock{
				Font:     "Times-Roman",
				FontSize: 12,
				Color:    layout.Color{R: 0, G: 0, B: 0},
			},
			want: false,
		},
		{
			name: "size difference within tolerance",
			block1: layout.TextBlock{
				Font:     "Helvetica",
				FontSize: 12,
				Color:    layout.Color{R: 0, G: 0, B: 0},
			},
			block2: layout.TextBlock{
				Font:     "Helvetica",
				FontSize: 12.5,
				Color:    layout.Color{R: 0, G: 0, B: 0},
			},
			want: true,
		},
		{
			name: "size difference beyond tolerance",
			block1: layout.TextBlock{
				Font:     "Helvetica",
				FontSize: 12,
				Color:    layout.Color{R: 0, G: 0, B: 0},
			},
			block2: layout.TextBlock{
				Font:     "Helvetica",
				FontSize: 14,
				Color:    layout.Color{R: 0, G: 0, B: 0},
			},
			want: false,
		},
		{
			name: "different color",
			block1: layout.TextBlock{
				Font:     "Helvetica",
				FontSize: 12,
				Color:    layout.Color{R: 0, G: 0, B: 0},
			},
			block2: layout.TextBlock{
				Font:     "Helvetica",
				FontSize: 12,
				Color:    layout.Color{R: 1, G: 0, B: 0},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CanMergeTextBlocks(tt.block1, tt.block2)
			if got != tt.want {
				t.Errorf("CanMergeTextBlocks() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMergeContentBlocksAcrossPages(t *testing.T) {
	tb1 := layout.TextBlock{
		Text:     "Block1",
		Font:     "Helvetica",
		FontSize: 12,
		Color:    layout.Color{R: 0, G: 0, B: 0},
	}
	tb2 := layout.TextBlock{
		Text:     "Block2",
		Font:     "Helvetica",
		FontSize: 12,
		Color:    layout.Color{R: 0, G: 0, B: 0},
	}
	tb3 := layout.TextBlock{
		Text:     "Block3",
		Font:     "Times-Roman", // Different font
		FontSize: 12,
		Color:    layout.Color{R: 0, G: 0, B: 0},
	}

	tests := []struct {
		name       string
		pageBlocks map[int][]layout.ContentBlock
		wantLen    int
	}{
		{
			name:       "empty map",
			pageBlocks: nil,
			wantLen:    0,
		},
		{
			name: "mergeable blocks",
			pageBlocks: map[int][]layout.ContentBlock{
				0: {tb1, tb2},
			},
			wantLen: 1, // tb1 and tb2 merged
		},
		{
			name: "non-mergeable blocks",
			pageBlocks: map[int][]layout.ContentBlock{
				0: {tb1, tb3},
			},
			wantLen: 2, // Different fonts, not merged
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeContentBlocksAcrossPages(tt.pageBlocks)
			if len(got) != tt.wantLen {
				t.Errorf("MergeContentBlocksAcrossPages() length = %d, want %d", len(got), tt.wantLen)
			}
		})
	}
}
