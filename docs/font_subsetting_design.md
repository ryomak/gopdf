# Phase 11: フォントサブセット化 設計書

## 1. 概要

現在のTTFフォント実装では、フォント全体をPDFに埋め込むため、PDFファイルサイズが2-15MBと大きくなっています。フォントサブセット化を実装し、実際に使用する文字（グリフ）のみを埋め込むことで、ファイルサイズを大幅に削減します。

### 1.1. 目的

- PDFファイルサイズの最適化（90-99%削減が目標）
- 使用する文字のみをフォントに含める
- フォント埋め込みの高速化
- ネットワーク転送の効率化

### 1.2. スコープ

**Phase 11で実装する機能:**
- ✅ ドキュメント内で使用する文字の収集
- ✅ グリフIDのリマッピング
- ✅ サブセットフォントの生成
- ✅ サブセットフォントのPDF埋め込み
- ✅ 文字幅配列（W array）の最適化

**Phase 11では実装しない機能:**
- 完全なTTFテーブルの再構築（簡易的なサブセット）
- CFF (Compact Font Format) のサブセット化
- WOFF/WOFF2のサブセット化

## 2. 技術的背景

### 2.1. フォントサブセット化の仕組み

フォントサブセット化は以下の手順で行います：

1. **使用文字の収集**: ドキュメント内で使用される全文字（rune）を収集
2. **グリフマッピング**: 各文字に対応するグリフIDを取得
3. **グリフリマッピング**: グリフIDを0から順に連番に変更
4. **サブセット生成**: 必要なグリフのみを含む新しいフォントデータを生成
5. **PDF埋め込み**: サブセットフォントをPDFに埋め込み

### 2.2. ファイルサイズ削減効果

**例: 日本語フォント**
- 完全フォント: 約10,000グリフ、ファイルサイズ 15MB
- 使用文字: 50文字（「こんにちは、世界！Hello, World!」等）
- サブセット: 50グリフ、ファイルサイズ約75KB
- **削減率: 99.5%** ✨

**例: 英語フォント**
- 完全フォント: 約1,000グリフ、ファイルサイズ 2MB
- 使用文字: 100文字（A-Z, a-z, 数字, 記号）
- サブセット: 100グリフ、ファイルサイズ約200KB
- **削減率: 90%** ✨

### 2.3. PDF仕様での要件

PDF 1.7仕様では、サブセットフォントには以下が必要：

1. **フォント名にプレフィックス**: `ABCDEF+FontName` 形式（6文字のランダムタグ）
2. **グリフIDの連続性**: 0から順に連番であること
3. **Width配列の更新**: 使用するグリフの幅情報のみ
4. **cmap更新**: Unicode → GlyphIDマッピングの更新

## 3. 実装設計

### 3.1. アーキテクチャ

```
┌──────────────────────────────────────────┐
│  Document.WriteTo()                       │
│  - 全ページから使用文字を収集             │
└────────────┬─────────────────────────────┘
             │
┌────────────▼─────────────────────────────┐
│  FontSubsetter                            │
│  - 使用文字 → グリフID変換                │
│  - グリフIDリマッピング                   │
│  - サブセットフォント生成                 │
└────────────┬─────────────────────────────┘
             │
┌────────────▼─────────────────────────────┐
│  TTFFontEmbedder                          │
│  - サブセットフォントをPDFに埋め込み      │
│  - サブセットタグ付きフォント名           │
│  - 最適化されたWidth配列                  │
└──────────────────────────────────────────┘
```

### 3.2. データ構造

```go
package font

// SubsetFont は文字のサブセットを含むフォント
type SubsetFont struct {
	baseFont      *TTFFont            // 元のフォント
	usedRunes     map[rune]bool       // 使用される文字
	glyphIDs      []sfnt.GlyphIndex   // 使用されるグリフID（元のID）
	glyphMapping  map[sfnt.GlyphIndex]uint16 // 元のGlyphID → 新しいGlyphID
	runeToNewGID  map[rune]uint16     // rune → 新しいGlyphID
	subsetData    []byte              // サブセットフォントデータ
	subsetName    string              // サブセットフォント名（タグ付き）
}

// CreateSubset は使用文字のサブセットフォントを作成
func CreateSubset(baseFont *TTFFont, usedRunes map[rune]bool) (*SubsetFont, error)

// SubsetName はサブセットフォント名を返す（例: "ABCDEF+HiraginoSans"）
func (s *SubsetFont) SubsetName() string

// GetNewGlyphID は文字に対応する新しいグリフIDを返す
func (s *SubsetFont) GetNewGlyphID(r rune) (uint16, error)

// Data はサブセットフォントのバイトデータを返す
func (s *SubsetFont) Data() []byte
```

