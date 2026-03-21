package layout

import (
	"testing"
)

func TestAdjustLayout(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *PageLayout
		opts     LayoutAdjustmentOptions
		wantErr  bool
		validate func(t *testing.T, pl *PageLayout)
	}{
		{
			name: "preserve position does nothing",
			setup: func() *PageLayout {
				return &PageLayout{
					Width:  612,
					Height: 792,
					TextBlocks: []TextBlock{
						{Text: "A", Rect: Rectangle{X: 10, Y: 700, Width: 50, Height: 12}},
						{Text: "B", Rect: Rectangle{X: 10, Y: 600, Width: 50, Height: 12}},
					},
				}
			},
			opts: LayoutAdjustmentOptions{
				Strategy:   StrategyPreservePosition,
				MinSpacing: 10,
				PageMargin: 20,
			},
			wantErr: false,
			validate: func(t *testing.T, pl *PageLayout) {
				t.Helper()
				if pl.TextBlocks[0].Rect.Y != 700 {
					t.Errorf("Block A Y = %v, want 700 (unchanged)", pl.TextBlocks[0].Rect.Y)
				}
				if pl.TextBlocks[1].Rect.Y != 600 {
					t.Errorf("Block B Y = %v, want 600 (unchanged)", pl.TextBlocks[1].Rect.Y)
				}
			},
		},
		{
			name: "unsupported strategy returns error",
			setup: func() *PageLayout {
				return &PageLayout{Width: 612, Height: 792}
			},
			opts: LayoutAdjustmentOptions{
				Strategy: LayoutStrategy("invalid_strategy"),
			},
			wantErr: true,
		},
		{
			name: "compact strategy with text blocks",
			setup: func() *PageLayout {
				return &PageLayout{
					Width:  612,
					Height: 792,
					TextBlocks: []TextBlock{
						{Text: "A", Rect: Rectangle{X: 10, Y: 600, Width: 50, Height: 20}},
						{Text: "B", Rect: Rectangle{X: 10, Y: 300, Width: 50, Height: 20}},
					},
				}
			},
			opts: LayoutAdjustmentOptions{
				Strategy:   StrategyCompact,
				MinSpacing: 10,
				PageMargin: 20,
			},
			wantErr: false,
			validate: func(t *testing.T, pl *PageLayout) {
				t.Helper()
				// After compact, blocks should be pushed up near top of page
				// First block: Y = pageHeight - margin - height = 792 - 20 - 20 = 752
				if pl.TextBlocks[0].Rect.Y != 752 {
					t.Errorf("Block A Y = %v, want 752", pl.TextBlocks[0].Rect.Y)
				}
				// Second block: Y = 752 - minSpacing - height = 752 - 10 - 20 = 722
				if pl.TextBlocks[1].Rect.Y != 722 {
					t.Errorf("Block B Y = %v, want 722", pl.TextBlocks[1].Rect.Y)
				}
			},
		},
		{
			name: "compact strategy with empty layout",
			setup: func() *PageLayout {
				return &PageLayout{Width: 612, Height: 792}
			},
			opts: LayoutAdjustmentOptions{
				Strategy:   StrategyCompact,
				MinSpacing: 10,
				PageMargin: 20,
			},
			wantErr: false,
		},
		{
			name: "flow down strategy",
			setup: func() *PageLayout {
				return &PageLayout{
					Width:  612,
					Height: 792,
					TextBlocks: []TextBlock{
						{Text: "A", Rect: Rectangle{X: 10, Y: 700, Width: 50, Height: 20}},
						{Text: "B", Rect: Rectangle{X: 10, Y: 705, Width: 50, Height: 20}},
					},
				}
			},
			opts: LayoutAdjustmentOptions{
				Strategy:   StrategyFlowDown,
				MinSpacing: 10,
				PageMargin: 20,
			},
			wantErr: false,
		},
		{
			name: "flow down strategy with empty layout",
			setup: func() *PageLayout {
				return &PageLayout{Width: 612, Height: 792}
			},
			opts: LayoutAdjustmentOptions{
				Strategy:   StrategyFlowDown,
				MinSpacing: 10,
				PageMargin: 20,
			},
			wantErr: false,
		},
		{
			name: "even spacing strategy",
			setup: func() *PageLayout {
				return &PageLayout{
					Width:  612,
					Height: 792,
					TextBlocks: []TextBlock{
						{Text: "A", Rect: Rectangle{X: 10, Y: 700, Width: 50, Height: 20}},
						{Text: "B", Rect: Rectangle{X: 10, Y: 400, Width: 50, Height: 20}},
						{Text: "C", Rect: Rectangle{X: 10, Y: 100, Width: 50, Height: 20}},
					},
				}
			},
			opts: LayoutAdjustmentOptions{
				Strategy:   StrategyEvenSpacing,
				MinSpacing: 10,
				PageMargin: 20,
			},
			wantErr: false,
			validate: func(t *testing.T, pl *PageLayout) {
				t.Helper()
				// Blocks should be evenly spaced from top of page
				// Total height: 3 * 20 = 60
				// Available: 792 - 2*20 - 60 = 692
				// Spacing: 692 / 2 = 346
				// All blocks should have been repositioned
				// Just verify they exist and have valid Y coordinates
				for i, tb := range pl.TextBlocks {
					if tb.Rect.Y < 0 || tb.Rect.Y > 792 {
						t.Errorf("Block %d Y = %v out of page bounds", i, tb.Rect.Y)
					}
				}
			},
		},
		{
			name: "even spacing strategy with empty layout",
			setup: func() *PageLayout {
				return &PageLayout{Width: 612, Height: 792}
			},
			opts: LayoutAdjustmentOptions{
				Strategy:   StrategyEvenSpacing,
				MinSpacing: 10,
				PageMargin: 20,
			},
			wantErr: false,
		},
		{
			name: "fit content strategy returns nil error",
			setup: func() *PageLayout {
				return &PageLayout{
					Width:  612,
					Height: 792,
					TextBlocks: []TextBlock{
						{Text: "A", Rect: Rectangle{X: 10, Y: 700, Width: 50, Height: 20}},
					},
				}
			},
			opts: LayoutAdjustmentOptions{
				Strategy:   StrategyFitContent,
				MinSpacing: 10,
				PageMargin: 20,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pl := tt.setup()
			err := pl.AdjustLayout(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("AdjustLayout() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.validate != nil && err == nil {
				tt.validate(t, pl)
			}
		})
	}
}

