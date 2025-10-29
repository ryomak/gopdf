# 画像座標抽出の問題

## 調査日
2025-10-29

## 問題の概要
PDFから画像を抽出する際、異常な座標値（例: x=1900.0, y=-101221.9）が計算され、画像が正しく配置されない。

## 再現方法
```bash
cd ~/Desktop/aaa
go run main.go Receipt-2021-3422.pdf output.pdf
```

出力：
```
--- ブロック 13 ---
Type: image
Position: (1900.0, -101221.9)  # 異常なY座標
Size: 126.0x30.7
Transform: a=126.00 b=0.00 c=0.00 d=30.75 e=1900.00 f=-101221.88
```

## 原因分析

### CTM（Current Transformation Matrix）
画像の配置は、CTM `[a b c d e f]` で決定されます：
- a, d: X/Y方向のスケール（画像のサイズ）
- b, c: 回転/シアー
- e, f: X/Yの変換（位置）

### 現在の実装（internal/content/image_extractor.go:182-191）
```go
minX, minY, maxX, maxY := currentCTM.TransformRect(0, 0, 1, 1)
x := minX  // = e = 1900.0
y := minY  // = f = -101221.9 (異常値)
```

### 問題のCTM
```
a=126.00 b=0.00 c=0.00 d=30.75 e=1900.00 f=-101221.88
```

- `e=1900.0`: X座標（ページ幅612を超えているが、まだ理解できる範囲）
- `f=-101221.88`: Y座標（ページ高さ792に対して異常に小さい負の値）

### 考えられる原因

#### 1. CTMの累積計算ミス
`cm`オペレーターが複数回適用される場合、CTMは以下のように累積されます：
```go
currentCTM = currentCTM.Multiply(newCTM)
```

グラフィックス状態スタック（`q`/`Q`オペレーター）の処理が正しく行われていない可能性があります。

#### 2. ページ変換行列の考慮不足
ページ全体に適用される変換行列（MediaBox, CropBox等）が考慮されていない可能性があります。

#### 3. 画像配置方法の誤解釈
PDFの画像配置には複数のパターンがあります：
- 通常: `[width 0 0 height x y]`
- 上下反転: `[width 0 0 -height x (y+height)]`
- 複雑な変換: 回転、シアーを含む

現在の実装は単純なケースのみを想定しています。

#### 4. 座標系の解釈ミス
PDF座標系は左下原点ですが、画像データは左上原点です。この変換が正しく行われていない可能性があります。

## 影響範囲
- ✅ テキスト抽出（正常動作）
- ❌ 画像抽出の座標計算（異常値）
- ❌ PDF再構成時の画像配置（ページ外に配置される）

## 回避策

### 一時的な対策（main.go）
```go
if drawY < -1000 || drawY > pageHeight+1000 {
    // 異常な座標の画像はスキップ
    fmt.Printf("警告: 画像の座標が異常なためスキップ\n")
    continue
}
```

### 推奨される対応
1. 元のPDFファイルをPDFビューアで開いて、画像が正しく表示されることを確認
2. pdftotext等の外部ツールで画像座標を確認
3. gopdfではなく、外部ツール（pdfimages等）で画像を抽出

## 今後の対応

### Phase 1: デバッグ情報の追加
`internal/content/image_extractor.go`にログ出力を追加：
- 各`cm`オペレーターの値
- グラフィックス状態スタックの状態
- 最終的なCTMの値

### Phase 2: CTM計算の検証
標準的なPDFファイルでCTM計算を検証：
- 単純な画像配置（回転なし）
- 回転した画像
- 上下反転した画像

### Phase 3: 修正実装
正しいCTM計算ロジックを実装：
- ページ変換行列の考慮
- グラフィックス状態スタックの正しい処理
- 座標系変換の明示化

### Phase 4: テストケースの追加
様々なパターンの画像配置をテスト：
- 通常配置
- 回転・反転
- 複雑な変換

## 関連資料
- PDF Reference 1.7, Section 4.2: Graphics Objects
- PDF Reference 1.7, Section 8.3.4: Transformation Matrices
- internal/content/image_extractor.go: 画像抽出処理
- internal/content/graphics_state.go: CTM計算
- docs/pdf_reconstruction_issue_investigation.md: テキスト座標問題（解決済み）

## 既知の制限
現時点では、一部のPDFファイルで画像座標が正しく抽出できません。PDFビューアで正しく表示される画像でも、gopdfで抽出すると異常な座標になる場合があります。

この問題は、ライブラリ側の修正が必要です。
