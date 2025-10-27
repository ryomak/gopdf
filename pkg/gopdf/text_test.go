package gopdf

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ryomak/gopdf/internal/font"
)

// TestPageSetFont はフォント設定をテストする
func TestPageSetFont(t *testing.T) {
	doc := New()
	page := doc.AddPage(A4, Portrait)

	err := page.SetFont(font.Helvetica, 12)
	if err != nil {
		t.Fatalf("SetFont() failed: %v", err)
	}

	// フォントが設定されたことを確認
	if page.currentFont == nil {
		t.Error("Expected font to be set")
	}
}

// TestPageDrawText はテキスト描画をテストする
func TestPageDrawText(t *testing.T) {
	doc := New()
	page := doc.AddPage(A4, Portrait)

	// フォントを設定せずに描画しようとするとエラー
	err := page.DrawText("Hello", 100, 700)
	if err == nil {
		t.Error("DrawText() should fail without font set")
	}

	// フォントを設定
	page.SetFont(font.Helvetica, 12)

	// テキストを描画
	err = page.DrawText("Hello, World!", 100, 700)
	if err != nil {
		t.Fatalf("DrawText() failed: %v", err)
	}

	// コンテンツが追加されたことを確認
	if page.content.Len() == 0 {
		t.Error("Expected content to be added")
	}
}

// TestDocumentWithText はテキスト付きPDFの生成をテストする
func TestDocumentWithText(t *testing.T) {
	doc := New()
	page := doc.AddPage(A4, Portrait)

	// フォントを設定してテキストを描画
	page.SetFont(font.Helvetica, 12)
	page.DrawText("Hello, World!", 100, 700)

	var buf bytes.Buffer
	err := doc.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo() failed: %v", err)
	}

	output := buf.String()

	// PDFの必須要素を確認
	if !strings.Contains(output, "%PDF-1.7") {
		t.Error("Output should contain PDF header")
	}

	// フォント辞書が含まれることを確認
	if !strings.Contains(output, "/Font") {
		t.Error("Output should contain Font dictionary")
	}

	// Helveticaフォントが含まれることを確認
	if !strings.Contains(output, "Helvetica") {
		t.Error("Output should contain Helvetica font")
	}

	// テキストが含まれることを確認
	if !strings.Contains(output, "Hello, World!") {
		t.Error("Output should contain the text")
	}

	// BT/ET（テキストオブジェクト）が含まれることを確認
	if !strings.Contains(output, "BT") {
		t.Error("Output should contain BT operator")
	}
	if !strings.Contains(output, "ET") {
		t.Error("Output should contain ET operator")
	}

	// Tf（フォント設定）が含まれることを確認
	if !strings.Contains(output, "Tf") {
		t.Error("Output should contain Tf operator")
	}

	// Tj（テキスト表示）が含まれることを確認
	if !strings.Contains(output, "Tj") {
		t.Error("Output should contain Tj operator")
	}
}

// TestMultipleTextDrawing は複数のテキスト描画をテストする
func TestMultipleTextDrawing(t *testing.T) {
	doc := New()
	page := doc.AddPage(A4, Portrait)

	page.SetFont(font.Helvetica, 12)
	page.DrawText("Line 1", 100, 700)
	page.DrawText("Line 2", 100, 680)
	page.DrawText("Line 3", 100, 660)

	var buf bytes.Buffer
	err := doc.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo() failed: %v", err)
	}

	output := buf.String()

	// 全てのテキストが含まれることを確認
	if !strings.Contains(output, "Line 1") {
		t.Error("Output should contain 'Line 1'")
	}
	if !strings.Contains(output, "Line 2") {
		t.Error("Output should contain 'Line 2'")
	}
	if !strings.Contains(output, "Line 3") {
		t.Error("Output should contain 'Line 3'")
	}
}

// TestDifferentFonts は異なるフォントの使用をテストする
func TestDifferentFonts(t *testing.T) {
	doc := New()
	page := doc.AddPage(A4, Portrait)

	// Helveticaで描画
	page.SetFont(font.Helvetica, 12)
	page.DrawText("Helvetica text", 100, 700)

	// Times-Romanに変更して描画
	page.SetFont(font.TimesRoman, 14)
	page.DrawText("Times text", 100, 680)

	var buf bytes.Buffer
	err := doc.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo() failed: %v", err)
	}

	output := buf.String()

	// 両方のフォントが含まれることを確認
	if !strings.Contains(output, "Helvetica") {
		t.Error("Output should contain Helvetica font")
	}
	if !strings.Contains(output, "Times-Roman") {
		t.Error("Output should contain Times-Roman font")
	}
}