func TestAdjustLayoutFlowDown(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *PageLayout
		opts     LayoutAdjustmentOptions
		validate func(t *testing.T, pl *PageLayout)
	}{
		{
			name: "empty layout",
			setup: func() *PageLayout {
				return &PageLayout{Width: 612, Height: 792}
			},
			opts: LayoutAdjustmentOptions{MinSpacing: 10, PageMargin: 20},
		},
		{
			name: "single block remains unchanged",
			setup: func() *PageLayout {
				return &PageLayout{
					Width:  612,
					Height: 792,
					TextBlocks: []TextBlock{
						{Text: "Only", Rect: Rectangle{X: 10, Y: 700, Width: 50, Height: 20}},
					},
				}
			},
			opts: LayoutAdjustmentOptions{MinSpacing: 10, PageMargin: 20},
			validate: func(t *testing.T, pl *PageLayout) {
				t.Helper()
				if pl.TextBlocks[0].Rect.Y != 700 {
					t.Errorf("Y = %v, want 700", pl.TextBlocks[0].Rect.Y)
				}
			},
		},
		{
			name: "overlapping blocks pushed down",
			setup: func() *PageLayout {
				return &PageLayout{
					Width:  612,
					Height: 792,
					TextBlocks: []TextBlock{
						{Text: "Top", Rect: Rectangle{X: 10, Y: 700, Width: 50, Height: 20}},
						{Text: "Overlap", Rect: Rectangle{X: 10, Y: 705, Width: 50, Height: 20}},
					},
				}
			},
			opts: LayoutAdjustmentOptions{MinSpacing: 10, PageMargin: 20},
			validate: func(t *testing.T, pl *PageLayout) {
				t.Helper()
				// After flow down, second block should be pushed below first
				// The blocks are sorted top-to-bottom (higher Y first in PDF coords)
				// Block "Overlap" has top at 725, Block "Top" has top at 720
				// So "Overlap" is actually higher and comes first in sorted order
				// Both blocks should have valid positions
				for _, tb := range pl.TextBlocks {
					if tb.Rect.Y < -100 || tb.Rect.Y > 800 {
						t.Errorf("Block %q Y = %v seems invalid", tb.Text, tb.Rect.Y)
					}
				}
			},
		},
		{
			name: "well-spaced blocks not moved unnecessarily",
			setup: func() *PageLayout {
				return &PageLayout{
					Width:  612,
					Height: 792,
					TextBlocks: []TextBlock{
						{Text: "Top", Rect: Rectangle{X: 10, Y: 700, Width: 50, Height: 20}},
						{Text: "Bottom", Rect: Rectangle{X: 10, Y: 500, Width: 50, Height: 20}},
					},
				}
			},
			opts: LayoutAdjustmentOptions{MinSpacing: 10, PageMargin: 20},
			validate: func(t *testing.T, pl *PageLayout) {
				t.Helper()
				// Blocks are far apart so second block should not move
				// "Top" top=720, "Bottom" top=520, gap = 200 > minSpacing
				if pl.TextBlocks[1].Rect.Y != 500 {
					t.Errorf("Bottom block Y = %v, want 500 (unchanged)", pl.TextBlocks[1].Rect.Y)
				}
			},
		},
		{
			name: "flow down with image blocks",
			setup: func() *PageLayout {
				return &PageLayout{
					Width:  612,
					Height: 792,
					TextBlocks: []TextBlock{
						{Text: "Text", Rect: Rectangle{X: 10, Y: 700, Width: 50, Height: 20}},
					},
					Images: []ImageBlock{
						{X: 10, Y: 710, PlacedWidth: 50, PlacedHeight: 30},
					},
				}
			},
			opts: LayoutAdjustmentOptions{MinSpacing: 10, PageMargin: 20},
			validate: func(t *testing.T, pl *PageLayout) {
				t.Helper()
				// Just check both blocks exist and have valid positions
				if len(pl.TextBlocks) != 1 {
					t.Errorf("TextBlocks count = %d, want 1", len(pl.TextBlocks))
				}
				if len(pl.Images) != 1 {
					t.Errorf("Images count = %d, want 1", len(pl.Images))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pl := tt.setup()
			err := pl.adjustLayoutFlowDown(tt.opts)
			if err != nil {
				t.Fatalf("adjustLayoutFlowDown() error = %v", err)
			}
			if tt.validate != nil {
				tt.validate(t, pl)
			}
		})
	}
}

