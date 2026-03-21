package layout

import (
	"math"
	"testing"
)

func TestMoveBlock(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() *PageLayout
		blockType ContentBlockType
		index     int
		offsetX   float64
		offsetY   float64
		wantErr   bool
		validate  func(t *testing.T, pl *PageLayout)
	}{
		{
			name: "move text block",
			setup: func() *PageLayout {
				return &PageLayout{
					TextBlocks: []TextBlock{
						{Text: "Hello", Rect: Rectangle{X: 10, Y: 100, Width: 50, Height: 12}},
					},
				}
			},
			blockType: ContentBlockTypeText,
			index:     0,
			offsetX:   5,
			offsetY:   -10,
			wantErr:   false,
			validate: func(t *testing.T, pl *PageLayout) {
				t.Helper()
				if pl.TextBlocks[0].Rect.X != 15 {
					t.Errorf("X = %v, want 15", pl.TextBlocks[0].Rect.X)
				}
				if pl.TextBlocks[0].Rect.Y != 90 {
					t.Errorf("Y = %v, want 90", pl.TextBlocks[0].Rect.Y)
				}
			},
		},
		{
			name: "move image block",
			setup: func() *PageLayout {
				return &PageLayout{
					Images: []ImageBlock{
						{X: 20, Y: 200, PlacedWidth: 100, PlacedHeight: 80},
					},
				}
			},
			blockType: ContentBlockTypeImage,
			index:     0,
			offsetX:   10,
			offsetY:   20,
			wantErr:   false,
			validate: func(t *testing.T, pl *PageLayout) {
				t.Helper()
				if pl.Images[0].X != 30 {
					t.Errorf("X = %v, want 30", pl.Images[0].X)
				}
				if pl.Images[0].Y != 220 {
					t.Errorf("Y = %v, want 220", pl.Images[0].Y)
				}
			},
		},
		{
			name: "text block index out of range negative",
			setup: func() *PageLayout {
				return &PageLayout{
					TextBlocks: []TextBlock{
						{Text: "Hello", Rect: Rectangle{X: 10, Y: 100, Width: 50, Height: 12}},
					},
				}
			},
			blockType: ContentBlockTypeText,
			index:     -1,
			offsetX:   0,
			offsetY:   0,
			wantErr:   true,
		},
		{
			name: "text block index out of range high",
			setup: func() *PageLayout {
				return &PageLayout{
					TextBlocks: []TextBlock{
						{Text: "Hello", Rect: Rectangle{X: 10, Y: 100, Width: 50, Height: 12}},
					},
				}
			},
			blockType: ContentBlockTypeText,
			index:     5,
			offsetX:   0,
			offsetY:   0,
			wantErr:   true,
		},
		{
			name: "image block index out of range negative",
			setup: func() *PageLayout {
				return &PageLayout{
					Images: []ImageBlock{
						{X: 20, Y: 200, PlacedWidth: 100, PlacedHeight: 80},
					},
				}
			},
			blockType: ContentBlockTypeImage,
			index:     -1,
			offsetX:   0,
			offsetY:   0,
			wantErr:   true,
		},
		{
			name: "image block index out of range high",
			setup: func() *PageLayout {
				return &PageLayout{
					Images: []ImageBlock{
						{X: 20, Y: 200, PlacedWidth: 100, PlacedHeight: 80},
					},
				}
			},
			blockType: ContentBlockTypeImage,
			index:     2,
			offsetX:   0,
			offsetY:   0,
			wantErr:   true,
		},
		{
			name: "unsupported block type",
			setup: func() *PageLayout {
				return &PageLayout{}
			},
			blockType: ContentBlockType("unknown"),
			index:     0,
			offsetX:   0,
			offsetY:   0,
			wantErr:   true,
		},
		{
			name: "move text block with zero offset",
			setup: func() *PageLayout {
				return &PageLayout{
					TextBlocks: []TextBlock{
						{Text: "Stay", Rect: Rectangle{X: 10, Y: 100, Width: 50, Height: 12}},
					},
				}
			},
			blockType: ContentBlockTypeText,
			index:     0,
			offsetX:   0,
			offsetY:   0,
			wantErr:   false,
			validate: func(t *testing.T, pl *PageLayout) {
				t.Helper()
				if pl.TextBlocks[0].Rect.X != 10 {
					t.Errorf("X = %v, want 10", pl.TextBlocks[0].Rect.X)
				}
				if pl.TextBlocks[0].Rect.Y != 100 {
					t.Errorf("Y = %v, want 100", pl.TextBlocks[0].Rect.Y)
				}
			},
		},
		{
			name: "move second text block among multiple",
			setup: func() *PageLayout {
				return &PageLayout{
					TextBlocks: []TextBlock{
						{Text: "First", Rect: Rectangle{X: 10, Y: 200, Width: 50, Height: 12}},
						{Text: "Second", Rect: Rectangle{X: 10, Y: 100, Width: 50, Height: 12}},
					},
				}
			},
			blockType: ContentBlockTypeText,
			index:     1,
			offsetX:   20,
			offsetY:   30,
			wantErr:   false,
			validate: func(t *testing.T, pl *PageLayout) {
				t.Helper()
				// First block unchanged
				if pl.TextBlocks[0].Rect.X != 10 || pl.TextBlocks[0].Rect.Y != 200 {
					t.Errorf("First block should be unchanged")
				}
				// Second block moved
				if pl.TextBlocks[1].Rect.X != 30 {
					t.Errorf("X = %v, want 30", pl.TextBlocks[1].Rect.X)
				}
				if pl.TextBlocks[1].Rect.Y != 130 {
					t.Errorf("Y = %v, want 130", pl.TextBlocks[1].Rect.Y)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pl := tt.setup()
			err := pl.MoveBlock(tt.blockType, tt.index, tt.offsetX, tt.offsetY)
			if (err != nil) != tt.wantErr {
				t.Errorf("MoveBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.validate != nil && err == nil {
				tt.validate(t, pl)
			}
		})
	}
}

func TestResizeBlock(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() *PageLayout
		blockType ContentBlockType
		index     int
		newWidth  float64
		newHeight float64
		wantErr   bool
		validate  func(t *testing.T, pl *PageLayout)
	}{
		{
			name: "resize text block",
			setup: func() *PageLayout {
				return &PageLayout{
					TextBlocks: []TextBlock{
						{Text: "Hello", Rect: Rectangle{X: 10, Y: 100, Width: 50, Height: 12}},
					},
				}
			},
			blockType: ContentBlockTypeText,
			index:     0,
			newWidth:  100,
			newHeight: 24,
			wantErr:   false,
			validate: func(t *testing.T, pl *PageLayout) {
				t.Helper()
				if pl.TextBlocks[0].Rect.Width != 100 {
					t.Errorf("Width = %v, want 100", pl.TextBlocks[0].Rect.Width)
				}
				if pl.TextBlocks[0].Rect.Height != 24 {
					t.Errorf("Height = %v, want 24", pl.TextBlocks[0].Rect.Height)
				}
				// Position unchanged
				if pl.TextBlocks[0].Rect.X != 10 || pl.TextBlocks[0].Rect.Y != 100 {
					t.Error("Position should be unchanged")
				}
			},
		},
		{
			name: "resize image block",
			setup: func() *PageLayout {
				return &PageLayout{
					Images: []ImageBlock{
						{X: 20, Y: 200, PlacedWidth: 100, PlacedHeight: 80},
					},
				}
			},
			blockType: ContentBlockTypeImage,
			index:     0,
			newWidth:  200,
			newHeight: 160,
			wantErr:   false,
			validate: func(t *testing.T, pl *PageLayout) {
				t.Helper()
				if pl.Images[0].PlacedWidth != 200 {
					t.Errorf("PlacedWidth = %v, want 200", pl.Images[0].PlacedWidth)
				}
				if pl.Images[0].PlacedHeight != 160 {
					t.Errorf("PlacedHeight = %v, want 160", pl.Images[0].PlacedHeight)
				}
				// Position unchanged
				if pl.Images[0].X != 20 || pl.Images[0].Y != 200 {
					t.Error("Position should be unchanged")
				}
			},
		},
		{
			name: "text block index out of range",
			setup: func() *PageLayout {
				return &PageLayout{
					TextBlocks: []TextBlock{
						{Text: "Hello", Rect: Rectangle{X: 10, Y: 100, Width: 50, Height: 12}},
					},
				}
			},
			blockType: ContentBlockTypeText,
			index:     3,
			newWidth:  100,
			newHeight: 24,
			wantErr:   true,
		},
		{
			name: "text block negative index",
			setup: func() *PageLayout {
				return &PageLayout{
					TextBlocks: []TextBlock{
						{Text: "Hello", Rect: Rectangle{X: 10, Y: 100, Width: 50, Height: 12}},
					},
				}
			},
			blockType: ContentBlockTypeText,
			index:     -1,
			newWidth:  100,
			newHeight: 24,
			wantErr:   true,
		},
		{
			name: "image block index out of range",
			setup: func() *PageLayout {
				return &PageLayout{
					Images: []ImageBlock{
						{X: 20, Y: 200, PlacedWidth: 100, PlacedHeight: 80},
					},
				}
			},
			blockType: ContentBlockTypeImage,
			index:     1,
			newWidth:  200,
			newHeight: 160,
			wantErr:   true,
		},
		{
			name: "unsupported block type",
			setup: func() *PageLayout {
				return &PageLayout{}
			},
			blockType: ContentBlockType("custom"),
			index:     0,
			newWidth:  100,
			newHeight: 50,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pl := tt.setup()
			err := pl.ResizeBlock(tt.blockType, tt.index, tt.newWidth, tt.newHeight)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResizeBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.validate != nil && err == nil {
				tt.validate(t, pl)
			}
		})
	}
}

