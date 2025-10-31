package content

import (
	"github.com/ryomak/gopdf/internal/core"
	"github.com/ryomak/gopdf/internal/reader"
	"github.com/ryomak/gopdf/internal/utils"
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

// ImageBlock は画像の配置情報（位置情報付き）
type ImageBlock struct {
	ImageInfo            // 画像情報
	X         float64    // 配置X座標
	Y         float64    // 配置Y座標
	PlacedWidth  float64 // 配置された幅
	PlacedHeight float64 // 配置された高さ
	Transform Matrix     // 変換行列
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

	xobjects, err := utils.MustExtractAs[core.Dictionary](xobjectsObj, "xobjects")
	if err != nil {
		return nil, err
	}

	var images []ImageInfo

	// 各XObjectを処理
	for name, obj := range xobjects {
		ref, ok := utils.ExtractAs[*core.Reference](obj)
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

// ExtractImagesWithPosition は位置情報付きで画像を抽出する
func (e *ImageExtractor) ExtractImagesWithPosition(page core.Dictionary, operations []Operation) ([]ImageBlock, error) {
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

	xobjects, err := utils.MustExtractAs[core.Dictionary](xobjectsObj, "xobjects")
	if err != nil {
		return nil, err
	}

	// グラフィックス状態スタック
	gsStack := []GraphicsState{NewGraphicsState()}

	var images []ImageBlock

	// コンテンツストリームを解析
	for _, op := range operations {
		switch op.Operator {
		case "cm": // 変換行列の変更
			if len(op.Operands) == 6 {
				a := toFloat64(op.Operands[0])
				b := toFloat64(op.Operands[1])
				c := toFloat64(op.Operands[2])
				d := toFloat64(op.Operands[3])
				e := toFloat64(op.Operands[4])
				f := toFloat64(op.Operands[5])

				matrix := Matrix{A: a, B: b, C: c, D: d, E: e, F: f}
				currentGS := &gsStack[len(gsStack)-1]
				currentGS.CTM = currentGS.CTM.Multiply(matrix)
			}

		case "Do": // XObjectの描画
			if len(op.Operands) == 1 {
				name, ok := utils.ExtractAs[core.Name](op.Operands[0])
				if !ok {
					continue
				}

				// 画像XObjectを取得
				xobjRef, ok := utils.ExtractAs[*core.Reference](xobjects[name])
				if !ok {
					continue
				}

				imgXObj, err := e.reader.GetImageXObject(xobjRef)
				if err != nil {
					continue // 画像でない場合はスキップ
				}

				// 現在のCTMを取得
				currentCTM := gsStack[len(gsStack)-1].CTM

				// 画像のデフォルトサイズは1x1の単位正方形
				// CTMで実際の位置とサイズが決まる

				// 1x1の単位正方形の4隅を変換して、実際の座標とサイズを計算
				minX, minY, maxX, maxY := currentCTM.TransformRect(0, 0, 1, 1)

				// 画像の左下座標（PDF座標系）
				x := minX
				y := minY

				// 画像のサイズ
				width := maxX - minX
				height := maxY - minY

				// 異常な座標値の検出
				// PDFページサイズの妥当な範囲を超える場合は異常と判定
				// 一般的なPDFページサイズは最大でも数千ポイント程度
				const maxReasonableCoordinate = 10000.0
				isAbnormal := false

				// CTM自体の異常値チェック
				if currentCTM.E < -maxReasonableCoordinate || currentCTM.E > maxReasonableCoordinate ||
					currentCTM.F < -maxReasonableCoordinate || currentCTM.F > maxReasonableCoordinate {
					isAbnormal = true
				}

				// 計算された座標の異常値チェック
				if x < -maxReasonableCoordinate || x > maxReasonableCoordinate ||
					y < -maxReasonableCoordinate || y > maxReasonableCoordinate {
					isAbnormal = true
				}

				// 異常な座標の場合はスキップ
				// 注: 異常な座標の画像は元PDFの問題であり、正しく表示できない
				if isAbnormal {
					// デバッグ情報として記録（将来的にはloggerを使用）
					// TODO: ロガーを使用する
					// fmt.Printf("Warning: Abnormal image coordinates detected: X=%.2f, Y=%.2f, CTM.F=%.2f\n", x, y, currentCTM.F)

					// 一旦、異常な座標でも画像を含める（デバッグ用）
					// continue // この画像をスキップ
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
				info.Format = detectImageFormat(imgXObj.Filter, info.Data)

				images = append(images, ImageBlock{
					ImageInfo:    info,
					X:            x,
					Y:            y,
					PlacedWidth:  width,
					PlacedHeight: height,
					Transform:    currentCTM,
				})
			}

		case "q": // グラフィックス状態の保存
			gsStack = append(gsStack, gsStack[len(gsStack)-1].Clone())

		case "Q": // グラフィックス状態の復元
			if len(gsStack) > 1 {
				gsStack = gsStack[:len(gsStack)-1]
			}
		}
	}

	return images, nil
}

// toFloat64 はcore.Objectをfloat64に変換する
func toFloat64(obj core.Object) float64 {
	switch v := obj.(type) {
	case core.Integer:
		return float64(v)
	case core.Real:
		return float64(v)
	default:
		return 0
	}
}
