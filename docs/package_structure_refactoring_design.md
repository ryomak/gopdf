# gopdf パッケージ構造リファクタリング設計書

## 1. 目的

gopdfをモジュール化し、以下を実現する：
1. **機能の分離**: reader, markdown, layout, ocr, translator などをサブパッケージに分割
2. **共通型の共有**: `internal/model` に共通型を配置し、複数のパッケージで利用
3. **ユーザビリティの維持**: root パッケージにエイリアスを配置し、簡潔なAPIを提供

## 2. 現状の問題点

- root パッケージに多数の機能が混在（41ファイル）
- layout.go が 849行と大きすぎる
- 共通型が layout/ パッケージに配置されている
- 機能ごとの責務が不明確

## 3. 設計方針

### 3.1. パッケージ構成

```
gopdf/
├── document.go          # コア: Document型（既存）
├── page.go              # コア: Page型（既存）
├── font.go              # コア: Font型（既存）
├── graphics.go          # コア: Color等（既存）
├── constants.go         # コア: PageSize等（既存）
├── image.go             # コア: Image型（既存）
├── encryption.go        # コア: Encryption型（既存）
├── metadata.go          # コア: Metadata型（既存）
├── ruby.go              # コア: RubyText型（既存）
├── types.go             # エイリアス: internal/model の型
├── aliases.go           # エイリアス: サブパッケージの型と関数
│
├── reader/              # サブパッケージ: PDF読み込み
│   ├── reader.go
│   ├── text.go
│   ├── image.go
│   ├── layout.go
│   └── metadata.go
│
├── markdown/            # サブパッケージ: Markdown変換
│   ├── markdown.go
│   ├── renderer.go
│   └── styles.go
│
├── layout/              # サブパッケージ: レイアウト調整（既存）
│   ├── layout.go
│   ├── blocks.go
│   ├── adjustment.go
│   ├── strategies.go
│   ├── operations.go
│   └── image_utils.go
│
├── ocr/                 # サブパッケージ: OCR/TextLayer
│   ├── text_layer.go
│   └── ocr.go
│
├── translator/          # サブパッケージ: PDF翻訳
│   ├── translator.go
│   └── renderer.go
│
└── internal/
    ├── model/           # 新規: 共通型定義
    │   ├── geometry.go  # Rectangle
    │   ├── color.go     # Color
    │   ├── text.go      # TextElement
    │   ├── image.go     # ImageInfo, ImageFormat
    │   └── image_utils.go
    ├── core/            # PDF基本オブジェクト（既存）
    ├── font/            # フォント管理（既存）
    ├── writer/          # PDF生成（既存）
    ├── reader/          # PDF解析（既存）
    ├── content/         # コンテンツ処理（既存）
    ├── security/        # セキュリティ（既存）
    └── utils/           # ユーティリティ（既存）
```

### 3.2. 依存関係

```
┌─────────────────────────┐
│   gopdf (root)          │ ← ユーザーが使う公開API（エイリアスのみ）
└─────────────────────────┘
         ↓ import
┌─────────────────────────┐
│ reader/  markdown/      │
│ layout/  ocr/           │ ← サブパッケージ
│ translator/             │
└─────────────────────────┘
         ↓ import
┌─────────────────────────┐
│   internal/model        │ ← 共通型定義
└─────────────────────────┘
         ↓ import (from other internals)
┌─────────────────────────┐
│ internal/core/          │
│ internal/reader/        │ ← 内部実装
│ internal/content/       │
│ ...                     │
└─────────────────────────┘
```

## 4. internal/model パッケージの設計

### 4.1. 配置する型

**geometry.go**
```go
package model

// Rectangle は矩形領域を表す
type Rectangle struct {
    X      float64 // 左下X座標
    Y      float64 // 左下Y座標
    Width  float64 // 幅
    Height float64 // 高さ
}
```

**color.go**
```go
package model

// Color は色を表す（RGB形式）
type Color struct {
    R, G, B float64 // 0.0-1.0 の範囲
}
```

