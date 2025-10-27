# Phase 6: PDF解析・読み込み 設計書

## 1. 概要

PDF解析・読み込み機能（Reader Layer）を実装し、既存のPDFファイルを読み込んで構造を解析できるようにする。

### 1.1. 目的

- 既存PDFファイルの読み込み
- PDFオブジェクトの解析とデシリアライズ
- ドキュメント構造（Catalog, Pages, Page）の取得
- Phase 7（テキスト抽出）の基盤となる機能を提供

### 1.2. スコープ

**Phase 6で実装する機能:**
- ✅ PDFファイルからの基本構造の読み込み
- ✅ Trailer, Xref tableの解析
- ✅ 間接オブジェクトの読み込み
- ✅ Dictionary, Array, Stream等のデシリアライズ
- ✅ Catalog, Pages, Pageオブジェクトの取得
- ✅ ページ数の取得
- ✅ 基本的なメタデータ（Info辞書）の取得

**Phase 6では実装しない機能:**
- テキスト抽出（Phase 7で実装）
- 画像抽出（Phase 7で実装）
- 暗号化PDFの復号（将来の拡張）
- PDF 2.0のObject Streams（まずはPDF 1.7のxref tableに集中）

## 2. PDF読み込みの技術的背景

### 2.1. PDFファイルの読み込みフロー

PDF仕様（ISO 32000-1）に基づく標準的な読み込みフロー：

```
1. ファイル末尾から %%EOF を探す
   ↓
2. startxref キーワードを見つけ、xrefオフセットを取得
   ↓
3. xrefオフセット位置にシークし、xref tableを読む
   ↓
4. trailerを読み、/Root (Catalog) への参照を取得
   ↓
5. Catalogを起点に、必要なオブジェクトを順次読み込む
   ↓
6. オブジェクト参照 (N M R) を解決してデータを取得
```

### 2.2. PDF構造のおさらい

```
%PDF-1.7                          ← Header
<binary objects>                  ← Body
xref                              ← Cross Reference Table
0 6
0000000000 65535 f
0000000015 00000 n
...
trailer                           ← Trailer
<< /Size 6 /Root 1 0 R >>
startxref
531
%%EOF                             ← End of File
```

### 2.3. 主要なPDFオブジェクト型

| 型 | 表記例 | 説明 |
|----|-------|------|
| Null | `null` | 値なし |
| Boolean | `true` / `false` | 真偽値 |
| Integer | `42` | 整数 |
| Real | `3.14` | 実数 |
| String | `(Hello)` または `<48656C6C6F>` | 文字列 |
| Name | `/Type` | 名前（識別子） |
| Array | `[1 2 3]` | 配列 |
| Dictionary | `<< /Type /Page >>` | 辞書 |
| Stream | Dictionary + `stream ... endstream` | バイナリストリーム |
| Reference | `1 0 R` | 間接オブジェクトへの参照 |

## 3. 実装アプローチ

### 3.1. アーキテクチャ

```
┌──────────────────────────────────────┐
│  gopdf.Open(path)                    │  ← 公開API
│  gopdf.Reader                        │
└──────────────┬───────────────────────┘
               │
┌──────────────▼───────────────────────┐
│  internal/reader/reader.go           │  ← Reader本体
│  - Xref table管理                     │
│  - オブジェクトキャッシュ              │
│  - Catalog/Pages/Pageへのアクセス     │
└──────────────┬───────────────────────┘
               │
┌──────────────▼───────────────────────┐
│  internal/reader/parser.go           │  ← パーサー
│  - オブジェクトのパース                │
│  - Dictionary/Array/Stream解析       │
└──────────────┬───────────────────────┘
               │
┌──────────────▼───────────────────────┐
│  internal/reader/lexer.go            │  ← 字句解析
│  - トークン化                          │
│  - 空白・コメントの処理                 │
└──────────────────────────────────────┘
```

### 3.2. Reader構造体設計

