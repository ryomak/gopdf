# パッケージ再構成 設計書

## 1. 現状の問題点

### ルートパッケージの肥大化
- 30以上のGoファイルがルートに存在
- 責務が混在（読み込み、書き込み、レイアウト、翻訳、マークダウン等）
- 新機能追加時にファイルが増え続ける

### 重複と冗長性
- `layout/` サブパッケージと `layout.go` の重複
- `internal/util`, `internal/utils`, `internal/testutil` の混在

## 2. 新しいパッケージ構成

```
gopdf/
├── gopdf.go              # ファサード: New(), Open() 等の主要API
├── types.go              # 公開型: PageSize, Color, StandardFont 等
├── errors.go             # エラー型
│
├── document/             # ドキュメント・ページ管理
│   ├── document.go       # Document 構造体
│   ├── page.go           # Page 構造体
│   ├── options.go        # ドキュメントオプション
│   └── encryption.go     # 暗号化設定
│
├── reader/               # PDF読み込み（公開API）
│   ├── reader.go         # PDFReader 構造体
│   ├── text.go           # テキスト抽出
│   ├── image.go          # 画像抽出
│   └── metadata.go       # メタデータ抽出
│
├── layout/               # レイアウト処理（既存を拡張）
│   ├── types.go          # PageLayout, TextBlock, ImageBlock
│   ├── extract.go        # レイアウト抽出
│   ├── render.go         # レイアウト描画
│   ├── adjustment.go     # レイアウト調整
│   └── strategies.go     # 調整戦略
│
├── text/                 # テキスト処理
│   ├── fitting.go        # テキストフィッティング
│   ├── sort.go           # テキストソート
│   ├── layer.go          # OCRテキストレイヤー
│   └── ruby.go           # ルビ注釈
│
├── font/                 # フォント（公開API）
│   ├── standard.go       # 標準14フォント
│   ├── ttf.go            # TTFフォント
│   └── default.go        # デフォルト日本語フォント
│
├── image/                # 画像処理（公開API）
│   ├── image.go          # Image 構造体
│   ├── jpeg.go           # JPEG処理
│   └── png.go            # PNG処理
│
├── translate/            # 翻訳機能
│   ├── translator.go     # Translator インターフェース
│   ├── options.go        # 翻訳オプション
│   └── render.go         # 翻訳結果描画
│
├── markdown/             # Markdown処理
│   ├── parser.go         # Markdownパーサー
│   └── renderer.go       # PDF描画
│
├── internal/             # 内部実装（非公開）
│   ├── core/             # PDFオブジェクトモデル
│   ├── parser/           # PDF解析（旧reader）
│   ├── writer/           # PDF書き込み
│   ├── content/          # コンテンツストリーム処理
│   ├── security/         # 暗号化
│   └── utils/            # ユーティリティ
│
├── examples/             # サンプル（既存）
├── cmd/                  # CLIツール（既存）
└── docs/                 # ドキュメント（既存）
```

## 3. 公開API設計

### ファサード（gopdf パッケージ）
```go
package gopdf

import (
    "github.com/ryomak/gopdf/document"
    "github.com/ryomak/gopdf/reader"
)

// ドキュメント作成
func New() *document.Document
func NewWithOptions(opts document.Options) *document.Document

// PDF読み込み
func Open(path string) (*reader.Reader, error)
func OpenReader(r io.ReadSeeker) (*reader.Reader, error)

// 型エイリアス（後方互換性）
type Document = document.Document
type Page = document.Page
type PDFReader = reader.Reader
```

### document パッケージ
```go
package document

type Document struct { ... }

func (d *Document) AddPage(size PageSize, orientation Orientation) *Page
func (d *Document) WriteTo(w io.Writer) error
func (d *Document) SetMetadata(m Metadata) error
func (d *Document) SetEncryption(opts EncryptionOptions) error

type Page struct { ... }

func (p *Page) SetFont(font StandardFont, size float64) error
func (p *Page) SetTTFFont(font *font.TTFFont, size float64) error
func (p *Page) DrawText(text string, x, y float64) error
func (p *Page) DrawImage(img *image.Image, x, y, w, h float64) error
func (p *Page) DrawLine(x1, y1, x2, y2 float64) error
// ... その他描画メソッド
```

