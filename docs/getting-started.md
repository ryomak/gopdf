# はじめに

gopdfのインストールから基本的な使い方までを説明します。

## 動作環境

- Go 1.18以上
- Pure Go（CGO不要）

## インストール

```bash
go get github.com/ryomak/gopdf
```

## 基本的な使い方

### PDF作成の基本フロー

1. ドキュメントを作成
2. ページを追加
3. コンテンツを描画
4. ファイルに出力

```go
package main

import (
    "os"
    "log"
    "github.com/ryomak/gopdf"
)

func main() {
    // 1. ドキュメントを作成
    doc := gopdf.New()

    // 2. ページを追加
    page := doc.AddPage(gopdf.PageSizeA4, gopdf.Portrait)

    // 3. コンテンツを描画
    page.SetFont(gopdf.FontHelvetica, 24)
    page.DrawText("Hello, World!", 100, 750)

    // 4. ファイルに出力
    file, err := os.Create("output.pdf")
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    if err := doc.WriteTo(file); err != nil {
        log.Fatal(err)
    }
}
```

## ページサイズと向き

### ページサイズの指定

```go
// A4サイズ（縦向き）
page := doc.AddPage(gopdf.PageSizeA4, gopdf.Portrait)

// Letterサイズ（横向き）
page := doc.AddPage(gopdf.PageSizeLetter, gopdf.Landscape)

// プレゼンテーション用（16:9）
page := doc.AddPage(gopdf.PageSizePresentation16x9, gopdf.Portrait)
```

### 利用可能なページサイズ

| 定数 | サイズ (pt) | サイズ (mm/inch) |
|------|-------------|------------------|
| `PageSizeA4` | 595 x 842 | 210mm x 297mm |
| `PageSizeLetter` | 612 x 792 | 8.5" x 11" |
| `PageSizeLegal` | 612 x 1008 | 8.5" x 14" |
| `PageSizeA3` | 842 x 1191 | 297mm x 420mm |
| `PageSizeA5` | 420 x 595 | 148mm x 210mm |
| `PageSizePresentation16x9` | 720 x 405 | 16:9 |
| `PageSizePresentation4x3` | 720 x 540 | 4:3 |

### 向き

| 定数 | 説明 |
|------|------|
| `Portrait` | 縦向き |
| `Landscape` | 横向き |

## 複数ページの作成

```go
doc := gopdf.New()

// ページ1
page1 := doc.AddPage(gopdf.PageSizeA4, gopdf.Portrait)
page1.SetFont(gopdf.FontHelvetica, 24)
page1.DrawText("Page 1", 100, 750)

// ページ2
page2 := doc.AddPage(gopdf.PageSizeA4, gopdf.Portrait)
page2.SetFont(gopdf.FontHelvetica, 24)
page2.DrawText("Page 2", 100, 750)

// ページ数を取得
fmt.Printf("Total pages: %d\n", doc.PageCount())
```

## 座標系について

gopdfの座標系は、PDFの標準座標系に従います：

- **原点 (0, 0)**: ページの左下
- **X軸**: 右方向が正
- **Y軸**: 上方向が正
- **単位**: ポイント (pt)、1pt = 1/72 inch

```
        ↑ Y (上方向が正)
        |
        |
        |  (100, 750) ← テキストの描画位置
        |
        |
(0,0)---+---------------→ X (右方向が正)
        左下が原点
```

### 単位変換

```go
// ミリメートルからポイントへ
func mmToPt(mm float64) float64 {
    return mm * 72.0 / 25.4
}

// インチからポイントへ
func inchToPt(inch float64) float64 {
    return inch * 72.0
}
```

## 出力方法

### ファイルに出力

```go
file, err := os.Create("output.pdf")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

doc.WriteTo(file)
```

### バイト配列に出力

```go
var buf bytes.Buffer
doc.WriteTo(&buf)
pdfData := buf.Bytes()
```

### HTTPレスポンスに出力

```go
func handler(w http.ResponseWriter, r *http.Request) {
    doc := gopdf.New()
    // ... PDFを作成 ...

    w.Header().Set("Content-Type", "application/pdf")
    w.Header().Set("Content-Disposition", "attachment; filename=document.pdf")
    doc.WriteTo(w)
}
```

## 次のステップ

- [テキストとフォント](text-and-fonts.md) - テキスト描画の詳細
- [図形描画](graphics.md) - 線、矩形、円の描画
- [画像の埋め込み](images.md) - JPEG/PNG画像の使用
