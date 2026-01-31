# Markdown変換機能設計書

## 1. 概要

Markdownファイルを入力として、PDFドキュメントやプレゼンテーションスライドを生成する機能を実装する。
CommonMark仕様とGitHub Flavored Markdown (GFM) に対応し、柔軟なスタイルカスタマイズを提供する。

## 2. 要件

### 2.1. 機能要件（requirements.md FR-7より）

#### FR-7.1: Markdown解析
- CommonMark仕様準拠
- GitHub Flavored Markdown (GFM) の拡張構文対応
  - テーブル、タスクリスト、取り消し線など
- 対応要素：
  - 見出し（H1-H6）
  - 段落
  - リスト（箇条書き、番号付き）
  - コードブロック
  - 引用
  - テーブル
  - 画像
  - リンク

#### FR-7.2: Markdown to PDF変換
- ドキュメント形式のPDF生成
- 見出しレベルに応じたスタイル自動適用
- コードブロックは等幅フォントで表示
- シンタックスハイライト（オプション）
- 画像の自動埋め込み
- リンクをPDFアノテーションとして保持

#### FR-7.3: Markdown to Slide変換
- プレゼンテーションスライド形式のPDF生成
- 水平線（`---`）またはH1見出しでスライド区切り
- 16:9または4:3のアスペクト比対応
- スライドタイトル、本文、箇条書き、コードブロック、画像の配置
- シンプルなテーマ（配色、フォント、レイアウト）

#### FR-7.4: スタイルカスタマイズ
- スタイル設定（フォント、色、余白、行間）
- テンプレートベースのレイアウト

#### FR-7.5: メタデータ対応
- フロントマター（YAML、TOML）からメタデータ読み取り
- PDF Infoへの反映

## 3. アーキテクチャ設計

### 3.1. パッケージ構成

```
gopdf/
├── markdown.go                    # 公開API
├── internal/
│   └── markdown/
│       ├── parser.go              # Markdown解析
│       ├── ast.go                 # 抽象構文木（AST）定義
│       ├── renderer.go            # レンダラーインターフェース
│       ├── pdf_renderer.go        # PDF変換レンダラー
│       ├── slide_renderer.go      # スライド変換レンダラー
│       ├── styles.go              # スタイル定義
│       ├── frontmatter.go         # フロントマター解析
│       └── syntax_highlight.go    # シンタックスハイライト（オプション）
└── examples/
    └── 16_markdown_conversion/
```

### 3.2. データフロー

```
Markdownファイル
    ↓
[Parser] → AST（抽象構文木）
    ↓
[Renderer] → gopdf Document
    ↓
PDFファイル
```

## 4. 詳細設計

### 4.1. 公開API (`markdown.go`)

```go
package gopdf

// MarkdownOptions はMarkdown変換のオプション
type MarkdownOptions struct {
    // Mode: "document" or "slide"
    Mode string

    // PageSize: スライドモードの場合のサイズ
    PageSize PageSize

    // Style: スタイル設定
    Style *MarkdownStyle

    // FontPath: カスタムフォントへのパス
    FontPath string

    // ImageBasePath: 画像ファイルのベースパス
    ImageBasePath string
}

// MarkdownStyle はMarkdownのスタイル設定
type MarkdownStyle struct {
    // H1-H6のフォントサイズ
    H1Size, H2Size, H3Size, H4Size, H5Size, H6Size float64

    // 本文のフォントサイズ
    BodySize float64

    // コードブロックのフォントサイズ
    CodeSize float64

    // 行間
    LineSpacing float64

    // 余白
    Margin struct {
        Top, Right, Bottom, Left float64
    }

    // 色設定
    TextColor      Color
    HeadingColor   Color
    CodeBackground Color
    LinkColor      Color
}

// NewMarkdownDocument はMarkdownからPDFドキュメントを生成
func NewMarkdownDocument(markdown string, opts *MarkdownOptions) (*Document, error)

// NewMarkdownDocumentFromFile はMarkdownファイルからPDFドキュメントを生成
func NewMarkdownDocumentFromFile(filepath string, opts *MarkdownOptions) (*Document, error)

// DefaultMarkdownStyle はデフォルトのMarkdownスタイルを返す
func DefaultMarkdownStyle() *MarkdownStyle

// DefaultSlideStyle はデフォルトのスライドスタイルを返す
func DefaultSlideStyle() *MarkdownStyle
```

### 4.2. Markdown Parser

#### 4.2.1. ライブラリの選定

**採用候補: `github.com/gomarkdown/markdown`**
- Pure Go実装
- CommonMark準拠
- GFM拡張サポート
- ASTベースの柔軟な処理
- アクティブにメンテナンスされている

