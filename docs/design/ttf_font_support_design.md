# Phase 10: TTFフォント対応 設計書

## 1. 概要

TrueType Font (TTF) およびOpenType Font (OTF) のサポートを実装し、日本語を含む多言語テキストの描画を可能にする。

### 1.1. 目的

- TrueTypeフォントの読み込みと解析
- PDFへのフォント埋め込み
- Unicode文字列の描画
- 日本語・中国語・韓国語などの多言語対応
- カスタムフォントの使用

### 1.2. スコープ

**Phase 10で実装する機能:**
- ✅ TTF/OTFファイルのパース
- ✅ PDFへのフォント埋め込み（Type0/CIDFont）
- ✅ Unicode文字列の描画
- ✅ グリフメトリクス（文字幅）の取得
- ✅ 基本的な日本語テキスト描画

**Phase 10では実装しない機能:**
- フォントサブセット化（将来の最適化）
- 縦書き
- 複雑なグリフ配置（OpenTypeのGSUB/GPOS）
- 合字（ligature）
- カーニング

## 2. 技術的背景

### 2.1. TrueTypeフォントの構造

TTFファイルは複数のテーブルで構成される：

| テーブル | 必須 | 説明 |
|---------|------|------|
| head | ✅ | フォントヘッダー |
| hhea | ✅ | 水平ヘッダー |
| maxp | ✅ | 最大値プロファイル |
| name | ✅ | 名前テーブル |
| cmap | ✅ | 文字コードマッピング |
| glyf | ✅ | グリフデータ |
| loca | ✅ | グリフ位置 |
| hmtx | ✅ | 水平メトリクス |
| post | ✅ | PostScript情報 |
| OS/2 | 推奨 | OS/2とWindows固有メトリクス |

### 2.2. PDFでのTrueTypeフォント埋め込み

PDFでは2つの方法でTrueTypeフォントを埋め込む：

#### 2.2.1. Simple Font（非推奨）
- 1バイトエンコーディング
- ASCII/Latin-1のみ
- 日本語など非対応

#### 2.2.2. Composite Font（推奨）
- Type0フォント
- CIDFont (Character ID Font)
- Unicode対応
- 多言語対応

```
Type0 Font
  ├─ DescendantFont (CIDFont)
  │   └─ FontDescriptor
  │       └─ FontFile2 (埋め込みTTFデータ)
  └─ Encoding (Identity-H)
```

### 2.3. CIDフォントとUnicode

**CID (Character ID)**
- グリフの一意識別子
- PDFの内部表現

**Unicode → CID変換**
- cmapテーブルを使用
- ToUnicode CMapをPDFに埋め込み

## 3. 実装設計

### 3.1. アーキテクチャ

```
┌─────────────────────────────────────┐
│  gopdf.LoadTTF(path) → *TTFFont     │  ← 公開API
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│  internal/font/ttf.go                │  ← TTFローダー
│  - TTFFont struct                    │
│  - LoadTTF(path)                     │
│  - GetGlyphWidth(rune)               │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│  golang.org/x/image/font/sfnt       │  ← 外部ライブラリ
│  - TTFパース                         │
│  - グリフ情報                        │
└─────────────────────────────────────┘
```

### 3.2. データ構造

```go
package gopdf

// TTFFont はTrueTypeフォント
type TTFFont struct {
	name     string
	data     []byte          // フォントファイルの内容
	font     *sfnt.Font      // パースされたフォント
	glyphMap map[rune]uint16 // rune → GlyphID
}

// LoadTTF はTTFフォントを読み込む
func LoadTTF(path string) (*TTFFont, error)

// LoadTTFFromBytes はバイト列からTTFフォントを読み込む
func LoadTTFFromBytes(data []byte) (*TTFFont, error)

// Name はフォント名を返す
func (f *TTFFont) Name() string

// GlyphWidth はルーン（文字）の幅を返す
func (f *TTFFont) GlyphWidth(r rune, size float64) float64
```

### 3.3. PDFへの埋め込み

