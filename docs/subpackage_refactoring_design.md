# gopdf サブパッケージ分割設計書

## 1. 目的

gopdfパッケージをサブパッケージに分割し、以下を実現する：
1. **見通しの改善**: 機能ごとに独立したパッケージにする
2. **責務の明確化**: 各ファイルの行数を適切に保つ（目安: 300行以下）
3. **利用の明確化**: import pathで機能が明確になる

## 2. 現状の問題点

### 2.1. ルートパッケージのファイル数
- **41ファイル**が混在（コア機能 + Reader + Markdown + Layout + OCR等）
- 機能が見つけにくい

### 2.2. 大きすぎるファイル
| ファイル | 行数 | 状態 |
|---------|------|------|
| layout.go | 849行 | 要分割 ⚠️ |
| internal/reader/reader.go | 767行 | 要分割 ⚠️ |
| page.go | 652行 | 分割検討 |
| metadata.go | 450行 | 分割検討 |
| internal/content/extractor.go | 431行 | 分割検討 |
| internal/reader/lexer.go | 429行 | OK（字句解析は1ファイルが自然） |

## 3. リファクタリング方針

### 3.1. サブパッケージ構成

```
gopdf/
├── document.go          # コア: Document型
├── page.go              # コア: Page型（必要なら分割）
├── font.go              # コア: Font型
├── graphics.go          # コア: Color等
├── constants.go         # コア: PageSize等
├── image.go             # コア: Image型
├── encryption.go        # コア: Encryption型
├── metadata.go          # コア: Metadata型（必要なら分割）
├── ruby.go              # コア: RubyText型
├── deprecated.go        # コア: 非推奨関数
│
├── reader/              # 新規サブパッケージ: PDF読み込み
│   ├── reader.go        # PDFReader型と基本メソッド
│   ├── text.go          # テキスト抽出
│   ├── image.go         # 画像抽出
│   ├── layout.go        # レイアウト抽出
│   ├── metadata.go      # メタデータ抽出
│   └── encryption.go    # 暗号化情報
│
├── markdown/            # 新規サブパッケージ: Markdown変換
│   ├── markdown.go      # Markdown → PDF変換
│   ├── renderer.go      # レンダリングロジック
│   └── styles.go        # スタイル定義
│
├── layout/              # 新規サブパッケージ: レイアウト調整
│   ├── layout.go        # PageLayout型と基本機能
│   ├── blocks.go        # TextBlock, ImageBlock
│   ├── adjustment.go    # レイアウト調整ロジック
│   ├── strategy.go      # LayoutStrategy実装
│   ├── overlap.go       # オーバーラップ検出
│   └── fitting.go       # テキストフィッティング
│
├── ocr/                 # 新規サブパッケージ: OCR/TextLayer
│   ├── text_layer.go    # TextLayer型
│   └── ocr.go           # OCRResult型
│
├── translator/          # 新規サブパッケージ: PDF翻訳
│   ├── translator.go    # Translator インターフェース
│   └── renderer.go      # レイアウトレンダリング
│
└── internal/
    ├── core/
    ├── font/
    ├── image/
    ├── writer/
    ├── security/
    ├── markdown/
    ├── utils/
    ├── reader/          # 内部ファイルも分割
    │   ├── reader.go            # Reader型と基本機能（300行以下）
    │   ├── reader_catalog.go    # カタログ関連
    │   ├── reader_pages.go      # ページツリー関連
    │   ├── reader_resources.go  # リソース関連
    │   ├── lexer.go             # 字句解析（429行 → そのまま）
    │   ├── parser.go            # 構文解析（325行 → そのまま）
    │   └── encryption.go        # 暗号化（168行 → そのまま）
    └── content/         # 内部ファイルも分割
        ├── extractor.go             # テキスト抽出コア（200行程度に削減）
        ├── extractor_text.go        # テキスト抽出詳細
        ├── extractor_position.go    # 位置計算
        ├── image_extractor.go       # 画像抽出（234行 → そのまま）
        ├── stream_parser.go         # ストリーム解析（144行 → そのまま）
        ├── graphics_state.go        # グラフィック状態（97行 → そのまま）
        ├── font_info.go             # フォント情報（159行 → そのまま）
        └── tounicode.go             # Unicode変換（293行 → そのまま）
```

### 3.2. 破壊的変更の内容

**Before (現在):**
```go
import "github.com/ryomak/gopdf"

reader, err := gopdf.Open("test.pdf")
doc, err := gopdf.NewMarkdownDocument(md, nil)
layout, err := reader.ExtractPageLayout(1)
```