```go
// internal/markdown/parser.go
package markdown

import (
    "github.com/gomarkdown/markdown"
    "github.com/gomarkdown/markdown/ast"
    "github.com/gomarkdown/markdown/parser"
)

// Parser はMarkdownパーサー
type Parser struct {
    parser *parser.Parser
}

// NewParser は新しいMarkdownパーサーを作成
func NewParser() *Parser {
    extensions := parser.CommonExtensions | parser.AutoHeadingIDs
    p := parser.NewWithExtensions(extensions)
    return &Parser{parser: p}
}

// Parse はMarkdownテキストを解析してASTを返す
func (p *Parser) Parse(markdown []byte) ast.Node {
    return p.parser.Parse(markdown)
}
```

#### 4.2.2. フロントマター解析

```go
// internal/markdown/frontmatter.go
package markdown

// Frontmatter はMarkdownのフロントマター
type Frontmatter struct {
    Title    string
    Author   string
    Date     string
    Keywords []string
    Custom   map[string]interface{}
}

// ParseFrontmatter はMarkdownからフロントマターを抽出
func ParseFrontmatter(markdown string) (*Frontmatter, string, error)
```

### 4.3. PDF Renderer

```go
// internal/markdown/pdf_renderer.go
package markdown

import (
    "github.com/gomarkdown/markdown/ast"
    "github.com/ryomak/gopdf"
)

// PDFRenderer はMarkdownをPDFドキュメントに変換
type PDFRenderer struct {
    doc           *gopdf.Document
    currentPage   *gopdf.Page
    style         *gopdf.MarkdownStyle
    currentY      float64
    imageBasePath string
}

// NewPDFRenderer は新しいPDFレンダラーを作成
func NewPDFRenderer(style *gopdf.MarkdownStyle, imageBasePath string) *PDFRenderer

// Render はASTをPDFドキュメントに変換
func (r *PDFRenderer) Render(root ast.Node) (*gopdf.Document, error)

// 各ノードタイプの処理メソッド
func (r *PDFRenderer) renderHeading(node *ast.Heading) error
func (r *PDFRenderer) renderParagraph(node *ast.Paragraph) error
func (r *PDFRenderer) renderList(node *ast.List) error
func (r *PDFRenderer) renderCodeBlock(node *ast.CodeBlock) error
func (r *PDFRenderer) renderTable(node *ast.Table) error
func (r *PDFRenderer) renderImage(node *ast.Image) error
func (r *PDFRenderer) renderLink(node *ast.Link) error
```

### 4.4. Slide Renderer

```go
// internal/markdown/slide_renderer.go
package markdown

// SlideRenderer はMarkdownをスライドPDFに変換
type SlideRenderer struct {
    doc           *gopdf.Document
    currentSlide  *gopdf.Page
    style         *gopdf.MarkdownStyle
    pageSize      gopdf.PageSize
    imageBasePath string
    slides        []ast.Node // スライドごとのASTノード
}

// NewSlideRenderer は新しいスライドレンダラーを作成
func NewSlideRenderer(
    pageSize gopdf.PageSize,
    style *gopdf.MarkdownStyle,
    imageBasePath string,
) *SlideRenderer

// Render はASTをスライドPDFに変換
func (r *SlideRenderer) Render(root ast.Node) (*gopdf.Document, error)

// splitSlides はASTをスライドごとに分割
func (r *SlideRenderer) splitSlides(root ast.Node) []ast.Node

// renderSlide は1つのスライドをレンダリング
func (r *SlideRenderer) renderSlide(slideNode ast.Node) error

// スライド要素の配置メソッド
func (r *SlideRenderer) renderSlideTitle(text string) error
func (r *SlideRenderer) renderSlideBody(nodes []ast.Node) error
func (r *SlideRenderer) renderSlideBullets(items []ast.Node) error
```

## 5. 実装フェーズ

### Phase 1: 基礎実装
1. プレゼンテーションページサイズの追加
2. 基本的なMarkdownパーサーの統合
3. シンプルなPDFレンダラーの実装（見出し、段落のみ）

### Phase 2: ドキュメント変換
1. リスト、コードブロック、引用の対応
2. 画像埋め込み
3. リンクアノテーション
4. スタイルカスタマイズ

### Phase 3: スライド変換
1. スライド分割ロジック
2. スライドレイアウトエンジン
3. 16:9/4:3対応
4. テーマシステム

### Phase 4: 拡張機能
1. テーブル描画
2. シンタックスハイライト（オプション）
3. フロントマター対応
4. テンプレートシステム

## 6. テスト計画

