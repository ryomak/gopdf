# テキストと画像の統合的なコンテンツグルーピング 設計書

## 1. 概要

### 1.1. 目的

テキストと画像を統合的にグルーピングして、より自然なコンテンツブロックを提供する。

### 1.2. ユーザーの要望

- 文章と画像をそれぞれのレベルでグループ化したい
- 画像が挟まっているときはテキストブロックを分けて欲しい
- ページが別れた時も分ける（これは既にページ単位で処理しているので対応済み）

### 1.3. 現状の問題

**現在の`ExtractPageTextBlocks`:**
- テキストのみをグルーピング
- 画像の位置を考慮していない
- 画像が文章の間に挟まっていても、テキストが一つのブロックになってしまう

**例:**
```
テキスト行1
テキスト行2
[画像]       ← ここで区切って欲しい
テキスト行3
テキスト行4
```

現在はテキスト行1-4が全て1つのブロックになってしまう。

## 2. 設計方針

### 2.1. 基本方針

1. **画像位置を考慮したテキストグルーピング**
   - テキスト行のY座標範囲と画像のY座標範囲を比較
   - 画像が挟まっている場合、その位置でテキストブロックを分割

2. **統合的なコンテンツブロック**
   - テキストブロックと画像ブロックを混在させたリストを返す
   - Y座標順にソート（上から下）

3. **既存のAPIとの互換性**
   - `ExtractPageTextBlocks`は維持（テキストのみ）
   - 新しいAPI `ExtractPageContentBlocks`を追加（テキスト+画像）

### 2.2. データ構造

既存の`ContentBlock`インターフェースを活用：

```go
// ContentBlock はテキストブロックと画像ブロックの統一インターフェース
type ContentBlock interface {
    Bounds() Rectangle
    Type() ContentBlockType
    Position() (x, y float64)
}

// TextBlock と ImageBlock は既に ContentBlock を実装している
```

### 2.3. 新しいAPI

```go
// ExtractPageContentBlocks はテキストと画像を統合したコンテンツブロックを抽出
func (r *PDFReader) ExtractPageContentBlocks(pageNum int) ([]ContentBlock, error)

// ExtractAllContentBlocks は全ページのコンテンツブロックを抽出
func (r *PDFReader) ExtractAllContentBlocks() (map[int][]ContentBlock, error)
```

## 3. アルゴリズム

### 3.1. 画像を考慮したテキストグルーピング

```go
func (r *PDFReader) groupTextElementsWithImages(
    elements []layout.TextElement,
    images []layout.ImageBlock,
) []layout.TextBlock {
    // 1. 通常の行グルーピング
    lines := groupElementsByLine(elements)

    // 2. 画像のY座標範囲を取得
    imageRanges := getImageYRanges(images)

    // 3. 行をブロックにグルーピング（画像を考慮）
    var blocks []layout.TextBlock
    currentBlock := [][]layout.TextElement{}

    for i, line := range lines {
        lineY := getLineYRange(line)

        // 前の行からこの行までの間に画像があるかチェック
        if len(currentBlock) > 0 && hasImageBetween(currentBlock, line, imageRanges) {
            // 画像が挟まっているのでブロックを分割
            blocks = append(blocks, createTextBlockFromLines(currentBlock))
            currentBlock = [][]layout.TextElement{line}
        } else if shouldMergeLines(prevLine, line) {
            currentBlock = append(currentBlock, line)
        } else {
            // 行間が広いので新しいブロック
            if len(currentBlock) > 0 {
                blocks = append(blocks, createTextBlockFromLines(currentBlock))
            }
            currentBlock = [][]layout.TextElement{line}
        }
    }

    if len(currentBlock) > 0 {
        blocks = append(blocks, createTextBlockFromLines(currentBlock))
    }

    return blocks
}
```

### 3.2. 画像が挟まっているかの判定

