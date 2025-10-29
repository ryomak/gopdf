# ToUnicode CMap Issue - TTF Font Text Extraction

## 調査日
2025-10-29

## 問題の概要
TTFフォントで書き込まれた日本語テキストをPDFから抽出すると、文字化けが発生する。

## 根本原因

### 書き込み時の処理（page.go:383-398）
```go
func (p *Page) textToGlyphIndices(text string, ttfFont *TTFFont) (string, error) {
    var result string
    for _, r := range text {
        // 文字 → グリフインデックスに変換
        glyphIndex, err := ttfFont.internal.GetGlyphIndex(r)
        if err != nil {
            return "", fmt.Errorf("failed to get glyph index: %w", err)
        }
        // 4桁の16進数としてエンコード
        result += fmt.Sprintf("%04X", glyphIndex)
    }
    return result, nil
}
```

**書き込まれる内容:**
- Unicode文字 → グリフインデックスに変換
- 例: '日' (U+65E5) → グリフインデックス 1234 → "1234" (16進数)

### ToUnicode CMapの問題（internal/writer/ttf_embed.go:150-212）

現在のToUnicode CMapは**Identity Mapping**を使用：
```
<0000> <FFFF> <0000>  // コード X を Unicode U+X にマッピング
```

これは「コードポイント → Unicode」のマッピングであり、「グリフインデックス → Unicode」ではない。

### 抽出時の処理（internal/content/extractor.go:263-284）

```go
func (e *TextExtractor) getTextString(obj core.Object) string {
    // ToUnicode CMapを使用してデコード
    if e.currentFontInfo != nil && e.currentFontInfo.ToUnicodeCMap != nil {
        result := e.currentFontInfo.ToUnicodeCMap.LookupString(data)
        if result != "" {
            return result
        }
    }
    // フォールバック
    return decodePDFString(data)
}
```

**抽出時の問題:**
1. PDFには「グリフインデックス」が格納されている（例: 1234）
2. ToUnicode CMapで1234を検索 → Identity Mappingでは U+1234（別の文字）を返す
3. 元の文字（'日' U+65E5）が復元できない

## 具体例

### 書き込み
- 入力: "日本語" (U+65E5, U+672C, U+8A9E)
- グリフインデックス: 1234, 5678, 9ABC
- PDFに格納: <12345678

9ABC> (16進数)

### 抽出（現在の実装）
- PDFから読み取り: <12345678>
- ToUnicode CMap（Identity）: U+1234 → 'ሴ', U+5678 → '噸', U+9ABC → 'ꪼ'
- 結果: **文字化け**

### 正しい動作
- ToUnicode CMap: グリフインデックス 1234 → U+65E5, 5678 → U+672C, 9ABC → U+8A9E
- 結果: "日本語" (正しく復元)

## 解決策

### Option 1: グリフベースのToUnicode CMap（推奨）

使用された文字を追跡し、グリフインデックスベースのToUnicode CMapを生成：

```go
// TTFFontに使用文字を追跡するフィールドを追加
type TTFFont struct {
    internal *font.TTFFont
    usedGlyphs map[uint16]rune  // グリフインデックス → Unicode
}

// DrawText時に使用文字を記録
func (p *Page) DrawText(text string, x, y float64) error {
    for _, r := range text {
        glyphIndex, _ := ttfFont.internal.GetGlyphIndex(r)
        ttfFont.usedGlyphs[glyphIndex] = r
    }
    // ...
}

// ToUnicode CMap生成時に使用文字のマッピングを追加
func (e *TTFFontEmbedder) createToUnicodeCMap(ttfFont *TTFFont) (*core.Reference, error) {
    var buf bytes.Buffer
    // ヘッダー
    buf.WriteString(...)

    // 使用されたグリフのマッピングを生成
    buf.WriteString(fmt.Sprintf("%d beginbfchar\n", len(ttfFont.usedGlyphs)))
    for glyphIndex, unicode := range ttfFont.usedGlyphs {
        fmt.Fprintf(&buf, "<%04X> <%04X>\n", glyphIndex, unicode)
    }
    buf.WriteString("endbfchar\n")
    // ...
}
```

### Option 2: CID-keyed Font（代替案）

CIDFontとしてTTFフォントを扱い、CID = Unicodeとする方法。ただし、これはPDF仕様に準拠するが、実装が複雑。

### Option 3: 標準フォントの使用（回避策）

日本語テキストには標準フォントを使わず、埋め込みTTFフォントのみを使用。

## 実装計画

### Phase 1: TTFFontの拡張
1. `TTFFont`構造体に`usedGlyphs map[uint16]rune`を追加
2. `DrawText`で使用文字を記録
3. `usedGlyphs`をスレッドセーフにする（sync.Mutex）

### Phase 2: ToUnicode CMap生成の修正
1. `createToUnicodeCMap`を修正
2. 使用されたグリフのみをマッピング
3. Identity Mappingを削除または補完として使用

### Phase 3: テストとドキュメント
1. 日本語テキストの書き込み・抽出テストを追加
2. 他の言語（中国語、韓国語、絵文字など）のテスト
3. READMEに制限事項を記載

## 回避策（現在）

現時点では、TTFフォントで書き込んだPDFから正しくテキストを抽出することはできません。

回避策：
1. PDF作成時に標準フォント（ASCII範囲）のみを使用
2. 抽出不要な場合はTTFフォントを使用
3. 外部ツール（pdftotext等）で抽出

## 影響範囲

- ✅ TTFフォントでのPDF生成（正常に動作）
- ✅ 標準フォントでのPDF生成・抽出（正常に動作）
- ❌ TTFフォント使用PDFからのテキスト抽出（文字化け）
- ❌ PDF翻訳機能（TTFフォント使用時）
- ❌ PDF検索機能（TTFフォント使用時）

## 参考資料

- PDF Reference 1.7, Section 9.10: Extraction of Text Content
- PDF Reference 1.7, Section 5.9: ToUnicode CMaps
- https://github.com/jung-kurt/gofpdf/issues/234 (similar issue in gofpdf)
- Adobe Technical Note #5014: CMap and CIDFont Files Specification

## 次のステップ

1. GitHubイシューを作成
2. Phase 1の実装（TTFFontの拡張）
3. Phase 2の実装（ToUnicode CMap修正）
4. テストケースの追加
5. ドキュメントの更新
