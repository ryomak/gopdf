# ToUnicodeマッピングサポート 設計書

## 1. 概要

PDFのTTFフォントやCIDフォントから日本語などのUnicodeテキストを正しく抽出するため、ToUnicodeマッピング（CMap）のサポートを追加する。

### 1.1. 問題

現在の実装では、TTFフォントで書かれた日本語テキストが文字化けする：
- CIDフォント（Character ID）の数値がそのまま出力される
- ToUnicodeマッピングを参照していない
- 例：`こんにちは` → `\x00&\x00O\x00H...` のような文字化け

### 1.2. 原因

TTFフォントやCIDフォントでは：
1. テキストがCID（Character ID）としてエンコードされる
2. CIDからUnicodeへの変換にはToUnicode CMapが必要
3. 現在の`TextExtractor`はToUnicodeマッピングを読み取っていない

## 2. ToUnicodeマッピングとは

### 2.1. PDFでの表現

```
/Font <<
  /Type /Font
  /Subtype /Type0
  /BaseFont /HiraginoSans-W3
  /Encoding /Identity-H
  /DescendantFonts [3 0 R]
  /ToUnicode 4 0 R  ← これがToUnicode CMapへの参照
>>
```

### 2.2. ToUnicode CMapの形式

```
/CIDInit /ProcSet findresource begin
12 dict begin
begincmap
/CIDSystemInfo <<
  /Registry (Adobe)
  /Ordering (UCS)
  /Supplement 0
>> def
/CMapName /Adobe-Identity-UCS def
/CMapType 2 def
1 begincodespacerange
<0000> <ffff>
endcodespacerange

100 beginbfchar
<0026> <0048>  # CID 0x0026 → Unicode 0x0048 (H)
<004f> <0065>  # CID 0x004f → Unicode 0x0065 (e)
<0048> <006c>  # CID 0x0048 → Unicode 0x006c (l)
...
<035c> <3053>  # CID 0x035c → Unicode 0x3053 (こ)
<039c> <3093>  # CID 0x039c → Unicode 0x3093 (ん)
endbfchar

10 beginbfrange
<1000> <10ff> <4e00>  # 範囲マッピング
endbfrange

endcmap
```

## 3. 設計

### 3.1. アーキテクチャ

```
┌─────────────────┐
│   TextExtractor │
└────────┬────────┘
         │
         │ 使用
         ▼
┌─────────────────┐
│  FontManager    │ ← 新規追加
│  - フォント情報  │
│  - ToUnicodeマップ│
└────────┬────────┘
         │
         │ 読み込み
         ▼
┌─────────────────┐
│   ToUnicodeCMap │ ← 新規追加
│  - CID→Unicode  │
│  - パーサー      │
└─────────────────┘
```

### 3.2. データ構造

#### 3.2.1. ToUnicodeCMap

```go
package content

// ToUnicodeCMap はCIDからUnicodeへのマッピング
type ToUnicodeCMap struct {
	// bfchar: 個別のCIDマッピング
	charMap map[uint16]rune

	// bfrange: 範囲マッピング
	ranges []CIDRange
}

// CIDRange はCIDの範囲マッピング
type CIDRange struct {
	StartCID  uint16
	EndCID    uint16
	StartChar rune
}

// Lookup はCIDをUnicodeに変換
func (cm *ToUnicodeCMap) Lookup(cid uint16) (rune, bool) {
	// 1. charMapで検索
	if r, ok := cm.charMap[cid]; ok {
		return r, true
	}

	// 2. rangesで検索
	for _, rang := range cm.ranges {
		if cid >= rang.StartCID && cid <= rang.EndCID {
			offset := cid - rang.StartCID
			return rang.StartChar + rune(offset), true
		}
	}

	return 0, false
}
```

#### 3.2.2. FontInfo

```go
package content

// FontInfo はフォント情報
type FontInfo struct {
	Name         string
	ToUnicodeCMap *ToUnicodeCMap // nilの場合は通常のエンコーディング
}
```

### 3.3. TextExtractorの拡張

