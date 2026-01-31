# DrawText/DrawTextUTF8 共通化設計書

## 目的
`DrawText`と`DrawTextUTF8`の重複コードを共通化し、保守性とテスタビリティを向上させる。

## 現状の問題点

### コードの重複
`page.go`内の`DrawText`（50-71行目）と`DrawTextUTF8`（300-318行目）は、ほぼ同じ処理フローを持つ：

**共通処理:**
1. フォントが設定されているかチェック
2. フォントキーを取得
3. `BT` (Begin Text) を出力
4. フォントとサイズを設定 (`/%s %.2f Tf`)
5. 位置を設定 (`%.2f %.2f Td`)
6. テキストを描画 (`Tj`)
7. `ET` (End Text) を出力

**差異:**
| 項目 | DrawText | DrawTextUTF8 |
|------|----------|--------------|
| フォント型 | `*font.StandardFont` | `*TTFFont` |
| エンコーディング | `escapeString(text)` | `textToHexString(text)` |
| Tj記法 | `(%s) Tj` (括弧) | `<%s> Tj` (山括弧) |

### 保守性の問題
- 一方を修正しても、もう一方に反映されないリスク
- ロジックの追加・変更が2箇所に必要
- テストも2倍必要

## 提案: 共通化による改善

### アプローチ1: 内部ヘルパー関数（推奨）

共通ロジックを抽出し、差分をパラメータ化する。

```go
// drawTextInternal は DrawText と DrawTextUTF8 の共通ロジック
func (p *Page) drawTextInternal(
    text string,
    x, y float64,
    fontKey string,
    encodedText string,
    useBrackets bool, // true: (), false: <>
) {
    fmt.Fprintf(&p.content, "BT\n")
    fmt.Fprintf(&p.content, "/%s %.2f Tf\n", fontKey, p.fontSize)
    fmt.Fprintf(&p.content, "%.2f %.2f Td\n", x, y)

    if useBrackets {
        fmt.Fprintf(&p.content, "(%s) Tj\n", encodedText)
    } else {
        fmt.Fprintf(&p.content, "<%s> Tj\n", encodedText)
    }

    fmt.Fprintf(&p.content, "ET\n")
}
```

**使用例:**
```go
func (p *Page) DrawText(text string, x, y float64) error {
    if p.currentFont == nil {
        return fmt.Errorf("no font set; call SetFont before DrawText")
    }

    fontKey := p.getFontKey(*p.currentFont)
    encodedText := p.escapeString(text)
    p.drawTextInternal(text, x, y, fontKey, encodedText, true)

    return nil
}

func (p *Page) DrawTextUTF8(text string, x, y float64) error {
    if p.currentTTFFont == nil {
        return fmt.Errorf("no TTF font set; call SetTTFFont before DrawTextUTF8")
    }

    fontKey := p.getTTFFontKey(p.currentTTFFont)
    encodedText := p.textToHexString(text)
    p.drawTextInternal(text, x, y, fontKey, encodedText, false)

    return nil
}
```

### アプローチ2: Genericsを使用（検討）

Go 1.18+のGenericsを活用する案。

```go
type TextEncoder interface {
    EncodeText(text string) string
    UsesBrackets() bool
}

func (p *Page) drawText[E TextEncoder](
    text string,
    x, y float64,
    fontKey string,
    encoder E,
) {
    encodedText := encoder.EncodeText(text)
    useBrackets := encoder.UsesBrackets()

    p.drawTextInternal(text, x, y, fontKey, encodedText, useBrackets)
}
```

**評価**:
- より柔軟だが、現状の要件に対してはオーバーエンジニアリング
- 将来的にエンコーディング方式が増える場合は有用

### アプローチ3: 戦略パターン

エンコーディング戦略を分離する。

```go
type textEncodingStrategy interface {
    encodeText(text string) string
    format() string // "()" or "<>"
}

type standardFontStrategy struct{ page *Page }
func (s *standardFontStrategy) encodeText(text string) string {
    return s.page.escapeString(text)
}
func (s *standardFontStrategy) format() string { return "(%s)" }

type ttfFontStrategy struct{ page *Page }
func (s *ttfFontStrategy) encodeText(text string) string {
    return s.page.textToHexString(text)
}
func (s *ttfFontStrategy) format() string { return "<%s>" }
```

