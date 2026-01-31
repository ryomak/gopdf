# 暗号化とパスワード保護

PDFの暗号化、パスワード保護、アクセス権限の設定方法を説明します。

## 概要

gopdfは以下のセキュリティ機能をサポートしています：

- **ユーザーパスワード**: PDFを開くために必要なパスワード
- **オーナーパスワード**: 権限変更に必要なパスワード
- **アクセス権限**: 印刷、コピー、編集などの制限
- **暗号化強度**: 40ビット/128ビット RC4

## PDF暗号化

### 基本的な暗号化

```go
doc := gopdf.New()
page := doc.AddPage(gopdf.PageSizeA4, gopdf.Portrait)

page.SetFont(gopdf.FontHelvetica, 24)
page.DrawText("Confidential Document", 100, 750)

// 暗号化を設定
doc.SetEncryption(gopdf.EncryptionOptions{
    UserPassword:  "user123",
    OwnerPassword: "owner456",
})

file, _ := os.Create("encrypted.pdf")
defer file.Close()
doc.WriteTo(file)
```

### EncryptionOptions

```go
type EncryptionOptions struct {
    UserPassword  string      // PDFを開くためのパスワード
    OwnerPassword string      // 権限変更用パスワード
    Permissions   Permissions // アクセス権限
    KeyLength     int         // 暗号化キー長（40 or 128）
}
```

## パスワードの種類

### ユーザーパスワード

PDFを開くときに要求されるパスワードです。

```go
doc.SetEncryption(gopdf.EncryptionOptions{
    UserPassword: "password123",
})
```

### オーナーパスワード

権限を変更したり、暗号化を解除するためのパスワードです。
オーナーパスワードだけを設定すると、PDFは開けますが権限が制限されます。

```go
doc.SetEncryption(gopdf.EncryptionOptions{
    OwnerPassword: "owner456",
    Permissions: gopdf.RestrictedPermissions(),
})
```

### 両方のパスワード

```go
doc.SetEncryption(gopdf.EncryptionOptions{
    UserPassword:  "user123",   // 開くためのパスワード
    OwnerPassword: "owner456",  // 権限変更用パスワード
})
```

## アクセス権限

### Permissions構造体

```go
type Permissions struct {
    Print          bool // 印刷許可
    PrintHighQuality bool // 高品質印刷許可
    Copy           bool // テキスト・画像のコピー許可
    Modify         bool // ドキュメントの編集許可
    Annotate       bool // 注釈の追加許可
    FillForms      bool // フォーム入力許可
    ExtractForAccessibility bool // アクセシビリティ用抽出許可
    Assemble       bool // ページ追加・削除許可
}
```

### プリセット権限

```go
// デフォルト（すべて許可）
gopdf.DefaultPermissions()

// 制限付き（すべて禁止）
gopdf.RestrictedPermissions()

// 印刷のみ許可
gopdf.PrintOnlyPermissions()
```

### カスタム権限

```go
doc.SetEncryption(gopdf.EncryptionOptions{
    UserPassword:  "user123",
    OwnerPassword: "owner456",
    Permissions: gopdf.Permissions{
        Print:     true,  // 印刷許可
        Copy:      false, // コピー禁止
        Modify:    false, // 編集禁止
        Annotate:  true,  // 注釈許可
        FillForms: true,  // フォーム入力許可
    },
})
```

## 暗号化強度

```go
// 40ビット RC4（互換性重視）
doc.SetEncryption(gopdf.EncryptionOptions{
    UserPassword: "password",
    KeyLength:    40,
})

// 128ビット RC4（推奨）
doc.SetEncryption(gopdf.EncryptionOptions{
    UserPassword: "password",
    KeyLength:    128,
})
```

| キー長 | 説明 |
|-------|------|
| 40ビット | 古いPDFリーダーとの互換性がある。セキュリティは低い |
| 128ビット | 推奨。より強力な暗号化 |

## 暗号化PDFの読み込み

### 暗号化チェック

