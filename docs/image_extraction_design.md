# Phase 9: 画像抽出 設計書

## 1. 概要

PDFファイルから埋め込まれた画像を抽出する機能を実装する。画像リソースを解析し、画像データを取得してファイルとして保存可能にする。

### 1.1. 目的

- PDFに埋め込まれた画像の抽出
- 画像の位置・サイズ情報の取得
- 画像フォーマットの保持（JPEG, PNG）
- 画像データへのアクセスAPI

### 1.2. スコープ

**Phase 9で実装する機能:**
- ✅ ページリソースからの画像XObject解析
- ✅ JPEG画像の抽出（DCTDecode）
- ✅ PNG画像の抽出（FlateDecode）
- ✅ 画像メタデータ（幅、高さ、色空間）
- ✅ 公開API

**Phase 9では実装しない機能:**
- 画像の配置位置の正確な計算（コンテンツストリームからの抽出）
- CCITT FAX、JBIG2などの他の画像フォーマット
- アルファチャンネル（SMask）の処理
- インラインイメージ（BI/ID/EI）

## 2. PDFの画像リソース

### 2.1. 画像XObject

PDFの画像は XObject として格納される：

```
4 0 obj
<<
  /Type /XObject
  /Subtype /Image
  /Width 1024
  /Height 768
  /ColorSpace /DeviceRGB
  /BitsPerComponent 8
  /Filter /DCTDecode
  /Length 45678
>>
stream
...JPEG data...
endstream
endobj
```

### 2.2. ページリソース

ページの画像はResourcesの/XObjectで参照される：

```
5 0 obj  % Page object
<<
  /Type /Page
  /Resources <<
    /XObject <<
      /Im1 4 0 R    % 画像XObjectへの参照
      /Im2 7 0 R
    >>
  >>
  /Contents 6 0 R
>>
endobj
```

### 2.3. 画像フィルター

| フィルター | 説明 | 対応 |
|-----------|-----|------|
| DCTDecode | JPEG圧縮 | ✅ Phase 9 |
| FlateDecode | PNG/Zlib圧縮 | ✅ Phase 9 |
| CCITTFaxDecode | FAX圧縮 | ❌ 将来 |
| JBIG2Decode | JBIG2圧縮 | ❌ 将来 |
| JPXDecode | JPEG2000 | ❌ 将来 |

## 3. 設計

### 3.1. データ構造

```go
package gopdf

// ImageInfo は画像の情報
type ImageInfo struct {
	Name        string      // リソース名（例: "Im1"）
	Width       int         // 画像の幅
	Height      int         // 画像の高さ
	ColorSpace  string      // 色空間（DeviceRGB, DeviceGray, DeviceCMYK）
	BitsPerComp int         // ビット深度
	Filter      string      // 圧縮フィルター
	Data        []byte      // 画像データ
	Format      ImageFormat // 画像フォーマット
}

// ImageFormat は画像フォーマット
type ImageFormat string

const (
	ImageFormatJPEG    ImageFormat = "jpeg"
	ImageFormatPNG     ImageFormat = "png"
	ImageFormatUnknown ImageFormat = "unknown"
)
```

### 3.2. 公開API

```go
package gopdf

// ExtractImages はページから画像を抽出する
func (r *PDFReader) ExtractImages(pageNum int) ([]ImageInfo, error)

// ExtractAllImages は全ページから画像を抽出する
func (r *PDFReader) ExtractAllImages() (map[int][]ImageInfo, error)

// SaveImage は画像をファイルに保存する
func (img *ImageInfo) SaveImage(filename string) error

// ToImage は画像をimage.Imageに変換する
func (img *ImageInfo) ToImage() (image.Image, error)
```

### 3.3. 実装アーキテクチャ

```
┌─────────────────────────────────────┐
│  gopdf.PDFReader.ExtractImages()    │  ← 公開API
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│  internal/content/image_extractor.go│  ← 画像抽出
│  - ImageExtractor                    │
│  - ExtractImages(page) → []ImageInfo│
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│  internal/reader/reader.go          │  ← Reader拡張
│  - GetPageResources(page)           │
│  - GetImageXObject(ref)             │
└─────────────────────────────────────┘
```

## 4. 実装の詳細

### 4.1. ページリソースの取得