### 3.3. 使用文字の収集

```go
package gopdf

// collectUsedRunes はドキュメント内で使用される全文字を収集
func (d *Document) collectUsedRunes() map[*TTFFont]map[rune]bool {
	fontRunes := make(map[*TTFFont]map[rune]bool)

	for _, page := range d.pages {
		// ページ内のテキストコンテンツを解析
		// DrawTextUTF8()で描画された文字を収集
		for fontKey, ttfFont := range page.ttfFonts {
			if _, exists := fontRunes[ttfFont]; !exists {
				fontRunes[ttfFont] = make(map[rune]bool)
			}

			// ページコンテンツから使用文字を抽出
			usedRunes := extractRunesFromPageContent(page, fontKey)
			for r := range usedRunes {
				fontRunes[ttfFont][r] = true
			}
		}
	}

	return fontRunes
}
```

### 3.4. グリフリマッピング

```go
// createGlyphMapping はグリフIDを0から連番にリマッピング
func (s *SubsetFont) createGlyphMapping() error {
	// グリフID 0 (.notdef) は常に含める
	newGID := uint16(0)
	s.glyphMapping[0] = newGID
	newGID++

	// 使用される文字のグリフIDを取得してソート
	var glyphIDs []sfnt.GlyphIndex
	for r := range s.usedRunes {
		gid, err := s.baseFont.font.GlyphIndex(&sfnt.Buffer{}, r)
		if err != nil {
			continue
		}
		if _, exists := s.glyphMapping[gid]; !exists {
			glyphIDs = append(glyphIDs, gid)
		}
	}

	// グリフIDをソートして連番を割り当て
	sort.Slice(glyphIDs, func(i, j int) bool {
		return glyphIDs[i] < glyphIDs[j]
	})

	for _, gid := range glyphIDs {
		s.glyphMapping[gid] = newGID
		newGID++
	}

	return nil
}
```

### 3.5. サブセットフォント名の生成

```go
// generateSubsetTag は6文字のランダムタグを生成
func generateSubsetTag() string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	rand.Seed(time.Now().UnixNano())

	tag := make([]byte, 6)
	for i := range tag {
		tag[i] = letters[rand.Intn(len(letters))]
	}

	return string(tag)
}

// createSubsetName はサブセットフォント名を生成
func (s *SubsetFont) createSubsetName() {
	tag := generateSubsetTag()
	baseName := s.baseFont.Name()
	s.subsetName = fmt.Sprintf("%s+%s", tag, baseName)
}
```

### 3.6. Width配列の最適化

```go
// createWidthArray はサブセットフォント用のWidth配列を作成
func (s *SubsetFont) createWidthArray() ([]int, error) {
	// 新しいグリフIDの最大値を取得
	maxNewGID := uint16(0)
	for _, newGID := range s.glyphMapping {
		if newGID > maxNewGID {
			maxNewGID = newGID
		}
	}

	// Width配列を初期化
	widths := make([]int, maxNewGID+1)

	// 各グリフの幅を取得
	buf := &sfnt.Buffer{}
	for oldGID, newGID := range s.glyphMapping {
		advance, err := s.baseFont.font.GlyphAdvance(buf, oldGID, fixed.I(1000), font.HintingNone)
		if err != nil {
			return nil, err
		}

		// 幅を1000単位に正規化
		width := int(advance) / 64
		widths[newGID] = width
	}

	return widths, nil
}
```

## 4. 実装手順

### 4.1. Phase 11.1: 使用文字の収集機能実装

