# テキスト編集API 設計書

## 1. 概要

PDFから抽出したテキストブロックを編集し、新しいPDFを生成する機能を提供する。

### 1.1. 目的

- 既存PDFのテキストを編集可能にする
- 元のレイアウト、フォント、座標を維持しながらテキストのみを変更
- 簡単なAPIで編集から再生成までを実現

### 1.2. ユースケース

```go
// 1. PDFを開く
reader, _ := gopdf.Open("input.pdf")
defer reader.Close()

// 2. ページレイアウトを取得
layout, _ := reader.ExtractPageLayout(0)

// 3. テキストを編集
for i := range layout.TextBlocks {
    if strings.Contains(layout.TextBlocks[i].Text, "Old Text") {
        layout.TextBlocks[i].Text = strings.ReplaceAll(
            layout.TextBlocks[i].Text,
            "Old Text",
            "New Text",
        )
    }
}

// 4. 新しいPDFを生成
doc := gopdf.NewFromLayout(layout)
doc.WriteTo(file)
```

## 2. 設計方針

### 2.1. 基本方針

1. **既存のPageLayout構造を活用**
   - 新しいデータ構造を作らず、既存のTextBlock, ImageBlockを使用
   - TextBlock.Textを編集可能にする

2. **座標とフォント情報を保持**
   - TextBlock.Rect（位置とサイズ）
   - TextBlock.Font, FontSize
   - 元のレイアウトを可能な限り維持

3. **シンプルなAPI**
   - `NewFromLayout(layout *PageLayout) *Document`
   - レイアウトから直接PDFを生成

### 2.2. 制約

**編集可能なもの:**
- TextBlock.Text（テキスト内容）

**編集しないもの（元の値を使用）:**
- 座標（X, Y）
- フォント名、サイズ
- 色
- 画像

**注意点:**
- テキストが長くなった場合、元の境界を超える可能性
- フォントが変わると文字幅が変わる
- 複雑なレイアウト（回転、変換）は未対応

## 3. API設計

### 3.1. 新しい関数

```go
// NewFromLayout は PageLayout から新しいPDFドキュメントを作成
func NewFromLayout(layout *PageLayout, opts ...LayoutToDocumentOption) *Document

// LayoutToDocumentOption はレイアウトからドキュメント生成時のオプション
type LayoutToDocumentOption func(*layoutToDocumentOptions)

// デフォルトオプション
type layoutToDocumentOptions struct {
    // テキストが境界を超えた場合の動作
    OverflowBehavior OverflowBehavior
    // フォールバックフォント
    FallbackFont Font
}

// OverflowBehavior はテキストがブロック境界を超えた場合の動作
type OverflowBehavior int

const (
    // OverflowClip はテキストをクリップ（デフォルト）
    OverflowClip OverflowBehavior = iota
    // OverflowShrink はフォントサイズを縮小して収める
    OverflowShrink
    // OverflowIgnore は境界を無視して全て描画
    OverflowIgnore
)
```

### 3.2. オプション関数

```go
// WithOverflowBehavior はテキストオーバーフロー時の動作を設定
func WithOverflowBehavior(behavior OverflowBehavior) LayoutToDocumentOption {
    return func(opts *layoutToDocumentOptions) {
        opts.OverflowBehavior = behavior
    }
}

// WithFallbackFont はフォールバックフォントを設定
func WithFallbackFont(font Font) LayoutToDocumentOption {
    return func(opts *layoutToDocumentOptions) {
        opts.FallbackFont = font
    }
}
```

## 4. 実装計画

### 4.1. Phase 1: NewFromLayout の基本実装

```go
// NewFromLayout は PageLayout から新しいPDFドキュメントを作成
func NewFromLayout(layout *PageLayout, opts ...LayoutToDocumentOption) *Document {
    // オプションを適用
    options := &layoutToDocumentOptions{
        OverflowBehavior: OverflowClip,
        FallbackFont:     FontHelvetica,
    }
    for _, opt := range opts {
        opt(options)
    }

    // 新しいドキュメントを作成
    doc := New()

    // ページを追加
    orientation := Portrait
    if layout.Width > layout.Height {
        orientation = Landscape
    }
    pageSize := PageSize{Width: layout.Width, Height: layout.Height}
    page := doc.AddPage(pageSize, orientation)

    // コンテンツブロックを描画（上から下の順）
    blocks := layout.SortedContentBlocks()
    for _, block := range blocks {
        switch block.Type() {
        case ContentBlockTypeText:
            tb := block.(TextBlock)
            drawTextBlock(page, tb, options)

        case ContentBlockTypeImage:
            ib := block.(ImageBlock)
            drawImageBlock(page, ib)
        }
    }

    return doc
}
```