**After (リファクタリング後):**
```go
import (
    "github.com/ryomak/gopdf"
    "github.com/ryomak/gopdf/reader"
    "github.com/ryomak/gopdf/markdown"
    "github.com/ryomak/gopdf/layout"
)

r, err := reader.Open("test.pdf")
doc, err := markdown.NewDocument(md, nil)
l, err := r.ExtractPageLayout(1) // returns *layout.PageLayout
```

## 4. 詳細設計

### 4.1. reader/ パッケージ

**責務**: PDF読み込みとデータ抽出

**公開型:**
```go
package reader

type PDFReader struct {
    impl *internal_reader.Reader
    closer io.Closer
}

type TextElement struct {
    Text   string
    X, Y   float64
    Width, Height float64
    Font   string
    Size   float64
}

type ImageInfo struct {
    Name   string
    Width, Height int
    Data   []byte
    Format ImageFormat
}

type EncryptionInfo struct {
    Filter string
    V, R, Length int
    P int32
    IsOwner bool
}
```

**公開関数:**
```go
func Open(path string) (*PDFReader, error)
func OpenReader(r io.ReadSeeker) (*PDFReader, error)

func (r *PDFReader) Close() error
func (r *PDFReader) PageCount() int
func (r *PDFReader) Info() gopdf.Metadata

// text.go
func (r *PDFReader) ExtractPageText(pageNum int) (string, error)
func (r *PDFReader) ExtractText() ([]string, error)
func (r *PDFReader) ExtractTextElements(pageNum int) ([]TextElement, error)

// image.go
func (r *PDFReader) ExtractImages(pageNum int) ([]ImageInfo, error)

// layout.go
func (r *PDFReader) ExtractPageLayout(pageNum int) (*layout.PageLayout, error)
func (r *PDFReader) ExtractAllLayouts() ([]*layout.PageLayout, error)

// encryption.go
func (r *PDFReader) GetEncryptionInfo() (*EncryptionInfo, error)
```

**ファイル分割:**
- `reader.go` (100行): PDFReader型、Open関数
- `text.go` (100行): テキスト抽出メソッド
- `image.go` (50行): 画像抽出メソッド
- `layout.go` (50行): レイアウト抽出メソッド
- `metadata.go` (30行): メタデータ抽出
- `encryption.go` (30行): 暗号化情報

**合計**: 約360行 → 6ファイルに分割

---

### 4.2. markdown/ パッケージ

**責務**: Markdown → PDF変換

**公開型:**
```go
package markdown

type Options struct {
    PageSize     gopdf.PageSize
    Margins      Margins
    DefaultFont  interface{}
    Style        *Style
    Mode         Mode
}

type Style struct {
    H1Size, H2Size, H3Size float64
    // ... その他のスタイル設定
}

type Mode string
const (
    ModeDocument Mode = "document"
    ModeSlide    Mode = "slide"
)
```

**公開関数:**
```go
func NewDocument(markdownText string, opts *Options) (*gopdf.Document, error)
func NewDocumentFromFile(filepath string, opts *Options) (*gopdf.Document, error)
func DefaultStyle() *Style
func SlideStyle() *Style
```

**ファイル分割:**
- `markdown.go` (150行): 変換メインロジック
- `renderer.go` (150行): レンダリング詳細
- `styles.go` (50行): スタイル定義

**合計**: 約350行 → 3ファイルに分割

---

### 4.3. layout/ パッケージ

**責務**: PDFレイアウトの表現と調整

**公開型:**
```go
package layout

type PageLayout struct {
    PageNum    int
    Width, Height float64
    TextBlocks []TextBlock
    Images     []ImageBlock
}

type TextBlock struct {
    Text     string
    Elements []TextElement
    Rect     Rectangle
    Font     string
    FontSize float64
    Color    gopdf.Color
}

type ImageBlock struct {
    ImageInfo reader.ImageInfo
    X, Y, PlacedWidth, PlacedHeight float64
}

type Rectangle struct {
    X, Y, Width, Height float64
}

type ContentBlock interface {
    Bounds() Rectangle
    Type() ContentBlockType
    Position() (x, y float64)
}

type Strategy string
const (
    StrategyFlowDown    Strategy = "flow_down"
    StrategyFitContent  Strategy = "fit_content"
)

type AdjustmentOptions struct {
    Strategy     Strategy
    MinFontSize  float64
    MaxFontSize  float64
    LineSpacing  float64
}
```

