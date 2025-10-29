package gopdf

import (
	"testing"
)

// TestMoveBlock_TextBlock はTextBlockの移動テスト
func TestMoveBlock_TextBlock(t *testing.T) {
	layout := &PageLayout{
		Width:  595,
		Height: 842,
		TextBlocks: []TextBlock{
			{
				Text: "Block 1",
				Rect: Rectangle{X: 100, Y: 700, Width: 200, Height: 50},
			},
			{
				Text: "Block 2",
				Rect: Rectangle{X: 100, Y: 600, Width: 200, Height: 50},
			},
		},
	}

	// Block 0を右に50、下に30移動
	err := layout.MoveBlock(ContentBlockTypeText, 0, 50, -30)
	if err != nil {
		t.Fatalf("MoveBlock failed: %v", err)
	}

	// 位置が正しく更新されていることを確認
	if layout.TextBlocks[0].Rect.X != 150 {
		t.Errorf("TextBlocks[0].Rect.X = %f, want 150", layout.TextBlocks[0].Rect.X)
	}
	if layout.TextBlocks[0].Rect.Y != 670 {
		t.Errorf("TextBlocks[0].Rect.Y = %f, want 670", layout.TextBlocks[0].Rect.Y)
	}

	// Block 1は影響を受けないことを確認
	if layout.TextBlocks[1].Rect.X != 100 {
		t.Errorf("TextBlocks[1].Rect.X = %f, want 100", layout.TextBlocks[1].Rect.X)
	}
	if layout.TextBlocks[1].Rect.Y != 600 {
		t.Errorf("TextBlocks[1].Rect.Y = %f, want 600", layout.TextBlocks[1].Rect.Y)
	}
}

// TestMoveBlock_ImageBlock はImageBlockの移動テスト
func TestMoveBlock_ImageBlock(t *testing.T) {
	layout := &PageLayout{
		Width:  595,
		Height: 842,
		Images: []ImageBlock{
			{X: 100, Y: 500, PlacedWidth: 300, PlacedHeight: 200},
			{X: 100, Y: 250, PlacedWidth: 300, PlacedHeight: 200},
		},
	}

	// Image 0を移動
	err := layout.MoveBlock(ContentBlockTypeImage, 0, 25, -10)
	if err != nil {
		t.Fatalf("MoveBlock failed: %v", err)
	}

	if layout.Images[0].X != 125 {
		t.Errorf("Images[0].X = %f, want 125", layout.Images[0].X)
	}
	if layout.Images[0].Y != 490 {
		t.Errorf("Images[0].Y = %f, want 490", layout.Images[0].Y)
	}

	// Image 1は影響を受けないことを確認
	if layout.Images[1].X != 100 {
		t.Errorf("Images[1].X = %f, want 100", layout.Images[1].X)
	}
}

// TestMoveBlock_InvalidIndex は無効なインデックスのテスト
func TestMoveBlock_InvalidIndex(t *testing.T) {
	layout := &PageLayout{
		TextBlocks: []TextBlock{
			{Rect: Rectangle{X: 100, Y: 700}},
		},
	}

	// 範囲外のインデックス
	err := layout.MoveBlock(ContentBlockTypeText, 10, 50, 50)
	if err == nil {
		t.Error("Expected error for invalid index, got nil")
	}

	// 負のインデックス
	err = layout.MoveBlock(ContentBlockTypeText, -1, 50, 50)
	if err == nil {
		t.Error("Expected error for negative index, got nil")
	}
}

// TestResizeBlock_TextBlock はTextBlockのリサイズテスト
func TestResizeBlock_TextBlock(t *testing.T) {
	layout := &PageLayout{
		TextBlocks: []TextBlock{
			{
				Text: "Block 1",
				Rect: Rectangle{X: 100, Y: 700, Width: 200, Height: 50},
			},
		},
	}

	// Block 0をリサイズ
	err := layout.ResizeBlock(ContentBlockTypeText, 0, 400, 100)
	if err != nil {
		t.Fatalf("ResizeBlock failed: %v", err)
	}

	if layout.TextBlocks[0].Rect.Width != 400 {
		t.Errorf("TextBlocks[0].Rect.Width = %f, want 400", layout.TextBlocks[0].Rect.Width)
	}
	if layout.TextBlocks[0].Rect.Height != 100 {
		t.Errorf("TextBlocks[0].Rect.Height = %f, want 100", layout.TextBlocks[0].Rect.Height)
	}

	// 位置は変わらないことを確認
	if layout.TextBlocks[0].Rect.X != 100 {
		t.Errorf("TextBlocks[0].Rect.X = %f, want 100", layout.TextBlocks[0].Rect.X)
	}
	if layout.TextBlocks[0].Rect.Y != 700 {
		t.Errorf("TextBlocks[0].Rect.Y = %f, want 700", layout.TextBlocks[0].Rect.Y)
	}
}