```go
func hasImageBetween(
    prevLines [][]layout.TextElement,
    currLine []layout.TextElement,
    imageRanges []YRange,
) bool {
    if len(prevLines) == 0 {
        return false
    }

    // 前のブロックの最下端
    prevMinY := minY(prevLines[len(prevLines)-1])

    // 現在の行の最上端
    currMaxY := maxY(currLine)

    // この間に画像があるかチェック
    for _, imgRange := range imageRanges {
        // 画像が前のブロックと現在の行の間にある
        if imgRange.Max < prevMinY && imgRange.Min > currMaxY {
            return true
        }

        // 画像が行のY座標範囲と重なっている
        if overlapsY(imgRange, prevMinY, currMaxY) {
            return true
        }
    }

    return false
}

type YRange struct {
    Min float64 // 下端
    Max float64 // 上端
}

func getImageYRanges(images []layout.ImageBlock) []YRange {
    ranges := make([]YRange, len(images))
    for i, img := range images {
        ranges[i] = YRange{
            Min: img.Y,
            Max: img.Y + img.PlacedHeight,
        }
    }
    return ranges
}

func overlapsY(imgRange YRange, y1, y2 float64) bool {
    // y1 > y2 と仮定（PDFは下が原点）
    rangeMin := math.Min(y1, y2)
    rangeMax := math.Max(y1, y2)

    return !(imgRange.Max < rangeMin || imgRange.Min > rangeMax)
}
```

### 3.3. 統合したコンテンツブロックの返却

```go
func (r *PDFReader) ExtractPageContentBlocks(pageNum int) ([]ContentBlock, error) {
    // PageLayoutを取得（テキストと画像の両方）
    layout, err := r.ExtractPageLayout(pageNum)
    if err != nil {
        return nil, err
    }

    // 既に layout.ContentBlocks() がY座標順にソートして返してくれる
    return layout.SortedContentBlocks(), nil
}
```

## 4. 実装計画

### 4.1. Phase 1: ヘルパー関数の実装

```go
// layout.go に追加

// YRange はY座標の範囲
type YRange struct {
    Min float64 // 下端
    Max float64 // 上端
}

// getImageYRanges は画像のY座標範囲を取得
func getImageYRanges(images []layout.ImageBlock) []YRange

// getLineYRange は行のY座標範囲を取得
func getLineYRange(line []layout.TextElement) YRange

// hasImageBetween は2つのテキスト範囲の間に画像があるかチェック
func hasImageBetween(prevBlockLastLine []layout.TextElement, currLine []layout.TextElement, imageRanges []YRange) bool

// overlapsY はY座標範囲が重なっているかチェック
func overlapsY(range1, range2 YRange) bool
```

### 4.2. Phase 2: groupTextElements の改善

```go
// 既存の groupTextElements を修正
func (r *PDFReader) groupTextElements(elements []layout.TextElement) []layout.TextBlock {
    // 画像情報が必要なので、別の関数に委譲
    return r.groupTextElementsWithImages(elements, nil)
}

// 新しい関数（画像を考慮）
func (r *PDFReader) groupTextElementsWithImages(
    elements []layout.TextElement,
    images []layout.ImageBlock,
) []layout.TextBlock {
    // 実装
}
```

### 4.3. Phase 3: ExtractPageLayout の改善

```go
// ExtractPageLayout 内で groupTextElementsWithImages を使用
func (r *PDFReader) ExtractPageLayout(pageNum int) (*PageLayout, error) {
    // ... 既存のコード ...

    // 画像を先に抽出
    imageBlocks, err := imageExtractor.ExtractImagesWithPosition(page, operations)
    if err != nil {
        return nil, err
    }

    convertedImageBlocks := convertImageBlocks(imageBlocks)

    // テキストをグループ化（画像を考慮）
    textBlocks := r.groupTextElementsWithImages(
        convertTextElements(textElements),
        convertedImageBlocks,
    )

    return &PageLayout{
        PageNum:    pageNum,
        Width:      width,
        Height:     height,
        TextBlocks: textBlocks,
        Images:     convertedImageBlocks,
    }, nil
}
```

### 4.4. Phase 4: 新しいAPIの追加

```go
// reader.go に追加

// ExtractPageContentBlocks はテキストと画像を統合したコンテンツブロックを抽出
func (r *PDFReader) ExtractPageContentBlocks(pageNum int) ([]ContentBlock, error) {
    layout, err := r.ExtractPageLayout(pageNum)
    if err != nil {
        return nil, err
    }

    return layout.SortedContentBlocks(), nil
}

// ExtractAllContentBlocks は全ページのコンテンツブロックを抽出
func (r *PDFReader) ExtractAllContentBlocks() (map[int][]ContentBlock, error) {
    pageCount := r.PageCount()
    result := make(map[int][]ContentBlock)

    for i := 0; i < pageCount; i++ {
        blocks, err := r.ExtractPageContentBlocks(i)
        if err != nil {
            return nil, err
        }
        result[i] = blocks
    }

    return result, nil
}
```

