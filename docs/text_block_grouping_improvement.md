# TextBlockグルーピング改善 調査・設計書

## 1. 問題の概要

### 1.1. 現状の問題

`ExtractPageTextBlocks`のグルーピングロジックが厳しすぎて、同じ段落の文章が複数のブロックに分割されてしまう。

### 1.2. ユーザーの要望

- 基本的に何行も続くもの（改行を含むもの）をひとつのブロックとして扱いたい
- 段落が変わる、または画像が挟まるときにブロックを切り上げて欲しい

## 2. 現在の実装の問題点

### 2.1. 実装位置

- ファイル: `layout.go:176-221`
- 関数: `groupTextElements()`

### 2.2. 現在のアルゴリズム

```go
// 1. Y座標でソート（上から下）
// 2. 前の要素との距離を計算
xDist := math.Abs(elem.X - (prev.X + prev.Width))
yDist := math.Abs(elem.Y - prev.Y)

// 3. 近接判定
if yDist < threshold && xDist < prev.Size*2 {
    // 同じブロック
}
```

### 2.3. 問題点の詳細

#### 問題1: Y座標の閾値が厳しすぎる

```go
threshold := 5.0 // ピクセル単位の閾値
```

- **影響**: 行間（通常12-20ポイント）が5.0ピクセルを超えると別ブロックになる
- **結果**: 改行を含む段落が複数のブロックに分割される
- **設計書との乖離**: 設計書では「フォントサイズ * 1.5」と記載されているが、実装では固定値5.0

#### 問題2: 1段階のグルーピング

現在の実装は要素を直接ブロックにグルーピングしている。

**設計書の方針（text_block_grouping_design.md）:**
1. 行単位のグルーピング（Y座標が閾値内なら同じ行）
2. ブロック単位のグルーピング（連続する行で行間が閾値以内ならブロックに含める）

**現在の実装:**
- 1段階で処理しているため、行の概念がない
- 改行の判定ができない

#### 問題3: X座標の判定が不適切

```go
xDist := math.Abs(elem.X - (prev.X + prev.Width))
if xDist < prev.Size*2 {
    // 同じブロック
}
```

- **想定**: 同じ行内の単語を繋げる判定
- **問題**: 段落の行同士をグルーピングする判定としては不適切
- **改善案**: 各行の左端のX座標を比較し、段落の始まりが揃っているかを判定

## 3. 改善方針

### 3.1. 設計書の方針に従う

`text_block_grouping_design.md`に記載されているアルゴリズムを正しく実装する。

### 3.2. 2段階グルーピング

#### ステップ1: 行単位のグルーピング

```go
// Y座標が近い（フォントサイズ*0.5以内）要素を同じ行とする
func groupByLine(elements []layout.TextElement) [][]layout.TextElement
```

#### ステップ2: ブロック単位のグルーピング

```go
// 連続する行で以下の条件を満たす場合、同じブロックとする：
// 1. 行間が閾値以内（フォントサイズ * 1.5）
// 2. X座標の左端が近い（±50ポイント以内）
func groupLinesByBlock(lines [][]layout.TextElement) []layout.TextBlock
```

### 3.3. 具体的な判定条件

#### 行の判定（同じ行か？）

```go
yDiff := math.Abs(elem1.Y - elem2.Y)
avgSize := (elem1.Size + elem2.Size) / 2
lineThreshold := avgSize * 0.5

if yDiff < lineThreshold {
    // 同じ行
}
```

#### ブロックの判定（同じ段落か？）

```go
// 行間を計算（PDFは下が原点なので注意）
prevMinY := minY(prevLine)  // 前の行の下端
currMaxY := maxY(currLine)  // 現在の行の上端
lineSpacing := prevMinY - currMaxY

// フォントサイズの平均
avgSize := (avgFontSize(prevLine) + avgFontSize(currLine)) / 2

// 判定条件
lineSpacingThreshold := avgSize * 1.5
xThreshold := 50.0

prevLeftX := minX(prevLine)
currLeftX := minX(currLine)
xDiff := math.Abs(prevLeftX - currLeftX)

if lineSpacing <= lineSpacingThreshold && xDiff <= xThreshold {
    // 同じ段落
}
```

