package model

// ImageFormat は画像フォーマットを表す
type ImageFormat string

const (
	// ImageFormatJPEG はJPEG形式
	ImageFormatJPEG ImageFormat = "jpeg"
	// ImageFormatPNG はPNG形式
	ImageFormatPNG ImageFormat = "png"
	// ImageFormatUnknown は不明な形式
	ImageFormatUnknown ImageFormat = "unknown"
)

// ImageInfo は画像の情報を保持する
type ImageInfo struct {
	Name        string      // 画像名
	Width       int         // 幅（ピクセル）
	Height      int         // 高さ（ピクセル）
	ColorSpace  string      // 色空間（DeviceRGB, DeviceGray, etc.）
	BitsPerComp int         // ビット深度
	Filter      string      // 圧縮フィルター
	Data        []byte      // 画像データ
	Format      ImageFormat // 画像フォーマット
}
