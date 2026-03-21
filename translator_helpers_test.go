package gopdf

import (
	"bytes"
	"fmt"
	"testing"
)

func TestPDFTranslatorOptions_getTargetFontName(t *testing.T) {
	tests := []struct {
		name     string
		opts     PDFTranslatorOptions
		expected string
	}{
		{
			name:     "explicit font name takes priority",
			opts:     PDFTranslatorOptions{TargetFontName: "MyFont", TargetFont: FontHelvetica},
			expected: "MyFont",
		},
		{
			name:     "falls back to font.Name()",
			opts:     PDFTranslatorOptions{TargetFont: FontHelvetica},
			expected: "Helvetica",
		},
		{
			name:     "returns empty when no font set",
			opts:     PDFTranslatorOptions{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.opts.getTargetFontName()
			if got != tt.expected {
				t.Errorf("getTargetFontName() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestPDFTranslatorOptions_getTranslateUnit(t *testing.T) {
	tests := []struct {
		name     string
		unit     TranslateUnit
		expected TranslateUnit
	}{
		{"block", TranslateUnitBlock, TranslateUnitBlock},
		{"line", TranslateUnitLine, TranslateUnitLine},
		{"sentence", TranslateUnitSentence, TranslateUnitSentence},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := PDFTranslatorOptions{TranslateUnit: tt.unit}
			got := opts.getTranslateUnit()
			if got != tt.expected {
				t.Errorf("getTranslateUnit() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDefaultPDFTranslatorOptions(t *testing.T) {
	opts := DefaultPDFTranslatorOptions(FontHelvetica)

	if opts.TargetFont != FontHelvetica {
		t.Errorf("TargetFont = %v, want FontHelvetica", opts.TargetFont)
	}
	if !opts.KeepImages {
		t.Error("KeepImages should be true by default")
	}
	if !opts.KeepLayout {
		t.Error("KeepLayout should be true by default")
	}
	if opts.Translator != nil {
		t.Error("Translator should be nil by default")
	}
}

func TestNewTranslatorOptions(t *testing.T) {
	translator := TranslateFunc(func(text string) (string, error) {
		return "translated", nil
	})

	opts := NewTranslatorOptions(
		WithTranslatorFunc(translator),
		WithTranslatorTargetFont(FontHelvetica),
		WithTranslatorTargetFontName("CustomName"),
		WithTranslatorKeepImages(false),
		WithTranslatorKeepLayout(false),
		WithTranslatorUnit(TranslateUnitSentence),
	)

	if opts.Translator == nil {
		t.Error("Translator should not be nil")
	}
	if opts.TargetFont != FontHelvetica {
		t.Errorf("TargetFont = %v, want FontHelvetica", opts.TargetFont)
	}
	if opts.TargetFontName != "CustomName" {
		t.Errorf("TargetFontName = %q, want %q", opts.TargetFontName, "CustomName")
	}
	if opts.KeepImages {
		t.Error("KeepImages should be false")
	}
	if opts.KeepLayout {
		t.Error("KeepLayout should be false")
	}
	if opts.TranslateUnit != TranslateUnitSentence {
		t.Errorf("TranslateUnit = %v, want TranslateUnitSentence", opts.TranslateUnit)
	}
}

func TestWithTranslatorTargetBoldFont(t *testing.T) {
	opts := NewTranslatorOptions(
		WithTranslatorTargetBoldFont(FontHelveticaBold),
	)
	if opts.TargetBoldFont != FontHelveticaBold {
		t.Errorf("TargetBoldFont = %v, want FontHelveticaBold", opts.TargetBoldFont)
	}
}

func TestWithTranslatorFittingOptions(t *testing.T) {
	fitOpts := NewFitOptions(WithFitMaxFontSize(20))
	opts := NewTranslatorOptions(
		WithTranslatorFittingOptions(fitOpts),
	)
	if opts.FittingOptions.MaxFontSize != 20 {
		t.Errorf("MaxFontSize = %f, want 20", opts.FittingOptions.MaxFontSize)
	}
}

func TestSelectASCIIFont(t *testing.T) {
	tests := []struct {
		name         string
		textBlock    TextBlock
		wantFontName string
	}{
		{
			name: "known Helvetica font",
			textBlock: TextBlock{
				Text: "Hello",
				Font: "Helvetica",
			},
			wantFontName: "Helvetica",
		},
		{
			name: "bold text with unknown font falls back to Helvetica-Bold",
			textBlock: TextBlock{
				Text:   "Bold",
				Font:   "UnknownFont",
				IsBold: true,
			},
			wantFontName: "Helvetica-Bold",
		},
		{
			name: "non-bold text with unknown font falls back to Helvetica",
			textBlock: TextBlock{
				Text: "Normal",
				Font: "UnknownFont",
			},
			wantFontName: "Helvetica",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			font, name, err := selectASCIIFont(tt.textBlock)
			if err != nil {
				t.Fatalf("selectASCIIFont() error: %v", err)
			}
			if name != tt.wantFontName {
				t.Errorf("font name = %q, want %q", name, tt.wantFontName)
			}
			if font == nil {
				t.Error("font should not be nil")
			}
		})
	}
}

func TestSelectNonASCIIFont(t *testing.T) {
	tests := []struct {
		name      string
		textBlock TextBlock
		opts      PDFTranslatorOptions
		wantErr   bool
	}{
		{
			name:      "no target font returns error",
			textBlock: TextBlock{Text: "\u65E5\u672C\u8A9E"},
			opts:      PDFTranslatorOptions{},
			wantErr:   true,
		},
		{
			name:      "with target font succeeds",
			textBlock: TextBlock{Text: "\u65E5\u672C\u8A9E"},
			opts:      PDFTranslatorOptions{TargetFont: FontHelvetica},
			wantErr:   false,
		},
		{
			name:      "bold with bold font",
			textBlock: TextBlock{Text: "\u65E5\u672C\u8A9E", IsBold: true},
			opts:      PDFTranslatorOptions{TargetFont: FontHelvetica, TargetBoldFont: FontHelveticaBold},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := selectNonASCIIFont(tt.textBlock, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("selectNonASCIIFont() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCalculateLineX(t *testing.T) {
	rect := Rectangle{X: 100, Y: 200, Width: 300, Height: 50}

	tests := []struct {
		name      string
		line      string
		fontSize  float64
		alignment Align
		wantX     float64
	}{
		{
			name:      "left alignment",
			line:      "Hello",
			fontSize:  12,
			alignment: AlignLeft,
			wantX:     100, // rect.X
		},
		{
			name:      "center alignment",
			line:      "Hello",
			fontSize:  12,
			alignment: AlignCenter,
			// lineWidth = 5 * 12 * 0.6 = 36, center = 100 + (300-36)/2 = 232
			wantX: 232,
		},
		{
			name:      "right alignment",
			line:      "Hello",
			fontSize:  12,
			alignment: AlignRight,
			// lineWidth = 5 * 12 * 0.6 = 36, right = 100 + 300 - 36 = 364
			wantX: 364,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateLineX(tt.line, tt.fontSize, "Helvetica", rect, tt.alignment)
			if got != tt.wantX {
				t.Errorf("calculateLineX() = %f, want %f", got, tt.wantX)
			}
		})
	}
}

func TestMapToStandardFont(t *testing.T) {
	tests := []struct {
		name     string
		fontName string
		isBold   bool
		wantOK   bool
	}{
		{"Helvetica", "Helvetica", false, true},
		{"Helvetica-Bold", "Helvetica-Bold", false, true},
		{"unknown font", "CompletelyUnknown", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ok := mapToStandardFont(tt.fontName, tt.isBold)
			if ok != tt.wantOK {
				t.Errorf("mapToStandardFont(%q, %v) ok = %v, want %v",
					tt.fontName, tt.isBold, ok, tt.wantOK)
			}
		})
	}
}

func TestLoadImageFromImageInfo(t *testing.T) {
	tests := []struct {
		name    string
		info    ImageInfo
		wantErr bool
	}{
		{
			name:    "empty data",
			info:    ImageInfo{},
			wantErr: true,
		},
		{
			name: "valid image with defaults",
			info: ImageInfo{
				Width:  100,
				Height: 100,
				Data:   []byte{0xFF, 0xD8, 0xFF}, // JPEG-like
			},
			wantErr: false,
		},
		{
			name: "image with flate data prefix",
			info: ImageInfo{
				Width:  50,
				Height: 50,
				Data:   []byte{0x78, 0x9C, 0x01}, // zlib header
			},
			wantErr: false,
		},
		{
			name: "image with explicit colorspace",
			info: ImageInfo{
				Width:      100,
				Height:     100,
				Data:       []byte{0x01, 0x02, 0x03},
				ColorSpace: "DeviceGray",
				BitsPerComp: 4,
				Filter:     "DCTDecode",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img, err := loadImageFromImageInfo(tt.info)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadImageFromImageInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if img.Width != tt.info.Width {
				t.Errorf("Width = %d, want %d", img.Width, tt.info.Width)
			}
			if img.Height != tt.info.Height {
				t.Errorf("Height = %d, want %d", img.Height, tt.info.Height)
			}
			// Check defaults are applied
			if img.ColorSpace == "" {
				t.Error("ColorSpace should not be empty")
			}
			if img.BitsPerComponent == 0 {
				t.Error("BitsPerComponent should not be 0")
			}
		})
	}
}

func TestTranslateTextBlocks(t *testing.T) {
	t.Run("nil translator returns error", func(t *testing.T) {
		err := TranslateTextBlocks(nil, nil)
		if err == nil {
			t.Error("expected error for nil translator")
		}
	})

	t.Run("successful translation", func(t *testing.T) {
		blocks := []TextBlock{
			{Text: "Hello"},
			{Text: "World"},
		}
		translator := TranslateFunc(func(text string) (string, error) {
			return "translated:" + text, nil
		})
		err := TranslateTextBlocks(blocks, translator)
		if err != nil {
			t.Fatalf("TranslateTextBlocks() error: %v", err)
		}
		if blocks[0].Text != "translated:Hello" {
			t.Errorf("blocks[0].Text = %q, want %q", blocks[0].Text, "translated:Hello")
		}
		if blocks[1].Text != "translated:World" {
			t.Errorf("blocks[1].Text = %q, want %q", blocks[1].Text, "translated:World")
		}
	})

	t.Run("translation error propagates", func(t *testing.T) {
		blocks := []TextBlock{
			{Text: "Hello"},
		}
		translator := TranslateFunc(func(text string) (string, error) {
			return "", fmt.Errorf("translation failed")
		})
		err := TranslateTextBlocks(blocks, translator)
		if err == nil {
			t.Error("expected error when translator fails")
		}
	})
}

func TestRenderLayout(t *testing.T) {
	layout := &PageLayout{
		Width:  595,
		Height: 842,
		TextBlocks: []TextBlock{
			{
				Text:     "Test render",
				Font:     "Helvetica",
				FontSize: 12,
				Rect: Rectangle{
					X:      50,
					Y:      700,
					Width:  200,
					Height: 50,
				},
			},
		},
	}

	doc := New()
	opts := PDFTranslatorOptions{
		KeepImages: true,
		KeepLayout: true,
		TargetFont: FontHelvetica,
		FittingOptions: DefaultFitOptions(),
	}

	page, err := RenderLayout(doc, layout, opts)
	if err != nil {
		t.Fatalf("RenderLayout() error: %v", err)
	}
	if page == nil {
		t.Fatal("page should not be nil")
	}

	// Verify the document can be written
	var buf bytes.Buffer
	if err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo() failed: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("output should not be empty")
	}
}

func TestRenderLayout_WithoutImages(t *testing.T) {
	layout := &PageLayout{
		Width:  595,
		Height: 842,
	}

	doc := New()
	opts := PDFTranslatorOptions{
		KeepImages: false,
		KeepLayout: false,
	}

	page, err := RenderLayout(doc, layout, opts)
	if err != nil {
		t.Fatalf("RenderLayout() error: %v", err)
	}
	if page == nil {
		t.Fatal("page should not be nil")
	}
}

func TestTranslatePDFToWriter_NoTranslator(t *testing.T) {
	// Create a simple PDF
	doc := New()
	page := doc.AddPage(PageSizeA4, Portrait)
	_ = page.SetFont(FontHelvetica, 12)
	_ = page.DrawText("Original", 100, 700)

	var inputBuf bytes.Buffer
	if err := doc.WriteTo(&inputBuf); err != nil {
		t.Fatal(err)
	}

	// Translate without translator (just re-render)
	var outputBuf bytes.Buffer
	opts := PDFTranslatorOptions{
		TargetFont:     FontHelvetica,
		KeepImages:     true,
		KeepLayout:     true,
		FittingOptions: DefaultFitOptions(),
	}

	err := TranslatePDFToWriter(bytes.NewReader(inputBuf.Bytes()), &outputBuf, opts)
	if err != nil {
		t.Fatalf("TranslatePDFToWriter() error: %v", err)
	}
	if outputBuf.Len() == 0 {
		t.Error("output should not be empty")
	}
}

func TestSelectFont(t *testing.T) {
	tests := []struct {
		name      string
		textBlock TextBlock
		opts      PDFTranslatorOptions
		wantErr   bool
	}{
		{
			name:      "ASCII text uses standard font",
			textBlock: TextBlock{Text: "Hello", Font: "Helvetica"},
			opts:      PDFTranslatorOptions{},
			wantErr:   false,
		},
		{
			name:      "non-ASCII text without target font fails",
			textBlock: TextBlock{Text: "\u65E5\u672C\u8A9E"},
			opts:      PDFTranslatorOptions{},
			wantErr:   true,
		},
		{
			name:      "non-ASCII text with target font succeeds",
			textBlock: TextBlock{Text: "\u65E5\u672C\u8A9E"},
			opts:      PDFTranslatorOptions{TargetFont: FontHelvetica},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := selectFont(tt.textBlock, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("selectFont() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
