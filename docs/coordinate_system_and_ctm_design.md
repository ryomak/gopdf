# PDF座標系とCTMの正しい処理設計

## 1. 問題の概要

### 1.1 現象
Receipt-2021-3422.pdfを読み取ると、テキストの上下が反転して出力される。

- 元のPDFで一番下にある「Page 1 of 1」がY=754.0として抽出される
- これはページ高さ792の95.2%の位置で、ページ上部を示す
- しかし実際には視覚的に一番下に表示されているべき

### 1.2 根本原因
PDF内で使用されているCTM（Current Transformation Matrix）により、Y軸が反転した座標系が使われているが、現在の実装ではこれを考慮していない。

## 2. PDFの座標系の仕様

### 2.1 標準PDF座標系
- 原点：ページ左下 (0, 0)
- X軸：右方向が正
- Y軸：上方向が正
- ページサイズ：例えばLetterサイズは612×792ポイント

```
     Y=792 ─────────────────
           │               │
           │               │  標準座標系
           │               │  Y軸は上向き
           │               │
     Y=0   └───────────────┘
          X=0             X=612
```

### 2.2 CTM（Current Transformation Matrix）
PDFのコンテンツストリームでは、`cm`オペレータを使って座標変換を適用できる。

```
a b c d e f cm
```

これは以下の変換行列を表す：
```
[ a  b  0 ]
[ c  d  0 ]
[ e  f  1 ]
```

座標変換：
```
x' = a*x + c*y + e
y' = b*x + d*y + f
```

### 2.3 Y軸反転のCTM
Receipt-2021-3422.pdfで使用されているCTM：
```
.23999999 0 0 -.23999999 0 792 cm
```

この変換の意味：
- a = 0.24: X軸のスケール
- b = 0: X軸の傾き
- c = 0: Y軸の傾き
- **d = -0.24**: Y軸のスケール（**負の値でY軸反転**）
- e = 0: X方向の移動
- f = 792: Y方向の移動

このCTMの効果：
1. Y軸が反転される（d < 0）
2. 原点がページ上端(0, 792)に移動
3. 0.24倍にスケールされる

変換後の座標系：
```
     Y=0   ─────────────────  ← 元のY=792の位置
           │               │
           │               │  反転座標系
           │               │  Y軸は下向き
           │               │
     Y=?   └───────────────┘  ← 元のY=0の位置
          X=0
```

## 3. 具体的な問題事例の分析

### 3.1 Receipt-2021-3422.pdfのコンテンツストリーム

```
.23999999 0 0 -.23999999 0 792 cm    ← ページレベルのCTM（Y軸反転）
q
...
BT
/F4 7.5 Tf
1 0 0 -1 543.57813 754 Tm              ← テキストマトリックス（これもY反転）
<0176> Tj
...
```

### 3.2 座標の意味

「Page 1 of 1」のテキストマトリックス：`1 0 0 -1 543.57813 754 Tm`
- 位置: (543.57813, 754)
- テキストマトリックスのd=-1もY軸を反転している

この座標754は**CTM適用後の座標空間**での値。

### 3.3 なぜY軸反転が使われるのか

多くのグラフィックスシステム（画像、スクリーン座標など）は原点が左上でY軸が下向き。
このようなシステムから生成されたPDFでは、Y軸を反転させた座標系を使うことで、
元の座標系をそのまま使える。

## 4. 現在の実装の問題点

### 4.1 TextExtractor
- CTMの存在を認識している（`graphicsState.CTM`）
- しかし、テキスト座標の抽出時にCTMを適用していない
- `createTextElement()`で`textMatrix[4]`, `textMatrix[5]`をそのまま使用

### 4.2 ソート処理
- `PageLayout.SortedContentBlocks()`で上から下にソート
- 標準PDF座標系（Y軸上向き）を前提としている
- Y軸が反転した座標系を考慮していない

### 4.3 試みた解決策とその問題

#### 試行1: CTMを座標に適用
```go
x, y := e.graphicsState.CTM.TransformPoint(e.textMatrix[4], e.textMatrix[5])
```
結果：Y=2546（ページ高さの321%）← 不正確

#### 試行2: CTMの逆変換を適用
```go
inverseCTM := e.graphicsState.CTM.Inverse()
x, y := inverseCTM.TransformPoint(e.textMatrix[4], e.textMatrix[5])
```
結果：Y=-2546（負の値）← 不正確

#### 問題の本質
テキストマトリックスの座標は**既にCTM適用後の座標空間にある**ため、
さらにCTMを適用すると二重変換になる。

## 5. 正しい解決策の設計

### 5.1 基本方針