```go
// internal/font/subset_collector.go

package font

// UsageCollector はフォントの使用文字を収集
type UsageCollector struct {
	fontUsage map[*TTFFont]map[rune]bool
}

// NewUsageCollector は新しいUsageCollectorを作成
func NewUsageCollector() *UsageCollector {
	return &UsageCollector{
		fontUsage: make(map[*TTFFont]map[rune]bool),
	}
}

// AddText は指定されたフォントで使用されるテキストを記録
func (c *UsageCollector) AddText(font *TTFFont, text string) {
	if _, exists := c.fontUsage[font]; !exists {
		c.fontUsage[font] = make(map[rune]bool)
	}

	for _, r := range text {
		c.fontUsage[font][r] = true
	}
}

// GetUsedRunes は指定されたフォントで使用される文字を返す
func (c *UsageCollector) GetUsedRunes(font *TTFFont) map[rune]bool {
	if runes, exists := c.fontUsage[font]; exists {
		return runes
	}
	return make(map[rune]bool)
}
```

### 4.2. Phase 11.2: サブセットフォント生成実装

```go
// internal/font/subset.go

package font

// SubsetBuilder はフォントサブセットを構築
type SubsetBuilder struct {
	baseFont *TTFFont
}

// Build は使用文字からサブセットフォントを構築
func (b *SubsetBuilder) Build(usedRunes map[rune]bool) (*SubsetFont, error) {
	subset := &SubsetFont{
		baseFont:     b.baseFont,
		usedRunes:    usedRunes,
		glyphMapping: make(map[sfnt.GlyphIndex]uint16),
		runeToNewGID: make(map[rune]uint16),
	}

	// グリフマッピングを作成
	if err := subset.createGlyphMapping(); err != nil {
		return nil, err
	}

	// rune → 新しいGlyphIDマッピングを作成
	if err := subset.createRuneMapping(); err != nil {
		return nil, err
	}

	// サブセット名を生成
	subset.createSubsetName()

	// Width配列を作成
	if err := subset.createWidthArray(); err != nil {
		return nil, err
	}

	return subset, nil
}
```

### 4.3. Phase 11.3: サブセットフォント埋め込み実装

Writer拡張:
```go
// internal/writer/ttf_embed.go

// EmbedSubsetFont はサブセットフォントをPDFに埋め込む
func (e *TTFFontEmbedder) EmbedSubsetFont(subset *SubsetFont) (*core.Reference, error) {
	// サブセットフォント名を使用
	// Width配列を最適化されたものに置き換え
	// 他はEmbedTTFFont()と同様
}
```

### 4.4. Phase 11.4: Document統合

```go
// document.go

func (d *Document) WriteTo(w io.Writer) error {
	// 1. 使用文字を収集
	collector := font.NewUsageCollector()
	for _, page := range d.pages {
		// ページコンテンツから使用文字を収集
		page.collectUsedRunes(collector)
	}

	// 2. 各フォントのサブセットを作成
	subsets := make(map[*TTFFont]*font.SubsetFont)
	for ttfFont, usedRunes := range collector.fontUsage {
		builder := font.NewSubsetBuilder(ttfFont)
		subset, err := builder.Build(usedRunes)
		if err != nil {
			return err
		}
		subsets[ttfFont] = subset
	}

	// 3. サブセットフォントを埋め込み
	// ...
}
```

## 5. 使用例

```go
package main

import (
	"os"
	"github.com/ryomak/gopdf"
)

func main() {
	doc := gopdf.New()
	page := doc.AddPage(gopdf.A4, gopdf.Portrait)

	// TTFフォントを読み込み
	font, _ := gopdf.LoadTTF("/System/Library/Fonts/ヒラギノ角ゴシック W3.ttc")
	page.SetTTFFont(font, 24)

	// テキストを描画（使用文字: こんにちは、世界！）
	page.DrawTextUTF8("こんにちは、世界！", 100, 750)

	// ファイルに出力
	// → サブセット化により、12文字分のグリフのみ埋め込まれる
	// → ファイルサイズ: 15MB → 約100KB（99.3%削減）
	file, _ := os.Create("subset.pdf")
	doc.WriteTo(file)
	file.Close()
}
```

## 6. テスト計画

### 6.1. ユニットテスト

