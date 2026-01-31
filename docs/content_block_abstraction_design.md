# コンテンツブロック抽象化とPDF再構成機能 設計書

## 1. 概要

PDFから抽出したコンテンツ（テキスト、画像）を統一的に扱い、それらを使って新しいPDFを生成する機能を提供する。

### 1.1. 目的

- テキストブロックと画像ブロックを統一的に扱うインターフェースの提供
- PDFから抽出したレイアウト情報を使った新規PDF生成
- 既存PDFの再構成・変換の基盤機能

### 1.2. ユースケース

```go
// 1. 既存PDFからレイアウトを抽出
reader, _ := gopdf.Open("input.pdf")
layout, _ := reader.ExtractPageLayout(0)

// 2. コンテンツを操作（テキスト翻訳など）
for i := range layout.TextBlocks {
    layout.TextBlocks[i].Text = translate(layout.TextBlocks[i].Text)
}

// 3. 新しいPDFとして出力
doc := gopdf.New()
page := doc.AddPage(gopdf.CustomSize(layout.Width, layout.Height), gopdf.Portrait)
page.RenderLayout(layout)
doc.WriteTo(outputFile)
```

### 1.3. スコープ

**実装する機能:**
- ✅ ContentBlockインターフェースの定義
- ✅ TextBlockとImageBlockのContentBlock実装
- ✅ PageLayout出力機能
- ✅ 位置情報を保持したレンダリング

**実装しない機能:**
- フォント自動置き換え（ユーザーが明示的に指定）
- 複雑なテキストフロー調整
- 自動リサイズ・リフロー

## 2. 現状の課題

### 2.1. 既存機能

**入力側（抽出）:**
```go
// 既に実装済み
layout, _ := reader.ExtractPageLayout(0)
// => PageLayout{TextBlocks, Images}
```

**出力側（描画）:**
```go
// 個別の描画APIのみ
page.DrawText("Hello", 100, 100)
page.DrawImage(img, 100, 200, 50, 50)
```

### 2.2. 問題点

- TextBlockとImageBlockを統一的に扱えない
- 抽出したレイアウトをそのまま出力できない
- 位置情報を保持したままPDF再構成が困難

## 3. 設計

### 3.1. データ構造

#### 3.1.1. ContentBlockインターフェース

```go
package gopdf

// ContentBlock はページ内のコンテンツブロックを表す統一インターフェース
type ContentBlock interface {
    // Bounds はブロックの境界矩形を返す
    Bounds() Rectangle

    // Type はブロックの種類を返す
    Type() ContentBlockType

    // Position はブロックの配置位置を返す（左下座標）
    Position() (x, y float64)
}

// ContentBlockType はコンテンツブロックの種類
type ContentBlockType string

const (
    ContentBlockTypeText  ContentBlockType = "text"
    ContentBlockTypeImage ContentBlockType = "image"
)
```

#### 3.1.2. TextBlockの拡張

```go
// TextBlock はContentBlockインターフェースを実装
func (tb TextBlock) Bounds() Rectangle {
    return tb.Bounds
}

func (tb TextBlock) Type() ContentBlockType {
    return ContentBlockTypeText
}

func (tb TextBlock) Position() (float64, float64) {
    return tb.Bounds.X, tb.Bounds.Y
}
```

#### 3.1.3. ImageBlockの拡張

```go
// ImageBlock はContentBlockインターフェースを実装
func (ib ImageBlock) Bounds() Rectangle {
    return Rectangle{
        X:      ib.X,
        Y:      ib.Y,
        Width:  ib.PlacedWidth,
        Height: ib.PlacedHeight,
    }
}

func (ib ImageBlock) Type() ContentBlockType {
    return ContentBlockTypeImage
}

func (ib ImageBlock) Position() (float64, float64) {
    return ib.X, ib.Y
}
```

#### 3.1.4. PageLayoutの拡張