**座標の変換は行わず、ソート処理で座標系の向きを考慮する**

理由：
1. テキストマトリックスの座標は既に正しい空間にある
2. 問題はソート順序が間違っていること
3. Y軸が反転している場合、ソート順序を反転させれば良い

### 5.2 具体的な設計

#### 5.2.1 PageLayoutにCTM情報を追加
```go
type PageLayout struct {
    PageNum    int
    Width      float64
    Height     float64
    TextBlocks []TextBlock
    Images     []ImageBlock
    PageCTM    *Matrix  // ページレベルのCTM
}
```

#### 5.2.2 ページレベルのCTMを記録
TextExtractorで、最初の`cm`オペレータ（かつq/Qスタックが空の時）を記録：
```go
type TextExtractor struct {
    ...
    pageLevelCTM *Matrix  // ページレベルのCTM（最初のcm）
}
```

#### 5.2.3 Y軸反転を検出してソート順序を調整
```go
func (pl *PageLayout) SortedContentBlocks() []ContentBlock {
    blocks := pl.ContentBlocks()

    // Y軸が反転しているかチェック（CTMのd成分が負）
    yAxisFlipped := false
    if pl.PageCTM != nil && pl.PageCTM.D < 0 {
        yAxisFlipped = true
    }

    sort.Slice(blocks, func(i, j int) bool {
        topI := blocks[i].Bounds().Y + blocks[i].Bounds().Height
        topJ := blocks[j].Bounds().Y + blocks[j].Bounds().Height

        if yAxisFlipped {
            return topI < topJ  // Y軸反転時：Y値が小さい方が視覚的に上
        }
        return topI > topJ      // 標準：Y値が大きい方が視覚的に上
    })

    return blocks
}
```

### 5.3 設計の利点

1. **シンプル**: 座標変換の複雑な計算が不要
2. **正確**: 抽出された座標をそのまま使用
3. **拡張性**: 他の種類のCTM（回転など）にも対応可能
4. **後方互換性**: 標準座標系のPDFも正しく動作

## 6. 実装計画

### 6.1 ステップ1: PageLayoutの拡張
- [x] `PageLayout`に`PageCTM`フィールドを追加
- [ ] 型定義の整合性を確認

### 6.2 ステップ2: CTMの記録
- [x] `TextExtractor`に`pageLevelCTM`フィールドを追加
- [x] `cm`オペレータ処理でページレベルCTMを記録
- [x] `GetPageLevelCTM()`メソッドを追加

### 6.3 ステップ3: PageLayoutへのCTM設定
- [x] `ExtractPageLayout()`で`TextExtractor.GetPageLevelCTM()`を取得
- [x] `PageLayout.PageCTM`に設定

### 6.4 ステップ4: ソート処理の修正
- [x] `SortedContentBlocks()`でY軸反転を検出
- [x] Y軸反転時にソート順序を反転

### 6.5 ステップ5: RenderLayoutでの座標変換
- [x] `RenderLayout()`でY軸反転を検出
- [x] Y軸反転時に座標を標準座標系に変換
- [x] 変換式: `newY = pageHeight - oldY - blockHeight`
- [x] `setPageFont()`の型アサーション修正（font.StandardFont → StandardFont）

### 6.6 ステップ6: テスト
- [x] Receipt-2021-3422.pdfで動作確認 ✅
- [x] 既存のテストが全て通ることを確認 ✅
- [x] 標準座標系のPDFでも正しく動作することを確認 ✅

### 6.7 ステップ7: ドキュメント更新
- [ ] coordinate_system.mdを更新
- [ ] 変更内容をcommit

## 7. テストケース

### 7.1 Receipt-2021-3422.pdf（Y軸反転）
期待される出力順序（上から下）：
1. "Date paid October 8, 2025..." (Y=13.0, 視覚的に上部)
2. "billing@suno.com..." (Y=120.0)
3. ...
12. "Page 1 of 1" (Y=754.0, 視覚的に下部)

### 7.2 標準座標系のPDF
既存のテストケースが全て通ること

## 8. 参考資料

- PDF Reference 1.7, Section 4.2.2 "Common Transformations"
- PDF Reference 1.7, Section 8.3 "Text State Parameters and Operators"
- docs/ctm_coordinate_transformation.md（既存の調査結果）

## 9. 今後の拡張可能性

### 9.1 回転への対応
CTMで回転が適用されている場合（bやcが非ゼロ）の処理

### 9.2 複雑な変換への対応
複数のCTMが組み合わさっている場合の処理

### 9.3 テキストマトリックスとCTMの統合
より正確な座標計算が必要な場合の対応