```go
package reader

import (
    "io"
    "github.com/ryomak/gopdf/internal/core"
)

// Reader はPDFファイルを読み込み、解析する
type Reader struct {
    r         io.ReadSeeker      // ファイルのシーク可能なリーダー
    xref      map[int]xrefEntry  // オブジェクト番号 -> オフセット
    trailer   core.Dictionary    // Trailer辞書
    catalog   core.Dictionary    // Catalogオブジェクト
    objCache  map[int]core.Object // オブジェクトキャッシュ
}

// xrefEntry はクロスリファレンステーブルのエントリ
type xrefEntry struct {
    offset     int64  // ファイル内バイトオフセット
    generation int    // 世代番号
    inUse      bool   // 使用中かどうか
}

// NewReader は新しいReaderを作成する
func NewReader(r io.ReadSeeker) (*Reader, error)

// GetObject はオブジェクト番号からオブジェクトを取得する
func (r *Reader) GetObject(objNum int) (core.Object, error)

// GetCatalog はCatalogオブジェクトを返す
func (r *Reader) GetCatalog() (core.Dictionary, error)

// GetPageCount はページ数を返す
func (r *Reader) GetPageCount() (int, error)

// GetPage は指定されたページ番号のPageオブジェクトを返す（0-indexed）
func (r *Reader) GetPage(pageNum int) (core.Dictionary, error)

// GetInfo はInfo辞書（メタデータ）を返す
func (r *Reader) GetInfo() (core.Dictionary, error)
```

### 3.3. Parser設計

```go
package reader

import (
    "io"
    "github.com/ryomak/gopdf/internal/core"
)

// Parser はPDFオブジェクトをパースする
type Parser struct {
    lexer *Lexer
}

// NewParser は新しいParserを作成する
func NewParser(r io.Reader) *Parser

// ParseObject は次のオブジェクトをパースする
func (p *Parser) ParseObject() (core.Object, error)

// ParseDictionary は辞書をパースする
func (p *Parser) ParseDictionary() (core.Dictionary, error)

// ParseArray は配列をパースする
func (p *Parser) ParseArray() (core.Array, error)

// ParseStream はストリームをパースする
func (p *Parser) ParseStream(dict core.Dictionary) (*core.Stream, error)

// ParseIndirectObject は間接オブジェクトをパースする
func (p *Parser) ParseIndirectObject() (int, int, core.Object, error)
```

### 3.4. Lexer設計

```go
package reader

import (
    "io"
)

// TokenType はトークンの種類
type TokenType int

const (
    TokenEOF TokenType = iota
    TokenInteger      // 123
    TokenReal         // 3.14
    TokenString       // (text) or <hex>
    TokenName         // /Name
    TokenKeyword      // obj, endobj, stream, etc.
    TokenDictStart    // <<
    TokenDictEnd      // >>
    TokenArrayStart   // [
    TokenArrayEnd     // ]
    TokenRef          // R
)

// Token はトークン
type Token struct {
    Type  TokenType
    Value interface{}
}

// Lexer はPDFバイトストリームをトークン化する
type Lexer struct {
    r   io.ByteReader
    buf []byte
    pos int
}

// NewLexer は新しいLexerを作成する
func NewLexer(r io.Reader) *Lexer

// NextToken は次のトークンを返す
func (l *Lexer) NextToken() (Token, error)

// SkipWhitespace は空白文字とコメントをスキップする
func (l *Lexer) SkipWhitespace() error

// PeekByte は次のバイトを先読みする（消費しない）
func (l *Lexer) PeekByte() (byte, error)
```

## 4. 実装の詳細

### 4.1. Xref Table解析

```go
// parseXref はクロスリファレンステーブルをパースする
func (r *Reader) parseXref(offset int64) error {
    // offset位置にシーク
    r.r.Seek(offset, io.SeekStart)

    // "xref" キーワードを確認
    // サブセクションをループ処理
    //   - 開始オブジェクト番号と個数を読む
    //   - 各エントリを読む（offset generation n/f）
    //   - xrefマップに格納

    return nil
}
```

### 4.2. Trailer解析

```go
// parseTrailer はTrailerをパースする
func (r *Reader) parseTrailer() error {
    // "trailer" キーワードを確認
    // Dictionaryをパースして r.trailer に格納
    // /Root, /Size, /Info などを取得

    return nil
}
```

### 4.3. オブジェクトの読み込み

```go
// GetObject はオブジェクト番号からオブジェクトを取得する
func (r *Reader) GetObject(objNum int) (core.Object, error) {
    // キャッシュをチェック
    if obj, ok := r.objCache[objNum]; ok {
        return obj, nil
    }

    // xrefからオフセットを取得
    entry, ok := r.xref[objNum]
    if !ok {
        return nil, fmt.Errorf("object %d not found", objNum)
    }

    // オフセット位置にシーク
    r.r.Seek(entry.offset, io.SeekStart)

    // 間接オブジェクトをパース
    parser := NewParser(r.r)
    num, gen, obj, err := parser.ParseIndirectObject()

    // キャッシュに保存
    r.objCache[objNum] = obj

    return obj, nil
}
```

