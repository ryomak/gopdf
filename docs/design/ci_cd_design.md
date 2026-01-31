# CI/CD設計書

## 概要
gopdfプロジェクトのための継続的インテグレーション/継続的デリバリー（CI/CD）パイプラインの設計。

## 目的
- コードの品質を自動的に検証
- プルリクエストとmainブランチへのマージ時にテストを実行
- バージョンタグ作成時に自動リリース
- マルチプラットフォームバイナリの自動ビルドと配布

## GitHub Actions ワークフロー

### 1. テスト実行ワークフロー (test.yml)

#### トリガー
- `main`ブランチへのプッシュ
- プルリクエストの作成・更新

#### ジョブ構成

##### テストジョブ
- **実行環境**: Ubuntu Latest
- **Go バージョン**: 1.21, 1.22（マトリックス戦略）
- **ステップ**:
  1. コードのチェックアウト
  2. Go環境のセットアップ
  3. Goモジュールのキャッシュ（ビルド高速化）
  4. 依存関係のダウンロード
  5. テスト実行（race detector有効、カバレッジ計測）
  6. Codecovへのカバレッジアップロード（Go 1.22のみ）

##### Lintジョブ
- **実行環境**: Ubuntu Latest
- **Go バージョン**: 1.22
- **ステップ**:
  1. コードのチェックアウト
  2. Go環境のセットアップ
  3. golangci-lintの実行

#### テストオプション
```bash
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
```
- `-v`: 詳細出力
- `-race`: データ競合検出
- `-coverprofile`: カバレッジプロファイル生成
- `-covermode=atomic`: アトミックカバレッジモード

### 2. リリースワークフロー (release.yml)

#### トリガー
- `v*`パターンのタグプッシュ（例: v1.0.0, v2.1.3）

#### ジョブ構成

##### テストジョブ（リリース前）
- リリース前にすべてのテストを実行
- 失敗した場合はリリースを中止

##### リリースジョブ
- **依存**: テストジョブが成功
- **GoReleaser使用**: クロスプラットフォームビルドとリリース自動化
- **パーミッション**: `contents: write`（GitHub Releases作成のため）

## GoReleaser設定 (.goreleaser.yml)

### ビルド設定
```yaml
builds:
  - env:
      - CGO_ENABLED=0  # 静的バイナリ生成
    goos: [linux, windows, darwin]
    goarch: [amd64, arm64]
```

#### サポートプラットフォーム
- **OS**: Linux, Windows, macOS
- **アーキテクチャ**: amd64 (x86_64), arm64

#### ビルドフラグ
```
-s -w                          # シンボル削除、バイナリサイズ削減
-X main.version={{.Version}}   # バージョン情報埋め込み
-X main.commit={{.Commit}}     # コミットハッシュ埋め込み
-X main.date={{.Date}}         # ビルド日時埋め込み
```

### アーカイブ設定
- **形式**: tar.gz（Windows: zip）
- **命名規則**: `gopdf_<OS>_<Arch>.tar.gz`
  - 例: `gopdf_Linux_x86_64.tar.gz`

### チェンジログ生成
自動的に以下のカテゴリに分類：
- **Features** (`feat:` プレフィックス)
- **Bug Fixes** (`fix:` プレフィックス)
- **Others** (その他)

除外パターン：
- `docs:` - ドキュメント変更
- `test:` - テスト変更
- `chore:` - 雑務
- `ci:` - CI変更

### GitHub Releases
- **リポジトリ**: ryomak/gopdf
- **ドラフト**: 無効（自動公開）
- **プレリリース**: 自動判定
- **リリース名**: `gopdf-v{version}`

## 使用方法

### 通常の開発フロー
1. フィーチャーブランチで開発
2. プルリクエスト作成
3. 自動的にテストとLintが実行
4. レビュー後、mainブランチにマージ
5. mainブランチでも再度テスト実行

### リリースフロー
1. リリース準備ができたら、バージョンタグを作成
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```
2. 自動的にテスト実行
3. テスト成功後、GoReleaserが実行
4. マルチプラットフォームバイナリのビルド
5. GitHub Releasesへの自動公開
6. チェンジログの自動生成

## キャッシュ戦略
Goモジュールとビルドキャッシュを活用して、CI実行時間を短縮：
```yaml
~/.cache/go-build  # ビルドキャッシュ
~/go/pkg/mod       # モジュールキャッシュ
```

キャッシュキー: `{OS}-go-{version}-{go.sum hash}`

## セキュリティ
- **GITHUB_TOKEN**: GitHub Actionsが自動的に提供
- **権限**: 必要最小限（contents: write）
- **CGO無効**: 静的バイナリでセキュリティリスク軽減

## モニタリング
- テスト結果はGitHub Actions UIで確認
- カバレッジレポートはCodecovで追跡
- リリース成果物はGitHub Releasesページで確認

## 今後の拡張案
1. **パフォーマンステスト**: ベンチマーク結果の追跡
2. **ドキュメント自動生成**: godocの自動デプロイ
3. **コンテナイメージ**: Docker imageの自動ビルド
4. **依存関係スキャン**: セキュリティ脆弱性チェック
5. **自動バージョンバンプ**: セマンティックバージョニング自動化
