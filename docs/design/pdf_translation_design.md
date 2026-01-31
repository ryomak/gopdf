# Phase 11: PDF翻訳機能 設計書

## 1. 概要

既存PDFのレイアウトを維持したまま、テキストコンテンツを翻訳した新しいPDFを生成する機能を実装する。

### 1.1. 目的

- 既存PDFから要素（テキスト、画像）とそのレイアウト情報を抽出
- 抽出した情報を元に、レイアウトを保持したまま翻訳後のPDFを生成
- 画像のサイズと位置を正確に保持
- テキストを矩形領域内に自動フィッティング

### 1.2. ユースケース

```
元のPDF（英語）             翻訳後のPDF（日本語）
┌──────────────┐         ┌──────────────┐
│ Technical Report │         │ 技術レポート      │
│                  │         │                  │
│ [Image: Graph]   │   →    │ [Image: Graph]   │
│                  │         │                  │
│ This document... │         │ このドキュメント...│
└──────────────┘         └──────────────┘
```

### 1.3. スコープ

**Phase 11で実装する機能:**
- ✅ ページレイアウトの完全解析（テキスト＋画像の位置・サイズ）
- ✅ 画像位置情報の取得（コンテンツストリーム解析）
- ✅ テキスト自動フィッティング（矩形領域内で自動サイズ調整）
- ✅ 翻訳インターフェース定義
- ✅ レイアウト保持型PDF生成
- ✅ 多言語対応（日本語TTFフォント対応）

**Phase 11では実装しない機能:**
- 実際の翻訳処理（外部API連携は使用側で実装）
- 複雑な表構造の完全再現
- 回転・斜体などの複雑な変形の完全再現
- アノテーション（リンク、コメント）の保持

## 2. アーキテクチャ

### 2.1. 全体フロー

```
┌─────────────────────────────────────────────────────┐
│ 1. 元PDFを読み込み                                    │
│    reader := gopdf.Open("original.pdf")             │
└───────────────────┬─────────────────────────────────┘
                    │
┌───────────────────▼─────────────────────────────────┐
│ 2. ページレイアウトを解析                             │
│    layout := reader.ExtractPageLayout(0)            │
│                                                      │
│    layout.TextBlocks  → []TextBlock                 │
│    layout.Images      → []ImageBlock                │
└───────────────────┬─────────────────────────────────┘
                    │
┌───────────────────▼─────────────────────────────────┐
│ 3. テキストを翻訳（ユーザー実装）                     │
│    for _, block := range layout.TextBlocks {        │
│        block.Text = translate(block.Text)           │
│    }                                                 │
└───────────────────┬─────────────────────────────────┘
                    │
┌───────────────────▼─────────────────────────────────┐
│ 4. レイアウトを保持してPDF生成                        │
│    translator := gopdf.NewPDFTranslator()           │
│    translator.RenderLayout(layout, japaneseFont)    │
│    translator.WriteTo("translated.pdf")             │
└─────────────────────────────────────────────────────┘
```

### 2.2. モジュール構成

```
gopdf/
├── layout.go                    # 公開API: PageLayout, TextBlock, ImageBlock
├── layout_extractor.go          # レイアウト抽出ロジック
├── text_fitting.go              # テキスト自動フィッティング
├── translator.go                # 翻訳インターフェースとPDF生成
├── internal/content/
│   ├── image_extractor.go       # 拡張: 画像位置情報
│   └── graphics_state.go        # CTM追跡機能
└── examples/10_pdf_translation/ # サンプル実装
```

## 3. データ構造

### 3.1. PageLayout（ページレイアウト）

