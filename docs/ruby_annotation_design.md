# Phase 12: PDFルビ追加機能 設計書

## 1. 概要

既存のPDFファイルを読み込み、指定したテキストにルビ（振り仮名）を追加して新しいPDFを生成する機能を実装します。

### 1.1. 目的

- 既存PDFの日本語テキストにルビを追加
- テキスト位置を自動検出してルビを配置
- 元のPDFレイアウトを保持
- 辞書ベースの自動ルビ振り機能

### 1.2. スコープ

**Phase 12で実装する機能:**
- ✅ 基本的なルビ描画機能
- ✅ グループルビ（複数文字に1つのルビ）
- ✅ モノルビ（1文字に1つのルビ）
- ✅ PDFページの読み込みと重ね合わせ
- ✅ テキスト検索・位置特定
- ✅ ルビ配置の自動計算
- ✅ ルビ辞書による自動振り

**Phase 12では実装しない機能:**
- タグ付きPDF（PDF/UA対応）
- 縦書きルビ
- 熟語ルビの高度な配置ルール
- OCRによるテキスト認識

## 2. 技術的背景

### 2.1. ルビの種類

#### グループルビ（総ルビ）
複数の親文字に対して1つのルビを振る方式

```
  とうきょう
   東  京
```

#### モノルビ（パラルビ）
1文字ごとに1つのルビを振る方式

```
  に ほん ご
  日 本  語
```

### 2.2. ルビの配置ルール

#### 位置関係
- ルビは親文字の上部に配置（横書きの場合）
- 親文字とルビの間隔: 0.5～2.0pt
- ルビサイズ: 親文字の50%が一般的

#### 配置方法
1. **中央揃え**: ルビを親文字の中央に配置（最も一般的）
2. **左揃え**: ルビを親文字の左端に配置
3. **右揃え**: ルビを親文字の右端に配置
4. **均等配置**: ルビを親文字の幅に均等配置

### 2.3. PDF実装方法

既存PDFにルビを追加する2つのアプローチ：

#### アプローチ1: ページ重ね合わせ（Form XObject） ⭐採用
```
┌─────────────────────┐
│  新しいPDFページ     │
│  ┌───────────────┐  │
│  │ ルビレイヤー   │  │ ← 新規追加
│  └───────────────┘  │
│  ┌───────────────┐  │
│  │ 元のPDFページ  │  │ ← Form XObjectとして埋め込み
│  └───────────────┘  │
└─────────────────────┘
```

**メリット:**
- ✅ 元のPDFを完全に保持
- ✅ フォント情報を保持
- ✅ 実装が比較的簡単

**デメリット:**
- ❌ ファイルサイズが増加

#### アプローチ2: コンテンツストリーム直接編集
元のPDFのコンテンツストリームを解析して、直接ルビを挿入

**メリット:**
- ✅ ファイルサイズ最小

**デメリット:**
- ❌ 実装が非常に複雑
- ❌ 元のPDFを破壊する可能性

→ **アプローチ1を採用**

## 3. 実装設計

### 3.1. アーキテクチャ

```
┌──────────────────────────────────────┐
│  gopdf.Open(pdf)                      │
│  - 既存PDFを読み込み                  │
└────────────┬─────────────────────────┘
             │
┌────────────▼─────────────────────────┐
│  reader.ExtractPageTextElements()     │
│  - テキストと位置情報を抽出           │
└────────────┬─────────────────────────┘
             │
┌────────────▼─────────────────────────┐
│  RubyAnnotator                        │
│  - テキストを検索して位置を特定       │
│  - ルビを配置する位置を計算           │
└────────────┬─────────────────────────┘
             │
┌────────────▼─────────────────────────┐
│  Document.ImportPage()                │
│  - 元のページをForm XObjectとして取込 │
│  - ルビレイヤーを追加                 │
└────────────┬─────────────────────────┘
             │
┌────────────▼─────────────────────────┐
│  Document.WriteTo()                   │
│  - 新しいPDFとして出力                │
└──────────────────────────────────────┘
```

### 3.2. データ構造

