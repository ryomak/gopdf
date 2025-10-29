# gopdf パッケージリファクタリング設計書

## 1. 目的

gopdfパッケージの公開APIを整理し、ユーザーが使用すべき型・メソッドのみを露出させる。
内部実装の詳細は`internal`パッケージに隠蔽し、見通しの良いAPIを提供する。

## 2. 現状の問題点

### 2.1. 内部型の露出

**問題:**
- `font.StandardFont`型がPageの公開フィールドに露出
- `font.TTFFont`型がTTFFont構造体に露出
- `reader.Reader`型がPDFReader構造体に露出

**影響:**
- ユーザーが内部実装に依存する可能性
- 内部実装の変更が破壊的変更になる
- APIの見通しが悪い

### 2.2. ヘルパーメソッドの公開

**問題:**
以下のメソッドが公開されているが、ユーザーは直接使用しない:
- `Page.drawTextInternal()`
- `Page.getFontKey()`
- `Page.getTTFFontKey()`
- `Page.escapeString()`
- `Page.textToHexString()`
- `Page.textToGlyphIndices()`
- `Page.getCurrentFontName()`

**影響:**
- APIが肥大化
- ユーザーが混乱する
- 内部実装の変更が困難

### 2.3. 構造体フィールドの露出

**問題:**
Pageの内部フィールドが露出:
```go
type Page struct {
    content        bytes.Buffer        // 内部バッファ
    currentFont    *font.StandardFont  // 内部状態
    currentTTFFont *TTFFont            // 内部状態
    fonts          map[string]font.StandardFont // 内部マップ
}
```

**影響:**
- カプセル化の破壊
- 内部状態の整合性が保証できない

## 3. リファクタリング方針

### 3.1. 基本原則

1. **最小公開APIの原則**
   - ユーザーが直接使用する型・メソッドのみをroot packageで公開
   - 内部実装はすべて`internal`に配置

2. **DTOパターンの採用**
   - 内部型を直接露出せず、公開用のDTO（Data Transfer Object）を提供
   - 内部型とDTOの変換ロジックをinternal内で実装

3. **循環参照の回避**
   - 構造体は内部で定義し、公開APIではインターフェースまたはDTOを使用
   - 依存関係を一方向に保つ

### 3.2. パッケージ構成

```
gopdf/
├── document.go          # 公開API: Document型とメソッド
├── page.go              # 公開API: Page型とメソッド
├── font.go              # 公開API: StandardFont定数とTTFFont型
├── reader.go            # 公開API: PDFReader型とメソッド
├── graphics.go          # 公開API: Color, Line等の型
├── constants.go         # 公開API: PageSize等の定数
├── image.go             # 公開API: Image型
├── metadata.go          # 公開API: Metadata型
├── layout.go            # 公開API: Layout関連の型
├── ruby.go              # 公開API: RubyText関連の型
├── text_layer.go        # 公開API: TextLayer関連の型
├── markdown.go          # 公開API: Markdown関連の型
├── translator.go        # 公開API: Translation関連の型
├── encryption.go        # 公開API: Encryption関連の型
│
└── internal/
    ├── core/            # PDF基本オブジェクト
    ├── font/            # フォント内部実装
    ├── image/           # 画像内部実装
    ├── reader/          # PDFリーダー内部実装
    ├── writer/          # PDFライター内部実装
    ├── content/         # コンテンツストリーム処理
    ├── security/        # セキュリティ・暗号化
    ├── markdown/        # Markdown解析
    ├── utils/           # ユーティリティ
    │
    └── pageimpl/        # 新規: Pageの内部実装
        ├── page_impl.go         # Page構造体の内部実装
        ├── text_drawing.go      # テキスト描画の内部ロジック
        ├── graphics_drawing.go  # 図形描画の内部ロジック
        ├── image_drawing.go     # 画像描画の内部ロジック
        └── helpers.go           # 内部ヘルパー関数
```

