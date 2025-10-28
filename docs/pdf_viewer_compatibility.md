# PDF Viewer Compatibility Notes

## TTF Font Embedding Issues

### macOS Preview

一部のmacOS Previewバージョンで、TTFフォントを使用した日本語PDFが正しく表示されない問題が報告されています。

**症状:**
- テキストが白/透明で表示されない
- ただし、テキストの選択・コピーは可能
- ToUnicodeマップは正常に機能

**原因:**
macOS Previewの特定のバージョンが、以下のTTF埋め込み方式に完全対応していない可能性：
- Type0/CIDFont/CIDFontType2の組み合わせ
- Identity-Hエンコーディング

**実装済みの対策:**
1. テキスト色の明示的な設定 (`0 0 0 rg`)
2. ToUnicodeマップの完全なidentity mapping
3. CIDToGIDMap: Identityの追加
4. FontFile2によるフォントファイル全体の埋め込み

**代替ビューア:**
以下のPDFビューアでは正常に表示されることを確認：
- Adobe Acrobat Reader
- Google Chrome
- Firefox
- Safari（ブラウザ）

**検証方法:**
```bash
# テキスト抽出が正常に動作するか確認
go run examples/10_pdf_translation/main.go
open -a 'Google Chrome' examples/10_pdf_translation/japanese.pdf
```

### 今後の改善案

1. **W array (glyph widths) の実装**
   - 現在は DW (default width) のみ使用
   - 各グリフの正確な幅を設定することで互換性が向上する可能性

2. **代替フォント埋め込み方式の検討**
   - Type3フォントの使用
   - CFF（Compact Font Format）フォントの使用

3. **PDF/A準拠**
   - より厳密なPDF標準に準拠することで互換性向上

## 関連Issue

- #10_pdf_translation: Japanese PDF rendering issues in macOS Preview
- Commit: d431074, 2c5853b, 8584f29

## 参考資料

- PDF Reference 1.7 - Section 9: Text
- Adobe Technical Note #5411: ToUnicode Mapping File Tutorial
- PDF Reference - CIDFont Dictionaries
