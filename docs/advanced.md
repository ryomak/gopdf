# 高度な機能

ルビ（ふりがな）、OCRテキストレイヤー、レイアウト調整などの高度な機能を説明します。

## ルビ（ふりがな）

日本語テキストにふりがな（ルビ）を付ける機能です。

### 基本的な使い方

```go
// 日本語フォントを読み込み
jpFont, _ := gopdf.LoadSystemJapaneseFont()

doc := gopdf.New()
page := doc.AddPage(gopdf.PageSizeA4, gopdf.Portrait)
page.SetTTFFont(jpFont, 24)

// ルビ付きテキストを作成
rubyText := gopdf.NewRubyText("漢字", "かんじ")

// デフォルトスタイルでルビを描画
style := gopdf.DefaultRubyStyle()
page.DrawRuby(rubyText, 100, 750, style)
```

### RubyStyle

ルビの表示スタイルをカスタマイズできます：

```go
style := gopdf.DefaultRubyStyle()

// ルビのサイズ比率（親文字に対する比率）
style.SizeRatio = 0.5  // 50%

// 配置
style.Alignment = gopdf.RubyAlignCenter  // 中央揃え
style.Alignment = gopdf.RubyAlignLeft    // 左揃え
style.Alignment = gopdf.RubyAlignRight   // 右揃え

// 親文字との間隔
style.Offset = 2.0  // ポイント
```

### コピー時の動作

ActualText属性を使用して、コピー時のテキストを制御できます：

```go
// 漢字のみコピーされる（デフォルト）
style.CopyMode = gopdf.RubyCopyModeBase

// ひらがなのみコピーされる
style.CopyMode = gopdf.RubyCopyModeRuby

// 「漢字(かんじ)」形式でコピーされる
style.CopyMode = gopdf.RubyCopyModeBoth
```

### 複数のルビを連続描画

```go
rubyTexts := []gopdf.RubyText{
    gopdf.NewRubyText("東", "とう"),
    gopdf.NewRubyText("京", "きょう"),
    gopdf.NewRubyText("駅", "えき"),
}

page.DrawRubyTexts(rubyTexts, 100, 750, style)
```

### ActualText付きの描画

コピー時のテキストを明示的に指定：

```go
rubyText := gopdf.NewRubyText("日本語", "にほんご")
actualText := "日本語(にほんご)"

page.DrawRubyWithActualText(rubyText, 100, 750, style, actualText)
```

## OCRテキストレイヤー

画像ベースのPDFに透明なテキストレイヤーを追加し、検索・コピー可能にします。

### 基本的な使い方

```go
doc := gopdf.New()
page := doc.AddPage(gopdf.PageSizeA4, gopdf.Portrait)

// 背景画像を配置
img, _ := gopdf.LoadJPEG(imageReader)
page.DrawImage(img, 0, 0, 595, 842)

// 透明なテキストを追加
page.AddInvisibleText("Hello, World!", 50, 750, 200, 20)
```

### OCR結果からテキストレイヤーを追加

```go
// OCR結果の構造体
ocrResult := gopdf.OCRResult{
    Text: "This is the full document text",
    Words: []gopdf.OCRWord{
        {Text: "This", X: 10, Y: 10, Width: 50, Height: 20},
        {Text: "is", X: 70, Y: 10, Width: 20, Height: 20},
        {Text: "the", X: 100, Y: 10, Width: 30, Height: 20},
        // ... 他の単語
    },
}

// 画像サイズとPDFサイズを指定してテキストレイヤーに変換
imageWidth := 1920.0
imageHeight := 1080.0
pdfWidth := 595.0
pdfHeight := 842.0

textLayer := ocrResult.ToTextLayer(imageWidth, imageHeight, pdfWidth, pdfHeight)

// ページにテキストレイヤーを追加
page.AddTextLayer(textLayer)
```

### 単語単位でテキストレイヤーを追加

```go
words := []gopdf.TextLayerWord{
    {Text: "Hello", X: 50, Y: 750, Width: 80, Height: 20},
    {Text: "World", X: 140, Y: 750, Width: 80, Height: 20},
}

page.AddTextLayerWords(words)
```

## レイアウト調整

抽出したレイアウトを操作・調整する機能です。

### レイアウトの取得

```go
reader, _ := gopdf.Open("document.pdf")
layouts, _ := reader.ExtractAllLayouts()

layout := layouts[0] // 最初のページ
```

### ブロックの移動

```go
// テキストブロックを移動（dx, dy）
layout.MoveBlock(gopdf.ContentBlockTypeText, 0, 20, -30)

// 画像ブロックを移動
layout.MoveBlock(gopdf.ContentBlockTypeImage, 0, 10, 10)
```

### ブロックのリサイズ

```go
// テキストブロックのサイズを変更
layout.ResizeBlock(gopdf.ContentBlockTypeText, 0, 400, 35)

// 画像ブロックのサイズを変更
layout.ResizeBlock(gopdf.ContentBlockTypeImage, 0, 200, 150)
```

### 重なり検出

```go
overlaps := layout.DetectOverlaps()

for _, overlap := range overlaps {
    fmt.Printf("Overlap detected: %s[%d] and %s[%d]\n",
        overlap.Type1, overlap.Index1,
        overlap.Type2, overlap.Index2)
}
```