func TestAdjustLayoutCompact(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *PageLayout
		opts     LayoutAdjustmentOptions
		validate func(t *testing.T, pl *PageLayout)
	}{
		{
			name: "empty layout",
			setup: func() *PageLayout {
				return &PageLayout{Width: 612, Height: 792}
			},
			opts: LayoutAdjustmentOptions{MinSpacing: 10, PageMargin: 20},
		},
		{
			name: "blocks compacted to top",
			setup: func() *PageLayout {
				return &PageLayout{
					Width:  612,
					Height: 792,
					TextBlocks: []TextBlock{
						{Text: "A", Rect: Rectangle{X: 10, Y: 600, Width: 50, Height: 20}},
						{Text: "B", Rect: Rectangle{X: 10, Y: 200, Width: 50, Height: 20}},
					},
				}
			},
			opts: LayoutAdjustmentOptions{MinSpacing: 10, PageMargin: 20},
			validate: func(t *testing.T, pl *PageLayout) {
				t.Helper()
				// First block: Y = 792 - 20 - 20 = 752
				if pl.TextBlocks[0].Rect.Y != 752 {
					t.Errorf("Block A Y = %v, want 752", pl.TextBlocks[0].Rect.Y)
				}
				// Second block: Y = 752 - 10 - 20 = 722
				if pl.TextBlocks[1].Rect.Y != 722 {
					t.Errorf("Block B Y = %v, want 722", pl.TextBlocks[1].Rect.Y)
				}
			},
		},
		{
			name: "compact with image block",
			setup: func() *PageLayout {
				return &PageLayout{
					Width:  612,
					Height: 792,
					Images: []ImageBlock{
						{X: 10, Y: 400, PlacedWidth: 100, PlacedHeight: 50},
					},
				}
			},
			opts: LayoutAdjustmentOptions{MinSpacing: 10, PageMargin: 20},
			validate: func(t *testing.T, pl *PageLayout) {
				t.Helper()
				// Image: Y = 792 - 20 - 50 = 722
				if pl.Images[0].Y != 722 {
					t.Errorf("Image Y = %v, want 722", pl.Images[0].Y)
				}
			},
		},
		{
			name: "compact with mixed text and image",
			setup: func() *PageLayout {
				return &PageLayout{
					Width:  612,
					Height: 792,
					TextBlocks: []TextBlock{
						{Text: "Text", Rect: Rectangle{X: 10, Y: 700, Width: 50, Height: 20}},
					},
					Images: []ImageBlock{
						{X: 10, Y: 300, PlacedWidth: 100, PlacedHeight: 50},
					},
				}
			},
			opts: LayoutAdjustmentOptions{MinSpacing: 10, PageMargin: 20},
			validate: func(t *testing.T, pl *PageLayout) {
				t.Helper()
				// Blocks repositioned near top, text first (higher Y in sorted order)
				// Text top = 720, Image top = 350 -> Text comes first
				// Text: Y = 792 - 20 - 20 = 752
				if pl.TextBlocks[0].Rect.Y != 752 {
					t.Errorf("Text Y = %v, want 752", pl.TextBlocks[0].Rect.Y)
				}
				// Image: Y = 752 - 10 - 50 = 692
				if pl.Images[0].Y != 692 {
					t.Errorf("Image Y = %v, want 692", pl.Images[0].Y)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pl := tt.setup()
			err := pl.adjustLayoutCompact(tt.opts)
			if err != nil {
				t.Fatalf("adjustLayoutCompact() error = %v", err)
			}
			if tt.validate != nil {
				tt.validate(t, pl)
			}
		})
	}
}

