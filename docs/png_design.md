# PNG画像対応 設計書

## 概要

Phase 5では、PNG画像のPDF埋め込み機能を実装します。PNGはPDFで直接サポートされていないため、デコードして再エンコードする必要があります。

## 要件

- PNG画像をPDFに埋め込むことができる
- 透明度（アルファチャンネル）を適切に処理できる
- RGB、RGBA、Grayscale、Paletteカラーをサポート
- JPEGと同じAPIで使用できる（`LoadPNG`, `LoadPNGFile`）
- 既存の `DrawImage` メソッドで描画できる

## PDF仕様とPNGの扱い

### PNGはPDFで直接サポートされていない

PDF仕様では、以下のフォーマットが直接サポートされています：
- JPEG (DCTDecode)
- JPEG 2000 (JPXDecode)

**PNG、GIF、TIFFは直接サポートされていません**。

### PNG埋め込みの2つのアプローチ

#### アプローチ1: PNG分解（複雑、高効率）
PNGファイルを分解して各要素をPDFオブジェクトにマッピング：
- IDAT (画像データ) → PDFバイトストリーム
- PLTE (パレット) → カラースペース定義
- iCCP (カラープロファイル) → ICCベースカラースペース
- tRNS (透明度) → SMaskストリーム

**メリット**: 最小サイズ、元のPNG圧縮を保持
**デメリット**: 実装が複雑、すべてのPNGチャンクに対応が必要

#### アプローチ2: デコード＆再エンコード（シンプル、実装容易）
Go標準ライブラリでPNGをデコードし、Raw pixel dataをFlateDecodeで圧縮：
1. `image/png` でPNGをデコード
2. Pixel dataを抽出
3. FlateDecode (Zlib) で圧縮
4. PDF Image XObjectとして埋め込み

**メリット**: 実装がシンプル、Go標準ライブラリを活用
**デメリット**: ファイルサイズが大きくなる可能性

### Phase 5での採用アプローチ

**アプローチ2（デコード＆再エンコード）を採用**

理由：
- Go標準ライブラリの活用
- 実装が明確でメンテナンスしやすい
- すべてのPNG形式に対応可能
- テストが容易

## PNG Image XObject構造

### 透明度なしPNG（RGB/Grayscale）

```
10 0 obj
<<
  /Type /XObject
  /Subtype /Image
  /Width 640
  /Height 480
  /ColorSpace /DeviceRGB
  /BitsPerComponent 8
  /Length 12345
  /Filter /FlateDecode
>>
stream
...compressed pixel data...
endstream
endobj
```

### 透明度ありPNG（RGBA）

透明度（アルファチャンネル）は、SMask（Soft Mask）として別のImage XObjectで表現：

```
10 0 obj  % メインイメージ
<<
  /Type /XObject
  /Subtype /Image
  /Width 640
  /Height 480
  /ColorSpace /DeviceRGB
  /BitsPerComponent 8
  /Length 12345
  /Filter /FlateDecode
  /SMask 11 0 R  % アルファチャンネルへの参照
>>
stream
...RGB data...
endstream
endobj

11 0 obj  % SMask（アルファチャンネル）
<<
  /Type /XObject
  /Subtype /Image
  /Width 640
  /Height 480
  /ColorSpace /DeviceGray
  /BitsPerComponent 8
  /Length 5678
  /Filter /FlateDecode
>>
stream
...alpha data...
endstream
endobj
```

### Indexed Color Space（パレット）

PNGのパレットカラーは、PDF Indexed Color Spaceにマッピング：

```
/ColorSpace [/Indexed /DeviceRGB 255 <hex-string-of-palette>]
```

## API設計

### LoadPNG関数

```go
// LoadPNG loads a PNG image from a reader
func LoadPNG(r io.Reader) (*Image, error)

// LoadPNGFile loads a PNG image from a file path
func LoadPNGFile(path string) (*Image, error)
```

### Image構造の拡張

```go
// Image represents an image that can be embedded in a PDF
type Image struct {
    Width            int
    Height           int
    Data             []byte
    ColorSpace       string
    BitsPerComponent int
    Filter           string      // "DCTDecode" or "FlateDecode"
    SMask            *Image      // アルファチャンネル（透明度）
    Palette          []byte      // パレットデータ（Indexed color用）
}
```

## 実装計画

### Phase 5.1: PNG解析とデコード

1. `internal/image/png` パッケージ作成
2. Go標準 `image/png` を使用してPNGデコード
3. 画像情報抽出（幅、高さ、カラーモデル）
4. テスト作成（TDD）

