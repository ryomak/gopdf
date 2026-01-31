# 図形描画機能 設計書

## 概要

Phase 3では、PDFの基本的な図形描画機能を実装します。PDF 1.7仕様のグラフィックスオペレーターを使用して、線、矩形、円などの図形を描画できるようにします。

## 要件（FR-1.5より）

- 基本的な図形（線、矩形、円、ベジェ曲線など）を描画できる
- 線の太さ、色、塗りつぶしの色を指定できる

## PDF仕様

### グラフィックス状態オペレーター

| オペレーター | 説明 | 例 |
|------------|------|-----|
| `w` | 線の太さ（Line width） | `2.0 w` (2ポイント) |
| `J` | 線の端の形状（Line cap style）| `0 J` (Butt cap), `1 J` (Round cap), `2 J` (Square cap) |
| `j` | 線の結合形状（Line join style）| `0 j` (Miter join), `1 j` (Round join), `2 j` (Bevel join) |
| `RG` | ストローク色（RGB）| `1.0 0.0 0.0 RG` (赤) |
| `rg` | 塗りつぶし色（RGB）| `0.0 1.0 0.0 rg` (緑) |

### パス構築オペレーター

| オペレーター | 説明 | 例 |
|------------|------|-----|
| `m` | moveto（パス開始点）| `100 100 m` |
| `l` | lineto（直線）| `200 200 l` |
| `re` | rectangle（矩形パス）| `100 100 200 150 re` (x, y, width, height) |
| `c` | curveto（3次ベジェ曲線）| `x1 y1 x2 y2 x3 y3 c` |
| `h` | closepath（パスを閉じる）| `h` |

### パス描画オペレーター

| オペレーター | 説明 | 用途 |
|------------|------|------|
| `S` | stroke（線を描画）| 枠線のみ |
| `s` | close and stroke | パスを閉じて線を描画 |
| `f` | fill（塗りつぶし）| 塗りつぶしのみ |
| `F` | fill（f と同じ）| 塗りつぶしのみ |
| `B` | fill and stroke | 塗りつぶし＋枠線 |
| `b` | close, fill and stroke | パスを閉じて塗りつぶし＋枠線 |

### 円の描画

PDFには直接円を描く命令がないため、4つのベジェ曲線で近似します。
近似式: `κ = 4 * (√2 - 1) / 3 ≈ 0.5522847498`

## API設計

### 色の表現

```go
// Color represents an RGB color in PDF
type Color struct {
    R, G, B float64 // 0.0 〜 1.0
}

// 定義済み色
var (
    Black   = Color{0, 0, 0}
    White   = Color{1, 1, 1}
    Red     = Color{1, 0, 0}
    Green   = Color{0, 1, 0}
    Blue    = Color{0, 0, 1}
)

// NewRGB creates a color from 8-bit RGB values (0-255)
func NewRGB(r, g, b uint8) Color
```

### 線スタイルの表現

```go
// LineCapStyle represents the shape at the ends of lines
type LineCapStyle int

const (
    ButtCap   LineCapStyle = 0 // 端を切り落とす
    RoundCap  LineCapStyle = 1 // 端を丸める
    SquareCap LineCapStyle = 2 // 端を四角く延長
)

// LineJoinStyle represents the shape at corners of paths
type LineJoinStyle int

const (
    MiterJoin LineJoinStyle = 0 // 尖った結合
    RoundJoin LineJoinStyle = 1 // 丸い結合
    BevelJoin LineJoinStyle = 2 // 面取り結合
)
```

### Pageへのメソッド追加

```go
// グラフィックス状態の設定
func (p *Page) SetLineWidth(width float64)
func (p *Page) SetStrokeColor(c Color)
func (p *Page) SetFillColor(c Color)
func (p *Page) SetLineCap(cap LineCapStyle)
func (p *Page) SetLineJoin(join LineJoinStyle)

// 基本図形の描画
func (p *Page) DrawLine(x1, y1, x2, y2 float64)
func (p *Page) DrawRectangle(x, y, width, height float64)
func (p *Page) FillRectangle(x, y, width, height float64)
func (p *Page) DrawAndFillRectangle(x, y, width, height float64)
func (p *Page) DrawCircle(centerX, centerY, radius float64)
func (p *Page) FillCircle(centerX, centerY, radius float64)
func (p *Page) DrawAndFillCircle(centerX, centerY, radius float64)

// 低レベルパス操作（将来的な拡張用）
func (p *Page) MoveTo(x, y float64)
func (p *Page) LineTo(x, y float64)
func (p *Page) ClosePath()
func (p *Page) Stroke()
func (p *Page) Fill()
func (p *Page) FillAndStroke()
```

## 実装例

### 線の描画

```go
page.SetLineWidth(2.0)
page.SetStrokeColor(gopdf.Red)
page.DrawLine(100, 100, 300, 200)
```

生成されるPDFコンテンツストリーム:
```
2.00 w
1.00 0.00 0.00 RG
100.00 100.00 m
300.00 200.00 l
S
```

### 矩形の描画

```go
// 枠線のみ
page.SetLineWidth(1.5)
page.SetStrokeColor(gopdf.Blue)
page.DrawRectangle(100, 200, 150, 100)

// 塗りつぶしのみ
page.SetFillColor(gopdf.Color{R: 1.0, G: 1.0, B: 0.0}) // 黄色
page.FillRectangle(300, 200, 150, 100)

// 枠線＋塗りつぶし
page.SetStrokeColor(gopdf.Black)
page.SetFillColor(gopdf.Color{R: 0.8, G: 0.8, B: 0.8}) // 灰色
page.DrawAndFillRectangle(500, 200, 150, 100)
```

### 円の描画

```go
page.SetStrokeColor(gopdf.Red)
page.SetFillColor(gopdf.Color{R: 1.0, G: 0.8, B: 0.8}) // 薄い赤
page.DrawAndFillCircle(300, 400, 50) // 中心(300, 400)、半径50
```

## 実装計画

### Phase 3.1: 色と線スタイル
1. Color型の定義 (pkg/gopdf/graphics.go)
2. LineCapStyle, LineJoinStyle型の定義
3. Pageへの状態管理フィールド追加
4. SetLineWidth, SetStrokeColor, SetFillColor メソッド
5. テスト作成と実装（TDD）

### Phase 3.2: 線と矩形
1. DrawLine メソッド実装
2. DrawRectangle, FillRectangle, DrawAndFillRectangle メソッド実装
3. テスト作成と実装（TDD）

### Phase 3.3: 円
1. 円のベジェ曲線近似アルゴリズム実装
2. DrawCircle, FillCircle, DrawAndFillCircle メソッド実装
3. テスト作成と実装（TDD）

### Phase 3.4: 統合テストとサンプル
1. examples/03_graphics サンプルコード作成
2. 統合テスト（実際にPDFを生成して確認）
3. ドキュメント更新（README.md, docs/progress.md）

## テスト戦略

### ユニットテスト
- 各メソッドが正しいPDFオペレーターを生成することを確認
- 色値の範囲チェック（0.0〜1.0）
- 線の太さの妥当性確認

### 統合テスト
- 実際にPDFを生成し、以下を確認：
  - PDF Reader で開けること
  - 図形が期待通りに描画されること
  - 色が正しく表示されること

## 参考資料

- PDF 1.7 仕様書 8章: Graphics
  - 8.2: Graphics State
  - 8.5: Path Construction and Painting
- PDF Reference 6th Edition (Adobe)
  - Appendix D: Operator Summary
