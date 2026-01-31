# ContentBlockインターフェース改善設計書

## 1. 概要

型アサーションを排除し、インターフェースメソッドとジェネリクスを組み合わせて型安全なブロック操作を実現する。

## 2. 現状の問題

### 2.1. 型アサーションの多用

以下の箇所で`ok`チェックなしの型アサーションが使われており、パニックのリスクがある：

```go
// page_split.go:43-45
pageLayout.TextBlocks = append(pageLayout.TextBlocks, block.(TextBlock))
pageLayout.Images = append(pageLayout.Images, block.(ImageBlock))

// layout/operations.go:141-147
tb := block.(TextBlock)
tb.Rect.Y = newY
ib := block.(ImageBlock)
ib.Y = newY

// layout/strategies.go:36-41
tb := block.(TextBlock)
ib := block.(ImageBlock)
```

### 2.2. 問題点

1. **型安全性**: 型アサーションは実行時エラーのリスク
2. **冗長性**: switch文で`Type()`をチェックした後にアサーションが必要
3. **保守性**: 新しいブロック型を追加する際、全箇所で修正が必要

## 3. 設計

### 3.1. ContentBlockインターフェースの拡張

```go
type ContentBlock interface {
    // 既存メソッド
    Bounds() Rectangle
    Type() ContentBlockType
    Position() (x, y float64)

    // 新規メソッド
    WithY(y float64) ContentBlock      // 新しいY座標でコピーを返す
    AddToLayout(pl *PageLayout)        // 自身を適切なスライスに追加
}
```

### 3.2. TextBlockの実装

```go
// WithY は新しいY座標でTextBlockのコピーを返す
func (tb TextBlock) WithY(y float64) ContentBlock {
    newTB := tb
    newTB.Rect.Y = y
    return newTB
}

// AddToLayout はTextBlockをPageLayoutのTextBlocksに追加する
func (tb TextBlock) AddToLayout(pl *PageLayout) {
    pl.TextBlocks = append(pl.TextBlocks, tb)
}
```

### 3.3. ImageBlockの実装

```go
// WithY は新しいY座標でImageBlockのコピーを返す
func (ib ImageBlock) WithY(y float64) ContentBlock {
    newIB := ib
    newIB.Y = y
    return newIB
}

// AddToLayout はImageBlockをPageLayoutのImagesに追加する
func (ib ImageBlock) AddToLayout(pl *PageLayout) {
    pl.Images = append(pl.Images, ib)
}
```

### 3.4. 使用例: 型アサーション排除

**Before (型アサーション使用):**
```go
switch block.Type() {
case ContentBlockTypeText:
    tb := block.(TextBlock)
    tb.Rect.Y = newY
    currentPage.TextBlocks = append(currentPage.TextBlocks, tb)
case ContentBlockTypeImage:
    ib := block.(ImageBlock)
    ib.Y = newY
    currentPage.Images = append(currentPage.Images, ib)
}
```

**After (インターフェースメソッド使用):**
```go
newBlock := block.WithY(newY)
newBlock.AddToLayout(currentPage)
```

## 4. 影響範囲

### 4.1. 修正が必要なファイル

| ファイル | 修正内容 |
|---------|---------|
| `layout/layout.go` | `ContentBlock`インターフェースに`WithY`, `AddToLayout`を追加 |
| `layout/blocks.go` | `TextBlock`, `ImageBlock`に新メソッドを実装 |
| `page_split.go` | 型アサーションを`AddToLayout`に置換 |
| `layout/operations.go` | 型アサーションを`WithY`, `AddToLayout`に置換 |
| `layout/strategies.go` | 型アサーションを`WithY`, `AddToLayout`に置換 |
| `translator.go` | 型アサーションを新メソッドに置換（okチェックありなので優先度低） |

### 4.2. 後方互換性

- インターフェースにメソッドを追加するため、外部で`ContentBlock`を実装している場合は破壊的変更
- ただし、このパッケージ内でのみ使用されているため、実質的な影響は軽微

## 5. テスト計画

### 5.1. ユニットテスト

```go
func TestTextBlock_WithY(t *testing.T) {
    tb := TextBlock{
        Text: "Hello",
        Rect: Rectangle{X: 100, Y: 200, Width: 300, Height: 50},
    }

    newBlock := tb.WithY(500)

    // 元のブロックは変更されない
    assert.Equal(t, 200.0, tb.Rect.Y)

    // 新しいブロックは新しいY座標を持つ
    _, newY := newBlock.Position()
    assert.Equal(t, 500.0, newY)

    // 型がTextBlockであること
    assert.Equal(t, ContentBlockTypeText, newBlock.Type())
}

func TestImageBlock_WithY(t *testing.T) {
    ib := ImageBlock{
        X: 100,
        Y: 200,
        PlacedWidth: 300,
        PlacedHeight: 400,
    }

    newBlock := ib.WithY(600)

    // 元のブロックは変更されない
    assert.Equal(t, 200.0, ib.Y)

    // 新しいブロックは新しいY座標を持つ
    _, newY := newBlock.Position()
    assert.Equal(t, 600.0, newY)
}

func TestContentBlock_AddToLayout(t *testing.T) {
    pl := &PageLayout{
        Width: 595,
        Height: 842,
    }

    tb := TextBlock{Text: "Test", Rect: Rectangle{X: 100, Y: 700}}
    ib := ImageBlock{X: 100, Y: 500, PlacedWidth: 200, PlacedHeight: 150}

    tb.AddToLayout(pl)
    ib.AddToLayout(pl)

    assert.Equal(t, 1, len(pl.TextBlocks))
    assert.Equal(t, 1, len(pl.Images))
    assert.Equal(t, "Test", pl.TextBlocks[0].Text)
}
```

## 6. 実装手順

1. `layout/layout.go`: インターフェースに新メソッドを追加
2. `layout/blocks.go`: `TextBlock`, `ImageBlock`に実装を追加
3. `layout/blocks_test.go`: 新メソッドのテストを追加
4. `layout/operations.go`: 型アサーションを置換
5. `layout/strategies.go`: 型アサーションを置換
6. `page_split.go`: 型アサーションを置換
7. `make ci`で全テスト通過を確認

## 7. 将来の拡張

### 7.1. 追加可能なメソッド

```go
type ContentBlock interface {
    // ...既存メソッド...

    // 将来追加可能
    WithPosition(x, y float64) ContentBlock  // X, Y両方を変更
    WithBounds(rect Rectangle) ContentBlock  // 境界全体を変更
    Clone() ContentBlock                     // 完全なコピー
}
```

### 7.2. ジェネリクスとの組み合わせ

```go
// 型を保持したままブロックを変換
func MapBlocks[T ContentBlock](blocks []T, fn func(T) T) []T {
    result := make([]T, len(blocks))
    for i, b := range blocks {
        result[i] = fn(b)
    }
    return result
}
```