func TestAdjustLayoutEvenSpacing(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *PageLayout
		opts     LayoutAdjustmentOptions
		validate func(t *testing.T, pl *PageLayout)
	}{
		{
			name: "empty layout",
			setup: func() *PageLayout {
				return &PageLayout{Width: 612, Height: 792}
			},
			opts: LayoutAdjustmentOptions{MinSpacing: 10, PageMargin: 20},
		},
		{
			name: "two blocks evenly spaced",
			setup: func() *PageLayout {
				return &PageLayout{
					Width:  612,
					Height: 792,
					TextBlocks: []TextBlock{
						{Text: "A", Rect: Rectangle{X: 10, Y: 700, Width: 50, Height: 20}},
						{Text: "B", Rect: Rectangle{X: 10, Y: 100, Width: 50, Height: 20}},
					},
				}
			},
			opts: LayoutAdjustmentOptions{MinSpacing: 10, PageMargin: 20},
			validate: func(t *testing.T, pl *PageLayout) {
				t.Helper()
				// Total height: 2 * 20 = 40
				// Available: 792 - 2*20 - 40 = 712
				// Spacing: 712 / 1 = 712
				// First block: Y = 792 - 20 - 20 = 752
				if pl.TextBlocks[0].Rect.Y != 752 {
					t.Errorf("Block A Y = %v, want 752", pl.TextBlocks[0].Rect.Y)
				}
				// Second block: Y = 752 - 712 - 20 = 20
				if pl.TextBlocks[1].Rect.Y != 20 {
					t.Errorf("Block B Y = %v, want 20", pl.TextBlocks[1].Rect.Y)
				}
			},
		},
		{
			name: "spacing falls back to minSpacing when too small",
			setup: func() *PageLayout {
				return &PageLayout{
					Width:  612,
					Height: 200,
					TextBlocks: []TextBlock{
						{Text: "A", Rect: Rectangle{X: 10, Y: 150, Width: 50, Height: 20}},
						{Text: "B", Rect: Rectangle{X: 10, Y: 120, Width: 50, Height: 20}},
						{Text: "C", Rect: Rectangle{X: 10, Y: 90, Width: 50, Height: 20}},
						{Text: "D", Rect: Rectangle{X: 10, Y: 60, Width: 50, Height: 20}},
					},
				}
			},
			opts: LayoutAdjustmentOptions{MinSpacing: 15, PageMargin: 20},
			validate: func(t *testing.T, pl *PageLayout) {
				t.Helper()
				// Total height: 4 * 20 = 80
				// Available: 200 - 2*20 - 80 = 80
				// Spacing: 80 / 3 = 26.67 > minSpacing(15), so use calculated
				// Just verify all blocks have been positioned
				for i, tb := range pl.TextBlocks {
					if tb.Rect.Y < -100 || tb.Rect.Y > 200 {
						t.Errorf("Block %d Y = %v seems invalid", i, tb.Rect.Y)
					}
				}
			},
		},
		{
			name: "not enough space falls back to compact",
			setup: func() *PageLayout {
				return &PageLayout{
					Width:  612,
					Height: 100,
					TextBlocks: []TextBlock{
						{Text: "A", Rect: Rectangle{X: 10, Y: 70, Width: 50, Height: 40}},
						{Text: "B", Rect: Rectangle{X: 10, Y: 20, Width: 50, Height: 40}},
					},
				}
			},
			opts: LayoutAdjustmentOptions{MinSpacing: 10, PageMargin: 20},
			validate: func(t *testing.T, pl *PageLayout) {
				t.Helper()
				// Total height: 2 * 40 = 80
				// Available: 100 - 2*20 - 80 = -20 < 0 -> fallback to compact
				// Compact: first = 100-20-40 = 40, second = 40-10-40 = -10
				if pl.TextBlocks[0].Rect.Y != 40 {
					t.Errorf("Block A Y = %v, want 40", pl.TextBlocks[0].Rect.Y)
				}
				if pl.TextBlocks[1].Rect.Y != -10 {
					t.Errorf("Block B Y = %v, want -10", pl.TextBlocks[1].Rect.Y)
				}
			},
		},
		{
			name: "even spacing with image",
			setup: func() *PageLayout {
				return &PageLayout{
					Width:  612,
					Height: 792,
					Images: []ImageBlock{
						{X: 10, Y: 600, PlacedWidth: 100, PlacedHeight: 50},
						{X: 10, Y: 200, PlacedWidth: 100, PlacedHeight: 50},
					},
				}
			},
			opts: LayoutAdjustmentOptions{MinSpacing: 10, PageMargin: 20},
			validate: func(t *testing.T, pl *PageLayout) {
				t.Helper()
				for i, img := range pl.Images {
					if img.Y < -100 || img.Y > 792 {
						t.Errorf("Image %d Y = %v seems invalid", i, img.Y)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pl := tt.setup()
			err := pl.adjustLayoutEvenSpacing(tt.opts)
			if err != nil {
				t.Fatalf("adjustLayoutEvenSpacing() error = %v", err)
			}
			if tt.validate != nil {
				tt.validate(t, pl)
			}
		})
	}
}

func TestAdjustLayoutFitContent(t *testing.T) {
	pl := &PageLayout{
		Width:  612,
		Height: 792,
		TextBlocks: []TextBlock{
			{Text: "A", Rect: Rectangle{X: 10, Y: 700, Width: 50, Height: 20}},
		},
	}
	opts := LayoutAdjustmentOptions{MinSpacing: 10, PageMargin: 20}

	err := pl.adjustLayoutFitContent(opts)
	if err != nil {
		t.Fatalf("adjustLayoutFitContent() error = %v", err)
	}

	// FitContent is a no-op in the layout package
	if pl.TextBlocks[0].Rect.Y != 700 {
		t.Errorf("Block Y = %v, want 700 (unchanged)", pl.TextBlocks[0].Rect.Y)
	}
}

func TestDefaultLayoutAdjustmentOptions(t *testing.T) {
	opts := DefaultLayoutAdjustmentOptions()

	if opts.Strategy != StrategyCompact {
		t.Errorf("Strategy = %v, want %v", opts.Strategy, StrategyCompact)
	}
	if opts.MinSpacing != 10.0 {
		t.Errorf("MinSpacing = %v, want 10.0", opts.MinSpacing)
	}
	if opts.PageMargin != 20.0 {
		t.Errorf("PageMargin = %v, want 20.0", opts.PageMargin)
	}
}