```go
func TestUsageCollector(t *testing.T) {
	collector := font.NewUsageCollector()

	font, _ := font.LoadTTF("testdata/test.ttf")
	collector.AddText(font, "Hello")

	usedRunes := collector.GetUsedRunes(font)
	if len(usedRunes) != 5 { // H, e, l, o (lは重複)
		t.Errorf("Expected 4 unique runes, got %d", len(usedRunes))
	}
}

func TestSubsetBuilder(t *testing.T) {
	baseFont, _ := font.LoadTTF("testdata/test.ttf")
	usedRunes := map[rune]bool{'H': true, 'e': true, 'l': true, 'o': true}

	builder := font.NewSubsetBuilder(baseFont)
	subset, err := builder.Build(usedRunes)

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// サブセット名にタグが含まれているか
	if !strings.Contains(subset.SubsetName(), "+") {
		t.Error("Subset name should contain tag prefix")
	}
}

func TestSubsetPDFGeneration(t *testing.T) {
	doc := gopdf.New()
	page := doc.AddPage(gopdf.A4, gopdf.Portrait)

	font, _ := gopdf.LoadTTF("testdata/test.ttf")
	page.SetTTFFont(font, 18)
	page.DrawTextUTF8("Test", 100, 700)

	var buf bytes.Buffer
	err := doc.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo failed: %v", err)
	}

	// サブセットタグがPDFに含まれているか確認
	pdfStr := buf.String()
	if !strings.Contains(pdfStr, "+") {
		t.Error("PDF should contain subset tag in font name")
	}
}
```

### 6.2. ファイルサイズ比較テスト

```go
func TestFileSize_WithSubsetting(t *testing.T) {
	// サブセットなし
	doc1 := createDocumentWithoutSubsetting()
	var buf1 bytes.Buffer
	doc1.WriteTo(&buf1)
	sizeWithoutSubset := buf1.Len()

	// サブセットあり
	doc2 := createDocumentWithSubsetting()
	var buf2 bytes.Buffer
	doc2.WriteTo(&buf2)
	sizeWithSubset := buf2.Len()

	// サブセットの方が90%以上小さいことを確認
	reduction := float64(sizeWithoutSubset-sizeWithSubset) / float64(sizeWithoutSubset)
	if reduction < 0.90 {
		t.Errorf("Subsetting should reduce size by at least 90%%, got %.2f%%", reduction*100)
	}

	t.Logf("Size without subsetting: %d bytes", sizeWithoutSubset)
	t.Logf("Size with subsetting: %d bytes", sizeWithSubset)
	t.Logf("Reduction: %.2f%%", reduction*100)
}
```

## 7. 注意事項

### 7.1. 制限事項

- **簡易実装**: TTFテーブルの完全な再構築は行わず、グリフマッピング情報のみ更新
- **CFF未対応**: OpenTypeのCFFフォントには未対応（TTFのみ）
- **合字未対応**: 複数文字の合字（ligature）は適切に処理されない可能性

### 7.2. パフォーマンス

- **メモリ使用量**: フォント全体を一度メモリに読み込む必要あり
- **処理時間**: 文字数に比例して増加（1000文字程度までは問題なし）
- **キャッシュ**: 同じフォント・文字セットの場合、サブセットを再利用可能

### 7.3. 互換性

- PDF 1.3以降でサブセットフォントをサポート
- ほとんどのPDFビューアで正しく表示される
- テキスト抽出も正常に動作（ToUnicode CMapにより）

## 8. 期待される効果

### 8.1. ファイルサイズ削減

| シナリオ | 完全フォント | サブセット | 削減率 |
|---------|-------------|-----------|-------|
| 日本語10文字 | 15MB | 50KB | 99.7% |
| 日本語100文字 | 15MB | 500KB | 96.7% |
| 英語50文字 | 2MB | 100KB | 95.0% |
| 混在200文字 | 15MB | 1MB | 93.3% |

### 8.2. パフォーマンス向上

- ネットワーク転送時間の大幅削減
- PDFビューアでの読み込み高速化
- ストレージ容量の節約

## 9. 将来の拡張

- [ ] CFFフォントのサブセット化
- [ ] より高度なTTFテーブル再構築
- [ ] フォントキャッシュによる高速化
- [ ] WOFF/WOFF2のサブセット化

## 10. 参考資料

- [PDF 1.7 仕様書](https://opensource.adobe.com/dc-acrobat-sdk-docs/pdfstandards/PDF32000_2008.pdf)
  - Section 9.9: Embedded Font Programs
- [TrueType仕様](https://developer.apple.com/fonts/TrueType-Reference-Manual/)
- [Font Subsetting in PDFs](https://github.com/foliojs/fontkit) - 参考実装