```go
// PageLayout にコンテンツブロックを統一的に取得するメソッドを追加
func (pl *PageLayout) ContentBlocks() []ContentBlock {
    var blocks []ContentBlock

    // TextBlocksを追加
    for _, tb := range pl.TextBlocks {
        blocks = append(blocks, tb)
    }

    // ImageBlocksを追加
    for _, ib := range pl.Images {
        blocks = append(blocks, ib)
    }

    // Y座標でソート（上から下）
    sort.Slice(blocks, func(i, j int) bool {
        _, yi := blocks[i].Position()
        _, yj := blocks[j].Position()
        return yi > yj
    })

    return blocks
}

// SortedContentBlocks はコンテンツブロックをソート順で返す
// ソート順: Y座標（上から下）、同じY座標ならX座標（左から右）
func (pl *PageLayout) SortedContentBlocks() []ContentBlock {
    blocks := pl.ContentBlocks()

    sort.Slice(blocks, func(i, j int) bool {
        xi, yi := blocks[i].Position()
        xj, yj := blocks[j].Position()

        // Y座標で比較（上から下）
        if math.Abs(yi-yj) > 1.0 {
            return yi > yj
        }

        // X座標で比較（左から右）
        return xi < xj
    })

    return blocks
}
```

### 3.2. 出力機能

#### 3.2.1. Page.RenderLayout API

```go
package gopdf

// RenderLayout はPageLayoutをページにレンダリングする
// レイアウト内のすべてのコンテンツブロック（テキスト・画像）を描画する
func (p *Page) RenderLayout(layout *PageLayout) error {
    // コンテンツブロックをソート順で取得
    blocks := layout.SortedContentBlocks()

    for _, block := range blocks {
        if err := p.RenderContentBlock(block); err != nil {
            return fmt.Errorf("failed to render block: %w", err)
        }
    }

    return nil
}

// RenderContentBlock は単一のコンテンツブロックをレンダリング
func (p *Page) RenderContentBlock(block ContentBlock) error {
    switch block.Type() {
    case ContentBlockTypeText:
        return p.renderTextBlock(block.(TextBlock))
    case ContentBlockTypeImage:
        return p.renderImageBlock(block.(ImageBlock))
    default:
        return fmt.Errorf("unsupported content block type: %s", block.Type())
    }
}

// renderTextBlock はTextBlockをレンダリング
func (p *Page) renderTextBlock(block TextBlock) error {
    // フォントが設定されていない場合のデフォルト処理
    if p.currentFont == nil && p.currentTTFFont == nil {
        // デフォルトフォントを設定
        if err := p.SetFont(font.Helvetica, block.FontSize); err != nil {
            return err
        }
    }

    // テキストを描画
    x, y := block.Position()
    return p.DrawText(block.Text, x, y)
}

// renderImageBlock はImageBlockをレンダリング
func (p *Page) renderImageBlock(block ImageBlock) error {
    // 画像データからimage.Imageを生成
    img, err := block.ToImage()
    if err != nil {
        return fmt.Errorf("failed to convert image: %w", err)
    }

    // 画像を描画
    return p.DrawImage(img, block.X, block.Y, block.PlacedWidth, block.PlacedHeight)
}
```

#### 3.2.2. 高度なレンダリングオプション

```go
// RenderLayoutOptions はレンダリング時のオプション
type RenderLayoutOptions struct {
    // PreserveFont はフォントを保持するか（デフォルト: false）
    PreserveFont bool

    // PreserveColor は色を保持するか（デフォルト: false）
    PreserveColor bool

    // DefaultFont はデフォルトフォント
    DefaultFont StandardFont

    // DefaultFontSize はデフォルトフォントサイズ
    DefaultFontSize float64

    // ScaleFactor は拡大縮小率（デフォルト: 1.0）
    ScaleFactor float64

    // OffsetX, OffsetY はオフセット（デフォルト: 0）
    OffsetX float64
    OffsetY float64
}

// RenderLayoutWithOptions はオプション付きでレイアウトをレンダリング
func (p *Page) RenderLayoutWithOptions(layout *PageLayout, opts RenderLayoutOptions) error {
    // デフォルト値の設定
    if opts.ScaleFactor == 0 {
        opts.ScaleFactor = 1.0
    }
    if opts.DefaultFontSize == 0 {
        opts.DefaultFontSize = 12
    }

    blocks := layout.SortedContentBlocks()

    for _, block := range blocks {
        // 位置を調整
        x, y := block.Position()
        x = x*opts.ScaleFactor + opts.OffsetX
        y = y*opts.ScaleFactor + opts.OffsetY

        switch block.Type() {
        case ContentBlockTypeText:
            tb := block.(TextBlock)

            // フォント設定
            fontSize := tb.FontSize * opts.ScaleFactor
            if opts.DefaultFont != "" {
                p.SetFont(opts.DefaultFont, fontSize)
            } else if p.currentFont == nil && p.currentTTFFont == nil {
                p.SetFont(font.Helvetica, fontSize)
            }

            // テキスト描画
            if err := p.DrawText(tb.Text, x, y); err != nil {
                return err
            }

        case ContentBlockTypeImage:
            ib := block.(ImageBlock)
            img, err := ib.ToImage()
            if err != nil {
                return fmt.Errorf("failed to convert image: %w", err)
            }

            w := ib.PlacedWidth * opts.ScaleFactor
            h := ib.PlacedHeight * opts.ScaleFactor

            if err := p.DrawImage(img, x, y, w, h); err != nil {
                return err
            }
        }
    }

    return nil
}
```

