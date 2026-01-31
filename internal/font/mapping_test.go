package font

import (
	"testing"
)

func TestMapToStandardFont(t *testing.T) {
	tests := []struct {
		name     string
		fontName string
		isBold   bool
		wantFont StandardFont
		wantOk   bool
	}{
		{
			name:     "direct match Helvetica",
			fontName: "Helvetica",
			isBold:   false,
			wantFont: Helvetica,
			wantOk:   true,
		},
		{
			name:     "direct match Times-Bold",
			fontName: "Times-Bold",
			isBold:   true,
			wantFont: TimesBold,
			wantOk:   true,
		},
		{
			name:     "subset prefix match",
			fontName: "BCDEEE+Helvetica-Bold",
			isBold:   true,
			wantFont: HelveticaBold,
			wantOk:   true,
		},
		{
			name:     "partial match Helvetica",
			fontName: "SomeHelveticaFont",
			isBold:   false,
			wantFont: Helvetica,
			wantOk:   true,
		},
		{
			name:     "partial match Helvetica bold",
			fontName: "SomeHelveticaFont",
			isBold:   true,
			wantFont: HelveticaBold,
			wantOk:   true,
		},
		{
			name:     "partial match Times",
			fontName: "TimesNewRoman",
			isBold:   false,
			wantFont: TimesRoman,
			wantOk:   true,
		},
		{
			name:     "partial match Courier",
			fontName: "CourierNew",
			isBold:   false,
			wantFont: Courier,
			wantOk:   true,
		},
		{
			name:     "Symbol",
			fontName: "Symbol",
			isBold:   false,
			wantFont: Symbol,
			wantOk:   true,
		},
		{
			name:     "ZapfDingbats",
			fontName: "ZapfDingbats",
			isBold:   false,
			wantFont: ZapfDingbats,
			wantOk:   true,
		},
		{
			name:     "Dingbats partial",
			fontName: "SomeDingbats",
			isBold:   false,
			wantFont: ZapfDingbats,
			wantOk:   true,
		},
		{
			name:     "unknown font",
			fontName: "UnknownFont",
			isBold:   false,
			wantFont: "",
			wantOk:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFont, gotOk := MapToStandardFont(tt.fontName, tt.isBold)
			if gotFont != tt.wantFont {
				t.Errorf("MapToStandardFont() font = %v, want %v", gotFont, tt.wantFont)
			}
			if gotOk != tt.wantOk {
				t.Errorf("MapToStandardFont() ok = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestContainsFont(t *testing.T) {
	tests := []struct {
		name     string
		fontName string
		target   string
		want     bool
	}{
		{
			name:     "exact match",
			fontName: "Helvetica",
			target:   "Helvetica",
			want:     true,
		},
		{
			name:     "case insensitive",
			fontName: "HELVETICA",
			target:   "helvetica",
			want:     true,
		},
		{
			name:     "partial match",
			fontName: "Arial-Bold",
			target:   "Arial",
			want:     true,
		},
		{
			name:     "no match",
			fontName: "Arial",
			target:   "Helvetica",
			want:     false,
		},
		{
			name:     "target longer than fontName",
			fontName: "Hel",
			target:   "Helvetica",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ContainsFont(tt.fontName, tt.target)
			if got != tt.want {
				t.Errorf("ContainsFont() = %v, want %v", got, tt.want)
			}
		})
	}
}