```go
package gopdf

// PageLayout はページの完全なレイアウト情報
type PageLayout struct {
    PageNum    int              // ページ番号
    Width      float64          // ページ幅
    Height     float64          // ページ高さ
    TextBlocks []TextBlock      // テキストブロック
    Images     []ImageBlock     // 画像ブロック
}

// TextBlock はテキストの論理的なブロック
type TextBlock struct {
    Text      string           // テキスト内容
    Elements  []TextElement    // 構成要素（Phase 8で定義済み）
    Bounds    Rectangle        // バウンディングボックス
    Font      string           // 主要フォント
    FontSize  float64          // 主要フォントサイズ
    Color     Color            // テキスト色
}

// ImageBlock は画像の配置情報
type ImageBlock struct {
    ImageInfo                  // 画像データ（Phase 9で定義済み）
    X         float64         // 配置X座標
    Y         float64         // 配置Y座標
    Width     float64         // 表示幅
    Height    float64         // 表示高さ
    Transform Matrix          // 変換行列
}

// Rectangle は矩形領域
type Rectangle struct {
    X      float64  // 左下X座標
    Y      float64  // 左下Y座標
    Width  float64  // 幅
    Height float64  // 高さ
}

// Matrix は変換行列（PDF仕様のCTM）
type Matrix struct {
    A, B, C, D, E, F float64  // [a b c d e f]
}

// Color はRGB色
type Color struct {
    R, G, B float64  // 0.0 ~ 1.0
}
```

### 3.2. Translator Interface

```go
package gopdf

// Translator はテキスト翻訳のインターフェース
type Translator interface {
    // Translate はテキストを翻訳する
    Translate(text string) (string, error)
}

// TranslateFunc は関数型Translator
type TranslateFunc func(string) (string, error)

func (f TranslateFunc) Translate(text string) (string, error) {
    return f(text)
}
```

## 4. 公開API

### 4.1. レイアウト抽出

```go
package gopdf

// ExtractPageLayout はページの完全なレイアウト情報を抽出
func (r *PDFReader) ExtractPageLayout(pageNum int) (*PageLayout, error)

// ExtractAllLayouts は全ページのレイアウトを抽出
func (r *PDFReader) ExtractAllLayouts() (map[int]*PageLayout, error)
```

### 4.2. テキストブロック化

```go
// GroupTextElements はTextElementsをTextBlocksにグループ化
func GroupTextElements(elements []TextElement) []TextBlock
```

### 4.3. テキスト自動フィッティング

```go
// FitTextOptions はテキストフィッティングのオプション
type FitTextOptions struct {
    MaxFontSize   float64  // 最大フォントサイズ
    MinFontSize   float64  // 最小フォントサイズ
    LineSpacing   float64  // 行間倍率（1.0 = フォントサイズと同じ）
    Padding       float64  // パディング
    AllowShrink   bool     // 縮小を許可
    AllowGrow     bool     // 拡大を許可
    Alignment     TextAlign // テキスト配置
}

type TextAlign int

const (
    AlignLeft TextAlign = iota
    AlignCenter
    AlignRight
)

// FitText は矩形領域内にテキストをフィッティング
func FitText(text string, bounds Rectangle, font string, opts FitTextOptions) (*FittedText, error)

// FittedText はフィッティング結果
type FittedText struct {
    Lines      []string  // 改行されたテキスト
    FontSize   float64   // 調整後のフォントサイズ
    LineHeight float64   // 行の高さ
}
```

### 4.4. PDF翻訳

```go
// PDFTranslatorOptions は翻訳オプション
type PDFTranslatorOptions struct {
    Translator     Translator       // 翻訳インターフェース
    TargetFont     *Font            // ターゲット言語のフォント
    FittingOptions FitTextOptions   // テキストフィッティングオプション
    KeepImages     bool             // 画像を保持（デフォルト: true）
    KeepLayout     bool             // レイアウトを保持（デフォルト: true）
}

// TranslatePDF はPDFを翻訳
func TranslatePDF(inputPath string, outputPath string, opts PDFTranslatorOptions) error

// TranslatePage はページを翻訳して新しいPageを返す
func TranslatePage(layout *PageLayout, opts PDFTranslatorOptions) (*Page, error)
```

## 5. 実装の詳細

### 5.1. 画像位置情報の取得

画像の位置情報を取得するには、コンテンツストリームの`Do`オペレーターと現在の変換行列（CTM）を追跡する必要があります。

#### 5.1.1. グラフィックスステート拡張