### 4.2. Phase 2: テキストブロック描画

```go
// drawTextBlock はテキストブロックをページに描画
func drawTextBlock(page *Page, block TextBlock, opts *layoutToDocumentOptions) error {
    // フォントを設定
    fontName := block.Font
    if fontName == "" {
        fontName = string(opts.FallbackFont)
    }

    // 標準フォントの場合
    if isStandardFont(fontName) {
        page.SetFont(mapToStandardFont(fontName), block.FontSize)
    } else {
        // TTFフォントの場合（簡易的には標準フォントにフォールバック）
        page.SetFont(opts.FallbackFont, block.FontSize)
    }

    // テキストを描画
    // 複数行の場合は改行を考慮
    lines := strings.Split(block.Text, "\n")
    lineHeight := block.FontSize * 1.2

    y := block.Rect.Y + block.Rect.Height // 上端から開始

    for _, line := range lines {
        if line == "" {
            y -= lineHeight
            continue
        }

        // テキスト幅をチェック
        textWidth := estimateTextWidth(line, block.FontSize, fontName)

        // オーバーフロー処理
        if textWidth > block.Rect.Width {
            switch opts.OverflowBehavior {
            case OverflowShrink:
                // フォントサイズを縮小
                adjustedSize := block.FontSize * (block.Rect.Width / textWidth)
                page.SetFont(mapToStandardFont(fontName), adjustedSize)
            case OverflowClip:
                // クリップ（何もしない、はみ出す）
            case OverflowIgnore:
                // 無視して描画
            }
        }

        // テキストを描画
        page.DrawText(line, block.Rect.X, y)
        y -= lineHeight

        // 境界を超えた場合は停止
        if y < block.Rect.Y && opts.OverflowBehavior == OverflowClip {
            break
        }
    }

    return nil
}

// isStandardFont は標準フォントかどうかを判定
func isStandardFont(fontName string) bool {
    standardFonts := []string{
        "Helvetica", "Helvetica-Bold", "Helvetica-Oblique", "Helvetica-BoldOblique",
        "Times-Roman", "Times-Bold", "Times-Italic", "Times-BoldItalic",
        "Courier", "Courier-Bold", "Courier-Oblique", "Courier-BoldOblique",
        "Symbol", "ZapfDingbats",
    }
    for _, sf := range standardFonts {
        if strings.Contains(fontName, sf) || strings.Contains(sf, fontName) {
            return true
        }
    }
    return false
}

// mapToStandardFont は内部フォント名を標準フォントにマップ
func mapToStandardFont(fontName string) Font {
    if strings.Contains(fontName, "Bold") {
        return FontHelveticaBold
    }
    if strings.Contains(fontName, "Italic") || strings.Contains(fontName, "Oblique") {
        return FontHelvetica // Italicは簡易的には通常フォント
    }
    if strings.Contains(fontName, "Times") {
        return FontTimesRoman
    }
    if strings.Contains(fontName, "Courier") {
        return FontCourier
    }
    return FontHelvetica // デフォルト
}
```

### 4.3. Phase 3: 画像ブロック描画

```go
// drawImageBlock は画像ブロックをページに描画
func drawImageBlock(page *Page, block ImageBlock) error {
    // 画像データからImageを作成
    var img *Image
    var err error

    switch block.Format {
    case ImageFormatJPEG:
        img, err = LoadJPEGData(block.Data)
    case ImageFormatPNG:
        img, err = LoadPNGData(block.Data)
    default:
        // 未対応フォーマットはスキップ
        return nil
    }

    if err != nil {
        return err
    }

    // 画像を描画
    page.DrawImage(img, block.X, block.Y, block.PlacedWidth, block.PlacedHeight)

    return nil
}
```

## 5. 使用例

### 5.1. 基本的なテキスト編集

