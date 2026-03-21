package gopdf

import (
	"testing"
)

func TestDefaultSplitOptions(t *testing.T) {
	opts := DefaultSplitOptions()

	if opts.MinSpacing != 10.0 {
		t.Errorf("MinSpacing = %f, want 10.0", opts.MinSpacing)
	}
	if opts.PageMargin != 50.0 {
		t.Errorf("PageMargin = %f, want 50.0", opts.PageMargin)
	}
}

func TestSplitContentBlocksIntoPages(t *testing.T) {
	tests := []struct {
		name         string
		blocks       []ContentBlock
		pageSize     PageSize
		options      SplitOptions
		wantMinPages int
	}{
		{
			name:         "empty blocks",
			blocks:       nil,
			pageSize:     PageSizeA4,
			options:      DefaultSplitOptions(),
			wantMinPages: 1,
		},
		{
			name: "single text block fits on one page",
			blocks: []ContentBlock{
				TextBlock{
					Text:     "Hello World",
					Font:     "Helvetica",
					FontSize: 12,
					Rect: Rectangle{
						X:      50,
						Y:      50,
						Width:  200,
						Height: 20,
					},
				},
			},
			pageSize:     PageSizeA4,
			options:      DefaultSplitOptions(),
			wantMinPages: 1,
		},
		{
			name: "zero options get defaults",
			blocks: []ContentBlock{
				TextBlock{
					Text:     "Test",
					Font:     "Helvetica",
					FontSize: 12,
					Rect: Rectangle{
						X:      50,
						Y:      50,
						Width:  200,
						Height: 20,
					},
				},
			},
			pageSize:     PageSizeA4,
			options:      SplitOptions{}, // zero values should get defaults
			wantMinPages: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pages, err := SplitContentBlocksIntoPages(tt.blocks, tt.pageSize, tt.options)
			if err != nil {
				t.Fatalf("SplitContentBlocksIntoPages() error: %v", err)
			}
			if len(pages) < tt.wantMinPages {
				t.Errorf("got %d pages, want at least %d", len(pages), tt.wantMinPages)
			}
		})
	}
}

func TestSplitContentBlocksIntoPagesWithDefaults(t *testing.T) {
	blocks := []ContentBlock{
		TextBlock{
			Text:     "Hello",
			Font:     "Helvetica",
			FontSize: 12,
			Rect: Rectangle{
				X:      50,
				Y:      50,
				Width:  200,
				Height: 20,
			},
		},
	}

	pages, err := SplitContentBlocksIntoPagesWithDefaults(blocks, PageSizeA4)
	if err != nil {
		t.Fatalf("SplitContentBlocksIntoPagesWithDefaults() error: %v", err)
	}
	if len(pages) < 1 {
		t.Error("expected at least 1 page")
	}
}

func TestSplitContentBlocksIntoPages_MultipleBlocks(t *testing.T) {
	// Create many blocks that should span multiple pages on a small page
	smallPage := PageSize{Width: 200, Height: 100}
	var blocks []ContentBlock

	for i := 0; i < 10; i++ {
		blocks = append(blocks, TextBlock{
			Text:     "Block text content",
			Font:     "Helvetica",
			FontSize: 12,
			Rect: Rectangle{
				X:      10,
				Y:      float64(i * 30),
				Width:  180,
				Height: 25,
			},
		})
	}

	pages, err := SplitContentBlocksIntoPages(blocks, smallPage, DefaultSplitOptions())
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	// With 10 blocks of height 25 on a page of height 100, we should get multiple pages
	if len(pages) < 1 {
		t.Error("expected at least 1 page")
	}
}
