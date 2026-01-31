# プレゼンテーション用ページサイズ設計書

## 1. 概要

PDF生成においてプレゼンテーション用のページサイズを追加する。
PowerPointやKeynoteなどのプレゼンテーションツールで使用される標準的なアスペクト比に対応する。

## 2. 要件

### 2.1. 対応するアスペクト比

- **16:9 (Widescreen)**: 現代的なプレゼンテーションの標準
- **4:3 (Standard)**: 従来型プレゼンテーションの標準

### 2.2. ページサイズの決定

PowerPointの標準サイズを参考に、以下のサイズを採用：

#### 16:9 Widescreen
- 幅: 10 inch = 720 points (1 inch = 72 points)
- 高さ: 5.625 inch = 405 points
- アスペクト比: 16:9 = 1.777...

#### 4:3 Standard
- 幅: 10 inch = 720 points
- 高さ: 7.5 inch = 540 points
- アスペクト比: 4:3 = 1.333...

## 3. 実装設計

### 3.1. 定数の追加

`constants.go` に以下の定数を追加：

```go
var (
    // PageSizePresentation16x9 size: 10in x 5.625in (Widescreen)
    PageSizePresentation16x9 = PageSize{Width: 720.0, Height: 405.0}

    // PageSizePresentation4x3 size: 10in x 7.5in (Standard)
    PageSizePresentation4x3 = PageSize{Width: 720.0, Height: 540.0}
)
```

### 3.2. 命名規則

- `PageSizePresentation16x9`: 16:9のワイドスクリーン形式
- `PageSizePresentation4x3`: 4:3の標準形式
- プレフィックス `PageSizePresentation` で統一

### 3.3. 既存の構造体との互換性

既存の `PageSize` 構造体（Width, Height）をそのまま使用するため、
既存コードとの互換性は完全に保たれる。

```go
type PageSize struct {
    Width  float64
    Height float64
}
```

## 4. 使用例

### 4.1. 16:9プレゼンテーションの作成

```go
doc := gopdf.New()
page := doc.NewPage(gopdf.PageSizePresentation16x9, gopdf.Portrait)

// プレゼンテーションコンテンツの描画
page.SetFont(gopdf.FontHelvetica, 48)
page.DrawText("タイトル", 50, 300)
```

### 4.2. 4:3プレゼンテーションの作成

```go
doc := gopdf.New()
page := doc.NewPage(gopdf.PageSizePresentation4x3, gopdf.Portrait)

// プレゼンテーションコンテンツの描画
page.SetFont(gopdf.FontHelvetica, 48)
page.DrawText("タイトル", 50, 400)
```

## 5. テスト計画

### 5.1. ユニットテスト

- `constants_test.go` にサイズ定義の検証テストを追加
- アスペクト比の計算テスト
- Orientation (Portrait/Landscape) との組み合わせテスト

### 5.2. 統合テスト

- 16:9プレゼンテーションPDFの生成
- 4:3プレゼンテーションPDFの生成
- PDF Readerでの表示確認

### 5.3. サンプルコード

`examples/16_presentation_sizes/` にサンプルを作成

## 6. 参考資料

### 6.1. PowerPoint標準サイズ

- **16:9 Widescreen**: 10" x 5.625" (25.4cm x 14.29cm)
- **4:3 Standard**: 10" x 7.5" (25.4cm x 19.05cm)

### 6.2. PDF単位

- 1 point = 1/72 inch
- 1 inch = 72 points
- 1 cm = 28.35 points

### 6.3. 計算

```
16:9 Widescreen:
  Width:  10 inch × 72 = 720 points
  Height: 5.625 inch × 72 = 405 points
  Ratio: 720 / 405 = 1.777... ≈ 16/9

4:3 Standard:
  Width:  10 inch × 72 = 720 points
  Height: 7.5 inch × 72 = 540 points
  Ratio: 720 / 540 = 1.333... = 4/3
```

## 7. マイルストーン

- [x] 設計書作成
- [ ] `constants.go` への実装
- [ ] テストコード作成
- [ ] サンプルコード作成
- [ ] ドキュメント更新

## 8. 今後の拡張

### 8.1. その他のアスペクト比

将来的に以下のサイズも検討可能：
- **16:10**: 10" x 6.25" (720 x 450 points)
- **A4横**: 既存のPageSizeA4をLandscapeで使用

### 8.2. Markdown to Slide機能との連携

Markdown to Slide変換機能（Phase 14）で、これらのプレゼンテーションサイズを
デフォルトサイズとして使用する。
