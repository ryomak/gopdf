# OCR Text Layer Example

このサンプルは、画像ベースのPDFにテキストレイヤーを追加して、検索・コピー可能にする機能を実演します。

## 機能

1. **透明テキストレイヤー**: 画像の上に透明なテキストを配置
2. **検索可能PDF**: テキスト検索が可能になる
3. **コピー可能**: 画像から文字をコピーできる
4. **OCR統合**: OCR APIの結果を簡単に埋め込める

## ユースケース

- スキャンした文書をテキスト検索可能にする
- OCR結果をPDFに埋め込む
- 画像ベースのPDFに隠しテキストを追加
- 翻訳サービスと統合（画像+翻訳テキスト）

## 実行方法

```bash
cd examples/12_ocr_text_layer
go run main.go
```

## 出力ファイル

- `simple_invisible.pdf`: シンプルな透明テキストの例
- `ocr_searchable.pdf`: OCR結果をシミュレートした例
- `multiple_words.pdf`: 複数の単語を個別に配置した例

## コード概要

### 1. 簡単な透明テキスト

```go
doc := gopdf.New()
page := doc.AddPage(gopdf.PageSizeA4, gopdf.Portrait)

// 画像を配置
img, _ := gopdf.LoadJPEG("scan.jpg")
page.DrawImage(img, 0, 0, gopdf.PageSizeA4.Width, gopdf.PageSizeA4.Height)

// 透明テキストを追加（検索・コピー可能）
page.AddInvisibleText("Hello, World!", 50, 750, 200, 20)
page.AddInvisibleText("Second line", 50, 700, 200, 20)

doc.SaveToFile("searchable.pdf")
```

### 2. OCR結果の埋め込み

```go
// OCR API から結果を取得（擬似コード）
ocrResult := gopdf.OCRResult{
    Text: "Full document text",
    Words: []gopdf.OCRWord{
        {
            Text:       "Hello",
            Confidence: 0.99,
            Bounds: gopdf.Rectangle{
                X:      10,   // ピクセル座標
                Y:      10,
                Width:  100,
                Height: 20,
            },
        },
        // ... more words
    },
}

// ピクセル座標をPDF座標に変換
imageWidth := 1000  // 元画像の幅
imageHeight := 1000 // 元画像の高さ
textLayer := ocrResult.ToTextLayer(
    imageWidth, imageHeight,
    gopdf.PageSizeA4.Width, gopdf.PageSizeA4.Height,
)

// PDFに追加
page.AddTextLayer(textLayer)
```

### 3. 手動での単語配置

```go
// 単語とその位置を手動で指定
words := []gopdf.TextLayerWord{
    {
        Text: "Hello",
        Bounds: gopdf.Rectangle{
            X: 50, Y: 700, Width: 50, Height: 12,
        },
    },
    {
        Text: "World",
        Bounds: gopdf.Rectangle{
            X: 110, Y: 700, Width: 50, Height: 12,
        },
    },
}

page.AddTextLayerWords(words)
```

## OCR APIとの統合例

### Google Cloud Vision API

```go
import (
    vision "cloud.google.com/go/vision/apiv1"
    "github.com/ryomak/gopdf"
)

func processImageWithVision(imagePath string) (*gopdf.TextLayer, error) {
    ctx := context.Background()
    client, _ := vision.NewImageAnnotatorClient(ctx)
    defer client.Close()

    // OCR実行
    image := vision.NewImageFromURI(imagePath)
    annotation, _ := client.DetectDocumentText(ctx, image, nil)

    // gopdf.OCRWordに変換
    var ocrWords []gopdf.OCRWord
    for _, word := range annotation.Pages[0].Blocks[0].Paragraphs[0].Words {
        vertices := word.BoundingBox.Vertices
        ocrWords = append(ocrWords, gopdf.OCRWord{
            Text:       extractText(word),
            Confidence: word.Confidence,
            Bounds: gopdf.Rectangle{
                X:      float64(vertices[0].X),
                Y:      float64(vertices[0].Y),
                Width:  float64(vertices[2].X - vertices[0].X),
                Height: float64(vertices[2].Y - vertices[0].Y),
            },
        })
    }

    // TextLayerに変換
    result := gopdf.OCRResult{Words: ocrWords}
    return result.ToTextLayer(imageWidth, imageHeight, pdfWidth, pdfHeight), nil
}
```

### Tesseract OCR

```go
import (
    "github.com/otiai10/gosseract/v2"
    "github.com/ryomak/gopdf"
)

func processImageWithTesseract(imagePath string) (*gopdf.TextLayer, error) {
    client := gosseract.NewClient()
    defer client.Close()

    client.SetImage(imagePath)

    // 単語の境界ボックスを取得
    boxes, _ := client.GetBoundingBoxes(gosseract.RIL_WORD)

    // gopdf.TextLayerWordに変換
    var words []gopdf.TextLayerWord
    for _, box := range boxes {
        // ピクセル座標をPDF座標に変換
        pdfBounds := gopdf.ConvertPixelToPDFRect(
            gopdf.Rectangle{
                X:      float64(box.Box.Min.X),
                Y:      float64(box.Box.Min.Y),
                Width:  float64(box.Box.Dx()),
                Height: float64(box.Box.Dy()),
            },
            imageWidth, imageHeight,
            pdfWidth, pdfHeight,
        )

        words = append(words, gopdf.TextLayerWord{
            Text:   box.Word,
            Bounds: pdfBounds,
        })
    }

    return gopdf.NewTextLayer(words), nil
}
```

