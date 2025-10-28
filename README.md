# gopdf

Pure GoでPDF生成・解析を行う高機能ライブラリ

## 概要

`gopdf` は、CGOを使用せず、Go標準ライブラリのみで動作するPDFライブラリです。PDF 1.7（ISO 32000-1:2008）仕様に準拠し、PDFの生成と解析を行います。

### 特徴

- **Pure Go**: CGO不要、外部ライブラリへの依存なし
- **シンプルなAPI**: 直感的で使いやすいAPI設計
- **型安全**: Goの型システムを活用した安全な設計
- **テスト駆動**: 高いテストカバレッジ
- **標準準拠**: PDF 1.7仕様に準拠

## 主な機能

### PDF生成

- PDFドキュメント生成
- ページ管理（追加、サイズ指定）
- 標準ページサイズ（A4, Letter, Legal, A3, A5）
- ページ向き（Portrait, Landscape）

### テキスト描画

- 標準Type1フォント対応（14種類）: Helvetica系、Times系、Courier系、Symbol、ZapfDingbats
- TTF/OTFフォント対応: TrueType/OpenTypeフォントの埋め込み
- Unicode対応: 日本語・中国語・韓国語などの多言語テキスト描画
- フォントサイズ指定

### 図形描画

- 線の描画
- 矩形の描画（枠線のみ/塗りつぶしのみ/両方）
- 円の描画（枠線のみ/塗りつぶしのみ/両方）
- グラフィックス状態設定（線の太さ、色、スタイル）

### 画像埋め込み

- JPEG画像: DCTDecodeフィルター、RGB/Grayscale/CMYK対応
- PNG画像: FlateDecodeフィルター、RGB/Grayscale対応
- 透明度サポート: アルファチャンネル（SMask）対応
- 画像のサイズ・位置指定

### PDF解析

- PDFファイルの読み込み
- ページ数・メタデータの取得
- テキスト抽出（位置情報付き）
- 画像抽出

## インストール

```bash
go get github.com/ryomak/gopdf
```

## 使い方

### 基本的な使い方

```go
package main

import (
    "os"
    "github.com/ryomak/gopdf"
    "github.com/ryomak/gopdf/internal/font"
)

func main() {
    // 新規PDFドキュメントを作成
    doc := gopdf.New()

    // A4サイズの縦向きページを追加
    page := doc.AddPage(gopdf.A4, gopdf.Portrait)

    // フォントを設定してテキストを描画
    page.SetFont(font.Helvetica, 24)
    page.DrawText("Hello, World!", 100, 750)

    page.SetFont(font.TimesRoman, 14)
    page.DrawText("gopdf - Pure Go PDF library", 100, 720)

    // ファイルに出力
    file, _ := os.Create("output.pdf")
    defer file.Close()

    doc.WriteTo(file)
}
```

### TTFフォント（日本語テキスト）

```go
package main

import (
    "os"
    "github.com/ryomak/gopdf"
)

func main() {
    // 新規PDFドキュメントを作成
    doc := gopdf.New()
    page := doc.AddPage(gopdf.A4, gopdf.Portrait)

    // TTFフォントを読み込み
    font, err := gopdf.LoadTTF("/System/Library/Fonts/ヒラギノ角ゴシック W3.ttc")
    if err != nil {
        panic(err)
    }

    // TTFフォントを設定
    page.SetTTFFont(font, 24)

    // 日本語テキストを描画
    page.DrawTextUTF8("こんにちは、世界！", 100, 750)
    page.DrawTextUTF8("gopdf supports Japanese text!", 100, 720)

    // ファイルに出力
    file, _ := os.Create("japanese.pdf")
    defer file.Close()
    doc.WriteTo(file)
}
```

### テキスト抽出

```go
package main

import (
    "fmt"
    "log"
    "github.com/ryomak/gopdf"
)

func main() {
    // PDFファイルを開く
    reader, err := gopdf.Open("document.pdf")
    if err != nil {
        log.Fatal(err)
    }
    defer reader.Close()

    // ページ数を取得
    pageCount := reader.PageCount()
    fmt.Printf("Total pages: %d\n", pageCount)

    // 特定のページのテキストを抽出
    text, err := reader.ExtractPageText(0) // 0-indexed
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Page 1 text:\n%s\n", text)

    // 全ページのテキストを抽出
    allText, err := reader.ExtractText()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("All text:\n%s\n", allText)
}
```

### サンプルコード

詳細なサンプルコードは [`examples/`](examples/) ディレクトリを参照してください。

- [`01_empty_page`](examples/01_empty_page): 空白ページの作成
- [`02_hello_world`](examples/02_hello_world): テキスト描画と複数フォントの使用
- [`03_graphics`](examples/03_graphics): 図形描画（線、矩形、円）と色の設定
- [`04_images`](examples/04_images): JPEG画像の埋め込みとレイアウト
- [`05_png_images`](examples/05_png_images): PNG画像の埋め込みと透明度（アルファチャンネル）
- [`06_read_pdf`](examples/06_read_pdf): PDFファイルの読み込み、情報取得、テキスト抽出
- [`07_structured_text`](examples/07_structured_text): 構造的テキスト抽出（位置情報付き）
- [`08_extract_images`](examples/08_extract_images): 画像抽出とファイル保存
- [`09_ttf_fonts`](examples/09_ttf_fonts): TTF/OTFフォント使用、Unicode/日本語テキスト描画

## 開発

### 必要な環境

- Go 1.18以上

### テストの実行

```bash
go test ./...
```

### ビルド

```bash
go build ./...
```

## アーキテクチャ

gopdfは以下のレイヤー構造で設計されています：

```
┌─────────────────────────────────────┐
│        API Layer (pkg/gopdf)        │
├─────────────────────────────────────┤
│  Content Layer (描画・抽出機能)      │
├─────────────────────────────────────┤
│  Document Layer (文書管理)           │
├─────────────────────────────────────┤
│  Writer Layer     │  Reader Layer    │
│  (生成・出力)      │  (解析・読込)     │
├──────────────────┼──────────────────┤
│  Font Layer (フォント管理)           │
├─────────────────────────────────────┤
│  Core Layer (PDF基本オブジェクト)    │
└─────────────────────────────────────┘
```

詳細は [docs/architecture.md](docs/architecture.md) を参照してください。

## ドキュメント

詳細な設計ドキュメントは [`docs/`](docs/) ディレクトリを参照してください。

- [アーキテクチャ設計書](docs/architecture.md)
- [PDFバイナリ仕様メモ](docs/pdf_spec_notes.md)

## ライセンス

MIT License（予定）

## 参考

- [PDF 1.7 仕様書（ISO 32000-1:2008）](https://opensource.adobe.com/dc-acrobat-sdk-docs/pdfstandards/PDF32000_2008.pdf)
- [PDF Association](https://pdfa.org/)

## 貢献

現在、このプロジェクトは開発初期段階です。

## 関連プロジェクト

- [pdfcpu](https://github.com/pdfcpu/pdfcpu) - Go製PDFプロセッサ
- [gofpdf](https://github.com/jung-kurt/gofpdf) - PDF生成ライブラリ
- [unipdf](https://github.com/unidoc/unipdf) - 商用PDFライブラリ