```go
package content

// GraphicsState は現在のグラフィックス状態
type GraphicsState struct {
    CTM           Matrix       // Current Transformation Matrix
    TextState     TextState    // テキスト状態（Phase 7で定義済み）
    ColorSpace    string       // 色空間
    StrokeColor   [3]float64   // 線の色
    FillColor     [3]float64   // 塗りつぶし色
    LineWidth     float64      // 線幅
}

// Matrix は変換行列
type Matrix struct {
    A, B, C, D, E, F float64
}

// Identity は単位行列
func Identity() Matrix {
    return Matrix{A: 1, B: 0, C: 0, D: 1, E: 0, F: 0}
}

// Multiply は行列の乗算
func (m Matrix) Multiply(other Matrix) Matrix {
    return Matrix{
        A: m.A*other.A + m.B*other.C,
        B: m.A*other.B + m.B*other.D,
        C: m.C*other.A + m.D*other.C,
        D: m.C*other.B + m.D*other.D,
        E: m.E*other.A + m.F*other.C + other.E,
        F: m.E*other.B + m.F*other.D + other.F,
    }
}

// TransformPoint は座標を変換
func (m Matrix) TransformPoint(x, y float64) (float64, float64) {
    return m.A*x + m.C*y + m.E, m.B*x + m.D*y + m.F
}
```

#### 5.1.2. 画像抽出の拡張

`internal/content/image_extractor.go`を拡張：

```go
package content

// ImageExtractorV2 は位置情報付き画像抽出
type ImageExtractorV2 struct {
    reader     *reader.Reader
    operations []Operation  // コンテンツストリームのオペレーション
}

// ExtractImagesWithPosition は位置情報付きで画像を抽出
func (e *ImageExtractorV2) ExtractImagesWithPosition(page core.Dictionary) ([]ImageBlock, error) {
    // グラフィックス状態スタック
    gsStack := []GraphicsState{{CTM: Identity()}}

    var images []ImageBlock

    // コンテンツストリームを解析
    for _, op := range e.operations {
        switch op.Operator {
        case "cm":  // 変換行列の変更
            if len(op.Operands) == 6 {
                matrix := Matrix{
                    A: op.Operands[0].(float64),
                    B: op.Operands[1].(float64),
                    C: op.Operands[2].(float64),
                    D: op.Operands[3].(float64),
                    E: op.Operands[4].(float64),
                    F: op.Operands[5].(float64),
                }
                currentGS := &gsStack[len(gsStack)-1]
                currentGS.CTM = currentGS.CTM.Multiply(matrix)
            }

        case "Do":  // XObjectの描画
            if len(op.Operands) == 1 {
                name := op.Operands[0].(core.Name)

                // 画像XObjectを取得
                imgXObj, err := e.getImageXObject(page, name)
                if err != nil {
                    continue
                }

                // 現在のCTMを取得
                currentCTM := gsStack[len(gsStack)-1].CTM

                // 画像のデフォルトは1x1の単位正方形
                // CTMでスケール・位置が決まる
                x, y := currentCTM.TransformPoint(0, 0)
                w := currentCTM.A  // スケール
                h := currentCTM.D  // スケール

                images = append(images, ImageBlock{
                    ImageInfo: imgXObj.ToImageInfo(string(name)),
                    X:         x,
                    Y:         y,
                    Width:     w,
                    Height:    h,
                    Transform: currentCTM,
                })
            }

        case "q":  // グラフィックス状態の保存
            gsStack = append(gsStack, gsStack[len(gsStack)-1])

        case "Q":  // グラフィックス状態の復元
            if len(gsStack) > 1 {
                gsStack = gsStack[:len(gsStack)-1]
            }
        }
    }

    return images, nil
}
```

### 5.2. テキストのグループ化

近接するTextElementsを論理的なTextBlockにグループ化：

