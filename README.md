# gopdf

Pure GoでPDF生成・解析・翻訳を行う高機能ライブラリ

[![Go Reference](https://pkg.go.dev/badge/github.com/ryomak/gopdf.svg)](https://pkg.go.dev/github.com/ryomak/gopdf)
[![Test](https://github.com/ryomak/gopdf/actions/workflows/test.yml/badge.svg)](https://github.com/ryomak/gopdf/actions/workflows/test.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/ryomak/gopdf)](https://github.com/ryomak/gopdf)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## 概要

`gopdf` は、CGOを使用せず、Go標準ライブラリのみで動作するPDFライブラリです。PDF 1.7（ISO 32000-1:2008）仕様に準拠し、PDFの生成・解析・翻訳を行います。

### 特徴

- **Pure Go**: CGO不要、外部ライブラリへの依存なし
- **シンプルなAPI**: 直感的で使いやすいAPI設計
- **型安全**: Goの型システムを活用した安全な設計
- **PDF翻訳**: レイアウトを保持したままテキストを翻訳
- **日本語フォント内蔵**: Koruriフォントが埋め込み済み

## 主な機能

### PDF生成
- 標準フォント（14種類）とTTFフォント対応
- 日本語・中国語・韓国語などの多言語テキスト
- 図形描画（線、矩形、円）
- JPEG/PNG画像埋め込み（透明度対応）

### PDF解析
- テキスト抽出（位置・フォント情報付き）
- 画像抽出
- メタデータ取得

### PDF翻訳
- レイアウト保持翻訳
- フォント・サイズの自動保持
- 文単位/行単位の翻訳オプション

## インストール

```bash
go get github.com/ryomak/gopdf
```

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
// 内蔵の日本語フォント（Koruri）を使用
jpFont, _ := gopdf.DefaultJapaneseFont()
page.SetTTFFont(jpFont, 24)
page.DrawText("こんにちは、世界！", 100, 750)
```

### PDF翻訳

```go
// 翻訳辞書
dict := map[string]string{
    "Hello": "こんにちは",
    "World": "世界",
}

jpFont, _ := gopdf.DefaultJapaneseFont()

opts := gopdf.PDFTranslatorOptions{
    Translator: gopdf.TranslateFunc(func(text string) (string, error) {
        if t, ok := dict[text]; ok {
            return t, nil
        }
        return text, nil
    }),
    TargetFont:    jpFont,
    TargetFontName: "Koruri",
    TranslateUnit: gopdf.TranslateUnitSentence, // 文単位で翻訳
    KeepLayout:    true,
    KeepImages:    true,
}

gopdf.TranslatePDF("input.pdf", "output.pdf", opts)
```

## API

### 主要な型

```go
// Font インターフェース - StandardFont と *TTFFont が実装
type Font interface {
    Name() string
}

// 標準フォント
gopdf.FontHelvetica
gopdf.FontHelveticaBold
gopdf.FontTimesRoman
gopdf.FontCourier
// ... 他14種類

// TTFフォント
font, err := gopdf.LoadTTF("path/to/font.ttf")
jpFont, err := gopdf.DefaultJapaneseFont() // 内蔵Koruri
```

### 翻訳オプション

```go
type PDFTranslatorOptions struct {
    Translator     Translator    // 翻訳関数
    TargetFont     Font          // 非ASCII用フォント
    TargetFontName string        // フォント名
    TranslateUnit  TranslateUnit // 翻訳単位
    KeepLayout     bool          // レイアウト保持
    KeepImages     bool          // 画像保持
}

// 翻訳単位
gopdf.TranslateUnitBlock    // ブロック全体
gopdf.TranslateUnitLine     // 行単位
gopdf.TranslateUnitSentence // 文単位（. 。 ! ? で区切り）
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
| `09_ttf_fonts` | TTFフォント・日本語テキスト |
| `10_pdf_translation` | PDF翻訳 |

## 開発

```bash
# テスト
make test

# CI（テスト + lint）
make ci
```

## ライセンス

MIT License

内蔵フォント（Koruri）: Apache License 2.0

## 関連プロジェクト

- [pdfcpu](https://github.com/pdfcpu/pdfcpu) - Go製PDFプロセッサ
- [gofpdf](https://github.com/jung-kurt/gofpdf) - PDF生成ライブラリ