**text.go**
```go
package model

// TextElement は単一のテキスト要素を表す
type TextElement struct {
    Text   string  // テキスト内容
    X      float64 // X座標
    Y      float64 // Y座標
    Width  float64 // 幅
    Height float64 // 高さ
    Font   string  // フォント名
    Size   float64 // フォントサイズ
}
```

**image.go**
```go
package model

// ImageFormat は画像フォーマットを表す
type ImageFormat string

const (
    ImageFormatJPEG    ImageFormat = "jpeg"
    ImageFormatPNG     ImageFormat = "png"
    ImageFormatUnknown ImageFormat = "unknown"
)

// ImageInfo は画像の情報を保持する
type ImageInfo struct {
    Name        string
    Width       int
    Height      int
    ColorSpace  string
    BitsPerComp int
    Filter      string
    Data        []byte
    Format      ImageFormat
}
```

**image_utils.go**
```go
package model

// SaveImage は画像をファイルに保存する
func (img *ImageInfo) SaveImage(filename string) error { ... }

// ToImage は画像をimage.Imageに変換する
func (img *ImageInfo) ToImage() (image.Image, error) { ... }
```

### 4.2. 依存関係のルール

- `internal/model` は標準ライブラリのみに依存
- 他の internal パッケージに依存しない
- 純粋なデータ型とそのメソッドのみを配置

## 5. layout/ パッケージの更新

### 5.1. internal/model を使用

```go
// layout/blocks.go
package layout

import "github.com/ryomak/gopdf/internal/model"

// TextBlock はテキストの論理的なブロック
type TextBlock struct {
    Text     string              // テキスト内容
    Elements []model.TextElement // internal/model を使用
    Rect     model.Rectangle     // internal/model を使用
    Font     string
    FontSize float64
    Color    model.Color         // internal/model を使用
}

// ImageBlock は画像の配置情報
type ImageBlock struct {
    model.ImageInfo              // internal/model を使用
    X            float64
    Y            float64
    PlacedWidth  float64
    PlacedHeight float64
}
```

### 5.2. 後方互換性のためのエイリアス（オプション）

layout パッケージにもエイリアスを残すことで、既存コードの動作を保証：

```go
// layout/types.go
package layout

import "github.com/ryomak/gopdf/internal/model"

// 後方互換性のためのエイリアス
type (
    Rectangle   = model.Rectangle
    Color       = model.Color
    TextElement = model.TextElement
    ImageInfo   = model.ImageInfo
    ImageFormat = model.ImageFormat
)

const (
    ImageFormatJPEG    = model.ImageFormatJPEG
    ImageFormatPNG     = model.ImageFormatPNG
    ImageFormatUnknown = model.ImageFormatUnknown
)
```

## 6. サブパッケージの作成

### 6.1. reader/ パッケージ

**reader/reader.go**
```go
package reader

import (
    "io"
    "github.com/ryomak/gopdf/internal/model"
    "github.com/ryomak/gopdf/layout"
)

// PDFReader はPDFファイルを読み込むための構造体
type PDFReader struct {
    impl   *internal_reader.Reader
    closer io.Closer
}

// Open はPDFファイルを開く
func Open(path string) (*PDFReader, error) { ... }

// OpenReader はio.ReadSeekerからPDFを開く
func OpenReader(r io.ReadSeeker) (*PDFReader, error) { ... }

// PageCount はページ数を返す
func (r *PDFReader) PageCount() int { ... }

// Close はリーダーを閉じる
func (r *PDFReader) Close() error { ... }
```

**reader/text.go**
```go
package reader

import "github.com/ryomak/gopdf/internal/model"

// ExtractPageText はページのテキストを抽出
func (r *PDFReader) ExtractPageText(pageNum int) (string, error) { ... }

// ExtractTextElements はテキスト要素を抽出
func (r *PDFReader) ExtractTextElements(pageNum int) ([]model.TextElement, error) { ... }
```

**reader/layout.go**
```go
package reader

import "github.com/ryomak/gopdf/layout"

// ExtractPageLayout はページレイアウトを抽出
func (r *PDFReader) ExtractPageLayout(pageNum int) (*layout.PageLayout, error) { ... }
```