### Phase 5.2: Pixel Data抽出

1. RGBAデータの抽出
2. Grayscaleデータの抽出
3. Paletteデータの抽出
4. アルファチャンネルの分離
5. テスト作成（TDD）

### Phase 5.3: FlateEncode圧縮

1. Zlib/Deflate圧縮実装
2. Go標準 `compress/zlib` を使用
3. 圧縮レベル設定（デフォルト、最良、最速）
4. テスト作成（TDD）

### Phase 5.4: LoadPNG API実装

1. `LoadPNG` 関数実装
2. `LoadPNGFile` 関数実装
3. カラーモデル別の処理
   - RGBA → RGB + SMask
   - Gray, GrayAlpha
   - NRGBA, NRGBA64
   - Paletted
4. テスト作成（TDD）

### Phase 5.5: Document統合

1. SMask対応のImage XObject生成
2. Indexed Color Space対応
3. FlateDecode Filter対応
4. Documentでの画像リソース管理更新
5. テスト作成（TDD）

### Phase 5.6: サンプルとドキュメント

1. `examples/05_png_images` サンプル作成
2. 透明度ありPNGのサンプル
3. 統合テスト
4. README.md更新

## カラーモデル対応

### Go image.Imageのカラーモデル

| カラーモデル | 説明 | PDF変換 |
|------------|------|---------|
| `color.RGBA` | RGB + Alpha | RGB + SMask |
| `color.NRGBA` | Non-premultiplied RGBA | RGB + SMask |
| `color.Gray` | Grayscale | DeviceGray |
| `color.GrayAlpha` | Gray + Alpha | DeviceGray + SMask |
| `color.Palette` | Indexed color | Indexed ColorSpace |

### 変換処理

```go
switch img.ColorModel() {
case color.RGBAModel, color.NRGBA64Model:
    // RGB + アルファチャンネル分離
    rgbData, alphaData := separateRGBA(img)

case color.GrayModel:
    // Grayscale
    grayData := extractGray(img)

case color.Palette:
    // Indexed color
    indexedData, palette := extractPalette(img)
}
```

## 圧縮効率の考慮

### FlateDecode圧縮レベル

Go `compress/zlib` の圧縮レベル：
- `zlib.NoCompression` (0) - 圧縮なし
- `zlib.BestSpeed` (1) - 最速
- `zlib.DefaultCompression` (-1) - デフォルト
- `zlib.BestCompression` (9) - 最良圧縮

**採用**: `zlib.DefaultCompression` をデフォルトとする

### PNG vs JPEG選択のガイドライン

| 画像タイプ | 推奨フォーマット | 理由 |
|-----------|----------------|------|
| 写真 | JPEG | 高圧縮率、ファイルサイズ小 |
| スクリーンショット | PNG | テキストがシャープ、ロスレス |
| グラフ・図表 | PNG | エッジがクリーン、ロスレス |
| 透明度あり | PNG | JPEGは透明度非対応 |

## エラーハンドリング

以下のエラーケースに対応：
- ファイルが存在しない
- PNGフォーマットが不正
- サポートされていないカラーモデル
- メモリ不足（巨大画像）
- 圧縮失敗

## テスト戦略

### ユニットテスト

- PNG デコードのテスト
- Pixel data 抽出のテスト
- FlateDecode 圧縮のテスト
- LoadPNG API のテスト

### 統合テスト

- 実際のPNG画像を使用したPDF生成
- 透明度ありPNGの埋め込み
- 複数PNGの混在
- PDFリーダーでの表示確認

### テストデータ

以下のテスト画像を用意：
- RGB PNG（透明度なし）
- RGBA PNG（透明度あり）
- Grayscale PNG
- Indexed PNG（パレット）
- 半透明PNG（アルファチャンネル値が0-255）

## 制限事項（Phase 5）

以下は当フェーズでは対応しません：
- interlaced PNG（インターレース）の最適化
- アニメーションPNG（APNG）
- 16-bit PNG（8-bitに変換）
- iCCPカラープロファイル
- テキストチャンク（tEXt, iTXt, zTXt）

## 参考資料

- PDF 1.7 仕様書（ISO 32000-1:2008）
  - 7.4: Filters
  - 8.9.5: Image Dictionaries
  - 11.6.5: Soft-Mask Images
  - 8.6.6: Indexed Color Spaces
- PNG仕様（ISO/IEC 15948:2003）
- Go標準ライブラリ `image/png`
- Go標準ライブラリ `compress/zlib`
