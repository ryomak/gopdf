#!/bin/bash
# Noto Sans CJK JPフォントをダウンロードするスクリプト

set -e

# Google Fontsから直接ダウンロード
# 注意: Google FontsのAPIを使用するため、静的TTFファイルのURLが変更される可能性があります
# その場合は手動でダウンロードしてください

echo "==================================="
echo "Noto Sans JP Font Download Script"
echo "==================================="
echo ""
echo "このスクリプトはNoto Sans JP Regularフォントをダウンロードします。"
echo ""
echo "手動でダウンロードする場合:"
echo "1. https://fonts.google.com/noto/specimen/Noto+Sans+JP にアクセス"
echo "2. 'Get font' → 'Download all' をクリック"
echo "3. ZIPを解凍して static/NotoSansJP-Regular.ttf を取得"
echo "4. このディレクトリに NotoSansJP-Regular.ttf として配置"
echo ""
echo "フォントサイズ: 約1.5-4MB"
echo "ライセンス: SIL Open Font License 1.1 (商用利用可)"
echo ""

# GitHub releases から最新のフォントをダウンロード
# Noto Sans CJK JP は大きいので、Noto Sans JP（日本語専用版）を使用
FONT_URL="https://github.com/notofonts/noto-cjk/raw/main/Sans/SubsetOTF/JP/NotoSansCJKjp-Regular.otf"
FONT_FILE="NotoSansJP-Regular.otf"
LICENSE_URL="https://raw.githubusercontent.com/notofonts/noto-cjk/main/LICENSE"
LICENSE_FILE="LICENSE.txt"

echo "Attempting to download from GitHub..."
echo "URL: $FONT_URL"
echo ""

if curl -L -f -o "$FONT_FILE" "$FONT_URL" 2>/dev/null; then
    echo "✓ Font downloaded: $FONT_FILE"
    ls -lh "$FONT_FILE"

    # OTFをTTFに変換する必要がある場合の注意
    echo ""
    echo "⚠️  注意: ダウンロードしたフォントはOTF形式です。"
    echo "    gopdfではTTF形式が必要です。"
    echo "    fonttools等でTTFに変換してください:"
    echo "    $ pip install fonttools"
    echo "    $ ttx -f $FONT_FILE"
    echo ""
else
    echo "✗ 自動ダウンロードに失敗しました"
    echo ""
    echo "手動でダウンロードしてください:"
    echo "1. https://fonts.google.com/noto/specimen/Noto+Sans+JP"
    echo "2. ZIPをダウンロードして解凍"
    echo "3. static/NotoSansJP-Regular.ttf をこのディレクトリに配置"
    echo ""
    exit 1
fi

echo "Downloading license file..."
curl -L -o "$LICENSE_FILE" "$LICENSE_URL"

if [ -f "$LICENSE_FILE" ]; then
    echo "✓ License downloaded: $LICENSE_FILE"
else
    echo "✗ Failed to download license"
fi

echo ""
echo "==================================="
echo "Download complete!"
echo "==================================="
echo ""
echo "ライセンス: SIL Open Font License 1.1"
echo "商用利用: 可能"
echo "再配布: 可能（ライセンス表示が必要）"
echo ""
