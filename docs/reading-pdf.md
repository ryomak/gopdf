# PDF読み込み・解析

既存のPDFを読み込み、テキストや画像を抽出する方法を説明します。

## PDFを開く

### ファイルパスから

```go
reader, err := gopdf.Open("document.pdf")
if err != nil {
    log.Fatal(err)
}
```

### io.Readerから

```go
file, _ := os.Open("document.pdf")
defer file.Close()

reader, err := gopdf.OpenReader(file)
if err != nil {
    log.Fatal(err)
}
```

## 基本情報の取得

### ページ数

```go
pageCount := reader.PageCount()
fmt.Printf("Total pages: %d\n", pageCount)
```

### メタデータ

```go
info := reader.Info()

fmt.Printf("Title: %s\n", info.Title)
fmt.Printf("Author: %s\n", info.Author)
fmt.Printf("Subject: %s\n", info.Subject)
fmt.Printf("Keywords: %s\n", info.Keywords)
fmt.Printf("Creator: %s\n", info.Creator)
fmt.Printf("Producer: %s\n", info.Producer)
```

## テキスト抽出

### 全ページからテキスト抽出

```go
text, err := reader.ExtractText()
if err != nil {
    log.Fatal(err)
}
fmt.Println(text)
```

### 特定ページからテキスト抽出

```go
// ページ番号は0から始まる
text, err := reader.ExtractPageText(0) // 最初のページ
if err != nil {
    log.Fatal(err)
}
fmt.Println(text)
```

## 位置情報付きテキスト抽出

テキストの位置、サイズ、フォント情報を取得できます。

### 特定ページから

```go
elements, err := reader.ExtractPageTextElements(0)
if err != nil {
    log.Fatal(err)
}

for _, elem := range elements {
    fmt.Printf("Text: %s\n", elem.Text)
    fmt.Printf("  Position: (%.2f, %.2f)\n", elem.X, elem.Y)
    fmt.Printf("  Size: %.2f x %.2f\n", elem.Width, elem.Height)
    fmt.Printf("  Font: %s, Size: %.2f\n", elem.Font, elem.Size)
}
```

### 全ページから

```go
allElements, err := reader.ExtractAllTextElements()
if err != nil {
    log.Fatal(err)
}

for pageNum, elements := range allElements {
    fmt.Printf("=== Page %d ===\n", pageNum+1)
    for _, elem := range elements {
        fmt.Printf("  %s at (%.2f, %.2f)\n", elem.Text, elem.X, elem.Y)
    }
}
```

### 読み順にソート

```go
elements, _ := reader.ExtractPageTextElements(0)

// 上から下、左から右の順にソート
sorted := gopdf.SortTextElements(elements)

// 文字列に変換
text := gopdf.TextElementsToString(sorted)
fmt.Println(text)
```

## 画像抽出

### 全ページから画像抽出

```go
allImages, err := reader.ExtractAllImages()
if err != nil {
    log.Fatal(err)
}

for pageNum, images := range allImages {
    fmt.Printf("Page %d: %d images\n", pageNum+1, len(images))

    for i, img := range images {
        fmt.Printf("  Image %d:\n", i+1)
        fmt.Printf("    Name: %s\n", img.Name)
        fmt.Printf("    Size: %d x %d\n", img.Width, img.Height)
        fmt.Printf("    ColorSpace: %s\n", img.ColorSpace)
        fmt.Printf("    Format: %s\n", img.Format)
    }
}
```

### 画像の保存

```go
allImages, _ := reader.ExtractAllImages()

for pageNum, images := range allImages {
    for i, img := range images {
        // 生データとして保存
        filename := fmt.Sprintf("page%d_image%d.raw", pageNum+1, i+1)
        img.SaveImage(filename)

        // または、Go標準のimage.Imageに変換
        stdImg, err := img.ToImage()
        if err != nil {
            continue
        }

        // JPEGとして保存
        jpegFile, _ := os.Create(fmt.Sprintf("page%d_image%d.jpg", pageNum+1, i+1))
        jpeg.Encode(jpegFile, stdImg, nil)
        jpegFile.Close()
    }
}
```

