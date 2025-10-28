package gopdf

import (
	"math"
	"testing"
)

func TestDefaultTextLayer(t *testing.T) {
	layer := DefaultTextLayer()

	if layer.RenderMode != TextRenderInvisible {
		t.Errorf("DefaultTextLayer().RenderMode = %v, want %v",
			layer.RenderMode, TextRenderInvisible)
	}
	if layer.Opacity != 0.0 {
		t.Errorf("DefaultTextLayer().Opacity = %f, want 0.0", layer.Opacity)
	}
	if len(layer.Words) != 0 {
		t.Errorf("DefaultTextLayer().Words length = %d, want 0", len(layer.Words))
	}
}

func TestNewTextLayer(t *testing.T) {
	words := []TextLayerWord{
		{Text: "Hello", Bounds: Rectangle{X: 10, Y: 20, Width: 50, Height: 12}},
		{Text: "World", Bounds: Rectangle{X: 70, Y: 20, Width: 50, Height: 12}},
	}

	layer := NewTextLayer(words)

	if len(layer.Words) != 2 {
		t.Errorf("NewTextLayer() returned %d words, want 2", len(layer.Words))
	}
	if layer.RenderMode != TextRenderInvisible {
		t.Errorf("NewTextLayer().RenderMode = %v, want %v",
			layer.RenderMode, TextRenderInvisible)
	}
}

func TestTextLayer_AddWord(t *testing.T) {
	layer := DefaultTextLayer()

	word1 := TextLayerWord{Text: "First", Bounds: Rectangle{X: 10, Y: 20, Width: 40, Height: 12}}
	word2 := TextLayerWord{Text: "Second", Bounds: Rectangle{X: 60, Y: 20, Width: 50, Height: 12}}

	layer.AddWord(word1)
	if len(layer.Words) != 1 {
		t.Errorf("After adding 1 word, length = %d, want 1", len(layer.Words))
	}

	layer.AddWord(word2)
	if len(layer.Words) != 2 {
		t.Errorf("After adding 2 words, length = %d, want 2", len(layer.Words))
	}

	if layer.Words[0].Text != "First" {
		t.Errorf("First word text = %q, want %q", layer.Words[0].Text, "First")
	}
	if layer.Words[1].Text != "Second" {
		t.Errorf("Second word text = %q, want %q", layer.Words[1].Text, "Second")
	}
}

func TestConvertPixelToPDFCoords(t *testing.T) {
	tests := []struct {
		name                 string
		pixelX, pixelY       float64
		imageWidth, imageHeight int
		pdfWidth, pdfHeight  float64
		wantX, wantY         float64
	}{
		{
			name:        "Top-left corner",
			pixelX:      0,
			pixelY:      0,
			imageWidth:  1000,
			imageHeight: 1000,
			pdfWidth:    595,
			pdfHeight:   842,
			wantX:       0,
			wantY:       842, // PDF座標では下が0なので、上端はpdfHeight
		},
		{
			name:        "Bottom-right corner",
			pixelX:      1000,
			pixelY:      1000,
			imageWidth:  1000,
			imageHeight: 1000,
			pdfWidth:    595,
			pdfHeight:   842,
			wantX:       595,
			wantY:       0, // PDF座標では下が0
		},
		{
			name:        "Center",
			pixelX:      500,
			pixelY:      500,
			imageWidth:  1000,
			imageHeight: 1000,
			pdfWidth:    595,
			pdfHeight:   842,
			wantX:       297.5,  // 595 * 0.5
			wantY:       421,    // 842 - (842 * 0.5)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotX, gotY := ConvertPixelToPDFCoords(
				tt.pixelX, tt.pixelY,
				tt.imageWidth, tt.imageHeight,
				tt.pdfWidth, tt.pdfHeight,
			)

			if math.Abs(gotX-tt.wantX) > 0.1 {
				t.Errorf("ConvertPixelToPDFCoords() gotX = %f, want %f", gotX, tt.wantX)
			}
			if math.Abs(gotY-tt.wantY) > 0.1 {
				t.Errorf("ConvertPixelToPDFCoords() gotY = %f, want %f", gotY, tt.wantY)
			}
		})
	}
}

