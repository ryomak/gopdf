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

## ステータス

🚧 **開発中** - Phase 4 (画像埋め込み) 完了

現在、基本的なPDF生成、テキスト描画、図形描画、画像埋め込み機能が実装されています。

### 実装済み機能

- ✅ 基本的なPDFドキュメント生成
- ✅ ページ管理（追加、サイズ指定）
- ✅ 標準ページサイズ（A4, Letter, Legal, A3, A5）
- ✅ ページ向き（Portrait, Landscape）
- ✅ テキスト描画
- ✅ 標準Type1フォント対応（14種類）
  - Helvetica系、Times系、Courier系、Symbol、ZapfDingbats
- ✅ フォントサイズ指定
- ✅ **図形描画**
  - 線の描画
  - 矩形の描画（枠線のみ/塗りつぶしのみ/両方）
  - 円の描画（枠線のみ/塗りつぶしのみ/両方）
- ✅ **グラフィックス状態設定**
  - 線の太さ設定
  - ストローク色・塗りつぶし色設定（RGB）
  - 線の端スタイル（Butt, Round, Square）
  - 線の結合スタイル（Miter, Round, Bevel）
- ✅ **画像埋め込み**
  - JPEG画像の読み込みと埋め込み
  - RGB/Grayscale/CMYKカラースペース対応
  - 画像のサイズ・位置指定
  - 複数画像の配置
  - 画像の重複排除（同じ画像の再利用）
- ✅ PDF 1.7準拠の出力

### 今後の実装予定

- [ ] PNG画像対応
- [ ] TTFフォント対応
- [ ] PDF解析・読み込み
- [ ] テキスト抽出

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

### サンプルコード

詳細なサンプルコードは [`examples/`](examples/) ディレクトリを参照してください。

- [`01_empty_page`](examples/01_empty_page): 空白ページの作成
- [`02_hello_world`](examples/02_hello_world): テキスト描画と複数フォントの使用
- [`03_graphics`](examples/03_graphics): 図形描画（線、矩形、円）と色の設定
- [`04_images`](examples/04_images): JPEG画像の埋め込みとレイアウト

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

- [要件定義書](docs/requirements.md)
- [アーキテクチャ設計書](docs/architecture.md)
- [プロジェクト構造設計書](docs/structure.md)
- [PDFバイナリ仕様メモ](docs/pdf_spec_notes.md)
- [開発進捗](docs/progress.md)

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