```go
package gopdf

// GroupTextElements はTextElementsをTextBlocksにグループ化
func GroupTextElements(elements []TextElement) []TextBlock {
    if len(elements) == 0 {
        return nil
    }

    // 1. Y座標でソート（上から下）
    sorted := make([]TextElement, len(elements))
    copy(sorted, elements)
    sort.Slice(sorted, func(i, j int) bool {
        return sorted[i].Y > sorted[j].Y
    })

    var blocks []TextBlock
    var currentBlock []TextElement
    threshold := 5.0  // ピクセル単位の閾値

    for i, elem := range sorted {
        if i == 0 {
            currentBlock = []TextElement{elem}
            continue
        }

        // 前の要素との距離を計算
        prev := sorted[i-1]
        xDist := math.Abs(elem.X - (prev.X + prev.Width))
        yDist := math.Abs(elem.Y - prev.Y)

        // 近接している場合は同じブロック
        if yDist < threshold && xDist < prev.Size*2 {
            currentBlock = append(currentBlock, elem)
        } else {
            // 新しいブロック
            if len(currentBlock) > 0 {
                blocks = append(blocks, createTextBlock(currentBlock))
            }
            currentBlock = []TextElement{elem}
        }
    }

    // 最後のブロック
    if len(currentBlock) > 0 {
        blocks = append(blocks, createTextBlock(currentBlock))
    }

    return blocks
}

// createTextBlock はTextElementsからTextBlockを作成
func createTextBlock(elements []TextElement) TextBlock {
    // バウンディングボックスを計算
    minX, minY := elements[0].X, elements[0].Y
    maxX, maxY := elements[0].X+elements[0].Width, elements[0].Y+elements[0].Height

    var text strings.Builder
    var totalSize float64

    for i, elem := range elements {
        if i > 0 {
            text.WriteString(" ")
        }
        text.WriteString(elem.Text)

        totalSize += elem.Size

        minX = math.Min(minX, elem.X)
        minY = math.Min(minY, elem.Y)
        maxX = math.Max(maxX, elem.X+elem.Width)
        maxY = math.Max(maxY, elem.Y+elem.Height)
    }

    avgSize := totalSize / float64(len(elements))

    return TextBlock{
        Text:     text.String(),
        Elements: elements,
        Bounds: Rectangle{
            X:      minX,
            Y:      minY,
            Width:  maxX - minX,
            Height: maxY - minY,
        },
        Font:     elements[0].Font,
        FontSize: avgSize,
        Color:    Color{R: 0, G: 0, B: 0},  // デフォルト黒
    }
}
```

### 5.3. テキスト自動フィッティング

```go
package gopdf

import (
    "strings"
    "unicode/utf8"
)

// FitText は矩形領域内にテキストをフィッティング
func FitText(text string, bounds Rectangle, fontName string, opts FitTextOptions) (*FittedText, error) {
    // パディングを考慮
    availWidth := bounds.Width - opts.Padding*2
    availHeight := bounds.Height - opts.Padding*2

    if availWidth <= 0 || availHeight <= 0 {
        return nil, fmt.Errorf("bounds too small")
    }

    // 2分探索でフォントサイズを決定
    minSize := opts.MinFontSize
    maxSize := opts.MaxFontSize
    bestFit := &FittedText{}

    for maxSize-minSize > 0.1 {
        midSize := (minSize + maxSize) / 2
        lineHeight := midSize * opts.LineSpacing

        // テキストを改行
        lines := wrapText(text, availWidth, fontName, midSize)
        totalHeight := float64(len(lines)) * lineHeight

        if totalHeight <= availHeight {
            // 収まる場合
            bestFit = &FittedText{
                Lines:      lines,
                FontSize:   midSize,
                LineHeight: lineHeight,
            }
            if opts.AllowGrow {
                minSize = midSize  // もっと大きくできるか試す
            } else {
                break
            }
        } else {
            // 収まらない場合
            maxSize = midSize  // 小さくする
        }
    }

    if bestFit.FontSize == 0 {
        return nil, fmt.Errorf("text does not fit in bounds")
    }

    return bestFit, nil
}

// wrapText はテキストを指定幅で改行
func wrapText(text string, maxWidth float64, fontName string, fontSize float64) []string {
    words := strings.Fields(text)
    var lines []string
    var currentLine strings.Builder

    for _, word := range words {
        testLine := currentLine.String()
        if testLine != "" {
            testLine += " "
        }
        testLine += word

        // テキスト幅を計算
        width := estimateTextWidth(testLine, fontSize, fontName)

        if width <= maxWidth {
            if currentLine.Len() > 0 {
                currentLine.WriteString(" ")
            }
            currentLine.WriteString(word)
        } else {
            // 現在の行を確定
            if currentLine.Len() > 0 {
                lines = append(lines, currentLine.String())
                currentLine.Reset()
            }
            currentLine.WriteString(word)
        }
    }

    if currentLine.Len() > 0 {
        lines = append(lines, currentLine.String())
    }

    return lines
}

// estimateTextWidth はテキストの幅を概算
func estimateTextWidth(text string, fontSize float64, fontName string) float64 {
    // ASCII文字は幅0.5em、日本語は1em
    totalWidth := 0.0
    for _, r := range text {
        if r < 128 {
            totalWidth += 0.5
        } else {
            totalWidth += 1.0
        }
    }
    return totalWidth * fontSize
}
```