### 3.3. 画像データ変換機能

ImageBlockからimage.Imageを取得する機能を追加：

```go
package gopdf

import (
    "bytes"
    "compress/zlib"
    "fmt"
    "image"
    "image/color"
    "image/jpeg"
)

// ToImage は画像データをimage.Imageに変換する
func (ib *ImageBlock) ToImage() (image.Image, error) {
    // ImageInfoのToImageを利用
    return ib.ImageInfo.ToImage()
}

// ToImage は画像データをimage.Imageに変換する
func (ii *ImageInfo) ToImage() (image.Image, error) {
    switch ii.Format {
    case ImageFormatJPEG:
        return jpeg.Decode(bytes.NewReader(ii.Data))

    case ImageFormatPNG:
        // FlateDecodeの画像をデコード
        return ii.decodeFlateImage()

    default:
        return nil, fmt.Errorf("unsupported image format: %s", ii.Format)
    }
}

// decodeFlateImage はFlateDecode画像をデコードする
func (ii *ImageInfo) decodeFlateImage() (image.Image, error) {
    // zlibで展開
    reader, err := zlib.NewReader(bytes.NewReader(ii.Data))
    if err != nil {
        return nil, fmt.Errorf("failed to create zlib reader: %w", err)
    }
    defer reader.Close()

    // 展開されたデータを読み込み
    var buf bytes.Buffer
    if _, err := buf.ReadFrom(reader); err != nil {
        return nil, fmt.Errorf("failed to read image data: %w", err)
    }

    rawData := buf.Bytes()

    // 色空間に応じて画像を作成
    switch ii.ColorSpace {
    case "DeviceRGB":
        return ii.createRGBImage(rawData)
    case "DeviceGray":
        return ii.createGrayImage(rawData)
    default:
        return nil, fmt.Errorf("unsupported color space: %s", ii.ColorSpace)
    }
}

// createRGBImage はRGB画像を作成
func (ii *ImageInfo) createRGBImage(data []byte) (image.Image, error) {
    img := image.NewRGBA(image.Rect(0, 0, ii.Width, ii.Height))

    bytesPerPixel := 3
    expectedSize := ii.Width * ii.Height * bytesPerPixel

    if len(data) < expectedSize {
        return nil, fmt.Errorf("insufficient image data: got %d bytes, expected %d", len(data), expectedSize)
    }

    for y := 0; y < ii.Height; y++ {
        for x := 0; x < ii.Width; x++ {
            offset := (y*ii.Width + x) * bytesPerPixel
            r := data[offset]
            g := data[offset+1]
            b := data[offset+2]
            img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
        }
    }

    return img, nil
}

// createGrayImage はグレースケール画像を作成
func (ii *ImageInfo) createGrayImage(data []byte) (image.Image, error) {
    img := image.NewGray(image.Rect(0, 0, ii.Width, ii.Height))

    expectedSize := ii.Width * ii.Height

    if len(data) < expectedSize {
        return nil, fmt.Errorf("insufficient image data: got %d bytes, expected %d", len(data), expectedSize)
    }

    for y := 0; y < ii.Height; y++ {
        for x := 0; x < ii.Width; x++ {
            offset := y*ii.Width + x
            gray := data[offset]
            img.Set(x, y, color.Gray{Y: gray})
        }
    }

    return img, nil
}
```

## 4. 実装計画

### 4.1. Phase 1: インターフェースとメソッド定義

1. `ContentBlock`インターフェースを定義
2. `TextBlock`に`Bounds()`, `Type()`, `Position()`メソッドを追加
3. `ImageBlock`に`Bounds()`, `Type()`, `Position()`メソッドを追加
4. `PageLayout.ContentBlocks()`メソッドを実装

### 4.2. Phase 2: 画像変換機能

1. `ImageInfo.ToImage()`を実装
2. `ImageInfo.decodeFlateImage()`を実装
3. `ImageBlock.ToImage()`を実装

