package gopdf

import (
	"bytes"
	"testing"

	"github.com/ryomak/gopdf/internal/font"
)

func TestExtractPageLayout(t *testing.T) {
	// テスト用PDFを生成
	doc := New()
	page := doc.AddPage(A4, Portrait)

	// テキストを追加
	page.SetFont(font.Helvetica, 12)
	page.DrawText("Hello", 100, 700)
	page.DrawText("World", 200, 700)
	page.DrawText("Second Line", 100, 680)

	// PDFをバッファに書き込み
	var buf bytes.Buffer
	if err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("Failed to write PDF: %v", err)
	}

	// PDFを読み込み
	reader, err := OpenReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("Failed to open PDF: %v", err)
	}
	defer reader.Close()

	// レイアウトを抽出
	layout, err := reader.ExtractPageLayout(0)
	if err != nil {
		t.Fatalf("ExtractPageLayout failed: %v", err)
	}

	// ページ番号を検証
	if layout.PageNum != 0 {
		t.Errorf("PageNum = %d, want 0", layout.PageNum)
	}

	// ページサイズを検証
	if layout.Width != 595.0 || layout.Height != 842.0 {
		t.Errorf("Page size = %.1f x %.1f, want 595.0 x 842.0", layout.Width, layout.Height)
	}

	// テキストブロックが存在することを検証
	if len(layout.TextBlocks) == 0 {
		t.Error("Expected at least one text block")
	}

	t.Logf("Found %d text blocks", len(layout.TextBlocks))
	for i, block := range layout.TextBlocks {
		t.Logf("Block %d: %q at (%.1f, %.1f)", i, block.Text, block.Bounds.X, block.Bounds.Y)
	}
}

func TestExtractAllLayouts(t *testing.T) {
	// 複数ページのPDFを生成
	doc := New()

	// ページ1
	page1 := doc.AddPage(A4, Portrait)
	page1.SetFont(font.Helvetica, 12)
	page1.DrawText("Page 1", 100, 700)

	// ページ2
	page2 := doc.AddPage(A4, Portrait)
	page2.SetFont(font.Helvetica, 12)
	page2.DrawText("Page 2", 100, 700)

	// PDFをバッファに書き込み
	var buf bytes.Buffer
	if err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("Failed to write PDF: %v", err)
	}

	// PDFを読み込み
	reader, err := OpenReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("Failed to open PDF: %v", err)
	}
	defer reader.Close()

	// 全レイアウトを抽出
	layouts, err := reader.ExtractAllLayouts()
	if err != nil {
		t.Fatalf("ExtractAllLayouts failed: %v", err)
	}

	// ページ数を検証
	if len(layouts) != 2 {
		t.Errorf("Expected 2 pages, got %d", len(layouts))
	}

	// 各ページを検証
	for i := 0; i < 2; i++ {
		layout, ok := layouts[i]
		if !ok {
			t.Errorf("Page %d not found in layouts", i)
			continue
		}
		if layout.PageNum != i {
			t.Errorf("Page %d: PageNum = %d, want %d", i, layout.PageNum, i)
		}
		t.Logf("Page %d: %d text blocks", i, len(layout.TextBlocks))
	}
}

func TestGroupTextElements(t *testing.T) {
	reader := &PDFReader{}

	// テスト用TextElements
	elements := []TextElement{
		{Text: "Hello", X: 100, Y: 700, Width: 30, Height: 12, Size: 12},
		{Text: "World", X: 135, Y: 700, Width: 30, Height: 12, Size: 12},
		{Text: "Line2", X: 100, Y: 680, Width: 30, Height: 12, Size: 12},
	}

	blocks := reader.groupTextElements(elements)

	// 少なくとも1つのブロックが作成されることを検証
	if len(blocks) == 0 {
		t.Fatal("Expected at least one text block")
	}

	// 最初のブロックのテキストを確認
	t.Logf("Created %d blocks", len(blocks))
	for i, block := range blocks {
		t.Logf("Block %d: %q", i, block.Text)
	}
}

func TestCreateTextBlock(t *testing.T) {
	elements := []TextElement{
		{Text: "Hello", X: 100, Y: 700, Width: 30, Height: 12, Font: "Helvetica", Size: 12},
		{Text: "World", X: 135, Y: 700, Width: 30, Height: 12, Font: "Helvetica", Size: 12},
	}

	block := createTextBlock(elements)

	// テキストが結合されていることを検証
	expectedText := "Hello World"
	if block.Text != expectedText {
		t.Errorf("Text = %q, want %q", block.Text, expectedText)
	}

	// フォントが設定されていることを検証
	if block.Font != "Helvetica" {
		t.Errorf("Font = %q, want %q", block.Font, "Helvetica")
	}

	// フォントサイズが設定されていることを検証
	if block.FontSize != 12.0 {
		t.Errorf("FontSize = %.1f, want 12.0", block.FontSize)
	}

	// バウンディングボックスが正しく計算されていることを検証
	if block.Bounds.X != 100 {
		t.Errorf("Bounds.X = %.1f, want 100.0", block.Bounds.X)
	}
	if block.Bounds.Y != 700 {
		t.Errorf("Bounds.Y = %.1f, want 700.0", block.Bounds.Y)
	}
}

func TestGetPageSize(t *testing.T) {
	// テスト用PDFを生成
	doc := New()
	doc.AddPage(A4, Portrait)

	var buf bytes.Buffer
	if err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("Failed to write PDF: %v", err)
	}

	// PDFを読み込み
	reader, err := OpenReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("Failed to open PDF: %v", err)
	}
	defer reader.Close()

	// ページを取得
	page, err := reader.r.GetPage(0)
	if err != nil {
		t.Fatalf("Failed to get page: %v", err)
	}

	// ページサイズを取得
	width, height := reader.getPageSize(page)

	// A4サイズを検証
	if width != 595.0 || height != 842.0 {
		t.Errorf("Page size = %.1f x %.1f, want 595.0 x 842.0", width, height)
	}
}