```go
package writer

// embedTTFFont はTTFフォントをPDFに埋め込む
func (w *Writer) embedTTFFont(font *TTFFont) (*FontObject, error) {
	// 1. FontDescriptorオブジェクト作成
	fontDescriptor := w.createFontDescriptor(font)

	// 2. FontFile2ストリーム作成（TTFデータ）
	fontFile := w.createFontFile2(font.data)

	// 3. CIDFontオブジェクト作成
	cidFont := w.createCIDFont(font, fontDescriptor)

	// 4. Type0フォントオブジェクト作成
	type0Font := w.createType0Font(font, cidFont)

	// 5. ToUnicode CMap作成
	toUnicode := w.createToUnicodeCMap(font)

	return type0Font, nil
}

// createType0Font はType0フォントオブジェクトを作成
func (w *Writer) createType0Font(font *TTFFont, cidFont *core.Reference) *core.Dictionary {
	return &core.Dictionary{
		core.Name("Type"):            core.Name("Font"),
		core.Name("Subtype"):         core.Name("Type0"),
		core.Name("BaseFont"):        core.Name(font.Name()),
		core.Name("Encoding"):        core.Name("Identity-H"),
		core.Name("DescendantFonts"): core.Array{cidFont},
		core.Name("ToUnicode"):       toUnicodeRef,
	}
}

// createCIDFont はCIDFontオブジェクトを作成
func (w *Writer) createCIDFont(font *TTFFont, descriptor *core.Reference) *core.Dictionary {
	return &core.Dictionary{
		core.Name("Type"):           core.Name("Font"),
		core.Name("Subtype"):        core.Name("CIDFontType2"),
		core.Name("BaseFont"):       core.Name(font.Name()),
		core.Name("CIDSystemInfo"):  w.createCIDSystemInfo(),
		core.Name("FontDescriptor"): descriptor,
		core.Name("W"):              w.createWidthArray(font),
	}
}
```

### 3.4. ToUnicode CMap

```go
// createToUnicodeCMap はToUnicode CMapを生成
func (w *Writer) createToUnicodeCMap(font *TTFFont) []byte {
	var buf bytes.Buffer

	buf.WriteString(`/CIDInit /ProcSet findresource begin
12 dict begin
begincmap
/CIDSystemInfo
<< /Registry (Adobe)
/Ordering (UCS)
/Supplement 0
>> def
/CMapName /Adobe-Identity-UCS def
/CMapType 2 def
1 begincodespacerange
<0000> <FFFF>
endcodespacerange
`)

	// Unicode → CID マッピング
	buf.WriteString("100 beginbfchar\n")
	// ... マッピングを書き込む
	buf.WriteString("endbfchar\n")

	buf.WriteString(`endcmap
CMapName currentdict /CMap defineresource pop
end
end`)

	return buf.Bytes()
}
```

## 4. 実装手順

### 4.1. Phase 10.1: TTFパーサー実装

```go
package font

import (
	"golang.org/x/image/font/sfnt"
)

type TTFFont struct {
	name     string
	data     []byte
	font     *sfnt.Font
	glyphMap map[rune]uint16
}

func LoadTTF(path string) (*TTFFont, error) {
	// 1. ファイル読み込み
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// 2. sfntでパース
	font, err := sfnt.Parse(data)
	if err != nil {
		return nil, err
	}

	// 3. cmapテーブルからglyphMapを構築
	glyphMap := make(map[rune]uint16)
	// ... cmapの解析

	return &TTFFont{
		name:     extractFontName(font),
		data:     data,
		font:     font,
		glyphMap: glyphMap,
	}, nil
}
```

### 4.2. Phase 10.2: フォント埋め込み実装

Writer拡張:
```go
func (w *Writer) EmbedTTFFont(font *TTFFont) error {
	// FontDescriptor, CIDFont, Type0Font オブジェクト作成
	// ...
}
```

### 4.3. Phase 10.3: Unicode対応

