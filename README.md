# gopdf

Pure GoでPDF生成・解析・翻訳を行う高機能ライブラリ

[![Go Reference](https://pkg.go.dev/badge/github.com/ryomak/gopdf.svg)](https://pkg.go.dev/github.com/ryomak/gopdf)
[![Test](https://github.com/ryomak/gopdf/actions/workflows/test.yml/badge.svg)](https://github.com/ryomak/gopdf/actions/workflows/test.yml)
[![Go Version](https://img.shields.io/github/go-mod-go-version/ryomak/gopdf)](https://github.com/ryomak/gopdf)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## 概要

`gopdf` は、CGOを使用せず、Go標準ライブラリのみで動作するPDFライブラリです。PDF 1.7（ISO 32000-1:2008）仕様に準拠し、PDFの生成・解析・翻訳を行います。

### 特徴

- **Pure Go**: CGO不要、外部ライブラリへの依存なし
- **シンプルなAPI**: 直感的で使いやすいAPI設計
- **型安全**: Goの型システムを活用した安全な設計
- **PDF翻訳**: レイアウトを保持したままテキストを翻訳
- **日本語対応**: システムフォントまたは任意のTTFフォントを使用可能
- **CLIツール**: コマンドラインから直接PDF操作が可能

## 主な機能

### PDF生成
- 標準フォント（14種類）とTTFフォント対応
- 日本語・中国語・韓国語などの多言語テキスト
- 図形描画（線、矩形、円）
- JPEG/PNG画像埋め込み（透明度対応）
- Markdown → PDF変換

### PDF解析
- テキスト抽出（位置・フォント情報付き）
- 画像抽出
- メタデータ取得

### PDF翻訳
- レイアウト保持翻訳
- フォント・サイズの自動保持
- 文単位/行単位の翻訳オプション

## インストール

### ライブラリとして使用

```bash
go get github.com/ryomak/gopdf
```

### CLIツールとして使用

```bash
go install github.com/ryomak/gopdf/cmd/gopdf@latest
```

## CLIツール

コマンドラインからPDF操作ができるCLIツールを提供しています。

### 主なコマンド

```bash
# PDF情報表示
gopdf info document.pdf
gopdf info document.pdf --json

# テキスト抽出
gopdf extract text document.pdf
gopdf extract text document.pdf --format json

# 画像抽出
gopdf extract images document.pdf --output ./images/

# PDF暗号化
gopdf encrypt input.pdf output.pdf --user-password "secret"

# PDF復号
gopdf decrypt protected.pdf decrypted.pdf --password "secret"

# Markdown → PDF変換
gopdf markdown README.md output.pdf
gopdf markdown slides.md presentation.pdf --mode slide

# PDF翻訳（外部翻訳コマンド使用）
gopdf translate input.pdf output.pdf --font japanese.ttf --command "trans -b :ja"
```

詳細は [CLIドキュメント](docs/cli.md) を参照してください。

## クイックスタート

### PDF生成

```go
package main

import (
    "os"
    "github.com/ryomak/gopdf"
)

func main() {
    doc := gopdf.New()
    page := doc.AddPage(gopdf.PageSizeA4, gopdf.Portrait)

    // 標準フォントでテキスト描画
    page.SetFont(gopdf.FontHelvetica, 24)
    page.DrawText("Hello, World!", 100, 750)

    // ファイルに出力
    file, _ := os.Create("output.pdf")
    defer file.Close()
    doc.WriteTo(file)
}
```

### 日本語テキスト

```go
// 方法1: システムフォントを使用（macOS: Hiragino, Linux: Noto, Windows: Yu Gothic）
jpFont, err := gopdf.LoadSystemJapaneseFont()
if err != nil {
    log.Fatal("Japanese font not found on system")
}
page.SetFont(jpFont, 24)
page.DrawText("こんにちは、世界！", 100, 750)

// 方法2: 任意のTTFフォントを指定
jpFont, err := gopdf.LoadTTF("/path/to/NotoSansJP-Regular.ttf")
```

### PDF翻訳

```go
// 翻訳辞書
dict := map[string]string{
    "Hello": "こんにちは",
    "World": "世界",
}

jpFont, _ := gopdf.LoadSystemJapaneseFont()

opts := gopdf.PDFTranslatorOptions{
    Translator: gopdf.TranslateFunc(func(text string) (string, error) {
        if t, ok := dict[text]; ok {
            return t, nil
        }
        return text, nil
    }),
    TargetFont:    jpFont,  // TargetFontNameは自動取得
    TranslateUnit: gopdf.TranslateUnitSentence, // 文単位で翻訳
    KeepLayout:    true,
    KeepImages:    true,
}

gopdf.TranslatePDF("input.pdf", "output.pdf", opts)
```

### Markdown → PDF

```go
doc, err := gopdf.NewMarkdownDocumentFromFile("input.md", &gopdf.MarkdownOptions{
    Mode:        gopdf.MarkdownModeDocument,
    PageSize:    gopdf.PageSizeA4,
    Orientation: gopdf.Portrait,
})
if err != nil {
    log.Fatal(err)
}

file, _ := os.Create("output.pdf")
defer file.Close()
doc.WriteTo(file)
```

## API

### フォント

```go
// 標準フォント（14種類、埋め込み不要）
gopdf.FontHelvetica
gopdf.FontHelveticaBold
gopdf.FontTimesRoman
gopdf.FontCourier
// ... 他

// TTFフォント
font, err := gopdf.LoadTTF("path/to/font.ttf")           // ファイルから
font, err := gopdf.LoadTTFFromBytes(data)                // バイト列から
font, err := gopdf.LoadSystemJapaneseFont()              // システムフォント
```

### 翻訳オプション

```go
type PDFTranslatorOptions struct {
    Translator     Translator    // 翻訳関数
    TargetFont     Font          // 非ASCII用フォント
    TargetFontName string        // フォント名（省略可、自動取得）
    TranslateUnit  TranslateUnit // 翻訳単位
    KeepLayout     bool          // レイアウト保持
    KeepImages     bool          // 画像保持
}

// 翻訳単位
gopdf.TranslateUnitBlock    // ブロック全体
gopdf.TranslateUnitLine     // 行単位
gopdf.TranslateUnitSentence // 文単位（. 。 ! ? で区切り）
```

#### Functional Options パターン（v0.20.0+）

構造体の代わりに、`WithTranslatorXXX` 関数を使ってオプションを設定できます：

```go
jpFont, _ := gopdf.LoadSystemJapaneseFont()

opts := gopdf.NewTranslatorOptions(
    gopdf.WithTranslatorFunc(myTranslator),
    gopdf.WithTranslatorTargetFont(jpFont),
    gopdf.WithTranslatorUnit(gopdf.TranslateUnitSentence),
    gopdf.WithTranslatorKeepLayout(true),
    gopdf.WithTranslatorKeepImages(true),
)

gopdf.TranslatePDF("input.pdf", "output.pdf", opts)
```

### テキストフィッティングオプション

```go
// Functional Options パターン
opts := gopdf.NewFitOptions(
    gopdf.WithFitMaxFontSize(24),
    gopdf.WithFitMinFontSize(8),
    gopdf.WithFitLineSpacing(1.2),
    gopdf.WithFitAlignment(gopdf.AlignCenter),
)
```

### フォント設定

```go
// 統一API（StandardFont / TTFFont 両対応）
page.SetFont(gopdf.FontHelvetica, 12)  // StandardFont
page.SetFont(ttfFont, 12)               // TTFFont
```

## アーキテクチャ

```
┌─────────────────────────────────────┐
│        API Layer (gopdf/)           │  公開API
├─────────────────────────────────────┤
│  Content Layer                      │  描画・抽出・翻訳
│  (internal/content/)                │
├─────────────────────────────────────┤
│  Writer Layer     │  Reader Layer   │  生成・解析
│  (internal/writer)│  (internal/reader)
├───────────────────┴─────────────────┤
│  Font Layer (internal/font/)        │  フォント管理
├─────────────────────────────────────┤
│  Core Layer (internal/core/)        │  PDF基本オブジェクト
└─────────────────────────────────────┘
```

## サンプル

[`examples/`](examples/) ディレクトリを参照してください。

| サンプル | 説明 |
|---------|------|
| `01_empty_page` | 空白ページの作成 |
| `02_hello_world` | テキスト描画 |
| `03_graphics` | 図形描画 |
| `04_images` | JPEG画像埋め込み |
| `05_png_images` | PNG画像（透明度対応） |
| `06_read_pdf` | PDF読み込み・テキスト抽出 |
| `07_structured_text` | テキスト抽出（位置情報付き） |
| `08_extract_images` | 画像抽出 |
| `09_ttf_fonts` | TTFフォント・日本語テキスト |
| `10_pdf_translation` | PDF翻訳 |
| `11_ruby_annotation` | ルビ（振り仮名） |
| `12_ocr_text_layer` | OCRテキストレイヤー |
| `13_encryption` | PDF暗号化 |
| `14_decrypt_pdf` | PDF復号 |
| `15_metadata` | メタデータ設定・取得 |
| `16_presentation_sizes` | プレゼンサイズ（16:9, 4:3） |
| `17_layout_adjustment` | レイアウト調整 |
| `18_markdown` | Markdown → PDF変換 |

## 開発

```bash
# テスト
make test

# CI（テスト + lint）
make ci
```

## ライセンス

MIT License
