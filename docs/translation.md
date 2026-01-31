# PDF翻訳

既存PDFのレイアウトを保持しながら、テキストを翻訳する方法を説明します。

## 概要

gopdfのPDF翻訳機能は、以下を実現します：

- 元のレイアウト（テキストの位置、サイズ）を保持
- フォント情報を維持
- 画像を保持
- カスタム翻訳関数で任意の翻訳エンジンを使用可能

## 基本的な使い方

### TranslatePDF関数

```go
err := gopdf.TranslatePDF(inputPath, outputPath, options)
```

### 簡単な例

```go
package main

import (
    "log"
    "github.com/ryomak/gopdf"
)

func main() {
    // 日本語フォントを読み込み
    jpFont, err := gopdf.LoadSystemJapaneseFont()
    if err != nil {
        log.Fatal(err)
    }

    // 翻訳辞書
    dict := map[string]string{
        "Hello":  "こんにちは",
        "World":  "世界",
        "Sample": "サンプル",
    }

    // 翻訳関数を定義
    translator := gopdf.TranslateFunc(func(text string) (string, error) {
        if translated, ok := dict[text]; ok {
            return translated, nil
        }
        return text, nil // 翻訳がなければ元のテキストを返す
    })

    // 翻訳オプション
    opts := gopdf.PDFTranslatorOptions{
        Translator:    translator,
        TargetFont:    jpFont,
        TranslateUnit: gopdf.TranslateUnitSentence,
        KeepLayout:    true,
        KeepImages:    true,
    }

    // 翻訳実行
    err = gopdf.TranslatePDF("input.pdf", "output.pdf", opts)
    if err != nil {
        log.Fatal(err)
    }
}
```

## PDFTranslatorOptions

翻訳の動作を制御するオプション：

```go
type PDFTranslatorOptions struct {
    Translator     Translator    // 翻訳関数（必須）
    TargetFont     Font          // 翻訳後のフォント（非ASCII用）
    TargetFontName string        // フォント名（省略可、自動取得）
    TranslateUnit  TranslateUnit // 翻訳単位
    KeepLayout     bool          // レイアウト保持
    KeepImages     bool          // 画像保持
}
```

### オプションの詳細

| オプション | 説明 | デフォルト |
|-----------|------|-----------|
| `Translator` | 翻訳関数（必須） | - |
| `TargetFont` | 翻訳後テキスト用フォント | nil |
| `TargetFontName` | フォント名（省略時は自動取得） | "" |
| `TranslateUnit` | 翻訳単位 | TranslateUnitBlock |
| `KeepLayout` | レイアウトを保持 | false |
| `KeepImages` | 画像を保持 | false |

## 翻訳単位 (TranslateUnit)

テキストをどの単位で翻訳するかを指定します：

```go
// ブロック全体を1つの翻訳単位として扱う
gopdf.TranslateUnitBlock

// 行単位で翻訳
gopdf.TranslateUnitLine

// 文単位で翻訳（. 。 ! ? で区切り）
gopdf.TranslateUnitSentence
```

### 使い分け

| 翻訳単位 | 用途 |
|---------|------|
| `TranslateUnitBlock` | 段落全体を翻訳したい場合 |
| `TranslateUnitLine` | 行ごとに翻訳したい場合 |
| `TranslateUnitSentence` | 文単位で正確に翻訳したい場合（推奨） |

## カスタム翻訳関数

### 関数型アダプタ

```go
translator := gopdf.TranslateFunc(func(text string) (string, error) {
    // 翻訳処理
    return translatedText, nil
})
```

### 外部翻訳APIの使用例

```go
import (
    "encoding/json"
    "net/http"
    "bytes"
)

// 翻訳API呼び出し関数
func translateWithAPI(text string) (string, error) {
    payload := map[string]string{
        "text":   text,
        "source": "en",
        "target": "ja",
    }
    body, _ := json.Marshal(payload)

    resp, err := http.Post(
        "https://api.translation-service.com/translate",
        "application/json",
        bytes.NewBuffer(body),
    )
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    var result struct {
        TranslatedText string `json:"translated_text"`
    }
    json.NewDecoder(resp.Body).Decode(&result)

    return result.TranslatedText, nil
}

// 使用
translator := gopdf.TranslateFunc(translateWithAPI)
```

### キャッシュ付き翻訳関数

```go
type CachedTranslator struct {
    cache map[string]string
    api   func(string) (string, error)
}

func (t *CachedTranslator) Translate(text string) (string, error) {
    if cached, ok := t.cache[text]; ok {
        return cached, nil
    }

    translated, err := t.api(text)
    if err != nil {
        return "", err
    }

    t.cache[text] = translated
    return translated, nil
}

// 使用
cached := &CachedTranslator{
    cache: make(map[string]string),
    api:   translateWithAPI,
}
translator := gopdf.TranslateFunc(cached.Translate)
```

## レイアウト情報の活用

翻訳前にレイアウトを確認・分析できます：

```go
reader, _ := gopdf.Open("input.pdf")
layouts, _ := reader.ExtractAllLayouts()

for pageNum, layout := range layouts {
    fmt.Printf("Page %d:\n", pageNum+1)
    fmt.Printf("  Size: %.2f x %.2f\n", layout.Width, layout.Height)

    for _, block := range layout.TextBlocks {
        fmt.Printf("  Text: %s\n", block.Text)
        fmt.Printf("    Position: (%.2f, %.2f)\n", block.X, block.Y)
        fmt.Printf("    Font: %s, Size: %.2f\n", block.FontName, block.FontSize)
    }
}
```

## 完全な例

```go
package main

import (
    "log"
    "strings"
    "github.com/ryomak/gopdf"
)

func main() {
    // 日本語フォントを読み込み
    jpFont, err := gopdf.LoadSystemJapaneseFont()
    if err != nil {
        log.Fatal("Japanese font not found:", err)
    }

    // 翻訳辞書（実際のアプリではAPIを使用）
    translations := map[string]string{
        "Introduction":     "はじめに",
        "Chapter 1":        "第1章",
        "Summary":          "まとめ",
        "This is a sample": "これはサンプルです",
    }

    // 翻訳関数
    translator := gopdf.TranslateFunc(func(text string) (string, error) {
        // 完全一致
        if t, ok := translations[text]; ok {
            return t, nil
        }

        // 部分一致（簡易的な置換）
        result := text
        for en, ja := range translations {
            result = strings.ReplaceAll(result, en, ja)
        }

        return result, nil
    })

    // オプション設定
    opts := gopdf.PDFTranslatorOptions{
        Translator:    translator,
        TargetFont:    jpFont,
        TranslateUnit: gopdf.TranslateUnitSentence,
        KeepLayout:    true,
        KeepImages:    true,
    }

    // 翻訳実行
    err = gopdf.TranslatePDF("english_document.pdf", "japanese_document.pdf", opts)
    if err != nil {
        log.Fatal("Translation failed:", err)
    }

    log.Println("Translation completed successfully!")
}
```

## 注意事項

- 翻訳後のテキストが元の領域に収まらない場合、フォントサイズが自動調整されることがあります
- 複雑なレイアウト（表、グラフ内のテキストなど）は正確に翻訳されない場合があります
- 画像内のテキストは翻訳されません

## サンプルコード

完全な例は[examples/10_pdf_translation](https://github.com/ryomak/gopdf/tree/main/examples/10_pdf_translation)を参照してください。

## 次のステップ

- [Markdown変換](markdown.md) - MarkdownからPDFへの変換
- [高度な機能](advanced.md) - レイアウト調整機能
