package gopdf

import (
	"testing"
)

// TestColorCreation はColor型の作成をテストする
func TestColorCreation(t *testing.T) {
	// 基本色の定義
	black := Color{R: 0, G: 0, B: 0}
	if black.R != 0 || black.G != 0 || black.B != 0 {
		t.Errorf("Black color should be (0,0,0), got (%f,%f,%f)", black.R, black.G, black.B)
	}

	white := Color{R: 1, G: 1, B: 1}
	if white.R != 1 || white.G != 1 || white.B != 1 {
		t.Errorf("White color should be (1,1,1), got (%f,%f,%f)", white.R, white.G, white.B)
	}

	red := Color{R: 1, G: 0, B: 0}
	if red.R != 1 || red.G != 0 || red.B != 0 {
		t.Errorf("Red color should be (1,0,0), got (%f,%f,%f)", red.R, red.G, red.B)
	}
}

// TestNewRGB はRGB値からColor作成をテストする
func TestNewRGB(t *testing.T) {
	tests := []struct {
		name     string
		r, g, b  uint8
		expected Color
	}{
		{"Black", 0, 0, 0, Color{0, 0, 0}},
		{"White", 255, 255, 255, Color{1, 1, 1}},
		{"Red", 255, 0, 0, Color{1, 0, 0}},
		{"Green", 0, 255, 0, Color{0, 1, 0}},
		{"Blue", 0, 0, 255, Color{0, 0, 1}},
		{"Gray", 128, 128, 128, Color{128.0 / 255.0, 128.0 / 255.0, 128.0 / 255.0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			color := NewRGB(tt.r, tt.g, tt.b)
			// 浮動小数点の比較には誤差を許容
			epsilon := 0.001
			if abs(color.R-tt.expected.R) > epsilon ||
				abs(color.G-tt.expected.G) > epsilon ||
				abs(color.B-tt.expected.B) > epsilon {
				t.Errorf("NewRGB(%d,%d,%d) = (%f,%f,%f), want (%f,%f,%f)",
					tt.r, tt.g, tt.b,
					color.R, color.G, color.B,
					tt.expected.R, tt.expected.G, tt.expected.B)
			}
		})
	}
}

// TestPredefinedColors は定義済み色をテストする
func TestPredefinedColors(t *testing.T) {
	if Black.R != 0 || Black.G != 0 || Black.B != 0 {
		t.Error("Black color is incorrect")
	}
	if White.R != 1 || White.G != 1 || White.B != 1 {
		t.Error("White color is incorrect")
	}
	if Red.R != 1 || Red.G != 0 || Red.B != 0 {
		t.Error("Red color is incorrect")
	}
	if Green.R != 0 || Green.G != 1 || Green.B != 0 {
		t.Error("Green color is incorrect")
	}
	if Blue.R != 0 || Blue.G != 0 || Blue.B != 1 {
		t.Error("Blue color is incorrect")
	}
}

// TestLineCapStyle は線の端スタイルをテストする
func TestLineCapStyle(t *testing.T) {
	if ButtCap != 0 {
		t.Errorf("ButtCap should be 0, got %d", ButtCap)
	}
	if RoundCap != 1 {
		t.Errorf("RoundCap should be 1, got %d", RoundCap)
	}
	if SquareCap != 2 {
		t.Errorf("SquareCap should be 2, got %d", SquareCap)
	}
}

// TestLineJoinStyle は線の結合スタイルをテストする
func TestLineJoinStyle(t *testing.T) {
	if MiterJoin != 0 {
		t.Errorf("MiterJoin should be 0, got %d", MiterJoin)
	}
	if RoundJoin != 1 {
		t.Errorf("RoundJoin should be 1, got %d", RoundJoin)
	}
	if BevelJoin != 2 {
		t.Errorf("BevelJoin should be 2, got %d", BevelJoin)
	}
}

// abs は浮動小数点数の絶対値を返す
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// TestPageSetLineWidth はSetLineWidthメソッドをテストする
func TestPageSetLineWidth(t *testing.T) {
	doc := New()
	page := doc.AddPage(PageSizeA4, Portrait)

	// 線の太さを設定
	page.SetLineWidth(2.5)

	// コンテンツストリームに正しいオペレーターが含まれることを確認
	content := page.content.String()
	expected := "2.50 w\n"
	if content != expected {
		t.Errorf("SetLineWidth() content = %q, want %q", content, expected)
	}
}

// TestPageSetStrokeColor はSetStrokeColorメソッドをテストする
func TestPageSetStrokeColor(t *testing.T) {
	doc := New()
	page := doc.AddPage(PageSizeA4, Portrait)

	// ストローク色を赤に設定
	page.SetStrokeColor(Red)

	// コンテンツストリームに正しいオペレーターが含まれることを確認
	content := page.content.String()
	expected := "1.00 0.00 0.00 RG\n"
	if content != expected {
		t.Errorf("SetStrokeColor() content = %q, want %q", content, expected)
	}
}