## レイアウト抽出

テキストブロックと画像の位置情報を含むレイアウトを抽出できます。PDF翻訳などに使用します。

```go
layouts, err := reader.ExtractAllLayouts()
if err != nil {
    log.Fatal(err)
}

for pageNum, layout := range layouts {
    fmt.Printf("=== Page %d ===\n", pageNum+1)
    fmt.Printf("Size: %.2f x %.2f\n", layout.Width, layout.Height)

    // テキストブロック
    for _, block := range layout.TextBlocks {
        fmt.Printf("Text: %s\n", block.Text)
        fmt.Printf("  Position: (%.2f, %.2f)\n", block.X, block.Y)
        fmt.Printf("  Size: %.2f x %.2f\n", block.Width, block.Height)
    }

    // 画像ブロック
    for _, img := range layout.ImageBlocks {
        fmt.Printf("Image at (%.2f, %.2f)\n", img.X, img.Y)
        fmt.Printf("  Size: %.2f x %.2f\n", img.Width, img.Height)
    }
}
```

## 暗号化PDFの読み込み

暗号化されたPDFを読み込む場合は、パスワードで認証が必要です。

```go
file, _ := os.Open("encrypted.pdf")
defer file.Close()

reader, err := gopdf.OpenReader(file)
if err != nil {
    log.Fatal(err)
}

// 暗号化チェック
if reader.IsEncrypted() {
    // パスワードで認証
    err := reader.AuthenticateWithPassword("password123")
    if err != nil {
        log.Fatal("Invalid password")
    }
}

// 認証後、通常通り読み込み
text, _ := reader.ExtractText()
fmt.Println(text)
```

## 完全な例

```go
package main

import (
    "fmt"
    "log"
    "github.com/ryomak/gopdf"
)

func main() {
    reader, err := gopdf.Open("sample.pdf")
    if err != nil {
        log.Fatal(err)
    }

    // 基本情報
    fmt.Printf("Pages: %d\n", reader.PageCount())

    info := reader.Info()
    if info.Title != "" {
        fmt.Printf("Title: %s\n", info.Title)
    }
    if info.Author != "" {
        fmt.Printf("Author: %s\n", info.Author)
    }

    // テキスト抽出
    fmt.Println("\n=== Text Content ===")
    for i := 0; i < reader.PageCount(); i++ {
        text, err := reader.ExtractPageText(i)
        if err != nil {
            continue
        }
        fmt.Printf("--- Page %d ---\n%s\n", i+1, text)
    }

    // 画像抽出
    fmt.Println("\n=== Images ===")
    allImages, _ := reader.ExtractAllImages()
    for pageNum, images := range allImages {
        for i, img := range images {
            fmt.Printf("Page %d, Image %d: %dx%d (%s)\n",
                pageNum+1, i+1, img.Width, img.Height, img.ColorSpace)
        }
    }
}
```

## サンプルコード

- [examples/06_read_pdf](https://github.com/ryomak/gopdf/tree/main/examples/06_read_pdf) - PDF読み込み・テキスト抽出
- [examples/07_structured_text](https://github.com/ryomak/gopdf/tree/main/examples/07_structured_text) - 位置情報付きテキスト抽出
- [examples/08_extract_images](https://github.com/ryomak/gopdf/tree/main/examples/08_extract_images) - 画像抽出
- [examples/14_decrypt_pdf](https://github.com/ryomak/gopdf/tree/main/examples/14_decrypt_pdf) - 暗号化PDF復号

## 次のステップ

- [PDF翻訳](translation.md) - レイアウト保持翻訳
- [メタデータ](metadata.md) - メタデータの詳細