### 3.4. テキストの結合方法

```go
// 行内: スペース区切り
for j, elem := range line {
    if j > 0 {
        result.WriteString(" ")
    }
    result.WriteString(elem.Text)
}

// 行間: 改行
for i, line := range lines {
    if i > 0 {
        result.WriteString("\n")
    }
    // 行のテキスト
}
```

## 4. 実装計画

### 4.1. ヘルパー関数の追加

```go
// minY: 要素リストの最小Y座標を取得
func minY(elements []layout.TextElement) float64

// maxY: 要素リストの最大Y座標を取得
func maxY(elements []layout.TextElement) float64

// minX: 要素リストの最小X座標を取得
func minX(elements []layout.TextElement) float64

// avgFontSize: 要素リストの平均フォントサイズを取得
func avgFontSize(elements []layout.TextElement) float64
```

### 4.2. 行グルーピング関数

```go
// groupByLine はTextElementを行単位でグルーピング
func groupByLine(elements []layout.TextElement) [][]layout.TextElement {
    // Y座標でソート
    // 近接する要素を同じ行にグループ化
}
```

### 4.3. ブロックグルーピング関数の書き換え

```go
// groupTextElements はTextElementsをTextBlocksにグループ化
func (r *PDFReader) groupTextElements(elements []layout.TextElement) []layout.TextBlock {
    // 1. 行単位でグルーピング
    lines := groupByLine(elements)

    // 2. ブロック単位でグルーピング
    var blocks []layout.TextBlock
    currentBlock := [][]layout.TextElement{lines[0]}

    for i := 1; i < len(lines); i++ {
        if shouldMergeLines(lines[i-1], lines[i]) {
            currentBlock = append(currentBlock, lines[i])
        } else {
            blocks = append(blocks, createTextBlock(currentBlock))
            currentBlock = [][]layout.TextElement{lines[i]}
        }
    }

    blocks = append(blocks, createTextBlock(currentBlock))
    return blocks
}
```

### 4.4. テキスト結合の改善

```go
// combineBlockText はブロック内のテキストを結合（改行を保持）
func combineBlockText(lines [][]layout.TextElement) string {
    var result strings.Builder

    for i, line := range lines {
        if i > 0 {
            result.WriteString("\n")  // 行間は改行
        }

        for j, elem := range line {
            if j > 0 {
                result.WriteString(" ")  // 単語間はスペース
            }
            result.WriteString(elem.Text)
        }
    }

    return result.String()
}
```

## 5. テスト計画

### 5.1. ユニットテスト

- 行グルーピングのテスト
  - 同じY座標の要素が同じ行になるか
  - 異なるY座標の要素が別の行になるか

- ブロックグルーピングのテスト
  - 行間が近い行が同じブロックになるか
  - 行間が遠い行が別ブロックになるか
  - X座標が異なる行が別ブロックになるか

### 5.2. 統合テスト

既存のサンプルPDFでテスト：
- `examples/15_japanese_text_blocks/main.go`を実行
- 段落が適切にグルーピングされているか確認

## 6. 将来の拡張

### 6.1. 画像を考慮したグルーピング

現在はテキストのみでグルーピングしているが、将来的には：
- 画像の位置を取得
- 画像が挟まっている場合、前後でブロックを分ける

### 6.2. 閾値の調整可能化

現在はハードコードされた閾値を使用しているが、将来的には：
- オプション構造体で閾値を指定可能にする
- PDFの種類に応じて自動調整

## 7. 参考

- 元の設計書: `docs/text_block_grouping_design.md`
- 実装ファイル: `layout.go`
- テストファイル: `layout_test.go`
