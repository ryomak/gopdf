# RenderLayoutにおけるブロック順序の問題

## 問題の概要

`RenderLayout`関数でPDFを再構築する際、ブロックの順序が期待と異なる場合がある。
ユーザーが「本来一番上にあって欲しいものが下で出力されている」と報告。

## 現在の実装

### translator.go:139-204

```go
func RenderLayout(doc *Document, layout *PageLayout, opts PDFTranslatorOptions) (*Page, error) {
    // ...

    // 1. 画像を全て配置
    if opts.KeepImages {
        for _, img := range layout.Images {
            page.DrawImage(pdfImage, img.X, img.Y, img.PlacedWidth, img.PlacedHeight)
        }
    }

    // 2. テキストを全て配置
    if opts.KeepLayout {
        for _, block := range layout.TextBlocks {
            // テキストを描画
        }
    }
}
```

### 問題点

1. **画像とテキストが分離されている**
   - 先に全ての画像を描画し、次に全てのテキストを描画
   - 元のPDFでの描画順序が保持されない

2. **`layout.TextBlocks`の順序**
   - `groupTextElementsWithImages`で作成される順序を使用
   - Y座標の降順（上から下）でソートされているはず
   - しかし、明示的なソートが行われていない

3. **`ContentBlocks`が使われていない**
   - `PageLayout.SortedContentBlocks()`が画像とテキストを統合して正しい順序を返す
   - しかし、`RenderLayout`ではこれを使用していない

## 調査結果

### 基本的な座標系は正しい

テスト結果：
```
Block 0: Text="TOP TEXT (Y=750)", Rect.Y=750.00, Height=12.00, Top=762.00
Block 1: Text="MIDDLE TEXT (Y=400)", Rect.Y=400.00, Height=12.00, Top=412.00
Block 2: Text="BOTTOM TEXT (Y=100)", Rect.Y=100.00, Height=12.00, Top=112.00
```

- `Rect.Y`は左下の座標（PDF座標系）
- `Rect.Y + Rect.Height`は上端の座標
- Y座標が大きいほど上にある
- ブロックの順序は正しい（TOP → MIDDLE → BOTTOM）

### 描画位置の計算も正しい

translator.go:184-199:
```go
// 上から下に描画（Y座標が大きい方から小さい方へ）
y := block.Rect.Y + block.Rect.Height - fitted.LineHeight
for _, line := range fitted.Lines {
    _ = drawPageText(page, opts.TargetFont, line, x, y)
    y -= fitted.LineHeight  // 下に移動
}
```

- `block.Rect.Y + block.Rect.Height`でブロックの上端を取得
- そこから`fitted.LineHeight`を引いて最初の行の位置
- 各行を描画後、Y座標を減らして下に移動
- これは正しい

## 推測される問題のケース

### ケース1: 画像とテキストの重なり

元のPDF:
```
[画像A] ← 上部
[テキストB]
[画像C]
[テキストD] ← 下部
```

現在の実装での出力:
```
[画像A]
[画像C]  ← 先に全ての画像を描画
[テキストB]
[テキストD]
```

これにより、視覚的に順序が変わって見える可能性がある。

### ケース2: TextBlocksの順序が不安定

`layout.TextBlocks`の作成時にソートが行われていない場合、順序が不定になる。
特に、複雑なレイアウトや、Y座標が近いブロックが複数ある場合。

## 解決策

### 解決策1: ContentBlocksを使用する（推奨）

```go
func RenderLayout(doc *Document, layout *PageLayout, opts PDFTranslatorOptions) (*Page, error) {
    customSize := PageSize{Width: layout.Width, Height: layout.Height}
    page := doc.AddPage(customSize, Portrait)

    // ContentBlocksを使用して、画像とテキストを正しい順序で描画
    contentBlocks := layout.SortedContentBlocks()

    for _, block := range contentBlocks {
        switch block.Type {
        case ContentBlockTypeImage:
            if opts.KeepImages {
                // 画像を描画
                img := block.ImageBlock
                pdfImage, err := loadImageFromImageInfo(img.ImageInfo)
                if err == nil {
                    page.DrawImage(pdfImage, img.X, img.Y, img.PlacedWidth, img.PlacedHeight)
                }
            }

        case ContentBlockTypeText:
            if opts.KeepLayout {
                // テキストを描画
                textBlock := block.TextBlock
                // ... (既存のテキスト描画ロジック)
            }
        }
    }

    return page, nil
}
```

**メリット:**
- 画像とテキストが元の順序で描画される
- `SortedContentBlocks()`が正しい順序を保証
- より正確なPDF再構築

### 解決策2: TextBlocksを明示的にソート

```go
// テキストを配置
if opts.KeepLayout {
    // TextBlocksを明示的にY座標でソート（上から下）
    sortedBlocks := make([]TextBlock, len(layout.TextBlocks))
    copy(sortedBlocks, layout.TextBlocks)
    sort.Slice(sortedBlocks, func(i, j int) bool {
        // Y座標の上端で比較（上から下）
        topI := sortedBlocks[i].Rect.Y + sortedBlocks[i].Rect.Height
        topJ := sortedBlocks[j].Rect.Y + sortedBlocks[j].Rect.Height
        return topI > topJ
    })

    for _, block := range sortedBlocks {
        // ...
    }
}
```

**メリット:**
- 既存のコードへの影響が少ない
- 順序が明示的に保証される

**デメリット:**
- 画像との統合順序は保証されない

## 推奨実装

**解決策1（ContentBlocksを使用）を推奨**

理由:
1. 画像とテキストの正しい順序が保証される
2. `ContentBlocks`は既に実装済み
3. より正確なPDF再構築が可能

## 次のステップ

1. `RenderLayout`を`ContentBlocks`を使用するように修正
2. 既存のテストを実行して動作確認
3. 複雑なレイアウト（画像+テキスト）のテストケースを追加
4. ドキュメントを更新

## 関連ファイル

- `translator.go:138-204` - RenderLayout実装
- `layout.go:202-251` - groupTextElementsWithImages
- `layout/layout.go:44-91` - SortedContentBlocks実装
- `docs/coordinate_system.md` - 座標系の説明
