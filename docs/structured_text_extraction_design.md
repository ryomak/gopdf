# Phase 8: 構造的テキスト取得 設計書

## 1. 概要

PDFから抽出したテキストの位置情報、フォント情報、サイズなどを含む構造的なデータとして取得できるAPIを提供する。

### 1.1. 目的

- テキストの位置情報を含む構造的なデータ取得
- テキストの読み順序の推定
- テキストのフォント・サイズ情報の提供
- より高度なテキスト処理の基盤

### 1.2. スコープ

**Phase 8で実装する機能:**
- ✅ TextElementの公開型作成
- ✅ 位置情報付きテキスト抽出API
- ✅ テキストの並び替え機能（読み順序推定）
- ✅ ページ単位・全ページ対応

**Phase 8では実装しない機能:**
- 段落検出
- 表構造の認識
- 複雑なレイアウト解析（複数カラムなど）
- テキストスタイル（太字、斜体）の検出

## 2. 現状の課題

Phase 7で実装したテキスト抽出APIは文字列のみを返すため、以下の情報が失われています：

```go
// 現在のAPI
text, err := reader.ExtractPageText(0)
// => "Hello World" (位置情報なし)
```

### 2.1. 必要な情報

- **位置情報**: X, Y座標
- **サイズ**: フォントサイズ
- **フォント**: フォント名
- **境界**: Width, Height（テキストの範囲）
- **読み順序**: ページ内での論理的な順序

## 3. 設計

### 3.1. データ構造

#### 3.1.1. 公開型

```go
package gopdf

// TextElement はテキスト要素の位置とスタイル情報
type TextElement struct {
    Text   string  // テキスト内容
    X      float64 // X座標（左下原点）
    Y      float64 // Y座標（左下原点）
    Width  float64 // テキストの幅（概算）
    Height float64 // テキストの高さ（フォントサイズ）
    Font   string  // フォント名
    Size   float64 // フォントサイズ
}

// TextBlock は論理的なテキストブロック（将来拡張用）
type TextBlock struct {
    Elements []TextElement // テキスト要素のリスト
    Bounds   Rectangle     // ブロック全体の境界
    Text     string        // 結合されたテキスト
}

// Rectangle は矩形領域
type Rectangle struct {
    X      float64
    Y      float64
    Width  float64
    Height float64
}
```

### 3.2. 公開API

```go
package gopdf

// ExtractPageTextElements は位置情報付きテキスト要素を抽出
func (r *PDFReader) ExtractPageTextElements(pageNum int) ([]TextElement, error)

// ExtractAllTextElements は全ページのテキスト要素を抽出
func (r *PDFReader) ExtractAllTextElements() (map[int][]TextElement, error)

// SortTextElements はテキスト要素を読み順序でソート
func SortTextElements(elements []TextElement) []TextElement

// TextElementsToString はテキスト要素を文字列に変換
func TextElementsToString(elements []TextElement) string
```

### 3.3. テキストの並び替えアルゴリズム

PDFのコンテンツストリームは描画順序であり、読み順序ではない。読み順序を推定する必要がある。

#### 3.3.1. 基本アルゴリズム

```
1. Y座標でグループ化（行の検出）
   - 同じ行と判定する閾値: フォントサイズの50%程度

2. 各行内でX座標でソート（左から右）

3. 行をY座標の降順でソート（上から下）
   - PDFは左下原点なので、Y座標が大きい方が上
```

#### 3.3.2. 実装

