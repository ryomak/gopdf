# 画像埋め込み機能 設計書

## 概要

Phase 4では、PDFへの画像埋め込み機能を実装します。JPEG画像を中心に、PNG画像のサポートも検討します。

## 要件

- JPEG画像をPDFに埋め込むことができる
- 画像のサイズ（幅・高さ）を指定できる
- 画像の位置（X, Y座標）を指定できる
- 複数の画像を同一ページに配置できる
- 画像の縦横比を維持できる

## PDF仕様

### Image XObject

PDFにおける画像は、Image XObjectとして表現されます（PDF 1.7仕様 8.9.5節）。

#### JPEG画像の基本構造

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
  /Filter /DCTDecode
>>
stream
...JPEG binary data...
endstream
endobj
```

#### 必須フィールド

| フィールド | 説明 | 例 |
|-----------|------|-----|
| `/Type` | オブジェクトタイプ | `/XObject` |
| `/Subtype` | サブタイプ | `/Image` |
| `/Width` | 画像幅（ピクセル） | `640` |
| `/Height` | 画像高さ（ピクセル） | `480` |
| `/ColorSpace` | 色空間 | `/DeviceRGB`, `/DeviceGray`, `/DeviceCMYK` |
| `/BitsPerComponent` | ピクセルあたりのビット数 | `8` |
| `/Filter` | 圧縮フィルター | `/DCTDecode` (JPEG) |

### ページコンテンツでの画像表示

画像をページに表示するには、以下のオペレーターを使用します：

```
q                           % グラフィックス状態を保存
width 0 0 height x y cm    % 変換行列（位置とサイズ）
/Im1 Do                     % 画像を描画
Q                           % グラフィックス状態を復元
```

#### 変換行列の説明

`a b c d e f cm` の形式で指定します：
- `a`: X方向のスケール（幅）
- `b`: X方向の歪み（通常は0）
- `c`: Y方向の歪み（通常は0）
- `d`: Y方向のスケール（高さ）
- `e`: X座標（位置）
- `f`: Y座標（位置）

例：幅200, 高さ150の画像を (100, 500) に配置
```
q
200 0 0 150 100 500 cm
/Im1 Do
Q
```

### カラースペース

JPEGは以下のカラースペースをサポート：
- `/DeviceGray` - グレースケール（1チャンネル）
- `/DeviceRGB` - RGB（3チャンネル）
- `/DeviceCMYK` - CMYK（4チャンネル）

## API設計

### 画像の読み込みと管理

```go
// Image represents an image that can be embedded in a PDF
type Image struct {
    Width  int
    Height int
    Data   []byte
    ColorSpace ColorSpace
    BitsPerComponent int
}

// ColorSpace represents the color space of an image
type ColorSpace string

const (
    ColorSpaceGray ColorSpace = "DeviceGray"
    ColorSpaceRGB  ColorSpace = "DeviceRGB"
    ColorSpaceCMYK ColorSpace = "DeviceCMYK"
)

// LoadJPEG loads a JPEG image from a reader
func LoadJPEG(r io.Reader) (*Image, error)

// LoadJPEGFile loads a JPEG image from a file path
func LoadJPEGFile(path string) (*Image, error)
```

### Pageへのメソッド追加

```go
// DrawImage draws an image at the specified position with the specified size
func (p *Page) DrawImage(img *Image, x, y, width, height float64) error

// DrawImageAspectFit draws an image maintaining aspect ratio within the given bounds
func (p *Page) DrawImageAspectFit(img *Image, x, y, maxWidth, maxHeight float64) error
```

### Documentレベルでの画像管理

同じ画像が複数回使用される場合、画像リソースを再利用する必要があります。

```go
type Document struct {
    pages  []*Page
    images map[string]*imageResource  // 画像の再利用管理
}

type imageResource struct {
    image    *Image
    objNum   int  // PDFオブジェクト番号
    resourceName string  // /Im1, /Im2, etc.
}
```

## JPEG画像の解析

JPEGファイルの構造を解析して、以下の情報を抽出する必要があります：

### JPEG構造

```
FF D8           - SOI (Start of Image)
FF E0 ... FF    - JFIF/EXIF header (optional)
FF C0 ...       - SOF (Start of Frame) - 画像情報を含む
  - 高さ
  - 幅
  - カラーコンポーネント数
  - ビット深度
FF DA ...       - SOS (Start of Scan) - 画像データ開始
... data ...
FF D9           - EOI (End of Image)
```

### SOF (Start of Frame) マーカー

```
FF C0           - マーカー
00 11           - セグメント長
08              - ビット深度 (通常8)
01 E0           - 高さ (480)
02 80           - 幅 (640)
03              - コンポーネント数 (3=RGB, 1=Gray, 4=CMYK)
```

## 実装計画

### Phase 4.1: JPEG解析パッケージ

1. `internal/image/jpeg` パッケージ作成
2. JPEG SOFマーカー解析実装
3. 画像情報抽出（幅、高さ、色空間）
4. テスト作成（TDD）

### Phase 4.2: Image構造とAPI

1. `Image`型の定義
2. `LoadJPEG`関数実装
3. `LoadJPEGFile`関数実装
4. テスト作成（TDD）

### Phase 4.3: 画像埋め込み実装

1. Documentへの画像リソース管理追加
2. `Page.DrawImage`メソッド実装
3. `Page.DrawImageAspectFit`メソッド実装（縦横比維持）
4. PDF出力時の画像XObject生成
5. テスト作成（TDD）

### Phase 4.4: PNG対応（オプション）

PNGの場合、以下の選択肢があります：
1. PNG → JPEG変換して埋め込み（画質劣化）
2. PNG → Raw pixel data + FlateDecode（ファイルサイズ増大）
3. Go標準の `image/png` で解凍してから埋め込み

Phase 4では、まずJPEGに注力し、PNGは将来的な拡張として検討します。

### Phase 4.5: サンプルとドキュメント

1. `examples/04_images` サンプル作成
2. テスト用のサンプル画像準備
3. README.md更新
4. 統合テスト

## エラーハンドリング

以下のエラーケースに対応：
- ファイルが存在しない
- JPEGフォーマットが不正
- サポートされていないJPEGタイプ（プログレッシブJPEGなど）
- 画像サイズが大きすぎる（メモリ制限）

## テスト戦略

### ユニットテスト

- JPEG解析のテスト（各マーカーの解析）
- Image構造のテスト
- DrawImageメソッドのテスト（コンテンツストリーム生成）

### 統合テスト

- 実際のJPEG画像を使用したPDF生成
- 複数画像の埋め込み
- PDFリーダーでの表示確認

### テストデータ

以下のテスト画像を用意：
- グレースケールJPEG（小サイズ、例: 100x100）
- RGB JPEG（中サイズ、例: 640x480）
- 縦長・横長の画像

## 参考資料

- PDF 1.7 仕様書（ISO 32000-1:2008）
  - 7.4: Filters
  - 8.9.5: Image Dictionaries
- JPEG仕様（ITU-T T.81）
- Go標準ライブラリ `image/jpeg`

## 制限事項（Phase 4）

以下は当フェーズでは対応しません：
- PNG画像（将来的に対応予定）
- プログレッシブJPEG
- JPEG 2000
- 画像の回転・反転
- 画像のマスク・透明度
- 画像の圧縮レベル調整
