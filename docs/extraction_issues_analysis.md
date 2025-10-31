# PDF抽出機能の問題分析

## 調査日
2025-10-31

## テスト対象
- ファイル: Receipt-2021-3422.pdf
- ページ数: 1ページ
- PageCTM: [0.24 0.00 0.00 -0.24 0.00 792.00]
  - スケール: 0.24 (24%)
  - Y軸反転: D = -0.24 (負の値)

## 発見された問題

### 1. ⚠️ 画像座標の異常値（最優先）

**現象:**
```
Block 0 [image]:
  Position: (1900.00, 101983.13)  ← 異常に大きい
  Transform: [126.00 0.00 0.00 30.75 1900.00 -101221.88]
            ↑ 正常     ↑ 正常            ↑ 異常なF値
```

**原因:**
1. PageCTM.D = -0.24 (Y軸反転あり)
2. Transform.F = -101221.88 (異常値)
3. layout.go:118で座標変換を実行：
   ```go
   convertedImageBlocks[i].Y = height - convertedImageBlocks[i].Y - convertedImageBlocks[i].PlacedHeight
   ```
4. しかし、元のY座標が既に異常値のため、変換後も異常値になる

**影響:**
- 画像がページ外の異常な位置に配置される
- SortedContentBlocks()で最初のブロックとして扱われる（Y座標が最大のため）
- PDF再構成時に画像が表示されない、または異常な位置に表示される

**対策案:**
1. **異常値の検出とフィルタリング**
   - ページ境界を大きく超える座標値を検出
   - 閾値を設定（例: -1000 < Y < pageHeight + 1000）
   - 異常値の場合、警告を出力してスキップまたはデフォルト位置に配置

2. **Transform行列の検証**
   - CTMのF値が異常な場合（|F| > pageHeight * 10など）を検出
   - 元のPDFの問題としてログに記録

3. **座標変換の改善**
   - PageCTMとローカルCTMを正しく合成
   - ページレベルの変換を考慮した座標計算

### 2. ⚠️ 文字化け（ToUnicode CMap問題）

**現象:**
```
Block 10:
  Text: "billing@suno.com : \x1b ó4\x009\x0010 Ciel101\n
         United States 9 Òl\n
         Cambridge, Massachusetts 02138 É\x1f è\n
         Floor 4 ‰1810012\n
         17 Dunster Street ÅÄ\n y\nSuno Bill to"
```

**文字化け部分:**
- `\x1b` (制御文字 ESC)
- `ó4\x009\x0010 Ciel101`
- `Òl`
- `É\x1f è`
- `‰1810012`
- `ÅÄ`

**原因:**
1. 元のPDFが複数の特殊フォント（F9-F19）を使用
2. これらのフォントにToUnicode CMAPが欠落または不正
3. pdf_reconstruction_status.mdでも同じ問題を指摘

**影響:**
- テキスト抽出時に既に文字化け
- 再生成PDFにも文字化けが引き継がれる
- 読み取り不能な文字が含まれる

**対策案:**
1. **ToUnicode CMAP処理の改善**
   - より多くのエンコーディング形式に対応
   - フォールバック処理の追加

2. **警告メッセージの追加**
   - ToUnicode CMAPがないフォントを検出時に警告
   - ユーザーに元PDFの問題を通知

3. **制御文字のフィルタリング**
   - 表示不可能な制御文字を除外
   - またはUnicode置換文字(U+FFFD)に変換

### 3. ✓ テキストブロックのグルーピング（概ね正常）

**結果:**
- 12個のテキストブロックに正しく分割
- 近接する行が適切にグループ化されている
- 行間の判定が機能している

**例:**
```
Block 2: "Visa - 2420 October 8, 2025 $10.00 2021\x003422"
Block 3: "Payment method Date Amount paid Receipt number"
```
これらは別ブロックとして正しく認識

### 4. ⚠️ 座標変換とソート順序

**PageCTMの影響:**
```
PageCTM: [0.24 0.00 0.00 -0.24 0.00 792.00]
```

**座標変換処理:**
layout.go:103-120でY軸反転を処理：
```go
if pageCTM != nil && pageCTM.D < 0 {
    // TextBlocks
    for i := range textBlocks {
        textBlocks[i].Rect.Y = height - textBlocks[i].Rect.Y - textBlocks[i].Rect.Height
        for j := range textBlocks[i].Elements {
            textBlocks[i].Elements[j].Y = height - textBlocks[i].Elements[j].Y
        }
    }
    // ImageBlocks
    for i := range convertedImageBlocks {
        convertedImageBlocks[i].Y = height - convertedImageBlocks[i].Y - convertedImageBlocks[i].PlacedHeight
    }
}
```

**問題:**
1. PageCTMのスケール(0.24)が考慮されていない
2. 画像の異常座標が正しく変換されない
3. 変換後の座標検証がない

**対策案:**
1. PageCTMの完全な変換行列を適用
2. 変換後の座標を検証（ページ境界内かチェック）
3. 異常値を検出した場合の処理を追加

### 5. ⚠️ SortedContentBlocks()の順序

**現在の順序:**
```
Block 0 [image]: Y=101983.13  ← 異常値のため最初に
Block 1 [text]:  Y=698.00
Block 2 [text]:  Y=570.00
...
Block 12 [text]: Y=30.50
```

**問題:**
- 画像の異常座標により、ソート順が崩れる
- 本来はページ内の適切な位置にあるべき

**対策:**
1. 異常値のブロックを除外してからソート
2. またはページ境界内のブロックのみをソート対象とする

## 修正の優先順位

### Phase 1: 画像座標の異常値対策（最優先）
1. 座標の異常値検出ロジックを追加
2. 異常値のフィルタリングまたは修正
3. 警告メッセージの出力

### Phase 2: 座標変換の改善
1. PageCTMの完全な適用
2. スケールを考慮した変換
3. 変換後の検証

### Phase 3: 文字化け対策
1. ToUnicode CMAP処理の改善
2. フォールバック処理
3. 制御文字のフィルタリング

### Phase 4: テキスト編集機能の設計
1. TextBlockの編集API設計
2. 座標計算の保持
3. PDF再構成時の適用

## テスト計画

### 1. 異常値検出テスト
```go
func TestDetectAbnormalCoordinates(t *testing.T) {
    // 異常に大きいY座標を検出できることを確認
}
```

### 2. 座標変換テスト
```go
func TestCoordinateTransformationWithPageCTM(t *testing.T) {
    // PageCTMを考慮した正しい座標変換を確認
}
```

### 3. 文字エンコーディングテスト
```go
func TestControlCharacterFiltering(t *testing.T) {
    // 制御文字のフィルタリングを確認
}
```

## 参考資料
- docs/pdf_reconstruction_status.md
- docs/coordinate_system_and_ctm_design.md
- docs/image_coordinate_issue.md