### 4.3. Phase 3: レンダリング機能

1. `Page.RenderLayout()`を実装
2. `Page.RenderContentBlock()`を実装
3. `Page.renderTextBlock()`を実装
4. `Page.renderImageBlock()`を実装

### 4.4. Phase 4: テストと検証

1. ユニットテストの作成
2. 統合テストの作成
3. サンプルプログラムの作成

## 5. テスト計画

### 5.1. ユニットテスト

```go
// TestContentBlockInterface はContentBlockインターフェースのテスト
func TestContentBlockInterface(t *testing.T) {
    // TextBlockのテスト
    tb := TextBlock{
        Bounds: Rectangle{X: 100, Y: 200, Width: 300, Height: 50},
    }

    var block ContentBlock = tb
    assert.Equal(t, ContentBlockTypeText, block.Type())

    x, y := block.Position()
    assert.Equal(t, 100.0, x)
    assert.Equal(t, 200.0, y)
}

// TestPageLayoutContentBlocks はContentBlocks()のテスト
func TestPageLayoutContentBlocks(t *testing.T) {
    layout := &PageLayout{
        TextBlocks: []TextBlock{
            {Bounds: Rectangle{X: 100, Y: 700}},
        },
        Images: []ImageBlock{
            {X: 100, Y: 500},
        },
    }

    blocks := layout.ContentBlocks()
    assert.Equal(t, 2, len(blocks))

    // ソート順を確認（Y座標が大きい方が先）
    _, y1 := blocks[0].Position()
    _, y2 := blocks[1].Position()
    assert.True(t, y1 > y2)
}
```

### 5.2. 統合テスト

```go
// TestRenderLayout はレイアウトレンダリングの統合テスト
func TestRenderLayout(t *testing.T) {
    // 1. 元のPDFを作成
    doc1 := gopdf.New()
    page1 := doc1.AddPage(gopdf.A4, gopdf.Portrait)
    page1.SetFont(font.Helvetica, 12)
    page1.DrawText("Hello World", 100, 700)

    var buf1 bytes.Buffer
    doc1.WriteTo(&buf1)

    // 2. PDFを読み込んでレイアウトを抽出
    reader, _ := gopdf.OpenReader(bytes.NewReader(buf1.Bytes()))
    layout, _ := reader.ExtractPageLayout(0)

    // 3. 新しいPDFにレンダリング
    doc2 := gopdf.New()
    page2 := doc2.AddPage(gopdf.A4, gopdf.Portrait)
    page2.SetFont(font.Helvetica, 12)
    page2.RenderLayout(layout)

    var buf2 bytes.Buffer
    doc2.WriteTo(&buf2)

    // 4. 読み込んで検証
    reader2, _ := gopdf.OpenReader(bytes.NewReader(buf2.Bytes()))
    text, _ := reader2.ExtractPageText(0)

    assert.Contains(t, text, "Hello World")
}
```

### 5.3. サンプル

`examples/16_pdf_reconstruction/main.go`:

```go
package main

import (
    "log"
    "os"

    "github.com/ryomak/gopdf"
)

func main() {
    // 1. 既存PDFを読み込み
    reader, err := gopdf.Open("input.pdf")
    if err != nil {
        log.Fatal(err)
    }
    defer reader.Close()

    // 2. レイアウトを抽出
    layout, err := reader.ExtractPageLayout(0)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Page layout: %d text blocks, %d images",
        len(layout.TextBlocks), len(layout.Images))

    // 3. コンテンツブロックを表示
    blocks := layout.SortedContentBlocks()
    for i, block := range blocks {
        x, y := block.Position()
        log.Printf("Block %d: type=%s, position=(%.1f, %.1f)",
            i, block.Type(), x, y)
    }

    // 4. 新しいPDFを作成
    doc := gopdf.New()
    page := doc.AddPage(
        gopdf.CustomSize(layout.Width, layout.Height),
        gopdf.Portrait,
    )

    // デフォルトフォントを設定
    page.SetFont(gopdf.FontHelvetica, 12)

    // 5. レイアウトをレンダリング
    if err := page.RenderLayout(layout); err != nil {
        log.Fatal(err)
    }

    // 6. ファイルに保存
    file, err := os.Create("output.pdf")
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    if err := doc.WriteTo(file); err != nil {
        log.Fatal(err)
    }

    log.Println("PDF reconstruction completed: output.pdf")
}
```

## 6. 注意事項

### 6.1. フォント

