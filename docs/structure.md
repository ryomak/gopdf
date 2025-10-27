# gopdf プロジェクト構造設計書

## 1. 概要

本ドキュメントは、`gopdf` プロジェクトのディレクトリ構造、パッケージ構成、およびファイルの配置方針を定義する。

## 2. ディレクトリ構造

```
gopdf/
├── go.mod                        # Go モジュール定義
├── go.sum                        # 依存関係のチェックサム
├── README.md                     # プロジェクト概要・使用方法
├── CLAUDE.md                     # AI開発指示書
├── LICENSE                       # ライセンス（MIT想定）
│
├── docs/                         # ドキュメント
│   ├── requirements.md           # 要件定義書
│   ├── architecture.md           # アーキテクチャ設計書
│   ├── structure.md              # 本ドキュメント
│   ├── progress.md               # 進捗管理・実装状況
│   ├── api_design.md             # API設計詳細
│   └── pdf_spec_notes.md         # PDF仕様のメモ・参考資料
│
├── pkg/                          # 公開パッケージ
│   └── gopdf/                    # メインの公開API
│       ├── gopdf.go              # パッケージのエントリーポイント
│       ├── document.go           # Document型の定義
│       ├── page.go               # Page型の定義
│       ├── options.go            # オプション型の定義
│       ├── constants.go          # 定数（ページサイズ、色など）
│       ├── document_test.go
│       ├── page_test.go
│       └── README.md             # パッケージの使い方
│
├── internal/                     # 内部パッケージ（外部非公開）
│   │
│   ├── core/                     # PDF基本オブジェクト
│   │   ├── object.go             # Object インターフェース、基本型
│   │   ├── dictionary.go         # Dictionary 型
│   │   ├── array.go              # Array 型
│   │   ├── stream.go             # Stream 型
│   │   ├── reference.go          # Reference 型（間接参照）
│   │   ├── name.go               # Name 型
│   │   ├── catalog.go            # Catalog 辞書
│   │   ├── pages.go              # Pages ツリー
│   │   ├── object_test.go
│   │   └── README.md
│   │
│   ├── font/                     # フォント管理
│   │   ├── font.go               # Font インターフェース
│   │   ├── standard.go           # 標準フォント実装
│   │   ├── standard_metrics.go   # 標準フォントのメトリクス
│   │   ├── ttf.go                # TTFフォント解析
│   │   ├── ttf_parser.go         # TTF内部パーサー
│   │   ├── subset.go             # サブセット化
│   │   ├── encoding.go           # エンコーディング処理
│   │   ├── font_test.go
│   │   └── README.md
│   │
│   ├── writer/                   # PDF生成・書き込み
│   │   ├── writer.go             # Writer 型、出力制御
│   │   ├── serializer.go         # オブジェクトのシリアライズ
│   │   ├── xref.go               # クロスリファレンステーブル生成
│   │   ├── trailer.go            # トレーラー生成
│   │   ├── compress.go           # ストリーム圧縮
│   │   ├── writer_test.go
│   │   └── README.md
│   │
│   ├── reader/                   # PDF解析・読み込み
│   │   ├── reader.go             # Reader 型、読み込み制御
│   │   ├── lexer.go              # 字句解析（トークナイザー）
│   │   ├── parser.go             # 構文解析（パーサー）
│   │   ├── xref_reader.go        # xref解析
│   │   ├── trailer_reader.go     # trailer解析
│   │   ├── decrypt.go            # 暗号化PDF復号
│   │   ├── reader_test.go
│   │   └── README.md
│   │
│   ├── content/                  # コンテンツ生成・抽出
│   │   ├── stream.go             # コンテンツストリーム操作
│   │   ├── graphics_state.go     # グラフィック状態管理
│   │   ├── text_drawer.go        # テキスト描画
│   │   ├── image_drawer.go       # 画像描画
│   │   ├── shape_drawer.go       # 図形描画
│   │   ├── text_extractor.go     # テキスト抽出
│   │   ├── image_extractor.go    # 画像抽出
│   │   ├── link_extractor.go     # リンク抽出
│   │   ├── operators.go          # PDFオペレーター定義
│   │   ├── content_test.go
│   │   └── README.md
│   │
│   ├── util/                     # ユーティリティ
│   │   ├── buffer.go             # バッファ操作
│   │   ├── color.go              # 色変換
│   │   ├── units.go              # 単位変換（ポイント、mm、inchなど）
│   │   └── util_test.go
│   │
│   └── testutil/                 # テスト用ユーティリティ
│       ├── pdf_validator.go      # PDF検証ヘルパー
│       └── mock.go               # モック
│
├── examples/                     # サンプルコード
│   ├── 01_hello_world/           # 最小構成のPDF生成
│   │   └── main.go
│   ├── 02_text_styling/          # テキストスタイリング
│   │   └── main.go
│   ├── 03_shapes/                # 図形描画
│   │   └── main.go
│   ├── 04_images/                # 画像埋め込み
│   │   └── main.go
│   ├── 05_ttf_font/              # TTFフォント使用
│   │   └── main.go
│   ├── 06_read_pdf/              # PDF読み込み・解析
│   │   └── main.go
│   ├── 07_extract_text/          # テキスト抽出
│   │   └── main.go
│   └── 08_modify_existing/       # 既存PDF変更
│       └── main.go
│
├── testdata/                     # テストデータ
│   ├── pdfs/                     # 既存PDFサンプル
│   │   ├── simple.pdf
│   │   ├── with_images.pdf
│   │   ├── encrypted.pdf
│   │   └── japanese.pdf
│   ├── fonts/                    # テスト用フォント
│   │   └── test.ttf
│   └── images/                   # テスト用画像
│       ├── test.jpg
│       └── test.png
│
├── scripts/                      # 開発用スクリプト
│   ├── test.sh                   # テスト実行
│   ├── bench.sh                  # ベンチマーク実行
│   └── lint.sh                   # Lint実行
│
└── .github/                      # GitHub設定（将来）
    └── workflows/
        └── ci.yml                # CI/CD設定
```

