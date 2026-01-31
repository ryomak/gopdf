# テキストとフォント

テキストの描画とフォントの使用方法を説明します。

## 標準フォント

PDFには14種類の標準フォントが組み込まれており、フォントファイルを埋め込まずに使用できます。

### 標準フォント一覧

| ファミリー | Regular | Bold | Oblique/Italic | Bold Oblique/Italic |
|-----------|---------|------|----------------|---------------------|
| Helvetica | `FontHelvetica` | `FontHelveticaBold` | `FontHelveticaOblique` | `FontHelveticaBoldOblique` |
| Times | `FontTimesRoman` | `FontTimesBold` | `FontTimesItalic` | `FontTimesBoldItalic` |
| Courier | `FontCourier` | `FontCourierBold` | `FontCourierOblique` | `FontCourierBoldOblique` |
| Symbol | `FontSymbol` | - | - | - |
| ZapfDingbats | `FontZapfDingbats` | - | - | - |

### 使用例

```go
page := doc.AddPage(gopdf.PageSizeA4, gopdf.Portrait)

// Helvetica（サンセリフ体）
page.SetFont(gopdf.FontHelvetica, 24)
page.DrawText("Hello, World!", 100, 750)

// Helvetica Bold
page.SetFont(gopdf.FontHelveticaBold, 24)
page.DrawText("Bold Text", 100, 720)

// Times Roman（セリフ体）
page.SetFont(gopdf.FontTimesRoman, 18)
page.DrawText("Times Roman", 100, 690)

// Courier（等幅フォント）
page.SetFont(gopdf.FontCourier, 14)
page.DrawText("Monospace Text", 100, 660)
```

## TrueTypeフォント (TTF)

日本語などのマルチバイト文字を描画するには、TTFフォントを使用します。

### フォントの読み込み

```go
// ファイルパスから読み込み
font, err := gopdf.LoadTTF("/path/to/font.ttf")
if err != nil {
    log.Fatal(err)
}

// バイト配列から読み込み
fontData, _ := os.ReadFile("/path/to/font.ttf")
font, err := gopdf.LoadTTFFromBytes(fontData)
```

### 日本語システムフォントの自動検出

`LoadSystemJapaneseFont()`は、OSに応じて日本語フォントを自動的に検出します：

- **macOS**: Hiragino Kaku Gothic
- **Linux**: Noto Sans CJK JP
- **Windows**: Yu Gothic

```go
jpFont, err := gopdf.LoadSystemJapaneseFont()
if err != nil {
    log.Fatal("Japanese font not found on system")
}

page.SetTTFFont(jpFont, 24)
page.DrawText("こんにちは、世界！", 100, 750)
```

### TTFフォントの使用例

```go
// フォントを読み込み
ttfFont, err := gopdf.LoadTTF("/path/to/NotoSansJP-Regular.ttf")
if err != nil {
    log.Fatal(err)
}

page := doc.AddPage(gopdf.PageSizeA4, gopdf.Portrait)

// TTFフォントを設定
page.SetTTFFont(ttfFont, 18)

// 日本語テキストを描画
page.DrawText("日本語テキスト", 50, 750)
page.DrawText("English and 日本語", 50, 720)
```

### テキスト幅の計算

TTFフォントでは、テキストの幅を計算できます：

```go
ttfFont, _ := gopdf.LoadTTF("/path/to/font.ttf")

// テキスト幅を計算（ポイント単位）
width, err := ttfFont.TextWidth("Hello, World!", 18)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Text width: %.2f pt\n", width)
```

## テキストの描画

### 基本的な描画

```go
// フォントとサイズを設定
page.SetFont(gopdf.FontHelvetica, 24)

// 座標 (x, y) にテキストを描画
page.DrawText("Hello, World!", 100, 750)
```

### 複数行のテキスト

現在、自動改行機能はありません。複数行のテキストは手動で描画します：

```go
page.SetFont(gopdf.FontHelvetica, 14)

lines := []string{
    "Line 1: First line of text",
    "Line 2: Second line of text",
    "Line 3: Third line of text",
}

y := 750.0
lineHeight := 20.0

for _, line := range lines {
    page.DrawText(line, 100, y)
    y -= lineHeight
}
```

### テキストの中央揃え

TTFフォントを使用して、テキストを中央揃えにする例：

```go
ttfFont, _ := gopdf.LoadTTF("/path/to/font.ttf")
page.SetTTFFont(ttfFont, 24)

text := "Centered Text"
fontSize := 24.0
pageWidth := 595.0 // A4幅

// テキスト幅を計算
textWidth, _ := ttfFont.TextWidth(text, fontSize)

// 中央の X 座標を計算
x := (pageWidth - textWidth) / 2

page.DrawText(text, x, 750)
```

## サンプルコード

完全な例は[examples/02_hello_world](https://github.com/ryomak/gopdf/tree/main/examples/02_hello_world)と[examples/09_ttf_fonts](https://github.com/ryomak/gopdf/tree/main/examples/09_ttf_fonts)を参照してください。

## 次のステップ

- [図形描画](graphics.md) - 線、矩形、円の描画
- [高度な機能](advanced.md) - ルビ（ふりがな）の描画
