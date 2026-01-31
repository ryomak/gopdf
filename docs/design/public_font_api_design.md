# フォント公開API設計書

## 1. 背景と目的

### 1.1. 問題
現在の実装では、`internal/font`パッケージにある`StandardFont`型を直接使用しているため、外部ユーザーがこの型を利用できない。

```go
// 現在の実装（外部から使えない）
import "github.com/ryomak/gopdf/internal/font"

page.SetFont(font.Helvetica, 24)  // エラー：internalパッケージにアクセスできない
```

### 1.2. 目的
- `internal/font`パッケージの`StandardFont`定数を公開APIとして提供
- 外部ユーザーが簡単にフォントを指定できるインターフェースを提供
- 既存の内部実装を変更せず、公開レイヤーのみ追加

## 2. 設計方針

### 2.1. レイヤー分離
- **内部実装**: `internal/font` - 実装詳細、PDF仕様に基づく処理
- **公開API**: `gopdf` - ユーザー向けのシンプルなインターフェース

### 2.2. 型定義
`gopdf`パッケージに`StandardFont`型を定義し、内部実装との橋渡しを行う。

```go
// gopdf/font.go
package gopdf

// StandardFont represents one of the 14 standard PDF fonts
type StandardFont string

// 14種類の標準フォント定数
const (
    Helvetica           StandardFont = "Helvetica"
    HelveticaBold       StandardFont = "Helvetica-Bold"
    // ... 他のフォント
)
```

### 2.3. 内部変換
公開APIの`StandardFont`を内部実装の`internal/font.StandardFont`に変換する。

## 3. API設計

### 3.1. 公開フォント定数

```go
package gopdf

// StandardFont represents one of the 14 standard PDF fonts.
// These fonts are built into PDF viewers and don't need to be embedded.
type StandardFont string

// Sans-serif fonts
const (
    Helvetica           StandardFont = "Helvetica"
    HelveticaBold       StandardFont = "Helvetica-Bold"
    HelveticaOblique    StandardFont = "Helvetica-Oblique"
    HelveticaBoldOblique StandardFont = "Helvetica-BoldOblique"
)

// Serif fonts
const (
    TimesRoman      StandardFont = "Times-Roman"
    TimesBold       StandardFont = "Times-Bold"
    TimesItalic     StandardFont = "Times-Italic"
    TimesBoldItalic StandardFont = "Times-BoldItalic"
)

// Monospace fonts
const (
    Courier           StandardFont = "Courier"
    CourierBold       StandardFont = "Courier-Bold"
    CourierOblique    StandardFont = "Courier-Oblique"
    CourierBoldOblique StandardFont = "Courier-BoldOblique"
)

// Symbol fonts
const (
    Symbol       StandardFont = "Symbol"
    ZapfDingbats StandardFont = "ZapfDingbats"
)
```

### 3.2. Page.SetFont メソッド

```go
// SetFont sets the current font and size for subsequent text operations.
func (p *Page) SetFont(f StandardFont, size float64) error {
    // 公開APIの型を内部実装の型に変換
    internalFont := font.StandardFont(f)

    p.currentFont = &internalFont
    p.fontSize = size

    // フォントをページのフォントリストに追加
    if p.fonts == nil {
        p.fonts = make(map[string]font.StandardFont)
    }
    fontKey := p.getFontKey(internalFont)
    p.fonts[fontKey] = internalFont

    return nil
}
```

### 3.3. 使用例

```go
package main

import (
    "os"
    "github.com/ryomak/gopdf"
)

func main() {
    // 新規PDFドキュメントを作成
    doc := gopdf.New()

    // A4サイズの縦向きページを追加
    page := doc.AddPage(gopdf.A4, gopdf.Portrait)

    // フォントを設定してテキストを描画
    page.SetFont(gopdf.Helvetica, 24)
    page.DrawText("Hello, World!", 100, 750)

    page.SetFont(gopdf.TimesRoman, 14)
    page.DrawText("gopdf - Pure Go PDF library", 100, 720)

    // ファイルに出力
    file, _ := os.Create("output.pdf")
    defer file.Close()

    doc.WriteTo(file)
}
```

## 4. 実装手順

### 4.1. font.go の作成
1. `gopdf/font.go` ファイルを作成
2. `StandardFont` 型と14種類の定数を定義
3. ヘルパー関数（必要に応じて）を実装

### 4.2. page.go の修正
1. `SetFont` メソッドの引数型を `gopdf.StandardFont` に変更
2. 内部で `internal/font.StandardFont` に変換
3. 他のフォント関連メソッドも同様に修正

### 4.3. テストの追加
1. `font_test.go` を作成
2. 公開APIの動作確認テストを追加
3. 既存のテストを更新（型変更に対応）

### 4.4. ドキュメント更新
1. README.md のサンプルコードを更新
2. GoDocコメントを追加

## 5. 下位互換性

### 5.1. 破壊的変更
`Page.SetFont` の引数型が変更されるため、既存のコードは影響を受ける。ただし、定数名が同じため、importを変更するだけで動作する。

```go
// 変更前
import "github.com/ryomak/gopdf/internal/font"
page.SetFont(font.Helvetica, 24)

// 変更後
import "github.com/ryomak/gopdf"
page.SetFont(gopdf.Helvetica, 24)
```

### 5.2. 移行ガイド
1. `internal/font` のimportを削除
2. `font.` プレフィックスを `gopdf.` に変更
3. または、定数を直接使用（`gopdf.Helvetica` → `Helvetica`）

## 6. 将来の拡張

### 6.1. TTFフォントのサポート
TTFフォントも同様に公開APIとして提供する。

```go
// 将来的な実装例
type TTFFont struct {
    internal *font.TTFFont
}

func LoadTTFFont(path string) (*TTFFont, error) {
    // ...
}

func (p *Page) SetTTFFont(f *TTFFont, size float64) error {
    // ...
}
```

### 6.2. フォント情報の取得
フォントのメトリクス情報を取得するAPIを追加。

```go
func (f StandardFont) Metrics() FontMetrics {
    // ...
}
```

## 7. テスト戦略

### 7.1. ユニットテスト
- 各フォント定数が正しい値を持つことを確認
- `SetFont` メソッドの動作確認
- 内部変換が正しく行われることを確認

### 7.2. 統合テスト
- 実際のPDF生成で各フォントが正しく使用されることを確認
- PDF viewerでの表示確認

### 7.3. テストコード例

```go
func TestStandardFontConstants(t *testing.T) {
    tests := []struct {
        name string
        font StandardFont
        want string
    }{
        {"Helvetica", Helvetica, "Helvetica"},
        {"Times-Roman", TimesRoman, "Times-Roman"},
        // ...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if string(tt.font) != tt.want {
                t.Errorf("got %s, want %s", tt.font, tt.want)
            }
        })
    }
}
```

## 8. まとめ

この設計により：
- ✅ 外部ユーザーが標準フォントを簡単に使用できる
- ✅ 内部実装の詳細を隠蔽できる
- ✅ 将来的な拡張に対応できる柔軟な設計
- ✅ シンプルで直感的なAPI
