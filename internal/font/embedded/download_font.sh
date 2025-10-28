#!/bin/bash
# Koruriフォントをダウンロードするスクリプト

set -e

echo "==================================="
echo "Koruri Font Download Script"
echo "==================================="
echo ""
echo "このスクリプトはKoruri Regularフォントをダウンロードします。"
echo ""
echo "Koruriについて:"
echo "- M+ FONTS と Open Sans を合成した日本語フォント"
echo "- 英数字部分が美しい Open Sans ベース"
echo "- 軽量で読みやすい"
echo ""
echo "ライセンス: Apache License 2.0 (商用利用可)"
echo ""

# GitHub から最新のフォントをダウンロード
FONT_URL="https://github.com/Koruri/Koruri/raw/master/Koruri-Regular.ttf"
FONT_FILE="Koruri-Regular.ttf"
LICENSE_URL="https://raw.githubusercontent.com/Koruri/Koruri/master/LICENSE"
LICENSE_FILE="LICENSE.txt"

echo "Downloading from GitHub..."
echo "URL: $FONT_URL"
echo ""

if curl -L -f -o "$FONT_FILE" "$FONT_URL"; then
    echo "✓ Font downloaded: $FONT_FILE"
    ls -lh "$FONT_FILE"

    # ファイル形式の確認
    if command -v file &> /dev/null; then
        echo ""
        echo "File type:"
        file "$FONT_FILE"
    fi
else
    echo "✗ 自動ダウンロードに失敗しました"
    echo ""
    echo "手動でダウンロードしてください:"
    echo "1. https://github.com/Koruri/Koruri にアクセス"
    echo "2. Koruri-Regular.ttf をダウンロード"
    echo "3. このディレクトリに配置"
    echo ""
    exit 1
fi

echo ""
echo "Downloading license file..."
if curl -L -f -o "$LICENSE_FILE" "$LICENSE_URL"; then
    echo "✓ License downloaded: $LICENSE_FILE"
else
    echo "⚠️  License download failed (non-fatal)"
fi

echo ""
echo "==================================="
echo "Download complete!"
echo "==================================="
echo ""
echo "フォント: Koruri Regular"
echo "ライセンス: Apache License 2.0"
echo "商用利用: 可能"
echo "再配布: 可能（ライセンス文書の添付が必要）"
echo "埋め込み: 可能"
echo ""
echo "Next: go build でビルド時にフォントが埋め込まれます"
echo ""
