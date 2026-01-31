# ページ跨ぎブロック統合とページ再生成 設計書

## 1. 概要

### 1.1. 目的

複数ページにまたがるコンテンツを統合して扱えるようにし、テキスト編集後にページを再生成できるようにする。

### 1.2. ユーザーの要望

1. **ページを跨いだブロック統合**
   - `ExtractAllContentBlocks`でページ境界を無視してブロックをくっつけたい
   - 同じフォントサイズ・フォント名のテキストなら統合
   - サイズが違う場合は分けてもOK

2. **ページ再生成機能**
   - テキストを編集（翻訳など）した後、自動的にページを再分割したい
   - 文字が増えたらページも増やせるようにしたい

### 1.3. ユースケース

**PDF翻訳ワークフロー:**
```go
// 1. 全ページのコンテンツを統合して取得（ページを跨いで統合）
blocks := reader.ExtractAllContentBlocksFlattened(true)

// 2. テキストブロックを翻訳
for _, block := range blocks {
    if block.Type() == gopdf.ContentBlockTypeText {
        tb := block.(gopdf.TextBlock)
        tb.Text = translate(tb.Text)
        // テキストが長くなっても問題ない
    }
}

// 3. ページに再分割
pages := gopdf.SplitContentBlocksIntoPages(blocks, gopdf.PageSizeA4)

// 4. 新しいPDFを生成
doc := gopdf.New()
for _, page := range pages {
    p := doc.AddPage(gopdf.PageSizeA4, gopdf.Portrait)
    // ブロックを配置
}
```

## 2. 現状の問題

### 2.1. 現在の動作

`ExtractAllContentBlocks()` は `map[int][]ContentBlock` を返す：
```go
map[0][]ContentBlock{
    TextBlock{Text: "First paragraph on page 1"},
    TextBlock{Text: "Second paragraph"},
},
map[1][]ContentBlock{
    TextBlock{Text: "continues from page 1"}, // ページ跨ぎだが別ブロック
    TextBlock{Text: "Third paragraph"},
}
```

### 2.2. 問題点

- ページ境界で段落が分断される
- 統合処理を自分で実装する必要がある
- ページ再生成機能がない

## 3. 設計

### 3.1. 新しいAPI

#### 3.1.1. フラット化されたブロック抽出

```go
// ExtractAllContentBlocksFlattened はページを跨いでブロックを統合して返す
// mergeAcrossPagesがtrueの場合、連続するテキストブロックで、フォント属性が同じなら統合される
// falseの場合、ページ境界を保持したまま単純にフラット化される
func (r *PDFReader) ExtractAllContentBlocksFlattened(mergeAcrossPages bool) ([]ContentBlock, error)
```

#### 3.1.2. ページ分割（既存機能の公開API化）

```go
// SplitContentBlocksIntoPages はコンテンツブロックをページに分割する
// 既存の layout.PageLayout.SplitIntoPages を使いやすくラップ
func SplitContentBlocksIntoPages(blocks []ContentBlock, pageSize PageSize, options SplitOptions) ([]*PageLayout, error)

// SplitOptions はページ分割のオプション
type SplitOptions struct {
    MinSpacing  float64 // ブロック間の最小間隔（デフォルト: 10.0）
    PageMargin  float64 // ページ端からのマージン（デフォルト: 50.0）
}
```

### 3.2. ブロック統合アルゴリズム

