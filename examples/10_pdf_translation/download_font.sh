#!/bin/bash
# Noto Sans JP 静的フォントのダウンロードスクリプト

set -e

echo "Downloading Noto Sans JP (Static Font)..."

# 既存のフォントファイルを削除
if [ -f "NotoSansJP-Regular.ttf" ]; then
    echo "Removing existing font file..."
    rm -f NotoSansJP-Regular.ttf
fi

# 一時ディレクトリ
TMP_DIR=$(mktemp -d)
cd "$TMP_DIR"

# GitHub Releasesから静的フォントをダウンロード（直接ダウンロードURL）
echo "Downloading from GitHub (Noto CJK Sans)..."
curl -L "https://github.com/notofonts/noto-cjk/raw/main/Sans/OTF/Japanese/NotoSansCJKjp-Regular.otf" -o font.otf

# ダウンロードしたフォントをコピー
if [ -f "font.otf" ]; then
    TARGET_DIR="$OLDPWD"
    mv font.otf "$TARGET_DIR/NotoSansJP-Regular.ttf"
    cd "$TARGET_DIR"
    rm -rf "$TMP_DIR"

    # ファイルサイズを確認
    if [ -f "NotoSansJP-Regular.ttf" ]; then
        SIZE=$(du -h NotoSansJP-Regular.ttf | awk '{print $1}')
        SIZE_BYTES=$(du -k NotoSansJP-Regular.ttf | awk '{print $1}')
        echo "✓ Font downloaded successfully: NotoSansJP-Regular.ttf ($SIZE)"

        # 10MB以上なら警告
        if [ "$SIZE_BYTES" -gt 10240 ]; then
            echo "⚠ Warning: Font file is larger than 10MB. This might not work correctly."
            echo "   Expected size: ~3-6MB for static OTF font"
        fi
        exit 0
    fi
fi

echo "Failed with direct download, trying ZIP archive..."

# ZIPアーカイブからダウンロード
curl -L "https://github.com/google/fonts/raw/main/ofl/notosansjp/static/NotoSansJP-Regular.ttf" -o font2.ttf

if [ -f "font2.ttf" ] && [ -s "font2.ttf" ]; then
    TARGET_DIR="$OLDPWD"
    mv font2.ttf "$TARGET_DIR/NotoSansJP-Regular.ttf"
    cd "$TARGET_DIR"
    rm -rf "$TMP_DIR"

    SIZE=$(du -h NotoSansJP-Regular.ttf | awk '{print $1}')
    echo "✓ Font downloaded successfully: NotoSansJP-Regular.ttf ($SIZE)"
    exit 0
fi

echo "✗ All download methods failed. Please download manually."
echo "
Manual download instructions:
1. Go to: https://github.com/notofonts/noto-cjk/tree/main/Sans/OTF/Japanese
2. Download: NotoSansCJKjp-Regular.otf
3. Rename to: NotoSansJP-Regular.ttf
4. Place in: examples/10_pdf_translation/
"
exit 1