- 元のPDFのフォント情報は完全には保持されない
- ユーザーが明示的にフォントを設定する必要がある
- TTFフォントの場合、同じフォントファイルが必要

### 6.2. レイアウトの精度

- 複雑なレイアウト（多段組み、表など）は完全には再現できない
- テキストの配置は元の位置情報に基づく
- フォントサイズや幅が異なると、はみ出す可能性がある

### 6.3. 画像

- JPEG画像はそのまま利用可能
- FlateDecode画像は再エンコードが必要
- 一部の色空間（CMYK、ICCベース）は未対応

### 6.4. パフォーマンス

- 大量のコンテンツブロックがある場合、処理時間がかかる
- 画像の変換・再エンコードは重い処理

## 7. テキストはみ出し対策

翻訳後のテキストが元の領域からはみ出る問題への対策：

### 7.1. 既存の対策（実装済み）

#### FitTextによる自動調整

`text_fitting.go`の`FitText`関数を使用して、テキストを領域内に収める：

```go
// 既にtranslator.goで使用されている例
fitted, err := FitText(block.Text, block.Rect, fontName, FitTextOptions{
    MaxFontSize: 24.0,
    MinFontSize: 6.0,
    LineSpacing: 1.2,
    AllowShrink: true,  // フォントサイズを縮小
    AllowGrow:   false, // フォントサイズを拡大しない
    Alignment:   AlignLeft,
})
```

**動作:**
- 2分探索でフォントサイズを自動調整
- テキストを自動改行
- 領域に収まる最大のフォントサイズを見つける

**メリット:**
- 既に実装済み
- レイアウトを崩さない
- 複数行対応

**デメリット:**
- フォントが小さくなりすぎる可能性
- 元のフォントサイズから大きく変わる場合がある

### 7.2. 新規対策（将来実装）

#### オプション1: AutoLayoutモード

はみ出したブロックを自動的に下にずらす：

```go
type AutoLayoutOptions struct {
    EnableAutoShift bool    // 自動位置調整を有効化
    MinSpacing      float64 // ブロック間の最小間隔
    PageMargin      float64 // ページ下端からのマージン
}

// PageLayoutを自動調整
func (pl *PageLayout) AutoAdjustLayout(opts AutoLayoutOptions) error {
    // 上から順に処理
    blocks := pl.SortedContentBlocks()

    var currentY float64 = pl.Height - opts.PageMargin

    for i, block := range blocks {
        bounds := block.GetBounds()

        // 前のブロックと重なる場合
        if bounds.Y + bounds.Height > currentY {
            // 位置を調整
            newY := currentY - bounds.Height

            // TextBlockの場合
            if tb, ok := block.(TextBlock); ok {
                tb.Rect.Y = newY
                pl.TextBlocks[i] = tb
            }
            // ImageBlockの場合
            if ib, ok := block.(ImageBlock); ok {
                ib.Y = newY
                pl.Images[i] = ib
            }
        }

        // 次のブロックの配置位置を更新
        currentY = block.GetBounds().Y - opts.MinSpacing
    }

    return nil
}
```

**メリット:**
- レイアウトの重なりを防ぐ
- テキストが読みやすくなる

**デメリット:**
- 元のレイアウトから大きく変わる
- ページをまたぐ処理が必要

#### オプション2: 手動調整API

ユーザーがブロック位置を手動で調整：

```go
// 特定のブロックを移動
func (pl *PageLayout) MoveTextBlock(index int, offsetX, offsetY float64) {
    if index < len(pl.TextBlocks) {
        pl.TextBlocks[index].Rect.X += offsetX
        pl.TextBlocks[index].Rect.Y += offsetY
    }
}

// ブロックをリサイズ
func (pl *PageLayout) ResizeTextBlock(index int, newWidth, newHeight float64) {
    if index < len(pl.TextBlocks) {
        pl.TextBlocks[index].Rect.Width = newWidth
        pl.TextBlocks[index].Rect.Height = newHeight
    }
}
```

**使用例:**
```go
layout, _ := reader.ExtractPageLayout(0)

// テキストを翻訳
for i := range layout.TextBlocks {
    layout.TextBlocks[i].Text = translate(layout.TextBlocks[i].Text)

    // 翻訳後のテキストが長い場合、領域を拡大
    if len(layout.TextBlocks[i].Text) > originalLength * 1.5 {
        layout.ResizeTextBlock(i,
            layout.TextBlocks[i].Rect.Width,
            layout.TextBlocks[i].Rect.Height * 1.5)
    }
}

// 自動レイアウト調整
layout.AutoAdjustLayout(AutoLayoutOptions{
    EnableAutoShift: true,
    MinSpacing:      10.0,
    PageMargin:      20.0,
})

// レンダリング
page.RenderLayout(layout)
```