```go
func mergeContentBlocksAcrossPages(pageBlocks map[int][]ContentBlock) []ContentBlock {
    // 1. ページ順にソートして統合リストを作成
    var allBlocks []ContentBlock
    for pageNum := 0; pageNum < len(pageBlocks); pageNum++ {
        allBlocks = append(allBlocks, pageBlocks[pageNum]...)
    }

    // 2. 連続するテキストブロックを統合
    var merged []ContentBlock
    var currentTextBlock *TextBlock

    for _, block := range allBlocks {
        switch block.Type() {
        case ContentBlockTypeText:
            tb := block.(TextBlock)

            if currentTextBlock == nil {
                // 新しいテキストブロック開始
                currentTextBlock = &tb
            } else if canMergeTextBlocks(*currentTextBlock, tb) {
                // 前のブロックと統合可能
                currentTextBlock.Text += "\n" + tb.Text
                currentTextBlock.Elements = append(currentTextBlock.Elements, tb.Elements...)
                // 境界を拡張
                updateBounds(currentTextBlock, tb)
            } else {
                // 統合できないので前のブロックを確定
                merged = append(merged, *currentTextBlock)
                currentTextBlock = &tb
            }

        case ContentBlockTypeImage:
            // 画像が来たらテキストブロックを確定
            if currentTextBlock != nil {
                merged = append(merged, *currentTextBlock)
                currentTextBlock = nil
            }
            merged = append(merged, block)
        }
    }

    // 最後のテキストブロックを追加
    if currentTextBlock != nil {
        merged = append(merged, *currentTextBlock)
    }

    return merged
}

// canMergeTextBlocks は2つのテキストブロックが統合可能か判定
func canMergeTextBlocks(block1, block2 TextBlock) bool {
    // フォント名が同じ
    if block1.Font != block2.Font {
        return false
    }

    // フォントサイズが近い（±1ポイントの差は許容）
    sizeDiff := math.Abs(block1.FontSize - block2.FontSize)
    if sizeDiff > 1.0 {
        return false
    }

    // 色が同じ
    if block1.Color != block2.Color {
        return false
    }

    return true
}

// updateBounds は境界を拡張
func updateBounds(target *TextBlock, source TextBlock) {
    minX := math.Min(target.Rect.X, source.Rect.X)
    minY := math.Min(target.Rect.Y, source.Rect.Y)

    maxX := math.Max(
        target.Rect.X + target.Rect.Width,
        source.Rect.X + source.Rect.Width,
    )
    maxY := math.Max(
        target.Rect.Y + target.Rect.Height,
        source.Rect.Y + source.Rect.Height,
    )

    target.Rect = Rectangle{
        X:      minX,
        Y:      minY,
        Width:  maxX - minX,
        Height: maxY - minY,
    }
}
```

### 3.3. ページ分割API（既存機能の公開）

```go
// SplitContentBlocksIntoPages はブロックをページに分割
func SplitContentBlocksIntoPages(
    blocks []ContentBlock,
    pageSize PageSize,
    options SplitOptions,
) ([]*PageLayout, error) {
    // デフォルト値の設定
    if options.MinSpacing == 0 {
        options.MinSpacing = 10.0
    }
    if options.PageMargin == 0 {
        options.PageMargin = 50.0
    }

    // PageLayoutを作成
    layout := &PageLayout{
        Width:  pageSize.Width,
        Height: pageSize.Height,
    }

    // ブロックをTextBlocksとImagesに分類
    for _, block := range blocks {
        switch block.Type() {
        case ContentBlockTypeText:
            layout.TextBlocks = append(layout.TextBlocks, block.(TextBlock))
        case ContentBlockTypeImage:
            layout.Images = append(layout.Images, block.(ImageBlock))
        }
    }

    // 既存の SplitIntoPages を使用
    return layout.SplitIntoPages(pageSize.Height, options.MinSpacing, options.PageMargin)
}
```

## 4. 実装計画

### 4.1. Phase 1: ブロック統合機能

```go
// reader.go に追加

// ExtractAllContentBlocksFlattened はページを跨いでブロックを統合
func (r *PDFReader) ExtractAllContentBlocksFlattened(mergeAcrossPages bool) ([]ContentBlock, error) {
    // 全ページのブロックを取得
    pageBlocks, err := r.ExtractAllContentBlocks()
    if err != nil {
        return nil, err
    }

    if !mergeAcrossPages {
        // ページ境界を保持したまま単純にフラット化
        return flattenContentBlocks(pageBlocks), nil
    }

    // ページを跨いで統合して返す
    return mergeContentBlocksAcrossPages(pageBlocks), nil
}
```

### 4.2. Phase 2: ヘルパー関数の実装