```go
package gopdf

// RubyType はルビの種類
type RubyType int

const (
    RubyTypeGroup RubyType = iota // グループルビ
    RubyTypeMono                   // モノルビ
)

// RubyAlignment はルビの配置方法
type RubyAlignment int

const (
    RubyAlignCenter RubyAlignment = iota // 中央揃え
    RubyAlignLeft                         // 左揃え
    RubyAlignRight                        // 右揃え
    RubyAlignJustify                      // 均等配置
)

// RubyStyle はルビのスタイル設定
type RubyStyle struct {
    Type      RubyType      // ルビの種類
    Alignment RubyAlignment // 配置方法
    Offset    float64       // 親文字との間隔（pt）
    SizeRatio float64       // 親文字に対するサイズ比率（デフォルト0.5）
}

// DefaultRubyStyle はデフォルトのルビスタイル
var DefaultRubyStyle = RubyStyle{
    Type:      RubyTypeGroup,
    Alignment: RubyAlignCenter,
    Offset:    1.0,
    SizeRatio: 0.5,
}

// RubyAnnotation はルビ情報
type RubyAnnotation struct {
    Base  string     // 親文字
    Ruby  string     // ルビテキスト
    Style RubyStyle  // スタイル設定
}

// RubyDictionary はルビ辞書
type RubyDictionary map[string]string

// RubyAnnotator はPDFにルビを追加
type RubyAnnotator struct {
    reader     *PDFReader
    dictionary RubyDictionary
    style      RubyStyle
}

// NewRubyAnnotator は新しいRubyAnnotatorを作成
func NewRubyAnnotator(reader *PDFReader) *RubyAnnotator

// SetDictionary はルビ辞書を設定
func (a *RubyAnnotator) SetDictionary(dict RubyDictionary)

// SetStyle はルビスタイルを設定
func (a *RubyAnnotator) SetStyle(style RubyStyle)

// AddRuby は指定テキストにルビを追加
func (a *RubyAnnotator) AddRuby(pageNum int, base, ruby string) error

// AddRubyAuto は辞書を使って自動的にルビを追加
func (a *RubyAnnotator) AddRubyAuto(pageNum int) error

// Generate はルビ付きPDFを生成
func (a *RubyAnnotator) Generate() (*Document, error)
```

### 3.3. ページインポート機能

既存PDFのページをForm XObjectとして新しいPDFに取り込む

```go
// Document拡張

// ImportPage は既存PDFのページを新しいページとして取り込む
func (d *Document) ImportPage(reader *PDFReader, pageNum int) (*Page, error) {
    // 1. 元のページを取得
    sourcePage, err := reader.GetPage(pageNum)
    if err != nil {
        return nil, err
    }

    // 2. 元のページのサイズを取得
    width, height := reader.GetPageSize(pageNum)

    // 3. 新しいページを作成
    page := d.AddPageWithSize(width, height)

    // 4. 元のページをForm XObjectとして埋め込み
    formXObject := createFormXObject(sourcePage)
    page.drawFormXObject(formXObject, 0, 0, width, height)

    return page, nil
}
```

### 3.4. ルビ描画機能

```go
// Page拡張

// DrawRuby はルビ付きテキストを描画
func (p *Page) DrawRuby(base, ruby string, x, y float64, style RubyStyle) error {
    if p.currentTTFFont == nil {
        return fmt.Errorf("no TTF font set")
    }

    baseSize := p.fontSize
    rubySize := baseSize * style.SizeRatio

    // 親文字の幅を計算
    baseWidth, err := p.currentTTFFont.TextWidth(base, baseSize)
    if err != nil {
        return err
    }

    // ルビの幅を計算
    rubyWidth, err := p.currentTTFFont.TextWidth(ruby, rubySize)
    if err != nil {
        return err
    }

    // ルビの配置位置を計算
    rubyX := calculateRubyX(x, baseWidth, rubyWidth, style.Alignment)
    rubyY := y + baseSize + style.Offset

    // 親文字を描画
    p.DrawTextUTF8(base, x, y)

    // ルビを描画
    originalSize := p.fontSize
    p.SetTTFFont(p.currentTTFFont, rubySize)
    p.DrawTextUTF8(ruby, rubyX, rubyY)
    p.SetTTFFont(p.currentTTFFont, originalSize)

    return nil
}

// calculateRubyX はルビのX座標を計算
func calculateRubyX(baseX, baseWidth, rubyWidth float64, align RubyAlignment) float64 {
    switch align {
    case RubyAlignCenter:
        return baseX + (baseWidth-rubyWidth)/2
    case RubyAlignLeft:
        return baseX
    case RubyAlignRight:
        return baseX + baseWidth - rubyWidth
    case RubyAlignJustify:
        // 均等配置は複雑なので後で実装
        return baseX + (baseWidth-rubyWidth)/2
    default:
        return baseX
    }
}
```

### 3.5. テキスト検索・マッチング

```go
// RubyAnnotator実装

// findTextPosition はテキストの位置を検索
func (a *RubyAnnotator) findTextPosition(pageNum int, searchText string) ([]TextElement, error) {
    // ページからテキスト要素を抽出
    elements, err := a.reader.ExtractPageTextElements(pageNum)
    if err != nil {
        return nil, err
    }

    // 検索テキストにマッチする要素を探す
    var matches []TextElement
    for _, elem := range elements {
        if strings.Contains(elem.Text, searchText) {
            matches = append(matches, elem)
        }
    }

    return matches, nil
}

// AddRuby は指定テキストにルビを追加
func (a *RubyAnnotator) AddRuby(pageNum int, base, ruby string) error {
    // テキスト位置を検索
    positions, err := a.findTextPosition(pageNum, base)
    if err != nil {
        return err
    }

    if len(positions) == 0 {
        return fmt.Errorf("text not found: %s", base)
    }

    // 各位置にルビを追加（後でGenerate()時に描画）
    for _, pos := range positions {
        a.annotations = append(a.annotations, rubyAnnotation{
            pageNum:  pageNum,
            base:     base,
            ruby:     ruby,
            x:        pos.X,
            y:        pos.Y,
            baseSize: pos.Size,
        })
    }

    return nil
}
```

