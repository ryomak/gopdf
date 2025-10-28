# Ruby Annotation (Furigana) Example

このサンプルは、日本語テキストにルビ（ふりがな）を追加する機能を実演します。

## 機能

1. **基本的なルビ描画**: 親文字の上にルビテキストを配置
2. **配置オプション**: 中央揃え、左揃え、右揃え
3. **サイズ調整**: ルビのサイズ比率を調整可能
4. **ActualText対応**: PDFからテキストをコピーする時の動作を制御
5. **複数ルビ**: 複数のルビテキストを連続して描画

## 実行方法

### 前提条件: 日本語フォントのダウンロード

日本語を表示するには、日本語対応のTTFフォントが必要です。

#### Noto Sans JPフォントをダウンロード

**重要要件**:
- ✅ **TTF（TrueType Font）形式**のファイルが必要
- ✅ **静的フォント（Static Font）**を使用（可変フォントは不可）

**ダウンロード方法**:

1. https://fonts.google.com/noto/specimen/Noto+Sans+JP にアクセス
2. 右上の **"Get font"** → **"Download all"** をクリック
3. ダウンロードしたZIPファイルを解凍
4. **`static/NotoSansJP-Regular.ttf`** を探す
   - ⚠️ **必ず `static/` フォルダ内のファイル**を使用してください
5. `examples/11_ruby_annotation/` ディレクトリに `NotoSansJP-Regular.ttf` としてコピー

**ファイルサイズの確認**:
- ✅ 正しいファイル: 約 1.5MB ~ 4MB
- ❌ 間違ったファイル: 9MB以上（可変フォント）

確認コマンド:
```bash
ls -lh NotoSansJP-Regular.ttf
file NotoSansJP-Regular.ttf  # "TrueType Font data" と表示されるべき
```

### サンプルを実行

```bash
cd examples/11_ruby_annotation
go run main.go
```

## 出力ファイル

- `ruby_basic.pdf`: 基本的なルビの例
- `ruby_alignment.pdf`: 配置オプションの例
- `ruby_actualtext.pdf`: ActualTextコピーモードの例
- `ruby_multiple.pdf`: 複数のルビテキストの例

## コード概要

### 1. 基本的なルビ描画

```go
// ルビテキストを作成
rubyText := gopdf.NewRubyText("東京", "とうきょう")

// スタイルを設定
style := gopdf.DefaultRubyStyle()

// ルビを描画
width, err := page.DrawRuby(rubyText, x, y, style)
```

### 2. ルビスタイルのカスタマイズ

```go
style := gopdf.RubyStyle{
    Alignment: gopdf.RubyAlignCenter,  // 中央揃え
    Offset:    1.0,                     // 親文字との間隔（pt）
    SizeRatio: 0.5,                     // 親文字の50%サイズ
    CopyMode:  gopdf.RubyCopyBase,      // 親文字のみコピー
}
```

### 3. ActualTextでコピー動作を制御

```go
style := gopdf.DefaultRubyStyle()

// 親文字のみコピー（デフォルト）
style.CopyMode = gopdf.RubyCopyBase
page.DrawRubyWithActualText(rubyText, x, y, style)
// コピー結果: "東京"

// ルビのみコピー
style.CopyMode = gopdf.RubyCopyRuby
page.DrawRubyWithActualText(rubyText, x, y, style)
// コピー結果: "とうきょう"

// 両方コピー
style.CopyMode = gopdf.RubyCopyBoth
page.DrawRubyWithActualText(rubyText, x, y, style)
// コピー結果: "東京(とうきょう)"
```

### 4. 複数のルビテキストを描画

```go
// ヘルパー関数で複数のルビテキストを作成
texts := gopdf.NewRubyTextPairs(
    "私", "わたし",
    "日本", "にほん",
    "住", "す",
)

// 一度に描画
totalWidth, err := page.DrawRubyTexts(texts, x, y, style, true)

// 後に続くテキストを描画
page.DrawTextUTF8("んでいます。", x+totalWidth, y)
```

## API リファレンス

### RubyText 型

```go
type RubyText struct {
    Base string // 親文字（漢字など）
    Ruby string // ルビテキスト（ひらがななど）
}
```

### RubyStyle 型

```go
type RubyStyle struct {
    Alignment RubyAlignment // 配置方法
    Offset    float64       // 親文字との間隔（pt）
    SizeRatio float64       // 親文字に対するサイズ比率（0.0-1.0）
    CopyMode  RubyCopyMode  // コピー時の動作
}
```

### RubyAlignment 定数

```go
const (
    RubyAlignCenter  // 中央揃え（デフォルト）
    RubyAlignLeft    // 左揃え
    RubyAlignRight   // 右揃え
)
```

### RubyCopyMode 定数

```go
const (
    RubyCopyBase  // 親文字のみコピー（デフォルト）
    RubyCopyRuby  // ルビのみコピー
    RubyCopyBoth  // 両方コピー（親文字(ルビ)形式）
)
```

### Page メソッド

#### DrawRuby

```go
func (p *Page) DrawRuby(rubyText RubyText, x, y float64, style RubyStyle) (float64, error)
```

ルビ（ふりがな）テキストを描画します。描画した幅を返します。

#### DrawRubyWithActualText

```go
func (p *Page) DrawRubyWithActualText(rubyText RubyText, x, y float64, style RubyStyle) (float64, error)
```

ActualTextサポート付きでルビを描画します。PDFからテキストをコピーする時の動作を`CopyMode`で制御できます。

#### DrawRubyTexts

```go
func (p *Page) DrawRubyTexts(texts []RubyText, x, y float64, style RubyStyle, useActualText bool) (float64, error)
```

複数のルビテキストを連続して描画します。合計幅を返します。

### ヘルパー関数

#### NewRubyText

```go
func NewRubyText(base, ruby string) RubyText
```

単一のRubyTextを作成します。

#### NewRubyTextPairs

```go
func NewRubyTextPairs(pairs ...string) []RubyText
```

複数のRubyTextをペアから作成します。偶数個の文字列を受け取り、2つずつペアにしてRubyTextのスライスを返します。

例: `NewRubyTextPairs("私", "わたし", "日本", "にほん")` → `[{Base:"私", Ruby:"わたし"}, {Base:"日本", Ruby:"にほん"}]`

#### DefaultRubyStyle

```go
func DefaultRubyStyle() RubyStyle
```

デフォルトのRubyStyleを返します:
- Alignment: `RubyAlignCenter`
- Offset: `1.0`
- SizeRatio: `0.5`
- CopyMode: `RubyCopyBase`

## 制限事項

- 横書き（水平テキスト）のみ対応
- 縦書きテキストには未対応
- ルビテキストの自動折り返しは未対応
- 親文字が1文字でない場合のルビの細かい配置制御は未対応

## トラブルシューティング

### 文字化けする

→ TTFフォントが読み込まれていません。`NotoSansJP-Regular.ttf` が正しい場所にあるか確認してください。

### フォントが見つからない

```
Warning: TTF font not found. Skipping Japanese examples
```

→ `NotoSansJP-Regular.ttf` をダウンロードして、`examples/11_ruby_annotation/` ディレクトリに配置してください。

### PDFからコピーしてもActualTextが反映されない

→ 使用しているPDFビューアによっては、ActualTextのサポートが不完全な場合があります。Adobe Acrobat ReaderやMac標準のプレビューアプリで試してください。

## 参考

- 設計書: [docs/ruby_annotation_design.md](../../docs/ruby_annotation_design.md)
- Noto Sans JP: https://fonts.google.com/noto/specimen/Noto+Sans+JP
