# PDF再構成時の座標問題の調査と修正

## 調査日
2025-10-29

## 問題の概要
PDFを読み取ってTextBlockに変換し、再度PDFとして出力すると、全く異なる構成のPDFが生成される問題が発生した。

## 原因分析

### 1. PDFの座標系
- PDFの標準座標系は左下を原点(0,0)とする
- X軸: 左から右に増加
- Y軸: 下から上に増加

### 2. TextBlockとTextElementの関係

#### TextElement
- PDFから抽出された個々のテキスト要素
- `X, Y`: テキストのベースライン位置（PDF座標系）
- `Size`: フォントサイズ
- `Text`: テキスト内容

#### TextBlock
- 複数のTextElementをグループ化した論理的なブロック
- `Text`: 全TextElementを結合した文字列
- `Elements[]`: 構成するTextElementのリスト
- `Rect`: バウンディングボックス（全TextElementを囲む矩形）
  - `X, Y`: 左下座標（全要素の最小X, 最小Y）
  - `Width, Height`: 矩形のサイズ

### 3. 問題の根本原因

**誤った実装:**
```go
// TextBlockの境界座標を使用
bounds := block.Bounds()
x := bounds.X
y := bounds.Y

// 結合されたテキスト全体を単一の位置に描画
newPage.DrawText(textBlock.Text, x, y)
```

**問題点:**
1. TextBlock.Textは複数のTextElementを結合した文字列
2. 元のPDFでは各TextElementが異なる位置に配置されていた
3. 単一の位置に全テキストを描画すると、元の配置が失われる
4. フォントサイズも固定値を使用していたため、元のサイズが失われる

## 修正内容

### 修正後の実装
```go
// 各TextElementを個別に処理
for _, elem := range textBlock.Elements {
    // 元のフォントサイズを使用
    newPage.SetTTFFont(jpFont, elem.Size)

    // 元の位置に描画（PDF座標系: 左下が原点）
    x := elem.X
    y := elem.Y

    if elem.Text != "" {
        newPage.DrawText(elem.Text, x, y)
    }
}
```

### 修正のポイント
1. **個別描画**: TextBlock.Textではなく、各TextElementを個別に描画
2. **元の位置を保持**: 各TextElementの元の座標（elem.X, elem.Y）を使用
3. **フォントサイズの保持**: 各TextElementの元のフォントサイズ（elem.Size）を使用

## 座標変換の詳細

### TextElementの座標
`TextElement.Y`は`textMatrix[5]`から取得される。これはPDFのテキスト描画マトリックスのY成分で、テキストのベースライン位置を表す。

**ソースコード（internal/content/extractor.go:225-234）:**
```go
func (e *TextExtractor) createTextElement(text string) TextElement {
    return TextElement{
        Text: text,
        X:    e.textMatrix[4], // e
        Y:    e.textMatrix[5], // f (ベースライン位置)
        Font: e.currentFont,
        Size: e.fontSize,
    }
}
```

### TextBlockの境界計算
`TextBlock.Rect`は全TextElementのバウンディングボックス。

**ソースコード（layout.go:324-365）:**
```go
func createTextBlockFromLines(lines [][]layout.TextElement) layout.TextBlock {
    // 全要素を収集
    var allElements []layout.TextElement
    for _, line := range lines {
        allElements = append(allElements, line...)
    }

    // バウンディングボックスを計算
    minX, minY := allElements[0].X, allElements[0].Y
    maxX, maxY := allElements[0].X+allElements[0].Width, allElements[0].Y+allElements[0].Height

    for _, elem := range allElements {
        minX = math.Min(minX, elem.X)
        minY = math.Min(minY, elem.Y)
        maxX = math.Max(maxX, elem.X+elem.Width)
        maxY = math.Max(maxY, elem.Y+elem.Height)
    }

    return layout.TextBlock{
        Text:     text,
        Elements: allElements,
        Rect: layout.Rectangle{
            X:      minX,  // 最小X座標
            Y:      minY,  // 最小Y座標
            Width:  maxX - minX,
            Height: maxY - minY,
        },
        ...
    }
}
```

## DrawTextの座標系

**ソースコード（page.go:74-76）:**
```go
// DrawText draws text at the specified position.
// The position (x, y) is in PDF units (points), where (0, 0) is the bottom-left corner.
func (p *Page) DrawText(text string, x, y float64) error
```

- `(x, y)`: PDF座標系での位置（左下原点）
- `y`: テキストのベースライン位置

## 画像の処理について

調査の過程で、画像ブロックにも対応する必要があることが判明したが、現時点ではgopdfに画像描画APIが実装されていない。

### 今後の課題
- `Page.DrawImage()` メソッドの実装
- 画像座標の検証機能（異常な座標値のフィルタリング）

**サンプル画像ブロック:**
```
--- ブロック 13 ---
Type: image
Position: (1900.0, -101221.9)  // 異常な座標
Size: 126.0x30.7
Image: X7
Size: 1000x247
```

## テスト結果

### 修正前
- TextBlock単位で結合されたテキストを描画
- 元の配置が失われる
- フォントサイズが固定値

### 修正後
- 各TextElementを元の位置とフォントサイズで描画
- 元のPDFに近い配置を再現
- ファイルサイズ: 元70KB → 再構成1.8MB（フォント埋め込みによる増加）

## まとめ

### 学んだこと
1. PDFの座標系は常に左下原点
2. TextBlockは論理的なグルーピングであり、配置情報としては不十分
3. 正確な再構成には各TextElementの位置情報が必要
4. フォントサイズも要素ごとに保持する必要がある

### ベストプラクティス
PDFの再構成を行う場合は、以下の点に注意：
1. TextBlock.Textではなく、TextBlock.Elementsを使用
2. 各TextElementの元の座標とフォントサイズを保持
3. 座標系の変換は不要（どちらも左下原点のPDF座標系）
4. 画像の座標値を検証し、異常値はスキップ

## 参考資料
- PDF Reference 1.7, Section 8.3: Text Objects
- PDF Reference 1.7, Section 8.4: Text Positioning
- internal/content/extractor.go: TextElement抽出処理
- layout.go: TextBlock生成処理
- page.go: DrawText実装
