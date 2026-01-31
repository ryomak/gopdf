# テキストブロックグルーピング機能 設計書

## 1. 概要

PDFから抽出したTextElement（一文字や小さいテキスト単位）を、段落やブロック単位でグルーピングする機能を提供する。

### 1.1. 目的

- 一文字単位ではなく、意味のある単位（段落、ブロック）でテキストを取得
- 近接する複数の行を1つのブロックとしてまとめる
- テキストの論理的な構造を推定

### 1.2. 背景

現在の`ExtractPageTextElements`は小さい単位（一文字や単語レベル）でTextElementを返すため、ユーザーがブロック単位で処理したい場合に不便である。

**現状:**
```go
elements := reader.ExtractPageTextElements(0)
// => 100個以上の小さいTextElementが返る（一文字ずつなど）
```

**要望:**
```go
blocks := reader.ExtractPageTextBlocks(0)
// => 意味のある単位（段落など）でグルーピングされたTextBlockが返る
```

### 1.3. スコープ

**実装する機能:**
- TextBlock型の定義
- 近接するTextElementをブロックにグルーピング
- グルーピングアルゴリズムの実装
- ページ単位・全ページ対応

**実装しない機能:**
- 表構造の認識
- 複雑な多段組みレイアウト解析
- セマンティックな段落検出（タイトル、本文の区別など）

## 2. 現状の課題

### 2.1. 現在の実装

`ExtractPageTextElements`は個別のTextElementを返す：

```go
type TextElement struct {
    Text   string  // "あ" や "Hello" など小さい単位
    X      float64
    Y      float64
    Width  float64
    Height float64
    Font   string
    Size   float64
}
```

### 2.2. 問題点

- 一文字ずつ処理が必要で効率が悪い
- 関連するテキストがバラバラになる
- 段落やブロックの境界が不明確

## 3. 設計

### 3.1. データ構造

#### 3.1.1. TextBlock型（公開型）

```go
package gopdf

// TextBlock は論理的なテキストブロック（段落やまとまり）
type TextBlock struct {
    Elements []TextElement // 含まれるテキスト要素
    Bounds   Rectangle     // ブロック全体の境界
    Text     string        // 結合されたテキスト
}

// Rectangle は矩形領域（既にstructured_text_extraction_design.mdで定義済み）
type Rectangle struct {
    X      float64 // 左下X座標
    Y      float64 // 左下Y座標
    Width  float64 // 幅
    Height float64 // 高さ
}
```

### 3.2. 公開API

```go
package gopdf

// ExtractPageTextBlocks はテキストをブロック単位でグルーピングして抽出
func (r *PDFReader) ExtractPageTextBlocks(pageNum int) ([]TextBlock, error)

// ExtractAllTextBlocks は全ページのテキストブロックを抽出
func (r *PDFReader) ExtractAllTextBlocks() (map[int][]TextBlock, error)

// GroupTextElements はTextElementをブロック単位でグルーピング
func GroupTextElements(elements []TextElement) []TextBlock
```

### 3.3. グルーピングアルゴリズム

#### 3.3.1. 基本方針

1. **行単位のグルーピング（既存の`groupByLine`を利用）**
   - Y座標が閾値内（フォントサイズ*0.5）なら同じ行

2. **ブロック単位のグルーピング（新規実装）**
   - 連続する行で、行間が閾値以内ならブロックに含める
   - 行間の閾値: フォントサイズ * 1.5（調整可能）
   - X座標の範囲が近い（左端が±50ポイント以内）

3. **境界の計算**
   - 最小X、最小Y、最大X、最大Yからブロックの境界を計算

#### 3.3.2. 疑似コード

```
GroupTextElements(elements):
  1. elementsを読み順序でソート (SortTextElements)
  2. 行単位でグルーピング (groupByLine)
  3. ブロック単位でグルーピング:
     - 最初の行から開始
     - 次の行との行間を計算
     - 行間が閾値以内 AND X座標範囲が近い → 同じブロック
     - そうでない → 新しいブロック
  4. 各ブロックの境界とテキストを計算
  5. TextBlockのリストを返す
```

#### 3.3.3. 実装詳細