### 4.4. 参照の解決

```go
// ResolveReference は参照を解決してオブジェクトを取得する
func (r *Reader) ResolveReference(ref *core.Reference) (core.Object, error) {
    return r.GetObject(ref.ObjectNumber)
}

// ResolveReferences は辞書内の参照を再帰的に解決する
func (r *Reader) ResolveReferences(dict core.Dictionary) error {
    for key, value := range dict {
        if ref, ok := value.(*core.Reference); ok {
            obj, err := r.ResolveReference(ref)
            if err != nil {
                return err
            }
            dict[key] = obj
        }
    }
    return nil
}
```

## 5. 公開API設計

### 5.1. 基本的な使い方

```go
package main

import (
    "fmt"
    "github.com/ryomak/gopdf"
)

func main() {
    // PDFファイルを開く
    doc, err := gopdf.Open("input.pdf")
    if err != nil {
        panic(err)
    }
    defer doc.Close()

    // ページ数を取得
    pageCount := doc.PageCount()
    fmt.Printf("Pages: %d\n", pageCount)

    // メタデータを取得
    info := doc.Info()
    fmt.Printf("Title: %s\n", info.Title)
    fmt.Printf("Author: %s\n", info.Author)

    // ページを取得（Phase 7で抽出機能を追加）
    page := doc.Page(0)
    // text, err := page.ExtractText() // Phase 7で実装
}
```

### 5.2. gopdf パッケージへの追加

```go
package gopdf

import (
    "io"
    "os"
    "github.com/ryomak/gopdf/internal/reader"
)

// Reader はPDFを読み込むための構造体
type Reader struct {
    r      *reader.Reader
    closer io.Closer
}

// Open はファイルパスからPDFを開く
func Open(path string) (*Reader, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }

    r, err := reader.NewReader(file)
    if err != nil {
        file.Close()
        return nil, err
    }

    return &Reader{
        r:      r,
        closer: file,
    }, nil
}

// OpenReader はio.ReadSeekerからPDFを開く
func OpenReader(r io.ReadSeeker) (*Reader, error) {
    rd, err := reader.NewReader(r)
    if err != nil {
        return nil, err
    }

    return &Reader{r: rd}, nil
}

// Close はリーダーをクローズする
func (r *Reader) Close() error {
    if r.closer != nil {
        return r.closer.Close()
    }
    return nil
}

// PageCount はページ数を返す
func (r *Reader) PageCount() int {
    count, _ := r.r.GetPageCount()
    return count
}

// Info はメタデータを返す
func (r *Reader) Info() Metadata {
    infoDict, _ := r.r.GetInfo()
    // Dictionary から Metadata 構造体に変換
    return Metadata{
        Title:  getString(infoDict, "Title"),
        Author: getString(infoDict, "Author"),
        // ...
    }
}

// Metadata はPDFメタデータ
type Metadata struct {
    Title    string
    Author   string
    Subject  string
    Keywords string
    Creator  string
    Producer string
}
```

## 6. テスト計画

### 6.1. ユニットテスト

#### 6.1.1. Lexerのテスト

```go
func TestLexer_NextToken(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected []Token
    }{
        {
            name:  "Integer",
            input: "42",
            expected: []Token{{Type: TokenInteger, Value: 42}},
        },
        {
            name:  "Real",
            input: "3.14",
            expected: []Token{{Type: TokenReal, Value: 3.14}},
        },
        {
            name:  "String",
            input: "(Hello)",
            expected: []Token{{Type: TokenString, Value: "Hello"}},
        },
        {
            name:  "Name",
            input: "/Type",
            expected: []Token{{Type: TokenName, Value: "Type"}},
        },
        {
            name:  "Dictionary",
            input: "<< /Type /Page >>",
            expected: []Token{
                {Type: TokenDictStart},
                {Type: TokenName, Value: "Type"},
                {Type: TokenName, Value: "Page"},
                {Type: TokenDictEnd},
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            lexer := NewLexer(strings.NewReader(tt.input))
            for _, expected := range tt.expected {
                token, err := lexer.NextToken()
                if err != nil {
                    t.Fatalf("Unexpected error: %v", err)
                }
                if token.Type != expected.Type {
                    t.Errorf("Type = %v, want %v", token.Type, expected.Type)
                }
                // Value検証
            }
        })
    }
}
```