func TestDetectOverlaps(t *testing.T) {
	tests := []struct {
		name         string
		setup        func() *PageLayout
		wantOverlaps int
		validate     func(t *testing.T, overlaps []BlockOverlap)
	}{
		{
			name: "no blocks",
			setup: func() *PageLayout {
				return &PageLayout{Width: 612, Height: 792}
			},
			wantOverlaps: 0,
		},
		{
			name: "single text block no overlap",
			setup: func() *PageLayout {
				return &PageLayout{
					Width:  612,
					Height: 792,
					TextBlocks: []TextBlock{
						{Text: "Hello", Rect: Rectangle{X: 10, Y: 100, Width: 50, Height: 12}},
					},
				}
			},
			wantOverlaps: 0,
		},
		{
			name: "two non-overlapping text blocks",
			setup: func() *PageLayout {
				return &PageLayout{
					Width:  612,
					Height: 792,
					TextBlocks: []TextBlock{
						{Text: "Hello", Rect: Rectangle{X: 10, Y: 100, Width: 50, Height: 12}},
						{Text: "World", Rect: Rectangle{X: 10, Y: 50, Width: 50, Height: 12}},
					},
				}
			},
			wantOverlaps: 0,
		},
		{
			name: "two overlapping text blocks",
			setup: func() *PageLayout {
				return &PageLayout{
					Width:  612,
					Height: 792,
					TextBlocks: []TextBlock{
						{Text: "Hello", Rect: Rectangle{X: 10, Y: 100, Width: 50, Height: 20}},
						{Text: "World", Rect: Rectangle{X: 30, Y: 110, Width: 50, Height: 20}},
					},
				}
			},
			wantOverlaps: 1,
			validate: func(t *testing.T, overlaps []BlockOverlap) {
				t.Helper()
				if overlaps[0].Area <= 0 {
					t.Error("Expected positive overlap area")
				}
				// Overlap region: X [30,60] Y [110,120] = 30 * 10 = 300
				expected := 300.0
				if math.Abs(overlaps[0].Area-expected) > 0.01 {
					t.Errorf("Area = %v, want %v", overlaps[0].Area, expected)
				}
			},
		},
		{
			name: "text and image overlap",
			setup: func() *PageLayout {
				return &PageLayout{
					Width:  612,
					Height: 792,
					TextBlocks: []TextBlock{
						{Text: "Hello", Rect: Rectangle{X: 10, Y: 100, Width: 100, Height: 50}},
					},
					Images: []ImageBlock{
						{X: 50, Y: 120, PlacedWidth: 100, PlacedHeight: 50},
					},
				}
			},
			wantOverlaps: 1,
			validate: func(t *testing.T, overlaps []BlockOverlap) {
				t.Helper()
				// Overlap: X [50,110] Y [120,150] = 60 * 30 = 1800
				expected := 1800.0
				if math.Abs(overlaps[0].Area-expected) > 0.01 {
					t.Errorf("Area = %v, want %v", overlaps[0].Area, expected)
				}
			},
		},
		{
			name: "three blocks with multiple overlaps",
			setup: func() *PageLayout {
				return &PageLayout{
					Width:  612,
					Height: 792,
					TextBlocks: []TextBlock{
						{Text: "A", Rect: Rectangle{X: 0, Y: 0, Width: 100, Height: 100}},
						{Text: "B", Rect: Rectangle{X: 50, Y: 50, Width: 100, Height: 100}},
						{Text: "C", Rect: Rectangle{X: 80, Y: 80, Width: 100, Height: 100}},
					},
				}
			},
			wantOverlaps: 3, // A-B, A-C, B-C
		},
		{
			name: "adjacent blocks touching edges no overlap",
			setup: func() *PageLayout {
				return &PageLayout{
					Width:  612,
					Height: 792,
					TextBlocks: []TextBlock{
						{Text: "A", Rect: Rectangle{X: 0, Y: 0, Width: 50, Height: 50}},
						{Text: "B", Rect: Rectangle{X: 50, Y: 0, Width: 50, Height: 50}},
					},
				}
			},
			wantOverlaps: 0, // Touching but not overlapping (width=0)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pl := tt.setup()
			overlaps := pl.DetectOverlaps()
			if len(overlaps) != tt.wantOverlaps {
				t.Errorf("DetectOverlaps() got %d overlaps, want %d", len(overlaps), tt.wantOverlaps)
				return
			}
			if tt.validate != nil {
				tt.validate(t, overlaps)
			}
		})
	}
}

