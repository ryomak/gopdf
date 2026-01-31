# メタデータ

PDFのメタデータ（タイトル、作成者、キーワードなど）の設定と取得方法を説明します。

## 概要

PDFメタデータには以下の情報を含めることができます：

- **Title**: ドキュメントのタイトル
- **Author**: 作成者
- **Subject**: 件名・主題
- **Keywords**: キーワード（検索用）
- **Creator**: 作成アプリケーション
- **Producer**: PDF生成ツール
- **CreationDate**: 作成日時
- **ModDate**: 更新日時
- **Custom**: カスタムフィールド

## メタデータの設定

### Metadata構造体

```go
type Metadata struct {
    Title        string            // タイトル
    Author       string            // 作成者
    Subject      string            // 件名
    Keywords     string            // キーワード（カンマ区切り）
    Creator      string            // 作成アプリケーション
    Producer     string            // PDF生成ツール
    CreationDate time.Time         // 作成日時
    ModDate      time.Time         // 更新日時
    Custom       map[string]string // カスタムフィールド
}
```

### 基本的な設定

```go
doc := gopdf.New()
page := doc.AddPage(gopdf.PageSizeA4, gopdf.Portrait)

// コンテンツを追加
page.SetFont(gopdf.FontHelvetica, 24)
page.DrawText("Document with Metadata", 100, 750)

// メタデータを設定
doc.SetMetadata(gopdf.Metadata{
    Title:    "Annual Report 2024",
    Author:   "John Smith",
    Subject:  "Company Performance Summary",
    Keywords: "report, annual, 2024, performance",
})

file, _ := os.Create("report.pdf")
defer file.Close()
doc.WriteTo(file)
```

### 日時の設定

```go
now := time.Now()

doc.SetMetadata(gopdf.Metadata{
    Title:        "Document Title",
    Author:       "Author Name",
    CreationDate: now,
    ModDate:      now,
})
```

### カスタムフィールド

アプリケーション固有の情報を埋め込むことができます：

```go
doc.SetMetadata(gopdf.Metadata{
    Title:  "Project Document",
    Author: "Team Lead",
    Custom: map[string]string{
        "Department":  "Engineering",
        "ProjectID":   "PRJ-2024-001",
        "Version":     "1.0.0",
        "Confidential": "Yes",
    },
})
```

## メタデータの取得

### PDFからメタデータを読み込み

```go
reader, err := gopdf.Open("document.pdf")
if err != nil {
    log.Fatal(err)
}

info := reader.Info()

fmt.Printf("Title: %s\n", info.Title)
fmt.Printf("Author: %s\n", info.Author)
fmt.Printf("Subject: %s\n", info.Subject)
fmt.Printf("Keywords: %s\n", info.Keywords)
fmt.Printf("Creator: %s\n", info.Creator)
fmt.Printf("Producer: %s\n", info.Producer)
```

### 日時の取得

```go
info := reader.Info()

if !info.CreationDate.IsZero() {
    fmt.Printf("Created: %s\n", info.CreationDate.Format("2006-01-02 15:04:05"))
}

if !info.ModDate.IsZero() {
    fmt.Printf("Modified: %s\n", info.ModDate.Format("2006-01-02 15:04:05"))
}
```

### カスタムフィールドの取得

```go
info := reader.Info()

if info.Custom != nil {
    for key, value := range info.Custom {
        fmt.Printf("%s: %s\n", key, value)
    }
}
```

## 完全な例

### メタデータ付きPDFの作成