```go
// GroupTextElements はTextElementをブロック単位でグルーピング
func GroupTextElements(elements []TextElement) []TextBlock {
    if len(elements) == 0 {
        return nil
    }

    // 1. ソート
    sorted := SortTextElements(elements)

    // 2. 行単位でグルーピング
    lines := groupByLine(sorted)

    // 3. ブロック単位でグルーピング
    var blocks []TextBlock
    currentBlock := [][]TextElement{lines[0]}

    for i := 1; i < len(lines); i++ {
        prevLine := lines[i-1]
        currLine := lines[i]

        if shouldMergeLines(prevLine, currLine) {
            currentBlock = append(currentBlock, currLine)
        } else {
            // 現在のブロックを確定
            blocks = append(blocks, createTextBlock(currentBlock))
            // 新しいブロックを開始
            currentBlock = [][]TextElement{currLine}
        }
    }
    // 最後のブロックを追加
    blocks = append(blocks, createTextBlock(currentBlock))

    return blocks
}

// shouldMergeLines は2つの行を同じブロックにマージするべきか判定
func shouldMergeLines(prevLine, currLine []TextElement) bool {
    if len(prevLine) == 0 || len(currLine) == 0 {
        return false
    }

    // 行間を計算
    // prevLineの最小Y（下端）とcurrLineの最大Y（上端）の差
    prevMinY := minY(prevLine)
    currMaxY := maxY(currLine)
    lineSpacing := prevMinY - currMaxY

    // フォントサイズの平均
    avgSize := (avgFontSize(prevLine) + avgFontSize(currLine)) / 2

    // 行間の閾値: フォントサイズ * 1.5
    lineSpacingThreshold := avgSize * 1.5

    // X座標の範囲をチェック
    prevLeftX := minX(prevLine)
    currLeftX := minX(currLine)
    xDiff := math.Abs(prevLeftX - currLeftX)

    // X座標の差の閾値: 50ポイント
    xThreshold := 50.0

    // 条件: 行間が閾値以内 AND X座標が近い
    return lineSpacing <= lineSpacingThreshold && xDiff <= xThreshold
}

// createTextBlock は行のリストからTextBlockを作成
func createTextBlock(lines [][]TextElement) TextBlock {
    if len(lines) == 0 {
        return TextBlock{}
    }

    // 全要素を収集
    var allElements []TextElement
    for _, line := range lines {
        allElements = append(allElements, line...)
    }

    // 境界を計算
    bounds := calculateBounds(allElements)

    // テキストを結合
    text := combineBlockText(lines)

    return TextBlock{
        Elements: allElements,
        Bounds:   bounds,
        Text:     text,
    }
}

// calculateBounds は要素の境界を計算
func calculateBounds(elements []TextElement) Rectangle {
    if len(elements) == 0 {
        return Rectangle{}
    }

    minX := elements[0].X
    minY := elements[0].Y
    maxX := elements[0].X + elements[0].Width
    maxY := elements[0].Y + elements[0].Height

    for _, elem := range elements[1:] {
        minX = math.Min(minX, elem.X)
        minY = math.Min(minY, elem.Y)
        maxX = math.Max(maxX, elem.X+elem.Width)
        maxY = math.Max(maxY, elem.Y+elem.Height)
    }

    return Rectangle{
        X:      minX,
        Y:      minY,
        Width:  maxX - minX,
        Height: maxY - minY,
    }
}

// combineBlockText はブロック内のテキストを結合
func combineBlockText(lines [][]TextElement) string {
    var result strings.Builder

    for i, line := range lines {
        if i > 0 {
            result.WriteString("\n")
        }

        for j, elem := range line {
            if j > 0 {
                result.WriteString(" ")
            }
            result.WriteString(elem.Text)
        }
    }

    return result.String()
}

// ヘルパー関数
func minY(elements []TextElement) float64 {
    min := elements[0].Y
    for _, e := range elements[1:] {
        if e.Y < min {
            min = e.Y
        }
    }
    return min
}

func maxY(elements []TextElement) float64 {
    max := elements[0].Y
    for _, e := range elements[1:] {
        if e.Y > max {
            max = e.Y
        }
    }
    return max
}

func minX(elements []TextElement) float64 {
    min := elements[0].X
    for _, e := range elements[1:] {
        if e.X < min {
            min = e.X
        }
    }
    return min
}

func avgFontSize(elements []TextElement) float64 {
    if len(elements) == 0 {
        return 0
    }
    sum := 0.0
    for _, e := range elements {
        sum += e.Size
    }
    return sum / float64(len(elements))
}
```

## 4. 実装計画

### 4.1. Phase 1: TextBlock型とRectangle型の追加

`reader.go`に型を追加：

```go
// Rectangle は矩形領域
type Rectangle struct {
    X      float64
    Y      float64
    Width  float64
    Height float64
}

// TextBlock は論理的なテキストブロック
type TextBlock struct {
    Elements []TextElement
    Bounds   Rectangle
    Text     string
}
```

### 4.2. Phase 2: グルーピング機能の実装

`text_block.go`を新規作成：

```go
package gopdf

// GroupTextElements の実装
// shouldMergeLines の実装
// createTextBlock の実装
// calculateBounds の実装
// combineBlockText の実装
// ヘルパー関数の実装
```

### 4.3. Phase 3: 公開API実装

`reader.go`に追加：

```go
// ExtractPageTextBlocks はテキストブロックを抽出
func (r *PDFReader) ExtractPageTextBlocks(pageNum int) ([]TextBlock, error) {
    elements, err := r.ExtractPageTextElements(pageNum)
    if err != nil {
        return nil, err
    }
    return GroupTextElements(elements), nil
}

// ExtractAllTextBlocks は全ページのテキストブロックを抽出
func (r *PDFReader) ExtractAllTextBlocks() (map[int][]TextBlock, error) {
    pageCount := r.PageCount()
    result := make(map[int][]TextBlock)

    for i := 0; i < pageCount; i++ {
        blocks, err := r.ExtractPageTextBlocks(i)
        if err != nil {
            return nil, err
        }
        result[i] = blocks
    }

    return result, nil
}
```