func TestConvertPixelToPDFRect(t *testing.T) {
	tests := []struct {
		name                 string
		pixelRect            Rectangle
		imageWidth, imageHeight int
		pdfWidth, pdfHeight  float64
		want                 Rectangle
	}{
		{
			name: "Top-left rectangle",
			pixelRect: Rectangle{
				X:      0,
				Y:      0,
				Width:  100,
				Height: 50,
			},
			imageWidth:  1000,
			imageHeight: 1000,
			pdfWidth:    595,
			pdfHeight:   842,
			want: Rectangle{
				X:      0,
				Y:      842 - (50 * 842.0 / 1000.0), // Y座標は上端 - 高さ
				Width:  59.5,  // 100 * 595 / 1000
				Height: 42.1,  // 50 * 842 / 1000
			},
		},
		{
			name: "Center rectangle",
			pixelRect: Rectangle{
				X:      400,
				Y:      400,
				Width:  200,
				Height: 100,
			},
			imageWidth:  1000,
			imageHeight: 1000,
			pdfWidth:    595,
			pdfHeight:   842,
			want: Rectangle{
				X:      238,       // 400 * 595 / 1000
				Y:      421,       // (842 - 400*842/1000) - height = (842 - 336.8) - 84.2 = 505.2 - 84.2 = 421
				Width:  119,       // 200 * 595 / 1000
				Height: 84.2,      // 100 * 842 / 1000
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertPixelToPDFRect(
				tt.pixelRect,
				tt.imageWidth, tt.imageHeight,
				tt.pdfWidth, tt.pdfHeight,
			)

			if math.Abs(got.X-tt.want.X) > 1.0 {
				t.Errorf("ConvertPixelToPDFRect() X = %f, want %f", got.X, tt.want.X)
			}
			if math.Abs(got.Y-tt.want.Y) > 1.0 {
				t.Errorf("ConvertPixelToPDFRect() Y = %f, want %f", got.Y, tt.want.Y)
			}
			if math.Abs(got.Width-tt.want.Width) > 1.0 {
				t.Errorf("ConvertPixelToPDFRect() Width = %f, want %f", got.Width, tt.want.Width)
			}
			if math.Abs(got.Height-tt.want.Height) > 1.0 {
				t.Errorf("ConvertPixelToPDFRect() Height = %f, want %f", got.Height, tt.want.Height)
			}
		})
	}
}

func TestOCRResult_ToTextLayer(t *testing.T) {
	ocrResult := OCRResult{
		Text: "Hello World",
		Words: []OCRWord{
			{
				Text:       "Hello",
				Confidence: 0.99,
				Bounds: Rectangle{
					X:      10,
					Y:      10,
					Width:  100,
					Height: 20,
				},
			},
			{
				Text:       "World",
				Confidence: 0.98,
				Bounds: Rectangle{
					X:      120,
					Y:      10,
					Width:  100,
					Height: 20,
				},
			},
		},
	}

	imageWidth := 1000
	imageHeight := 1000
	pdfWidth := 595.0
	pdfHeight := 842.0

	layer := ocrResult.ToTextLayer(imageWidth, imageHeight, pdfWidth, pdfHeight)

	if len(layer.Words) != 2 {
		t.Fatalf("ToTextLayer() returned %d words, want 2", len(layer.Words))
	}

	if layer.Words[0].Text != "Hello" {
		t.Errorf("Word[0].Text = %q, want %q", layer.Words[0].Text, "Hello")
	}
	if layer.Words[1].Text != "World" {
		t.Errorf("Word[1].Text = %q, want %q", layer.Words[1].Text, "World")
	}

	// 座標が変換されていることを確認
	if layer.Words[0].Bounds.X <= 0 || layer.Words[0].Bounds.X > pdfWidth {
		t.Errorf("Word[0] X coordinate not properly converted: %f", layer.Words[0].Bounds.X)
	}
}

func TestTextRenderMode_Constants(t *testing.T) {
	tests := []struct {
		name string
		mode TextRenderMode
		want int
	}{
		{"Normal", TextRenderNormal, 0},
		{"Stroke", TextRenderStroke, 1},
		{"FillStroke", TextRenderFillStroke, 2},
		{"Invisible", TextRenderInvisible, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.mode) != tt.want {
				t.Errorf("%s mode = %d, want %d", tt.name, int(tt.mode), tt.want)
			}
		})
	}
}