## 5. テスト計画

### 5.1. ユニットテスト

```go
// テキストと画像が混在する場合
func TestGroupTextElementsWithImages(t *testing.T) {
    elements := []layout.TextElement{
        // 上のテキスト
        {Text: "Line 1", Y: 800},
        {Text: "Line 2", Y: 785},
        // [画像が Y=700-750 に配置]
        // 下のテキスト
        {Text: "Line 3", Y: 680},
        {Text: "Line 4", Y: 665},
    }

    images := []layout.ImageBlock{
        {Y: 700, PlacedHeight: 50}, // Y=700-750
    }

    blocks := groupTextElementsWithImages(elements, images)

    // 2つのブロックに分かれるべき
    if len(blocks) != 2 {
        t.Errorf("Expected 2 blocks, got %d", len(blocks))
    }
}
```

### 5.2. 統合テスト

```go
func TestExtractPageContentBlocks(t *testing.T) {
    // テキストと画像が混在するPDFを作成
    doc := gopdf.New()
    page := doc.AddPage(gopdf.PageSizeA4, gopdf.Portrait)

    page.SetFont(gopdf.FontHelvetica, 12)
    page.DrawText("Text before image", 50, 800)

    // 画像を配置
    img, _ := gopdf.LoadJPEGFile("test.jpg")
    page.DrawImage(img, 50, 700, 100, 50)

    page.DrawText("Text after image", 50, 680)

    // PDFを保存して読み込み
    // ...

    // コンテンツブロックを抽出
    blocks, err := reader.ExtractPageContentBlocks(0)

    // 3つのブロック: テキスト、画像、テキスト
    if len(blocks) != 3 {
        t.Errorf("Expected 3 blocks, got %d", len(blocks))
    }

    // 順序を確認
    if blocks[0].Type() != layout.ContentBlockTypeText {
        t.Error("First block should be text")
    }
    if blocks[1].Type() != layout.ContentBlockTypeImage {
        t.Error("Second block should be image")
    }
    if blocks[2].Type() != layout.ContentBlockTypeText {
        t.Error("Third block should be text")
    }
}
```

## 6. 使用例

```go
// 例1: ページ全体のコンテンツを順番に処理
reader, _ := gopdf.Open("document.pdf")
defer reader.Close()

blocks, _ := reader.ExtractPageContentBlocks(0)

for i, block := range blocks {
    switch block.Type() {
    case layout.ContentBlockTypeText:
        tb := block.(layout.TextBlock)
        fmt.Printf("Block %d [TEXT]: %s\n", i, tb.Text)

    case layout.ContentBlockTypeImage:
        ib := block.(layout.ImageBlock)
        fmt.Printf("Block %d [IMAGE]: %dx%d\n", i, ib.Width, ib.Height)
    }
}
```

```go
// 例2: ページ別に統合コンテンツを取得
allBlocks, _ := reader.ExtractAllContentBlocks()

for pageNum, blocks := range allBlocks {
    fmt.Printf("Page %d: %d blocks\n", pageNum, len(blocks))

    for _, block := range blocks {
        bounds := block.Bounds()
        fmt.Printf("  - %s at (%.1f, %.1f)\n",
            block.Type(), bounds.X, bounds.Y)
    }
}
```

## 7. 注意事項

### 7.1. 重なり判定の精度

画像とテキストの重なり判定は厳密ではなく、Y座標範囲のみを考慮している。
X座標も考慮したい場合は、将来的に改善が必要。

### 7.2. 既存APIへの影響

- `ExtractPageTextBlocks`: 引き続きテキストのみを返す（後方互換性維持）
- `ExtractPageLayout`: 内部で画像を考慮したグルーピングを使用するように改善
- 新しいAPI `ExtractPageContentBlocks`: テキスト+画像の統合ブロックを返す

### 7.3. パフォーマンス

画像の位置チェックが追加されるため、多数の画像がある場合は若干パフォーマンスが低下する可能性がある。
必要に応じて最適化を検討。

## 8. 参考資料

- [docs/text_block_grouping_design.md](./text_block_grouping_design.md)
- [docs/text_block_grouping_improvement.md](./text_block_grouping_improvement.md)
- [layout/layout.go](../layout/layout.go) - ContentBlock インターフェース定義