### ページ分割

コンテンツが1ページに収まらない場合、複数ページに分割：

```go
maxHeight := 700.0    // 1ページの最大高さ
topMargin := 50.0     // 上マージン
bottomMargin := 50.0  // 下マージン

pages, err := layout.SplitIntoPages(maxHeight, topMargin, bottomMargin)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Split into %d pages\n", len(pages))
```

### 自動レイアウト調整

```go
options := gopdf.LayoutAdjustmentOptions{
    Strategy:   gopdf.StrategyFlowDown,  // 下方向に流し込み
    MinSpacing: 10.0,                    // 最小間隔
    MaxWidth:   500.0,                   // 最大幅
}

err := layout.AdjustLayout(options)
if err != nil {
    log.Fatal(err)
}
```

### 調整戦略

```go
// 下方向に流し込み
gopdf.StrategyFlowDown

// 右方向に流し込み
gopdf.StrategyFlowRight

// コンパクトに配置
gopdf.StrategyCompact
```

## プレゼンテーションサイズ

プレゼンテーション用のページサイズを使用：

```go
// 16:9ワイドスクリーン
page := doc.AddPage(gopdf.PageSizePresentation16x9, gopdf.Portrait)

// 4:3スタンダード
page := doc.AddPage(gopdf.PageSizePresentation4x3, gopdf.Portrait)
```

### プレゼンテーションの作成

```go
doc := gopdf.New()

// タイトルスライド
page1 := doc.AddPage(gopdf.PageSizePresentation16x9, gopdf.Portrait)
page1.SetFont(gopdf.FontHelveticaBold, 48)
page1.DrawText("Presentation Title", 50, 250)
page1.SetFont(gopdf.FontHelvetica, 24)
page1.DrawText("Presenter Name", 50, 200)

// コンテンツスライド
page2 := doc.AddPage(gopdf.PageSizePresentation16x9, gopdf.Portrait)
page2.SetFont(gopdf.FontHelveticaBold, 36)
page2.DrawText("Slide Title", 50, 350)
page2.SetFont(gopdf.FontHelvetica, 18)
page2.DrawText("• Point 1", 70, 300)
page2.DrawText("• Point 2", 70, 270)
page2.DrawText("• Point 3", 70, 240)
```

## 完全な例：ルビ付き文書

```go
package main

import (
    "os"
    "log"
    "github.com/ryomak/gopdf"
)

func main() {
    jpFont, err := gopdf.LoadSystemJapaneseFont()
    if err != nil {
        log.Fatal(err)
    }

    doc := gopdf.New()
    page := doc.AddPage(gopdf.PageSizeA4, gopdf.Portrait)
    page.SetTTFFont(jpFont, 24)

    style := gopdf.DefaultRubyStyle()
    style.SizeRatio = 0.5
    style.Alignment = gopdf.RubyAlignCenter

    // タイトル
    page.SetTTFFont(jpFont, 32)
    title := []gopdf.RubyText{
        gopdf.NewRubyText("日", "に"),
        gopdf.NewRubyText("本", "ほん"),
        gopdf.NewRubyText("語", "ご"),
        gopdf.NewRubyText("入", "にゅう"),
        gopdf.NewRubyText("門", "もん"),
    }
    page.DrawRubyTexts(title, 100, 750, style)

    // 本文
    page.SetTTFFont(jpFont, 18)
    style.SizeRatio = 0.4

    y := 680.0
    sentences := [][]gopdf.RubyText{
        {
            gopdf.NewRubyText("私", "わたし"),
            gopdf.NewRubyText("は", ""),
            gopdf.NewRubyText("学", "がく"),
            gopdf.NewRubyText("生", "せい"),
            gopdf.NewRubyText("です", ""),
        },
        {
            gopdf.NewRubyText("東", "とう"),
            gopdf.NewRubyText("京", "きょう"),
            gopdf.NewRubyText("に", ""),
            gopdf.NewRubyText("住", "す"),
            gopdf.NewRubyText("んでいます", ""),
        },
    }

    for _, sentence := range sentences {
        page.DrawRubyTexts(sentence, 100, y, style)
        y -= 40
    }

    file, _ := os.Create("ruby_document.pdf")
    defer file.Close()
    doc.WriteTo(file)
}
```

## サンプルコード

- [examples/11_ruby_annotation](https://github.com/ryomak/gopdf/tree/main/examples/11_ruby_annotation) - ルビ
- [examples/12_ocr_text_layer](https://github.com/ryomak/gopdf/tree/main/examples/12_ocr_text_layer) - OCRテキストレイヤー
- [examples/16_presentation_sizes](https://github.com/ryomak/gopdf/tree/main/examples/16_presentation_sizes) - プレゼンテーション
- [examples/17_layout_adjustment](https://github.com/ryomak/gopdf/tree/main/examples/17_layout_adjustment) - レイアウト調整

## 関連ページ

- [テキストとフォント](text-and-fonts.md) - 基本的なテキスト描画
- [PDF翻訳](translation.md) - レイアウト保持翻訳
- [PDF読み込み・解析](reading-pdf.md) - レイアウト抽出
