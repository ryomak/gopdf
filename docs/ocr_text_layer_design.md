# OCR Text Layer Design

## 概要

画像ベースのPDF（スキャン文書など）に、テキストレイヤーを追加する機能の設計書。
OCR処理自体はユーザー側で実装し、gopdfはその結果をPDFに埋め込むインターフェースを提供する。

## 目的

- 画像のみのPDFから文字をコピー可能にする
- 検索可能なPDFを生成する
- OCR APIとの統合を容易にする

## ユースケース

### 1. スキャン文書のテキスト化

```go
// 画像をスキャン
images := []string{"scan_page1.jpg", "scan_page2.jpg"}

// OCR処理（外部APIやライブラリを使用）
ocrResults := performOCR(images) // ユーザー実装

// PDFに変換（テキストレイヤー付き）
doc := gopdf.New()
for i, imgPath := range images {
    page := doc.AddPage(gopdf.A4, gopdf.Portrait)

    // 画像を配置
    img, _ := gopdf.LoadJPEG(imgPath)
    page.DrawImage(img, 0, 0, gopdf.A4.Width, gopdf.A4.Height)

    // OCR結果からテキストレイヤーを追加
    page.AddTextLayer(ocrResults[i])
}

doc.SaveToFile("searchable.pdf")
```

### 2. 翻訳サービスとの統合

```go
// 既存の画像PDFを開く
reader, _ := gopdf.Open("scanned.pdf")

// 新しいPDFを作成
doc := gopdf.New()

for i := 0; i < reader.PageCount(); i++ {
    // ページの画像を抽出
    images, _ := reader.ExtractImages(i)

    // OCR処理
    ocrResult := ocrAPI.Process(images[0].Data)

    // 翻訳
    translatedText := translateAPI.Translate(ocrResult.Text)

    // 新しいページを作成
    page := doc.AddPage(gopdf.A4, gopdf.Portrait)
    page.DrawImageFromData(images[0].Data, ...)

    // 翻訳されたテキストレイヤーを追加
    page.AddTextLayer(TextLayer{
        Words: convertToWords(translatedText, ocrResult.Positions),
    })
}
```

## データ構造

### TextLayerWord

個々の単語の情報を保持する構造体。

```go
// TextLayerWord は1つの単語とその位置情報
type TextLayerWord struct {
    Text   string    // 単語のテキスト
    Bounds Rectangle // 位置と範囲（PDF座標系）
    Font   string    // フォント名（オプション）
}
```

### TextLayer

ページ全体のテキストレイヤー情報。

```go
// TextLayer はページのテキストレイヤー
type TextLayer struct {
    Words      []TextLayerWord // 単語のリスト
    RenderMode TextRenderMode  // レンダリングモード
    Opacity    float64         // 不透明度（0.0-1.0）
}

// TextRenderMode はテキストの描画モード
type TextRenderMode int

const (
    TextRenderInvisible TextRenderMode = 3 // 非表示（コピーのみ）
    TextRenderNormal    TextRenderMode = 0 // 通常表示
)
```

### OCRResult（ヘルパー構造体）

OCR APIからの結果を標準化する構造体。

```go
// OCRResult はOCR処理の結果
type OCRResult struct {
    Text  string           // 全体テキスト
    Words []OCRWord        // 個別の単語
}

// OCRWord はOCRで認識された単語
type OCRWord struct {
    Text       string    // 単語
    Confidence float64   // 信頼度（0.0-1.0）
    Bounds     Rectangle // 位置（ピクセル座標）
}
```

## API設計

### Page拡張メソッド

```go
// AddTextLayer はページにテキストレイヤーを追加する
func (p *Page) AddTextLayer(layer TextLayer) error

// AddTextLayerWords は個別の単語を追加する（簡易版）
func (p *Page) AddTextLayerWords(words []TextLayerWord) error

// AddInvisibleText は指定位置に透明テキストを追加
func (p *Page) AddInvisibleText(text string, x, y, width, height float64) error
```

### 座標変換ヘルパー

```go
// ConvertPixelToPDFCoords は画像のピクセル座標をPDF座標に変換
func ConvertPixelToPDFCoords(
    pixelX, pixelY float64,
    imageWidth, imageHeight int,
    pdfWidth, pdfHeight float64,
) (pdfX, pdfY float64)

// OCRResultToTextLayer はOCRResultをTextLayerに変換
func OCRResultToTextLayer(
    result OCRResult,
    imageWidth, imageHeight int,
    pdfWidth, pdfHeight float64,
) TextLayer
```

## PDF実装詳細

### テキストレイヤーの描画

PDFのテキスト描画モードを使用：

```
BT                          % Begin Text
/F1 12 Tf                   % Set font
3 Tr                        % Text rendering mode: invisible
1 0 0 1 100 700 Tm          % Text matrix
(Hello) Tj                  % Show text
ET                          % End Text
```

テキストレンダリングモード：
- `0`: Fill text (通常)
- `1`: Stroke text (輪郭)
- `2`: Fill then stroke
- `3`: Invisible (コピー・検索可能だが表示されない)

### レイヤー順序

1. 画像を描画（背景）
2. テキストレイヤーを描画（前面、通常は透明）

```go
// 内部実装イメージ
func (p *Page) AddTextLayer(layer TextLayer) error {
    // テキストレンダリングモードを設定
    fmt.Fprintf(&p.content, "BT\n")
    fmt.Fprintf(&p.content, "%d Tr\n", layer.RenderMode)

    // 不透明度を設定
    if layer.Opacity < 1.0 {
        fmt.Fprintf(&p.content, "/GS1 gs\n") // Graphics state
    }

    for _, word := range layer.Words {
        // フォントとサイズを設定
        // 位置を設定
        // テキストを描画
    }

    fmt.Fprintf(&p.content, "ET\n")
    return nil
}
```

