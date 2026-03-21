package layout

import (
	"testing"
)

func TestTextBlockBounds(t *testing.T) {
	tb := TextBlock{Rect: Rectangle{X: 10, Y: 20, Width: 100, Height: 50}}
	bounds := tb.Bounds()
	if bounds.X != 10 || bounds.Y != 20 || bounds.Width != 100 || bounds.Height != 50 {
		t.Errorf("Bounds() = %+v, want {10 20 100 50}", bounds)
	}
}

func TestTextBlockType(t *testing.T) {
	tb := TextBlock{}
	if tb.Type() != ContentBlockTypeText {
		t.Errorf("Type() = %v, want %v", tb.Type(), ContentBlockTypeText)
	}
}

func TestTextBlockPosition(t *testing.T) {
	tb := TextBlock{Rect: Rectangle{X: 15, Y: 25, Width: 100, Height: 50}}
	x, y := tb.Position()
	if x != 15 || y != 25 {
		t.Errorf("Position() = (%v, %v), want (15, 25)", x, y)
	}
}

func TestTextBlockWithY(t *testing.T) {
	tb := TextBlock{
		Text: "Hello",
		Rect: Rectangle{X: 10, Y: 100, Width: 50, Height: 12},
		Font: "Helvetica",
	}
	newBlock := tb.WithY(200)
	_, newY := newBlock.Position()
	if newY != 200 {
		t.Errorf("WithY() new Y = %v, want 200", newY)
	}
	// Original unchanged
	if tb.Rect.Y != 100 {
		t.Errorf("Original Y = %v, want 100 (unchanged)", tb.Rect.Y)
	}
	// Type preserved
	if newBlock.Type() != ContentBlockTypeText {
		t.Errorf("Type = %v, want text", newBlock.Type())
	}
}

func TestTextBlockAddToLayout(t *testing.T) {
	tb := TextBlock{Text: "Hello", Rect: Rectangle{X: 10, Y: 100, Width: 50, Height: 12}}
	pl := &PageLayout{}
	tb.AddToLayout(pl)
	if len(pl.TextBlocks) != 1 {
		t.Fatalf("TextBlocks count = %d, want 1", len(pl.TextBlocks))
	}
	if pl.TextBlocks[0].Text != "Hello" {
		t.Errorf("Text = %q, want %q", pl.TextBlocks[0].Text, "Hello")
	}
}

func TestImageBlockBounds(t *testing.T) {
	ib := ImageBlock{X: 10, Y: 20, PlacedWidth: 100, PlacedHeight: 50}
	bounds := ib.Bounds()
	if bounds.X != 10 || bounds.Y != 20 || bounds.Width != 100 || bounds.Height != 50 {
		t.Errorf("Bounds() = %+v, want {10 20 100 50}", bounds)
	}
}

func TestImageBlockType(t *testing.T) {
	ib := ImageBlock{}
	if ib.Type() != ContentBlockTypeImage {
		t.Errorf("Type() = %v, want %v", ib.Type(), ContentBlockTypeImage)
	}
}

func TestImageBlockPosition(t *testing.T) {
	ib := ImageBlock{X: 15, Y: 25, PlacedWidth: 100, PlacedHeight: 50}
	x, y := ib.Position()
	if x != 15 || y != 25 {
		t.Errorf("Position() = (%v, %v), want (15, 25)", x, y)
	}
}

func TestImageBlockWithY(t *testing.T) {
	ib := ImageBlock{
		X:            10,
		Y:            100,
		PlacedWidth:  200,
		PlacedHeight: 150,
	}
	newBlock := ib.WithY(300)
	_, newY := newBlock.Position()
	if newY != 300 {
		t.Errorf("WithY() new Y = %v, want 300", newY)
	}
	// Original unchanged
	if ib.Y != 100 {
		t.Errorf("Original Y = %v, want 100 (unchanged)", ib.Y)
	}
	if newBlock.Type() != ContentBlockTypeImage {
		t.Errorf("Type = %v, want image", newBlock.Type())
	}
}

func TestImageBlockAddToLayout(t *testing.T) {
	ib := ImageBlock{X: 10, Y: 100, PlacedWidth: 200, PlacedHeight: 150}
	pl := &PageLayout{}
	ib.AddToLayout(pl)
	if len(pl.Images) != 1 {
		t.Fatalf("Images count = %d, want 1", len(pl.Images))
	}
	if pl.Images[0].X != 10 {
		t.Errorf("X = %v, want 10", pl.Images[0].X)
	}
}

func TestTextElementGetters(t *testing.T) {
	te := TextElement{X: 10, Y: 20, Size: 14}
	if te.GetX() != 10 {
		t.Errorf("GetX() = %v, want 10", te.GetX())
	}
	if te.GetY() != 20 {
		t.Errorf("GetY() = %v, want 20", te.GetY())
	}
	if te.GetSize() != 14 {
		t.Errorf("GetSize() = %v, want 14", te.GetSize())
	}
}