```go
package main

import (
    "os"
    "time"
    "github.com/ryomak/gopdf"
)

func main() {
    doc := gopdf.New()

    // メタデータを設定
    doc.SetMetadata(gopdf.Metadata{
        Title:        "gopdf User Guide",
        Author:       "Development Team",
        Subject:      "PDF Generation Library Documentation",
        Keywords:     "gopdf, pdf, go, golang, library",
        Creator:      "gopdf documentation generator",
        CreationDate: time.Now(),
        ModDate:      time.Now(),
        Custom: map[string]string{
            "Version":      "1.0.0",
            "Language":     "Japanese",
            "DocumentType": "Technical Documentation",
        },
    })

    // ページを追加
    page := doc.AddPage(gopdf.PageSizeA4, gopdf.Portrait)

    page.SetFont(gopdf.FontHelveticaBold, 28)
    page.DrawText("gopdf User Guide", 100, 750)

    page.SetFont(gopdf.FontHelvetica, 14)
    page.DrawText("Version 1.0.0", 100, 720)
    page.DrawText("Development Team", 100, 700)

    // 保存
    file, _ := os.Create("user_guide.pdf")
    defer file.Close()
    doc.WriteTo(file)
}
```

### メタデータの読み込みと表示

```go
package main

import (
    "fmt"
    "log"
    "github.com/ryomak/gopdf"
)

func main() {
    reader, err := gopdf.Open("user_guide.pdf")
    if err != nil {
        log.Fatal(err)
    }

    info := reader.Info()

    fmt.Println("=== PDF Metadata ===")
    fmt.Println()

    // 標準フィールド
    if info.Title != "" {
        fmt.Printf("Title:    %s\n", info.Title)
    }
    if info.Author != "" {
        fmt.Printf("Author:   %s\n", info.Author)
    }
    if info.Subject != "" {
        fmt.Printf("Subject:  %s\n", info.Subject)
    }
    if info.Keywords != "" {
        fmt.Printf("Keywords: %s\n", info.Keywords)
    }
    if info.Creator != "" {
        fmt.Printf("Creator:  %s\n", info.Creator)
    }
    if info.Producer != "" {
        fmt.Printf("Producer: %s\n", info.Producer)
    }

    // 日時
    if !info.CreationDate.IsZero() {
        fmt.Printf("Created:  %s\n", info.CreationDate.Format("2006-01-02 15:04:05"))
    }
    if !info.ModDate.IsZero() {
        fmt.Printf("Modified: %s\n", info.ModDate.Format("2006-01-02 15:04:05"))
    }

    // カスタムフィールド
    if len(info.Custom) > 0 {
        fmt.Println()
        fmt.Println("=== Custom Fields ===")
        for key, value := range info.Custom {
            fmt.Printf("%s: %s\n", key, value)
        }
    }

    // ページ情報
    fmt.Println()
    fmt.Printf("Pages: %d\n", reader.PageCount())
}
```

## ユースケース

### ドキュメント管理システム

```go
// ドキュメントIDとバージョン管理
doc.SetMetadata(gopdf.Metadata{
    Title:  "Contract Agreement",
    Author: "Legal Department",
    Custom: map[string]string{
        "DocumentID":   "DOC-2024-12345",
        "Version":      "2.1",
        "Status":       "Approved",
        "ApprovedBy":   "Manager Name",
        "ApprovedDate": "2024-01-15",
    },
})
```

### SEO/検索最適化

```go
// 検索しやすいキーワードを設定
doc.SetMetadata(gopdf.Metadata{
    Title:    "Go言語によるPDF生成入門",
    Subject:  "プログラミング、PDF生成、Goライブラリ",
    Keywords: "Go, Golang, PDF, 生成, ライブラリ, プログラミング, チュートリアル",
})
```

### アーカイブ用

```go
// 長期保存用のメタデータ
doc.SetMetadata(gopdf.Metadata{
    Title:        "Meeting Minutes - Q4 2024",
    Author:       "Secretary",
    CreationDate: time.Now(),
    Custom: map[string]string{
        "MeetingDate":  "2024-12-15",
        "Location":     "Conference Room A",
        "Attendees":    "5",
        "RecordKeeper": "Archive System",
    },
})
```

## サンプルコード

完全な例は[examples/15_metadata](https://github.com/ryomak/gopdf/tree/main/examples/15_metadata)を参照してください。

## 次のステップ

- [高度な機能](advanced.md) - ルビ、OCR、レイアウト調整
- [PDF読み込み・解析](reading-pdf.md) - PDFの読み込み
