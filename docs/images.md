# 画像の埋め込み

JPEG/PNG画像をPDFに埋め込む方法を説明します。

## 対応フォーマット

| フォーマット | 特徴 |
|-------------|------|
| JPEG | 写真に適した非可逆圧縮、ファイルサイズが小さい |
| PNG | 透明度（アルファチャンネル）対応、可逆圧縮 |

## JPEG画像の埋め込み

### ファイルから読み込み

```go
// ファイルを開く
file, err := os.Open("photo.jpg")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

// JPEG画像を読み込み
img, err := gopdf.LoadJPEG(file)
if err != nil {
    log.Fatal(err)
}

// ページに描画
page.DrawImage(img, 100, 600, 200, 150)
```

### バイト配列から読み込み

```go
// バイト配列として読み込み
jpegData, _ := os.ReadFile("photo.jpg")

// 画像を読み込み
img, err := gopdf.LoadJPEG(bytes.NewReader(jpegData))
if err != nil {
    log.Fatal(err)
}

page.DrawImage(img, 100, 600, 200, 150)
```

### プログラムで生成した画像

```go
import (
    "bytes"
    "image"
    "image/color"
    "image/jpeg"
)

// 画像を生成
goImg := image.NewRGBA(image.Rect(0, 0, 200, 100))
for x := 0; x < 200; x++ {
    for y := 0; y < 100; y++ {
        goImg.Set(x, y, color.RGBA{
            R: uint8(x),
            G: uint8(y),
            B: 128,
            A: 255,
        })
    }
}

// JPEGにエンコード
var buf bytes.Buffer
jpeg.Encode(&buf, goImg, &jpeg.Options{Quality: 90})

// PDFに埋め込み
img, _ := gopdf.LoadJPEG(&buf)
page.DrawImage(img, 100, 600, 200, 100)
```

## PNG画像の埋め込み

PNGは透明度（アルファチャンネル）をサポートしています。

### ファイルから読み込み

```go
file, err := os.Open("logo.png")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

img, err := gopdf.LoadPNG(file)
if err != nil {
    log.Fatal(err)
}

page.DrawImage(img, 100, 600, 200, 200)
```

### 透明度のある画像

```go
import (
    "bytes"
    "image"
    "image/color"
    "image/png"
)

// 透明度のある画像を生成
goImg := image.NewRGBA(image.Rect(0, 0, 200, 100))
for x := 0; x < 200; x++ {
    for y := 0; y < 100; y++ {
        // 中心からの距離で透明度を決定
        alpha := uint8(255 - (x+y)/2)
        goImg.Set(x, y, color.RGBA{
            R: 255,
            G: 0,
            B: 0,
            A: alpha,
        })
    }
}

// PNGにエンコード
var buf bytes.Buffer
png.Encode(&buf, goImg)

// PDFに埋め込み
img, _ := gopdf.LoadPNG(&buf)
page.DrawImage(img, 100, 600, 200, 100)
```

## 画像の配置

### DrawImage関数

```go
page.DrawImage(image, x, y, width, height)
```

| パラメータ | 説明 |
|-----------|------|
| `image` | 読み込んだ画像 |
| `x` | 画像の左下X座標 |
| `y` | 画像の左下Y座標 |
| `width` | 表示幅（ポイント） |
| `height` | 表示高さ（ポイント） |

### アスペクト比を維持

```go
// 元の画像サイズを取得
originalWidth := 1920.0  // ピクセル
originalHeight := 1080.0

// 表示したい幅
targetWidth := 400.0

// アスペクト比を維持して高さを計算
targetHeight := targetWidth * (originalHeight / originalWidth)

page.DrawImage(img, 100, 500, targetWidth, targetHeight)
```

### 画像を中央に配置

```go
pageWidth := 595.0  // A4幅
imgWidth := 200.0
x := (pageWidth - imgWidth) / 2

page.DrawImage(img, x, 500, imgWidth, 150)
```

## 画像の重ね合わせ

PNG画像の透明度を利用して、画像を重ね合わせることができます：

```go
// 背景画像（JPEG）
bgImg, _ := gopdf.LoadJPEG(bgFile)
page.DrawImage(bgImg, 50, 500, 500, 300)

// 透明なロゴ（PNG）を重ねる
logoImg, _ := gopdf.LoadPNG(logoFile)
page.DrawImage(logoImg, 400, 520, 100, 100)
```

## 実践例：写真ギャラリー

```go
package main

import (
    "os"
    "github.com/ryomak/gopdf"
)

func main() {
    doc := gopdf.New()
    page := doc.AddPage(gopdf.PageSizeA4, gopdf.Portrait)

    // タイトル
    page.SetFont(gopdf.FontHelveticaBold, 24)
    page.DrawText("Photo Gallery", 50, 780)

    // 画像ファイル
    photos := []string{"photo1.jpg", "photo2.jpg", "photo3.jpg", "photo4.jpg"}

    // グリッド配置
    cols := 2
    imgWidth := 250.0
    imgHeight := 180.0
    startX := 50.0
    startY := 550.0
    padding := 20.0

    for i, photoPath := range photos {
        file, err := os.Open(photoPath)
        if err != nil {
            continue
        }

        img, err := gopdf.LoadJPEG(file)
        file.Close()
        if err != nil {
            continue
        }

        col := i % cols
        row := i / cols
        x := startX + float64(col)*(imgWidth+padding)
        y := startY - float64(row)*(imgHeight+padding)

        page.DrawImage(img, x, y, imgWidth, imgHeight)
    }

    file, _ := os.Create("gallery.pdf")
    defer file.Close()
    doc.WriteTo(file)
}
```

## サンプルコード

- [examples/04_images](https://github.com/ryomak/gopdf/tree/main/examples/04_images) - JPEG画像の埋め込み
- [examples/05_png_images](https://github.com/ryomak/gopdf/tree/main/examples/05_png_images) - PNG画像（透明度対応）

## 次のステップ

- [PDF読み込み・解析](reading-pdf.md) - 既存PDFからの画像抽出
- [PDF翻訳](translation.md) - 画像を保持した翻訳
