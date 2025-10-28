# gopdf

Pure GoでPDF生成・解析を行う高機能ライブラリ

[![Go Reference](https://pkg.go.dev/badge/github.com/ryomak/gopdf.svg)](https://pkg.go.dev/github.com/ryomak/gopdf)
[![Test](https://github.com/ryomak/gopdf/actions/workflows/test.yml/badge.svg)](https://github.com/ryomak/gopdf/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/ryomak/gopdf/branch/main/graph/badge.svg)](https://codecov.io/gh/ryomak/gopdf)
[![Go Version](https://img.shields.io/github/go-mod/go-version/ryomak/gopdf)](https://github.com/ryomak/gopdf)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

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

## API

### 主要な型とインターフェース

#### PDF生成

**Document**
```go
type Document struct { ... }

// 新規PDFドキュメントを作成
func New() *Document

// ページを追加
func (d *Document) AddPage(size PageSize, orientation Orientation) *Page

// PDFを出力
func (d *Document) WriteTo(w io.Writer) error

// 暗号化を設定
func (d *Document) SetEncryption(opts *EncryptionOptions)

// メタデータを設定
func (d *Document) SetMetadata(metadata *Metadata)
```

**Page**
```go
type Page struct { ... }

// ページサイズを取得
func (p *Page) Width() float64
func (p *Page) Height() float64

// テキスト描画
func (p *Page) SetFont(f font.StandardFont, size float64) error
func (p *Page) DrawText(text string, x, y float64) error  // 標準フォントとTTFフォントの両方に対応
func (p *Page) SetTTFFont(font *TTFFont, size float64) error
func (p *Page) DrawTextUTF8(text string, x, y float64) error  // Deprecated: DrawTextを使用してください

// 図形描画
func (p *Page) DrawLine(x1, y1, x2, y2 float64)
func (p *Page) DrawRect(x, y, width, height float64)
func (p *Page) FillRect(x, y, width, height float64)
func (p *Page) DrawCircle(x, y, r float64)
func (p *Page) FillCircle(x, y, r float64)

// 色設定
func (p *Page) SetStrokeColor(r, g, b float64)
func (p *Page) SetFillColor(r, g, b float64)
func (p *Page) SetLineWidth(width float64)

// 画像描画
func (p *Page) DrawImage(img *Image, x, y, width, height float64) error
func (p *Page) DrawJPEG(jpegData []byte, x, y, width, height float64) error
func (p *Page) DrawPNG(pngData []byte, x, y, width, height float64) error
```

#### PDF解析

**PDFReader**
```go
type PDFReader struct { ... }

// PDFファイルを開く
func Open(path string) (*PDFReader, error)
func OpenReader(r io.ReadSeeker) (*PDFReader, error)

// 基本情報を取得
func (r *PDFReader) PageCount() int
func (r *PDFReader) Info() Metadata

// テキスト抽出
func (r *PDFReader) ExtractText() (string, error)
func (r *PDFReader) ExtractPageText(pageIndex int) (string, error)
func (r *PDFReader) ExtractStructuredText(pageIndex int) ([]TextElement, error)

// 画像抽出
func (r *PDFReader) ExtractImages(pageIndex int) ([]ImageInfo, error)

// リソース解放
func (r *PDFReader) Close() error
```

**TextElement**（構造化テキスト抽出）
```go
type TextElement struct {
    Text   string  // テキスト内容
    X      float64 // X座標（左下原点）
    Y      float64 // Y座標（左下原点）
    Width  float64 // テキストの幅（概算）
    Height float64 // テキストの高さ（フォントサイズ）
    Font   string  // フォント名
    Size   float64 // フォントサイズ
}
```

**ImageInfo**（画像情報）
```go
type ImageInfo struct {
    Name        string      // リソース名（例: "Im1"）
    Width       int         // 画像の幅
    Height      int         // 画像の高さ
    ColorSpace  string      // 色空間（DeviceRGB, DeviceGray, DeviceCMYK）
    BitsPerComp int         // ビット深度
    Filter      string      // 圧縮フィルター
    Data        []byte      // 画像データ
    Format      ImageFormat // 画像フォーマット
}
```

#### 共通型

**PageSize**（ページサイズ）
```go
var (
    A4     PageSize  // 210 x 297 mm
    Letter PageSize  // 8.5 x 11 inch
    Legal  PageSize  // 8.5 x 14 inch
    A3     PageSize  // 297 x 420 mm
    A5     PageSize  // 148 x 210 mm
)
```

**Orientation**（ページ向き）
```go
const (
    Portrait  Orientation  // 縦向き
    Landscape Orientation  // 横向き
)
```

詳細なAPIドキュメントは [pkg.go.dev](https://pkg.go.dev/github.com/ryomak/gopdf) を参照してください。

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

    // 日本語テキストを描画（DrawTextは自動的にTTFフォントを使用）
    page.DrawText("こんにちは、世界！", 100, 750)
    page.DrawText("gopdf supports Japanese text!", 100, 720)

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
