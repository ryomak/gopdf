# デフォルト日本語フォント埋め込み機能 設計書

## 目的

gopdfライブラリで日本語テキストを使用する際、ユーザーが手動でTTFフォントを配置する必要がなく、デフォルトで日本語が使用できる機能を提供する。

## 背景

現在の実装（Phase 9完了時点）では：
- TTFフォントを使用して日本語を描画可能
- ユーザーがNoto Sans JP等のフォントを手動でダウンロードして配置する必要がある
- 初心者にとってハードルが高い

## 要件

### 機能要件

1. **フォントの埋め込み**
   - オープンソースの日本語フォントをライブラリに埋め込む
   - `go:embed`を使用してビルド時にバイナリに含める
   - ライセンスを遵守し、ドキュメントに明記

2. **デフォルトフォントAPI**
   - `gopdf.DefaultJapaneseFont()`関数を提供
   - 埋め込みフォントを自動的に読み込み
   - キャッシュ機構により初回のみ展開

3. **自動フォールバック機能**
   - 文字種を自動判定（ASCII/日本語/その他）
   - 標準フォントと日本語フォントを自動切り替え
   - `Page.DrawTextAuto()`関数で自動フォント選択

4. **下位互換性**
   - 既存のAPI（`LoadTTF`, `SetTTFFont`）は維持
   - ユーザーが独自フォントを使用することも可能

### 非機能要件

1. **パフォーマンス**
   - フォントの展開はメモリ上で実行
   - キャッシュにより2回目以降は高速化
   - サブセットフォント使用によりバイナリサイズを抑制

2. **ライセンス遵守**
   - 選定フォントのライセンスを遵守
   - LICENSE/NOTICEファイルに明記
   - 商用利用可能なライセンス

3. **メンテナンス性**
   - フォントファイルは`internal/font/embedded/`に配置
   - バージョン管理に含める
   - フォント更新手順をドキュメント化

## フォント選定

### 候補フォント比較

| フォント | ライセンス | サイズ | メリット | デメリット |
|---------|----------|-------|---------|----------|
| **Noto Sans JP** | SIL OFL 1.1 | フル: 4-8MB<br>サブセット: 1-2MB | Google製、高品質<br>7ウェイト対応 | フルセットは大きい |
| **M+ FONTS** | OFL | 1-3MB | 軽量、Google Fonts対応<br>4600+グリフ | ウェイトが少ない |
| **IPAゴシック** | IPAフォントライセンス | 3-4MB | JIS規格準拠 | 再配布条件がやや複雑 |

### 推奨選択: Noto Sans JP（サブセット版）

**理由:**
1. **ライセンス**: SIL OFL 1.1で商用利用・再配布可能
2. **品質**: Google製で高品質、広く使用されている
3. **対応文字**: JIS第一水準+一部第二水準で日常使用に十分
4. **サイズ**: サブセット版で1-2MBに抑制可能
5. **メンテナンス**: Google Fontsで継続的にメンテナンスされている

**使用するフォント:**
- Noto Sans CJK JP（サブセット版）
- レギュラーウェイトのみを初期実装
- 将来的に複数ウェイト対応を検討

## アーキテクチャ設計

### ディレクトリ構造

```
gopdf/
├── internal/
│   └── font/
│       ├── embedded/
│       │   ├── embed.go              # go:embed定義
│       │   ├── NotoSansJP-Regular.ttf
│       │   └── LICENSE.txt           # フォントライセンス
│       ├── default.go                # デフォルトフォントAPI
│       └── fallback.go               # フォールバック機構
├── auto_text.go                      # 自動フォント切り替えAPI
└── docs/
    └── licenses/
        └── NotoSansJP-LICENSE.txt
```

### コンポーネント設計

#### 1. フォント埋め込み (`internal/font/embedded/embed.go`)

