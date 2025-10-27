# gopdf アーキテクチャ設計書

## 1. 概要

本ドキュメントは、Pure GoでPDF生成・解析を行う `gopdf` ライブラリの全体アーキテクチャを定義する。

## 2. 設計原則

### 2.1. Pure Go
- CGO未使用、外部Cライブラリへの依存なし
- Go標準ライブラリとPure Goサードパーティライブラリのみ使用

### 2.2. Testability
- 各モジュールは独立してテスト可能
- インターフェースを活用した疎結合設計
- TDD（テスト駆動開発）を採用

### 2.3. Go Idioms
- エラーハンドリングはerrorを返す標準的な方法
- `io.Reader`/`io.Writer`インターフェースの活用
- goroutine-safeな設計（必要に応じてmutexを使用）

## 3. レイヤー構造

```
┌─────────────────────────────────────┐
│        API Layer (公開API)           │
├─────────────────────────────────────┤
│  Content Layer (描画・抽出機能)      │
│  - Text  - Image  - Shape            │
├─────────────────────────────────────┤
│  Document Layer (文書管理)           │
│  - Document  - Page  - Metadata      │
├─────────────────────────────────────┤
│  Writer Layer     │  Reader Layer    │
│  (生成・出力)      │  (解析・読込)     │
├──────────────────┼──────────────────┤
│  Font Layer (フォント管理)           │
│  - Standard  - TTF                   │
├─────────────────────────────────────┤
│  Core Layer (PDF基本オブジェクト)    │
│  - Object  - Stream  - Dictionary    │
└─────────────────────────────────────┘
```

## 4. モジュール設計

### 4.1. Core Layer (`internal/core`)

PDF仕様の低レベル要素を扱う。

**責務:**
- PDFオブジェクト（辞書、配列、ストリーム、名前、数値など）の表現
- オブジェクト参照とインダイレクトオブジェクトの管理
- クロスリファレンステーブル（xref）の構造

**主要型:**
```go
type Object interface{}
type Dictionary map[string]Object
type Array []Object
type Stream struct {
    Dict Dictionary
    Data []byte
}
type Reference struct {
    ObjectNumber   int
    GenerationNumber int
}
```

### 4.2. Font Layer (`internal/font`)

フォント管理と解析。

**責務:**
- 標準Type1フォントのサポート
- TTFフォントの解析（sfnt, glyf, cmapテーブルなど）
- フォントのサブセット化
- PDFへのフォント埋め込み

**主要型:**
```go
type Font interface {
    Name() string
    Encode(text string) ([]byte, error)
    GetWidth(text string) float64
}

type StandardFont struct {
    name string
    metrics *fontMetrics
}

type TTFFont struct {
    tables map[string][]byte
    // TTF解析データ
}
```

### 4.3. Writer Layer (`internal/writer`)

PDF生成・書き込み機能。

**責務:**
- PDFドキュメントのバイナリ出力
- オブジェクトのシリアライズ
- クロスリファレンステーブルの生成
- PDF構造の組み立て（header, body, xref, trailer）

**主要型:**
```go
type Writer struct {
    w io.Writer
    objOffset map[int]int64
    currentObjNum int
}

func NewWriter(w io.Writer) *Writer
func (w *Writer) WriteDocument(doc *Document) error
```

### 4.4. Reader Layer (`internal/reader`)

PDF解析・読み込み機能。

**責務:**
- PDFバイナリのパース
- クロスリファレンステーブルの解析
- オブジェクトの読み込みとデシリアライズ
- 暗号化PDFの基本的な復号

**主要型:**
```go
type Reader struct {
    r io.ReadSeeker
    xref *CrossReference
    trailer Dictionary
}

func NewReader(r io.ReadSeeker) (*Reader, error)
func (r *Reader) GetObject(ref Reference) (Object, error)
func (r *Reader) GetPageCount() (int, error)
```

### 4.5. Document Layer (`pkg/document`)

PDFドキュメントとページの管理。

**責務:**
- ドキュメントとページのライフサイクル管理
- ページツリー構造の管理
- メタデータ（Info辞書）の読み書き
- カスタムデータの埋め込み

**主要型:**
```go
type Document struct {
    pages []*Page
    metadata Metadata
    catalog *Catalog
    fonts map[string]Font
}

type Page struct {
    size PageSize
    orientation Orientation
    content *Content
    resources *Resources
}

type Metadata struct {
    Title    string
    Author   string
    Subject  string
    Keywords string
    Creator  string
}
```

### 4.6. Content Layer (`pkg/content`)

描画・抽出機能。

**責務:**
- テキスト描画（位置、フォント、色、変形）
- 画像描画（JPEG, PNG対応、拡大縮小、回転）
- 図形描画（線、矩形、円、ベジェ曲線）
- コンテンツストリームの生成・解析
- テキスト・画像・リンクの抽出

