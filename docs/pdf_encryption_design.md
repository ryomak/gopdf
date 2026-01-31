# PDF Encryption & Password Protection Design

## 概要

PDFにパスワード保護と権限制御を追加する機能の設計書。
RC4暗号化をサポートし、ユーザーパスワードとオーナーパスワードによるアクセス制御を実装する。

## 目的

- PDFファイルをパスワードで保護
- 印刷・コピー・編集などの権限を制御
- 既存のPDFリーダーとの互換性を確保

## PDF暗号化の基本概念

### 2種類のパスワード

1. **ユーザーパスワード（User Password）**
   - PDFを開くために必要
   - 設定された権限内での操作のみ可能

2. **オーナーパスワード（Owner Password）**
   - すべての操作が可能（権限制限なし）
   - パスワード変更や権限変更が可能

### 暗号化方式

**Phase 12では RC4 のみ実装**（最も広くサポートされている）:
- RC4 40-bit（PDF 1.2互換）
- RC4 128-bit（PDF 1.4互換）

**将来の拡張**:
- AES 128-bit（PDF 1.6）
- AES 256-bit（PDF 1.7 Extension Level 3）

## データ構造

### EncryptionOptions

```go
// EncryptionOptions はPDF暗号化のオプション
type EncryptionOptions struct {
    UserPassword  string       // ユーザーパスワード（PDFを開くために必要）
    OwnerPassword string       // オーナーパスワード（すべての権限）
    Permissions   Permissions  // アクセス権限
    KeyLength     int          // 暗号鍵の長さ（40 or 128 bits）
}
```

### Permissions

```go
// Permissions はPDFのアクセス権限
type Permissions struct {
    Print          bool // 印刷を許可
    Modify         bool // 内容の変更を許可
    Copy           bool // テキスト・グラフィックのコピーを許可
    Annotate       bool // 注釈の追加・変更を許可
    FillForms      bool // フォームフィールドの入力を許可
    ExtractContent bool // テキスト・グラフィックの抽出を許可
    Assemble       bool // ページの挿入・削除・回転を許可
    PrintHighQuality bool // 高解像度での印刷を許可
}
```

### 権限フラグ（PDF仕様）

```go
const (
    PermPrint          = 1 << 2  // bit 3: 印刷
    PermModify         = 1 << 3  // bit 4: 内容変更
    PermCopy           = 1 << 4  // bit 5: コピー
    PermAnnotate       = 1 << 5  // bit 6: 注釈
    PermFillForms      = 1 << 8  // bit 9: フォーム入力
    PermExtract        = 1 << 9  // bit 10: 抽出
    PermAssemble       = 1 << 10 // bit 11: ページ操作
    PermPrintHighQual  = 1 << 11 // bit 12: 高解像度印刷
)
```

## API設計

### Document メソッド

```go
// SetEncryption はPDFに暗号化を設定
func (d *Document) SetEncryption(opts EncryptionOptions) error

// RemoveEncryption は暗号化を解除（オーナーパスワード必要）
func (d *Document) RemoveEncryption(ownerPassword string) error
```

### 便利な権限プリセット

```go
// DefaultPermissions はデフォルトの権限（すべて許可）
func DefaultPermissions() Permissions

// RestrictedPermissions は制限された権限（閲覧のみ）
func RestrictedPermissions() Permissions

// PrintOnlyPermissions は印刷のみ許可
func PrintOnlyPermissions() Permissions
```

## PDF暗号化の仕組み

### 1. Encrypt辞書

```
/Encrypt <<
    /Filter /Standard
    /V 1                    % Version (1 = 40-bit RC4, 2 = 128-bit RC4)
    /R 2                    % Revision (2 = 40-bit, 3 = 128-bit)
    /O <hex>                % Owner password hash (32 bytes)
    /U <hex>                % User password hash (32 bytes)
    /P -44                  % Permissions (32-bit integer)
    /Length 40              % Key length in bits (40 or 128)
>>
```

### 2. 暗号化キー生成

**40-bit RC4（R=2）の場合**:

1. パディング定数を用意
```
padding = <28 BF 4E 5E 4E 75 8A 41 64 00 4E 56 FF FA 01 08
           2E 2E 00 B6 D0 68 3E 80 2F 0C A9 FE 64 53 69 7A>
```

2. MD5ハッシュ計算
```
hash = MD5(password + padding + O値 + P値 + FileID)
key = hash[0:5]  // 最初の5バイト（40 bits）
```

3. ストリーム暗号化
```
encrypted = RC4(key, plaintext)
```

### 3. O値（Owner Password）の計算

```
1. ownerPadded = (ownerPassword + padding)[0:32]
2. hash = MD5(ownerPadded)
3. if R >= 3:
      for i = 0 to 49:
          hash = MD5(hash)
4. key = hash[0:keyLength/8]
5. userPadded = (userPassword + padding)[0:32]
6. O = RC4(key, userPadded)
7. if R >= 3:
      for i = 1 to 19:
          key_i = XOR(key, i)
          O = RC4(key_i, O)
```

### 4. U値（User Password）の計算

```
1. if R == 2:
      U = RC4(encryptionKey, padding)
   else if R >= 3:
      hash = MD5(padding + FileID)
      U = RC4(encryptionKey, hash)
      for i = 1 to 19:
          key_i = XOR(encryptionKey, i)
          U = RC4(key_i, U)
```

## 実装計画

### Phase 1: RC4暗号化基盤