```go
type TextExtractor struct {
	operations []Operation
	reader     *reader.Reader  // ← 追加: フォント情報取得用
	page       core.Dictionary // ← 追加: ページリソース

	// テキスト状態
	textMatrix   [6]float64
	lineMatrix   [6]float64
	currentFont  string
	fontSize     float64
	fontInfo     *FontInfo // ← 追加: 現在のフォント情報
	// ...
}

// NewTextExtractor は新しいTextExtractorを作成する
func NewTextExtractor(operations []Operation, reader *reader.Reader, page core.Dictionary) *TextExtractor {
	return &TextExtractor{
		operations: operations,
		reader:     reader,
		page:       page,
	}
}
```

### 3.4. ToUnicode CMapのパース

```go
package content

// ParseToUnicodeCMap はToUnicode CMapをパースする
func ParseToUnicodeCMap(data []byte) (*ToUnicodeCMap, error) {
	cmap := &ToUnicodeCMap{
		charMap: make(map[uint16]rune),
	}

	// CMapをパース
	// beginbfchar/endbfchar の間を解析
	// beginbfrange/endbfrange の間を解析

	// 簡易実装：正規表現またはシンプルなパーサーで解析
	// <XXXX> <YYYY> の形式を抽出

	return cmap, nil
}

// parseBFChar は beginbfchar セクションをパース
func parseBFChar(data []byte) (map[uint16]rune, error) {
	// <0026> <0048> のようなペアを抽出
	// 最初の数値がCID、2番目がUnicode
	return nil, nil
}

// parseBFRange は beginbfrange セクションをパース
func parseBFRange(data []byte) ([]CIDRange, error) {
	// <1000> <10ff> <4e00> のような範囲を抽出
	return nil, nil
}
```

## 4. 実装計画

### 4.1. Phase 1: ToUnicode CMapパーサー

1. `internal/content/tounicode.go` を作成
2. `ToUnicodeCMap` 構造体を実装
3. `ParseToUnicodeCMap` 関数を実装
4. テストを作成

### 4.2. Phase 2: FontInfo管理

1. `internal/content/font_info.go` を作成
2. `FontInfo` 構造体を実装
3. フォントリソースからToUnicodeを取得する機能

### 4.3. Phase 3: TextExtractorの統合

1. `TextExtractor` にReaderとPageを追加
2. `Tf`命令時にフォント情報を読み込み
3. `Tj`/`TJ`命令時にToUnicodeマッピングを適用
4. 既存のエンコーディング処理と統合

### 4.4. Phase 4: テストと検証

1. TTFフォントでのテキスト抽出テスト
2. 日本語・英語混在テストcase
3. 既存テストの動作確認

## 5. API変更

### 5.1. 内部API（破壊的変更）

```go
// 変更前
func NewTextExtractor(operations []Operation) *TextExtractor

// 変更後
func NewTextExtractor(operations []Operation, reader *reader.Reader, page core.Dictionary) *TextExtractor
```

### 5.2. 公開API（変更なし）

```go
// reader.goの公開APIは変更なし
func (r *PDFReader) ExtractPageTextElements(pageNum int) ([]TextElement, error)
func (r *PDFReader) ExtractPageTextBlocks(pageNum int) ([]TextBlock, error)
```

## 6. 注意事項

### 6.1. パフォーマンス

- ToUnicode CMapのパースはキャッシュする
- フォントごとに1回だけパース

### 6.2. 互換性

- ToUnicodeがないフォントは従来の処理を使用
- 標準フォントには影響なし

### 6.3. 制限事項

- 複雑なCMap（縦書き、合成文字など）は初期実装では未対応
- 基本的なbfchar/bfrangeのみサポート

## 7. 参考資料

- [PDF 1.7 仕様書](https://opensource.adobe.com/dc-acrobat-sdk-docs/pdfstandards/PDF32000_2008.pdf)
  - Section 9.10: Extraction of Text Content
  - Section 9.10.3: ToUnicode CMaps
- [Adobe CMap and CIDFont Files Specification](https://www.adobe.com/content/dam/acom/en/devnet/font/pdfs/5014.CIDFont_Spec.pdf)
