# PDF座標系の仕様

## 概要

gopdfライブラリでは、抽出(Reader)と書き込み(Writer)の両方で**PDFの標準座標系**を使用しています。

## PDF座標系

### 基本仕様

- **原点**: 左下が原点 `(0, 0)`
- **X軸**: 右に行くほど値が大きくなる
- **Y軸**: 上に行くほど値が大きくなる
- **単位**: ポイント (1ポイント = 1/72インチ)

```
  ↑ Y軸
  |
  |  (ページの上部)
  |
  |
  |
  |
  |  (ページの下部)
  └──────────→ X軸
  (0,0)
  左下が原点
```

### コード内での実装

#### Writer (page.go:75)
```go
// DrawText draws text at the specified position.
// The position (x, y) is in PDF units (points), where (0, 0) is the bottom-left corner.
func (p *Page) DrawText(text string, x, y float64) error
```

#### Reader (text_sort.go:10, 27)
```go
// PDFの座標系（左下原点）を考慮し、上から下、左から右の順序にする
// PDFは左下原点なので、Y座標が大きい方が上
```

## 画像/OCRの座標系

画像処理やOCRで使用される座標系は、PDFとは異なる**左上原点**の座標系です。

### 画像座標系

- **原点**: 左上が原点 `(0, 0)`
- **X軸**: 右に行くほど値が大きくなる
- **Y軸**: 下に行くほど値が大きくなる（**PDFと逆**）

```
  (0,0)
  左上が原点
  └──────────→ X軸
  |
  |  (画像の上部)
  |
  |
  |
  |  (画像の下部)
  ↓ Y軸
```

### 座標変換

text_layer.go:53-73 で、画像座標系からPDF座標系への変換関数が提供されています。

```go
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
```

## 座標の扱い方

### テキストの配置

```go
// ページの下部にテキストを配置
page.DrawText("Bottom text", 100, 50)  // Y=50は下部

// ページの上部にテキストを配置
page.DrawText("Top text", 100, 750)    // Y=750は上部（A4サイズの場合）
```

### 画像の配置

```go
// DrawImageも同じPDF座標系を使用
// 左下の座標を指定
page.DrawImage(img, x, y, width, height)
```

### テキスト抽出時の座標

```go
// ExtractPageTextElements で取得したTextElementのX, Y座標もPDF座標系
elements, _ := reader.ExtractPageTextElements(0)
for _, elem := range elements {
	// elem.Y が大きいほど、ページの上部にある
	fmt.Printf("Text: %s at (%f, %f)\n", elem.Text, elem.X, elem.Y)
}
```

## ソート順序

### 上から下、左から右の順序

テキストブロックやコンテンツブロックをソートする際は、Y座標の降順でソートします。

```go
// Y座標の大きい方が上なので、降順でソート
sort.Slice(blocks, func(i, j int) bool {
	// Y座標で比較（上から下）
	if math.Abs(blocks[i].Y - blocks[j].Y) > threshold {
		return blocks[i].Y > blocks[j].Y  // 降順
	}
	// Y座標が同じ場合はX座標でソート（左から右）
	return blocks[i].X < blocks[j].X  // 昇順
})
```

参照:
- `layout/layout.go:58-86`
- `text_sort.go:10-92`

## 注意点

1. **Y座標の大小と視覚的な位置の関係**
   - Y座標が大きい = ページの上部
   - Y座標が小さい = ページの下部
   - 一般的なGUI座標系（左上原点）とは逆

2. **画像座標系との違い**
   - OCRや画像処理の結果を使用する場合は座標変換が必要
   - `ConvertPixelToPDFCoords` 関数を使用

3. **変換マトリックス**
   - PDFの描画では変換マトリックス(CTM)により座標変換が行われることがある
   - 画像配置時には `cm` オペレータで変換マトリックスを適用
   - 参照: `page.go:293-300`, `internal/content/image_extractor.go:182-193`

## 関連ファイル

- `page.go:75` - Writer側の座標系定義
- `text_layer.go:53-73` - 座標変換関数
- `text_sort.go:10-92` - PDF座標系でのソート処理
- `layout/layout.go:58-86` - コンテンツブロックのソート
- `docs/structured_text_extraction_design.md:418-420` - 座標系の説明
