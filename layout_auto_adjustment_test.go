package gopdf

import (
	"testing"
)

// TestAdjustLayout_StrategyFlowDown はFlowDown戦略のテスト
func TestAdjustLayout_StrategyFlowDown(t *testing.T) {
	layout := &PageLayout{
		Width:  595,
		Height: 842,
		TextBlocks: []TextBlock{
			{
				Text: "Block 1",
				Rect: Rectangle{X: 50, Y: 650, Width: 200, Height: 50},
			},
			{
				Text: "Block 2",
				Rect: Rectangle{X: 50, Y: 595, Width: 200, Height: 50}, // 間隔5pxで重なっている
			},
			{
				Text: "Block 3",
				Rect: Rectangle{X: 50, Y: 540, Width: 200, Height: 50}, // 間隔5pxで重なっている
			},
		},
	}

	// FlowDown戦略で自動調整（最小間隔10pxに修正）
	opts := LayoutAdjustmentOptions{
		Strategy:   StrategyFlowDown,
		MinSpacing: 10,
	}
	err := layout.AdjustLayout(opts)
	if err != nil {
		t.Fatalf("AdjustLayout failed: %v", err)
	}

	// Block 2とBlock 3が自動的に下に移動していることを確認
	// Block 1: Y=650 (bottom)
	// Block 2: top=650-10=640, height=50なので Y=640-50=590が期待値
	expectedY2 := float64(640 - 50)
	if layout.TextBlocks[1].Rect.Y != expectedY2 {
		t.Errorf("TextBlocks[1].Rect.Y = %f, want %f", layout.TextBlocks[1].Rect.Y, expectedY2)
	}

	// Block 2: Y=590 (bottom)
	// Block 3: top=590-10=580, height=50なので Y=580-50=530が期待値
	expectedY3 := float64(580 - 50)
	if layout.TextBlocks[2].Rect.Y != expectedY3 {
		t.Errorf("TextBlocks[2].Rect.Y = %f, want %f", layout.TextBlocks[2].Rect.Y, expectedY3)
	}
}

// TestAdjustLayout_StrategyCompact はCompact戦略のテスト
func TestAdjustLayout_StrategyCompact(t *testing.T) {
	layout := &PageLayout{
		Width:  595,
		Height: 842,
		TextBlocks: []TextBlock{
			{
				Text: "Block 1",
				Rect: Rectangle{X: 50, Y: 700, Width: 200, Height: 50},
			},
			{
				Text: "Block 2",
				Rect: Rectangle{X: 50, Y: 500, Width: 200, Height: 50}, // 大きな間隔
			},
		},
	}

	// Compact戦略で詰める
	opts := LayoutAdjustmentOptions{
		Strategy:   StrategyCompact,
		MinSpacing: 10,
		PageMargin: 20,
	}
	err := layout.AdjustLayout(opts)
	if err != nil {
		t.Fatalf("AdjustLayout failed: %v", err)
	}

	// Block 1: ページトップから配置される
	expectedY1 := layout.Height - opts.PageMargin - layout.TextBlocks[0].Rect.Height
	if layout.TextBlocks[0].Rect.Y != expectedY1 {
		t.Errorf("TextBlocks[0].Rect.Y = %f, want %f", layout.TextBlocks[0].Rect.Y, expectedY1)
	}

	// Block 2: Block 1の直下に配置される
	expectedY2 := layout.TextBlocks[0].Rect.Y - opts.MinSpacing - layout.TextBlocks[1].Rect.Height
	if layout.TextBlocks[1].Rect.Y != expectedY2 {
		t.Errorf("TextBlocks[1].Rect.Y = %f, want %f", layout.TextBlocks[1].Rect.Y, expectedY2)
	}
}

// TestAdjustLayout_WithImages は画像を含む場合のテスト
func TestAdjustLayout_WithImages(t *testing.T) {
	layout := &PageLayout{
		Width:  595,
		Height: 842,
		TextBlocks: []TextBlock{
			{
				Text: "Text Block",
				Rect: Rectangle{X: 50, Y: 700, Width: 200, Height: 50}, // top=750
			},
		},
		Images: []ImageBlock{
			{
				X:            50,
				Y:            695, // top=745, 間隔5pxで配置
				PlacedWidth:  200,
				PlacedHeight: 50,
			},
		},
	}

	opts := LayoutAdjustmentOptions{
		Strategy:   StrategyFlowDown,
		MinSpacing: 10,
	}
	err := layout.AdjustLayout(opts)
	if err != nil {
		t.Fatalf("AdjustLayout failed: %v", err)
	}

	// Imageが自動的に下に移動していることを確認
	// TextBlock: bottom=700
	// Image: 期待されるtop=700-10=690, height=50なので Y=690-50=640
	expectedY := float64(690 - 50)
	if layout.Images[0].Y != expectedY {
		t.Errorf("Images[0].Y = %f, want %f", layout.Images[0].Y, expectedY)
	}
}