## 3. パッケージ責務詳細

### 3.1. `pkg/gopdf` (公開API)

**目的:** ユーザーに提供する高レベルAPI

**公開する主要型:**
- `Document`: PDFドキュメント全体を表現
- `Page`: PDFの1ページ
- `PageSize`: ページサイズ（A4, Letter等）
- `Color`: 色の表現
- `Font`: フォント（内部実装は隠蔽）

**公開する主要関数:**
- `New() *Document`: 新規ドキュメント作成
- `Open(r io.Reader) (*Document, error)`: 既存PDF読み込み
- `(d *Document) NewPage(size PageSize, orientation Orientation) *Page`
- `(d *Document) WriteTo(w io.Writer) error`
- `(p *Page) DrawText(text string, x, y float64, opts ...TextOption) error`
- `(p *Page) DrawImage(img image.Image, x, y, w, h float64) error`
- `(p *Page) ExtractText() (string, error)`

### 3.2. `internal/core`

**目的:** PDF仕様の基本オブジェクトを実装

**内部型:**
- `Object`: すべてのPDFオブジェクトのインターフェース
- `Dictionary`: PDF辞書（`<< /Key /Value >>`）
- `Array`: PDF配列（`[ item1 item2 ]`）
- `Stream`: PDFストリーム（辞書+バイナリデータ）
- `Reference`: 間接参照（`1 0 R`）
- `Name`, `String`, `Integer`, `Real`, `Boolean`, `Null`

**責務:**
- PDF内部表現の型定義
- 型変換・検証

### 3.3. `internal/font`

**目的:** フォント管理と解析

**内部型:**
- `Font`: フォントインターフェース
- `StandardFont`: 標準Type1フォント
- `TTFFont`: TrueTypeフォント
- `FontDescriptor`: フォント記述子
- `Encoding`: エンコーディング情報