## 4. リファクタリング詳細

### 4.1. Pageの内部実装分離

**Before (現状):**
```go
// page.go (root package)
type Page struct {
    width          float64
    height         float64
    content        bytes.Buffer        // 露出
    currentFont    *font.StandardFont  // 内部型露出
    currentTTFFont *TTFFont
    fonts          map[string]font.StandardFont // 内部型露出
    // ... その他のフィールド
}

// ヘルパーメソッドが公開されている
func (p *Page) drawTextInternal(...) { }
func (p *Page) getFontKey(f font.StandardFont) string { }
```

**After (リファクタリング後):**
```go
// page.go (root package) - 公開API
type Page struct {
    impl *pageimpl.PageImpl  // 内部実装への参照（非公開）
}

// 公開メソッドのみ定義
func (p *Page) Width() float64 { return p.impl.Width() }
func (p *Page) Height() float64 { return p.impl.Height() }
func (p *Page) SetFont(f StandardFont, size float64) error { return p.impl.SetFont(f, size) }
func (p *Page) DrawText(text string, x, y float64) error { return p.impl.DrawText(text, x, y) }
// ... その他の公開メソッド

// internal/pageimpl/page_impl.go - 内部実装
package pageimpl

import (
    "bytes"
    "github.com/ryomak/gopdf/internal/font"
)

type PageImpl struct {
    width          float64
    height         float64
    content        bytes.Buffer
    currentFont    *font.StandardFont
    currentTTFFont *font.TTFFont
    fonts          map[string]*font.StandardFont
    // ... その他のフィールド
}

// 内部ヘルパーメソッド（非公開）
func (p *PageImpl) drawTextInternal(...) { }
func (p *PageImpl) getFontKey(f *font.StandardFont) string { }
```

**メリット:**
- 公開APIがシンプルになる
- 内部実装を自由に変更できる
- ユーザーが混乱するメソッドが見えなくなる

### 4.2. フォント型の分離

**Before (現状):**
```go
// ttf_font.go (root package)
type TTFFont struct {
    internal *font.TTFFont  // 内部型が露出
}
```

**After (リファクタリング後):**
```go
// font.go (root package) - 公開API
type StandardFont string  // そのまま

const (
    FontHelvetica StandardFont = "Helvetica"
    // ... その他のフォント定数
)

type TTFFont struct {
    impl *font.TTFFont  // 非公開フィールド
}

func LoadTTF(path string) (*TTFFont, error) {
    impl, err := font.LoadTTF(path)
    if err != nil {
        return nil, err
    }
    return &TTFFont{impl: impl}, nil
}

func (f *TTFFont) Name() string {
    return f.impl.Name()
}

func (f *TTFFont) TextWidth(text string, fontSize float64) (float64, error) {
    return f.impl.TextWidth(text, fontSize)
}

// 内部からの変換関数
func (f *TTFFont) Internal() *font.TTFFont {
    return f.impl
}
```

**メリット:**
- 内部型が完全に隠蔽される
- TTFFontは薄いラッパーとして機能
- 内部実装の変更がユーザーに影響しない

### 4.3. PDFReaderの分離

**Before (現状):**
```go
// reader.go (root package)
type PDFReader struct {
    r      *reader.Reader  // 内部型が露出
    closer io.Closer
}
```

**After (リファクタリング後):**
```go
// reader.go (root package) - 公開API
type PDFReader struct {
    impl   *reader.Reader  // 非公開フィールド
    closer io.Closer
}

func Open(path string) (*PDFReader, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    impl, err := reader.NewReader(f)
    if err != nil {
        f.Close()
        return nil, err
    }
    return &PDFReader{impl: impl, closer: f}, nil
}

func (r *PDFReader) PageCount() int {
    return r.impl.PageCount()
}

// ... その他の公開メソッド

// 内部からの変換関数（internal/内でのみ使用）
func (r *PDFReader) Internal() *reader.Reader {
    return r.impl
}
```