#### オプション3: 複数ページへの分割

テキストが長すぎる場合、複数ページに分割：

```go
type PageSplitOptions struct {
    MaxBlocksPerPage int     // 1ページあたりの最大ブロック数
    PageHeight       float64 // ページ高さ
}

// レイアウトを複数ページに分割
func (pl *PageLayout) SplitIntoPages(opts PageSplitOptions) ([]*PageLayout, error) {
    var pages []*PageLayout
    currentPage := &PageLayout{
        Width:  pl.Width,
        Height: opts.PageHeight,
    }

    currentY := opts.PageHeight

    for _, block := range pl.TextBlocks {
        blockHeight := block.Rect.Height

        // 現在のページに収まらない場合
        if currentY - blockHeight < 0 {
            // 新しいページを作成
            pages = append(pages, currentPage)
            currentPage = &PageLayout{
                Width:  pl.Width,
                Height: opts.PageHeight,
            }
            currentY = opts.PageHeight
        }

        // ブロックを追加
        block.Rect.Y = currentY - blockHeight
        currentPage.TextBlocks = append(currentPage.TextBlocks, block)
        currentY -= (blockHeight + 10) // 間隔
    }

    // 最後のページを追加
    if len(currentPage.TextBlocks) > 0 {
        pages = append(pages, currentPage)
    }

    return pages, nil
}
```

### 7.3. 推奨される使い方

```go
// 1. 基本: FitTextで自動調整（既存機能）
opts := DefaultPDFTranslatorOptions(targetFont, fontName)
opts.FittingOptions.AllowShrink = true
opts.FittingOptions.MinFontSize = 8.0 // 最小サイズを設定
TranslatePDF("input.pdf", "output.pdf", opts)

// 2. 高度: 手動でレイアウト調整（将来実装）
layout, _ := reader.ExtractPageLayout(0)

// 翻訳
for i := range layout.TextBlocks {
    layout.TextBlocks[i].Text = translate(layout.TextBlocks[i].Text)
}

// はみ出しチェックと調整
for i := range layout.TextBlocks {
    fitted, err := FitText(
        layout.TextBlocks[i].Text,
        layout.TextBlocks[i].Rect,
        fontName,
        opts.FittingOptions,
    )

    if err != nil {
        // 収まらない場合、領域を拡大
        estimatedHeight := EstimateTotalHeight(
            layout.TextBlocks[i].Text,
            layout.TextBlocks[i].Rect.Width,
            fontName,
            layout.TextBlocks[i].FontSize,
            1.2,
        )
        layout.ResizeTextBlock(i,
            layout.TextBlocks[i].Rect.Width,
            estimatedHeight)
    }
}

// 自動レイアウト調整でブロックをずらす
layout.AutoAdjustLayout(AutoLayoutOptions{
    EnableAutoShift: true,
    MinSpacing:      10.0,
})

// 複数ページに分割
pages, _ := layout.SplitIntoPages(PageSplitOptions{
    PageHeight: 842.0, // A4
})

// 各ページをレンダリング
for _, page := range pages {
    p := doc.AddPage(A4, Portrait)
    p.RenderLayout(page)
}
```

## 8. 拡張性

将来的な拡張：

1. **フォント自動マッピング**: 近いフォントを自動選択
2. **テキストフロー調整**: 幅に合わせて自動改行（実装済み）
3. **スタイル保持**: 太字、斜体などのスタイル情報
4. **アノテーション**: リンク、注釈の保持
5. **ベクター図形**: パスやシェイプの再構成
6. **AutoLayoutモード**: ブロックの自動位置調整（上記参照）
7. **複数ページ分割**: 長いコンテンツの自動ページ分割（上記参照）

## 8. 参考資料

- [docs/architecture.md](./architecture.md)
- [docs/structured_text_extraction_design.md](./structured_text_extraction_design.md)
- [docs/image_extraction_design.md](./image_extraction_design.md)
- [docs/text_block_grouping_design.md](./text_block_grouping_design.md)
- [PDF 1.7 仕様書](https://opensource.adobe.com/dc-acrobat-sdk-docs/pdfstandards/PDF32000_2008.pdf)