// TestPageSetFillColor はSetFillColorメソッドをテストする
func TestPageSetFillColor(t *testing.T) {
	doc := New()
	page := doc.AddPage(PageSizeA4, Portrait)

	// 塗りつぶし色を緑に設定
	page.SetFillColor(Green)

	// コンテンツストリームに正しいオペレーターが含まれることを確認
	content := page.content.String()
	expected := "0.00 1.00 0.00 rg\n"
	if content != expected {
		t.Errorf("SetFillColor() content = %q, want %q", content, expected)
	}
}

// TestPageSetLineCap はSetLineCapメソッドをテストする
func TestPageSetLineCap(t *testing.T) {
	doc := New()
	page := doc.AddPage(PageSizeA4, Portrait)

	// 線の端スタイルを丸めるに設定
	page.SetLineCap(RoundCap)

	// コンテンツストリームに正しいオペレーターが含まれることを確認
	content := page.content.String()
	expected := "1 J\n"
	if content != expected {
		t.Errorf("SetLineCap() content = %q, want %q", content, expected)
	}
}

// TestPageSetLineJoin はSetLineJoinメソッドをテストする
func TestPageSetLineJoin(t *testing.T) {
	doc := New()
	page := doc.AddPage(PageSizeA4, Portrait)

	// 線の結合スタイルを丸めるに設定
	page.SetLineJoin(RoundJoin)

	// コンテンツストリームに正しいオペレーターが含まれることを確認
	content := page.content.String()
	expected := "1 j\n"
	if content != expected {
		t.Errorf("SetLineJoin() content = %q, want %q", content, expected)
	}
}

// TestPageDrawLine はDrawLineメソッドをテストする
func TestPageDrawLine(t *testing.T) {
	doc := New()
	page := doc.AddPage(PageSizeA4, Portrait)

	// (100, 100) から (300, 200) まで線を描画
	page.DrawLine(100, 100, 300, 200)

	// コンテンツストリームに正しいオペレーターが含まれることを確認
	content := page.content.String()
	expected := "100.00 100.00 m\n300.00 200.00 l\nS\n"
	if content != expected {
		t.Errorf("DrawLine() content = %q, want %q", content, expected)
	}
}

// TestPageDrawLineWithStyle は線のスタイルを設定してから線を描画するテスト
func TestPageDrawLineWithStyle(t *testing.T) {
	doc := New()
	page := doc.AddPage(PageSizeA4, Portrait)

	// スタイルを設定
	page.SetLineWidth(2.0)
	page.SetStrokeColor(Red)
	page.DrawLine(100, 100, 300, 200)

	// コンテンツストリームに正しいオペレーターが含まれることを確認
	content := page.content.String()
	expected := "2.00 w\n1.00 0.00 0.00 RG\n100.00 100.00 m\n300.00 200.00 l\nS\n"
	if content != expected {
		t.Errorf("DrawLine() with style content = %q, want %q", content, expected)
	}
}

// TestPageDrawRectangle はDrawRectangleメソッドをテストする
func TestPageDrawRectangle(t *testing.T) {
	doc := New()
	page := doc.AddPage(PageSizeA4, Portrait)

	// (100, 200) から幅150, 高さ100の矩形を描画（枠線のみ）
	page.DrawRectangle(100, 200, 150, 100)

	// コンテンツストリームに正しいオペレーターが含まれることを確認
	content := page.content.String()
	expected := "100.00 200.00 150.00 100.00 re\nS\n"
	if content != expected {
		t.Errorf("DrawRectangle() content = %q, want %q", content, expected)
	}
}

// TestPageFillRectangle はFillRectangleメソッドをテストする
func TestPageFillRectangle(t *testing.T) {
	doc := New()
	page := doc.AddPage(PageSizeA4, Portrait)

	// (100, 200) から幅150, 高さ100の矩形を塗りつぶし
	page.FillRectangle(100, 200, 150, 100)

	// コンテンツストリームに正しいオペレーターが含まれることを確認
	content := page.content.String()
	expected := "100.00 200.00 150.00 100.00 re\nf\n"
	if content != expected {
		t.Errorf("FillRectangle() content = %q, want %q", content, expected)
	}
}

// TestPageDrawAndFillRectangle はDrawAndFillRectangleメソッドをテストする
func TestPageDrawAndFillRectangle(t *testing.T) {
	doc := New()
	page := doc.AddPage(PageSizeA4, Portrait)

	// (100, 200) から幅150, 高さ100の矩形を枠線＋塗りつぶし
	page.DrawAndFillRectangle(100, 200, 150, 100)

	// コンテンツストリームに正しいオペレーターが含まれることを確認
	content := page.content.String()
	expected := "100.00 200.00 150.00 100.00 re\nB\n"
	if content != expected {
		t.Errorf("DrawAndFillRectangle() content = %q, want %q", content, expected)
	}
}

