package gopdf

import (
	"testing"
)

func TestStandardFont_Name(t *testing.T) {
	tests := []struct {
		font     StandardFont
		expected string
	}{
		{FontHelvetica, "Helvetica"},
		{FontHelveticaBold, "Helvetica-Bold"},
		{FontHelveticaOblique, "Helvetica-Oblique"},
		{FontHelveticaBoldOblique, "Helvetica-BoldOblique"},
		{FontTimesRoman, "Times-Roman"},
		{FontTimesBold, "Times-Bold"},
		{FontTimesItalic, "Times-Italic"},
		{FontTimesBoldItalic, "Times-BoldItalic"},
		{FontCourier, "Courier"},
		{FontCourierBold, "Courier-Bold"},
		{FontCourierOblique, "Courier-Oblique"},
		{FontCourierBoldOblique, "Courier-BoldOblique"},
		{FontSymbol, "Symbol"},
		{FontZapfDingbats, "ZapfDingbats"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.font.Name()
			if got != tt.expected {
				t.Errorf("Name() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestStandardFont_IsFont(t *testing.T) {
	// Verify StandardFont implements Font interface
	var f Font = FontHelvetica
	f.isFont() // Should not panic
	if f.Name() != "Helvetica" {
		t.Errorf("Font.Name() = %q, want %q", f.Name(), "Helvetica")
	}

	// Directly call isFont on StandardFont value to cover it
	sf := FontHelvetica
	sf.isFont()
}

func TestStandardFont_StringConversion(t *testing.T) {
	// StandardFont is a string type, verify string conversion
	f := FontHelvetica
	s := string(f)
	if s != "Helvetica" {
		t.Errorf("string(FontHelvetica) = %q, want %q", s, "Helvetica")
	}
}

func TestAllStandardFontsAreDistinct(t *testing.T) {
	fonts := []StandardFont{
		FontHelvetica, FontHelveticaBold, FontHelveticaOblique, FontHelveticaBoldOblique,
		FontTimesRoman, FontTimesBold, FontTimesItalic, FontTimesBoldItalic,
		FontCourier, FontCourierBold, FontCourierOblique, FontCourierBoldOblique,
		FontSymbol, FontZapfDingbats,
	}

	seen := make(map[string]bool)
	for _, f := range fonts {
		name := f.Name()
		if seen[name] {
			t.Errorf("duplicate font name: %q", name)
		}
		seen[name] = true
	}

	if len(seen) != 14 {
		t.Errorf("expected 14 distinct fonts, got %d", len(seen))
	}
}
