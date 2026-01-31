package gopdf

import (
	"bytes"
	"strings"
	"testing"

)

// TestPageSetFont はフォント設定をテストする
func TestPageSetFont(t *testing.T) {
	doc := New()
	page := doc.AddPage(PageSizeA4, Portrait)

	err := page.SetFont(FontHelvetica, 12)
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
	page := doc.AddPage(PageSizeA4, Portrait)

	// フォントを設定せずに描画しようとするとエラー
	err := page.DrawText("Hello", 100, 700)
	if err == nil {
		t.Error("DrawText() should fail without font set")
	}

	// フォントを設定
	if err := page.SetFont(FontHelvetica, 12); err != nil {
		t.Fatalf("Failed to set font: %v", err)
	}

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
	page := doc.AddPage(PageSizeA4, Portrait)

	// フォントを設定してテキストを描画
	if err := page.SetFont(FontHelvetica, 12); err != nil {
		t.Fatalf("Failed to set font: %v", err)
	}
	if err := page.DrawText("Hello, World!", 100, 700); err != nil {
		t.Fatalf("Failed to draw text: %v", err)
	}

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
	page := doc.AddPage(PageSizeA4, Portrait)

	if err := page.SetFont(FontHelvetica, 12); err != nil {
		t.Fatalf("Failed to set font: %v", err)
	}
	if err := page.DrawText("Line 1", 100, 700); err != nil {
		t.Fatalf("Failed to draw text: %v", err)
	}
	_ = page.DrawText("Line 2", 100, 680)
	_ = page.DrawText("Line 3", 100, 660)

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
	page := doc.AddPage(PageSizeA4, Portrait)

	// Helveticaで描画
	if err := page.SetFont(FontHelvetica, 12); err != nil {
		t.Fatalf("Failed to set font: %v", err)
	}
	_ = page.DrawText("Helvetica text", 100, 700)

	// Times-Romanに変更して描画
	if err := page.SetFont(FontTimesRoman, 14); err != nil {
		t.Fatalf("Failed to set font: %v", err)
	}
	_ = page.DrawText("Times text", 100, 680)

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

// TestPage_drawTextInternal は drawTextInternal 関数のユニットテストを行う
func TestPage_drawTextInternal(t *testing.T) {
	tests := []struct {
		name          string
		x, y          float64
		fontKey       string
		encodedText   string
		useBrackets   bool
		fontSize      float64
		expectedParts []string
	}{
		{
			name:        "standard font with brackets",
			x:           100.0,
			y:           200.0,
			fontKey:     "F1",
			encodedText: "Hello",
			useBrackets: true,
			fontSize:    12.0,
			expectedParts: []string{
				"BT\n",
				"0 0 0 rg\n",
				"/F1 12.00 Tf\n",
				"100.00 200.00 Td\n",
				"(Hello) Tj\n",
				"ET\n",
			},
		},
		{
			name:        "TTF font with angle brackets",
			x:           50.0,
			y:           300.0,
			fontKey:     "F15",
			encodedText: "3053308230930306306F",
			useBrackets: false,
			fontSize:    14.0,
			expectedParts: []string{
				"BT\n",
				"0 0 0 rg\n",
				"/F15 14.00 Tf\n",
				"50.00 300.00 Td\n",
				"<3053308230930306306F> Tj\n",
				"ET\n",
			},
		},
		{
			name:        "with escaped characters in brackets",
			x:           10.0,
			y:           20.0,
			fontKey:     "F2",
			encodedText: "Hello \\(World\\)",
			useBrackets: true,
			fontSize:    10.0,
			expectedParts: []string{
				"BT\n",
				"0 0 0 rg\n",
				"/F2 10.00 Tf\n",
				"10.00 20.00 Td\n",
				"(Hello \\(World\\)) Tj\n",
				"ET\n",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := New()
			page := doc.AddPage(PageSizeA4, Portrait)
			page.fontSize = tt.fontSize

			page.drawTextInternal(tt.x, tt.y, tt.fontKey, tt.encodedText, tt.useBrackets)

			content := page.content.String()

			// 各期待される部分文字列が含まれているか確認
			for _, part := range tt.expectedParts {
				if !strings.Contains(content, part) {
					t.Errorf("content doesn't contain expected part:\nwant: %q\ngot: %q",
						part, content)
				}
			}

			// 全体の順序も確認
			expectedFull := strings.Join(tt.expectedParts, "")
			if content != expectedFull {
				t.Errorf("content doesn't match expected full output:\nwant: %q\ngot: %q",
					expectedFull, content)
			}
		})
	}
}

// TestPage_textEncodings はテキストエンコーディング関数をテストする
func TestPage_textEncodings(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		method   string // "escape" or "hex"
	}{
		{
			name:     "escape normal text",
			input:    "Hello World",
			expected: "Hello World",
			method:   "escape",
		},
		{
			name:     "escape special characters - parentheses",
			input:    "Hello (World)",
			expected: "Hello \\(World\\)",
			method:   "escape",
		},
		{
			name:     "escape backslash",
			input:    "C:\\path\\to\\file",
			expected: "C:\\\\path\\\\to\\\\file",
			method:   "escape",
		},
		{
			name:     "hex encoding for ASCII",
			input:    "Hello",
			expected: "00480065006C006C006F",
			method:   "hex",
		},
		{
			name:     "hex encoding for Japanese",
			input:    "こんにちは",
			expected: "30533093306B3061306F",
			method:   "hex",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := &Page{}

			var result string
			if tt.method == "escape" {
				result = page.escapeString(tt.input)
			} else {
				result = page.textToHexString(tt.input)
			}

			if result != tt.expected {
				t.Errorf("encoding failed:\ninput: %q\nwant: %q\ngot:  %q",
					tt.input, tt.expected, result)
			}
		})
	}
}
