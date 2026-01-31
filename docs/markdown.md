# Markdown変換

MarkdownファイルをPDFに変換する方法を説明します。

## 概要

gopdfは、Markdownテキストを直接PDFに変換できます。以下の要素をサポートしています：

- 見出し（H1〜H6）
- 段落
- 箇条書きリスト
- 番号付きリスト
- コードブロック
- 引用
- 太字・斜体
- 画像

## 基本的な使い方

### 文字列から変換

```go
markdown := `# タイトル

これは段落です。

## セクション1

- 項目1
- 項目2
- 項目3

## セクション2

1. 番号付き1
2. 番号付き2
`

doc, err := gopdf.NewMarkdownDocument(markdown, nil)
if err != nil {
    log.Fatal(err)
}

file, _ := os.Create("output.pdf")
defer file.Close()
doc.WriteTo(file)
```

### ファイルから変換

```go
doc, err := gopdf.NewMarkdownDocumentFromFile("document.md", nil)
if err != nil {
    log.Fatal(err)
}

file, _ := os.Create("output.pdf")
defer file.Close()
doc.WriteTo(file)
```

## MarkdownOptions

変換動作をカスタマイズするオプション：

```go
type MarkdownOptions struct {
    Mode        MarkdownMode   // 変換モード
    PageSize    PageSize       // ページサイズ
    Orientation Orientation    // 向き
    Style       *MarkdownStyle // スタイル設定
}
```

### オプションの使用例

```go
opts := &gopdf.MarkdownOptions{
    Mode:        gopdf.MarkdownModeDocument,
    PageSize:    gopdf.PageSizeA4,
    Orientation: gopdf.Portrait,
}

doc, err := gopdf.NewMarkdownDocument(markdown, opts)
```

## 変換モード

### MarkdownModeDocument

通常のドキュメントとして変換（A4などの用紙サイズ）：

```go
opts := &gopdf.MarkdownOptions{
    Mode:     gopdf.MarkdownModeDocument,
    PageSize: gopdf.PageSizeA4,
}
```

### MarkdownModeSlide

プレゼンテーションスライドとして変換：

```go
opts := &gopdf.MarkdownOptions{
    Mode:     gopdf.MarkdownModeSlide,
    PageSize: gopdf.PageSizePresentation16x9,
}
```

スライドモードでは、`---`（水平線）でページが区切られます。

## スタイルのカスタマイズ

`MarkdownStyle`で細かいスタイルを設定できます：

```go
style := gopdf.DefaultMarkdownStyle()

// 見出しサイズ
style.H1Size = 42
style.H2Size = 32
style.H3Size = 24

// 本文
style.BodySize = 14
style.LineSpacing = 1.5

// 色
style.HeadingColor = gopdf.Color{R: 0.1, G: 0.3, B: 0.7}
style.BodyColor = gopdf.ColorBlack
style.CodeBackgroundColor = gopdf.Color{R: 0.95, G: 0.95, B: 0.95}

// マージン
style.MarginTop = 72
style.MarginBottom = 72
style.MarginLeft = 72
style.MarginRight = 72

opts := &gopdf.MarkdownOptions{
    Mode:  gopdf.MarkdownModeDocument,
    Style: style,
}
```

### 利用可能なスタイル設定

| プロパティ | 説明 | デフォルト |
|-----------|------|-----------|
| `H1Size` | H1見出しのフォントサイズ | 32 |
| `H2Size` | H2見出しのフォントサイズ | 24 |
| `H3Size` | H3見出しのフォントサイズ | 20 |
| `H4Size` | H4見出しのフォントサイズ | 16 |
| `H5Size` | H5見出しのフォントサイズ | 14 |
| `H6Size` | H6見出しのフォントサイズ | 12 |
| `BodySize` | 本文のフォントサイズ | 12 |
| `CodeSize` | コードのフォントサイズ | 10 |
| `LineSpacing` | 行間 | 1.2 |
| `HeadingColor` | 見出しの色 | 黒 |
| `BodyColor` | 本文の色 | 黒 |
| `CodeBackgroundColor` | コードブロックの背景色 | 薄いグレー |
| `MarginTop` | 上マージン（pt） | 72 |
| `MarginBottom` | 下マージン（pt） | 72 |
| `MarginLeft` | 左マージン（pt） | 72 |
| `MarginRight` | 右マージン（pt） | 72 |