### 6.2. markdown/ パッケージ

**markdown/markdown.go**
```go
package markdown

import "github.com/ryomak/gopdf"

// Options はMarkdown変換のオプション
type Options struct {
    PageSize     gopdf.PageSize
    Margins      Margins
    DefaultFont  interface{}
    Style        *Style
    Mode         Mode
}

// NewDocument はMarkdownからPDFドキュメントを作成
func NewDocument(markdownText string, opts *Options) (*gopdf.Document, error) { ... }

// NewDocumentFromFile はファイルからMarkdown PDFを作成
func NewDocumentFromFile(filepath string, opts *Options) (*gopdf.Document, error) { ... }
```

### 6.3. ocr/ パッケージ

**ocr/text_layer.go**
```go
package ocr

import "github.com/ryomak/gopdf/internal/model"

// TextLayer はOCR用のテキストレイヤー
type TextLayer struct {
    Words      []TextLayerWord
    RenderMode TextRenderMode
    Opacity    float64
}

// TextLayerWord はテキストレイヤーの単語
type TextLayerWord struct {
    Text   string
    Bounds model.Rectangle
}
```

### 6.4. translator/ パッケージ

**translator/translator.go**
```go
package translator

import (
    "io"
    "github.com/ryomak/gopdf/layout"
)

// Translator はテキスト翻訳のインターフェース
type Translator interface {
    Translate(text string) (string, error)
}

// Options は翻訳オプション
type Options struct {
    Translator     Translator
    TargetFont     interface{}
    FontName       string
    PreserveLayout bool
    MaxFontSize    float64
    MinFontSize    float64
}

// TranslatePDF はPDFを翻訳
func TranslatePDF(inputPath, outputPath string, opts Options) error { ... }
```

## 7. gopdf パッケージのエイリアス

### 7.1. types.go（共通型のエイリアス）

```go
// gopdf/types.go
package gopdf

import "github.com/ryomak/gopdf/internal/model"

// 共通型のエイリアス
type (
    Rectangle   = model.Rectangle
    Color       = model.Color
    TextElement = model.TextElement
    ImageInfo   = model.ImageInfo
    ImageFormat = model.ImageFormat
)

// 定数のエイリアス
const (
    ImageFormatJPEG    = model.ImageFormatJPEG
    ImageFormatPNG     = model.ImageFormatPNG
    ImageFormatUnknown = model.ImageFormatUnknown
)
```

### 7.2. aliases.go（サブパッケージのエイリアス）

```go
// gopdf/aliases.go
package gopdf

import (
    "github.com/ryomak/gopdf/reader"
    "github.com/ryomak/gopdf/markdown"
    "github.com/ryomak/gopdf/layout"
    "github.com/ryomak/gopdf/ocr"
    "github.com/ryomak/gopdf/translator"
)

// layout パッケージのエイリアス
type (
    ContentBlock            = layout.ContentBlock
    ContentBlockType        = layout.ContentBlockType
    PageLayout              = layout.PageLayout
    TextBlock               = layout.TextBlock
    ImageBlock              = layout.ImageBlock
    BlockOverlap            = layout.BlockOverlap
    LayoutStrategy          = layout.LayoutStrategy
    LayoutAdjustmentOptions = layout.LayoutAdjustmentOptions
)

const (
    ContentBlockTypeText  = layout.ContentBlockTypeText
    ContentBlockTypeImage = layout.ContentBlockTypeImage

    StrategyPreservePosition = layout.StrategyPreservePosition
    StrategyCompact          = layout.StrategyCompact
    StrategyEvenSpacing      = layout.StrategyEvenSpacing
    StrategyFlowDown         = layout.StrategyFlowDown
    StrategyFitContent       = layout.StrategyFitContent
)

// reader パッケージのエイリアス
type PDFReader = reader.PDFReader

// Open はPDFファイルを開く（reader.Openのエイリアス）
func Open(path string) (*PDFReader, error) {
    return reader.Open(path)
}

// markdown パッケージのエイリアス
type (
    MarkdownOptions = markdown.Options
    MarkdownStyle   = markdown.Style
    MarkdownMode    = markdown.Mode
)

// NewMarkdownDocument はMarkdownからPDFを作成（markdown.NewDocumentのエイリアス）
func NewMarkdownDocument(markdownText string, opts *MarkdownOptions) (*Document, error) {
    return markdown.NewDocument(markdownText, opts)
}

// ocr パッケージのエイリアス
type (
    TextLayer     = ocr.TextLayer
    TextLayerWord = ocr.TextLayerWord
    OCRResult     = ocr.OCRResult
)

// translator パッケージのエイリアス
type (
    Translator        = translator.Translator
    TranslatorOptions = translator.Options
)

// TranslatePDF はPDFを翻訳（translator.TranslatePDFのエイリアス）
func TranslatePDF(inputPath, outputPath string, opts TranslatorOptions) error {
    return translator.TranslatePDF(inputPath, outputPath, opts)
}
```

