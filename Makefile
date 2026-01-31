.PHONY: help test lint ci deps build clean

# デフォルトターゲット
.DEFAULT_GOAL := help

help: ## ヘルプを表示
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

deps: ## 依存関係をダウンロード
	go mod download

test: ## テストを実行（カバレッジ付き）
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

lint: ## Lintを実行
	golangci-lint run

ci: deps test lint ## CI相当の処理を実行（deps → test → lint）
	@echo "✓ All CI checks passed!"

build: ## CLIをビルド（Gitタグからバージョンを取得）
	@VERSION=$$(git describe --tags --always --dirty 2>/dev/null || echo "dev"); \
	go build -ldflags "-s -w -X github.com/ryomak/gopdf/cmd/gopdf/cmd.Version=$$VERSION" -o gopdf ./cmd/gopdf

clean: ## ビルド成果物とキャッシュをクリーン
	go clean
	rm -f coverage.out gopdf
