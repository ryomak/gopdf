package content

import (
	"fmt"

	"github.com/ryomak/gopdf/internal/core"
	"github.com/ryomak/gopdf/internal/reader"
)

// ImageFormat は画像フォーマット
type ImageFormat string

const (
	ImageFormatJPEG    ImageFormat = "jpeg"
	ImageFormatPNG     ImageFormat = "png"
	ImageFormatUnknown ImageFormat = "unknown"
)

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
		// FlateDecode の場合はPNG相当
		return ImageFormatPNG
	default:
		return ImageFormatUnknown
	}
}
