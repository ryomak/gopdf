# Generics活用設計書

## 目的
gopdfコードベースにGenericsを導入し、型安全性の向上、コードの重複削減、保守性の向上を実現する。

## 調査結果サマリー
コードベース全体を調査した結果、以下の領域でGenericsの活用が有効であることが判明した：
1. スライス/マップ操作の汎用化（重複コード削減）
2. PDFオブジェクトの型抽出（型アサーション削減）
3. フォント型の抽象化（interface{}削減）
4. リソース管理の統一化

## Phase 1: 汎用ユーティリティ関数の作成

### 1.1 パッケージ構成
```
internal/
  utils/
    generics.go       # 汎用Generics関数
    generics_test.go  # テストコード
```

### 1.2 実装する汎用関数

#### Map関数
**目的**: スライス要素の変換を型安全に実施
```go
// Map は各要素に関数を適用して新しいスライスを返す
func Map[T, U any](items []T, fn func(T) U) []U {
    result := make([]U, len(items))
    for i, item := range items {
        result[i] = fn(item)
    }
    return result
}
```

**使用箇所**:
- `layout.go:184-229` - `convertTextElements`
- `layout.go:160-181` - `convertImageBlocks`

#### Filter関数
**目的**: 条件に合う要素のみを抽出
```go
// Filter は条件を満たす要素のみを返す
func Filter[T any](items []T, predicate func(T) bool) []T {
    result := make([]T, 0, len(items))
    for _, item := range items {
        if predicate(item) {
            result = append(result, item)
        }
    }
    return result
}
```

**使用箇所**:
- テキスト要素のフィルタリング
- 画像ブロックのフィルタリング

#### GroupBy関数
**目的**: キーによる要素のグループ化
```go
// GroupBy はキー関数でスライスをマップにグループ化する
func GroupBy[T any, K comparable](items []T, keyFunc func(T) K) map[K][]T {
    result := make(map[K][]T)
    for _, item := range items {
        key := keyFunc(item)
        result[key] = append(result[key], item)
    }
    return result
}
```

**使用箇所**:
- `text_sort.go:41-73` - Y座標によるグループ化

#### Reduce関数
**目的**: スライスを単一の値に集約
```go
// Reduce はスライスを単一の値に集約する
func Reduce[T, U any](items []T, initial U, fn func(U, T) U) U {
    result := initial
    for _, item := range items {
        result = fn(result, item)
    }
    return result
}
```

**使用箇所**:
- フィルターのチェーン適用 (`reader.go:558-580`)

#### Deduplicate関数
**目的**: 重複要素の排除
```go
// Deduplicate は重複を排除し、順序を保持したスライスを返す
func Deduplicate[T comparable](items []T) []T {
    seen := make(map[T]struct{})
    result := make([]T, 0, len(items))
    for _, item := range items {
        if _, exists := seen[item]; !exists {
            seen[item] = struct{}{}
            result = append(result, item)
        }
    }
    return result
}
```

**使用箇所**:
- `document.go:44-189` - フォント/画像の重複排除

#### Keys/Values関数
**目的**: マップのキー/値をスライスとして取得
```go
// Keys はマップのキーをスライスで返す
func Keys[K comparable, V any](m map[K]V) []K {
    keys := make([]K, 0, len(m))
    for k := range m {
        keys = append(keys, k)
    }
    return keys
}

// Values はマップの値をスライスで返す
func Values[K comparable, V any](m map[K]V) []V {
    values := make([]V, 0, len(m))
    for _, v := range m {
        values = append(values, v)
    }
    return values
}
```

#### GetOrDefault関数
**目的**: マップから安全に値を取得
```go
// GetOrDefault はマップから値を取得し、存在しない場合はデフォルト値を返す
func GetOrDefault[K comparable, V any](m map[K]V, key K, defaultValue V) V {
    if v, exists := m[key]; exists {
        return v
    }
    return defaultValue
}
```

### 1.3 テスト戦略
テーブルドリブンテストを採用し、以下をカバー：
- 正常系: 通常のスライス/マップ操作
- 境界値: 空スライス、nil、単一要素
- 型の多様性: int, string, struct, pointerなど複数の型でテスト

## Phase 2: スライス操作の置き換え

### 2.1 layout.go の改善

#### Before
```go
func convertTextElements(internalElements []content.TextElement) []TextElement {
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
    return elements
}
```

#### After
```go
func convertTextElements(internalElements []content.TextElement) []TextElement {
    return utils.Map(internalElements, func(elem content.TextElement) TextElement {
        return TextElement{
            Text:   elem.Text,
            X:      elem.X,
            Y:      elem.Y,
            Width:  estimateTextWidth(elem.Text, elem.Size, elem.Font),
            Height: elem.Size,
            Font:   elem.Font,
            Size:   elem.Size,
        }
    })
}
```