func TestSplitIntoPages(t *testing.T) {
	tests := []struct {
		name       string
		setup      func() *PageLayout
		maxHeight  float64
		minSpacing float64
		pageMargin float64
		wantPages  int
		validate   func(t *testing.T, pages []*PageLayout)
	}{
		{
			name: "single block fits on one page",
			setup: func() *PageLayout {
				return &PageLayout{
					Width:  612,
					Height: 792,
					TextBlocks: []TextBlock{
						{Text: "Hello", Rect: Rectangle{X: 10, Y: 700, Width: 50, Height: 12}},
					},
				}
			},
			maxHeight:  792,
			minSpacing: 10,
			pageMargin: 20,
			wantPages:  1,
			validate: func(t *testing.T, pages []*PageLayout) {
				t.Helper()
				if len(pages[0].TextBlocks) != 1 {
					t.Errorf("Expected 1 text block on page, got %d", len(pages[0].TextBlocks))
				}
			},
		},
		{
			name: "empty layout produces single empty page",
			setup: func() *PageLayout {
				return &PageLayout{
					Width:  612,
					Height: 792,
				}
			},
			maxHeight:  792,
			minSpacing: 10,
			pageMargin: 20,
			wantPages:  1,
		},
		{
			name: "blocks overflow to second page",
			setup: func() *PageLayout {
				return &PageLayout{
					Width:  612,
					Height: 100,
					TextBlocks: []TextBlock{
						{Text: "Block1", Rect: Rectangle{X: 10, Y: 70, Width: 50, Height: 30}},
						{Text: "Block2", Rect: Rectangle{X: 10, Y: 30, Width: 50, Height: 30}},
					},
				}
			},
			maxHeight:  100,
			minSpacing: 10,
			pageMargin: 20,
			wantPages:  2,
			validate: func(t *testing.T, pages []*PageLayout) {
				t.Helper()
				if len(pages[0].TextBlocks) != 1 {
					t.Errorf("Expected 1 text block on first page, got %d", len(pages[0].TextBlocks))
				}
				if len(pages[1].TextBlocks) != 1 {
					t.Errorf("Expected 1 text block on second page, got %d", len(pages[1].TextBlocks))
				}
			},
		},
		{
			name: "image blocks split across pages",
			setup: func() *PageLayout {
				return &PageLayout{
					Width:  612,
					Height: 100,
					Images: []ImageBlock{
						{X: 10, Y: 70, PlacedWidth: 50, PlacedHeight: 30},
						{X: 10, Y: 30, PlacedWidth: 50, PlacedHeight: 30},
					},
				}
			},
			maxHeight:  100,
			minSpacing: 10,
			pageMargin: 20,
			wantPages:  2,
			validate: func(t *testing.T, pages []*PageLayout) {
				t.Helper()
				if len(pages[0].Images) != 1 {
					t.Errorf("Expected 1 image on first page, got %d", len(pages[0].Images))
				}
				if len(pages[1].Images) != 1 {
					t.Errorf("Expected 1 image on second page, got %d", len(pages[1].Images))
				}
			},
		},
		{
			name: "mixed text and image blocks",
			setup: func() *PageLayout {
				return &PageLayout{
					Width:  612,
					Height: 100,
					TextBlocks: []TextBlock{
						{Text: "Hello", Rect: Rectangle{X: 10, Y: 70, Width: 50, Height: 30}},
					},
					Images: []ImageBlock{
						{X: 10, Y: 30, PlacedWidth: 50, PlacedHeight: 30},
					},
				}
			},
			maxHeight:  100,
			minSpacing: 10,
			pageMargin: 20,
			wantPages:  2,
			validate: func(t *testing.T, pages []*PageLayout) {
				t.Helper()
				// One page has text, the other has image
				totalText := len(pages[0].TextBlocks) + len(pages[1].TextBlocks)
				totalImages := len(pages[0].Images) + len(pages[1].Images)
				if totalText != 1 {
					t.Errorf("Expected 1 total text block, got %d", totalText)
				}
				if totalImages != 1 {
					t.Errorf("Expected 1 total image, got %d", totalImages)
				}
			},
		},
		{
			name: "all blocks fit on single page",
			setup: func() *PageLayout {
				return &PageLayout{
					Width:  612,
					Height: 792,
					TextBlocks: []TextBlock{
						{Text: "A", Rect: Rectangle{X: 10, Y: 700, Width: 50, Height: 12}},
						{Text: "B", Rect: Rectangle{X: 10, Y: 680, Width: 50, Height: 12}},
						{Text: "C", Rect: Rectangle{X: 10, Y: 660, Width: 50, Height: 12}},
					},
				}
			},
			maxHeight:  792,
			minSpacing: 5,
			pageMargin: 10,
			wantPages:  1,
			validate: func(t *testing.T, pages []*PageLayout) {
				t.Helper()
				if len(pages[0].TextBlocks) != 3 {
					t.Errorf("Expected 3 text blocks on page, got %d", len(pages[0].TextBlocks))
				}
			},
		},
		{
			name: "page width preserved in split pages",
			setup: func() *PageLayout {
				return &PageLayout{
					Width:  500,
					Height: 100,
					TextBlocks: []TextBlock{
						{Text: "A", Rect: Rectangle{X: 10, Y: 70, Width: 50, Height: 30}},
						{Text: "B", Rect: Rectangle{X: 10, Y: 30, Width: 50, Height: 30}},
					},
				}
			},
			maxHeight:  100,
			minSpacing: 10,
			pageMargin: 20,
			wantPages:  2,
			validate: func(t *testing.T, pages []*PageLayout) {
				t.Helper()
				for i, p := range pages {
					if p.Width != 500 {
						t.Errorf("Page %d width = %v, want 500", i, p.Width)
					}
					if p.Height != 100 {
						t.Errorf("Page %d height = %v, want 100", i, p.Height)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pl := tt.setup()
			pages, err := pl.SplitIntoPages(tt.maxHeight, tt.minSpacing, tt.pageMargin)
			if err != nil {
				t.Fatalf("SplitIntoPages() error = %v", err)
			}
			if len(pages) != tt.wantPages {
				t.Errorf("SplitIntoPages() got %d pages, want %d", len(pages), tt.wantPages)
				return
			}
			if tt.validate != nil {
				tt.validate(t, pages)
			}
		})
	}
}

func TestCalculateOverlapArea(t *testing.T) {
	tests := []struct {
		name  string
		b1    ContentBlock
		b2    ContentBlock
		want  float64
	}{
		{
			name: "no overlap separated horizontally",
			b1:   TextBlock{Rect: Rectangle{X: 0, Y: 0, Width: 10, Height: 10}},
			b2:   TextBlock{Rect: Rectangle{X: 20, Y: 0, Width: 10, Height: 10}},
			want: 0,
		},
		{
			name: "no overlap separated vertically",
			b1:   TextBlock{Rect: Rectangle{X: 0, Y: 0, Width: 10, Height: 10}},
			b2:   TextBlock{Rect: Rectangle{X: 0, Y: 20, Width: 10, Height: 10}},
			want: 0,
		},
		{
			name: "full overlap identical blocks",
			b1:   TextBlock{Rect: Rectangle{X: 0, Y: 0, Width: 10, Height: 10}},
			b2:   TextBlock{Rect: Rectangle{X: 0, Y: 0, Width: 10, Height: 10}},
			want: 100, // 10*10
		},
		{
			name: "partial overlap",
			b1:   TextBlock{Rect: Rectangle{X: 0, Y: 0, Width: 10, Height: 10}},
			b2:   TextBlock{Rect: Rectangle{X: 5, Y: 5, Width: 10, Height: 10}},
			want: 25, // 5*5
		},
		{
			name: "one block contained in another",
			b1:   TextBlock{Rect: Rectangle{X: 0, Y: 0, Width: 100, Height: 100}},
			b2:   TextBlock{Rect: Rectangle{X: 10, Y: 10, Width: 20, Height: 20}},
			want: 400, // 20*20
		},
		{
			name: "image and text overlap",
			b1:   TextBlock{Rect: Rectangle{X: 0, Y: 0, Width: 50, Height: 50}},
			b2:   ImageBlock{X: 25, Y: 25, PlacedWidth: 50, PlacedHeight: 50},
			want: 625, // 25*25
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateOverlapArea(tt.b1, tt.b2)
			if math.Abs(got-tt.want) > 0.01 {
				t.Errorf("calculateOverlapArea() = %v, want %v", got, tt.want)
			}
		})
	}
}
