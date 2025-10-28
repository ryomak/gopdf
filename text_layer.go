package gopdf

// TextRenderMode はPDFのテキストレンダリングモード
type TextRenderMode int

const (
	// TextRenderNormal は通常のテキスト表示（塗りつぶし）
	TextRenderNormal TextRenderMode = 0
	// TextRenderStroke はテキストの輪郭のみ表示
	TextRenderStroke TextRenderMode = 1
	// TextRenderFillStroke は塗りつぶしと輪郭の両方
	TextRenderFillStroke TextRenderMode = 2
	// TextRenderInvisible はテキストを非表示（コピー・検索は可能）
	TextRenderInvisible TextRenderMode = 3
)

// TextLayerWord は1つの単語とその位置情報
type TextLayerWord struct {
	Text   string    // 単語のテキスト
	Bounds Rectangle // 位置と範囲（PDF座標系）
}

// TextLayer はページのテキストレイヤー
type TextLayer struct {
	Words      []TextLayerWord // 単語のリスト
	RenderMode TextRenderMode  // レンダリングモード
	Opacity    float64         // 不透明度（0.0-1.0、デフォルト: 0.0 = 完全透明）
}

// DefaultTextLayer はデフォルトのTextLayerを作成（透明テキスト）
func DefaultTextLayer() TextLayer {
	return TextLayer{
		Words:      make([]TextLayerWord, 0),
		RenderMode: TextRenderInvisible,
		Opacity:    0.0,
	}
}

// NewTextLayer は指定された単語リストからTextLayerを作成
func NewTextLayer(words []TextLayerWord) TextLayer {
	return TextLayer{
		Words:      words,
		RenderMode: TextRenderInvisible,
		Opacity:    0.0,
	}
}

// AddWord はTextLayerに単語を追加
func (tl *TextLayer) AddWord(word TextLayerWord) {
	tl.Words = append(tl.Words, word)
}

// ConvertPixelToPDFCoords は画像のピクセル座標をPDF座標に変換
// 画像座標系: 左上が原点 (0,0)、右下が (imageWidth, imageHeight)
// PDF座標系: 左下が原点 (0,0)、右上が (pdfWidth, pdfHeight)
func ConvertPixelToPDFCoords(
	pixelX, pixelY float64,
	imageWidth, imageHeight int,
	pdfWidth, pdfHeight float64,
) (pdfX, pdfY float64) {
	// X方向のスケール
	scaleX := pdfWidth / float64(imageWidth)
	// Y方向のスケール
	scaleY := pdfHeight / float64(imageHeight)

	// X座標は同じ方向
	pdfX = pixelX * scaleX

	// Y座標は反転が必要（上下が逆）
	pdfY = pdfHeight - (pixelY * scaleY)

	return pdfX, pdfY
}

// ConvertPixelToPDFRect は画像のピクセル座標の矩形をPDF座標に変換
func ConvertPixelToPDFRect(
	pixelRect Rectangle,
	imageWidth, imageHeight int,
	pdfWidth, pdfHeight float64,
) Rectangle {
	// 左上の座標を変換
	x, y := ConvertPixelToPDFCoords(
		pixelRect.X, pixelRect.Y,
		imageWidth, imageHeight,
		pdfWidth, pdfHeight,
	)

	// スケールを計算
	scaleX := pdfWidth / float64(imageWidth)
	scaleY := pdfHeight / float64(imageHeight)

	// 幅と高さをスケール
	width := pixelRect.Width * scaleX
	height := pixelRect.Height * scaleY

	// PDF座標系では左下が原点なので、Y座標を調整
	y = y - height

	return Rectangle{
		X:      x,
		Y:      y,
		Width:  width,
		Height: height,
	}
}

// OCRWord はOCRで認識された単語（ピクセル座標）
type OCRWord struct {
	Text       string    // 単語
	Confidence float64   // 信頼度（0.0-1.0）
	Bounds     Rectangle // 位置（ピクセル座標、左上原点）
}

// OCRResult はOCR処理の結果
type OCRResult struct {
	Text  string    // 全体テキスト
	Words []OCRWord // 個別の単語
}

// ToTextLayer はOCRResultをTextLayerに変換
// imageWidth, imageHeight: 元画像のサイズ（ピクセル）
// pdfWidth, pdfHeight: PDFページのサイズ（ポイント）
func (r OCRResult) ToTextLayer(
	imageWidth, imageHeight int,
	pdfWidth, pdfHeight float64,
) TextLayer {
	words := make([]TextLayerWord, 0, len(r.Words))

	for _, ocrWord := range r.Words {
		// ピクセル座標をPDF座標に変換
		pdfBounds := ConvertPixelToPDFRect(
			ocrWord.Bounds,
			imageWidth, imageHeight,
			pdfWidth, pdfHeight,
		)

		words = append(words, TextLayerWord{
			Text:   ocrWord.Text,
			Bounds: pdfBounds,
		})
	}

	return NewTextLayer(words)
}
