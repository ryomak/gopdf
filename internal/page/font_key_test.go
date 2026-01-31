package page

import (
	"testing"

	"github.com/ryomak/gopdf/internal/font"
)

func TestGetFontKey(t *testing.T) {
	tests := []struct {
		name string
		font font.StandardFont
		want string
	}{
		{"Helvetica", font.Helvetica, "F1"},
		{"HelveticaBold", font.HelveticaBold, "F2"},
		{"HelveticaOblique", font.HelveticaOblique, "F3"},
		{"HelveticaBoldOblique", font.HelveticaBoldOblique, "F4"},
		{"TimesRoman", font.TimesRoman, "F5"},
		{"TimesBold", font.TimesBold, "F6"},
		{"TimesItalic", font.TimesItalic, "F7"},
		{"TimesBoldItalic", font.TimesBoldItalic, "F8"},
		{"Courier", font.Courier, "F9"},
		{"CourierBold", font.CourierBold, "F10"},
		{"CourierOblique", font.CourierOblique, "F11"},
		{"CourierBoldOblique", font.CourierBoldOblique, "F12"},
		{"Symbol", font.Symbol, "F13"},
		{"ZapfDingbats", font.ZapfDingbats, "F14"},
		{"Unknown", font.StandardFont("Unknown"), "F1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetFontKey(tt.font)
			if got != tt.want {
				t.Errorf("GetFontKey(%v) = %v, want %v", tt.font, got, tt.want)
			}
		})
	}
}

// mockTTFFontKeyGetter implements TTFFontKeyGetter for testing
type mockTTFFontKeyGetter struct {
	registeredKeys map[interface{}]string
	fontCount      int
}

func (m *mockTTFFontKeyGetter) GetRegisteredKey(fontPtr interface{}) string {
	if m.registeredKeys == nil {
		return ""
	}
	return m.registeredKeys[fontPtr]
}

func (m *mockTTFFontKeyGetter) GetFontCount() int {
	return m.fontCount
}

func TestGetTTFFontKey(t *testing.T) {
	type testFont struct{ name string }

	font1 := &testFont{name: "Font1"}
	font2 := &testFont{name: "Font2"}

	tests := []struct {
		name     string
		fontPtr  interface{}
		getter   TTFFontKeyGetter
		wantKey  string
	}{
		{
			name:    "registered font returns its key",
			fontPtr: font1,
			getter: &mockTTFFontKeyGetter{
				registeredKeys: map[interface{}]string{font1: "F15"},
				fontCount:      1,
			},
			wantKey: "F15",
		},
		{
			name:    "unregistered font with no existing fonts",
			fontPtr: font1,
			getter: &mockTTFFontKeyGetter{
				registeredKeys: nil,
				fontCount:      0,
			},
			wantKey: "F15",
		},
		{
			name:    "unregistered font with existing fonts",
			fontPtr: font2,
			getter: &mockTTFFontKeyGetter{
				registeredKeys: map[interface{}]string{font1: "F15"},
				fontCount:      1,
			},
			wantKey: "F16",
		},
		{
			name:    "unregistered font with multiple existing fonts",
			fontPtr: &testFont{name: "Font3"},
			getter: &mockTTFFontKeyGetter{
				registeredKeys: nil,
				fontCount:      3,
			},
			wantKey: "F18",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetTTFFontKey(tt.fontPtr, tt.getter)
			if got != tt.wantKey {
				t.Errorf("GetTTFFontKey() = %v, want %v", got, tt.wantKey)
			}
		})
	}
}
