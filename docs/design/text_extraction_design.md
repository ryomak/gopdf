# Phase 7: テキスト抽出 設計書

## 1. 概要

PDFファイルからテキストを抽出する機能を実装する。コンテンツストリームを解析し、テキスト描画オペレーターから文字列と位置情報を取得する。

### 1.1. 目的

- 既存PDFファイルからテキストを抽出
- ページ単位でのテキスト取得
- テキストの位置情報を考慮した抽出（読み順序の推定）
- Phase 6（PDF読み込み）の機能を拡張

### 1.2. スコープ

**Phase 7で実装する機能:**
- ✅ コンテンツストリームのパース
- ✅ テキスト描画オペレーターの解析
- ✅ テキストの抽出
- ✅ 基本的な位置情報の取得
- ✅ ページからのテキスト取得API

**Phase 7では実装しない機能:**
- 複雑なレイアウト解析（複数カラム、表など）
- フォントエンコーディングの完全な対応
- 縦書きテキスト
- テキストの詳細なスタイル情報（太字、斜体など）

## 2. PDFコンテンツストリームの技術的背景

### 2.1. コンテンツストリームとは

コンテンツストリームは、ページに描画する内容を記述したバイナリストリーム。PostScript言語風の命令（オペレーター）とオペランドで構成される。

**例:**
```
BT
/F1 12 Tf
100 700 Td
(Hello, World!) Tj
ET
```

### 2.2. テキスト描画オペレーター

| オペレーター | オペランド | 説明 |
|------------|-----------|------|
| `BT` | - | テキストオブジェクトの開始 |
| `ET` | - | テキストオブジェクトの終了 |
| `Tf` | font size | フォント選択 |
| `Td` | tx ty | テキスト位置移動 |
| `TD` | tx ty | テキスト位置移動（行送りも設定） |
| `Tm` | a b c d e f | テキストマトリックス設定 |
| `T*` | - | 次の行へ移動 |
| `Tj` | string | テキスト表示 |
| `TJ` | array | テキスト表示（配列、位置調整付き） |
| `'` | string | 次の行へ移動してテキスト表示 |
| `"` | aw ac string | 次の行へ移動してテキスト表示（間隔設定付き） |

### 2.3. テキスト状態パラメータ

テキスト描画には以下の状態が影響する：

- **Text matrix (Tm)**: テキストの位置と変形
- **Text line matrix (Tlm)**: 現在行の開始位置
- **Character spacing (Tc)**: 文字間隔
- **Word spacing (Tw)**: 単語間隔
- **Horizontal scaling (Tz)**: 水平スケーリング
- **Leading (TL)**: 行送り
- **Font (Tf)**: フォントとサイズ
- **Text rendering mode (Tr)**: 描画モード

### 2.4. 座標系

PDFの座標系は左下が原点 (0, 0)。テキストマトリックスで位置を指定。

```
[ a  b  0 ]
[ c  d  0 ]
[ e  f  1 ]
```

- `e, f`: テキストの位置（x, y）
- `a, d`: スケーリング
- `b, c`: 傾き（通常は0）

## 3. 実装アプローチ

### 3.1. アーキテクチャ

```
┌─────────────────────────────────────┐
│  gopdf.PDFReader.ExtractText(page)  │  ← 公開API
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│  internal/content/extractor.go      │  ← テキスト抽出
│  - TextExtractor                     │
│  - ExtractText(page) → []TextElement│
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│  internal/content/stream_parser.go  │  ← ストリームパーサー
│  - StreamParser                      │
│  - ParseContentStream() → Operations│
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│  internal/reader/reader.go          │  ← Reader拡張
│  - GetPageContents(page)            │
└─────────────────────────────────────┘
```

### 3.2. StreamParser設計

コンテンツストリームをパースし、オペレーションのリストを返す。

```go
package content

import (
    "github.com/ryomak/gopdf/internal/core"
)

// Operation はコンテンツストリームのオペレーション
type Operation struct {
    Operator string        // オペレーター名（例: "Tj", "Td"）
    Operands []core.Object // オペランド
}

// StreamParser はコンテンツストリームをパースする
type StreamParser struct {
    lexer *reader.Lexer
}

// NewStreamParser は新しいStreamParserを作成する
func NewStreamParser(data []byte) *StreamParser

// ParseOperations はコンテンツストリームからオペレーションを抽出する
func (p *StreamParser) ParseOperations() ([]Operation, error)
```

### 3.3. TextExtractor設計

オペレーションからテキストを抽出する。

