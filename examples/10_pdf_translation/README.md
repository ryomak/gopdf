# PDF Translation Example

このサンプルは、PDFのレイアウトを保持したまま、テキストを翻訳する機能を実演します。

## 機能

1. **レイアウト解析**: PDFからテキストと画像の位置情報を抽出
2. **テキスト翻訳**: ユーザー定義の翻訳関数でテキストを変換
3. **レイアウト保持**: 元のレイアウトを維持したまま新しいPDFを生成
4. **自動フィッティング**: 翻訳後のテキストを矩形領域内に自動調整

## 実行方法

### 日本語翻訳（デフォルト）

```bash
cd examples/10_pdf_translation
go run main.go
```

**gopdf v2では日本語フォント（Koruri）が埋め込まれているため、フォントファイルのダウンロードは不要です。**

実行すると、自動的に日本語翻訳されたPDFが生成されます。

## 出力ファイル

- `english.pdf`: 元の英語PDF（自動生成）
- `japanese.pdf`: 翻訳後のPDF

## コード概要

### 1. レイアウト解析

```go
reader, _ := gopdf.Open("english.pdf")
layout, _ := reader.ExtractPageLayout(0)

// layout.TextBlocks - テキストブロックのリスト
// layout.Images - 画像のリスト（位置情報付き）
```

### 2. 翻訳インターフェース

```go
translator := gopdf.TranslateFunc(func(text string) (string, error) {
    // 実際のアプリケーションでは翻訳APIを使用
    return translateAPI(text), nil
})
```

### 3. PDF生成

```go
opts := gopdf.PDFTranslatorOptions{
    Translator:     translator,
    TargetFont:     jpFont,         // TTFFont
    TargetFontName: "NotoSansJP",
    FittingOptions: gopdf.DefaultFitTextOptions(),
    KeepImages:     true,
    KeepLayout:     true,
}

gopdf.TranslatePDF("english.pdf", "japanese.pdf", opts)
```

## カスタマイズ

### 翻訳APIの統合

実際のアプリケーションでは、Google Translate APIなどを使用します：

```go
import "cloud.google.com/go/translate"

translator := gopdf.TranslateFunc(func(text string) (string, error) {
    ctx := context.Background()
    client, _ := translate.NewClient(ctx)
    defer client.Close()

    translations, _ := client.Translate(ctx, []string{text},
        language.Japanese, nil)

    return translations[0].Text, nil
})
```

### フィッティングオプション

```go
opts.FittingOptions = gopdf.FitTextOptions{
    MaxFontSize: 24.0,      // 最大フォントサイズ
    MinFontSize: 6.0,       // 最小フォントサイズ
    LineSpacing: 1.2,       // 行間（フォントサイズの倍率）
    Padding:     2.0,       // パディング
    AllowShrink: true,      // 縮小を許可
    AllowGrow:   false,     // 拡大を許可
    Alignment:   gopdf.AlignLeft, // 左寄せ/中央/右寄せ
}
```

### デフォルトフォントの使用

```go
// gopdf v2では埋め込みフォント（Koruri）を使用
jpFont, err := font.DefaultJapaneseFont()
if err != nil {
    log.Fatal(err)
}

opts := gopdf.PDFTranslatorOptions{
    TargetFont:     jpFont,
    TargetFontName: "Koruri",
    // ...
}
```

### カスタムTTFフォントの使用

独自のフォントを使用したい場合：

```go
// カスタムフォントを読み込み
jpFont, err := gopdf.LoadTTF("path/to/your/font.ttf")
if err != nil {
    log.Fatal(err)
}

opts := gopdf.PDFTranslatorOptions{
    TargetFont:     jpFont,
    TargetFontName: "YourFontName",
    // ...
}
```

## 制限事項

- 複雑な表構造の完全再現は未対応
- 回転・斜体などの複雑な変形は部分対応
- アノテーション（リンク、コメント）は保持されません
- 実際の翻訳処理は使用側で実装が必要

## トラブルシューティング

### 日本語が表示されない

gopdf v2では日本語フォント（Koruri）が自動的に埋め込まれているため、通常この問題は発生しません。

もし問題が発生した場合：
```bash
# gopdfを最新版に更新
go get -u github.com/ryomak/gopdf
go mod tidy
```

### テキストが切れる

→ `FittingOptions` の `MinFontSize` を小さくするか、`AllowShrink` を `true` にしてください。

### カスタムフォントを使いたい

→ 「カスタムTTFフォントの使用」セクションを参照してください。

## 参考

- 設計書: [docs/pdf_translation_design.md](../../docs/pdf_translation_design.md)
- デフォルトフォント: Koruri (https://koruri.github.io/)
- Koruriライセンス: Apache License 2.0
