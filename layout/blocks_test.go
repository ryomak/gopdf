package layout

import (
	"testing"
)

func TestTextBlock_WithY(t *testing.T) {
	tb := TextBlock{
		Text: "Hello",
		Rect: Rectangle{X: 100, Y: 200, Width: 300, Height: 50},
	}

	newBlock := tb.WithY(500)

	// 元のブロックは変更されない
	if tb.Rect.Y != 200 {
		t.Errorf("original TextBlock Y should be 200, got %f", tb.Rect.Y)
	}

	// 新しいブロックは新しいY座標を持つ
	_, newY := newBlock.Position()
	if newY != 500 {
		t.Errorf("new block Y should be 500, got %f", newY)
	}

	// 型がTextBlockであること
	if newBlock.Type() != ContentBlockTypeText {
		t.Errorf("new block type should be text, got %s", newBlock.Type())
	}

	// 他のフィールドは保持されていること
	newTB, ok := newBlock.(TextBlock)
	if !ok {
		t.Fatal("new block should be TextBlock")
	}
	if newTB.Text != "Hello" {
		t.Errorf("new block Text should be 'Hello', got %s", newTB.Text)
	}
	if newTB.Rect.X != 100 {
		t.Errorf("new block X should be 100, got %f", newTB.Rect.X)
	}
}

func TestImageBlock_WithY(t *testing.T) {
	ib := ImageBlock{
		X:            100,
		Y:            200,
		PlacedWidth:  300,
		PlacedHeight: 400,
	}

	newBlock := ib.WithY(600)

	// 元のブロックは変更されない
	if ib.Y != 200 {
		t.Errorf("original ImageBlock Y should be 200, got %f", ib.Y)
	}

	// 新しいブロックは新しいY座標を持つ
	_, newY := newBlock.Position()
	if newY != 600 {
		t.Errorf("new block Y should be 600, got %f", newY)
	}

	// 型がImageBlockであること
	if newBlock.Type() != ContentBlockTypeImage {
		t.Errorf("new block type should be image, got %s", newBlock.Type())
	}

	// 他のフィールドは保持されていること
	newIB, ok := newBlock.(ImageBlock)
	if !ok {
		t.Fatal("new block should be ImageBlock")
	}
	if newIB.X != 100 {
		t.Errorf("new block X should be 100, got %f", newIB.X)
	}
	if newIB.PlacedWidth != 300 {
		t.Errorf("new block PlacedWidth should be 300, got %f", newIB.PlacedWidth)
	}
}

func TestContentBlock_AddToLayout(t *testing.T) {
	pl := &PageLayout{
		Width:  595,
		Height: 842,
	}

	tb := TextBlock{Text: "Test", Rect: Rectangle{X: 100, Y: 700, Width: 200, Height: 50}}
	ib := ImageBlock{X: 100, Y: 500, PlacedWidth: 200, PlacedHeight: 150}

	tb.AddToLayout(pl)
	ib.AddToLayout(pl)

	if len(pl.TextBlocks) != 1 {
		t.Errorf("expected 1 TextBlock, got %d", len(pl.TextBlocks))
	}
	if len(pl.Images) != 1 {
		t.Errorf("expected 1 ImageBlock, got %d", len(pl.Images))
	}
	if pl.TextBlocks[0].Text != "Test" {
		t.Errorf("expected text 'Test', got %s", pl.TextBlocks[0].Text)
	}
}

func TestContentBlock_Interface(t *testing.T) {
	// TextBlockがContentBlockインターフェースを満たすことを確認
	var _ ContentBlock = TextBlock{}

	// ImageBlockがContentBlockインターフェースを満たすことを確認
	var _ ContentBlock = ImageBlock{}
}

func TestContentBlock_WithY_ChainedWithAddToLayout(t *testing.T) {
	pl := &PageLayout{
		Width:  595,
		Height: 842,
	}

	tb := TextBlock{Text: "Hello", Rect: Rectangle{X: 100, Y: 200, Width: 300, Height: 50}}
	ib := ImageBlock{X: 100, Y: 300, PlacedWidth: 200, PlacedHeight: 150}

	// WithYで新しいY座標を設定してからAddToLayout
	tb.WithY(700).AddToLayout(pl)
	ib.WithY(500).AddToLayout(pl)

	if len(pl.TextBlocks) != 1 {
		t.Fatalf("expected 1 TextBlock, got %d", len(pl.TextBlocks))
	}
	if len(pl.Images) != 1 {
		t.Fatalf("expected 1 ImageBlock, got %d", len(pl.Images))
	}

	// 新しいY座標が適用されていることを確認
	if pl.TextBlocks[0].Rect.Y != 700 {
		t.Errorf("expected TextBlock Y to be 700, got %f", pl.TextBlocks[0].Rect.Y)
	}
	if pl.Images[0].Y != 500 {
		t.Errorf("expected ImageBlock Y to be 500, got %f", pl.Images[0].Y)
	}
}