## 8. 使用例

### 8.1. サブパッケージを直接使う場合

```go
import (
    "github.com/ryomak/gopdf/reader"
    "github.com/ryomak/gopdf/layout"
)

// PDFを開く
r, err := reader.Open("test.pdf")
defer r.Close()

// レイアウトを抽出
layout, err := r.ExtractPageLayout(0)

// レイアウトを調整
opts := layout.DefaultLayoutAdjustmentOptions()
opts.Strategy = layout.StrategyFlowDown
layout.AdjustLayout(opts)
```

### 8.2. gopdf パッケージのエイリアスを使う場合

```go
import "github.com/ryomak/gopdf"

// PDFを開く（gopdf.Openエイリアス）
r, err := gopdf.Open("test.pdf")
defer r.Close()

// レイアウトを抽出
layout, err := r.ExtractPageLayout(0)

// レイアウトを調整
opts := gopdf.DefaultLayoutAdjustmentOptions()
opts.Strategy = gopdf.StrategyFlowDown
layout.AdjustLayout(opts)
```

どちらの方法でも使用可能！

## 9. マイグレーション手順

### Phase 1: internal/model パッケージの作成
1. `internal/model/` ディレクトリを作成
2. layout/ から共通型を移動
   - geometry.go (Rectangle)
   - color.go (Color)
   - text.go (TextElement)
   - image.go (ImageInfo, ImageFormat)
   - image_utils.go (ImageInfo のメソッド)
3. テストを作成

### Phase 2: layout/ パッケージの更新
1. `internal/model` をインポート
2. 型定義を `model.*` に変更
3. エイリアスを追加（後方互換性）
4. テストを実行

### Phase 3: reader/ パッケージの作成
1. root パッケージの reader 関連コードを移動
2. `internal/reader` を使用する実装
3. テストを移行

### Phase 4: markdown/, ocr/, translator/ パッケージの作成
1. それぞれのサブパッケージを作成
2. root パッケージから機能を移動
3. テストを移行

### Phase 5: gopdf パッケージのエイリアス作成
1. types.go を作成
2. aliases.go を作成
3. 既存のエイリアスと整合性を確認

### Phase 6: examples の更新
1. すべての examples を新しいAPIに更新
2. 動作確認

### Phase 7: テストとビルド
1. `go test ./...`
2. `make ci`
3. 最終確認

### Phase 8: コミット・プッシュ

## 10. 期待される効果

1. **責務の明確化**: 機能ごとにパッケージが分かれる
2. **コードの再利用性**: internal/model で共通型を共有
3. **保守性の向上**: 各パッケージが独立してテスト可能
4. **ユーザビリティ**: エイリアスにより簡潔なAPIを維持
5. **拡張性**: 新機能を追加しやすい構造

## 11. まとめ

このリファクタリングにより、gopdfは以下の構造になる：

- **root パッケージ**: コア機能 + エイリアス
- **サブパッケージ**: reader, markdown, layout, ocr, translator
- **internal/model**: 共通型定義
- **その他internal**: 内部実装

段階的に実装し、各フェーズで`make ci`を実行して品質を保証する。