**評価**:
- 拡張性が高いが、現状の2パターンのみでは複雑すぎる

## 採用案: アプローチ1（内部ヘルパー関数）

### 理由
1. **シンプルさ**: 最小限の変更で重複を削減
2. **可読性**: 既存のコードとの一貫性が保たれる
3. **テスタビリティ**: 内部関数のテストが容易
4. **段階的移行**: 既存のAPIを変更せずにリファクタリング可能

### 実装詳細

#### 1. `drawTextInternal` 関数の追加

```go
// drawTextInternal は DrawText と DrawTextUTF8 の共通ロジック
// このメソッドは内部実装用であり、外部から直接呼び出すべきではない
func (p *Page) drawTextInternal(
    text string,
    x, y float64,
    fontKey string,
    encodedText string,
    useBrackets bool,
) {
    fmt.Fprintf(&p.content, "BT\n")
    fmt.Fprintf(&p.content, "/%s %.2f Tf\n", fontKey, p.fontSize)
    fmt.Fprintf(&p.content, "%.2f %.2f Td\n", x, y)

    if useBrackets {
        fmt.Fprintf(&p.content, "(%s) Tj\n", encodedText)
    } else {
        fmt.Fprintf(&p.content, "<%s> Tj\n", encodedText)
    }

    fmt.Fprintf(&p.content, "ET\n")
}
```

#### 2. `DrawText` のリファクタリング

```go
// DrawText draws text at the specified position.
// The position (x, y) is in PDF units (points), where (0, 0) is the bottom-left corner.
func (p *Page) DrawText(text string, x, y float64) error {
    if p.currentFont == nil {
        return fmt.Errorf("no font set; call SetFont before DrawText")
    }

    fontKey := p.getFontKey(*p.currentFont)
    encodedText := p.escapeString(text)
    p.drawTextInternal(text, x, y, fontKey, encodedText, true)

    return nil
}
```

#### 3. `DrawTextUTF8` のリファクタリング

```go
// DrawTextUTF8 draws UTF-8 encoded text at the specified position using the current TTF font.
// This method supports Unicode characters including Japanese, Chinese, Korean, etc.
func (p *Page) DrawTextUTF8(text string, x, y float64) error {
    if p.currentTTFFont == nil {
        return fmt.Errorf("no TTF font set; call SetTTFFont before DrawTextUTF8")
    }

    fontKey := p.getTTFFontKey(p.currentTTFFont)
    encodedText := p.textToHexString(text)
    p.drawTextInternal(text, x, y, fontKey, encodedText, false)

    return nil
}
```

## メリット

### コード品質
- **重複削減**: 共通ロジックが1箇所に集約
- **一貫性**: 両メソッドが同じロジックを使用することが保証される
- **可読性**: 各メソッドの役割（バリデーション→エンコーディング→描画）が明確

### 保守性
- **単一責任**: 描画ロジックは`drawTextInternal`のみが担当
- **変更容易性**: PDF生成ロジックの変更が1箇所で済む
- **バグ修正**: 共通バグの修正が両メソッドに自動的に反映

### テスタビリティ
- **ユニットテスト**: `drawTextInternal`を直接テスト可能
- **テーブルドリブン**: エンコーディングパターンのテストが簡潔に

## テスト戦略

### 既存テストの維持
- `TestPageDrawText` (`text_test.go:28`)
- `TestPage_DrawTextUTF8` (`ttf_font_test.go:156`)

**重要**: これらのテストは変更なしでパスする必要がある（外部APIは不変）

### 新規テストの追加

#### 1. `drawTextInternal`のユニットテスト