**メリット:**
- reader.Reader型が完全に隠蔽される
- PDFReaderは薄いラッパーとして機能

### 4.4. 循環参照の回避

**問題のケース:**
```go
// internal/pageimpl/page_impl.go
import "github.com/ryomak/gopdf" // NGな循環参照

type PageImpl struct {
    // gopdf.Colorを使いたいが循環参照になる
}
```

**解決策1: 共通の型をinternal/typesに定義**
```go
// internal/types/types.go - 共通型の定義
package types

type Color struct {
    R, G, B float64
}

type Rectangle struct {
    X, Y, Width, Height float64
}

// ... その他の共通型

// graphics.go (root package) - 公開API
package gopdf

import "github.com/ryomak/gopdf/internal/types"

type Color = types.Color  // エイリアス

func NewRGB(r, g, b uint8) Color {
    return Color{
        R: float64(r) / 255.0,
        G: float64(g) / 255.0,
        B: float64(b) / 255.0,
    }
}

// internal/pageimpl/page_impl.go
package pageimpl

import "github.com/ryomak/gopdf/internal/types"

type PageImpl struct {
    strokeColor types.Color  // 循環参照なし
    fillColor   types.Color
}
```

**解決策2: DTOパターン**
```go
// internal/pageimpl/page_impl.go
package pageimpl

// 内部で使う型を定義
type internalColor struct {
    R, G, B float64
}

type PageImpl struct {
    strokeColor internalColor
}

// 公開API型から内部型への変換
func (p *PageImpl) SetStrokeColor(c gopdf.Color) {
    p.strokeColor = internalColor{R: c.R, G: c.G, B: c.B}
}
```

**推奨アプローチ:**
- 単純な値型（Color, Rectangle等）は `internal/types` で定義し、root packageでエイリアスとして公開
- 複雑な型やロジックを持つものはDTOパターンで変換

## 5. 実装手順

### Phase 1: 型の整理と移動準備
1. `internal/types` パッケージを作成
2. 共通の値型（Color, Rectangle等）を移動
3. root packageでエイリアスを定義

### Phase 2: Pageの内部実装分離
1. `internal/pageimpl` パッケージを作成
2. Pageの内部実装を`PageImpl`に移動
3. ヘルパーメソッドをすべて`PageImpl`に移動
4. root packageのPageを薄いラッパーに変更

### Phase 3: フォント型の分離
1. TTFFontの内部フィールドを非公開化
2. 変換関数を内部用に追加

### Phase 4: PDFReaderの分離
1. PDFReaderの内部フィールドを非公開化
2. 変換関数を内部用に追加

### Phase 5: テストの更新
1. 公開APIのテストを維持
2. 内部実装のテストを`internal/pageimpl`に移動
3. すべてのテストが通ることを確認

### Phase 6: ドキュメント更新
1. godocコメントを更新
2. examplesを更新
3. README.mdを更新

## 6. テスト戦略

### 6.1. 公開APIのテスト
```go
// page_test.go (root package)
func TestPage_DrawText(t *testing.T) {
    doc := New()
    page := doc.AddPage(PageSizeA4, Portrait)
    err := page.SetFont(FontHelvetica, 12)
    require.NoError(t, err)
    err = page.DrawText("Hello", 100, 100)
    require.NoError(t, err)
}
```

### 6.2. 内部実装のテスト
```go
// internal/pageimpl/page_impl_test.go
func TestPageImpl_DrawTextInternal(t *testing.T) {
    impl := NewPageImpl(595, 842)
    // 内部実装の詳細なテスト
}
```

### 6.3. 統合テスト
```go
// integration_test.go (root package)
func TestIntegration_CreatePDF(t *testing.T) {
    doc := New()
    page := doc.AddPage(PageSizeA4, Portrait)
    // ... 一連の操作
    var buf bytes.Buffer
    err := doc.WriteTo(&buf)
    require.NoError(t, err)
    // PDFの検証
}
```

