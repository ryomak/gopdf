# Embedded Japanese Font

このディレクトリには、gopdfにデフォルトで埋め込まれる日本語フォントが配置されます。

## フォントのダウンロード

以下のコマンドでフォントをダウンロードできます：

```bash
cd internal/font/embedded
./download_font.sh
```

## 手動ダウンロード

自動ダウンロードが失敗する場合は、手動でダウンロードしてください：

### Koruri（推奨）

1. https://github.com/Koruri/Koruri にアクセス
2. `Koruri-Regular.ttf` をダウンロード
3. このディレクトリに配置

### ファイルサイズ確認

Koruriフォントは軽量です：

```bash
ls -lh Koruri-Regular.ttf
file Koruri-Regular.ttf
```

`TrueType Font data` と表示されることを確認してください。

## ライセンス

使用フォント: Koruri（小瑠璃）
ライセンス: Apache License 2.0
構成: M+ FONTS + Open Sans
著作権: Koruri Project

詳細は `LICENSE.txt` を参照してください。

## 開発者向け情報

### フォント埋め込みの仕組み

`embed.go` で `go:embed` ディレクティブを使用してフォントをバイナリに埋め込みます：

```go
//go:embed Koruri-Regular.ttf
var KoruriRegular []byte
```

### ビルド時の注意

- フォントファイルがない場合、ビルドエラーになります
- 開発時は `download_font.sh` を実行してフォントを取得してください
- CIでは自動的にダウンロードされます（予定）

### フォントの更新

新しいバージョンのフォントに更新する場合：

1. 古いフォントファイルを削除
2. `download_font.sh` を再実行
3. テストを実行して互換性を確認
4. コミットして更新

## トラブルシューティング

### ビルドエラー: pattern Koruri-Regular.ttf: no matching files found

フォントファイルがダウンロードされていません。`./download_font.sh` を実行してください。

### フォントが読み込めない

- ファイル形式がTTFであることを確認
- OTFやWOFF2は未対応
- ファイルサイズを確認（Koruriは約1-2MB）