```go
package content

// TextElement はテキスト要素
type TextElement struct {
    Text string  // テキスト内容
    X    float64 // X座標
    Y    float64 // Y座標
    Font string  // フォント名
    Size float64 // フォントサイズ
}

// TextExtractor はテキストを抽出する
type TextExtractor struct {
    operations []Operation

    // テキスト状態
    textMatrix   [6]float64 // Current text matrix
    lineMatrix   [6]float64 // Current line matrix
    currentFont  string
    fontSize     float64
    charSpacing  float64
    wordSpacing  float64
    leading      float64
}

// NewTextExtractor は新しいTextExtractorを作成する
func NewTextExtractor(operations []Operation) *TextExtractor

// Extract はテキストを抽出する
func (e *TextExtractor) Extract() ([]TextElement, error)
```

### 3.4. 公開API設計

```go
package gopdf

// TextElement はテキスト要素（公開型）
type TextElement struct {
    Text string
    X    float64
    Y    float64
}

// ExtractText はページからテキストを抽出する
func (r *PDFReader) ExtractText(pageNum int) ([]TextElement, error)

// ExtractPageText はページ全体のテキストを文字列として抽出する
func (r *PDFReader) ExtractPageText(pageNum int) (string, error)
```

## 4. 実装の詳細

### 4.1. コンテンツストリームの取得

```go
// Reader拡張: ページのコンテンツストリームを取得
func (r *Reader) GetPageContents(page core.Dictionary) ([]byte, error) {
    // /Contents を取得
    contentsObj := page[core.Name("Contents")]

    // Referenceの場合は解決
    if ref, ok := contentsObj.(*core.Reference); ok {
        obj, err := r.GetObject(ref.ObjectNumber)
        if err != nil {
            return nil, err
        }
        contentsObj = obj
    }

    // Streamの場合
    if stream, ok := contentsObj.(*core.Stream); ok {
        // Filterがある場合は展開
        return decompressStream(stream)
    }

    // 配列の場合（複数のストリーム）
    if array, ok := contentsObj.(core.Array); ok {
        var allData []byte
        for _, item := range array {
            // 各ストリームを取得して連結
            // ...
        }
        return allData, nil
    }

    return nil, fmt.Errorf("invalid contents")
}

// decompressStream はストリームを展開する
func decompressStream(stream *core.Stream) ([]byte, error) {
    filter := stream.Dict[core.Name("Filter")]

    if filter == nil {
        // フィルターなし
        return stream.Data, nil
    }

    if filter == core.Name("FlateDecode") {
        // Zlib展開
        return zlibDecompress(stream.Data)
    }

    // その他のフィルターは未対応
    return nil, fmt.Errorf("unsupported filter: %v", filter)
}
```

### 4.2. オペレーションのパース

```go
func (p *StreamParser) ParseOperations() ([]Operation, error) {
    var operations []Operation
    var operands []core.Object

    for {
        token, err := p.lexer.NextToken()
        if err == io.EOF {
            break
        }
        if err != nil {
            return nil, err
        }

        // キーワード（オペレーター）の場合
        if token.Type == reader.TokenKeyword {
            op := Operation{
                Operator: token.Value.(string),
                Operands: operands,
            }
            operations = append(operations, op)
            operands = nil
            continue
        }

        // オペランドを解析
        obj, err := p.parseOperand(token)
        if err != nil {
            return nil, err
        }
        operands = append(operands, obj)
    }

    return operations, nil
}
```

### 4.3. テキスト抽出ロジック

```go
func (e *TextExtractor) Extract() ([]TextElement, error) {
    var elements []TextElement

    // 初期化
    e.resetTextState()

    for _, op := range e.operations {
        switch op.Operator {
        case "BT": // Begin text
            e.resetTextMatrices()

        case "ET": // End text
            // テキストオブジェクト終了

        case "Tf": // Set font
            e.currentFont = op.Operands[0].(core.Name).String()
            e.fontSize = float64(op.Operands[1].(core.Real))

        case "Td": // Move text position
            tx := getNumber(op.Operands[0])
            ty := getNumber(op.Operands[1])
            e.moveText(tx, ty)

        case "Tm": // Set text matrix
            e.setTextMatrix(op.Operands)

        case "Tj": // Show text
            text := op.Operands[0].(core.String).String()
            elem := e.createTextElement(text)
            elements = append(elements, elem)

        case "TJ": // Show text with positioning
            array := op.Operands[0].(core.Array)
            for _, item := range array {
                if str, ok := item.(core.String); ok {
                    elem := e.createTextElement(str.String())
                    elements = append(elements, elem)
                }
            }
        }
    }

    return elements, nil
}

func (e *TextExtractor) createTextElement(text string) TextElement {
    return TextElement{
        Text: text,
        X:    e.textMatrix[4], // e
        Y:    e.textMatrix[5], // f
        Font: e.currentFont,
        Size: e.fontSize,
    }
}
```

