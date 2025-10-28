# PDF Translation Example

このサンプルは、PDFのレイアウトを保持したまま、テキストを翻訳する機能を実演します。

## 機能

1. **レイアウト解析**: PDFからテキストと画像の位置情報を抽出
2. **テキスト翻訳**: ユーザー定義の翻訳関数でテキストを変換
3. **レイアウト保持**: 元のレイアウトを維持したまま新しいPDFを生成
4. **自動フィッティング**: 翻訳後のテキストを矩形領域内に自動調整

## 実行方法

### 英語のみの場合（標準フォント）

```bash
cd examples/10_pdf_translation
go run main.go
```

この場合、レイアウトは保持されますが、テキストは英語のままです。

### 日本語翻訳の場合（TTFフォント必要）

日本語を表示するには、日本語対応のTTFフォントが必要です。

#### 1. Noto Sans JPフォントをダウンロード

```bash
# Google Fontsから直接ダウンロード
# https://fonts.google.com/noto/specimen/Noto+Sans+JP

# またはwgetを使用（Linuxの場合）
wget https://github.com/google/fonts/raw/main/ofl/notosansjp/NotoSansJP%5Bwght%5D.ttf -O NotoSansJP-Regular.ttf
```

または手動でダウンロード：
1. https://fonts.google.com/noto/specimen/Noto+Sans+JP にアクセス
2. "Download family" をクリック
3. ZIPファイルを解凍
4. `NotoSansJP-Regular.ttf` を `examples/10_pdf_translation/` ディレクトリにコピー

#### 2. サンプルを実行

```bash
cd examples/10_pdf_translation
go run main.go
```

`NotoSansJP-Regular.ttf` が存在する場合、自動的に日本語翻訳が実行されます。

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

### 他のTTFフォント

```go
// 他のフォントを使用する場合
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

### 文字化けする

→ TTFフォントが読み込まれていません。`NotoSansJP-Regular.ttf` が正しい場所にあるか確認してください。

### フォントが見つからない

```
Warning: TTF font not found. Using Helvetica (English only)
```

→ `NotoSansJP-Regular.ttf` をダウンロードして、`examples/10_pdf_translation/` ディレクトリに配置してください。

### テキストが切れる

→ `FittingOptions` の `MinFontSize` を小さくするか、`AllowShrink` を `true` にしてください。

## 参考

- 設計書: [docs/pdf_translation_design.md](../../docs/pdf_translation_design.md)
- Noto Sans JP: https://fonts.google.com/noto/specimen/Noto+Sans+JP
