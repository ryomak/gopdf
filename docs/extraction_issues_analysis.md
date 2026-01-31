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

### 2. ⚠️ 日本語の文字化け（ToUnicode CMap不正問題）

**現象:**
```
元のPDF表示:           抽出結果:
上連雀4-9-10 Ciel101 → ó4910 Ciel101
栗栖              → Òl
〒1810012          → ‰1810012
東京都            → ÅÄ
```

**詳細分析:**
元のPDFを確認すると、PDFビューアでは**日本語が正しく表示**されています：
```
Cambridge, Massachusetts 02138
United States
billing@suno.com
Bill to
遼馬
栗栖
〒1810012
東京都
三鷹市
上連雀4-9-10 Ciel101
```

しかし、テキスト抽出すると**Latin文字に文字化け**します：
```
billing@suno.com :  ó4910 Ciel101
United States 9 Òl
Cambridge, Massachusetts 02138 É è
Floor 4 ‰1810012
17 Dunster Street ÅÄ
```

**根本原因:**
これは**ToUnicode CMAPの不正なマッピング**が原因です：

1. PDFのコンテンツストリームに `<0692>` などのグリフIDが記録
2. フォントのグリフテーブルには「上」の**形状データ**が正しく格納
3. しかし、ToUnicode CMAPには `0x0692 → U+00F3 (ó)` と**間違ったマッピング**

**PDFビューアと抽出の違い:**
```
PDFビューア（表示系）:
  グリフID 0x0692 → グリフ形状テーブル → "上"の形状 → 画面に表示 ✓

テキスト抽出（抽出系）:
  グリフID 0x0692 → ToUnicode CMap → U+00F3 (ó) → 文字化け ✗
```

**元のPDFの問題:**
- Type3フォント（F9-F19）を使用して日本語を描画
- ToUnicode CMAPが正しく生成されていない
- おそらくPDF生成ツール（Sunoの請求書システム）のバグ

**なぜ修正できないのか:**
1. グリフの形状から文字を逆引きする必要がある（OCR相当）
2. ToUnicode CMAPの情報が信頼できない
3. 元のPDFに正しいUnicode情報が存在しない

**影響:**
- テキスト抽出時に日本語が文字化け
- 検索、コピー、スクリーンリーダーで使用不可
- PDF再構成時も文字化けが引き継がれる

**実施した対策:**
1. ✅ **制御文字のフィルタリング** (layout.go:435, 518)
   - `\x1b`, `\x00`などの制御文字を除去
   - 表示可能な文字のみを残す

2. **ToUnicode CMAPは正しく読み取られている**
   - 実装は正常に動作
   - 問題はCMap自体の内容

**不可能な対策:**
1. ❌ **ToUnicode CMAPの修正**
   - グリフ形状からの文字認識が必要（OCR）
   - 自動化は困難

2. ❌ **正しい日本語への復元**
   - 元のPDFに正しい情報がない
   - 完全な復元は不可能

**推奨される対応:**
1. 元のPDFが**構造的な問題**を持っていることをユーザーに通知
2. PDF生成ツール（Suno）に報告することを推奨
3. 可能であれば、正しいToUnicode CMAPを持つPDFで再生成を依頼

**制限事項:**
- ✅ 通常のASCIIテキストは正しく抽出可能
- ✅ TTFフォントの日本語は正しく抽出可能（設計書通り実装済み）
- ❌ Type3フォント+不正なToUnicode CMAPは修正不可能

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
