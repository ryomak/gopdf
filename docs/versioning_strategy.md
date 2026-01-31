# バージョニング戦略

## 概要
gopdfプロジェクトでは、mainブランチへの変更時に自動的にバージョンタグを作成し、セマンティックバージョニング（Semantic Versioning）を適用する。

## バージョニングルール

### Semantic Versioning
- フォーマット: `vMAJOR.MINOR.PATCH` (例: v1.2.3)
- MAJOR: 破壊的変更（Breaking Changes）
- MINOR: 新機能追加（Features）
- PATCH: バグフィックス（Bug Fixes）

### Conventional Commits
コミットメッセージの接頭辞でバージョンの種類を判定：

| Commit Type | Version Bump | 例 |
|-------------|--------------|-----|
| `feat:` | MINOR | feat: Add new PDF parsing feature |
| `fix:` | PATCH | fix: Correct text encoding issue |
| `perf:` | PATCH | perf: Improve rendering performance |
| `BREAKING CHANGE:` | MAJOR | feat!: Change API signature |
| `docs:`, `chore:`, `style:`, `refactor:`, `test:` | なし | 既存のバージョンを維持 |

## 自動化の仕組み

### GitHub Actions ワークフロー

#### test.yml
mainブランチへのpushまたはPR時に実行：
1. **test**: 複数のGoバージョンでテストを実行
2. **lint**: golangci-lintでコード品質をチェック
3. **auto-tag** (mainへのpushのみ):
   - testとlintが成功した後に実行
   - 最新のタグを取得
   - 最新タグ以降のコミットメッセージを解析
   - Conventional Commitsに基づいてバージョンをインクリメント
   - 新しいバージョンタグを作成してGitHubにpush
   - GoReleaserでGitHubリリースを作成

#### release.yml
タグがpushされた時に実行（手動実行用）：
1. **test**: リリース前のテスト
2. **release**: GoReleaserでリリースを作成

### 使用ツール
- **GitHub Actions**: ワークフローの実行
- **シェルスクリプト**: バージョン計算とタグ作成（Pure Bash実装）
- **GoReleaser**: リリース成果物の作成

## 初期バージョン
- プロジェクトの最初のタグ: `v0.1.0`
- v1.0.0までは開発版として扱う

## 手動タグ作成（緊急時）
必要に応じて手動でタグを作成することも可能：
```bash
git tag -a v1.2.3 -m "Release v1.2.3"
git push origin v1.2.3
```

## タグの命名規則
- すべてのタグは `v` で始まる
- Goモジュールの標準に準拠
- 例: `v0.1.0`, `v1.0.0`, `v1.2.3`

## ワークフロー概要図
```
mainへのpush
    ↓
test.yml実行
    ├── test（複数Goバージョン）
    ├── lint
    └── auto-tag（mainのみ）
        ├── バージョン決定
        ├── タグ作成
        └── GitHubリリース作成
```

## 注意事項

### リリースの自動化について
- mainブランチへのpushで、テストとlintが成功すると自動的にタグとリリースが作成されます
- `feat:`, `fix:`, `perf:` のコミットがある場合のみバージョンがインクリメントされます
- `docs:`, `chore:`, `test:` などのコミットだけではタグは作成されません

### GITHUB_TOKENの制限
- GitHub Actionsの`GITHUB_TOKEN`で作成したタグは、新しいワークフローをトリガーしません
- そのため、test.ymlのauto-tagジョブ内でタグ作成とリリース作成を統合しています
- release.ymlは手動でタグをpushした場合に実行されます