## 使用例

### 例1: Google Cloud Vision APIとの統合

```go
package main

import (
    "context"
    vision "cloud.google.com/go/vision/apiv1"
    "github.com/ryomak/gopdf"
)

func main() {
    // 画像を読み込み
    imgPath := "scanned_page.jpg"
    img, _ := gopdf.LoadJPEG(imgPath)

    // Google Vision APIでOCR
    ctx := context.Background()
    client, _ := vision.NewImageAnnotatorClient(ctx)
    defer client.Close()

    image := vision.NewImageFromURI(imgPath)
    annotations, _ := client.DetectDocumentText(ctx, image, nil)

    // OCR結果を変換
    var words []gopdf.TextLayerWord
    for _, word := range annotations.Pages[0].Blocks[0].Paragraphs[0].Words {
        bounds := word.BoundingBox.Vertices
        words = append(words, gopdf.TextLayerWord{
            Text: getText(word.Symbols),
            Bounds: gopdf.Rectangle{
                X:      float64(bounds[0].X),
                Y:      float64(bounds[0].Y),
                Width:  float64(bounds[2].X - bounds[0].X),
                Height: float64(bounds[2].Y - bounds[0].Y),
            },
        })
    }

    // PDFを作成
    doc := gopdf.New()
    page := doc.AddPage(gopdf.A4, gopdf.Portrait)

    // 画像を配置
    page.DrawImage(img, 0, 0, gopdf.A4.Width, gopdf.A4.Height)

    // テキストレイヤーを追加
    layer := gopdf.TextLayer{
        Words:      words,
        RenderMode: gopdf.TextRenderInvisible,
        Opacity:    0.0,
    }
    page.AddTextLayer(layer)

    // 保存
    doc.SaveToFile("searchable.pdf")
}
```

### 例2: Tesseract OCRとの統合

```go
package main

import (
    "github.com/otiai10/gosseract/v2"
    "github.com/ryomak/gopdf"
)

func main() {
    client := gosseract.NewClient()
    defer client.Close()

    // 画像を設定
    imgPath := "scanned.jpg"
    client.SetImage(imgPath)

    // OCR実行
    text, _ := client.Text()
    boxes, _ := client.GetBoundingBoxes(gosseract.RIL_WORD)

    // 結果を変換
    var words []gopdf.TextLayerWord
    for _, box := range boxes {
        words = append(words, gopdf.TextLayerWord{
            Text: box.Word,
            Bounds: gopdf.Rectangle{
                X:      float64(box.Box.Min.X),
                Y:      float64(box.Box.Min.Y),
                Width:  float64(box.Box.Dx()),
                Height: float64(box.Box.Dy()),
            },
        })
    }

    // 座標を変換（ピクセル → PDF座標）
    img, _ := gopdf.LoadJPEG(imgPath)
    imageWidth := img.Width
    imageHeight := img.Height

    for i := range words {
        words[i].Bounds = gopdf.ConvertPixelToPDFRect(
            words[i].Bounds,
            imageWidth, imageHeight,
            gopdf.A4.Width, gopdf.A4.Height,
        )
    }

    // PDFを作成
    doc := gopdf.New()
    page := doc.AddPage(gopdf.A4, gopdf.Portrait)
    page.DrawImage(img, 0, 0, gopdf.A4.Width, gopdf.A4.Height)
    page.AddTextLayerWords(words)
    doc.SaveToFile("searchable.pdf")
}
```

### 例3: 簡易版（手動で位置指定）

```go
func main() {
    doc := gopdf.New()
    page := doc.AddPage(gopdf.A4, gopdf.Portrait)

    // 画像を配置
    img, _ := gopdf.LoadJPEG("scan.jpg")
    page.DrawImage(img, 0, 0, gopdf.A4.Width, gopdf.A4.Height)

    // 特定の位置に透明テキストを手動で追加
    // (画像上の特定箇所をコピー可能にする)
    page.AddInvisibleText("Hello World", 100, 700, 200, 20)
    page.AddInvisibleText("Second line", 100, 670, 200, 20)

    doc.SaveToFile("output.pdf")
}
```

## 実装計画

### Phase 1: 基本データ構造

- `TextLayerWord` 構造体
- `TextLayer` 構造体
- `TextRenderMode` 定数

### Phase 2: Page拡張

- `AddTextLayer()` メソッド
- `AddTextLayerWords()` メソッド
- `AddInvisibleText()` メソッド

### Phase 3: 座標変換ヘルパー

- `ConvertPixelToPDFCoords()` 関数
- `ConvertPixelToPDFRect()` 関数
- `OCRResultToTextLayer()` 関数

### Phase 4: 例とテスト

- サンプルコード（Tesseractとの統合例）
- ユニットテスト
- 統合テスト

## 制限事項

- OCR処理自体は提供しない（ユーザー側で実装）
- 複雑なレイアウト（表、複数カラムなど）は基本的なサポートのみ
- テキストの向き（回転）は今後の拡張
- フォントの自動選択は限定的

## 参考資料

- PDF Reference 1.7 - Text Objects
- PDF Reference 1.7 - Text Rendering Modes
- Google Cloud Vision API: https://cloud.google.com/vision/docs/ocr
- Tesseract OCR: https://github.com/tesseract-ocr/tesseract
