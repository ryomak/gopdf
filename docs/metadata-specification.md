# PDFメタデータ仕様

## 概要

PDFドキュメントのメタデータは、ドキュメント情報辞書（Document Information Dictionary）を使用して管理されます。この辞書はPDFファイルのtrailerセクションから`/Info`キーで参照されます。

## 準拠規格

- **PDF 1.7** (ISO 32000-1:2008)
- **注意**: PDF 2.0 (ISO 32000-2:2020) では、Info Dictionaryは非推奨となり、XMP（eXtensible Metadata Platform）メタデータが推奨されています。

## Document Information Dictionary（Info Dictionary）

### 標準エントリ

PDF 1.7仕様で定義されている標準エントリは以下の通りです：

| キー | 型 | 説明 | 例 |
|------|-----|------|-----|
| `Title` | text string | ドキュメントのタイトル | "Sample PDF Document" |
| `Author` | text string | ドキュメントの作成者（個人または組織） | "John Doe" |
| `Subject` | text string | ドキュメントの主題 | "PDF Metadata Example" |
| `Keywords` | text string | キーワード（カンマ区切り） | "PDF, metadata, gopdf" |
| `Creator` | text string | 元のドキュメントを作成したアプリケーション | "Microsoft Word" |
| `Producer` | text string | PDFに変換したソフトウェア | "gopdf v1.0.0" |
| `CreationDate` | date | ドキュメントの作成日時 | "D:20250129123045+09'00'" |
| `ModDate` | date | ドキュメントの最終更新日時 | "D:20250129123045+09'00'" |
| `Trapped` | name | トラッピング情報（印刷用） | `/True`, `/False`, `/Unknown` |

### 日付形式

PDFの日付形式は以下の通りです：

```
D:YYYYMMDDHHmmSSOHH'mm'
```

- `D:`: 日付を示すプレフィックス（必須）
- `YYYY`: 年（4桁）
- `MM`: 月（01-12）
- `DD`: 日（01-31）
- `HH`: 時（00-23）
- `mm`: 分（00-59）
- `SS`: 秒（00-59）
- `O`: UTCからのオフセット（`+`, `-`, `Z`）
- `HH'mm'`: オフセットの時間と分

例：
- `D:20250129123045+09'00'` - 2025年1月29日 12:30:45 JST
- `D:20250129` - 2025年1月29日（時刻なし）

### カスタムエントリ

PDF仕様は、上記の標準エントリ以外のカスタムエントリも許可しています。カスタムエントリを追加する場合は、以下の点に注意してください：

- キー名は名前オブジェクト（Name object）として定義
- 値は文字列（text string）、日付（date）、名前（name）のいずれか
- PDF/A準拠が必要な場合は、XMPメタデータとの整合性を保つ必要がある

## バイナリ構造

### Info Dictionary オブジェクト

```
5 0 obj
<<
  /Title (Sample PDF Document)
  /Author (John Doe)
  /Subject (PDF Metadata Example)
  /Keywords (PDF, metadata, gopdf)
  /Creator (gopdf)
  /Producer (gopdf v1.0.0)
  /CreationDate (D:20250129123045+09'00')
  /ModDate (D:20250129123045+09'00')
>>
endobj
```

### Trailer での参照

```
trailer
<<
  /Size 10
  /Root 1 0 R
  /Info 5 0 R
>>
```

## 実装上の注意点

1. **エンコーディング**:
   - テキスト文字列はPDFDocEncodingまたはUTF-16BEでエンコード
   - 非ASCII文字を含む場合は、BOM（`\xFE\xFF`）付きUTF-16BEを使用

2. **特殊文字のエスケープ**:
   - `(`, `)`, `\` は `\` でエスケープが必要
   - 例: `(Title with \(parentheses\))`

3. **省略可能**:
   - すべてのエントリは省略可能
   - 最小限の実装では、`Producer`と`CreationDate`を設定することが推奨される

4. **更新**:
   - PDFを更新する際は、`ModDate`を現在の日時に更新すべき
   - `CreationDate`は元のドキュメント作成日時を保持

## 参考資料

- [PDF Reference, version 1.7](https://www.adobe.com/devnet/pdf/pdf_reference.html)
- [ISO 32000-1:2008](https://www.iso.org/standard/51502.html)
- [PDF metadata - Prepressure](https://www.prepressure.com/pdf/basics/metadata)