```go
// SortTextElements はテキスト要素を読み順序でソート
func SortTextElements(elements []TextElement) []TextElement {
    if len(elements) == 0 {
        return elements
    }

    // 1. Y座標でグループ化（行の検出）
    lines := groupByLine(elements)

    // 2. 各行内でX座標でソート
    for _, line := range lines {
        sort.Slice(line, func(i, j int) bool {
            return line[i].X < line[j].X
        })
    }

    // 3. 行をY座標の降順でソート（上から下）
    sort.Slice(lines, func(i, j int) bool {
        return lines[i][0].Y > lines[j][0].Y
    })

    // フラット化
    result := make([]TextElement, 0, len(elements))
    for _, line := range lines {
        result = append(result, line...)
    }

    return result
}

// groupByLine は同じ行のテキスト要素をグループ化
func groupByLine(elements []TextElement) [][]TextElement {
    if len(elements) == 0 {
        return nil
    }

    // Y座標でソート
    sorted := make([]TextElement, len(elements))
    copy(sorted, elements)
    sort.Slice(sorted, func(i, j int) bool {
        return sorted[i].Y > sorted[j].Y
    })

    var lines [][]TextElement
    currentLine := []TextElement{sorted[0]}
    currentY := sorted[0].Y
    threshold := sorted[0].Size * 0.5 // 閾値

    for i := 1; i < len(sorted); i++ {
        elem := sorted[i]
        // Y座標の差が閾値以内なら同じ行
        if math.Abs(elem.Y-currentY) <= threshold {
            currentLine = append(currentLine, elem)
        } else {
            lines = append(lines, currentLine)
            currentLine = []TextElement{elem}
            currentY = elem.Y
            threshold = elem.Size * 0.5
        }
    }
    lines = append(lines, currentLine)

    return lines
}
```

### 3.4. 幅の計算

テキストの幅は正確に計算するのが難しいため、概算値を使用：

```go
// estimateTextWidth はテキストの幅を概算
func estimateTextWidth(text string, fontSize float64, font string) float64 {
    // 簡易的な幅計算
    // 英数字の平均幅は fontSizeの約60%
    avgCharWidth := fontSize * 0.6
    return float64(len(text)) * avgCharWidth
}
```

将来的には：
- フォントメトリクスを使用した正確な計算
- プロポーショナルフォントへの対応
- 文字ごとの幅情報

## 4. 実装計画

### 4.1. Phase 8.1: TextElementの公開型作成

`reader.go`に公開型を追加：

```go
package gopdf

// TextElement はテキスト要素の位置とスタイル情報
type TextElement struct {
    Text   string
    X      float64
    Y      float64
    Width  float64
    Height float64
    Font   string
    Size   float64
}
```

### 4.2. Phase 8.2: ExtractPageTextElements API実装

```go
// ExtractPageTextElements は位置情報付きテキスト要素を抽出
func (r *PDFReader) ExtractPageTextElements(pageNum int) ([]TextElement, error) {
    // ページを取得
    page, err := r.r.GetPage(pageNum)
    if err != nil {
        return nil, err
    }

    // コンテンツストリームを取得
    contentsData, err := r.r.GetPageContents(page)
    if err != nil {
        return nil, err
    }

    // パース
    parser := content.NewStreamParser(contentsData)
    operations, err := parser.ParseOperations()
    if err != nil {
        return nil, err
    }

    // 抽出
    extractor := content.NewTextExtractor(operations)
    internalElements, err := extractor.Extract()
    if err != nil {
        return nil, err
    }

    // 内部型から公開型に変換
    elements := make([]TextElement, len(internalElements))
    for i, elem := range internalElements {
        elements[i] = TextElement{
            Text:   elem.Text,
            X:      elem.X,
            Y:      elem.Y,
            Width:  estimateTextWidth(elem.Text, elem.Size, elem.Font),
            Height: elem.Size,
            Font:   elem.Font,
            Size:   elem.Size,
        }
    }

    return elements, nil
}
```

### 4.3. Phase 8.3: テキスト並び替え機能実装

`text_sort.go`を作成：

```go
package gopdf

import (
    "math"
    "sort"
)

// SortTextElements はテキスト要素を読み順序でソート
func SortTextElements(elements []TextElement) []TextElement {
    // 実装
}

// groupByLine は同じ行のテキスト要素をグループ化
func groupByLine(elements []TextElement) [][]TextElement {
    // 実装
}

// estimateTextWidth はテキストの幅を概算
func estimateTextWidth(text string, fontSize float64, font string) float64 {
    // 実装
}
```

### 4.4. Phase 8.4: テストとサンプル作成