**責務:**
- フォントファイルの解析
- グリフ情報の取得
- エンコーディング変換
- サブセット化

### 3.4. `internal/writer`

**目的:** PDFバイナリの生成と出力

**内部型:**
- `Writer`: PDF書き込みの制御
- `Serializer`: オブジェクトのシリアライズ
- `XRefTable`: クロスリファレンステーブル

**責務:**
- PDF構造の組み立て（header, body, xref, trailer）
- オブジェクトのバイナリ化
- ストリームの圧縮
- オフセットの管理

### 3.5. `internal/reader`

**目的:** PDFバイナリの解析と読み込み

**内部型:**
- `Reader`: PDF読み込みの制御
- `Lexer`: 字句解析
- `Parser`: 構文解析
- `CrossReference`: xrefテーブルの内部表現

**責務:**
- PDFファイルのパース
- オブジェクトの取得
- xref/trailerの解析
- 暗号化PDFの復号

### 3.6. `internal/content`

**目的:** ページコンテンツの生成と抽出

**内部型:**
- `ContentStream`: コンテンツストリーム
- `GraphicsState`: グラフィック状態
- `TextState`: テキスト状態
- `Operator`: PDFオペレーター

**責務:**
- コンテンツストリームの生成
- 描画オペレーターの発行
- コンテンツストリームの解析
- テキスト・画像の抽出

## 4. ファイル命名規則

### 4.1. 実装ファイル
- 小文字＋アンダースコア: `text_drawer.go`
- 1ファイル1主要型を原則とする
- 関連する型は同じファイルに配置可能

### 4.2. テストファイル
- 実装ファイル名 + `_test.go`: `text_drawer_test.go`
- テーブル駆動テストを推奨
- Example関数でドキュメント兼テスト

### 4.3. ドキュメント
- `README.md`: パッケージの概要・使い方
- `doc.go`: godocコメント（パッケージレベル）

## 5. import ルール

### 5.1. 公開API (`pkg/gopdf`)
- `internal/*` を直接import可能
- 外部パッケージは最小限に

### 5.2. `internal` パッケージ
- 同じ `internal` 配下のパッケージ間でのimport可能
- `pkg/gopdf` からimportされることを想定
- 循環importを避ける設計

### 5.3. import順序
1. 標準ライブラリ
2. 外部ライブラリ（Pure Goのみ）
3. 内部パッケージ

```go
import (
    "fmt"
    "io"

    "golang.org/x/image/font/sfnt"

    "github.com/ryomak/gopdf/internal/core"
    "github.com/ryomak/gopdf/internal/font"
)
```

## 6. テストデータ管理

### 6.1. testdata配置
- `testdata/` ディレクトリはGoツールチェインが特別扱い
- テストからは相対パス `testdata/` でアクセス可能

### 6.2. PDFサンプル
- シンプルなPDF: 基本的な機能テスト用
- 複雑なPDF: 実際のユースケースを想定
- 暗号化PDF: 復号機能のテスト用

## 7. 開発フロー

### 7.1. 新規機能開発
1. `docs/` に設計書を作成（または更新）
2. `internal/` に実装（TDD）
3. `pkg/gopdf` にAPIを公開
4. `examples/` にサンプルを追加
5. `docs/progress.md` に進捗を記録

### 7.2. テスト実行
```bash
# 全テスト
go test ./...

# カバレッジ
go test -cover ./...

# ベンチマーク
go test -bench=. ./...
```

### 7.3. コード品質
- `gofmt` でフォーマット
- `go vet` で静的解析
- `golangci-lint` でLint（推奨）

## 8. バージョニング

- セマンティックバージョニング採用: `v0.1.0`, `v1.0.0`...
- `v0.x.x`: 開発中、破壊的変更あり
- `v1.0.0`: 安定版、後方互換性を保証

## 9. ライセンス

- MIT License（予定）
- `LICENSE` ファイルに記載