**公開関数:**
```go
func DefaultAdjustmentOptions() AdjustmentOptions
func AdjustLayout(layout *PageLayout, opts AdjustmentOptions) error
func DetectOverlaps(layout *PageLayout) []BlockOverlap
```

**ファイル分割:**
- `layout.go` (150行): PageLayout型と基本機能
- `blocks.go` (150行): TextBlock, ImageBlock実装
- `adjustment.go` (200行): レイアウト調整ロジック
- `strategy.go` (150行): Strategy実装
- `overlap.go` (100行): オーバーラップ検出
- `fitting.go` (100行): テキストフィッティング

**合計**: 849行 → 6ファイルに分割

---

### 4.4. ocr/ パッケージ

**責務**: OCRとテキストレイヤー

**公開型:**
```go
package ocr

type TextLayer struct {
    Words      []TextLayerWord
    RenderMode TextRenderMode
    Opacity    float64
}

type TextLayerWord struct {
    Text   string
    Bounds layout.Rectangle
}

type OCRResult struct {
    Text  string
    Words []OCRWord
}

type OCRWord struct {
    Text       string
    Confidence float64
    Bounds     layout.Rectangle
}

type TextRenderMode int
const (
    RenderNormal TextRenderMode = iota
    RenderInvisible
)
```

**公開関数:**
```go
func DefaultTextLayer() TextLayer
func NewTextLayer(words []TextLayerWord) TextLayer
func ConvertPixelToPDFCoords(...) (float64, float64)
func ConvertPixelToPDFRect(...) layout.Rectangle
func (r OCRResult) ToTextLayer(...) TextLayer
```

**ファイル分割:**
- `text_layer.go` (80行): TextLayer型
- `ocr.go` (80行): OCRResult型

**合計**: 約160行 → 2ファイルに分割

---

### 4.5. translator/ パッケージ

**責務**: PDF翻訳とレイアウトレンダリング

**公開型:**
```go
package translator

type Translator interface {
    Translate(text string) (string, error)
}

type TranslateFunc func(string) (string, error)

func (f TranslateFunc) Translate(text string) (string, error) {
    return f(text)
}

type Options struct {
    Translator     Translator
    TargetFont     interface{}
    FontName       string
    PreserveLayout bool
    MaxFontSize    float64
    MinFontSize    float64
}
```

**公開関数:**
```go
func TranslatePDF(inputPath, outputPath string, opts Options) error
func TranslatePDFToWriter(input io.ReadSeeker, output io.Writer, opts Options) error
func RenderLayout(doc *gopdf.Document, layout *layout.PageLayout, opts Options) (*gopdf.Page, error)
func TranslateTextBlocks(blocks []layout.TextBlock, translator Translator) error
func DefaultOptions(targetFont interface{}, fontName string) Options
```

**ファイル分割:**
- `translator.go` (150行): 翻訳メインロジック
- `renderer.go` (100行): レイアウトレンダリング

**合計**: 約250行 → 2ファイルに分割

---

### 4.6. internal/reader/ の分割

**現状**: reader.go (767行) が大きすぎる

**分割案:**
```go
// reader.go (200行) - Reader型とコア機能
type Reader struct { ... }
func NewReader(r io.ReadSeeker) (*Reader, error)
func (r *Reader) GetObject(ref core.Reference) (core.Object, error)

// reader_catalog.go (150行) - カタログ関連
func (r *Reader) getCatalog() (core.Dictionary, error)
func (r *Reader) getPages() (core.Dictionary, error)

// reader_pages.go (200行) - ページツリー関連
func (r *Reader) PageCount() int
func (r *Reader) getPageNode(pageNum int) (core.Dictionary, error)
func (r *Reader) getAllPageNodes() ([]core.Dictionary, error)

// reader_resources.go (150行) - リソース関連
func (r *Reader) getResources(pageDict core.Dictionary) (core.Dictionary, error)
func (r *Reader) getFontInfo(resources core.Dictionary) map[string]*content.FontInfo
func (r *Reader) getImages(resources core.Dictionary) (map[string]core.Stream, error)

// reader_content.go (100行) - コンテンツストリーム
func (r *Reader) getContentStream(pageDict core.Dictionary) ([]byte, error)
```

**合計**: 767行 → 5ファイルに分割

---

### 4.7. internal/content/ の分割

**現状**: extractor.go (431行) が大きい