#### テストケース

```go
func TestExtractPageTextElements(t *testing.T) {
    // Writerで生成したPDFから抽出
    doc := gopdf.New()
    page := doc.AddPage(gopdf.A4, gopdf.Portrait)
    page.SetFont(font.Helvetica, 12)
    page.DrawText("Hello", 100, 700)
    page.DrawText("World", 200, 700)

    var buf bytes.Buffer
    doc.WriteTo(&buf)

    reader, _ := gopdf.OpenReader(bytes.NewReader(buf.Bytes()))
    elements, err := reader.ExtractPageTextElements(0)

    if err != nil {
        t.Fatalf("ExtractPageTextElements failed: %v", err)
    }

    if len(elements) != 2 {
        t.Fatalf("Expected 2 elements, got %d", len(elements))
    }

    // 位置情報の検証
    if elements[0].X != 100 {
        t.Errorf("Element[0].X = %f, want 100", elements[0].X)
    }
}

func TestSortTextElements(t *testing.T) {
    elements := []TextElement{
        {Text: "World", X: 200, Y: 700, Size: 12},
        {Text: "Hello", X: 100, Y: 700, Size: 12},
        {Text: "Bottom", X: 100, Y: 600, Size: 12},
    }

    sorted := gopdf.SortTextElements(elements)

    // 期待される順序: Hello, World, Bottom
    if sorted[0].Text != "Hello" {
        t.Errorf("First element = %q, want %q", sorted[0].Text, "Hello")
    }
    if sorted[1].Text != "World" {
        t.Errorf("Second element = %q, want %q", sorted[1].Text, "World")
    }
    if sorted[2].Text != "Bottom" {
        t.Errorf("Third element = %q, want %q", sorted[2].Text, "Bottom")
    }
}
```

#### サンプル

`examples/07_structured_text/main.go`:

```go
package main

import (
    "fmt"
    "log"
    "github.com/ryomak/gopdf"
)

func main() {
    reader, err := gopdf.Open("document.pdf")
    if err != nil {
        log.Fatal(err)
    }
    defer reader.Close()

    // 位置情報付きテキスト要素を取得
    elements, err := reader.ExtractPageTextElements(0)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Text elements with positions:")
    for i, elem := range elements {
        fmt.Printf("%d: '%s' at (%.1f, %.1f) size=%.1f font=%s\n",
            i, elem.Text, elem.X, elem.Y, elem.Size, elem.Font)
    }

    // 読み順序でソート
    sorted := gopdf.SortTextElements(elements)

    fmt.Println("\nSorted text:")
    for _, elem := range sorted {
        fmt.Printf("%s ", elem.Text)
    }
    fmt.Println()
}
```

## 5. テスト計画

### 5.1. ユニットテスト

- `TestExtractPageTextElements`: 基本的な抽出
- `TestExtractPageTextElements_MultipleElements`: 複数要素
- `TestSortTextElements`: ソート機能
- `TestGroupByLine`: 行グループ化
- `TestEstimateTextWidth`: 幅計算

### 5.2. 統合テスト

- Writerで生成したPDFからの抽出
- 既存PDFからの抽出
- 複数ページの処理

## 6. 注意事項

### 6.1. 座標系

PDFは左下が原点 (0, 0) で、上に行くほどY座標が大きくなる。一般的なGUIの座標系とは逆。

### 6.2. 幅の計算精度

現在の実装では幅は概算値。正確な値が必要な場合は：
- フォントメトリクスの読み込み
- 文字ごとの幅情報の使用
が必要（将来の拡張）

### 6.3. 複雑なレイアウト

複数カラム、表、回り込みなどの複雑なレイアウトは現在の簡易アルゴリズムでは正しく処理できない可能性がある。

## 7. 参考資料

- [PDF 1.7 仕様書](https://opensource.adobe.com/dc-acrobat-sdk-docs/pdfstandards/PDF32000_2008.pdf)
  - Section 9: Text
  - Section 9.4: Text Objects
- [docs/text_extraction_design.md](./text_extraction_design.md)