## サポートされるMarkdown要素

### 見出し

```markdown
# 見出し1
## 見出し2
### 見出し3
#### 見出し4
##### 見出し5
###### 見出し6
```

### 段落

```markdown
これは段落です。
空行で段落が区切られます。

これは別の段落です。
```

### 箇条書きリスト

```markdown
- 項目1
- 項目2
  - ネストされた項目
- 項目3
```

### 番号付きリスト

```markdown
1. 最初の項目
2. 2番目の項目
3. 3番目の項目
```

### コードブロック

````markdown
```go
func main() {
    fmt.Println("Hello, World!")
}
```
````

### 引用

```markdown
> これは引用文です。
> 複数行にまたがることができます。
```

### テキストスタイル

```markdown
**太字テキスト**
*斜体テキスト*
`インラインコード`
```

### 水平線（スライド区切り）

```markdown
---
```

## 完全な例

```go
package main

import (
    "os"
    "log"
    "github.com/ryomak/gopdf"
)

func main() {
    markdown := `# gopdf ドキュメント

## はじめに

gopdfは、Pure GoでPDF生成・解析を行うライブラリです。

### 特徴

- **Pure Go**: CGO不要
- **シンプルなAPI**: 直感的で使いやすい
- **日本語対応**: TTFフォントをサポート

## インストール

` + "```bash\n" + `go get github.com/ryomak/gopdf
` + "```\n" + `

## 使用例

` + "```go\n" + `doc := gopdf.New()
page := doc.AddPage(gopdf.PageSizeA4, gopdf.Portrait)
page.SetFont(gopdf.FontHelvetica, 24)
page.DrawText("Hello!", 100, 750)
` + "```\n" + `

> この例では、A4サイズのPDFを作成しています。

## まとめ

gopdfを使えば、Goで簡単にPDFを作成できます。
`

    // カスタムスタイル
    style := gopdf.DefaultMarkdownStyle()
    style.H1Size = 36
    style.H2Size = 28
    style.HeadingColor = gopdf.Color{R: 0.0, G: 0.3, B: 0.6}

    opts := &gopdf.MarkdownOptions{
        Mode:        gopdf.MarkdownModeDocument,
        PageSize:    gopdf.PageSizeA4,
        Orientation: gopdf.Portrait,
        Style:       style,
    }

    doc, err := gopdf.NewMarkdownDocument(markdown, opts)
    if err != nil {
        log.Fatal(err)
    }

    file, _ := os.Create("documentation.pdf")
    defer file.Close()
    doc.WriteTo(file)
}
```

## スライドの作成

```go
markdown := `# プレゼンテーションタイトル

発表者: 山田太郎

---

## 目次

1. はじめに
2. 本題
3. まとめ

---

## はじめに

- 背景
- 目的
- 範囲

---

## まとめ

ご清聴ありがとうございました。
`

opts := &gopdf.MarkdownOptions{
    Mode:     gopdf.MarkdownModeSlide,
    PageSize: gopdf.PageSizePresentation16x9,
}

doc, _ := gopdf.NewMarkdownDocument(markdown, opts)
```

## サンプルコード

完全な例は[examples/18_markdown](https://github.com/ryomak/gopdf/tree/main/examples/18_markdown)を参照してください。

## 次のステップ

- [暗号化とパスワード保護](encryption.md) - PDFのセキュリティ
- [メタデータ](metadata.md) - PDFメタデータの設定