```go
// Reader拡張: ページのResourcesを取得
func (r *Reader) GetPageResources(page core.Dictionary) (core.Dictionary, error) {
	resourcesObj, ok := page[core.Name("Resources")]
	if !ok {
		return nil, nil // Resourcesがない場合
	}

	// Referenceの場合は解決
	if ref, ok := resourcesObj.(*core.Reference); ok {
		obj, err := r.GetObject(ref.ObjectNumber)
		if err != nil {
			return nil, err
		}
		resourcesObj = obj
	}

	resources, ok := resourcesObj.(core.Dictionary)
	if !ok {
		return nil, fmt.Errorf("resources is not a dictionary")
	}

	return resources, nil
}
```

### 4.2. 画像XObjectの取得

```go
// GetImageXObject は画像XObjectを取得する
func (r *Reader) GetImageXObject(ref *core.Reference) (*ImageXObject, error) {
	obj, err := r.GetObject(ref.ObjectNumber)
	if err != nil {
		return nil, err
	}

	stream, ok := obj.(*core.Stream)
	if !ok {
		return nil, fmt.Errorf("image xobject is not a stream")
	}

	// /Typeと/Subtypeの確認
	subtype, _ := stream.Dict[core.Name("Subtype")].(core.Name)
	if subtype != "Image" {
		return nil, fmt.Errorf("not an image xobject")
	}

	// 画像情報を抽出
	img := &ImageXObject{
		Stream: stream,
	}

	// Width
	if w, ok := stream.Dict[core.Name("Width")].(core.Integer); ok {
		img.Width = int(w)
	}

	// Height
	if h, ok := stream.Dict[core.Name("Height")].(core.Integer); ok {
		img.Height = int(h)
	}

	// ColorSpace
	if cs, ok := stream.Dict[core.Name("ColorSpace")].(core.Name); ok {
		img.ColorSpace = string(cs)
	}

	// BitsPerComponent
	if bpc, ok := stream.Dict[core.Name("BitsPerComponent")].(core.Integer); ok {
		img.BitsPerComponent = int(bpc)
	}

	// Filter
	if filter, ok := stream.Dict[core.Name("Filter")].(core.Name); ok {
		img.Filter = string(filter)
	}

	return img, nil
}

type ImageXObject struct {
	Stream           *core.Stream
	Width            int
	Height           int
	ColorSpace       string
	BitsPerComponent int
	Filter           string
}
```

### 4.3. 画像抽出ロジック

```go
package content

import (
	"github.com/ryomak/gopdf/internal/core"
	"github.com/ryomak/gopdf/internal/reader"
)

// ImageExtractor は画像を抽出する
type ImageExtractor struct {
	reader *reader.Reader
}

// NewImageExtractor は新しいImageExtractorを作成する
func NewImageExtractor(r *reader.Reader) *ImageExtractor {
	return &ImageExtractor{reader: r}
}

// ExtractImages はページから画像を抽出する
func (e *ImageExtractor) ExtractImages(page core.Dictionary) ([]ImageInfo, error) {
	// Resourcesを取得
	resources, err := e.reader.GetPageResources(page)
	if err != nil || resources == nil {
		return nil, err
	}

	// /XObjectを取得
	xobjectsObj, ok := resources[core.Name("XObject")]
	if !ok {
		return nil, nil // XObjectがない
	}

	xobjects, ok := xobjectsObj.(core.Dictionary)
	if !ok {
		return nil, fmt.Errorf("xobjects is not a dictionary")
	}

	var images []ImageInfo

	// 各XObjectを処理
	for name, obj := range xobjects {
		ref, ok := obj.(*core.Reference)
		if !ok {
			continue
		}

		// 画像XObjectを取得
		imgXObj, err := e.reader.GetImageXObject(ref)
		if err != nil {
			continue // 画像でない場合はスキップ
		}

		// ImageInfoに変換
		info := ImageInfo{
			Name:        string(name),
			Width:       imgXObj.Width,
			Height:      imgXObj.Height,
			ColorSpace:  imgXObj.ColorSpace,
			BitsPerComp: imgXObj.BitsPerComponent,
			Filter:      imgXObj.Filter,
			Data:        imgXObj.Stream.Data,
		}

		// フォーマットを判定
		info.Format = detectImageFormat(imgXObj.Filter, info.Data)

		images = append(images, info)
	}

	return images, nil
}

// detectImageFormat は画像フォーマットを判定する
func detectImageFormat(filter string, data []byte) ImageFormat {
	switch filter {
	case "DCTDecode":
		return ImageFormatJPEG
	case "FlateDecode":
		// PNGかどうか確認（簡易的）
		return ImageFormatPNG
	default:
		return ImageFormatUnknown
	}
}

type ImageInfo struct {
	Name        string
	Width       int
	Height      int
	ColorSpace  string
	BitsPerComp int
	Filter      string
	Data        []byte
	Format      ImageFormat
}

type ImageFormat string

const (
	ImageFormatJPEG    ImageFormat = "jpeg"
	ImageFormatPNG     ImageFormat = "png"
	ImageFormatUnknown ImageFormat = "unknown"
)
```