// TestAdjustLayout_StrategyPreservePosition は位置保持戦略のテスト
func TestAdjustLayout_StrategyPreservePosition(t *testing.T) {
	layout := &PageLayout{
		Width:  595,
		Height: 842,
		TextBlocks: []TextBlock{
			{
				Text: "Block 1",
				Rect: Rectangle{X: 50, Y: 700, Width: 200, Height: 50},
			},
			{
				Text: "Block 2",
				Rect: Rectangle{X: 50, Y: 640, Width: 200, Height: 50},
			},
		},
	}

	originalY := layout.TextBlocks[1].Rect.Y

	// 位置保持戦略
	opts := LayoutAdjustmentOptions{
		Strategy: StrategyPreservePosition,
	}
	err := layout.AdjustLayout(opts)
	if err != nil {
		t.Fatalf("AdjustLayout failed: %v", err)
	}

	// 位置が変わっていないことを確認
	if layout.TextBlocks[1].Rect.Y != originalY {
		t.Errorf("Position changed: Y = %f, want %f", layout.TextBlocks[1].Rect.Y, originalY)
	}
}

// TestAdjustLayout_EmptyLayout は空のレイアウトのテスト
func TestAdjustLayout_EmptyLayout(t *testing.T) {
	layout := &PageLayout{
		Width:  595,
		Height: 842,
	}

	opts := DefaultLayoutAdjustmentOptions()
	err := layout.AdjustLayout(opts)
	if err != nil {
		t.Fatalf("AdjustLayout failed on empty layout: %v", err)
	}
}

// TestDefaultLayoutAdjustmentOptions はデフォルトオプションのテスト
func TestDefaultLayoutAdjustmentOptions(t *testing.T) {
	opts := DefaultLayoutAdjustmentOptions()

	if opts.MinSpacing != 10.0 {
		t.Errorf("MinSpacing = %f, want 10.0", opts.MinSpacing)
	}
	if opts.PageMargin != 20.0 {
		t.Errorf("PageMargin = %f, want 20.0", opts.PageMargin)
	}
	if opts.Strategy != StrategyCompact {
		t.Errorf("Strategy = %s, want %s", opts.Strategy, StrategyCompact)
	}
}

// TestAdjustLayout_TranslationUseCase は翻訳ユースケースのテスト
func TestAdjustLayout_TranslationUseCase(t *testing.T) {
	// 翻訳前の状態（ブロックが近い、上から下の順）
	layout := &PageLayout{
		Width:  595,
		Height: 842,
		TextBlocks: []TextBlock{
			{
				Text: "Short text",
				Rect: Rectangle{X: 50, Y: 680, Width: 200, Height: 20}, // top=700
			},
			{
				Text: "Next paragraph",
				Rect: Rectangle{X: 50, Y: 665, Width: 200, Height: 20}, // top=685, 間隔5px
			},
			{
				Text: "Third paragraph",
				Rect: Rectangle{X: 50, Y: 650, Width: 200, Height: 20}, // top=670, 間隔5px
			},
		},
	}

	// 自動調整（最小間隔10pxを確保）
	opts := LayoutAdjustmentOptions{
		Strategy:   StrategyFlowDown,
		MinSpacing: 10,
	}
	err := layout.AdjustLayout(opts)
	if err != nil {
		t.Fatalf("AdjustLayout failed: %v", err)
	}

	// 重なりがないことを確認
	overlaps := layout.DetectOverlaps()
	if len(overlaps) > 0 {
		t.Errorf("Expected no overlaps after adjustment, got %d", len(overlaps))
	}

	// ブロックが適切な間隔で配置されていることを確認
	// Block 0: bottom=680
	// Block 1: top should be 680-10=670
	block1Bottom := layout.TextBlocks[0].Rect.Y
	block2Top := layout.TextBlocks[1].Rect.Y + layout.TextBlocks[1].Rect.Height
	spacing := block1Bottom - block2Top
	if spacing < 10 {
		t.Errorf("Spacing between blocks 0-1 = %f, want >= 10", spacing)
	}

	// Block 1: bottom should be updated
	// Block 2: top should maintain >= 10px spacing
	block2Bottom := layout.TextBlocks[1].Rect.Y
	block3Top := layout.TextBlocks[2].Rect.Y + layout.TextBlocks[2].Rect.Height
	spacing2 := block2Bottom - block3Top
	if spacing2 < 10 {
		t.Errorf("Spacing between blocks 1-2 = %f, want >= 10", spacing2)
	}
}