**メリット**: ボイラープレートコード削減、意図の明確化

### 2.2 text_sort.go の改善

`groupByLine`関数をGenericsベースに書き換え（詳細は実装時に決定）

## Phase 3: リソース管理の統一化

### 3.1 document.go の改善

#### 現状の問題
- フォント、TTFフォント、画像で同じパターンのコードが3回繰り返されている
- 保守性が低く、バグの温床

#### 改善案: 汎用リソース収集関数
```go
// collectUniqueResources はページから重複のないリソースを収集する
func collectUniqueResources[K comparable, V any](
    pages []*Page,
    extractor func(*Page) map[K]V,
) (map[K]V, []K) {
    allResources := make(map[K]V)
    order := make([]K, 0)

    for _, page := range pages {
        resources := extractor(page)
        for key, value := range resources {
            if _, exists := allResources[key]; !exists {
                allResources[key] = value
                order = append(order, key)
            }
        }
    }

    return allResources, order
}
```

#### 使用例
```go
// フォント収集
allFonts, fontOrder := collectUniqueResources(d.pages, func(p *Page) map[string]font.StandardFont {
    return p.fonts
})

// 画像収集
allImages, imageOrder := collectUniqueResources(d.pages, func(p *Page) map[*Image]*Image {
    result := make(map[*Image]*Image)
    for _, img := range p.images {
        result[img] = img
    }
    return result
})
```

## Phase 4: PDFオブジェクト型抽出の改善

### 4.1 型アサーションヘルパー

#### ExtractAs関数
```go
// ExtractAs はcore.Objectを指定された型に安全に変換する
func ExtractAs[T any](obj core.Object) (T, bool) {
    v, ok := obj.(T)
    return v, ok
}

// MustExtractAs はcore.Objectを指定された型に変換し、失敗時はエラーを返す
func MustExtractAs[T any](obj core.Object) (T, error) {
    v, ok := obj.(T)
    if !ok {
        var zero T
        return zero, fmt.Errorf("expected type %T, got %T", zero, obj)
    }
    return v, nil
}
```

### 4.2 使用例

#### Before (`reader.go:203-206`)
```go
trailer, ok := trailerObj.(core.Dictionary)
if !ok {
    return fmt.Errorf("trailer should be dictionary, got %T", trailerObj)
}
```

#### After
```go
trailer, err := utils.MustExtractAs[core.Dictionary](trailerObj)
if err != nil {
    return fmt.Errorf("invalid trailer: %w", err)
}
```

**メリット**:
- 型チェックロジックの統一
- エラーメッセージの一貫性
- コード量の削減

## Phase 5: フォント型の抽象化（将来対応）

### 5.1 現状の問題
`translator.go`で`interface{}`を使用している：
```go
type PDFTranslatorOptions struct {
    TargetFont interface{} // font.StandardFont or *TTFFont
}
```

### 5.2 改善案（検討中）

#### Option 1: Font インターフェース
```go
type Font interface {
    Name() string
    // 共通メソッド
}

type PDFTranslatorOptions struct {
    TargetFont Font
}
```

#### Option 2: Generics
```go
type PDFTranslatorOptions[F font.StandardFont | *TTFFont] struct {
    TargetFont F
}
```

**判断**: Fontインターフェースの設計が必要なため、Phase 5で慎重に検討

## 実装順序

1. ✅ Phase 1: 汎用ユーティリティ関数パッケージ作成 + テスト
2. Phase 2: スライス操作の置き換え（低リスク、高効果）
3. Phase 3: リソース管理の統一化（中リスク、高効果）
4. Phase 4: PDFオブジェクト型抽出の改善（低リスク、中効果）
5. Phase 5: フォント型の抽象化（高リスク、要設計）

## メリット

### コード品質
- 型安全性の向上（実行時エラー削減）
- コードの可読性向上（意図の明確化）
- 重複コードの削減（DRY原則）

### 保守性
- 共通ロジックの一元管理
- テストのしやすさ向上
- バグ修正箇所の局所化

### パフォーマンス
- Genericsはコンパイル時に展開されるため、実行時オーバーヘッドなし
- インライン化による最適化の可能性

## リスクと対策

### リスク1: 既存コードの破壊
**対策**: TDDアプローチ、段階的な移行、十分なテストカバレッジ

### リスク2: 複雑性の増加
**対策**: シンプルな汎用関数から開始、ドキュメント整備

### リスク3: Go 1.18+ 依存
**対策**: go.modで明確にバージョン指定

## 次のステップ
1. Phase 1の実装とテスト
2. 既存コードの段階的置き換え
3. パフォーマンス測定とベンチマーク
4. ドキュメント更新

## 参考
- Go Generics公式ドキュメント: https://go.dev/doc/tutorial/generics
- Type Parameters Proposal: https://go.googlesource.com/proposal/+/refs/heads/master/design/43651-type-parameters.md
