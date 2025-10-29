package gopdf

import (
	"io"
	"os"
	"strings"

	"github.com/ryomak/gopdf/internal/content"
	"github.com/ryomak/gopdf/internal/reader"
)

// PDFReader はPDFを読み込むための構造体
type PDFReader struct {
	r      *reader.Reader
	closer io.Closer
}

// Open はファイルパスからPDFを開く
func Open(path string) (*PDFReader, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	r, err := reader.NewReader(file)
	if err != nil {
		file.Close()
		return nil, err
	}

	return &PDFReader{
		r:      r,
		closer: file,
	}, nil
}

// OpenReader はio.ReadSeekerからPDFを開く
func OpenReader(r io.ReadSeeker) (*PDFReader, error) {
	rd, err := reader.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &PDFReader{r: rd}, nil
}

// Close はリーダーをクローズする
func (r *PDFReader) Close() error {
	if r.closer != nil {
		return r.closer.Close()
	}
	return nil
}

// PageCount はページ数を返す
func (r *PDFReader) PageCount() int {
	count, _ := r.r.GetPageCount()
	return count
}

// Info はメタデータを返す
func (r *PDFReader) Info() Metadata {
	infoDict, err := r.r.GetInfo()
	if err != nil {
		return Metadata{}
	}

	return parseInfoDict(infoDict)
}

// TextElement はテキスト要素の位置とスタイル情報
type TextElement struct {
	Text   string  // テキスト内容
	X      float64 // X座標（左下原点）
	Y      float64 // Y座標（左下原点）
	Width  float64 // テキストの幅（概算）
	Height float64 // テキストの高さ（フォントサイズ）
	Font   string  // フォント名
	Size   float64 // フォントサイズ
}

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

// EncryptionInfo はPDF暗号化の情報
type EncryptionInfo struct {
	Filter  string // 暗号化フィルター（通常は "Standard"）
	V       int    // アルゴリズムバージョン（1 or 2）
	R       int    // リビジョン番号（2 or 3）
	Length  int    // 鍵長（ビット単位、40 or 128）
	P       int32  // パーミッションフラグ
	IsOwner bool   // オーナーとして認証されたか
}

// ExtractPageText は指定されたページのテキストを抽出する（0-indexed）
func (r *PDFReader) ExtractPageText(pageNum int) (string, error) {
	// ページを取得
	page, err := r.r.GetPage(pageNum)
	if err != nil {
		return "", err
	}

	// コンテンツストリームを取得
	contentsData, err := r.r.GetPageContents(page)
	if err != nil {
		return "", err
	}

	// コンテンツストリームをパース
	parser := content.NewStreamParser(contentsData)
	operations, err := parser.ParseOperations()
	if err != nil {
		return "", err
	}

	// テキストを抽出
	extractor := content.NewTextExtractor(operations)
	elements, err := extractor.Extract()
	if err != nil {
		return "", err
	}

	// テキスト要素を結合
	var texts []string
	for _, elem := range elements {
		texts = append(texts, elem.Text)
	}

	return strings.Join(texts, " "), nil
}

// ExtractText は全ページのテキストを抽出する
func (r *PDFReader) ExtractText() (string, error) {
	pageCount := r.PageCount()
	var allTexts []string

	for i := 0; i < pageCount; i++ {
		text, err := r.ExtractPageText(i)
		if err != nil {
			return "", err
		}
		allTexts = append(allTexts, text)
	}

	return strings.Join(allTexts, "\n\n"), nil
}

// ExtractPageTextElements は位置情報付きテキスト要素を抽出する（0-indexed）
func (r *PDFReader) ExtractPageTextElements(pageNum int) ([]TextElement, error) {
	// ページを取得
	page, err := r.r.GetPage(pageNum)
	if err != nil {
		return nil, err
	}

	// コンテンツストリームを取得
	contentsData, err := r.r.GetPageContents(page)
	if err != nil {
		return nil, err
	}

	// コンテンツストリームをパース
	parser := content.NewStreamParser(contentsData)
	operations, err := parser.ParseOperations()
	if err != nil {
		return nil, err
	}

	// テキストを抽出
	extractor := content.NewTextExtractor(operations)
	internalElements, err := extractor.Extract()
	if err != nil {
		return nil, err
	}

	// 内部型から公開型に変換
	elements := make([]TextElement, len(internalElements))
	for i, elem := range internalElements {
		elements[i] = TextElement{
			Text:   elem.Text,
			X:      elem.X,
			Y:      elem.Y,
			Width:  estimateTextWidth(elem.Text, elem.Size, elem.Font),
			Height: elem.Size,
			Font:   elem.Font,
			Size:   elem.Size,
		}
	}

	return elements, nil
}