```go
package main

import (
    "fmt"
    "os"
    "strings"

    "github.com/ryomak/gopdf"
)

func main() {
    // PDFを開く
    reader, err := gopdf.Open("input.pdf")
    if err != nil {
        panic(err)
    }
    defer reader.Close()

    // ページレイアウトを取得
    layout, err := reader.ExtractPageLayout(0)
    if err != nil {
        panic(err)
    }

    // テキストを編集
    for i := range layout.TextBlocks {
        // 特定のテキストを置換
        layout.TextBlocks[i].Text = strings.ReplaceAll(
            layout.TextBlocks[i].Text,
            "October 8, 2025",
            "October 10, 2025",
        )
    }

    // 新しいPDFを生成
    doc := gopdf.NewFromLayout(layout)

    // ファイルに保存
    file, err := os.Create("output.pdf")
    if err != nil {
        panic(err)
    }
    defer file.Close()

    if err := doc.WriteTo(file); err != nil {
        panic(err)
    }

    fmt.Println("PDF edited successfully!")
}
```

### 5.2. オーバーフロー処理

```go
// テキストが長くなった場合、フォントサイズを縮小
doc := gopdf.NewFromLayout(layout,
    gopdf.WithOverflowBehavior(gopdf.OverflowShrink),
)
```

### 5.3. フォールバックフォントの指定

```go
// 日本語フォントをフォールバックに使用
doc := gopdf.NewFromLayout(layout,
    gopdf.WithFallbackFont(gopdf.FontHelvetica), // TTFフォントも指定可能に拡張予定
)
```

## 6. テスト計画

### 6.1. ユニットテスト

```go
func TestNewFromLayout(t *testing.T) {
    // 簡単なレイアウトを作成
    layout := &gopdf.PageLayout{
        Width:  595,
        Height: 842,
        TextBlocks: []gopdf.TextBlock{
            {
                Text: "Hello World",
                Rect: gopdf.Rectangle{X: 50, Y: 800, Width: 100, Height: 20},
                Font: "Helvetica",
                FontSize: 12,
            },
        },
    }

    // PDFを生成
    doc := gopdf.NewFromLayout(layout)

    // ドキュメントが生成されることを確認
    if doc == nil {
        t.Fatal("Document should not be nil")
    }
}
```

### 6.2. 統合テスト

```go
func TestEditAndRegenerate(t *testing.T) {
    // テストPDFを作成
    originalDoc := gopdf.New()
    page := originalDoc.AddPage(gopdf.PageSizeA4, gopdf.Portrait)
    page.SetFont(gopdf.FontHelvetica, 12)
    page.DrawText("Original Text", 50, 800)

    // 一時ファイルに保存
    tmpFile, _ := os.CreateTemp("", "test-*.pdf")
    originalDoc.WriteTo(tmpFile)
    tmpFile.Close()

    // PDFを読み込んで編集
    reader, _ := gopdf.Open(tmpFile.Name())
    defer reader.Close()

    layout, _ := reader.ExtractPageLayout(0)
    layout.TextBlocks[0].Text = "Edited Text"

    // 新しいPDFを生成
    editedDoc := gopdf.NewFromLayout(layout)

    // 再度読み込んでテキストを確認
    editedFile, _ := os.CreateTemp("", "edited-*.pdf")
    editedDoc.WriteTo(editedFile)
    editedFile.Close()

    reader2, _ := gopdf.Open(editedFile.Name())
    defer reader2.Close()

    text, _ := reader2.ExtractPageText(0)

    if !strings.Contains(text, "Edited Text") {
        t.Errorf("Expected 'Edited Text', got %s", text)
    }
}
```

## 7. 制限事項と将来の拡張

### 7.1. 現在の制限

1. **フォント**
   - 標準フォントのみサポート
   - TTFフォントは今後対応予定

2. **レイアウト**
   - 単純なレイアウトのみ
   - 回転、特殊な変換は未対応

3. **テキストの長さ**
   - 元の境界を超える場合の処理は限定的

### 7.2. 将来の拡張

1. **TTFフォント対応**
   - 元PDFで使用されていたTTFフォントを保持
   - 新しいTTFフォントの指定

2. **より高度なレイアウト調整**
   - 自動リフロー
   - 行の自動折り返し

3. **インタラクティブな編集**
   - GUI エディタとの統合

## 8. 参考資料

- docs/text_block_grouping_design.md
- docs/unified_content_grouping_design.md
- docs/extraction_issues_analysis.md