```go
// layout.go に追加

// mergeContentBlocksAcrossPages はページを跨いでブロックを統合
func mergeContentBlocksAcrossPages(pageBlocks map[int][]ContentBlock) []ContentBlock

// canMergeTextBlocks は統合可能か判定
func canMergeTextBlocks(block1, block2 TextBlock) bool

// updateBounds は境界を更新
func updateBounds(target *TextBlock, source TextBlock)
```

### 4.3. Phase 3: ページ分割APIの公開

```go
// page_split.go (新規ファイル)

// SplitOptions はページ分割のオプション
type SplitOptions struct {
    MinSpacing float64
    PageMargin float64
}

// DefaultSplitOptions はデフォルトのオプション
func DefaultSplitOptions() SplitOptions {
    return SplitOptions{
        MinSpacing: 10.0,
        PageMargin: 50.0,
    }
}

// SplitContentBlocksIntoPages はブロックをページに分割
func SplitContentBlocksIntoPages(
    blocks []ContentBlock,
    pageSize PageSize,
    options SplitOptions,
) ([]*PageLayout, error)
```

## 5. テスト計画

### 5.1. ユニットテスト

```go
// ページ跨ぎのテキストブロック統合
func TestMergeContentBlocksAcrossPages(t *testing.T) {
    pageBlocks := map[int][]ContentBlock{
        0: {
            TextBlock{Text: "Part 1", Font: "Helvetica", FontSize: 12},
        },
        1: {
            TextBlock{Text: "Part 2", Font: "Helvetica", FontSize: 12},
        },
    }

    merged := mergeContentBlocksAcrossPages(pageBlocks)

    if len(merged) != 1 {
        t.Errorf("Expected 1 block, got %d", len(merged))
    }

    tb := merged[0].(TextBlock)
    if !strings.Contains(tb.Text, "Part 1") || !strings.Contains(tb.Text, "Part 2") {
        t.Errorf("Text not merged correctly: %s", tb.Text)
    }
}

// フォントサイズが違う場合は統合しない
func TestCannotMergeDifferentFontSizes(t *testing.T) {
    pageBlocks := map[int][]ContentBlock{
        0: {
            TextBlock{Text: "Title", Font: "Helvetica", FontSize: 18},
        },
        1: {
            TextBlock{Text: "Body", Font: "Helvetica", FontSize: 12},
        },
    }

    merged := mergeContentBlocksAcrossPages(pageBlocks)

    if len(merged) != 2 {
        t.Errorf("Expected 2 blocks, got %d", len(merged))
    }
}

// 画像が挟まる場合は分割
func TestImageSplitsTextBlocks(t *testing.T) {
    pageBlocks := map[int][]ContentBlock{
        0: {
            TextBlock{Text: "Before image", Font: "Helvetica", FontSize: 12},
            ImageBlock{Width: 100, Height: 100},
            TextBlock{Text: "After image", Font: "Helvetica", FontSize: 12},
        },
    }

    merged := mergeContentBlocksAcrossPages(pageBlocks)

    if len(merged) != 3 {
        t.Errorf("Expected 3 blocks, got %d", len(merged))
    }
}
```

### 5.2. 統合テスト

```go
func TestExtractAllContentBlocksFlattened(t *testing.T) {
    // 複数ページのPDFを作成
    doc := gopdf.New()

    page1 := doc.AddPage(gopdf.PageSizeA4, gopdf.Portrait)
    page1.SetFont(gopdf.FontHelvetica, 12)
    page1.DrawText("This is page 1", 50, 800)

    page2 := doc.AddPage(gopdf.PageSizeA4, gopdf.Portrait)
    page2.SetFont(gopdf.FontHelvetica, 12)
    page2.DrawText("This is page 2", 50, 800)

    // PDFを保存して読み込み
    // ...

    // フラット化して取得
    blocks, err := reader.ExtractAllContentBlocksFlattened()
    if err != nil {
        t.Fatal(err)
    }

    // 同じフォントサイズなので統合されるはず
    if len(blocks) != 1 {
        t.Errorf("Expected 1 merged block, got %d", len(blocks))
    }
}
```

## 6. 使用例