```go
// internal/security/rc4.go
type RC4Cipher struct {
    S [256]byte
    i, j byte
}

func NewRC4(key []byte) *RC4Cipher
func (r *RC4Cipher) Encrypt(data []byte) []byte
func (r *RC4Cipher) Decrypt(data []byte) []byte
```

### Phase 2: パスワード処理

```go
// internal/security/password.go

// ComputeEncryptionKey は暗号化キーを計算
func ComputeEncryptionKey(
    password string,
    o []byte,
    permissions int32,
    fileID []byte,
    revision int,
    keyLength int,
) []byte

// ComputeOwnerPassword はO値を計算
func ComputeOwnerPassword(
    ownerPassword, userPassword string,
    revision int,
    keyLength int,
) []byte

// ComputeUserPassword はU値を計算
func ComputeUserPassword(
    encryptionKey, fileID []byte,
    revision int,
) []byte

// AuthenticateUserPassword はユーザーパスワードを検証
func AuthenticateUserPassword(
    password string,
    u, o []byte,
    permissions int32,
    fileID []byte,
    revision int,
    keyLength int,
) bool

// AuthenticateOwnerPassword はオーナーパスワードを検証
func AuthenticateOwnerPassword(
    password string,
    o []byte,
    revision int,
    keyLength int,
) (string, bool) // userPassword, ok
```

### Phase 3: 権限管理

```go
// internal/security/permissions.go

// PermissionsToInt32 は権限構造体を32-bit整数に変換
func PermissionsToInt32(p Permissions) int32

// Int32ToPermissions は32-bit整数を権限構造体に変換
func Int32ToPermissions(flags int32) Permissions
```

### Phase 4: Document統合

```go
// document.go

type Document struct {
    // ...既存フィールド
    encryption *EncryptionOptions
}

func (d *Document) SetEncryption(opts EncryptionOptions) error {
    // バリデーション
    if opts.UserPassword == "" && opts.OwnerPassword == "" {
        return fmt.Errorf("at least one password must be set")
    }
    if opts.KeyLength != 40 && opts.KeyLength != 128 {
        return fmt.Errorf("key length must be 40 or 128")
    }

    d.encryption = &opts
    return nil
}
```

### Phase 5: Writer統合

```go
// internal/writer/writer.go

func (w *Writer) writeEncryptDict(fileID []byte, opts EncryptionOptions) error {
    // Encrypt辞書を生成
    // O値、U値、P値を計算
    // ストリームを暗号化
}
```

## 使用例

### 例1: 基本的なパスワード保護

```go
doc := gopdf.New()
page := doc.AddPage(gopdf.A4, gopdf.Portrait)
page.SetFont(font.Helvetica, 12)
page.DrawText("Confidential Document", 50, 800)

// パスワード保護を設定
err := doc.SetEncryption(gopdf.EncryptionOptions{
    UserPassword: "user123",     // PDFを開くパスワード
    OwnerPassword: "owner456",   // 管理者パスワード
    KeyLength: 128,              // 128-bit暗号化
    Permissions: gopdf.DefaultPermissions(), // すべて許可
})

doc.SaveToFile("protected.pdf")
```

### 例2: 閲覧のみ許可

```go
doc := gopdf.New()
// ... ページ追加

// 閲覧のみ許可（印刷・コピー・編集禁止）
err := doc.SetEncryption(gopdf.EncryptionOptions{
    UserPassword: "view123",
    OwnerPassword: "admin456",
    KeyLength: 128,
    Permissions: gopdf.RestrictedPermissions(), // 閲覧のみ
})

doc.SaveToFile("readonly.pdf")
```

### 例3: カスタム権限

```go
doc := gopdf.New()
// ... ページ追加

// カスタム権限: 印刷とコピーのみ許可
err := doc.SetEncryption(gopdf.EncryptionOptions{
    UserPassword: "user",
    OwnerPassword: "owner",
    KeyLength: 128,
    Permissions: gopdf.Permissions{
        Print:    true,  // 印刷OK
        Copy:     true,  // コピーOK
        Modify:   false, // 編集NG
        Annotate: false, // 注釈NG
    },
})

doc.SaveToFile("print-copy-only.pdf")
```

### 例4: パスワードなし（権限のみ）

```go
// オーナーパスワードのみ設定
// 誰でも開けるが、権限は制限される
err := doc.SetEncryption(gopdf.EncryptionOptions{
    OwnerPassword: "admin123",
    KeyLength: 128,
    Permissions: gopdf.Permissions{
        Print: true,
        Copy:  false,
    },
})
```

## 制限事項

### Phase 12での制限
- RC4暗号化のみサポート（AESは将来実装）
- PDF生成時の暗号化のみ（既存PDFの復号化は Phase 13以降）
- 40-bit と 128-bit のみ（256-bitは将来実装）

### セキュリティ上の注意
- RC4は現代の基準では弱い暗号化方式
- 重要なドキュメントには将来のAES実装を推奨
- パスワードの強度は重要（8文字以上、複雑な文字列を推奨）

## テスト計画

### ユニットテスト
- RC4暗号化・復号化
- パスワードハッシュ計算
- 権限フラグ変換
- O値・U値の計算

### 統合テスト
- パスワード保護PDFの生成
- 異なる権限設定での動作確認
- PDF Readerでの開封確認
- 権限制御の動作確認

## 参考資料

- PDF Reference 1.7, Section 3.5 "Encryption"
- PDF Reference 1.7, Section 3.5.2 "Standard Security Handler"
- RC4 Algorithm: https://en.wikipedia.org/wiki/RC4
- MD5 Hash: https://en.wikipedia.org/wiki/MD5
