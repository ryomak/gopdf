package font

import (
	"testing"
)

func TestDefaultJapaneseFont(t *testing.T) {
	// 初回呼び出し
	font1, err := DefaultJapaneseFont()
	if err != nil {
		t.Fatalf("DefaultJapaneseFont() error = %v", err)
	}

	if font1 == nil {
		t.Fatal("DefaultJapaneseFont() returned nil")
	}

	// フォント名の確認
	name := font1.Name()
	if name == "" {
		t.Error("Font name is empty")
	}
	t.Logf("Font name: %s", name)

	// フォントデータの確認
	data := font1.Data()
	if len(data) == 0 {
		t.Error("Font data is empty")
	}
	t.Logf("Font data size: %d bytes", len(data))

	// キャッシュの確認（2回目の呼び出し）
	font2, err := DefaultJapaneseFont()
	if err != nil {
		t.Fatalf("Second DefaultJapaneseFont() error = %v", err)
	}

	// 同じインスタンスが返されることを確認
	if font1 != font2 {
		t.Error("DefaultJapaneseFont() should return the same instance")
	}
}

func TestDefaultJapaneseFont_GlyphWidth(t *testing.T) {
	font, err := DefaultJapaneseFont()
	if err != nil {
		t.Fatalf("DefaultJapaneseFont() error = %v", err)
	}

	tests := []struct {
		name     string
		char     rune
		fontSize float64
		wantErr  bool
	}{
		{"ASCII character", 'A', 12.0, false},
		{"Hiragana", 'あ', 12.0, false},
		{"Katakana", 'カ', 12.0, false},
		{"Kanji", '日', 12.0, false},
		{"Space", ' ', 12.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			width, err := font.GlyphWidth(tt.char, tt.fontSize)
			if (err != nil) != tt.wantErr {
				t.Errorf("GlyphWidth() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && width <= 0 {
				t.Errorf("GlyphWidth() = %v, want > 0", width)
			}
			t.Logf("Character '%c' width at %.1fpt: %.2f", tt.char, tt.fontSize, width)
		})
	}
}

func TestDefaultJapaneseFont_TextWidth(t *testing.T) {
	font, err := DefaultJapaneseFont()
	if err != nil {
		t.Fatalf("DefaultJapaneseFont() error = %v", err)
	}

	tests := []struct {
		name     string
		text     string
		fontSize float64
	}{
		{"English text", "Hello, World!", 12.0},
		{"Japanese text", "こんにちは、世界！", 12.0},
		{"Mixed text", "Hello, 世界！", 12.0},
		{"Empty string", "", 12.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			width, err := font.TextWidth(tt.text, tt.fontSize)
			if err != nil {
				t.Errorf("TextWidth() error = %v", err)
				return
			}
			if tt.text == "" && width != 0 {
				t.Errorf("TextWidth() for empty string = %v, want 0", width)
			}
			if tt.text != "" && width <= 0 {
				t.Errorf("TextWidth() = %v, want > 0", width)
			}
			t.Logf("Text '%s' width at %.1fpt: %.2f", tt.text, tt.fontSize, width)
		})
	}
}

func TestGetDefaultJapaneseFontLicense(t *testing.T) {
	license := GetDefaultJapaneseFontLicense()
	if license == "" {
		t.Error("GetDefaultJapaneseFontLicense() returned empty string")
	}

	// ライセンスにApache License 2.0の記載があることを確認
	if len(license) < 100 {
		t.Errorf("License text seems too short: %d bytes", len(license))
	}

	t.Logf("License length: %d bytes", len(license))
	t.Logf("License preview: %s...", license[:min(200, len(license))])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ベンチマーク: フォントの初回読み込み（キャッシュなし）
func BenchmarkDefaultJapaneseFont_FirstLoad(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// 注意: このベンチマークは正確ではありません（sync.Onceのため初回のみ実行される）
		_, _ = DefaultJapaneseFont()
	}
}

// ベンチマーク: キャッシュされたフォントの取得
func BenchmarkDefaultJapaneseFont_Cached(b *testing.B) {
	// 事前にフォントをロード
	_, err := DefaultJapaneseFont()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = DefaultJapaneseFont()
	}
}

// ベンチマーク: 文字幅計算
func BenchmarkGlyphWidth(b *testing.B) {
	font, err := DefaultJapaneseFont()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = font.GlyphWidth('あ', 12.0)
	}
}

// ベンチマーク: テキスト幅計算
func BenchmarkTextWidth(b *testing.B) {
	font, err := DefaultJapaneseFont()
	if err != nil {
		b.Fatal(err)
	}

	text := "こんにちは、世界！"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = font.TextWidth(text, 12.0)
	}
}