**分割案:**
```go
// extractor.go (150行) - Extractor型とコア機能
type Extractor struct { ... }
func NewExtractor(contentStream []byte, resources core.Dictionary, ...) *Extractor
func (e *Extractor) Extract() ([]Element, error)

// extractor_text.go (150行) - テキスト抽出詳細
func (e *Extractor) handleTextOperator(op string, args []core.Object)
func (e *Extractor) processTextObject()

// extractor_position.go (130行) - 位置計算
func (e *Extractor) calculateTextPosition() (float64, float64)
func (e *Extractor) applyTextMatrix()
func (e *Extractor) transformCoordinates(x, y float64) (float64, float64)
```

**合計**: 431行 → 3ファイルに分割

---

## 5. マイグレーション手順

### Phase 1: 新規サブパッケージ作成（破壊的変更なし）
1. `reader/` パッケージを作成
   - 現在の `reader.go` の型をコピー
   - 新しいimport pathで動作確認
2. `markdown/` パッケージを作成
3. `layout/` パッケージを作成
4. `ocr/` パッケージを作成
5. `translator/` パッケージを作成

**この時点では**: ルートパッケージとサブパッケージの両方が存在

### Phase 2: ルートパッケージに非推奨マーカー追加
```go
// reader.go (root package)

// Deprecated: Use github.com/ryomak/gopdf/reader.Open instead.
func Open(path string) (*PDFReader, error) {
    return reader.Open(path)
}
```

### Phase 3: examples/ とテストの更新
- すべてのexamplesを新しいimport pathに変更
- テストを更新
- `make ci` で確認

### Phase 4: ルートパッケージのファイル削除
- deprecated.go に統合してから削除
- または直接削除

### Phase 5: ドキュメント更新
- README.md更新
- CHANGELOG.md に破壊的変更を記載
- マイグレーションガイド作成

---

## 6. 実装順序

1. **Phase 1-A**: layout/パッケージ作成（最優先、849行の分割）
   - layout.go → layout/layout.go, layout/blocks.go, layout/adjustment.go等
   - テスト移動
   - `make ci` 実行

2. **Phase 1-B**: reader/パッケージ作成
   - reader.go → reader/reader.go, reader/text.go等
   - テスト移動
   - `make ci` 実行

3. **Phase 1-C**: markdown/パッケージ作成
   - markdown.go, markdown_renderer.go → markdown/
   - テスト移動
   - `make ci` 実行

4. **Phase 1-D**: ocr/パッケージ作成
   - text_layer.go → ocr/
   - テスト移動
   - `make ci` 実行

5. **Phase 1-E**: translator/パッケージ作成
   - translator.go → translator/
   - テスト移動
   - `make ci` 実行

6. **Phase 2**: internal/の大きいファイル分割
   - internal/reader/reader.go → 5ファイルに分割
   - internal/content/extractor.go → 3ファイルに分割
   - テスト確認
   - `make ci` 実行

7. **Phase 3**: examplesとテスト更新
   - すべてのexamplesを新しいimport pathに変更
   - テストのimport更新
   - `make ci` 実行

8. **Phase 4**: ルートパッケージのクリーンアップ
   - 古いファイルを削除
   - deprecated.goに移行関数を追加（オプション）
   - `make ci` 実行

9. **Phase 5**: ドキュメント更新とpush
   - README.md, CHANGELOG.md更新
   - `make ci` 最終確認
   - git commit & push

---

## 7. テスト戦略

### 7.1. 各フェーズでのテスト
```bash
# 各サブパッケージ作成後
go test ./reader/...
go test ./markdown/...
go test ./layout/...

# 全体テスト
make ci
```

### 7.2. 破壊的変更の検証
- すべてのexamplesが新しいimportで動作することを確認
- 既存のテストを新しいimportに移行
- カバレッジが維持されることを確認

---

## 8. 期待される効果

### 8.1. ファイル数の削減（ルートパッケージ）
- **Before**: 41ファイル
- **After**: 約15ファイル（コア機能のみ）

### 8.2. 最大ファイルサイズの削減
- **Before**: 849行 (layout.go)
- **After**: 最大300行程度

### 8.3. importの明確化
```go
// 何をしているか一目瞭然
import (
    "github.com/ryomak/gopdf"
    "github.com/ryomak/gopdf/reader"
    "github.com/ryomak/gopdf/markdown"
)
```

### 8.4. godocの改善
- 各サブパッケージで独立したドキュメント
- 関係ない機能が混在しない

---

## 9. まとめ

このリファクタリングにより：
1. **見通しの改善**: 機能ごとに独立したパッケージ
2. **責務の明確化**: 各ファイル300行以下を目標
3. **保守性の向上**: 内部ファイルも適切に分割

破壊的変更ではあるが、より良いパッケージ構造になる。
段階的に実装し、各フェーズで`make ci`を実行して品質を保証する。