### 4.4. Phase 4: テストとサンプル作成

#### テストケース

`text_block_test.go`:

```go
package gopdf

import "testing"

// TestGroupTextElements_SingleBlock は1つのブロックのグルーピングをテスト
func TestGroupTextElements_SingleBlock(t *testing.T) {
    // 近接する3行
    elements := []TextElement{
        // 行1
        {Text: "a", X: 100, Y: 700, Size: 12},
        {Text: "b", X: 110, Y: 700, Size: 12},
        // 行2（行間が小さい）
        {Text: "c", X: 100, Y: 685, Size: 12},
        {Text: "d", X: 110, Y: 685, Size: 12},
        // 行3（行間が小さい）
        {Text: "e", X: 100, Y: 670, Size: 12},
    }

    blocks := GroupTextElements(elements)

    if len(blocks) != 1 {
        t.Fatalf("Expected 1 block, got %d", len(blocks))
    }

    if len(blocks[0].Elements) != 5 {
        t.Errorf("Expected 5 elements in block, got %d", len(blocks[0].Elements))
    }
}

// TestGroupTextElements_MultipleBlocks は複数ブロックのグルーピングをテスト
func TestGroupTextElements_MultipleBlocks(t *testing.T) {
    // 2つのブロック（行間が大きい）
    elements := []TextElement{
        // ブロック1
        {Text: "a", X: 100, Y: 700, Size: 12},
        {Text: "b", X: 100, Y: 685, Size: 12},
        // ブロック2（行間が大きい）
        {Text: "c", X: 100, Y: 600, Size: 12},
        {Text: "d", X: 100, Y: 585, Size: 12},
    }

    blocks := GroupTextElements(elements)

    if len(blocks) != 2 {
        t.Fatalf("Expected 2 blocks, got %d", len(blocks))
    }

    if len(blocks[0].Elements) != 2 {
        t.Errorf("Block 0: expected 2 elements, got %d", len(blocks[0].Elements))
    }

    if len(blocks[1].Elements) != 2 {
        t.Errorf("Block 1: expected 2 elements, got %d", len(blocks[1].Elements))
    }
}

// TestCalculateBounds は境界計算のテスト
func TestCalculateBounds(t *testing.T) {
    elements := []TextElement{
        {X: 100, Y: 700, Width: 50, Height: 12},
        {X: 110, Y: 685, Width: 40, Height: 12},
    }

    bounds := calculateBounds(elements)

    // minX = 100, maxX = 150, minY = 685, maxY = 712
    if bounds.X != 100 {
        t.Errorf("Bounds.X = %f, want 100", bounds.X)
    }
    if bounds.Y != 685 {
        t.Errorf("Bounds.Y = %f, want 685", bounds.Y)
    }
}
```

#### サンプル

`examples/08_text_blocks/main.go`:

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

    // テキストブロックを取得
    blocks, err := reader.ExtractPageTextBlocks(0)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Found %d text blocks\n\n", len(blocks))

    for i, block := range blocks {
        fmt.Printf("Block %d:\n", i+1)
        fmt.Printf("  Elements: %d\n", len(block.Elements))
        fmt.Printf("  Bounds: (%.1f, %.1f) - %.1fx%.1f\n",
            block.Bounds.X, block.Bounds.Y,
            block.Bounds.Width, block.Bounds.Height)
        fmt.Printf("  Text:\n%s\n\n", block.Text)
    }
}
```

## 5. テスト計画

### 5.1. ユニットテスト

- `TestGroupTextElements_SingleBlock`: 単一ブロック
- `TestGroupTextElements_MultipleBlocks`: 複数ブロック
- `TestGroupTextElements_DifferentXPositions`: X座標が異なる場合
- `TestShouldMergeLines`: 行のマージ判定
- `TestCalculateBounds`: 境界計算
- `TestCombineBlockText`: テキスト結合

### 5.2. 統合テスト

- Writerで生成したPDFからのブロック抽出
- 既存PDFからの抽出
- 複雑なレイアウトのPDF

## 6. 注意事項

### 6.1. 閾値の調整

行間の閾値とX座標の差の閾値は、PDFの種類によって調整が必要になる可能性がある。
将来的には設定可能にすることを検討。

### 6.2. 複雑なレイアウト

- 多段組み（2カラム以上）は正しく処理できない可能性
- 表組み、図表の回り込みなども対象外
- これらは将来の拡張課題

### 6.3. パフォーマンス

大量のTextElementがある場合、O(n^2)的な計算になる可能性。
必要に応じて最適化を検討。

## 7. 参考資料

- [docs/structured_text_extraction_design.md](./structured_text_extraction_design.md)
- [PDF 1.7 仕様書](https://opensource.adobe.com/dc-acrobat-sdk-docs/pdfstandards/PDF32000_2008.pdf)