## 5. テスト計画

### 5.1. ユニットテスト

#### 5.1.1. StreamParserのテスト

```go
func TestStreamParser_ParseOperations(t *testing.T) {
    tests := []struct {
        name     string
        stream   string
        expected []Operation
    }{
        {
            name: "Simple text",
            stream: "BT\n/F1 12 Tf\n100 700 Td\n(Hello) Tj\nET",
            expected: []Operation{
                {Operator: "BT", Operands: nil},
                {Operator: "Tf", Operands: []core.Object{core.Name("F1"), core.Integer(12)}},
                {Operator: "Td", Operands: []core.Object{core.Integer(100), core.Integer(700)}},
                {Operator: "Tj", Operands: []core.Object{core.String("Hello")}},
                {Operator: "ET", Operands: nil},
            },
        },
    }
    // ...
}
```

#### 5.1.2. TextExtractorのテスト

```go
func TestTextExtractor_Extract(t *testing.T) {
    operations := []Operation{
        {Operator: "BT"},
        {Operator: "Tf", Operands: []core.Object{core.Name("F1"), core.Real(12)}},
        {Operator: "Td", Operands: []core.Object{core.Real(100), core.Real(700)}},
        {Operator: "Tj", Operands: []core.Object{core.String("Hello")}},
        {Operator: "ET"},
    }

    extractor := NewTextExtractor(operations)
    elements, err := extractor.Extract()

    if err != nil {
        t.Fatalf("Extract failed: %v", err)
    }

    if len(elements) != 1 {
        t.Fatalf("Expected 1 element, got %d", len(elements))
    }

    if elements[0].Text != "Hello" {
        t.Errorf("Text = %q, want %q", elements[0].Text, "Hello")
    }
}
```

### 5.2. 統合テスト

```go
func TestPDFReader_ExtractText(t *testing.T) {
    // Writerで生成したPDFを読み込む
    doc := gopdf.New()
    page := doc.AddPage(gopdf.A4, gopdf.Portrait)
    page.SetFont(font.Helvetica, 12)
    page.DrawText("Test Text", 100, 700)

    var buf bytes.Buffer
    doc.WriteTo(&buf)

    // 読み込み
    reader, _ := gopdf.OpenReader(bytes.NewReader(buf.Bytes()))

    // テキスト抽出
    elements, err := reader.ExtractText(0)
    if err != nil {
        t.Fatalf("ExtractText failed: %v", err)
    }

    // 検証
    found := false
    for _, elem := range elements {
        if elem.Text == "Test Text" {
            found = true
            break
        }
    }

    if !found {
        t.Error("Expected text not found")
    }
}
```

## 6. 実装スケジュール

### 6.1. Phase 7.1: コンテンツストリーム仕様の調査
- ✅ PDF仕様書でコンテンツストリームの確認
- テキストオペレーターの仕様確認

### 6.2. Phase 7.2: StreamParser実装
- internal/content/stream_parser.go の実装
- オペレーションのパース
- テスト

### 6.3. Phase 7.3: TextExtractor実装
- internal/content/extractor.go の実装
- テキスト状態管理
- テキスト抽出ロジック
- テスト

### 6.4. Phase 7.4: 公開API実装
- PDFReader.ExtractText() の実装
- PDFReader.ExtractPageText() の実装
- Reader拡張（GetPageContents）

### 6.5. Phase 7.5: テストとサンプル
- 統合テストの作成
- examples/07_extract_text サンプルの作成

### 6.6. Phase 7.6: ドキュメント更新とコミット
- README.md の更新
- Git commit & push

## 7. 注意事項と制約

### 7.1. フォントエンコーディング

Phase 7では基本的なテキスト抽出に集中し、複雑なフォントエンコーディングは未対応：
- 標準エンコーディング（WinAnsiEncoding等）のみ対応
- CIDフォントは未対応
- カスタムエンコーディングは未対応

### 7.2. レイアウト解析

基本的な位置情報のみ取得：
- Y座標でソート → X座標でソート程度の簡易的な順序付け
- 複数カラム、表の認識は未対応
- 行の検出は簡易的

### 7.3. ストリーム展開

FlateDecode（Zlib）のみ対応：
- ASCIIHexDecode, ASCII85Decode等は未対応
- LZWDecode, RunLengthDecode等は未対応

## 8. 参考資料

- [PDF 1.7 仕様書](https://opensource.adobe.com/dc-acrobat-sdk-docs/pdfstandards/PDF32000_2008.pdf)
  - Section 9: Text（テキスト関連）
  - Section 14.8.2: Content Streams
- [pdf_spec_notes.md](./pdf_spec_notes.md)