**主要型:**
```go
type Content struct {
    stream []byte
    operations []Operation
}

type TextDrawer interface {
    DrawText(text string, x, y float64, options TextOptions) error
}

type ImageDrawer interface {
    DrawImage(img image.Image, x, y, w, h float64) error
}

type ShapeDrawer interface {
    DrawLine(x1, y1, x2, y2 float64, options LineOptions) error
    DrawRect(x, y, w, h float64, options RectOptions) error
    DrawCircle(x, y, r float64, options CircleOptions) error
}

type TextExtractor interface {
    ExtractText(page *Page) ([]TextElement, error)
}

type ImageExtractor interface {
    ExtractImages(page *Page) ([]ImageElement, error)
}
```

### 4.7. API Layer (`pkg/gopdf` or root package)

ユーザー向けの高レベルAPI。

**責務:**
- シンプルで直感的なAPI提供
- Builder patternやFluent interfaceの活用
- 使いやすいヘルパー関数

**使用例:**
```go
// 新規作成
doc := gopdf.New()
page := doc.NewPage(gopdf.A4, gopdf.Portrait)
page.SetFont("Helvetica", 12)
page.DrawText("Hello, World!", 100, 100)
doc.WriteTo(file)

// 既存PDF読み込み
doc, err := gopdf.Open("existing.pdf")
text, err := doc.Page(0).ExtractText()
```

## 5. ディレクトリ構造

```
gopdf/
├── go.mod
├── go.sum
├── README.md
├── CLAUDE.md
├── docs/
│   ├── requirements.md          # 要件定義書
│   ├── architecture.md          # 本ドキュメント
│   ├── structure.md             # プロジェクト構造設計
│   └── progress.md              # 進捗管理
├── pkg/
│   └── gopdf/                   # 公開API
│       ├── document.go
│       ├── page.go
│       ├── options.go
│       └── api.go
├── internal/
│   ├── core/                    # PDF基本オブジェクト
│   │   ├── object.go
│   │   ├── dictionary.go
│   │   ├── stream.go
│   │   └── reference.go
│   ├── font/                    # フォント管理
│   │   ├── font.go
│   │   ├── standard.go
│   │   └── ttf.go
│   ├── writer/                  # PDF生成
│   │   ├── writer.go
│   │   └── serializer.go
│   ├── reader/                  # PDF解析
│   │   ├── reader.go
│   │   ├── parser.go
│   │   └── lexer.go
│   └── content/                 # コンテンツ処理
│       ├── drawer.go
│       ├── extractor.go
│       └── stream.go
├── examples/                    # サンプルコード
│   ├── create_simple/
│   ├── draw_shapes/
│   ├── embed_ttf/
│   └── extract_text/
└── testdata/                    # テストデータ
    └── pdfs/
```

## 6. 実装フェーズ

### Phase 1: 基礎構築（MVP）
- Core Layer: 基本オブジェクト
- Document Layer: Document, Page
- Writer Layer: 基本的な出力機能
- 簡単なPDF生成（空ページ、シンプルなテキスト）

### Phase 2: 描画機能拡充
- Content Layer: テキスト描画
- Font Layer: 標準フォント
- 図形描画（線、矩形）

### Phase 3: 画像・高度な描画
- 画像描画（JPEG, PNG）
- 円、ベジェ曲線
- テキスト変形（回転、傾斜）

### Phase 4: PDF解析（読み込み）
- Reader Layer: パーサー実装
- テキスト抽出
- メタデータ読み込み

### Phase 5: フォント拡張
- TTFフォント解析
- フォント埋め込み・サブセット化
- 日本語対応

### Phase 6: 高度な機能
- 既存PDFへの追記
- 画像・リンク抽出
- 暗号化対応

## 7. テスト戦略

### 7.1. ユニットテスト
- 各パッケージに対して `*_test.go` を配置
- テーブル駆動テストを活用
- カバレッジ目標: 80%以上

### 7.2. 統合テスト
- 実際のPDF生成・解析を確認
- Adobe Reader等での開封確認
- testdataに既存PDFサンプルを配置

### 7.3. ベンチマーク
- 大量ページ、高解像度画像での性能測定
- メモリプロファイリング

## 8. 依存ライブラリ候補

- 画像処理: `image`, `image/jpeg`, `image/png` (標準ライブラリ)
- 圧縮: `compress/zlib`, `compress/flate` (標準ライブラリ)
- TTF解析: `golang.org/x/image/font/sfnt` (公式拡張、Pure Go)
- 暗号化: `crypto/md5`, `crypto/rc4`, `crypto/aes` (標準ライブラリ)

すべてPure Goである必要がある。

## 9. 今後の拡張性

- PDF/A対応（アーカイブ用途）
- フォーム（AcroForm）の読み書き
- アノテーション（注釈）
- 電子署名
- より高度な暗号化（AES256など）
