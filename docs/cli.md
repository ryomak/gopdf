# gopdf CLI

コマンドラインからPDF操作を行うCLIツール。

## インストール

```bash
go install github.com/ryomak/gopdf/cmd/gopdf@latest
```

または、ソースからビルド:

```bash
git clone https://github.com/ryomak/gopdf.git
cd gopdf
go build -o gopdf ./cmd/gopdf
```

## コマンド一覧

| コマンド | 説明 |
|---------|------|
| `info` | PDF情報の表示 |
| `extract text` | テキスト抽出 |
| `extract images` | 画像抽出 |
| `metadata get` | メタデータ表示 |
| `encrypt` | PDF暗号化 |
| `decrypt` | PDF復号 |
| `markdown` | Markdown→PDF変換 |
| `translate` | PDF翻訳 |
| `version` | バージョン表示 |

## 使用例

### PDF情報の表示

```bash
# 基本的な情報表示
gopdf info document.pdf

# JSON形式で出力
gopdf info document.pdf --json

# パスワード保護されたPDF
gopdf info protected.pdf --password "secret"
```

出力例:
```
File: document.pdf
Pages: 10
Encrypted: false

Page Sizes:
  Page 1: 595.00 x 842.00 pt
  Page 2: 595.00 x 842.00 pt
  ...

Metadata:
  Title: My Document
  Author: John Doe
  Created: 2025-01-15 10:30:00
```

### テキスト抽出

```bash
# プレーンテキストとして抽出
gopdf extract text document.pdf

# 特定のページのみ抽出（1ページ目）
gopdf extract text document.pdf --page 1

# ファイルに出力
gopdf extract text document.pdf --output text.txt

# テキストブロック形式で出力（位置情報付き）
gopdf extract text document.pdf --format blocks

# JSON形式で出力（座標情報付き）
gopdf extract text document.pdf --format json --output text.json
```

### 画像抽出

```bash
# 全ページの画像を抽出
gopdf extract images document.pdf --output ./images/

# 特定のページから抽出
gopdf extract images document.pdf --page 3 --output ./images/
```

### メタデータ表示

```bash
# テキスト形式
gopdf metadata get document.pdf

# JSON形式
gopdf metadata get document.pdf --json
```

### PDF暗号化

```bash
# 基本的な暗号化（128-bit）
gopdf encrypt input.pdf output.pdf --user-password "user123"

# オーナーパスワードも設定
gopdf encrypt input.pdf output.pdf \
  --user-password "user123" \
  --owner-password "owner456"

# 権限を制限（印刷・コピー・変更を禁止）
gopdf encrypt input.pdf output.pdf \
  --user-password "user123" \
  --no-print \
  --no-copy \
  --no-modify

# 40-bit暗号化（互換性重視）
gopdf encrypt input.pdf output.pdf \
  --user-password "user123" \
  --key-length 40
```

### PDF復号

```bash
gopdf decrypt protected.pdf decrypted.pdf --password "secret"
```

### Markdown変換

```bash
# ドキュメント形式でPDF生成
gopdf markdown README.md output.pdf

# プレゼンテーションスライド形式
gopdf markdown slides.md presentation.pdf --mode slide

# ページサイズと向きを指定
gopdf markdown document.md output.pdf \
  --page-size A4 \
  --orientation landscape

# 暗号化オプション付き
gopdf markdown document.md output.pdf \
  --user-password "secret"
```

対応モード:
- `document` - 通常のドキュメント（デフォルト）
- `slide` - プレゼンテーションスライド

対応ページサイズ:
- `A4`, `A3`, `A5`, `Letter`, `Legal`
- `16:9`, `4:3` - プレゼンテーション用

### PDF翻訳

外部の翻訳コマンドを使用してPDFを翻訳します。

```bash
# translate-shell (trans) を使用
gopdf translate input.pdf output.pdf \
  --font ./NotoSansJP-Regular.ttf \
  --command "trans -b :ja"

# カスタムスクリプトを使用
gopdf translate input.pdf output.pdf \
  --font ./font.ttf \
  --command "./my-translator.sh"

# 翻訳単位を指定
gopdf translate input.pdf output.pdf \
  --font ./font.ttf \
  --command "trans -b :ja" \
  --unit sentence  # block, line, sentence

# ドライラン（テキスト抽出のみ、翻訳なし）
gopdf translate input.pdf output.pdf --dry-run
```

翻訳コマンドの要件:
- 標準入力からテキストを受け取る
- 翻訳結果を標準出力に出力する

例: `trans -b :ja` は translate-shell を使用して日本語に翻訳します。

## グローバルオプション

| オプション | 説明 |
|-----------|------|
| `--help`, `-h` | ヘルプを表示 |
| `--verbose`, `-v` | 詳細出力モード |
| `--quiet`, `-q` | エラーのみ出力 |

## シェル補完

```bash
# Bash
gopdf completion bash > /etc/bash_completion.d/gopdf

# Zsh
gopdf completion zsh > "${fpath[1]}/_gopdf"

# Fish
gopdf completion fish > ~/.config/fish/completions/gopdf.fish

# PowerShell
gopdf completion powershell > gopdf.ps1
```

## 終了コード

| コード | 説明 |
|--------|------|
| 0 | 成功 |
| 1 | 一般的なエラー |

## 関連ドキュメント

- [PDF読み込み・解析](reading-pdf.md) - ライブラリとしての使用方法
- [暗号化とパスワード保護](encryption.md) - 暗号化の詳細
- [PDF翻訳](translation.md) - 翻訳機能の詳細
- [Markdown変換](markdown.md) - Markdown変換の詳細
