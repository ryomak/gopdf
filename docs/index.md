# gopdf ドキュメント

Pure GoでPDF生成・解析・翻訳を行う高機能ライブラリ

## 概要

`gopdf`は、CGOを使用せず、Go標準ライブラリのみで動作するPDFライブラリです。PDF 1.7（ISO 32000-1:2008）仕様に準拠し、PDFの生成・解析・翻訳を行います。

### 特徴

- **Pure Go**: CGO不要、外部ライブラリへの依存なし
- **シンプルなAPI**: 直感的で使いやすいAPI設計
- **型安全**: Goの型システムを活用した安全な設計
- **日本語対応**: システムフォントまたは任意のTTFフォントを使用可能
- **CLIツール**: コマンドラインから直接PDF操作が可能

## ドキュメント目次

### CLIツール

- [gopdf CLI](cli.md) - コマンドラインツールの使い方

### 入門

- [はじめに](getting-started.md) - インストールとクイックスタート

### 基本機能

- [テキストとフォント](text-and-fonts.md) - テキスト描画、標準フォント、TTFフォント
- [図形描画](graphics.md) - 線、矩形、円、色の設定
- [画像の埋め込み](images.md) - JPEG/PNG画像の埋め込み

### PDF操作

- [PDF読み込み・解析](reading-pdf.md) - テキスト抽出、画像抽出、メタデータ取得
- [PDF翻訳](translation.md) - レイアウト保持翻訳
- [Markdown変換](markdown.md) - MarkdownからPDFへの変換

### セキュリティ

- [暗号化とパスワード保護](encryption.md) - PDF暗号化、権限設定

### 高度な機能

- [メタデータ](metadata.md) - PDFメタデータの設定と取得
- [高度な機能](advanced.md) - ルビ、OCRテキストレイヤー、レイアウト調整

## クイックリファレンス

### 基本的な使い方

```go
package main

import (
    "os"
    "github.com/ryomak/gopdf"
)

func main() {
    // ドキュメント作成
    doc := gopdf.New()
    page := doc.AddPage(gopdf.PageSizeA4, gopdf.Portrait)

    // テキスト描画
    page.SetFont(gopdf.FontHelvetica, 24)
    page.DrawText("Hello, World!", 100, 750)

    // ファイルに出力
    file, _ := os.Create("output.pdf")
    defer file.Close()
    doc.WriteTo(file)
}
```

### ページサイズ一覧

| 定数 | サイズ | 説明 |
|------|--------|------|
| `PageSizeA4` | 595 x 842 pt | A4 (210mm x 297mm) |
| `PageSizeLetter` | 612 x 792 pt | Letter (8.5" x 11") |
| `PageSizeLegal` | 612 x 1008 pt | Legal (8.5" x 14") |
| `PageSizeA3` | 842 x 1191 pt | A3 (297mm x 420mm) |
| `PageSizeA5` | 420 x 595 pt | A5 (148mm x 210mm) |
| `PageSizePresentation16x9` | 720 x 405 pt | 16:9プレゼン |
| `PageSizePresentation4x3` | 720 x 540 pt | 4:3プレゼン |

### 標準フォント一覧

| フォント | 定数 |
|----------|------|
| Helvetica | `FontHelvetica`, `FontHelveticaBold`, `FontHelveticaOblique`, `FontHelveticaBoldOblique` |
| Times | `FontTimesRoman`, `FontTimesBold`, `FontTimesItalic`, `FontTimesBoldItalic` |
| Courier | `FontCourier`, `FontCourierBold`, `FontCourierOblique`, `FontCourierBoldOblique` |
| Symbol | `FontSymbol` |
| ZapfDingbats | `FontZapfDingbats` |

## サンプルコード

リポジトリの[examples/](https://github.com/ryomak/gopdf/tree/main/examples)ディレクトリに、各機能のサンプルコードがあります。

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

## ライセンス

MIT License