### 5.4. レイアウト保持型PDF生成

```go
package gopdf

// TranslatePDF はPDFを翻訳
func TranslatePDF(inputPath string, outputPath string, opts PDFTranslatorOptions) error {
    // 1. 元PDFを読み込み
    reader, err := Open(inputPath)
    if err != nil {
        return fmt.Errorf("failed to open input PDF: %w", err)
    }
    defer reader.Close()

    // 2. 新しいPDFドキュメントを作成
    doc := New()

    // 3. 各ページを処理
    pageCount := reader.GetPageCount()
    for i := 0; i < pageCount; i++ {
        layout, err := reader.ExtractPageLayout(i)
        if err != nil {
            return fmt.Errorf("failed to extract layout from page %d: %w", i, err)
        }

        // 4. テキストを翻訳
        if opts.Translator != nil {
            for j := range layout.TextBlocks {
                translated, err := opts.Translator.Translate(layout.TextBlocks[j].Text)
                if err != nil {
                    return fmt.Errorf("translation failed: %w", err)
                }
                layout.TextBlocks[j].Text = translated
            }
        }

        // 5. ページを生成
        page, err := RenderLayout(doc, layout, opts)
        if err != nil {
            return fmt.Errorf("failed to render page %d: %w", i, err)
        }
        _ = page
    }

    // 6. 出力
    file, err := os.Create(outputPath)
    if err != nil {
        return fmt.Errorf("failed to create output file: %w", err)
    }
    defer file.Close()

    return doc.WriteTo(file)
}

// RenderLayout はPageLayoutからPageを生成
func RenderLayout(doc *Document, layout *PageLayout, opts PDFTranslatorOptions) (*Page, error) {
    // カスタムサイズでページを追加
    page := doc.AddPageWithSize(layout.Width, layout.Height)

    // 画像を配置
    if opts.KeepImages {
        for _, img := range layout.Images {
            // 画像データからImageを作成
            pdfImage, err := loadImageFromData(img.Data, img.Format)
            if err != nil {
                continue
            }
            page.DrawImage(pdfImage, img.X, img.Y, img.Width, img.Height)
        }
    }

    // テキストを配置
    if opts.KeepLayout {
        for _, block := range layout.TextBlocks {
            // テキストをフィッティング
            fitted, err := FitText(block.Text, block.Bounds, opts.TargetFont.Name, opts.FittingOptions)
            if err != nil {
                // フィッティングできない場合は元のサイズを使用
                page.SetFont(opts.TargetFont, block.FontSize)
                page.DrawText(block.Text, block.Bounds.X, block.Bounds.Y)
                continue
            }

            // 複数行を描画
            page.SetFont(opts.TargetFont, fitted.FontSize)
            y := block.Bounds.Y + block.Bounds.Height - fitted.LineHeight
            for _, line := range fitted.Lines {
                page.DrawText(line, block.Bounds.X, y)
                y -= fitted.LineHeight
            }
        }
    }

    return page, nil
}
```

## 6. サンプルコード

### 6.1. 基本的な翻訳

```go
package main

import (
    "log"
    "github.com/ryomak/gopdf"
)

func main() {
    // 簡単な翻訳関数
    translator := gopdf.TranslateFunc(func(text string) (string, error) {
        // 実際のアプリケーションでは、Google Translate API等を使用
        translations := map[string]string{
            "Technical Report": "技術レポート",
            "Introduction":     "はじめに",
            "This document...": "このドキュメントは...",
        }
        if translated, ok := translations[text]; ok {
            return translated, nil
        }
        return text, nil
    })

    // 日本語フォントを読み込み
    jpFont, err := gopdf.LoadTTF("NotoSansJP-Regular.ttf")
    if err != nil {
        log.Fatal(err)
    }

    // 翻訳オプション
    opts := gopdf.PDFTranslatorOptions{
        Translator: translator,
        TargetFont: jpFont,
        FittingOptions: gopdf.FitTextOptions{
            MaxFontSize: 20,
            MinFontSize: 6,
            LineSpacing: 1.2,
            Padding:     2,
            AllowShrink: true,
            AllowGrow:   false,
            Alignment:   gopdf.AlignLeft,
        },
        KeepImages:  true,
        KeepLayout:  true,
    }

    // PDF翻訳
    err = gopdf.TranslatePDF("english.pdf", "japanese.pdf", opts)
    if err != nil {
        log.Fatal(err)
    }

    log.Println("Translation completed!")
}
```

