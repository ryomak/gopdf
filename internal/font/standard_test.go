package font

import (
	"testing"
)

// TestStandardFontNames は標準フォント名をテストする
func TestStandardFontNames(t *testing.T) {
	tests := []struct {
		name     string
		font     StandardFont
		wantName string
	}{
		{"Helvetica", Helvetica, "Helvetica"},
		{"Helvetica-Bold", HelveticaBold, "Helvetica-Bold"},
		{"Helvetica-Oblique", HelveticaOblique, "Helvetica-Oblique"},
		{"Helvetica-BoldOblique", HelveticaBoldOblique, "Helvetica-BoldOblique"},
		{"Times-Roman", TimesRoman, "Times-Roman"},
		{"Times-Bold", TimesBold, "Times-Bold"},
		{"Times-Italic", TimesItalic, "Times-Italic"},
		{"Times-BoldItalic", TimesBoldItalic, "Times-BoldItalic"},
		{"Courier", Courier, "Courier"},
		{"Courier-Bold", CourierBold, "Courier-Bold"},
		{"Courier-Oblique", CourierOblique, "Courier-Oblique"},
		{"Courier-BoldOblique", CourierBoldOblique, "Courier-BoldOblique"},
		{"Symbol", Symbol, "Symbol"},
		{"ZapfDingbats", ZapfDingbats, "ZapfDingbats"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.font.Name() != tt.wantName {
				t.Errorf("StandardFont.Name() = %q, want %q", tt.font.Name(), tt.wantName)
			}
		})
	}
}

// TestStandardFontType は標準フォントのタイプをテストする
func TestStandardFontType(t *testing.T) {
	font := Helvetica
	if font.Type() != "Type1" {
		t.Errorf("StandardFont.Type() = %q, want %q", font.Type(), "Type1")
	}
}

// TestStandardFontEncoding は標準フォントのエンコーディングをテストする
func TestStandardFontEncoding(t *testing.T) {
	font := Helvetica
	if font.Encoding() != "WinAnsiEncoding" {
		t.Errorf("StandardFont.Encoding() = %q, want %q", font.Encoding(), "WinAnsiEncoding")
	}
}

// TestStandardFontIsStandard は標準フォントの判定をテストする
func TestStandardFontIsStandard(t *testing.T) {
	tests := []struct {
		name string
		font StandardFont
		want bool
	}{
		{"Helvetica", Helvetica, true},
		{"Times-Roman", TimesRoman, true},
		{"Courier", Courier, true},
		{"Symbol", Symbol, true},
		{"ZapfDingbats", ZapfDingbats, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.font.IsStandard() != tt.want {
				t.Errorf("StandardFont.IsStandard() = %v, want %v", tt.font.IsStandard(), tt.want)
			}
		})
	}
}

// TestGetStandardFont は名前から標準フォントを取得するテストする
func TestGetStandardFont(t *testing.T) {
	tests := []struct {
		name      string
		fontName  string
		wantFont  StandardFont
		wantError bool
	}{
		{"Helvetica", "Helvetica", Helvetica, false},
		{"Times-Roman", "Times-Roman", TimesRoman, false},
		{"Invalid", "InvalidFont", StandardFont(""), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			font, err := GetStandardFont(tt.fontName)
			if tt.wantError {
				if err == nil {
					t.Error("GetStandardFont() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("GetStandardFont() unexpected error: %v", err)
				}
				if font != tt.wantFont {
					t.Errorf("GetStandardFont() = %q, want %q", font, tt.wantFont)
				}
			}
		})
	}
}
