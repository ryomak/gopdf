package gopdf

import (
	"testing"
)

// TestTextBlockContentBlock はTextBlockがContentBlockインターフェースを実装していることをテスト
func TestTextBlockContentBlock(t *testing.T) {
	tb := TextBlock{
		Text: "Hello World",
		Rect: Rectangle{
			X:      100,
			Y:      200,
			Width:  300,
			Height: 50,
		},
		Font:     "Helvetica",
		FontSize: 12,
	}

	// ContentBlockインターフェースとして扱えることを確認
	var block ContentBlock = tb

	// Type()のテスト
	if block.Type() != ContentBlockTypeText {
		t.Errorf("Type() = %s, want %s", block.Type(), ContentBlockTypeText)
	}

	// Position()のテスト
	x, y := block.Position()
	if x != 100.0 {
		t.Errorf("Position() x = %f, want 100.0", x)
	}
	if y != 200.0 {
		t.Errorf("Position() y = %f, want 200.0", y)
	}

	// Bounds()のテスト
	bounds := block.Bounds()
	if bounds.X != 100.0 || bounds.Y != 200.0 || bounds.Width != 300.0 || bounds.Height != 50.0 {
		t.Errorf("Bounds() = %+v, want {X:100 Y:200 Width:300 Height:50}", bounds)
	}
}

// TestImageBlockContentBlock はImageBlockがContentBlockインターフェースを実装していることをテスト
func TestImageBlockContentBlock(t *testing.T) {
	ib := ImageBlock{
		ImageInfo: ImageInfo{
			Name:   "Im1",
			Width:  800,
			Height: 600,
		},
		X:            150,
		Y:            250,
		PlacedWidth:  400,
		PlacedHeight: 300,
	}

	// ContentBlockインターフェースとして扱えることを確認
	var block ContentBlock = ib

	// Type()のテスト
	if block.Type() != ContentBlockTypeImage {
		t.Errorf("Type() = %s, want %s", block.Type(), ContentBlockTypeImage)
	}

	// Position()のテスト
	x, y := block.Position()
	if x != 150.0 {
		t.Errorf("Position() x = %f, want 150.0", x)
	}
	if y != 250.0 {
		t.Errorf("Position() y = %f, want 250.0", y)
	}

	// Bounds()のテスト
	bounds := block.Bounds()
	if bounds.X != 150.0 || bounds.Y != 250.0 || bounds.Width != 400.0 || bounds.Height != 300.0 {
		t.Errorf("Bounds() = %+v, want {X:150 Y:250 Width:400 Height:300}", bounds)
	}
}

// TestPageLayoutContentBlocks はPageLayout.ContentBlocks()のテスト
func TestPageLayoutContentBlocks(t *testing.T) {
	layout := &PageLayout{
		PageNum: 0,
		Width:   595,
		Height:  842,
		TextBlocks: []TextBlock{
			{
				Text: "First Block",
				Rect: Rectangle{
					X:      100,
					Y:      700,
					Width:  200,
					Height: 50,
				},
			},
			{
				Text: "Second Block",
				Rect: Rectangle{
					X:      100,
					Y:      500,
					Width:  200,
					Height: 50,
				},
			},
		},
		Images: []ImageBlock{
			{
				X:            100,
				Y:            600,
				PlacedWidth:  300,
				PlacedHeight: 200,
			},
		},
	}

	blocks := layout.ContentBlocks()

	// 総数の確認（2個のTextBlock + 1個のImageBlock = 3個）
	if len(blocks) != 3 {
		t.Errorf("ContentBlocks() returned %d blocks, want 3", len(blocks))
	}

	// ソート順の確認（Y座標の降順: 700 > 600 > 500）
	_, y1 := blocks[0].Position()
	_, y2 := blocks[1].Position()
	_, y3 := blocks[2].Position()

	if y1 != 700 {
		t.Errorf("blocks[0] Y = %f, want 700", y1)
	}
	if y2 != 600 {
		t.Errorf("blocks[1] Y = %f, want 600", y2)
	}
	if y3 != 500 {
		t.Errorf("blocks[2] Y = %f, want 500", y3)
	}
}

// TestPageLayoutSortedContentBlocks はPageLayout.SortedContentBlocks()のテスト
func TestPageLayoutSortedContentBlocks(t *testing.T) {
	layout := &PageLayout{
		PageNum: 0,
		Width:   595,
		Height:  842,
		TextBlocks: []TextBlock{
			{
				Text: "Right",
				Rect: Rectangle{
					X:      300, // 右側
					Y:      700,
					Width:  100,
					Height: 50,
				},
			},
			{
				Text: "Left",
				Rect: Rectangle{
					X:      100, // 左側
					Y:      700, // 同じY座標
					Width:  100,
					Height: 50,
				},
			},
			{
				Text: "Bottom",
				Rect: Rectangle{
					X:      100,
					Y:      500, // 下側
					Width:  100,
					Height: 50,
				},
			},
		},
	}

	blocks := layout.SortedContentBlocks()

	// ソート順の確認
	// 1. Y座標の降順（上から下）
	// 2. 同じY座標の場合、X座標の昇順（左から右）
	// 期待される順序: "Left" (100, 700) -> "Right" (300, 700) -> "Bottom" (100, 500)

	if len(blocks) != 3 {
		t.Fatalf("SortedContentBlocks() returned %d blocks, want 3", len(blocks))
	}

	// 最初のブロック: Left (100, 700)
	x1, y1 := blocks[0].Position()
	if x1 != 100 || y1 != 700 {
		t.Errorf("blocks[0] position = (%f, %f), want (100, 700)", x1, y1)
	}

	// 2番目のブロック: Right (300, 700)
	x2, y2 := blocks[1].Position()
	if x2 != 300 || y2 != 700 {
		t.Errorf("blocks[1] position = (%f, %f), want (300, 700)", x2, y2)
	}

	// 3番目のブロック: Bottom (100, 500)
	x3, y3 := blocks[2].Position()
	if x3 != 100 || y3 != 500 {
		t.Errorf("blocks[2] position = (%f, %f), want (100, 500)", x3, y3)
	}
}