// TestResizeBlock_ImageBlock はImageBlockのリサイズテスト
func TestResizeBlock_ImageBlock(t *testing.T) {
	layout := &PageLayout{
		Images: []ImageBlock{
			{X: 100, Y: 500, PlacedWidth: 300, PlacedHeight: 200},
		},
	}

	err := layout.ResizeBlock(ContentBlockTypeImage, 0, 400, 300)
	if err != nil {
		t.Fatalf("ResizeBlock failed: %v", err)
	}

	if layout.Images[0].PlacedWidth != 400 {
		t.Errorf("Images[0].PlacedWidth = %f, want 400", layout.Images[0].PlacedWidth)
	}
	if layout.Images[0].PlacedHeight != 300 {
		t.Errorf("Images[0].PlacedHeight = %f, want 300", layout.Images[0].PlacedHeight)
	}

	// 位置は変わらないことを確認
	if layout.Images[0].X != 100 {
		t.Errorf("Images[0].X = %f, want 100", layout.Images[0].X)
	}
	if layout.Images[0].Y != 500 {
		t.Errorf("Images[0].Y = %f, want 500", layout.Images[0].Y)
	}
}

// TestDetectOverlaps_NoOverlaps は重なりがない場合のテスト
func TestDetectOverlaps_NoOverlaps(t *testing.T) {
	layout := &PageLayout{
		TextBlocks: []TextBlock{
			{Rect: Rectangle{X: 100, Y: 700, Width: 200, Height: 50}},
			{Rect: Rectangle{X: 100, Y: 600, Width: 200, Height: 50}},
		},
	}

	overlaps := layout.DetectOverlaps()
	if len(overlaps) != 0 {
		t.Errorf("Expected no overlaps, got %d", len(overlaps))
	}
}

// TestDetectOverlaps_WithOverlaps は重なりがある場合のテスト
func TestDetectOverlaps_WithOverlaps(t *testing.T) {
	layout := &PageLayout{
		TextBlocks: []TextBlock{
			{Rect: Rectangle{X: 100, Y: 100, Width: 200, Height: 50}},
			{Rect: Rectangle{X: 150, Y: 120, Width: 200, Height: 50}}, // 重なる
		},
	}

	overlaps := layout.DetectOverlaps()
	if len(overlaps) != 1 {
		t.Errorf("Expected 1 overlap, got %d", len(overlaps))
	}

	if len(overlaps) > 0 {
		// 重なり面積が正の値であることを確認
		if overlaps[0].Area <= 0 {
			t.Errorf("Overlap area = %f, want > 0", overlaps[0].Area)
		}
	}
}

// TestDetectOverlaps_MixedBlocks はTextBlockとImageBlockが混在する場合のテスト
func TestDetectOverlaps_MixedBlocks(t *testing.T) {
	layout := &PageLayout{
		TextBlocks: []TextBlock{
			{Rect: Rectangle{X: 100, Y: 500, Width: 200, Height: 50}},
		},
		Images: []ImageBlock{
			{X: 150, Y: 510, PlacedWidth: 200, PlacedHeight: 50}, // TextBlockと重なる
		},
	}

	overlaps := layout.DetectOverlaps()
	if len(overlaps) != 1 {
		t.Errorf("Expected 1 overlap, got %d", len(overlaps))
	}
}