```go
// Page拡張
func (p *Page) DrawTextUTF8(text string, x, y float64) error {
	// UTF-8文字列をruneに変換
	// 各runeのグリフIDを取得
	// CID配列として出力
}
```

### 4.4. Phase 10.4: 日本語テキスト描画

```go
// SetTTFFont はTTFフォントを設定
func (p *Page) SetTTFFont(font *TTFFont, size float64) {
	p.currentTTFFont = font
	p.fontSize = size
}

// DrawJapaneseText は日本語テキストを描画
func (p *Page) DrawJapaneseText(text string, x, y float64) {
	// TTFフォントを使用してUnicodeテキストを描画
}
```

## 5. 使用例

```go
package main

import (
	"github.com/ryomak/gopdf"
)

func main() {
	doc := gopdf.New()
	page := doc.AddPage(gopdf.A4, gopdf.Portrait)

	// TTFフォント読み込み
	font, err := gopdf.LoadTTF("/System/Library/Fonts/ヒラギノ角ゴシック W3.ttc")
	if err != nil {
		panic(err)
	}

	// フォント設定
	page.SetTTFFont(font, 24)

	// 日本語テキスト描画
	page.DrawText("こんにちは、世界！", 100, 750)
	page.DrawText("Hello, World!", 100, 720)

	// 出力
	file, _ := os.Create("japanese.pdf")
	doc.WriteTo(file)
}
```

## 6. テスト計画

### 6.1. ユニットテスト

```go
func TestLoadTTF(t *testing.T) {
	font, err := font.LoadTTF("testdata/test.ttf")
	if err != nil {
		t.Fatalf("LoadTTF failed: %v", err)
	}

	if font.Name() == "" {
		t.Error("Font name is empty")
	}
}

func TestGlyphWidth(t *testing.T) {
	font, _ := font.LoadTTF("testdata/test.ttf")

	width := font.GlyphWidth('A', 12)
	if width <= 0 {
		t.Error("Glyph width should be positive")
	}
}

func TestJapaneseText(t *testing.T) {
	doc := gopdf.New()
	page := doc.AddPage(gopdf.A4, gopdf.Portrait)

	font, err := gopdf.LoadTTF("testdata/japanese.ttf")
	if err != nil {
		t.Skip("Japanese font not available")
	}

	page.SetTTFFont(font, 12)
	page.DrawText("こんにちは", 100, 700)

	var buf bytes.Buffer
	err = doc.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo failed: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("PDF output is empty")
	}
}
```

## 7. 注意事項

### 7.1. フォントライセンス

- フォントファイルの著作権に注意
- 埋め込み許可のあるフォントのみ使用
- 商用利用時はライセンス確認必須

### 7.2. ファイルサイズ

- TTFフォント全体を埋め込むとサイズが大きくなる（数MB）
- 将来的にサブセット化を実装して最適化
- 現在は使用する文字のみ埋め込む簡易サブセット

### 7.3. 互換性

- PDF 1.3以降でType0フォントをサポート
- Identity-Hエンコーディングを使用
- ほとんどのPDFビューアで表示可能

## 8. 将来の拡張

- [ ] フォントサブセット化（ファイルサイズ削減）
- [ ] 縦書き対応
- [ ] OpenType機能（GSUB/GPOS）
- [ ] Web フォント（WOFF/WOFF2）対応
- [ ] フォントキャッシュ

## 9. 参考資料

- [PDF 1.7 仕様書](https://opensource.adobe.com/dc-acrobat-sdk-docs/pdfstandards/PDF32000_2008.pdf)
  - Section 9.6: Type 0 Fonts
  - Section 9.7: Composite Fonts
  - Section 9.10: Embedded Font Programs
- [TrueType Font仕様](https://developer.apple.com/fonts/TrueType-Reference-Manual/)
- [OpenType仕様](https://docs.microsoft.com/en-us/typography/opentype/spec/)
- [golang.org/x/image/font/sfnt](https://pkg.go.dev/golang.org/x/image/font/sfnt)