## API リファレンス

### TextLayerWord 型

```go
type TextLayerWord struct {
    Text   string    // 単語のテキスト
    Bounds Rectangle // 位置と範囲（PDF座標系）
}
```

### TextLayer 型

```go
type TextLayer struct {
    Words      []TextLayerWord // 単語のリスト
    RenderMode TextRenderMode  // レンダリングモード
    Opacity    float64         // 不透明度（0.0-1.0）
}
```

### TextRenderMode 定数

```go
const (
    TextRenderNormal    // 通常表示
    TextRenderStroke    // 輪郭のみ
    TextRenderFillStroke // 塗りと輪郭
    TextRenderInvisible  // 非表示（コピー・検索は可能）
)
```

### OCRWord 型（ヘルパー）

```go
type OCRWord struct {
    Text       string    // 単語
    Confidence float64   // 信頼度（0.0-1.0）
    Bounds     Rectangle // 位置（ピクセル座標）
}
```

### OCRResult 型（ヘルパー）

```go
type OCRResult struct {
    Text  string    // 全体テキスト
    Words []OCRWord // 個別の単語
}
```

### Page メソッド

#### AddTextLayer

```go
func (p *Page) AddTextLayer(layer TextLayer) error
```

ページにテキストレイヤーを追加します。通常は透明にして画像の上に配置されます。

#### AddTextLayerWords

```go
func (p *Page) AddTextLayerWords(words []TextLayerWord) error
```

個別の単語を追加する簡易版メソッド。

#### AddInvisibleText

```go
func (p *Page) AddInvisibleText(text string, x, y, width, height float64) error
```

指定位置に透明テキストを追加します。画像の特定箇所をコピー・検索可能にする簡易メソッド。

### 座標変換関数

#### ConvertPixelToPDFCoords

```go
func ConvertPixelToPDFCoords(
    pixelX, pixelY float64,
    imageWidth, imageHeight int,
    pdfWidth, pdfHeight float64,
) (pdfX, pdfY float64)
```

画像のピクセル座標（左上原点）をPDF座標（左下原点）に変換します。

#### ConvertPixelToPDFRect

```go
func ConvertPixelToPDFRect(
    pixelRect Rectangle,
    imageWidth, imageHeight int,
    pdfWidth, pdfHeight float64,
) Rectangle
```

ピクセル座標の矩形をPDF座標の矩形に変換します。

#### OCRResult.ToTextLayer

```go
func (r OCRResult) ToTextLayer(
    imageWidth, imageHeight int,
    pdfWidth, pdfHeight float64,
) TextLayer
```

OCRResultをTextLayerに変換します。座標変換も自動で行います。

### ヘルパー関数

#### NewTextLayer

```go
func NewTextLayer(words []TextLayerWord) TextLayer
```

単語リストからTextLayerを作成します（デフォルト: 透明・非表示）。

#### DefaultTextLayer

```go
func DefaultTextLayer() TextLayer
```

デフォルトのTextLayer（透明・非表示）を作成します。

## 座標系について

### 画像（ピクセル座標）
- 原点: 左上 (0, 0)
- X軸: 右方向に増加
- Y軸: 下方向に増加

### PDF座標
- 原点: 左下 (0, 0)
- X軸: 右方向に増加
- Y軸: 上方向に増加

座標変換関数を使用すると、この違いを自動的に処理できます。

## 使用例：実践的なワークフロー

```go
package main

import (
    "github.com/ryomak/gopdf"
)

func createSearchablePDF(imagePath, outputPath string) error {
    // 1. OCR処理（外部APIまたはライブラリ）
    ocrResult := performOCR(imagePath) // ユーザー実装

    // 2. 画像サイズを取得
    img, _ := gopdf.LoadJPEG(imagePath)
    imageWidth := img.Width
    imageHeight := img.Height

    // 3. PDFを作成
    doc := gopdf.New()
    page := doc.AddPage(gopdf.PageSizeA4, gopdf.Portrait)

    // 4. 画像を配置
    page.DrawImage(img, 0, 0, gopdf.PageSizeA4.Width, gopdf.PageSizeA4.Height)

    // 5. テキストレイヤーを追加
    textLayer := ocrResult.ToTextLayer(
        imageWidth, imageHeight,
        gopdf.PageSizeA4.Width, gopdf.PageSizeA4.Height,
    )
    page.AddTextLayer(textLayer)

    // 6. 保存
    return doc.SaveToFile(outputPath)
}
```

## トラブルシューティング

### テキストが検索できない

→ テキストレイヤーが正しく追加されているか確認してください。`TextRenderInvisible`モードを使用していることを確認してください。

### テキストが画像とずれている

→ 座標変換が正しく行われていない可能性があります。`ConvertPixelToPDFRect`を使用して、ピクセル座標をPDF座標に変換してください。

### PDFが重い

→ 画像の解像度を下げるか、JPEGの品質を調整してください。テキストレイヤー自体はサイズにほとんど影響しません。

## 参考

- 設計書: [docs/ocr_text_layer_design.md](../../docs/ocr_text_layer_design.md)
- Google Cloud Vision: https://cloud.google.com/vision/docs/ocr
- Tesseract OCR: https://github.com/tesseract-ocr/tesseract
- Gosseract (Go binding): https://github.com/otiai10/gosseract