```go
package embedded

import _ "embed"

// NotoSansJPRegular は埋め込まれたNoto Sans JP Regularフォント
//
//go:embed NotoSansJP-Regular.ttf
var NotoSansJPRegular []byte

// License はフォントのライセンステキスト
//
//go:embed LICENSE.txt
var License string
```

#### 2. デフォルトフォントAPI (`internal/font/default.go`)

```go
package font

import (
    "sync"
    "github.com/ryomak/gopdf/internal/font/embedded"
)

var (
    defaultJPFont     *TTFFont
    defaultJPFontOnce sync.Once
    defaultJPFontErr  error
)

// DefaultJapaneseFont は埋め込まれた日本語フォントを返す
// 初回呼び出し時にフォントを読み込み、以降はキャッシュを返す
func DefaultJapaneseFont() (*TTFFont, error) {
    defaultJPFontOnce.Do(func() {
        defaultJPFont, defaultJPFontErr = loadEmbeddedFont(embedded.NotoSansJPRegular)
    })
    return defaultJPFont, defaultJPFontErr
}

func loadEmbeddedFont(data []byte) (*TTFFont, error) {
    // bytes.Readerを使用してメモリからフォントを読み込み
    // 既存のTTFFont実装を再利用
}
```

#### 3. フォールバック機構 (`internal/font/fallback.go`)

```go
package font

// CharType は文字の種類を表す
type CharType int

const (
    CharTypeASCII CharType = iota
    CharTypeJapanese
    CharTypeOther
)

// DetectCharType は文字の種類を判定する
func DetectCharType(r rune) CharType {
    switch {
    case r < 0x80:
        return CharTypeASCII
    case (r >= 0x3040 && r <= 0x309F) ||  // ひらがな
         (r >= 0x30A0 && r <= 0x30FF) ||  // カタカナ
         (r >= 0x4E00 && r <= 0x9FFF):    // 漢字（CJK統合漢字）
        return CharTypeJapanese
    default:
        return CharTypeOther
    }
}

// SplitByCharType はテキストを文字種で分割する
func SplitByCharType(text string) []TextSegment {
    // 文字種ごとにセグメント分割
}

type TextSegment struct {
    Text     string
    CharType CharType
}
```

#### 4. 自動テキスト描画API (`auto_text.go`)

```go
package gopdf

// DrawTextAuto はテキストを自動的にフォント切り替えして描画
// ASCII部分は標準フォント、日本語部分は埋め込み日本語フォントを使用
func (p *Page) DrawTextAuto(text string, x, y float64) error {
    segments := font.SplitByCharType(text)

    currentX := x
    for _, seg := range segments {
        switch seg.CharType {
        case font.CharTypeASCII:
            // 標準フォントで描画
            p.DrawText(seg.Text, currentX, y)
        case font.CharTypeJapanese:
            // 日本語フォントで描画
            jpFont, err := font.DefaultJapaneseFont()
            if err != nil {
                return err
            }
            p.SetTTFFont(jpFont, p.currentFontSize)
            p.DrawTextUTF8(seg.Text, currentX, y)
        }

        // 次のセグメントのX座標を計算
        width := calculateTextWidth(seg.Text)
        currentX += width
    }

    return nil
}

// SetDefaultJapaneseFont はページのデフォルト日本語フォントを設定
// これを呼ぶと以降のDrawTextは自動的に日本語対応になる
func (p *Page) SetDefaultJapaneseFont() error {
    jpFont, err := font.DefaultJapaneseFont()
    if err != nil {
        return err
    }
    p.defaultJPFont = jpFont
    return nil
}
```

## 実装計画

### Phase 1: フォント埋め込み基盤 ✓

1. **フォント取得とサブセット化**
   - Noto Sans JP Regularをダウンロード
   - 必要に応じてサブセット化（フルセットが小さければそのまま使用）
   - `internal/font/embedded/`に配置

2. **埋め込みAPI実装**
   - `embed.go`の実装
   - `default.go`の実装（`DefaultJapaneseFont`関数）
   - メモリからのフォント読み込み機能