### 3.6. ルビ辞書機能

```go
// LoadRubyDictionary はルビ辞書を読み込む
func LoadRubyDictionary(path string) (RubyDictionary, error) {
    // JSON形式の辞書を読み込み
    // {
    //   "東京": "とうきょう",
    //   "日本": "にほん",
    //   "振り仮名": "ふりがな"
    // }
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    var dict RubyDictionary
    if err := json.Unmarshal(data, &dict); err != nil {
        return nil, err
    }

    return dict, nil
}

// AddRubyAuto は辞書を使って自動的にルビを追加
func (a *RubyAnnotator) AddRubyAuto(pageNum int) error {
    if a.dictionary == nil {
        return fmt.Errorf("dictionary not set")
    }

    // ページからテキストを抽出
    elements, err := a.reader.ExtractPageTextElements(pageNum)
    if err != nil {
        return err
    }

    // 各テキスト要素に対して辞書を適用
    for _, elem := range elements {
        for base, ruby := range a.dictionary {
            if strings.Contains(elem.Text, base) {
                // マッチした場合、ルビを追加
                a.AddRuby(pageNum, base, ruby)
            }
        }
    }

    return nil
}
```

## 4. 実装手順

### 4.1. Phase 12.1: ルビ描画機能実装

基本的なルビ描画機能を実装

```go
// page.go に追加

func (p *Page) DrawRuby(base, ruby string, x, y float64, style RubyStyle) error
func (p *Page) DrawGroupRuby(base, ruby string, x, y float64) error
func (p *Page) DrawMonoRuby(base string, rubies []string, x, y float64) error
```

### 4.2. Phase 12.2: PDF読み込み＋ルビ追加機能実装

既存PDFにルビを追加する機能

```go
// ruby_annotator.go を新規作成

type RubyAnnotator struct { ... }
func NewRubyAnnotator(reader *PDFReader) *RubyAnnotator
func (a *RubyAnnotator) AddRuby(pageNum int, base, ruby string) error
func (a *RubyAnnotator) Generate() (*Document, error)
```

### 4.3. Phase 12.3: テキスト検索・マッチング機能実装

テキスト位置の自動検出

```go
func (a *RubyAnnotator) findTextPosition(pageNum int, text string) ([]TextElement, error)
func (a *RubyAnnotator) SetDictionary(dict RubyDictionary)
func (a *RubyAnnotator) AddRubyAuto(pageNum int) error
```

### 4.4. Phase 12.4: ページインポート機能実装

既存PDFページをForm XObjectとして取り込み

```go
// document.go に追加

func (d *Document) ImportPage(reader *PDFReader, pageNum int) (*Page, error)
```

## 5. 使用例

### 5.1. 基本的な使い方

```go
package main

import (
    "os"
    "github.com/ryomak/gopdf"
)

func main() {
    // 既存PDFを読み込み
    reader, err := gopdf.Open("input.pdf")
    if err != nil {
        panic(err)
    }
    defer reader.Close()

    // ルビアノテーターを作成
    annotator := gopdf.NewRubyAnnotator(reader)

    // ルビを追加
    annotator.AddRuby(0, "東京", "とうきょう")
    annotator.AddRuby(0, "日本", "にほん")

    // ルビ付きPDFを生成
    doc, err := annotator.Generate()
    if err != nil {
        panic(err)
    }

    // ファイルに出力
    file, _ := os.Create("output_with_ruby.pdf")
    defer file.Close()
    doc.WriteTo(file)
}
```

### 5.2. 辞書を使った自動ルビ振り

```go
// ルビ辞書を作成
dict := gopdf.RubyDictionary{
    "東京":   "とうきょう",
    "日本":   "にほん",
    "振り仮名": "ふりがな",
    "漢字":   "かんじ",
}

// または、JSONファイルから読み込み
// dict, err := gopdf.LoadRubyDictionary("ruby_dict.json")

reader, _ := gopdf.Open("input.pdf")
defer reader.Close()

annotator := gopdf.NewRubyAnnotator(reader)
annotator.SetDictionary(dict)

// 辞書を使って自動的にルビを追加
annotator.AddRubyAuto(0) // ページ0

doc, _ := annotator.Generate()

file, _ := os.Create("output_auto_ruby.pdf")
defer file.Close()
doc.WriteTo(file)
```

