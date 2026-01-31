package gopdf

import (
	"io"
	"math"
	"os"
	"sort"
	"strings"

	"github.com/ryomak/gopdf/internal/content"
	contentlayout "github.com/ryomak/gopdf/internal/content/layout"
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
	extractor := content.NewTextExtractor(operations, r.r, page)
	elements, err := extractor.Extract()
	if err != nil {
		return "", err
	}

	// テキスト要素を読み順にソートして結合
	return joinTextElements(elements), nil
}

// joinTextElements はテキスト要素を読み順に結合する
// 同じ行の要素は賢くスペースを挿入し、行が変わる場合は改行を挿入する
func joinTextElements(elements []content.TextElement) string {
	if len(elements) == 0 {
		return ""
	}

	// Y座標でグループ化（行ごと）
	lines := groupContentElementsByLine(elements)

	// 各行をY座標の降順（上から下）でソート
	sort.Slice(lines, func(i, j int) bool {
		return lines[i][0].Y > lines[j][0].Y
	})

	var result strings.Builder
	for i, line := range lines {
		// 行内をX座標でソート
		sort.Slice(line, func(a, b int) bool {
			return line[a].X < line[b].X
		})

		// 行内のテキストを結合
		lineText := joinContentLineElements(line)
		result.WriteString(lineText)

		// 最後の行でなければ改行を追加
		if i < len(lines)-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}

// groupContentElementsByLine はテキスト要素をY座標で行ごとにグループ化する
func groupContentElementsByLine(elements []content.TextElement) [][]content.TextElement {
	if len(elements) == 0 {
		return nil
	}

	// Y座標でソート
	sorted := make([]content.TextElement, len(elements))
	copy(sorted, elements)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Y > sorted[j].Y
	})

	var lines [][]content.TextElement
	currentLine := []content.TextElement{sorted[0]}
	currentY := sorted[0].Y
	// しきい値: フォントサイズの50%以内なら同じ行とみなす
	threshold := sorted[0].Size * 0.5
	if threshold < 1 {
		threshold = 3 // 最小しきい値
	}

	for i := 1; i < len(sorted); i++ {
		elem := sorted[i]
		if math.Abs(elem.Y-currentY) <= threshold {
			currentLine = append(currentLine, elem)
		} else {
			lines = append(lines, currentLine)
			currentLine = []content.TextElement{elem}
			currentY = elem.Y
			threshold = elem.Size * 0.5
			if threshold < 1 {
				threshold = 3
			}
		}
	}
	lines = append(lines, currentLine)

	return lines
}

// joinContentLineElements は同じ行のテキスト要素を結合する
// X座標の距離に基づいてスペースを挿入するかどうかを判断する
func joinContentLineElements(elements []content.TextElement) string {
	if len(elements) == 0 {
		return ""
	}

	var result strings.Builder
	var lastEndX float64
	var lastSize float64 = 10 // デフォルトサイズ

	for i, elem := range elements {
		text := elem.Text
		if text == "" {
			continue
		}

		if i > 0 && result.Len() > 0 {
			// 現在の要素との間隔を計算
			gap := elem.X - lastEndX

			// スペース幅の目安（フォントサイズの約40%）
			// PDFでは文字間調整で微妙な間隔が生じるため、余裕を持たせる
			// 典型的なスペース幅はフォントサイズの約25-30%だが、
			// カーニングなどで間隔が生じることがあるため、少し大きめに設定
			spaceWidth := lastSize * 0.4
			if spaceWidth < 2.5 {
				spaceWidth = 2.5
			}

			// 間隔がスペース幅より大きければスペースを挿入
			// ただし、前の文字または現在の文字がスペースで終わる/始まる場合は挿入しない
			resultStr := result.String()
			endsWithSpace := len(resultStr) > 0 && resultStr[len(resultStr)-1] == ' '
			startsWithSpace := len(text) > 0 && text[0] == ' '

			if gap > spaceWidth && !endsWithSpace && !startsWithSpace {
				result.WriteString(" ")
			}
		}

		result.WriteString(text)

		// 終端位置を更新（文字幅を概算）
		// フォントサイズの約60%を1文字の平均幅とする
		avgCharWidth := elem.Size * 0.6
		if avgCharWidth < 1 {
			avgCharWidth = 5 // デフォルト幅
		}
		lastEndX = elem.X + float64(len(text))*avgCharWidth
		lastSize = elem.Size
		if lastSize < 1 {
			lastSize = 10
		}
	}

	return result.String()
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
	extractor := content.NewTextExtractor(operations, r.r, page)
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
	return contentlayout.GroupTextElements(elements), nil
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

// ExtractPageContentBlocks はテキストと画像を統合したコンテンツブロックを抽出（0-indexed）
// 設計書: docs/unified_content_grouping_design.md
func (r *PDFReader) ExtractPageContentBlocks(pageNum int) ([]ContentBlock, error) {
	// PageLayoutを取得（テキストと画像の両方）
	pageLayout, err := r.ExtractPageLayout(pageNum)
	if err != nil {
		return nil, err
	}

	// Y座標順にソートして返す
	return pageLayout.SortedContentBlocks(), nil
}

// ExtractAllContentBlocks は全ページのコンテンツブロックを抽出
func (r *PDFReader) ExtractAllContentBlocks() (map[int][]ContentBlock, error) {
	pageCount := r.PageCount()
	result := make(map[int][]ContentBlock)

	for i := 0; i < pageCount; i++ {
		blocks, err := r.ExtractPageContentBlocks(i)
		if err != nil {
			return nil, err
		}
		result[i] = blocks
	}

	return result, nil
}

// ExtractAllContentBlocksFlattened はページを跨いでブロックを統合して返す
// mergeAcrossPagesがtrueの場合、連続するテキストブロックでフォント属性が同じなら統合される
// falseの場合、ページ境界を保持したまま単純にフラット化される
// 設計書: docs/cross_page_block_merging_design.md
func (r *PDFReader) ExtractAllContentBlocksFlattened(mergeAcrossPages bool) ([]ContentBlock, error) {
	// 全ページのブロックを取得
	pageBlocks, err := r.ExtractAllContentBlocks()
	if err != nil {
		return nil, err
	}

	if !mergeAcrossPages {
		// ページ境界を保持したまま単純にフラット化
		return flattenContentBlocks(pageBlocks), nil
	}

	// ページを跨いで統合
	return mergeContentBlocksAcrossPages(pageBlocks), nil
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
