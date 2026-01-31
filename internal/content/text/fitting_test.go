package text

import (
	"testing"

	"github.com/ryomak/gopdf/internal/content/layout"
)

func TestWrap(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		maxWidth  float64
		fontSize  float64
		wantLines int
	}{
		{
			name:      "empty text",
			text:      "",
			maxWidth:  100,
			fontSize:  10,
			wantLines: 1,
		},
		{
			name:      "single word",
			text:      "Hello",
			maxWidth:  100,
			fontSize:  10,
			wantLines: 1,
		},
		{
			name:      "multiple words",
			text:      "Hello World",
			maxWidth:  100,
			fontSize:  10,
			wantLines: 1,
		},
		{
			name:      "wrap needed",
			text:      "Hello World This is a long text",
			maxWidth:  50,
			fontSize:  10,
			wantLines: 5, // Each word wraps to its own line at 50px width with 10pt font
		},
		{
			name:      "with newlines",
			text:      "Line1\nLine2\nLine3",
			maxWidth:  100,
			fontSize:  10,
			wantLines: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := Wrap(tt.text, tt.maxWidth, "Helvetica", tt.fontSize, DefaultWidthEstimator)
			if len(lines) != tt.wantLines {
				t.Errorf("Wrap() = %d lines, want %d lines", len(lines), tt.wantLines)
			}
		})
	}
}

func TestFit(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		bounds  layout.Rectangle
		opts    FitOptions
		wantErr bool
	}{
		{
			name: "fits without change",
			text: "Hello",
			bounds: layout.Rectangle{
				X: 0, Y: 0, Width: 100, Height: 50,
			},
			opts:    DefaultFitOptions(),
			wantErr: false,
		},
		{
			name: "needs shrink",
			text: "This is a long text that needs wrapping",
			bounds: layout.Rectangle{
				X: 0, Y: 0, Width: 100, Height: 80,
			},
			opts:    DefaultFitOptions(),
			wantErr: false,
		},
		{
			name: "bounds too small",
			text: "Hello",
			bounds: layout.Rectangle{
				X: 0, Y: 0, Width: 1, Height: 1,
			},
			opts:    DefaultFitOptions(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Fit(tt.text, tt.bounds, "Helvetica", tt.opts, DefaultWidthEstimator)
			if (err != nil) != tt.wantErr {
				t.Errorf("Fit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAlign(t *testing.T) {
	tests := []struct {
		name  string
		align Align
		want  int
	}{
		{"AlignLeft", AlignLeft, 0},
		{"AlignCenter", AlignCenter, 1},
		{"AlignRight", AlignRight, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.align) != tt.want {
				t.Errorf("Align = %d, want %d", tt.align, tt.want)
			}
		})
	}
}

func TestWrapJapanese(t *testing.T) {
	tests := []struct {
		name           string
		text           string
		maxWidth       float64
		fontSize       float64
		wantLines      int
		checkNoSpaces  bool // 行内に不要なスペースがないことを確認
	}{
		{
			name:          "japanese short text",
			text:          "こんにちは",
			maxWidth:      200,
			fontSize:      12,
			wantLines:     1,
			checkNoSpaces: true,
		},
		{
			name:          "japanese text wrapping",
			text:          "これは長い日本語のテキストです。改行が必要になるはずです。",
			maxWidth:      100,
			fontSize:      12,
			wantLines:     7, // DefaultWidthEstimatorは1文字=fontSize*0.6なので100px/7.2≒14文字/行
			checkNoSpaces: true,
		},
		{
			name:          "japanese with newlines preserved",
			text:          "行1です\n行2です\n行3です",
			maxWidth:      200,
			fontSize:      12,
			wantLines:     3,
			checkNoSpaces: true,
		},
		{
			name:          "mixed japanese and english",
			text:          "日本語とEnglishの混在",
			maxWidth:      300, // より広い幅で1行に収める
			fontSize:      12,
			wantLines:     1,
			checkNoSpaces: false, // 英語部分はスペースがある可能性
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := Wrap(tt.text, tt.maxWidth, "NotoSansJP", tt.fontSize, DefaultWidthEstimator)
			if len(lines) != tt.wantLines {
				t.Errorf("Wrap() = %d lines %v, want %d lines", len(lines), lines, tt.wantLines)
			}
			// 日本語テキストの場合、不要なスペースがないことを確認
			if tt.checkNoSpaces {
				for _, line := range lines {
					// 日本語のみの行にスペースがないことを確認
					if containsJapanese(line) && !containsEnglish(line) {
						for _, r := range line {
							if r == ' ' {
								t.Errorf("Wrap() line contains unexpected space: %q", line)
							}
						}
					}
				}
			}
		})
	}
}

// containsEnglish は文字列に英字が含まれるかチェック
func containsEnglish(s string) bool {
	for _, r := range s {
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
			return true
		}
	}
	return false
}

func TestContainsJapanese(t *testing.T) {
	tests := []struct {
		name string
		text string
		want bool
	}{
		{"hiragana", "あいうえお", true},
		{"katakana", "アイウエオ", true},
		{"kanji", "漢字", true},
		{"english only", "Hello World", false},
		{"mixed", "Hello世界", true},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := containsJapanese(tt.text); got != tt.want {
				t.Errorf("containsJapanese(%q) = %v, want %v", tt.text, got, tt.want)
			}
		})
	}
}

func TestIsLineStartProhibited(t *testing.T) {
	tests := []struct {
		name string
		r    rune
		want bool
	}{
		{"comma", '、', true},
		{"period", '。', true},
		{"close paren", '）', true},
		{"close bracket", '」', true},
		{"small a", 'ぁ', true},
		{"normal a", 'あ', false},
		{"kanji", '漢', false},
		{"alpha", 'A', false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isLineStartProhibited(tt.r); got != tt.want {
				t.Errorf("isLineStartProhibited(%c) = %v, want %v", tt.r, got, tt.want)
			}
		})
	}
}