```go
file, _ := os.Open("encrypted.pdf")
defer file.Close()

reader, err := gopdf.OpenReader(file)
if err != nil {
    log.Fatal(err)
}

if reader.IsEncrypted() {
    fmt.Println("This PDF is encrypted")
}
```

### パスワード認証

```go
if reader.IsEncrypted() {
    err := reader.AuthenticateWithPassword("user123")
    if err != nil {
        log.Fatal("Invalid password")
    }
}

// 認証後、通常通り操作可能
text, _ := reader.ExtractText()
```

### 暗号化情報の取得

```go
if reader.IsEncrypted() {
    reader.AuthenticateWithPassword("password")

    encInfo := reader.GetEncryptionInfo()
    fmt.Printf("Key Length: %d bits\n", encInfo.KeyLength)
    fmt.Printf("Permissions:\n")
    fmt.Printf("  Print: %v\n", encInfo.Permissions.Print)
    fmt.Printf("  Copy: %v\n", encInfo.Permissions.Copy)
    fmt.Printf("  Modify: %v\n", encInfo.Permissions.Modify)
}
```

## 完全な例

### 暗号化PDFの作成

```go
package main

import (
    "os"
    "github.com/ryomak/gopdf"
)

func main() {
    doc := gopdf.New()
    page := doc.AddPage(gopdf.PageSizeA4, gopdf.Portrait)

    page.SetFont(gopdf.FontHelveticaBold, 24)
    page.DrawText("Confidential Report", 100, 750)

    page.SetFont(gopdf.FontHelvetica, 14)
    page.DrawText("This document contains sensitive information.", 100, 700)
    page.DrawText("Unauthorized access is prohibited.", 100, 680)

    // 暗号化設定
    doc.SetEncryption(gopdf.EncryptionOptions{
        UserPassword:  "viewonly123",
        OwnerPassword: "admin456",
        Permissions: gopdf.Permissions{
            Print:     true,
            Copy:      false,
            Modify:    false,
            Annotate:  false,
            FillForms: false,
        },
        KeyLength: 128,
    })

    file, _ := os.Create("confidential.pdf")
    defer file.Close()
    doc.WriteTo(file)
}
```

### 暗号化PDFの読み込みと復号

```go
package main

import (
    "fmt"
    "log"
    "os"
    "github.com/ryomak/gopdf"
)

func main() {
    file, _ := os.Open("confidential.pdf")
    defer file.Close()

    reader, err := gopdf.OpenReader(file)
    if err != nil {
        log.Fatal(err)
    }

    // 暗号化チェック
    if !reader.IsEncrypted() {
        fmt.Println("PDF is not encrypted")
        return
    }

    fmt.Println("PDF is encrypted")

    // パスワード認証
    password := "viewonly123"
    if err := reader.AuthenticateWithPassword(password); err != nil {
        log.Fatal("Authentication failed:", err)
    }

    fmt.Println("Authentication successful!")

    // 暗号化情報
    encInfo := reader.GetEncryptionInfo()
    fmt.Printf("Encryption: %d-bit\n", encInfo.KeyLength)
    fmt.Printf("Can print: %v\n", encInfo.Permissions.Print)
    fmt.Printf("Can copy: %v\n", encInfo.Permissions.Copy)

    // テキスト抽出
    text, _ := reader.ExtractText()
    fmt.Println("\nExtracted text:")
    fmt.Println(text)
}
```

## セキュリティ上の注意

- パスワードはソースコードにハードコードせず、環境変数や設定ファイルから読み込むことを推奨します
- 重要なドキュメントには128ビット暗号化を使用してください
- ユーザーパスワードとオーナーパスワードは異なるものを設定してください

## サンプルコード

- [examples/13_encryption](https://github.com/ryomak/gopdf/tree/main/examples/13_encryption) - PDF暗号化
- [examples/14_decrypt_pdf](https://github.com/ryomak/gopdf/tree/main/examples/14_decrypt_pdf) - PDF復号

## 次のステップ

- [メタデータ](metadata.md) - PDFメタデータの設定
- [PDF読み込み・解析](reading-pdf.md) - PDFの読み込み