```go
func TestPage_drawTextInternal(t *testing.T) {
    tests := []struct {
        name          string
        text          string
        x, y          float64
        fontKey       string
        encodedText   string
        useBrackets   bool
        expectedRegex string
    }{
        {
            name:          "standard font with brackets",
            text:          "Hello",
            x:             100.0,
            y:             200.0,
            fontKey:       "F1",
            encodedText:   "Hello",
            useBrackets:   true,
            expectedRegex: `BT\n/F1 \d+\.\d+ Tf\n100\.00 200\.00 Td\n\(Hello\) Tj\nET\n`,
        },
        {
            name:          "TTF font with angle brackets",
            text:          "こんにちは",
            x:             50.0,
            y:             300.0,
            fontKey:       "F15",
            encodedText:   "3053308230930306306F",
            useBrackets:   false,
            expectedRegex: `BT\n/F15 \d+\.\d+ Tf\n50\.00 300\.00 Td\n<3053308230930306306F> Tj\nET\n`,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            doc := NewDocument()
            page := doc.AddPage(PageSizeA4)
            page.fontSize = 12.0

            page.drawTextInternal(tt.text, tt.x, tt.y, tt.fontKey, tt.encodedText, tt.useBrackets)

            content := page.content.String()
            if !regexp.MustCompile(tt.expectedRegex).MatchString(content) {
                t.Errorf("content doesn't match expected pattern\ngot: %s\nexpected pattern: %s",
                    content, tt.expectedRegex)
            }
        })
    }
}
```

#### 2. エンコーディングのテスト

```go
func TestPage_textEncodings(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        method   string // "escape" or "hex"
    }{
        {
            name:     "escape special characters",
            input:    "Hello (World)",
            expected: "Hello \\(World\\)",
            method:   "escape",
        },
        {
            name:     "hex encoding for Japanese",
            input:    "こんにちは",
            expected: "3053308230930306306F",
            method:   "hex",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            page := &Page{}

            var result string
            if tt.method == "escape" {
                result = page.escapeString(tt.input)
            } else {
                result = page.textToHexString(tt.input)
            }

            if result != tt.expected {
                t.Errorf("encoding failed: got %s, want %s", result, tt.expected)
            }
        })
    }
}
```

### 統合テスト

既存のexampleコードが引き続き動作することを確認：
- `examples/01_hello_world`
- `examples/10_pdf_translation`（UTF-8テキスト使用）

## 実装順序（TDD）

1. ✅ 設計書作成（このファイル）
2. ⬜ テストの準備
   - 既存テストが全てパスすることを確認
   - 新規テストケースを作成（Red: まだ実装していないのでFail）
3. ⬜ `drawTextInternal`関数の実装（Green: テストをパス）
4. ⬜ `DrawText`のリファクタリング
5. ⬜ `DrawTextUTF8`のリファクタリング
6. ⬜ 全テストの実行と確認
7. ⬜ コードレビューとコミット

## リスクと対策

### リスク1: 既存の動作が壊れる
**対策**:
- 既存テストを変更せず、全てパスすることを確認
- 統合テスト（実際のPDF生成）で動作確認

### リスク2: パフォーマンスの劣化
**対策**:
- 関数呼び出しのオーバーヘッドは無視できるレベル
- 必要に応じてベンチマークテストを追加

### リスク3: APIの一貫性
**対策**:
- 外部APIは一切変更しない
- 内部実装のみをリファクタリング

## 将来の拡張性

### 1. テキスト装飾のサポート
共通化により、以下の機能追加が容易に：
- テキストカラー設定
- アンダーライン
- 太字/斜体（フォント切り替え）

```go
type TextStyle struct {
    Color     Color
    Underline bool
    // ...
}

func (p *Page) drawTextInternal(
    text string,
    x, y float64,
    fontKey string,
    encodedText string,
    useBrackets bool,
    style *TextStyle, // 追加
) {
    // スタイル適用ロジック
}
```

### 2. テキストレイアウト機能
- テキストの配置（左揃え、中央揃え、右揃え）
- 自動改行
- テキストボックス

### 3. 複数行テキスト描画
`drawTextInternal`を基盤として、複数行描画を実装可能：

```go
func (p *Page) DrawMultilineText(lines []string, x, y, lineHeight float64) error {
    currentY := y
    for _, line := range lines {
        // DrawText or DrawTextUTF8 を呼び出し
        currentY -= lineHeight
    }
    return nil
}
```

## 参考

- `generics_design.md`: Generics活用の全体方針
- `graphics_design.md`: 図形描画の設計パターン
- PDF 1.7仕様書 Section 9.4: Text Objects

## 結論

`drawTextInternal`を導入することで：
1. ✅ コードの重複を削減
2. ✅ 保守性を向上
3. ✅ テスタビリティを改善
4. ✅ 将来の拡張に備える

既存のAPIを変更せず、内部実装のみをリファクタリングするため、**リスクが低く、効果が高い**改善です。