#### 6.1.2. Parserのテスト

```go
func TestParser_ParseDictionary(t *testing.T) {
    input := "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] >>"
    parser := NewParser(strings.NewReader(input))

    dict, err := parser.ParseDictionary()
    if err != nil {
        t.Fatalf("Unexpected error: %v", err)
    }

    // /Type の検証
    if dict[core.Name("Type")] != core.Name("Page") {
        t.Error("Type should be Page")
    }

    // /Parent の検証（参照）
    // /MediaBox の検証（配列）
}
```

#### 6.1.3. Readerのテスト

```go
func TestReader_ParseXref(t *testing.T) {
    // 最小限のPDFを作成
    pdf := createMinimalPDF()

    reader, err := NewReader(bytes.NewReader(pdf))
    if err != nil {
        t.Fatalf("Failed to create reader: %v", err)
    }

    // xrefのエントリ数を確認
    if len(reader.xref) != 6 {
        t.Errorf("Expected 6 xref entries, got %d", len(reader.xref))
    }

    // Catalog取得
    catalog, err := reader.GetCatalog()
    if err != nil {
        t.Fatalf("Failed to get catalog: %v", err)
    }

    if catalog[core.Name("Type")] != core.Name("Catalog") {
        t.Error("Catalog Type should be Catalog")
    }
}
```

### 6.2. 統合テスト

```go
func TestOpen_RealPDF(t *testing.T) {
    // gopdfで生成したPDFを読み込む
    doc := gopdf.New()
    page := doc.AddPage(gopdf.A4, gopdf.Portrait)
    page.SetFont(font.Helvetica, 12)
    page.DrawText("Test", 100, 700)

    // PDFをバッファに書き込み
    var buf bytes.Buffer
    doc.WriteTo(&buf)

    // 読み込み
    reader, err := gopdf.OpenReader(bytes.NewReader(buf.Bytes()))
    if err != nil {
        t.Fatalf("Failed to open PDF: %v", err)
    }
    defer reader.Close()

    // ページ数確認
    if reader.PageCount() != 1 {
        t.Errorf("Expected 1 page, got %d", reader.PageCount())
    }
}
```

## 7. 実装スケジュール

### 7.1. Phase 6.1: PDF構造とオブジェクト仕様の調査
- ✅ pdf_spec_notes.md に既に記載済み
- 追加で必要な仕様の確認

### 7.2. Phase 6.2: Lexer実装（字句解析）
- internal/reader/lexer.go の実装
- トークン化のテスト

### 7.3. Phase 6.3: Parser実装（構文解析）
- internal/reader/parser.go の実装
- Dictionary, Array, Stream のパース
- テスト

### 7.4. Phase 6.4: Reader本体の実装
- internal/reader/reader.go の実装
- Xref table, Trailer の解析
- オブジェクト読み込みとキャッシュ

### 7.5. Phase 6.5: 公開API実装
- gopdf.Open, gopdf.OpenReader の実装
- Reader構造体とメソッド

### 7.6. Phase 6.6: テストとサンプル
- 統合テストの作成
- examples/06_read_pdf サンプルの作成

### 7.7. Phase 6.7: ドキュメント更新とコミット
- README.md の更新
- docs/progress.md の更新
- Git commit & push

## 8. 注意事項と制約

### 8.1. PDF 1.7に集中

Phase 6ではPDF 1.7（xref table形式）に集中する。以下は後回し：
- PDF 1.5以降のObject Streams (圧縮されたオブジェクト)
- PDF 2.0の新機能
- Linearized PDF（最適化されたWeb表示用PDF）

### 8.2. 暗号化PDFは未対応

暗号化されたPDFは将来の拡張として、Phase 6では非対応。

### 8.3. エラーハンドリング

不正なPDFに対して適切なエラーを返す：
- 構文エラー
- 不正なオブジェクト参照
- 破損したxref table

### 8.4. メモリ効率

大きなPDFファイルでもメモリ効率的に動作するよう：
- ストリームデータは必要になるまで読み込まない
- オブジェクトキャッシュのサイズ制限を検討

## 9. 参考資料

- [PDF 1.7 仕様書](https://opensource.adobe.com/dc-acrobat-sdk-docs/pdfstandards/PDF32000_2008.pdf)
- [pdf_spec_notes.md](./pdf_spec_notes.md) - 本プロジェクトのPDF仕様メモ
- [architecture.md](./architecture.md) - gopdfアーキテクチャ設計書
