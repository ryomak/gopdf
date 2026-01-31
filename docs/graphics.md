# 図形描画

線、矩形、円などの図形を描画する方法を説明します。

## 色の設定

### 定義済みの色

```go
gopdf.ColorBlack   // 黒
gopdf.ColorWhite   // 白
gopdf.ColorRed     // 赤
gopdf.ColorGreen   // 緑
gopdf.ColorBlue    // 青
```

### カスタム色

RGB値（0.0〜1.0）で色を指定します：

```go
// 赤
red := gopdf.Color{R: 1.0, G: 0.0, B: 0.0}

// 黄色
yellow := gopdf.Color{R: 1.0, G: 1.0, B: 0.0}

// 紫
purple := gopdf.Color{R: 0.5, G: 0.0, B: 0.5}

// グレー（50%）
gray := gopdf.Color{R: 0.5, G: 0.5, B: 0.5}
```

### 色の適用

```go
// 線の色を設定
page.SetStrokeColor(gopdf.ColorRed)

// 塗りつぶしの色を設定
page.SetFillColor(gopdf.Color{R: 1.0, G: 1.0, B: 0.0})
```

## 線の描画

### 基本的な線

```go
// 線の太さを設定
page.SetLineWidth(2.0)

// 線の色を設定
page.SetStrokeColor(gopdf.ColorBlack)

// (x1, y1) から (x2, y2) へ線を描画
page.DrawLine(100, 700, 500, 700)
```

### 線のスタイル

```go
// 線の端のスタイル
page.SetLineCap(gopdf.LineCapButt)   // 端が平ら（デフォルト）
page.SetLineCap(gopdf.LineCapRound)  // 端が丸い
page.SetLineCap(gopdf.LineCapSquare) // 端が四角く突き出る
```

### 複数の線を描画

```go
page.SetLineWidth(1.0)
page.SetStrokeColor(gopdf.ColorBlue)

// 複数の平行線
for i := 0; i < 5; i++ {
    y := 700.0 - float64(i)*20
    page.DrawLine(100, y, 500, y)
}
```

## 矩形（四角形）の描画

### 枠線のみ

```go
page.SetLineWidth(2.0)
page.SetStrokeColor(gopdf.ColorBlack)

// (x, y) を左下として、幅と高さを指定
page.DrawRectangle(100, 600, 200, 100)
```

### 塗りつぶしのみ

```go
page.SetFillColor(gopdf.Color{R: 0.9, G: 0.9, B: 0.9})

// 塗りつぶした矩形
page.FillRectangle(100, 500, 200, 100)
```

### 枠線と塗りつぶし

```go
page.SetLineWidth(2.0)
page.SetStrokeColor(gopdf.ColorBlack)
page.SetFillColor(gopdf.Color{R: 1.0, G: 1.0, B: 0.0})

// 枠線付きで塗りつぶした矩形
page.DrawAndFillRectangle(100, 400, 200, 100)
```

## 円の描画

### 枠線のみ

```go
page.SetLineWidth(2.0)
page.SetStrokeColor(gopdf.ColorRed)

// (cx, cy) を中心として、半径 r の円
page.DrawCircle(200, 300, 50)
```

### 塗りつぶしのみ

```go
page.SetFillColor(gopdf.ColorBlue)

// 塗りつぶした円
page.FillCircle(200, 200, 50)
```

### 枠線と塗りつぶし

```go
page.SetLineWidth(2.0)
page.SetStrokeColor(gopdf.ColorBlack)
page.SetFillColor(gopdf.ColorGreen)

// 枠線付きで塗りつぶした円
page.DrawAndFillCircle(200, 100, 50)
```

## 実践例：グラデーション風の表現

```go
// 青から緑へのグラデーション風の矩形
for i := 0; i < 10; i++ {
    // 色を徐々に変化
    ratio := float64(i) / 10.0
    color := gopdf.Color{
        R: 0.0,
        G: ratio,
        B: 1.0 - ratio,
    }

    page.SetFillColor(color)
    x := 100.0 + float64(i)*30
    page.FillRectangle(x, 500, 30, 100)
}
```

## 実践例：チャートの描画

```go
// 棒グラフ
data := []float64{80, 150, 100, 180, 120}
colors := []gopdf.Color{
    {R: 1.0, G: 0.2, B: 0.2},
    {R: 0.2, G: 1.0, B: 0.2},
    {R: 0.2, G: 0.2, B: 1.0},
    {R: 1.0, G: 1.0, B: 0.2},
    {R: 1.0, G: 0.2, B: 1.0},
}

barWidth := 50.0
spacing := 20.0
baseY := 200.0

for i, value := range data {
    x := 100.0 + float64(i)*(barWidth+spacing)
    page.SetFillColor(colors[i])
    page.FillRectangle(x, baseY, barWidth, value)
}
```

## 完全な例

```go
package main

import (
    "os"
    "github.com/ryomak/gopdf"
)

func main() {
    doc := gopdf.New()
    page := doc.AddPage(gopdf.PageSizeA4, gopdf.Portrait)

    // 背景色付きの矩形
    page.SetFillColor(gopdf.Color{R: 0.95, G: 0.95, B: 0.95})
    page.FillRectangle(50, 700, 500, 100)

    // タイトル
    page.SetFont(gopdf.FontHelveticaBold, 24)
    page.DrawText("Graphics Demo", 60, 760)

    // 線
    page.SetLineWidth(2.0)
    page.SetStrokeColor(gopdf.ColorRed)
    page.DrawLine(50, 680, 550, 680)

    // 矩形
    page.SetStrokeColor(gopdf.ColorBlue)
    page.SetFillColor(gopdf.Color{R: 0.8, G: 0.9, B: 1.0})
    page.DrawAndFillRectangle(50, 550, 150, 100)

    // 円
    page.SetStrokeColor(gopdf.ColorGreen)
    page.SetFillColor(gopdf.Color{R: 0.9, G: 1.0, B: 0.8})
    page.DrawAndFillCircle(350, 600, 60)

    file, _ := os.Create("graphics_demo.pdf")
    defer file.Close()
    doc.WriteTo(file)
}
```

## サンプルコード

完全な例は[examples/03_graphics](https://github.com/ryomak/gopdf/tree/main/examples/03_graphics)を参照してください。

## 次のステップ

- [画像の埋め込み](images.md) - JPEG/PNG画像の使用
- [PDF読み込み・解析](reading-pdf.md) - 既存PDFの読み込み