// TestPageRectangleWithStyle は矩形描画のスタイル設定をテストする
func TestPageRectangleWithStyle(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*Page)
		method   func(*Page)
		expected string
	}{
		{
			name: "DrawRectangle with stroke color",
			setup: func(p *Page) {
				p.SetStrokeColor(Blue)
				p.SetLineWidth(1.5)
			},
			method: func(p *Page) {
				p.DrawRectangle(100, 200, 150, 100)
			},
			expected: "0.00 0.00 1.00 RG\n1.50 w\n100.00 200.00 150.00 100.00 re\nS\n",
		},
		{
			name: "FillRectangle with fill color",
			setup: func(p *Page) {
				p.SetFillColor(Color{R: 1.0, G: 1.0, B: 0.0})
			},
			method: func(p *Page) {
				p.FillRectangle(300, 200, 150, 100)
			},
			expected: "1.00 1.00 0.00 rg\n300.00 200.00 150.00 100.00 re\nf\n",
		},
		{
			name: "DrawAndFillRectangle with both colors",
			setup: func(p *Page) {
				p.SetStrokeColor(ColorBlack)
				p.SetFillColor(Color{R: 0.8, G: 0.8, B: 0.8})
			},
			method: func(p *Page) {
				p.DrawAndFillRectangle(500, 200, 150, 100)
			},
			expected: "0.00 0.00 0.00 RG\n0.80 0.80 0.80 rg\n500.00 200.00 150.00 100.00 re\nB\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := New()
			page := doc.AddPage(PageSizeA4, Portrait)
			tt.setup(page)
			tt.method(page)

			content := page.content.String()
			if content != tt.expected {
				t.Errorf("content = %q, want %q", content, tt.expected)
			}
		})
	}
}

// TestPageDrawCircle はDrawCircleメソッドをテストする
func TestPageDrawCircle(t *testing.T) {
	doc := New()
	page := doc.AddPage(PageSizeA4, Portrait)

	// 中心(300, 400)、半径50の円を描画（枠線のみ）
	page.DrawCircle(300, 400, 50)

	// コンテンツストリームに円を近似するベジェ曲線のオペレーターが含まれることを確認
	// 円は4つのベジェ曲線で近似される
	content := page.content.String()

	// κ = 4 * (√2 - 1) / 3 ≈ 0.5522847498
	// 50 * κ ≈ 27.614237

	// 正確な値を確認するのではなく、必要なオペレーターが含まれることを確認
	if !containsSubstring(content, "m\n") { // moveto
		t.Error("DrawCircle() should contain moveto operator")
	}
	if !containsSubstring(content, "c\n") { // curveto (4回)
		t.Error("DrawCircle() should contain curveto operators")
	}
	if !containsSubstring(content, "S\n") { // stroke
		t.Error("DrawCircle() should contain stroke operator")
	}
}

// TestPageFillCircle はFillCircleメソッドをテストする
func TestPageFillCircle(t *testing.T) {
	doc := New()
	page := doc.AddPage(PageSizeA4, Portrait)

	// 中心(300, 400)、半径50の円を塗りつぶし
	page.FillCircle(300, 400, 50)

	content := page.content.String()

	// 塗りつぶしオペレーターを確認
	if !containsSubstring(content, "f\n") {
		t.Error("FillCircle() should contain fill operator")
	}
}

// TestPageDrawAndFillCircle はDrawAndFillCircleメソッドをテストする
func TestPageDrawAndFillCircle(t *testing.T) {
	doc := New()
	page := doc.AddPage(PageSizeA4, Portrait)

	// 中心(300, 400)、半径50の円を枠線＋塗りつぶし
	page.DrawAndFillCircle(300, 400, 50)

	content := page.content.String()

	// 枠線＋塗りつぶしオペレーターを確認
	if !containsSubstring(content, "B\n") {
		t.Error("DrawAndFillCircle() should contain fill and stroke operator")
	}
}

// TestCircleWithStyle は円描画のスタイル設定をテストする
func TestCircleWithStyle(t *testing.T) {
	doc := New()
	page := doc.AddPage(PageSizeA4, Portrait)

	// スタイルを設定
	page.SetStrokeColor(Red)
	page.SetFillColor(Color{R: 1.0, G: 0.8, B: 0.8})
	page.DrawAndFillCircle(300, 400, 50)

	content := page.content.String()

	// 色設定とオペレーターが含まれることを確認
	if !containsSubstring(content, "1.00 0.00 0.00 RG\n") {
		t.Error("Circle with style should contain stroke color setting")
	}
	if !containsSubstring(content, "1.00 0.80 0.80 rg\n") {
		t.Error("Circle with style should contain fill color setting")
	}
	if !containsSubstring(content, "B\n") {
		t.Error("Circle with style should contain fill and stroke operator")
	}
}

// containsSubstring は文字列が部分文字列を含むかチェックする
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && indexOfSubstring(s, substr) >= 0
}

// indexOfSubstring は部分文字列のインデックスを返す
func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
