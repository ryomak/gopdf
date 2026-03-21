package layout

import (
	"testing"
)

func TestGroupTextElements(t *testing.T) {
	tests := []struct {
		name       string
		elements   []TextElement
		wantBlocks int
		validate   func(t *testing.T, blocks []TextBlock)
	}{
		{
			name:       "empty elements returns nil",
			elements:   nil,
			wantBlocks: 0,
		},
		{
			name: "single element produces single block",
			elements: []TextElement{
				{Text: "Hello", X: 0, Y: 100, Width: 30, Height: 12, Font: "Helvetica", Size: 12},
			},
			wantBlocks: 1,
			validate: func(t *testing.T, blocks []TextBlock) {
				t.Helper()
				if blocks[0].Text != "Hello" {
					t.Errorf("Text = %q, want %q", blocks[0].Text, "Hello")
				}
			},
		},
		{
			name: "multiple elements same line same font grouped into one block",
			elements: []TextElement{
				{Text: "Hello", X: 0, Y: 100, Width: 30, Height: 12, Font: "Helvetica", Size: 12},
				{Text: "World", X: 40, Y: 100, Width: 30, Height: 12, Font: "Helvetica", Size: 12},
			},
			wantBlocks: 1,
		},
		{
			name: "multiple lines close together same font merged into one block",
			elements: []TextElement{
				{Text: "Line1", X: 0, Y: 100, Width: 30, Height: 12, Font: "Helvetica", Size: 12},
				{Text: "Line2", X: 0, Y: 85, Width: 30, Height: 12, Font: "Helvetica", Size: 12},
			},
			wantBlocks: 1,
			validate: func(t *testing.T, blocks []TextBlock) {
				t.Helper()
				if len(blocks[0].Elements) != 2 {
					t.Errorf("Elements = %d, want 2", len(blocks[0].Elements))
				}
			},
		},
		{
			name: "multiple lines with large spacing produce separate blocks",
			elements: []TextElement{
				{Text: "Block1", X: 0, Y: 200, Width: 30, Height: 12, Font: "Helvetica", Size: 12},
				{Text: "Block2", X: 0, Y: 100, Width: 30, Height: 12, Font: "Helvetica", Size: 12},
			},
			wantBlocks: 2,
		},
		{
			name: "different fonts produce separate blocks",
			elements: []TextElement{
				{Text: "Bold", X: 0, Y: 100, Width: 30, Height: 12, Font: "Helvetica-Bold", Size: 12},
				{Text: "Regular", X: 50, Y: 100, Width: 30, Height: 12, Font: "Helvetica", Size: 12},
			},
			wantBlocks: 2,
		},
		{
			name: "three lines same font close together produce one block",
			elements: []TextElement{
				{Text: "Line1", X: 0, Y: 100, Width: 30, Height: 12, Font: "Helvetica", Size: 12},
				{Text: "Line2", X: 0, Y: 85, Width: 30, Height: 12, Font: "Helvetica", Size: 12},
				{Text: "Line3", X: 0, Y: 70, Width: 30, Height: 12, Font: "Helvetica", Size: 12},
			},
			wantBlocks: 1,
			validate: func(t *testing.T, blocks []TextBlock) {
				t.Helper()
				if len(blocks[0].Elements) != 3 {
					t.Errorf("Elements = %d, want 3", len(blocks[0].Elements))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GroupTextElements(tt.elements)
			if len(got) != tt.wantBlocks {
				t.Errorf("GroupTextElements() returned %d blocks, want %d", len(got), tt.wantBlocks)
				return
			}
			if tt.validate != nil {
				tt.validate(t, got)
			}
		})
	}
}

func TestGroupTextElementsWithImages(t *testing.T) {
	tests := []struct {
		name       string
		elements   []TextElement
		images     []ImageBlock
		wantBlocks int
		validate   func(t *testing.T, blocks []TextBlock)
	}{
		{
			name:       "empty elements returns nil",
			elements:   nil,
			images:     nil,
			wantBlocks: 0,
		},
		{
			name: "elements with nil images delegates to GroupTextElements",
			elements: []TextElement{
				{Text: "Hello", X: 0, Y: 100, Width: 30, Height: 12, Font: "Helvetica", Size: 12},
			},
			images:     nil,
			wantBlocks: 1,
		},
		{
			name: "image between two lines splits into two blocks",
			elements: []TextElement{
				{Text: "Above", X: 0, Y: 150, Width: 30, Height: 12, Font: "Helvetica", Size: 12},
				{Text: "Below", X: 0, Y: 80, Width: 30, Height: 12, Font: "Helvetica", Size: 12},
			},
			images: []ImageBlock{
				{Y: 95, PlacedHeight: 50}, // Image spans Y 95-145, between the two lines
			},
			wantBlocks: 2,
			validate: func(t *testing.T, blocks []TextBlock) {
				t.Helper()
				if blocks[0].Text != "Above" {
					t.Errorf("First block text = %q, want %q", blocks[0].Text, "Above")
				}
				if blocks[1].Text != "Below" {
					t.Errorf("Second block text = %q, want %q", blocks[1].Text, "Below")
				}
			},
		},
		{
			name: "image not between lines does not split",
			elements: []TextElement{
				{Text: "Line1", X: 0, Y: 100, Width: 30, Height: 12, Font: "Helvetica", Size: 12},
				{Text: "Line2", X: 0, Y: 85, Width: 30, Height: 12, Font: "Helvetica", Size: 12},
			},
			images: []ImageBlock{
				{Y: 10, PlacedHeight: 30}, // Image is far below the text lines
			},
			wantBlocks: 1,
		},
		{
			name: "multiple images splitting three groups",
			elements: []TextElement{
				{Text: "Top", X: 0, Y: 300, Width: 30, Height: 12, Font: "Helvetica", Size: 12},
				{Text: "Mid", X: 0, Y: 200, Width: 30, Height: 12, Font: "Helvetica", Size: 12},
				{Text: "Bot", X: 0, Y: 100, Width: 30, Height: 12, Font: "Helvetica", Size: 12},
			},
			images: []ImageBlock{
				{Y: 215, PlacedHeight: 80}, // Between Top and Mid
				{Y: 115, PlacedHeight: 80}, // Between Mid and Bot
			},
			wantBlocks: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GroupTextElementsWithImages(tt.elements, tt.images)
			if len(got) != tt.wantBlocks {
				t.Errorf("GroupTextElementsWithImages() returned %d blocks, want %d", len(got), tt.wantBlocks)
				return
			}
			if tt.validate != nil {
				tt.validate(t, got)
			}
		})
	}
}