### 6.1. PDF翻訳ワークフロー

```go
// 1. PDFを開く
reader, _ := gopdf.Open("source.pdf")
defer reader.Close()

// 2. 全ページのコンテンツを統合して取得（ページを跨いで統合）
blocks, _ := reader.ExtractAllContentBlocksFlattened(true)

// 3. テキストブロックを翻訳
translatedBlocks := make([]gopdf.ContentBlock, len(blocks))
for i, block := range blocks {
    if block.Type() == gopdf.ContentBlockTypeText {
        tb := block.(gopdf.TextBlock)
        // 翻訳（テキストが長くなってもOK）
        tb.Text = translateText(tb.Text, "en", "ja")
        translatedBlocks[i] = tb
    } else {
        translatedBlocks[i] = block
    }
}

// 4. ページに再分割
pages, _ := gopdf.SplitContentBlocksIntoPages(
    translatedBlocks,
    gopdf.PageSizeA4,
    gopdf.DefaultSplitOptions(),
)

// 5. 新しいPDFを生成
doc := gopdf.New()
for _, pageLayout := range pages {
    page := doc.AddPage(gopdf.PageSizeA4, gopdf.Portrait)

    // テキストブロックを描画
    for _, tb := range pageLayout.TextBlocks {
        page.SetFont(gopdf.FontHelvetica, tb.FontSize)
        // 複数行対応
        lines := strings.Split(tb.Text, "\n")
        y := tb.Rect.Y + tb.Rect.Height
        for _, line := range lines {
            page.DrawText(line, tb.Rect.X, y)
            y -= tb.FontSize * 1.2
        }
    }
}

doc.WriteTo(outputFile)
```

### 6.2. コンテンツ抽出と再配置

```go
// 全コンテンツを統合して取得（ページを跨いで統合）
blocks, _ := reader.ExtractAllContentBlocksFlattened(true)

// テキストのみをフィルタリング
var textBlocks []gopdf.ContentBlock
for _, block := range blocks {
    if block.Type() == gopdf.ContentBlockTypeText {
        textBlocks = append(textBlocks, block)
    }
}

// 新しいページサイズで再配置
pages, _ := gopdf.SplitContentBlocksIntoPages(
    textBlocks,
    gopdf.PageSize{Width: 500, Height: 700}, // カスタムサイズ
    gopdf.SplitOptions{
        MinSpacing: 15.0,
        PageMargin: 60.0,
    },
)
```

## 7. 注意事項

### 7.1. ページ跨ぎ統合のオプション

`ExtractAllContentBlocksFlattened(mergeAcrossPages bool)` の引数で統合動作を制御：

- **`mergeAcrossPages = true`**: ページを跨いでテキストブロックを統合
  - 用途: PDF翻訳、コンテンツ抽出・再構成など
  - フォント属性が同じなら連続するブロックを統合

- **`mergeAcrossPages = false`**: ページ境界を保持したまま単純にフラット化
  - 用途: 1ページに収めたい場合、ページ構造を保持したい場合
  - `map[int][]ContentBlock` を `[]ContentBlock` に変換するだけ
  - ブロックの統合は行わない

### 7.2. 統合の条件

以下の条件を満たすテキストブロックのみ統合：
- フォント名が同じ
- フォントサイズが近い（±1ポイント）
- 色が同じ
- 画像が挟まっていない

### 7.2. 座標情報の扱い

統合後のブロックの座標は、元のページでの座標を保持するが、ページ再生成時には新しい座標が割り当てられる。

### 7.3. パフォーマンス

大量のページがある場合、統合処理に時間がかかる可能性がある。必要に応じて最適化を検討。

## 8. 将来の拡張

- カラム検出（2カラムレイアウトの処理）
- 見出しと本文の自動区別
- 表組みの保持
- フォントの自動マッピング（翻訳時）

## 9. 参考資料

- [docs/unified_content_grouping_design.md](./unified_content_grouping_design.md)
- [docs/layout_auto_adjustment_design.md](./layout_auto_adjustment_design.md)
- [layout/operations.go](../layout/operations.go) - SplitIntoPages実装