func TestPageLayoutContentBlocks(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() *PageLayout
		wantCount int
	}{
		{
			name: "empty layout",
			setup: func() *PageLayout {
				return &PageLayout{}
			},
			wantCount: 0,
		},
		{
			name: "text blocks only",
			setup: func() *PageLayout {
				return &PageLayout{
					TextBlocks: []TextBlock{
						{Text: "A", Rect: Rectangle{X: 0, Y: 100, Width: 50, Height: 12}},
						{Text: "B", Rect: Rectangle{X: 0, Y: 200, Width: 50, Height: 12}},
					},
				}
			},
			wantCount: 2,
		},
		{
			name: "image blocks only",
			setup: func() *PageLayout {
				return &PageLayout{
					Images: []ImageBlock{
						{X: 0, Y: 100, PlacedWidth: 50, PlacedHeight: 50},
					},
				}
			},
			wantCount: 1,
		},
		{
			name: "mixed text and image blocks sorted by Y descending",
			setup: func() *PageLayout {
				return &PageLayout{
					TextBlocks: []TextBlock{
						{Text: "Low", Rect: Rectangle{X: 0, Y: 50, Width: 50, Height: 12}},
						{Text: "High", Rect: Rectangle{X: 0, Y: 300, Width: 50, Height: 12}},
					},
					Images: []ImageBlock{
						{X: 0, Y: 150, PlacedWidth: 50, PlacedHeight: 50},
					},
				}
			},
			wantCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pl := tt.setup()
			blocks := pl.ContentBlocks()
			if len(blocks) != tt.wantCount {
				t.Errorf("ContentBlocks() count = %d, want %d", len(blocks), tt.wantCount)
				return
			}
			// Verify sorted by Y descending (higher Y first)
			for i := 1; i < len(blocks); i++ {
				_, yi := blocks[i-1].Position()
				_, yj := blocks[i].Position()
				if yi < yj {
					t.Errorf("ContentBlocks() not sorted: block %d Y=%v < block %d Y=%v", i-1, yi, i, yj)
				}
			}
		})
	}
}

func TestPageLayoutSortedContentBlocks(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() *PageLayout
		wantCount int
		validate  func(t *testing.T, blocks []ContentBlock)
	}{
		{
			name: "empty layout",
			setup: func() *PageLayout {
				return &PageLayout{}
			},
			wantCount: 0,
		},
		{
			name: "sorted by top edge descending then X ascending",
			setup: func() *PageLayout {
				return &PageLayout{
					TextBlocks: []TextBlock{
						{Text: "BottomLeft", Rect: Rectangle{X: 10, Y: 50, Width: 50, Height: 12}},
						{Text: "TopRight", Rect: Rectangle{X: 200, Y: 300, Width: 50, Height: 12}},
						{Text: "TopLeft", Rect: Rectangle{X: 10, Y: 300, Width: 50, Height: 12}},
					},
				}
			},
			wantCount: 3,
			validate: func(t *testing.T, blocks []ContentBlock) {
				t.Helper()
				// TopLeft and TopRight have same top edge (312), TopLeft should come first (X=10 < X=200)
				// BottomLeft has top edge 62, should be last
				if blocks[0].(TextBlock).Text != "TopLeft" {
					t.Errorf("First block = %q, want TopLeft", blocks[0].(TextBlock).Text)
				}
				if blocks[1].(TextBlock).Text != "TopRight" {
					t.Errorf("Second block = %q, want TopRight", blocks[1].(TextBlock).Text)
				}
				if blocks[2].(TextBlock).Text != "BottomLeft" {
					t.Errorf("Third block = %q, want BottomLeft", blocks[2].(TextBlock).Text)
				}
			},
		},
		{
			name: "blocks within epsilon treated as same line",
			setup: func() *PageLayout {
				return &PageLayout{
					TextBlocks: []TextBlock{
						{Text: "Right", Rect: Rectangle{X: 200, Y: 100, Width: 50, Height: 12}},
						{Text: "Left", Rect: Rectangle{X: 10, Y: 100.5, Width: 50, Height: 12}},
					},
				}
			},
			wantCount: 2,
			validate: func(t *testing.T, blocks []ContentBlock) {
				t.Helper()
				// Both tops are within epsilon (1.0): 112 and 112.5
				// So sorted by X: Left (X=10) before Right (X=200)
				if blocks[0].(TextBlock).Text != "Left" {
					t.Errorf("First block = %q, want Left", blocks[0].(TextBlock).Text)
				}
				if blocks[1].(TextBlock).Text != "Right" {
					t.Errorf("Second block = %q, want Right", blocks[1].(TextBlock).Text)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pl := tt.setup()
			blocks := pl.SortedContentBlocks()
			if len(blocks) != tt.wantCount {
				t.Errorf("SortedContentBlocks() count = %d, want %d", len(blocks), tt.wantCount)
				return
			}
			if tt.validate != nil {
				tt.validate(t, blocks)
			}
		})
	}
}

func TestContentBlockInterface(t *testing.T) {
	// Verify both TextBlock and ImageBlock satisfy ContentBlock interface
	var _ ContentBlock = TextBlock{}
	var _ ContentBlock = ImageBlock{}

	// Test that WithY returns correct concrete types via interface
	tb := TextBlock{Text: "test", Rect: Rectangle{X: 1, Y: 2, Width: 3, Height: 4}}
	newTB := tb.WithY(99)
	if _, ok := newTB.(TextBlock); !ok {
		t.Error("TextBlock.WithY() should return TextBlock")
	}

	ib := ImageBlock{X: 1, Y: 2, PlacedWidth: 3, PlacedHeight: 4}
	newIB := ib.WithY(99)
	if _, ok := newIB.(ImageBlock); !ok {
		t.Error("ImageBlock.WithY() should return ImageBlock")
	}
}
