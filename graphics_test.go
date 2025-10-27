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