### 6.2. レイアウト解析のみ

```go
package main

import (
    "fmt"
    "log"
    "github.com/ryomak/gopdf"
)

func main() {
    reader, err := gopdf.Open("document.pdf")
    if err != nil {
        log.Fatal(err)
    }
    defer reader.Close()

    // ページレイアウトを抽出
    layout, err := reader.ExtractPageLayout(0)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Page size: %.1f x %.1f\n", layout.Width, layout.Height)
    fmt.Printf("Text blocks: %d\n", len(layout.TextBlocks))
    fmt.Printf("Images: %d\n", len(layout.Images))

    // テキストブロックの情報
    for i, block := range layout.TextBlocks {
        fmt.Printf("\nText Block %d:\n", i+1)
        fmt.Printf("  Text: %s\n", block.Text)
        fmt.Printf("  Position: (%.1f, %.1f)\n", block.Bounds.X, block.Bounds.Y)
        fmt.Printf("  Size: %.1f x %.1f\n", block.Bounds.Width, block.Bounds.Height)
        fmt.Printf("  Font: %s, Size: %.1f\n", block.Font, block.FontSize)
    }

    // 画像の情報
    for i, img := range layout.Images {
        fmt.Printf("\nImage %d:\n", i+1)
        fmt.Printf("  Position: (%.1f, %.1f)\n", img.X, img.Y)
        fmt.Printf("  Size: %.1f x %.1f\n", img.Width, img.Height)
        fmt.Printf("  Original: %d x %d\n", img.ImageInfo.Width, img.ImageInfo.Height)
    }
}
```

## 7. テスト計画

### 7.1. ユニットテスト

```go
// layout_test.go
func TestExtractPageLayout(t *testing.T) {
    // Writerで画像＋テキストのPDFを生成
    // ExtractPageLayoutで抽出
    // TextBlocks、Imagesの数を検証
}

func TestGroupTextElements(t *testing.T) {
    // 近接するTextElementsが同じブロックになることを検証
}

func TestFitText(t *testing.T) {
    // 各種サイズの矩形でテキストがフィッティングされることを検証
}

func TestTranslatePDF(t *testing.T) {
    // 簡単な翻訳関数でPDF翻訳を実行
    // 生成されたPDFが読み込めることを検証
}
```

### 7.2. 統合テスト

- 実際の英語PDFを日本語に翻訳
- 画像位置の保持を確認
- テキストフィッティングの動作確認

## 8. 注意事項

### 8.1. 制約事項

- **複雑な変形**: 回転・斜体などの複雑な変形は部分的にのみ対応
- **表構造**: 表の完全な構造認識は未対応
- **フォント**: ターゲット言語のフォントは使用側で指定が必要
- **翻訳API**: 実際の翻訳処理は使用側で実装

### 8.2. パフォーマンス

- 大きなPDF（100ページ以上）の処理には時間がかかる可能性
- 画像が多い場合、メモリ使用量が増加
- ページ単位での並列処理は将来の拡張として検討

### 8.3. 推奨事項

- **フォント選択**: 日本語はNoto Sansなどの高品質なフォントを推奨
- **テストファイル**: 実際の翻訳前に小さなPDFでテスト
- **バックアップ**: 元のPDFは必ず保持

## 9. 参考資料

- [PDF 1.7 仕様書](https://opensource.adobe.com/dc-acrobat-sdk-docs/pdfstandards/PDF32000_2008.pdf)
  - Section 8: Graphics
  - Section 8.3: Coordinate Systems
  - Section 8.4: Graphic State
- [docs/text_extraction_design.md](./text_extraction_design.md)
- [docs/structured_text_extraction_design.md](./structured_text_extraction_design.md)
- [docs/image_extraction_design.md](./image_extraction_design.md)