### 6.1. ユニットテスト

- Markdown解析のテスト
- 各ノードタイプのレンダリングテスト
- スタイル適用のテスト
- フロントマター解析のテスト

### 6.2. 統合テスト

- サンプルMarkdownファイルからPDF生成
- 複雑なMarkdownドキュメントの変換
- スライド生成の検証
- PDF Readerでの表示確認

### 6.3. サンプルMarkdown

```markdown
---
title: "gopdfで生成したPDF"
author: "gopdf"
date: "2024-10-28"
---

# 見出し1

これは段落です。**太字**や*斜体*もサポートします。

## 見出し2

- 箇条書き1
- 箇条書き2
- 箇条書き3

### コードブロック

\`\`\`go
func main() {
    fmt.Println("Hello, World!")
}
\`\`\`

---

# 新しいスライド（スライドモード）

スライドの内容がここに入ります。
```

## 7. 使用例

### 7.1. ドキュメント変換

```go
// Markdownファイルを読み込んでPDF生成
doc, err := gopdf.NewMarkdownDocumentFromFile(
    "document.md",
    &gopdf.MarkdownOptions{
        Mode:          "document",
        PageSize:      gopdf.PageSizeA4,
        Style:         gopdf.DefaultMarkdownStyle(),
        ImageBasePath: "./images",
    },
)
if err != nil {
    log.Fatal(err)
}

f, _ := os.Create("output.pdf")
doc.WriteTo(f)
f.Close()
```

### 7.2. スライド変換

```go
// Markdownをプレゼンテーションスライドに変換
doc, err := gopdf.NewMarkdownDocumentFromFile(
    "slides.md",
    &gopdf.MarkdownOptions{
        Mode:          "slide",
        PageSize:      gopdf.PageSizePresentation16x9,
        Style:         gopdf.DefaultSlideStyle(),
        ImageBasePath: "./images",
    },
)
if err != nil {
    log.Fatal(err)
}

f, _ := os.Create("slides.pdf")
doc.WriteTo(f)
f.Close()
```

### 7.3. カスタムスタイル

```go
style := &gopdf.MarkdownStyle{
    H1Size:      48,
    H2Size:      36,
    H3Size:      24,
    BodySize:    12,
    CodeSize:    10,
    LineSpacing: 1.5,
    TextColor:   gopdf.RGB(0, 0, 0),
    HeadingColor: gopdf.RGB(0.2, 0.2, 0.8),
}

doc, err := gopdf.NewMarkdownDocument(markdownText, &gopdf.MarkdownOptions{
    Mode:  "document",
    Style: style,
})
```

## 8. 依存ライブラリ

### 8.1. 必須

- `github.com/gomarkdown/markdown`: Markdown解析（Pure Go）
  - License: BSD 2-Clause
  - 最新の安定版を使用

### 8.2. オプション

- シンタックスハイライト用のライブラリ（将来的に検討）
- YAML/TOMLパーサー（フロントマター用、標準ライブラリまたはPure Goのみ）

## 9. パフォーマンス考慮事項

- 大きなMarkdownファイルの処理
- 画像の最適化
- ページ遷移時のメモリ管理
- ストリーミング処理の活用

## 10. 制限事項

### 10.1. 初期実装での制限

- シンタックスハイライトは基本的なもののみ（または未対応）
- 複雑なテーブルレイアウトは簡略化
- 数式（LaTeX）は未対応
- アニメーション効果は未対応

### 10.2. 将来の拡張

- Mermaid図の対応
- LaTeX数式のレンダリング
- より高度なテーマシステム
- インタラクティブな要素（JavaScriptアクションなど）

## 11. 参考資料

- CommonMark Spec: https://commonmark.org/
- GitHub Flavored Markdown Spec: https://github.github.com/gfm/
- gomarkdown: https://github.com/gomarkdown/markdown
- Marp (Markdown Presentation): https://marp.app/

## 12. マイルストーン

### Phase 1（最小機能）
- [x] 設計書作成
- [ ] プレゼンテーションページサイズ実装
- [ ] gomarkdownライブラリの統合
- [ ] 基本的なドキュメント変換（見出し、段落）
- [ ] テストとサンプル

### Phase 2（ドキュメント変換）
- [ ] リスト、コードブロック、引用
- [ ] 画像とリンク
- [ ] スタイルカスタマイズ
- [ ] フロントマター対応

### Phase 3（スライド変換）
- [ ] スライド分割
- [ ] スライドレイアウト
- [ ] テーマシステム

### Phase 4（拡張）
- [ ] テーブル
- [ ] シンタックスハイライト
- [ ] 高度なスタイリング