### reader パッケージ
```go
package reader

type Reader struct { ... }

func (r *Reader) PageCount() int
func (r *Reader) ExtractPageText(pageNum int) (string, error)
func (r *Reader) ExtractPageLayout(pageNum int) (*layout.PageLayout, error)
func (r *Reader) ExtractImages(pageNum int) ([]image.Image, error)
func (r *Reader) Close() error
```

### layout パッケージ
```go
package layout

type PageLayout struct { ... }
type TextBlock struct { ... }
type ImageBlock struct { ... }

func (pl *PageLayout) SortedContentBlocks() []ContentBlock
func (pl *PageLayout) AdjustLayout(opts AdjustmentOptions) error
```

### translate パッケージ
```go
package translate

type Translator interface {
    Translate(text string) (string, error)
}

type Options struct {
    Translator Translator
    TargetFont interface{}
    // ...
}

func TranslatePDF(input, output string, opts Options) error
```

## 4. 移行計画

### Phase 1: サブパッケージ作成（既存コードと並存）
1. `document/`, `reader/`, `text/`, `translate/` 等を作成
2. 既存コードをコピー・整理
3. 新パッケージで動作確認

### Phase 2: ファサード作成
1. ルートの `gopdf.go` をファサードとして再構成
2. 型エイリアスで後方互換性を維持
3. Deprecatedアノテーションを追加

### Phase 3: 内部パッケージ整理
1. `internal/reader` → `internal/parser` にリネーム
2. `internal/util`, `utils`, `testutil` を統合
3. 不要なファイルを削除

### Phase 4: 旧コード削除
1. 非推奨APIの削除（メジャーバージョンアップ時）
2. 旧ファイルの削除

## 5. 後方互換性

### 型エイリアス
```go
// gopdf/deprecated.go
package gopdf

// Deprecated: Use document.Document instead
type Document = document.Document

// Deprecated: Use reader.Reader instead
type PDFReader = reader.Reader
```

### 関数ラッパー
```go
// gopdf/gopdf.go
package gopdf

// New creates a new Document
func New() *document.Document {
    return document.New()
}

// Open opens a PDF file
func Open(path string) (*reader.Reader, error) {
    return reader.Open(path)
}
```

## 6. 注意事項

- 各パッケージは循環参照を避ける
- `internal/` は外部から参照不可
- テストも新パッケージに移動
- examplesは新APIに更新

## 7. 実装状況

### 完了済み（Phase 1）

#### text/ パッケージ
- [x] `text/align.go` - テキスト配置（AlignLeft, AlignCenter, AlignRight）
- [x] `text/fitting.go` - テキストフィッティング（Fit, Wrap, FitOptions）
- [x] `text/fitting_test.go` - テスト

#### translate/ パッケージ
- [x] `translate/translator.go` - TranslatorインターフェースとFunc型
- [x] `translate/options.go` - 翻訳オプション
- [x] `translate/translator_test.go` - テスト

#### クリーンアップ
- [x] `test_coordinate_issue.go` 削除
- [x] `cmd/test_coordinate/` 削除
- [x] 空の `internal/util/`, `internal/testutil/` 確認済み（既に削除済み）

### 残りの作業

#### Phase 2: ルートパッケージ整理
- [ ] `layout.go` の型エイリアスと実装コードの分離
- [ ] `translator.go` を `translate/` パッケージへ移行
- [ ] `text_fitting.go` を `text/` パッケージへ移行
- [ ] ルートパッケージに型エイリアス追加

#### Phase 3: 追加パッケージ作成
- [ ] `document/` パッケージ作成
- [ ] `reader/` パッケージ作成（公開API）
- [ ] `font/` パッケージ作成
- [ ] `image/` パッケージ作成

#### Phase 4: 内部パッケージ整理
- [ ] `internal/reader` → `internal/parser` リネーム
- [ ] `internal/utils` 統合
