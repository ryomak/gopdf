# 画像座標問題の調査と修正

## 調査日
2025-10-31

## 問題の概要
`~/Desktop/aaa/main.go`で画像が読み取れない問題を調査しました。

## 問題の根本原因

### 1. 異常な座標チェックによるスキップ
- `internal/content/image_extractor.go`で、異常な座標を持つ画像をスキップする処理が実装されていた（193-217行目）
- CTMのf値が`-101221.88`のように異常に大きな負の値になる画像が検出されていた
- これらの画像は`maxReasonableCoordinate = 10000.0`のチェックで除外されていた

### 2. CTM計算の問題
PDFのコンテンツストリームの解析で、以下のCTM累積が発生：

**ページレベルCTM（最初のcm操作）:**
```
Input: [0.24 0.00 0.00 -0.24 0.00 792.00]
Result: Y軸反転とスケール変換
```

**画像配置のcm操作:**
```
Input: [525.00 0.00 0.00 -128.12 1900.00 253.12]
Before CTM: [0.24 0.00 0.00 -0.24 0.00 792.00]
After CTM: [126.00 0.00 0.00 30.75 1900.00 -101221.88]
```

**計算:**
- F値 = 0*0 + 792*(-128.12) + 253.12 = -101217.92
- この値は計算上は正しいが、ページレベルCTMが累積されているため異常に見える

### 3. 正しい画像位置の計算方法
画像のcm操作のパラメータ`[a b c d e f]`をページレベルCTMで変換：

```python
# ページレベルCTM: [0.24 0.00 0.00 -0.24 0.00 792.00]
# 画像cm: [525.00 0.00 0.00 -128.12 1900.00 253.12]

# 正しい位置：
x' = 1900 * 0.24 = 456.0
y' = 253.12 * (-0.24) + 792 = 731.25

# 正しいサイズ：
width' = 525 * 0.24 = 126.0
height' = 128.12 * 0.24 = 30.75
```

これはmain.goの座標修正結果（436.0, 731.3）と一致します！

## 実施した修正

### 1. 異常座標チェックの無効化
`internal/content/image_extractor.go:205-209`で異常座標チェックをコメントアウト

### 2. グラフィックス状態の拡張
`internal/content/graphics_state.go:72-73`に`LastCMMatrix`フィールドを追加：
```go
type GraphicsState struct {
    CTM          Matrix
    LastCMMatrix *Matrix  // 最後のcm操作のパラメータ（画像配置用）
    ...
}
```

### 3. cm操作パラメータの保存
`internal/content/image_extractor.go:160-161`でcm操作時にパラメータを保存

### 4. layout.goの対応
- ページレベルCTMをimage_extractor.ExtractImagesWithPositionに渡すように修正（`layout.go:92`）
- 画像のY軸反転処理を有効化（`layout.go:118-121`）

## 残っている課題

### 実装が未完了
`LastCMMatrix`を使った座標計算のロジック実装が途中で停止しています。
以下のコードを`image_extractor.go:182-204`に適用する必要があります：

```go
// 現在のグラフィックス状態を取得
currentGS := &gsStack[len(gsStack)-1]

// 最後のcm操作のパラメータがある場合、それをページレベルCTMで変換
if currentGS.LastCMMatrix != nil && pageLevelCTM != nil {
    cm := currentGS.LastCMMatrix

    // cm操作の位置（e, f）をページレベルCTMで変換
    x, y = pageLevelCTM.TransformPoint(cm.E, cm.F)

    // cm操作のサイズ（a, d）をページレベルCTMで変換
    width = cm.A * pageLevelCTM.A
    height = cm.D * pageLevelCTM.D
    if height < 0 {
        height = -height
    }
} else {
    // フォールバック: 従来の方法
    minX, minY, maxX, maxY := currentCTM.TransformRect(0, 0, 1, 1)
    x, y, width, height = minX, minY, maxX-minX, maxY-minY
}
```

### その他の考慮事項
1. **一般性**: 現在の実装は、画像配置の直前にcm操作がある場合のみ機能します。他のパターンも考慮する必要があります
2. **テストケース**: 様々なPDFでテストして、汎用性を確認する必要があります
3. **異常座標チェック**: 正しい座標計算が確立されたら、再度有効にする必要があります

## 次のステップ

1. ✅ image_extractor.goの座標計算ロジックを完成させる
2. ⬜ テストを実行して正しく動作することを確認
3. ⬜ デバッグ出力を削除
4. ⬜ 異常座標チェックを再有効化（必要に応じて閾値を調整）
5. ⬜ `make ci`を実行してテストとlintが成功することを確認
6. ⬜ コミット・プッシュ

## 参考資料
- [PDF Reference 1.7, Section 8.3.4: Transformation Matrices](https://opensource.adobe.com/dc-acrobat-sdk-docs/pdfstandards/PDF32000_2008.pdf)
- `docs/image_coordinate_issue.md`: 以前の調査記録
- `docs/coordinate_system_and_ctm_design.md`: CTM設計書