3. **テスト作成**
   - 埋め込みフォントの読み込みテスト
   - キャッシュ機構のテスト
   - 複数ゴルーチンからの同時アクセステスト

### Phase 2: 自動フォールバック機能 ✓

1. **文字種判定機能**
   - `DetectCharType`関数実装
   - `SplitByCharType`関数実装
   - Unicode範囲のテスト

2. **自動描画API**
   - `DrawTextAuto`関数実装
   - 混在テキストの幅計算
   - 自動フォント切り替え

3. **統合テスト**
   - ASCII+日本語混在テキストのテスト
   - 複数行テキストのテスト
   - 各種文字種の組み合わせテスト

### Phase 3: サンプル・ドキュメント作成 ✓

1. **サンプルコード**
   - `examples/11_default_japanese_font/`
   - 基本的な日本語テキスト描画
   - 自動フォント切り替えのデモ

2. **ドキュメント更新**
   - READMEに使用例追加
   - ライセンス情報の明記
   - API仕様書の更新

## API使用例

### 基本的な使用方法

```go
package main

import (
    "github.com/ryomak/gopdf"
)

func main() {
    doc := gopdf.New()
    page := doc.AddPage(gopdf.A4, gopdf.Portrait)

    // デフォルト日本語フォントを使用
    jpFont, _ := gopdf.DefaultJapaneseFont()
    page.SetTTFFont(jpFont, 16)
    page.DrawTextUTF8("こんにちは、世界！", 50, 800)

    doc.WriteTo(file)
}
```

### 自動フォント切り替え（推奨）

```go
func main() {
    doc := gopdf.New()
    page := doc.AddPage(gopdf.A4, gopdf.Portrait)

    // 日本語フォントを有効化
    page.SetDefaultJapaneseFont()

    // 自動的にフォント切り替え
    page.DrawTextAuto("Hello, 世界！", 50, 800)
    page.DrawTextAuto("This is 日本語 mixed text.", 50, 770)

    doc.WriteTo(file)
}
```

## フォントのサブセット化

### サブセット化の方針

1. **含める文字範囲**
   - ASCII（0x20-0x7E）
   - ひらがな（0x3040-0x309F）
   - カタカナ（0x30A0-0x30FF）
   - JIS第一水準漢字（2,965字）
   - 一部のJIS第二水準漢字（常用漢字等）
   - 基本的な記号・句読点

2. **サブセット化ツール**
   - pyftsubsetを使用（fonttools）
   - スクリプトで自動化

3. **目標サイズ**
   - 2MB以下を目標
   - gzip圧縮でさらに縮小可能

### サブセット化コマンド例

```bash
# fonttoolsをインストール
pip install fonttools

# サブセット化
pyftsubset NotoSansJP-Regular.ttf \
  --unicodes="U+0020-007E,U+3040-309F,U+30A0-30FF,U+4E00-9FFF" \
  --output-file=NotoSansJP-Regular-subset.ttf \
  --layout-features='*' \
  --flavor=woff2
```

## ライセンス管理

### 埋め込みフォントのライセンス表示

1. **ライセンスファイルの配置**
   - `internal/font/embedded/LICENSE.txt`に配置
   - Noto Sans JPのOFLライセンスを含める

2. **ドキュメントへの記載**
   - `README.md`にライセンス情報を追加
   - `docs/licenses/`に詳細ライセンスを配置

3. **コード内のコメント**
   - 埋め込みフォント使用箇所にライセンス言及
   - 著作権表示を含める

### Noto Sans JP ライセンス要約

