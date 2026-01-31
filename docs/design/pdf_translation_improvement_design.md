# PDF翻訳機能改善 設計書

## 1. 概要

PDFを読み込んで翻訳し、なるべく同じ出力をするための改善を実施する。

### 1.1. 現状の課題

1. **フォントスタイルの喪失**: 太字/斜体の情報が失われる
2. **画像座標の異常値**: 一部PDFで座標が異常値になる
3. **全テキストが同一フォント**: ASCII/非ASCII問わず同じフォントで描画

### 1.2. 改善方針

1. フォントスタイル情報（Bold/Italic）を抽出・保持して再生成
2. 画像座標の異常値検出とフォールバック処理
3. ASCII文字には標準フォント、非ASCII文字にはTTFフォントを使用

## 2. 設計

### 2.1. TextBlockにフォントスタイル情報を追加

```go
// TextBlock はテキストの論理的なブロック
type TextBlock struct {
    Text      string
    Elements  []TextElement
    Rect      Rectangle
    Font      string
    FontSize  float64
    Color     Color
    IsBold    bool    // 太字フラグ（新規追加）
    IsItalic  bool    // 斜体フラグ（新規追加）
}
```

### 2.2. フォントスタイルの判定

元のフォント名から太字/斜体を判定：

```go
func detectFontStyle(fontName string) (isBold, isItalic bool) {
    upper := strings.ToUpper(fontName)
    isBold = strings.Contains(upper, "BOLD") ||
             strings.Contains(upper, "-B") ||
             strings.HasSuffix(upper, "BD")
    isItalic = strings.Contains(upper, "ITALIC") ||
               strings.Contains(upper, "OBLIQUE") ||
               strings.Contains(upper, "-I") ||
               strings.Contains(upper, "-O")
    return
}
```

### 2.3. RenderLayoutの改善

文字種別に応じてフォントを選択：

```go
func renderTextWithSmartFont(page *Page, text string, x, y float64, fontSize float64, isBold bool, japaneseFont *TTFFont) error {
    // 文字単位でフォントを切り替える
    for _, char := range text {
        if char < 128 {
            // ASCII文字: 標準フォント
            if isBold {
                page.SetFont(FontHelveticaBold, fontSize)
            } else {
                page.SetFont(FontHelvetica, fontSize)
            }
        } else {
            // 非ASCII文字: 日本語フォント
            page.SetTTFFont(japaneseFont, fontSize)
        }
        page.DrawText(string(char), x, y)
        x += estimateCharWidth(char, fontSize)
    }
    return nil
}
```

### 2.4. 画像座標の異常値対策

RenderLayoutで座標が異常な場合のフォールバック：

```go
func validateImagePosition(x, y, width, height, pageWidth, pageHeight float64) (newX, newY float64, isValid bool) {
    const maxOffset = 10000.0

    // 異常値の検出
    if x < -maxOffset || x > pageWidth+maxOffset ||
       y < -maxOffset || y > pageHeight+maxOffset {
        // フォールバック: ページ内の適切な位置に配置
        return 50.0, pageHeight - height - 50.0, false
    }

    return x, y, true
}
```

## 3. 実装手順

### Phase 1: TextBlockのフォントスタイル情報追加
1. layout/blocks.go に IsBold, IsItalic フィールドを追加
2. layout.go の createTextBlockFromLines でフォントスタイルを判定・設定

### Phase 2: RenderLayoutの改善
1. 文字種別に応じたフォント選択ロジックを実装
2. 太字の場合は FontHelveticaBold を使用

### Phase 3: 画像座標の異常値対策
1. RenderLayoutで画像描画前に座標検証
2. 異常な場合はフォールバック位置に配置

## 4. テスト計画

1. 太字テキストを含むPDFの翻訳テスト
2. 日本語+英語混在テキストの翻訳テスト
3. 異常座標の画像を含むPDFの翻訳テスト