## 7. 破壊的変更の管理

### 7.1. 影響範囲
このリファクタリングは**破壊的変更**です:
- Pageの内部フィールドへの直接アクセスが不可能になる
- 一部のヘルパーメソッドが使えなくなる

### 7.2. マイグレーションガイド

**Before:**
```go
page := doc.AddPage(PageSizeA4, Portrait)
page.content.WriteString("直接書き込み")  // NG
```

**After:**
```go
page := doc.AddPage(PageSizeA4, Portrait)
// 公開APIを使用
page.DrawText("テキスト", 100, 100)
```

### 7.3. バージョン管理
- 現在が`v0.x.x`の場合、リファクタリング後に`v1.0.0`をリリース
- BREAKING CHANGESをCHANGELOG.mdに明記
- マイグレーションガイドを提供

## 8. メリットとトレードオフ

### 8.1. メリット
1. **APIの見通しが良くなる**
   - ユーザーが使うべきメソッドのみが見える
   - godocがシンプルになる

2. **内部実装の変更が容易**
   - 内部構造を自由に変更できる
   - 公開APIを維持したまま最適化できる

3. **カプセル化の強化**
   - 内部状態が保護される
   - 不正な状態遷移を防げる

4. **循環参照の回避**
   - パッケージ間の依存関係が明確になる
   - ビルドが安定する

### 8.2. トレードオフ
1. **初期実装コスト**
   - リファクタリングに時間がかかる
   - 既存コードの移動が必要

2. **ラッパーのオーバーヘッド**
   - メソッド呼び出しが1段増える（微小）
   - インライン化で最適化可能

3. **破壊的変更**
   - 既存ユーザーへの影響がある
   - マイグレーションガイドが必要

## 9. 実装例

### 9.1. 最小限の例

**Before:**
```go
// page.go
package gopdf

import "github.com/ryomak/gopdf/internal/font"

type Page struct {
    currentFont *font.StandardFont  // 内部型露出
}

func (p *Page) getFontKey(f font.StandardFont) string {  // 公開ヘルパー
    return string(f)
}
```

**After:**
```go
// page.go
package gopdf

import "github.com/ryomak/gopdf/internal/pageimpl"

type Page struct {
    impl *pageimpl.PageImpl  // 非公開
}

func (p *Page) SetFont(f StandardFont, size float64) error {
    return p.impl.SetFont(f, size)
}

// internal/pageimpl/page_impl.go
package pageimpl

import "github.com/ryomak/gopdf/internal/font"

type PageImpl struct {
    currentFont *font.StandardFont  // 内部で管理
}

func (p *PageImpl) SetFont(f string, size float64) error {
    // 実装
}

func (p *PageImpl) getFontKey(f *font.StandardFont) string {  // 非公開
    return f.Name()
}
```

## 10. 次のステップ

1. **設計レビュー**
   - このドキュメントをレビュー
   - チームメンバー（または本人）の承認

2. **プロトタイプ実装**
   - Phase 1から順次実装
   - 各フェーズごとにテストを実行

3. **段階的なリファクタリング**
   - 小さなPRに分割
   - 各PRでテストが通ることを確認

4. **ドキュメント整備**
   - マイグレーションガイド作成
   - CHANGELOG.md更新

## 11. まとめ

このリファクタリングにより、gopdfパッケージは以下のようになります:

- **明確な公開API**: ユーザーが使うべきメソッドのみ公開
- **カプセル化された内部実装**: `internal/pageimpl`等に隠蔽
- **循環参照のない設計**: `internal/types`で共通型を定義
- **保守性の向上**: 内部実装を自由に変更可能

このアプローチは、architecture.mdで定義されたレイヤー構造を正しく実装し、
Goのベストプラクティス（最小公開API、インターフェース活用、カプセル化）に従ったものです。