// TestDetectOverlaps_EdgeTouch はブロックが接触している場合のテスト（重なりとみなさない）
func TestDetectOverlaps_EdgeTouch(t *testing.T) {
	layout := &PageLayout{
		TextBlocks: []TextBlock{
			{Rect: Rectangle{X: 100, Y: 100, Width: 200, Height: 50}},
			{Rect: Rectangle{X: 100, Y: 150, Width: 200, Height: 50}}, // 下端が上端に接触
		},
	}

	overlaps := layout.DetectOverlaps()
	// 接触は重なりとみなさない
	if len(overlaps) != 0 {
		t.Errorf("Expected no overlaps for edge touch, got %d", len(overlaps))
	}
}

// TestSplitIntoPages_SinglePage は1ページに収まる場合のテスト
func TestSplitIntoPages_SinglePage(t *testing.T) {
	layout := &PageLayout{
		Width:  595,
		Height: 842,
		TextBlocks: []TextBlock{
			{Rect: Rectangle{X: 100, Y: 700, Width: 200, Height: 50}},
			{Rect: Rectangle{X: 100, Y: 600, Width: 200, Height: 50}},
		},
	}

	pages, err := layout.SplitIntoPages(842.0, 10.0, 20.0)
	if err != nil {
		t.Fatalf("SplitIntoPages failed: %v", err)
	}

	if len(pages) != 1 {
		t.Errorf("Expected 1 page, got %d", len(pages))
	}

	if len(pages) > 0 && len(pages[0].TextBlocks) != 2 {
		t.Errorf("Expected 2 blocks in page, got %d", len(pages[0].TextBlocks))
	}
}

// TestSplitIntoPages_MultiplePages は複数ページに分割される場合のテスト
func TestSplitIntoPages_MultiplePages(t *testing.T) {
	layout := &PageLayout{
		Width:  595,
		Height: 842,
		TextBlocks: []TextBlock{
			{Text: "Block 1", Rect: Rectangle{X: 100, Y: 800, Width: 200, Height: 300}},
			{Text: "Block 2", Rect: Rectangle{X: 100, Y: 450, Width: 200, Height: 300}},
			{Text: "Block 3", Rect: Rectangle{X: 100, Y: 100, Width: 200, Height: 300}},
		},
	}

	// ページ高さ500、マージン20なので、1ページに1ブロックずつ配置される
	pages, err := layout.SplitIntoPages(500.0, 10.0, 20.0)
	if err != nil {
		t.Fatalf("SplitIntoPages failed: %v", err)
	}

	if len(pages) < 2 {
		t.Errorf("Expected at least 2 pages, got %d", len(pages))
	}

	// 各ページのブロック数を確認
	totalBlocks := 0
	for _, page := range pages {
		totalBlocks += len(page.TextBlocks)
	}

	if totalBlocks != 3 {
		t.Errorf("Expected 3 total blocks across pages, got %d", totalBlocks)
	}
}

// TestSplitIntoPages_WithImages は画像を含む場合のテスト
func TestSplitIntoPages_WithImages(t *testing.T) {
	layout := &PageLayout{
		Width:  595,
		Height: 842,
		TextBlocks: []TextBlock{
			{Rect: Rectangle{X: 100, Y: 700, Width: 200, Height: 50}},
		},
		Images: []ImageBlock{
			{X: 100, Y: 400, PlacedWidth: 300, PlacedHeight: 200},
		},
	}

	pages, err := layout.SplitIntoPages(842.0, 10.0, 20.0)
	if err != nil {
		t.Fatalf("SplitIntoPages failed: %v", err)
	}

	if len(pages) != 1 {
		t.Errorf("Expected 1 page, got %d", len(pages))
	}

	if len(pages) > 0 {
		if len(pages[0].TextBlocks) != 1 {
			t.Errorf("Expected 1 text block, got %d", len(pages[0].TextBlocks))
		}
		if len(pages[0].Images) != 1 {
			t.Errorf("Expected 1 image block, got %d", len(pages[0].Images))
		}
	}
}

// TestSplitIntoPages_EmptyLayout は空のレイアウトの場合のテスト
func TestSplitIntoPages_EmptyLayout(t *testing.T) {
	layout := &PageLayout{
		Width:  595,
		Height: 842,
	}

	pages, err := layout.SplitIntoPages(842.0, 10.0, 20.0)
	if err != nil {
		t.Fatalf("SplitIntoPages failed: %v", err)
	}

	// 空のレイアウトでも1ページは返す
	if len(pages) != 1 {
		t.Errorf("Expected 1 empty page, got %d", len(pages))
	}
}
