# メタデータ機能設計書

## 目的

gopdfライブラリにPDFメタデータ（Document Information Dictionary）を設定する機能を追加する。

## 要件

### 機能要件

1. 標準メタデータフィールドの設定
   - Title（タイトル）
   - Author（著者）
   - Subject（主題）
   - Keywords（キーワード）
   - Creator（作成アプリケーション）
   - Producer（PDF生成ソフトウェア）
   - CreationDate（作成日時）
   - ModDate（更新日時）

2. カスタムメタデータフィールドの設定
   - 任意のキー・バリューペアを追加可能

3. 日付の自動設定
   - CreationDateが未設定の場合、自動的に現在時刻を設定
   - Producerが未設定の場合、"gopdf"を自動設定

### 非機能要件

1. Pure Go実装
2. テスタブルなコード設計
3. Genericsの活用
4. 既存コードとの整合性

## API設計

### Metadata構造体

```go
// Metadata represents PDF document metadata (Info dictionary)
type Metadata struct {
    // Standard fields
    Title        string
    Author       string
    Subject      string
    Keywords     string
    Creator      string
    Producer     string
    CreationDate time.Time
    ModDate      time.Time

    // Custom fields (key-value pairs)
    Custom map[string]string
}
```

### Documentへのメソッド追加

```go
// SetMetadata sets the document metadata
func (d *Document) SetMetadata(metadata Metadata) {
    d.metadata = &metadata
}

// GetMetadata returns the document metadata
// Returns nil if no metadata is set
func (d *Document) GetMetadata() *Metadata {
    return d.metadata
}
```

### 使用例

```go
doc := gopdf.New()

// メタデータの設定
metadata := gopdf.Metadata{
    Title:    "Sample Document",
    Author:   "John Doe",
    Subject:  "PDF Metadata Example",
    Keywords: "PDF, metadata, gopdf",
    Creator:  "My Application",
}
doc.SetMetadata(metadata)

// カスタムフィールド付き
metadata := gopdf.Metadata{
    Title:  "Sample Document",
    Author: "John Doe",
    Custom: map[string]string{
        "Department": "Engineering",
        "Project":    "gopdf",
    },
}
doc.SetMetadata(metadata)
```

## 内部実装設計

### 1. Document構造体の拡張

```go
type Document struct {
    pages      []*Page
    encryption *EncryptionOptions
    metadata   *Metadata  // 追加
}
```

### 2. WriteTo内でのInfo辞書の生成

`WriteTo`メソッド内で以下の処理を追加：

1. メタデータが設定されている場合、Info辞書オブジェクトを生成
2. デフォルト値の適用（Producer, CreationDate）
3. Trailerに`/Info`参照を追加

実装箇所：
- Catalogオブジェクト作成後、Trailer作成前

### 3. ヘルパー関数

```go
// formatPDFDate formats a time.Time to PDF date string
// Format: D:YYYYMMDDHHmmSSOHH'mm'
func formatPDFDate(t time.Time) string

// escapeString escapes special characters in PDF strings
// Escapes: (, ), \
func escapeString(s string) string

// encodeTextString encodes a string for PDF text string
// Uses UTF-16BE with BOM for non-ASCII characters
func encodeTextString(s string) core.String
```

### 4. オブジェクト番号の調整

Info辞書の追加により、オブジェクト番号の計算ロジックを調整：

```
現在: Catalog + Pages + (Content + Page) * ページ数 + フォント + 画像
↓
変更後: Info + Catalog + Pages + (Content + Page) * ページ数 + フォント + 画像
```

## データフロー

```
1. ユーザーコード
   ↓ SetMetadata()
2. Document.metadata に保存
   ↓
3. WriteTo() 実行
   ↓
4. メタデータのデフォルト値適用
   ↓
5. Info辞書オブジェクト生成
   ↓ pdfWriter.AddObject()
6. Trailerに /Info 参照追加
   ↓
7. PDFファイル出力
```

## テスト設計

### テストケース

1. **標準フィールドのテスト**
   - 各標準フィールドが正しく設定されること
   - 日付が正しいPDF形式でエンコードされること

2. **カスタムフィールドのテスト**
   - カスタムフィールドが正しく追加されること
   - 複数のカスタムフィールドが設定できること

3. **デフォルト値のテスト**
   - Producerが未設定の場合、"gopdf"が設定されること
   - CreationDateが未設定の場合、現在時刻が設定されること

4. **特殊文字のテスト**
   - `(`, `)`, `\`が正しくエスケープされること
   - 非ASCII文字（日本語など）が正しくエンコードされること

5. **メタデータ未設定のテスト**
   - メタデータを設定しない場合もPDFが正常に生成されること
   - Trailerに/Info参照が含まれないこと

6. **統合テスト**
   - メタデータ + 暗号化
   - メタデータ + ページ + テキスト描画

### テーブルドリブンテスト例

```go
func TestMetadata_Standard(t *testing.T) {
    tests := []struct {
        name     string
        metadata Metadata
        want     map[string]string
    }{
        {
            name: "all fields",
            metadata: Metadata{
                Title:   "Test Title",
                Author:  "Test Author",
                Subject: "Test Subject",
            },
            want: map[string]string{
                "Title":   "Test Title",
                "Author":  "Test Author",
                "Subject": "Test Subject",
            },
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // テスト実装
        })
    }
}
```

## マイルストーン

1. ヘルパー関数の実装（formatPDFDate, escapeString, encodeTextString）
2. Metadata構造体とメソッドの実装
3. WriteTo内でのInfo辞書生成ロジックの実装
4. テストの実装
5. ドキュメントの更新

## 制約・注意事項

1. PDF 1.7準拠（XMPメタデータは将来的な拡張として検討）
2. 暗号化との併用を考慮（Info辞書も暗号化対象）
3. 日付のタイムゾーン処理（time.Timeの情報を保持）
4. 既存のテストに影響を与えないこと

## 参考資料

- [metadata-specification.md](./metadata-specification.md)
- PDF Reference 1.7, Section 10.2.1 (Document Information Dictionary)