### 5.3. スタイルのカスタマイズ

```go
reader, _ := gopdf.Open("input.pdf")
annotator := gopdf.NewRubyAnnotator(reader)

// カスタムスタイルを設定
style := gopdf.RubyStyle{
    Type:      gopdf.RubyTypeMono,      // モノルビ
    Alignment: gopdf.RubyAlignCenter,   // 中央揃え
    Offset:    2.0,                     // 親文字との間隔
    SizeRatio: 0.4,                     // 親文字の40%サイズ
}

annotator.SetStyle(style)
annotator.AddRuby(0, "振り仮名", "ふりがな")

doc, _ := annotator.Generate()
// ...
```

### 5.4. 新規PDFでのルビ使用

```go
// 新規PDFでもルビが使える
doc := gopdf.New()
page := doc.AddPage(gopdf.A4, gopdf.Portrait)

font, _ := gopdf.LoadTTF("/path/to/font.ttf")
page.SetTTFFont(font, 16)

// ルビ付きテキストを描画
page.DrawGroupRuby("東京", "とうきょう", 100, 750)
page.DrawMonoRuby("日本語", []string{"に", "ほん", "ご"}, 100, 700)

file, _ := os.Create("new_with_ruby.pdf")
doc.WriteTo(file)
```

## 6. テスト計画

### 6.1. ユニットテスト

```go
func TestDrawRuby(t *testing.T) {
    doc := gopdf.New()
    page := doc.AddPage(gopdf.A4, gopdf.Portrait)

    font, _ := gopdf.LoadTTF(getTestTTFPath())
    page.SetTTFFont(font, 16)

    err := page.DrawGroupRuby("東京", "とうきょう", 100, 700)
    if err != nil {
        t.Fatalf("DrawGroupRuby failed: %v", err)
    }

    // PDFを生成して検証
    var buf bytes.Buffer
    doc.WriteTo(&buf)

    if buf.Len() == 0 {
        t.Error("PDF is empty")
    }
}

func TestRubyAnnotator(t *testing.T) {
    // テスト用PDFを作成
    testPDF := createTestPDF(t, "東京へ行く")

    reader, _ := gopdf.Open(testPDF)
    defer reader.Close()

    annotator := gopdf.NewRubyAnnotator(reader)
    err := annotator.AddRuby(0, "東京", "とうきょう")
    if err != nil {
        t.Fatalf("AddRuby failed: %v", err)
    }

    doc, err := annotator.Generate()
    if err != nil {
        t.Fatalf("Generate failed: %v", err)
    }

    var buf bytes.Buffer
    doc.WriteTo(&buf)

    if buf.Len() == 0 {
        t.Error("Generated PDF is empty")
    }
}

func TestRubyDictionary(t *testing.T) {
    dict := gopdf.RubyDictionary{
        "東京": "とうきょう",
        "日本": "にほん",
    }

    testPDF := createTestPDF(t, "日本の東京")
    reader, _ := gopdf.Open(testPDF)
    defer reader.Close()

    annotator := gopdf.NewRubyAnnotator(reader)
    annotator.SetDictionary(dict)
    annotator.AddRubyAuto(0)

    doc, _ := annotator.Generate()

    var buf bytes.Buffer
    doc.WriteTo(&buf)

    // PDFに両方のルビが含まれているか確認
    pdfStr := buf.String()
    if !strings.Contains(pdfStr, "とうきょう") {
        t.Error("PDF should contain ruby for 東京")
    }
    if !strings.Contains(pdfStr, "にほん") {
        t.Error("PDF should contain ruby for 日本")
    }
}
```

## 7. 注意事項

### 7.1. 制限事項

- **横書きのみ**: 縦書きテキストには未対応
- **単純な配置**: 複雑な熟語ルビには未対応
- **フォント制限**: TTFフォントが必要
- **Form XObject**: ファイルサイズが増加する可能性

### 7.2. パフォーマンス

- **テキスト検索**: ページ内の全テキスト要素をスキャン
- **位置計算**: 文字幅計算のため、TTFフォントメトリクス使用
- **メモリ**: 元のPDFページ全体をメモリに保持

### 7.3. 互換性

- PDF 1.4以降（Form XObject使用）
- ほとんどのPDFビューアで表示可能
- 印刷時も正しく出力される

## 8. 参考資料

- [W3C Ruby Annotation](https://www.w3.org/TR/ruby/)
- [JIS X 4051 日本語文書の組版方法](https://www.jisc.go.jp/)
- [PDF 1.7 仕様書](https://opensource.adobe.com/dc-acrobat-sdk-docs/pdfstandards/PDF32000_2008.pdf)
  - Section 8.10: Form XObjects