// ExtractAllTextElements は全ページのテキスト要素を抽出する
func (r *PDFReader) ExtractAllTextElements() (map[int][]TextElement, error) {
	pageCount := r.PageCount()
	result := make(map[int][]TextElement)

	for i := 0; i < pageCount; i++ {
		elements, err := r.ExtractPageTextElements(i)
		if err != nil {
			return nil, err
		}
		result[i] = elements
	}

	return result, nil
}

// estimateTextWidth はテキストの幅を概算する
func estimateTextWidth(text string, fontSize float64, font string) float64 {
	// 簡易的な幅計算
	// 英数字の平均幅は fontSizeの約60%
	avgCharWidth := fontSize * 0.6
	return float64(len(text)) * avgCharWidth
}

// ExtractPageTextBlocks はテキストブロックを抽出する（0-indexed）
func (r *PDFReader) ExtractPageTextBlocks(pageNum int) ([]TextBlock, error) {
	elements, err := r.ExtractPageTextElements(pageNum)
	if err != nil {
		return nil, err
	}
	return r.groupTextElements(elements), nil
}

// ExtractAllTextBlocks は全ページのテキストブロックを抽出する
func (r *PDFReader) ExtractAllTextBlocks() (map[int][]TextBlock, error) {
	pageCount := r.PageCount()
	result := make(map[int][]TextBlock)

	for i := 0; i < pageCount; i++ {
		blocks, err := r.ExtractPageTextBlocks(i)
		if err != nil {
			return nil, err
		}
		result[i] = blocks
	}

	return result, nil
}

// ExtractImages は指定されたページから画像を抽出する（0-indexed）
func (r *PDFReader) ExtractImages(pageNum int) ([]ImageInfo, error) {
	// ページを取得
	page, err := r.r.GetPage(pageNum)
	if err != nil {
		return nil, err
	}

	// ImageExtractorを使用
	extractor := content.NewImageExtractor(r.r)
	internalImages, err := extractor.ExtractImages(page)
	if err != nil {
		return nil, err
	}

	// 内部型から公開型に変換
	images := make([]ImageInfo, len(internalImages))
	for i, img := range internalImages {
		images[i] = ImageInfo{
			Name:        img.Name,
			Width:       img.Width,
			Height:      img.Height,
			ColorSpace:  img.ColorSpace,
			BitsPerComp: img.BitsPerComp,
			Filter:      img.Filter,
			Data:        img.Data,
			Format:      ImageFormat(img.Format),
		}
	}

	return images, nil
}

// ExtractAllImages は全ページから画像を抽出する
func (r *PDFReader) ExtractAllImages() (map[int][]ImageInfo, error) {
	pageCount := r.PageCount()
	result := make(map[int][]ImageInfo)

	for i := 0; i < pageCount; i++ {
		images, err := r.ExtractImages(i)
		if err != nil {
			return nil, err
		}
		if len(images) > 0 {
			result[i] = images
		}
	}

	return result, nil
}

// IsEncrypted はPDFが暗号化されているかどうかを確認する
func (r *PDFReader) IsEncrypted() bool {
	return r.r.IsEncrypted()
}

// AuthenticateWithPassword はパスワードを使用してPDFを認証する
// 認証に成功すると、暗号化されたコンテンツを読み取れるようになる
func (r *PDFReader) AuthenticateWithPassword(password string) error {
	return r.r.AuthenticateWithPassword(password)
}

// GetEncryptionInfo は暗号化情報を取得する
// PDFが暗号化されていない場合はnilを返す
func (r *PDFReader) GetEncryptionInfo() *EncryptionInfo {
	internalInfo := r.r.GetEncryptionInfo()
	if internalInfo == nil {
		return nil
	}

	// 内部のEncryptionInfoから公開APIのEncryptionInfoに変換
	return &EncryptionInfo{
		Filter:  internalInfo.Filter,
		V:       internalInfo.V,
		R:       internalInfo.R,
		Length:  internalInfo.Length,
		P:       internalInfo.P,
		IsOwner: internalInfo.IsOwner,
	}
}
