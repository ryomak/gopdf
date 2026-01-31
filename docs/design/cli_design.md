# gopdf CLI 設計書

## 概要

gopdfライブラリの機能をコマンドラインから簡単に利用できるCLIツール。

## コマンド構造

```
gopdf <command> [subcommand] [flags]
```

## 実装技術

- **CLI Framework**: [cobra](https://github.com/spf13/cobra)
- **出力整形**: 標準出力（JSON出力オプション対応）

## コマンド一覧

### 1. `info` - PDF情報表示

PDFファイルの基本情報を表示する。

```bash
gopdf info <file.pdf> [flags]

# 例
gopdf info document.pdf
gopdf info document.pdf --json
gopdf info document.pdf --password "secret"
```

**出力内容:**
- ファイル名
- ページ数
- 各ページのサイズ
- メタデータ（タイトル、作成者、作成日時等）
- 暗号化状態

**フラグ:**
- `--json`: JSON形式で出力
- `--password, -p`: パスワード保護されたPDFを開く

### 2. `extract` - コンテンツ抽出

PDFからテキストや画像を抽出する。

#### 2.1 `extract text` - テキスト抽出

```bash
gopdf extract text <file.pdf> [flags]

# 例
gopdf extract text document.pdf
gopdf extract text document.pdf --page 1
gopdf extract text document.pdf --output text.txt
gopdf extract text document.pdf --format blocks
```

**フラグ:**
- `--page, -p`: 抽出するページ番号（1始まり、指定しない場合は全ページ）
- `--output, -o`: 出力ファイル（指定しない場合は標準出力）
- `--format, -f`: 出力形式（`plain`|`blocks`|`json`）
  - `plain`: プレーンテキスト（デフォルト）
  - `blocks`: テキストブロック単位で出力
  - `json`: 座標情報付きJSON
- `--password`: パスワード

#### 2.2 `extract images` - 画像抽出

```bash
gopdf extract images <file.pdf> [flags]

# 例
gopdf extract images document.pdf --output ./images/
gopdf extract images document.pdf --page 1 --output ./images/
```

**フラグ:**
- `--page, -p`: 抽出するページ番号（指定しない場合は全ページ）
- `--output, -o`: 出力ディレクトリ（デフォルト: `./extracted_images/`）
- `--format, -f`: 出力形式（`original`|`png`|`jpg`）
- `--password`: パスワード

### 3. `translate` - PDF翻訳

PDFのレイアウトを維持したまま翻訳する。

```bash
gopdf translate <input.pdf> <output.pdf> [flags]

# 例
gopdf translate input.pdf output.pdf --font ./NotoSansJP-Regular.ttf
```

**フラグ:**
- `--font, -f`: 使用するTTFフォントファイル（必須）
- `--source-lang, -s`: 元の言語（デフォルト: `en`）
- `--target-lang, -t`: 翻訳先の言語（デフォルト: `ja`）
- `--translator`: 翻訳サービス（将来拡張用）
- `--password`: パスワード

**注意:** 翻訳機能を使うには、外部の翻訳関数を提供する必要がある。CLIでは翻訳関数をフックとして渡す方法（環境変数やコマンド指定）を検討。

### 4. `markdown` - Markdown変換

MarkdownファイルをPDFに変換する。

```bash
gopdf markdown <input.md> <output.pdf> [flags]

# 例
gopdf markdown README.md readme.pdf
gopdf markdown slides.md presentation.pdf --mode slide
```

**フラグ:**
- `--mode, -m`: 変換モード（`document`|`slide`）
- `--page-size`: ページサイズ（`A4`|`Letter`|`16:9`|`4:3`）
- `--orientation`: ページの向き（`portrait`|`landscape`）
- `--font`: 日本語フォント（TTFファイル）

### 5. `encrypt` - 暗号化

PDFにパスワード保護を設定する。

```bash
gopdf encrypt <input.pdf> <output.pdf> [flags]

# 例
gopdf encrypt document.pdf protected.pdf --user-password "user123"
gopdf encrypt document.pdf protected.pdf --user-password "user123" --owner-password "owner456"
```

**フラグ:**
- `--user-password`: ユーザーパスワード（開くためのパスワード）
- `--owner-password`: オーナーパスワード（権限制御用）
- `--algorithm`: 暗号化アルゴリズム（`rc4-128`|`aes-128`|`aes-256`）
- `--no-print`: 印刷禁止
- `--no-copy`: コピー禁止
- `--no-modify`: 編集禁止

### 6. `decrypt` - 復号

パスワード保護されたPDFを復号する。

```bash
gopdf decrypt <input.pdf> <output.pdf> --password <password>

# 例
gopdf decrypt protected.pdf decrypted.pdf --password "secret"
```

**フラグ:**
- `--password, -p`: パスワード（必須）

### 7. `metadata` - メタデータ操作

PDFのメタデータを読み取り・設定する。

#### 7.1 `metadata get` - メタデータ読み取り

```bash
gopdf metadata get <file.pdf> [flags]

# 例
gopdf metadata get document.pdf
gopdf metadata get document.pdf --json
```

**フラグ:**
- `--json`: JSON形式で出力
- `--password`: パスワード

#### 7.2 `metadata set` - メタデータ設定

```bash
gopdf metadata set <input.pdf> <output.pdf> [flags]

# 例
gopdf metadata set input.pdf output.pdf --title "My Document" --author "John Doe"
```

**フラグ:**
- `--title`: タイトル
- `--author`: 作成者
- `--subject`: 件名
- `--keywords`: キーワード（カンマ区切り）
- `--creator`: 作成アプリケーション
- `--password`: パスワード

### 8. `merge` - PDF結合

複数のPDFファイルを1つに結合する。

```bash
gopdf merge <output.pdf> <input1.pdf> <input2.pdf> ... [flags]

# 例
gopdf merge combined.pdf file1.pdf file2.pdf file3.pdf
```

**フラグ:**
- `--password`: 出力PDFにパスワードを設定

### 9. `split` - PDF分割

PDFファイルをページ単位で分割する。

```bash
gopdf split <input.pdf> <output-dir> [flags]

# 例
gopdf split document.pdf ./pages/
gopdf split document.pdf ./pages/ --pages 1-5,10,15-20
```

**フラグ:**
- `--pages, -p`: 分割するページ範囲（例: `1-5,10,15-20`）
- `--prefix`: 出力ファイルのプレフィックス（デフォルト: `page_`）
- `--password`: パスワード

### 10. `version` - バージョン表示

```bash
gopdf version
```

## ディレクトリ構造

```
cmd/
└── gopdf/
    ├── main.go           # エントリーポイント
    ├── root.go           # rootコマンド
    ├── info.go           # infoコマンド
    ├── extract.go        # extractコマンド
    ├── translate.go      # translateコマンド
    ├── markdown.go       # markdownコマンド
    ├── encrypt.go        # encryptコマンド
    ├── decrypt.go        # decryptコマンド
    ├── metadata.go       # metadataコマンド
    ├── merge.go          # mergeコマンド
    ├── split.go          # splitコマンド
    └── version.go        # versionコマンド
```

## 実装優先順位

### Phase 1（必須）
1. `info` - PDF情報表示
2. `extract text` - テキスト抽出
3. `extract images` - 画像抽出
4. `version` - バージョン表示

### Phase 2（高優先度）
5. `encrypt` - 暗号化
6. `decrypt` - 復号
7. `metadata` - メタデータ操作

### Phase 3（中優先度）
8. `markdown` - Markdown変換
9. `merge` - PDF結合
10. `split` - PDF分割

### Phase 4（低優先度）
11. `translate` - PDF翻訳（翻訳APIの統合が必要）

## エラーハンドリング

- ファイルが見つからない場合: 終了コード1、エラーメッセージを標準エラー出力
- パスワードが間違っている場合: 終了コード2、適切なエラーメッセージ
- 不正なPDF形式: 終了コード3、エラーメッセージ
- 成功: 終了コード0

## グローバルフラグ

- `--help, -h`: ヘルプを表示
- `--verbose, -v`: 詳細出力モード
- `--quiet, -q`: エラーのみ出力

## 将来の拡張

- シェル補完（bash, zsh, fish, powershell）
- 設定ファイルのサポート（`~/.gopdf.yaml`）
- プラグインシステム（翻訳サービスの追加など）