### 4.4. 画像保存機能

```go
package gopdf

import (
	"fmt"
	"os"
)

// SaveImage は画像をファイルに保存する
func (img *ImageInfo) SaveImage(filename string) error {
	// ファイルを作成
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// データを書き込み
	_, err = file.Write(img.Data)
	if err != nil {
		return fmt.Errorf("failed to write image data: %w", err)
	}

	return nil
}

// ToImage は画像をimage.Imageに変換する
func (img *ImageInfo) ToImage() (image.Image, error) {
	switch img.Format {
	case ImageFormatJPEG:
		return jpeg.Decode(bytes.NewReader(img.Data))
	case ImageFormatPNG:
		// FlateDecode画像をPNGに再エンコード
		// または直接デコード
		return decodeFlateImage(img)
	default:
		return nil, fmt.Errorf("unsupported image format: %s", img.Format)
	}
}
```

## 5. テスト計画

### 5.1. ユニットテスト

```go
func TestExtractImages(t *testing.T) {
	// Writerで画像入りPDFを生成
	doc := gopdf.New()
	page := doc.AddPage(gopdf.A4, gopdf.Portrait)

	// JPEG画像を追加
	jpegImg, _ := os.Open("test.jpg")
	page.DrawImage(jpegImg, 100, 700, 200, 150)

	var buf bytes.Buffer
	doc.WriteTo(&buf)

	// 読み込み
	reader, _ := gopdf.OpenReader(bytes.NewReader(buf.Bytes()))

	// 画像抽出
	images, err := reader.ExtractImages(0)
	if err != nil {
		t.Fatalf("ExtractImages failed: %v", err)
	}

	if len(images) != 1 {
		t.Fatalf("Expected 1 image, got %d", len(images))
	}

	img := images[0]
	if img.Format != gopdf.ImageFormatJPEG {
		t.Errorf("Format = %s, want JPEG", img.Format)
	}
}
```

### 5.2. サンプル

`examples/08_extract_images/main.go`:

```go
package main

import (
	"fmt"
	"log"
	"github.com/ryomak/gopdf"
)

func main() {
	reader, err := gopdf.Open("document.pdf")
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	// 画像を抽出
	images, err := reader.ExtractImages(0)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d images\n", len(images))

	// 各画像を保存
	for i, img := range images {
		filename := fmt.Sprintf("image_%d.%s", i+1, img.Format)
		if err := img.SaveImage(filename); err != nil {
			log.Printf("Failed to save image %d: %v", i+1, err)
			continue
		}
		fmt.Printf("Saved: %s (%dx%d, %s)\n",
			filename, img.Width, img.Height, img.ColorSpace)
	}
}
```

## 6. 注意事項

### 6.1. 画像フォーマット

- **JPEG (DCTDecode)**: データはそのままJPEGファイルとして保存可能
- **PNG (FlateDecode)**: 展開後のrawデータ。PNG形式への再エンコードが必要
- 他のフォーマットは現在未対応

### 6.2. 色空間

以下の色空間に対応：
- DeviceRGB
- DeviceGray
- DeviceCMYK（部分対応）

カスタム色空間（ICCベース、パターンなど）は未対応。

### 6.3. 画像の位置

現在の実装ではリソースから画像を抽出するのみ。
画像の配置位置を取得するには、コンテンツストリームの`Do`オペレーターを解析する必要がある（将来の拡張）。

## 7. 参考資料

- [PDF 1.7 仕様書](https://opensource.adobe.com/dc-acrobat-sdk-docs/pdfstandards/PDF32000_2008.pdf)
  - Section 8.9: Images
  - Section 8.9.5: Image Dictionaries
  - Section 7.8: Content Streams and Resources