- **ライセンス**: SIL Open Font License 1.1
- **著作権**: Copyright 2014-2021 Adobe (http://www.adobe.com/)
- **許可事項**:
  - 商用利用可
  - 改変可
  - 再配布可
  - バンドル可
- **制限事項**:
  - フォント自体の販売は不可
  - ライセンス表示が必要

## パフォーマンス考慮事項

### メモリ使用量

- 埋め込みフォント: 1-2MB（バイナリに含まれる）
- 実行時展開: 初回のみメモリに展開
- キャッシュ: `sync.Once`で1回のみ読み込み

### 起動時間への影響

- フォントはgo:embedで埋め込まれ、実行時には既にメモリに存在
- 初回の`DefaultJapaneseFont()`呼び出し時のみパース処理
- 約10-50ms程度のオーバーヘッド（フォントサイズに依存）

### バイナリサイズへの影響

- +1-2MB（サブセット版）
- gzip圧縮で約50%縮小可能
- 配布時の影響は軽微

## エラーハンドリング

### エラーケース

1. **フォント読み込み失敗**
   - 埋め込みフォントの破損（通常発生しない）
   - メモリ不足

2. **フォールバック失敗**
   - 日本語フォントが使用できない場合
   - 標準フォントでフォールバック

### エラー処理方針

```go
// エラーが発生した場合の挙動
jpFont, err := font.DefaultJapaneseFont()
if err != nil {
    // エラーログを出力
    log.Printf("Warning: Failed to load Japanese font: %v", err)
    // 標準フォントで継続（Helvetica等）
    page.SetFont(font.Helvetica, 16)
}
```

## テスト戦略

### ユニットテスト

1. **埋め込みフォントテスト**
   - `DefaultJapaneseFont()`の呼び出しテスト
   - キャッシュ機構のテスト
   - 並行アクセステスト

2. **文字種判定テスト**
   - ASCII文字の判定
   - ひらがな・カタカナ・漢字の判定
   - 混在テキストの分割テスト

3. **自動描画テスト**
   - 混在テキストの描画テスト
   - フォント切り替えの確認

### 統合テスト

1. **PDF生成テスト**
   - 日本語テキストのPDF生成
   - 混在テキストのPDF生成
   - PDFの妥当性確認

2. **パフォーマンステスト**
   - 大量テキストの描画速度
   - メモリ使用量の測定

## マイルストーン

### Week 1: フォント埋め込み基盤

- [ ] Noto Sans JPのダウンロードとサブセット化
- [ ] `internal/font/embedded/`の実装
- [ ] `DefaultJapaneseFont()`関数の実装
- [ ] ユニットテスト作成

### Week 2: 自動フォールバック機能

- [ ] 文字種判定機能の実装
- [ ] `DrawTextAuto()`関数の実装
- [ ] 統合テスト作成

### Week 3: サンプル・ドキュメント

- [ ] サンプルコード作成
- [ ] ドキュメント更新
- [ ] ライセンス情報の整備

## 将来の拡張

### 追加機能の候補

1. **複数ウェイト対応**
   - Bold, Light等の追加
   - ウェイト自動選択API

2. **他のフォント埋め込み**
   - 中国語（簡体字・繁体字）
   - 韓国語（ハングル）
   - その他のCJK言語

3. **カスタムフォント登録**
   - ユーザーが独自の埋め込みフォントを追加
   - フォントレジストリの実装

4. **動的フォントダウンロード**
   - オンデマンドでフォントをダウンロード
   - キャッシュディレクトリの管理
   - オフライン対応

## 参考資料

- [Noto Fonts](https://fonts.google.com/noto)
- [SIL Open Font License 1.1](https://scripts.sil.org/OFL)
- [Go embed package](https://pkg.go.dev/embed)
- [fonttools/fonttools](https://github.com/fonttools/fonttools)
- [Unicode Character Ranges](https://en.wikipedia.org/wiki/Unicode_block)

## まとめ

この設計により、gopdfユーザーは：
- フォントを手動で配置する必要がなくなる
- デフォルトで日本語が使用可能
- シンプルなAPIで自動フォント切り替え
- 商用利用可能なライセンス

実装の優先度は高く、ユーザーエクスペリエンスを大幅に向上させる重要な機能となる。
