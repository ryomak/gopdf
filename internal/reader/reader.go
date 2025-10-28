package reader

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/ryomak/gopdf/internal/core"
	"github.com/ryomak/gopdf/internal/utils"
)

// xrefEntry はクロスリファレンステーブルのエントリ
type xrefEntry struct {
	offset     int64 // ファイル内バイトオフセット
	generation int   // 世代番号
	inUse      bool  // 使用中かどうか
}

// Reader はPDFファイルを読み込み、解析する
type Reader struct {
	r          io.ReadSeeker       // ファイルのシーク可能なリーダー
	xref       map[int]xrefEntry   // オブジェクト番号 -> xrefエントリ
	trailer    core.Dictionary     // Trailer辞書
	objCache   map[int]core.Object // オブジェクトキャッシュ
	encryption *EncryptionInfo     // 暗号化情報（nil = 暗号化なし）
}

// NewReader は新しいReaderを作成する
func NewReader(r io.ReadSeeker) (*Reader, error) {
	reader := &Reader{
		r:        r,
		xref:     make(map[int]xrefEntry),
		objCache: make(map[int]core.Object),
	}

	// ファイルの解析
	if err := reader.parse(); err != nil {
		return nil, err
	}

	return reader, nil
}

// parse はPDFファイルを解析する
func (r *Reader) parse() error {
	// startxrefのオフセットを取得
	xrefOffset, err := r.findStartXref()
	if err != nil {
		return fmt.Errorf("failed to find startxref: %w", err)
	}

	// xrefテーブルとtrailerを解析
	if err := r.parseXrefAndTrailer(xrefOffset); err != nil {
		return fmt.Errorf("failed to parse xref and trailer: %w", err)
	}

	// 暗号化情報を検出
	if err := r.detectEncryption(); err != nil {
		return fmt.Errorf("failed to detect encryption: %w", err)
	}

	return nil
}

// detectEncryption はPDFの暗号化情報を検出する
func (r *Reader) detectEncryption() error {
	// Encrypt エントリをチェック
	encryptRef, hasEncrypt := r.trailer[core.Name("Encrypt")]
	if !hasEncrypt {
		// 暗号化されていない
		return nil
	}

	// Encrypt辞書を取得
	var encryptDict core.Dictionary
	if ref, ok := encryptRef.(*core.Reference); ok {
		// 参照の場合はオブジェクトを解決
		obj, err := r.ResolveReference(ref)
		if err != nil {
			return fmt.Errorf("failed to resolve Encrypt reference: %w", err)
		}
		var ok bool
		encryptDict, ok = obj.(core.Dictionary)
		if !ok {
			return fmt.Errorf("Encrypt is not a dictionary")
		}
	} else if dict, ok := encryptRef.(core.Dictionary); ok {
		// 直接辞書の場合
		encryptDict = dict
	} else {
		return fmt.Errorf("invalid Encrypt entry type: %T", encryptRef)
	}

	// File ID を取得
	var fileID []byte
	if idArray, ok := r.trailer[core.Name("ID")].(core.Array); ok && len(idArray) > 0 {
		if idStr, ok := idArray[0].(core.String); ok {
			fileID = []byte(idStr)
		}
	}
	if len(fileID) == 0 {
		return fmt.Errorf("missing or invalid File ID for encrypted PDF")
	}

	// 暗号化情報を解析
	encryptionInfo, err := parseEncryptDict(encryptDict, fileID)
	if err != nil {
		return fmt.Errorf("failed to parse Encrypt dictionary: %w", err)
	}

	r.encryption = encryptionInfo
	return nil
}

// findStartXref はstartxrefの値を取得する
func (r *Reader) findStartXref() (int64, error) {
	// ファイルの末尾にシーク
	if _, err := r.r.Seek(0, io.SeekEnd); err != nil {
		return 0, fmt.Errorf("failed to seek to end: %w", err)
	}

	// 末尾から1024バイト（またはファイルサイズ）を読む
	const bufSize = 1024
	buf := make([]byte, bufSize)

	// 現在位置を取得
	pos, _ := r.r.Seek(0, io.SeekCurrent)
	startPos := pos - bufSize
	if startPos < 0 {
		startPos = 0
	}

	// startPosにシークして読む
	if _, err := r.r.Seek(startPos, io.SeekStart); err != nil {
		return 0, fmt.Errorf("failed to seek: %w", err)
	}
	n, _ := r.r.Read(buf)
	buf = buf[:n]

	// "startxref" を探す
	startxrefIdx := bytes.LastIndex(buf, []byte("startxref"))
	if startxrefIdx == -1 {
		return 0, fmt.Errorf("startxref not found")
	}

	// startxrefの後の数値を読む
	afterStartxref := buf[startxrefIdx+len("startxref"):]
	scanner := bufio.NewScanner(bytes.NewReader(afterStartxref))
	scanner.Split(bufio.ScanWords)

	if !scanner.Scan() {
		return 0, fmt.Errorf("no offset after startxref")
	}

	offsetStr := scanner.Text()
	offset, err := strconv.ParseInt(offsetStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid startxref offset: %w", err)
	}

	return offset, nil
}

// parseXrefAndTrailer はxrefテーブルとtrailerを解析する
func (r *Reader) parseXrefAndTrailer(offset int64) error {
	// xrefオフセット位置にシーク
	if _, err := r.r.Seek(offset, io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek to xref: %w", err)
	}

	// "xref" キーワードを確認
	reader := bufio.NewReader(r.r)
	line, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	if !strings.HasPrefix(strings.TrimSpace(line), "xref") {
		return fmt.Errorf("expected 'xref' keyword, got %q", line)
	}

	// xrefサブセクションを読む
	for {
		// 次の行を読む
		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}

		line = strings.TrimSpace(line)

		// "trailer" キーワードに達したら終了
		if strings.HasPrefix(line, "trailer") {
			break
		}

		// サブセクションヘッダーをパース: "startNum count"
		parts := strings.Fields(line)
		if len(parts) != 2 {
			return fmt.Errorf("invalid xref subsection header: %q", line)
		}

		startNum, err := strconv.Atoi(parts[0])
		if err != nil {
			return fmt.Errorf("invalid xref start number: %w", err)
		}

		count, err := strconv.Atoi(parts[1])
		if err != nil {
			return fmt.Errorf("invalid xref count: %w", err)
		}

		// エントリを読む
		for i := 0; i < count; i++ {
			entryLine, err := reader.ReadString('\n')
			if err != nil {
				return err
			}

			// エントリをパース: "offset generation n/f"
			entryParts := strings.Fields(entryLine)
			if len(entryParts) != 3 {
				return fmt.Errorf("invalid xref entry: %q", entryLine)
			}

			offset, err := strconv.ParseInt(entryParts[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid xref offset: %w", err)
			}

			generation, err := strconv.Atoi(entryParts[1])
			if err != nil {
				return fmt.Errorf("invalid xref generation: %w", err)
			}

			inUse := entryParts[2] == "n"

			objNum := startNum + i
			r.xref[objNum] = xrefEntry{
				offset:     offset,
				generation: generation,
				inUse:      inUse,
			}
		}
	}

	// trailerを解析
	// 現在のreaderの残りをパーサーに渡す
	parser := NewParser(reader)

	trailerObj, err := parser.ParseObject()
	if err != nil {
		return fmt.Errorf("failed to parse trailer: %w", err)
	}

	trailer, err := utils.MustExtractAs[core.Dictionary](trailerObj, "trailer")
	if err != nil {
		return err
	}

	r.trailer = trailer

	return nil
}

// GetObject はオブジェクト番号からオブジェクトを取得する
func (r *Reader) GetObject(objNum int) (core.Object, error) {
	// キャッシュをチェック
	if obj, ok := r.objCache[objNum]; ok {
		return obj, nil
	}

	// xrefからエントリを取得
	entry, ok := r.xref[objNum]
	if !ok {
		return nil, fmt.Errorf("object %d not found in xref", objNum)
	}

	if !entry.inUse {
		return nil, fmt.Errorf("object %d is not in use", objNum)
	}

	// オフセット位置にシーク
	if _, err := r.r.Seek(entry.offset, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to seek to object: %w", err)
	}

	// 間接オブジェクトをパース
	parser := NewParser(r.r)
	num, gen, obj, err := parser.ParseIndirectObject()
	if err != nil {
		return nil, fmt.Errorf("failed to parse object %d: %w", objNum, err)
	}

	// オブジェクト番号と世代番号の確認
	if num != objNum {
		return nil, fmt.Errorf("object number mismatch: expected %d, got %d", objNum, num)
	}
	if gen != entry.generation {
		return nil, fmt.Errorf("generation number mismatch for object %d: expected %d, got %d", objNum, entry.generation, gen)
	}

	// 暗号化されている場合は復号化
	// ただし、Encrypt辞書自体は暗号化されていないのでスキップ
	if r.encryption != nil && r.encryption.Authenticated && !r.isEncryptObject(objNum) {
		obj = r.decryptObject(obj, objNum, gen)
	}

	// キャッシュに保存
	r.objCache[objNum] = obj

	return obj, nil
}

// ResolveReference は参照を解決してオブジェクトを取得する
func (r *Reader) ResolveReference(ref *core.Reference) (core.Object, error) {
	return r.GetObject(ref.ObjectNumber)
}

// GetCatalog はCatalogオブジェクトを返す
func (r *Reader) GetCatalog() (core.Dictionary, error) {
	// trailerから/Rootを取得
	rootRef, err := utils.MustExtractAs[*core.Reference](r.trailer[core.Name("Root")], "trailer /Root")
	if err != nil {
		return nil, err
	}

	// Catalogオブジェクトを取得
	catalogObj, err := r.GetObject(rootRef.ObjectNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get catalog: %w", err)
	}

	catalog, err := utils.MustExtractAs[core.Dictionary](catalogObj, "catalog")
	if err != nil {
		return nil, err
	}

	return catalog, nil
}

// GetPageCount はページ数を返す
func (r *Reader) GetPageCount() (int, error) {
	// Catalogを取得
	catalog, err := r.GetCatalog()
	if err != nil {
		return 0, err
	}

	// /Pagesを取得
	pagesRef, err := utils.MustExtractAs[*core.Reference](catalog[core.Name("Pages")], "catalog /Pages")
	if err != nil {
		return 0, err
	}

	pagesObj, err := r.GetObject(pagesRef.ObjectNumber)
	if err != nil {
		return 0, fmt.Errorf("failed to get pages: %w", err)
	}

	pages, err := utils.MustExtractAs[core.Dictionary](pagesObj, "pages")
	if err != nil {
		return 0, err
	}

	// /Countを取得
	countObj, ok := pages[core.Name("Count")]
	if !ok {
		return 0, fmt.Errorf("pages dictionary has no /Count")
	}

	count, err := utils.MustExtractAs[core.Integer](countObj, "pages /Count")
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

// GetPage は指定されたページ番号のPageオブジェクトを返す（0-indexed）
func (r *Reader) GetPage(pageNum int) (core.Dictionary, error) {
	// Catalogを取得
	catalog, err := r.GetCatalog()
	if err != nil {
		return nil, err
	}

	// /Pagesを取得
	pagesRef, err := utils.MustExtractAs[*core.Reference](catalog[core.Name("Pages")], "catalog /Pages")
	if err != nil {
		return nil, err
	}

	pagesObj, err := r.GetObject(pagesRef.ObjectNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get pages: %w", err)
	}

	pages, err := utils.MustExtractAs[core.Dictionary](pagesObj, "pages")
	if err != nil {
		return nil, err
	}

	// /Kidsから指定されたページを取得
	kidsObj, ok := pages[core.Name("Kids")]
	if !ok {
		return nil, fmt.Errorf("pages dictionary has no /Kids")
	}

	kids, err := utils.MustExtractAs[core.Array](kidsObj, "pages /Kids")
	if err != nil {
		return nil, err
	}

	if pageNum < 0 || pageNum >= len(kids) {
		return nil, fmt.Errorf("page number %d out of range [0, %d)", pageNum, len(kids))
	}

	// ページ参照を取得
	pageRef, err := utils.MustExtractAs[*core.Reference](kids[pageNum], "page reference")
	if err != nil {
		return nil, err
	}

	pageObj, err := r.GetObject(pageRef.ObjectNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get page %d: %w", pageNum, err)
	}

	page, err := utils.MustExtractAs[core.Dictionary](pageObj, "page")
	if err != nil {
		return nil, err
	}

	return page, nil
}

// GetInfo はInfo辞書（メタデータ）を返す
func (r *Reader) GetInfo() (core.Dictionary, error) {
	// trailerから/Infoを取得
	infoRef, ok := utils.ExtractAs[*core.Reference](r.trailer[core.Name("Info")])
	if !ok {
		// /Infoがない場合は空の辞書を返す
		return make(core.Dictionary), nil
	}

	// Infoオブジェクトを取得
	infoObj, err := r.GetObject(infoRef.ObjectNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get info: %w", err)
	}

	info, err := utils.MustExtractAs[core.Dictionary](infoObj, "info")
	if err != nil {
		return nil, err
	}

	return info, nil
}

// GetPageResources はページのResourcesを取得する
func (r *Reader) GetPageResources(page core.Dictionary) (core.Dictionary, error) {
	resourcesObj, ok := page[core.Name("Resources")]
	if !ok {
		return nil, nil // Resourcesがない場合
	}

	// Referenceの場合は解決
	if ref, ok := utils.ExtractAs[*core.Reference](resourcesObj); ok {
		obj, err := r.GetObject(ref.ObjectNumber)
		if err != nil {
			return nil, err
		}
		resourcesObj = obj
	}

	resources, err := utils.MustExtractAs[core.Dictionary](resourcesObj, "resources")
	if err != nil {
		return nil, err
	}

	return resources, nil
}

// ImageXObject は画像XObject
type ImageXObject struct {
	Stream           *core.Stream
	Width            int
	Height           int
	ColorSpace       string
	BitsPerComponent int
	Filter           string
}

// GetImageXObject は画像XObjectを取得する
func (r *Reader) GetImageXObject(ref *core.Reference) (*ImageXObject, error) {
	obj, err := r.GetObject(ref.ObjectNumber)
	if err != nil {
		return nil, err
	}

	stream, err := utils.MustExtractAs[*core.Stream](obj, "image xobject")
	if err != nil {
		return nil, err
	}

	// /Subtypeの確認
	subtype, _ := utils.ExtractAs[core.Name](stream.Dict[core.Name("Subtype")])
	if subtype != "Image" {
		return nil, fmt.Errorf("not an image xobject")
	}

	// 画像情報を抽出
	img := &ImageXObject{
		Stream: stream,
	}

	// Width
	if w, ok := utils.ExtractAs[core.Integer](stream.Dict[core.Name("Width")]); ok {
		img.Width = int(w)
	}

	// Height
	if h, ok := utils.ExtractAs[core.Integer](stream.Dict[core.Name("Height")]); ok {
		img.Height = int(h)
	}

	// ColorSpace
	if cs, ok := utils.ExtractAs[core.Name](stream.Dict[core.Name("ColorSpace")]); ok {
		img.ColorSpace = string(cs)
	}

	// BitsPerComponent
	if bpc, ok := utils.ExtractAs[core.Integer](stream.Dict[core.Name("BitsPerComponent")]); ok {
		img.BitsPerComponent = int(bpc)
	}

	// Filter
	if filter, ok := utils.ExtractAs[core.Name](stream.Dict[core.Name("Filter")]); ok {
		img.Filter = string(filter)
	}

	return img, nil
}

// GetPageContents はページのコンテンツストリームを取得してデコードする
func (r *Reader) GetPageContents(page core.Dictionary) ([]byte, error) {
	// /Contentsを取得
	contentsObj, ok := page[core.Name("Contents")]
	if !ok {
		// Contentsがない場合は空のバイト列を返す
		return []byte{}, nil
	}

	// Referenceの場合は解決
	if ref, ok := utils.ExtractAs[*core.Reference](contentsObj); ok {
		obj, err := r.GetObject(ref.ObjectNumber)
		if err != nil {
			return nil, fmt.Errorf("failed to get contents object: %w", err)
		}
		contentsObj = obj
	}

	// Streamの場合
	if stream, ok := utils.ExtractAs[*core.Stream](contentsObj); ok {
		return r.decodeStream(stream)
	}

	// Arrayの場合（複数のストリーム）
	if array, ok := utils.ExtractAs[core.Array](contentsObj); ok {
		var result []byte
		for _, item := range array {
			// 各要素を解決
			if ref, ok := utils.ExtractAs[*core.Reference](item); ok {
				obj, err := r.GetObject(ref.ObjectNumber)
				if err != nil {
					return nil, fmt.Errorf("failed to get stream from array: %w", err)
				}
				item = obj
			}

			// Streamをデコード
			if stream, ok := utils.ExtractAs[*core.Stream](item); ok {
				data, err := r.decodeStream(stream)
				if err != nil {
					return nil, err
				}
				result = append(result, data...)
				// ストリーム間に空白を追加
				result = append(result, ' ')
			}
		}
		return result, nil
	}

	return nil, fmt.Errorf("contents is neither a stream nor an array")
}

// decodeStream はストリームをデコードする
func (r *Reader) decodeStream(stream *core.Stream) ([]byte, error) {
	data := stream.Data

	// /Filterをチェック
	filterObj, hasFilter := stream.Dict[core.Name("Filter")]
	if !hasFilter {
		// フィルターがない場合はそのまま返す
		return data, nil
	}

	// Filterの解決
	if ref, ok := utils.ExtractAs[*core.Reference](filterObj); ok {
		obj, err := r.GetObject(ref.ObjectNumber)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve filter: %w", err)
		}
		filterObj = obj
	}

	// Filterが名前の場合
	if filterName, ok := utils.ExtractAs[core.Name](filterObj); ok {
		return r.applyFilter(data, string(filterName))
	}

	// Filterが配列の場合（複数のフィルター）
	if filterArray, ok := utils.ExtractAs[core.Array](filterObj); ok {
		for _, f := range filterArray {
			filterName, ok := utils.ExtractAs[core.Name](f)
			if !ok {
				continue
			}
			var err error
			data, err = r.applyFilter(data, string(filterName))
			if err != nil {
				return nil, err
			}
		}
		return data, nil
	}

	return data, nil
}

// applyFilter はフィルターを適用する
func (r *Reader) applyFilter(data []byte, filterName string) ([]byte, error) {
	switch filterName {
	case "FlateDecode":
		// zlibで解凍
		reader, err := zlib.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("failed to create zlib reader: %w", err)
		}
		defer reader.Close()

		var buf bytes.Buffer
		_, err = io.Copy(&buf, reader)
		if err != nil {
			return nil, fmt.Errorf("failed to decompress stream: %w", err)
		}

		return buf.Bytes(), nil

	default:
		// サポートしていないフィルターの場合はそのまま返す
		return data, nil
	}
}

// IsEncrypted returns true if the PDF is encrypted
func (r *Reader) IsEncrypted() bool {
	return r.encryption != nil
}

// IsAuthenticated returns true if the PDF has been successfully authenticated with a password
func (r *Reader) IsAuthenticated() bool {
	return r.encryption != nil && r.encryption.Authenticated
}

// AuthenticateWithPassword attempts to authenticate the PDF with the given password
// Returns an error if the PDF is not encrypted or if authentication fails
func (r *Reader) AuthenticateWithPassword(password string) error {
	if r.encryption == nil {
		return fmt.Errorf("PDF is not encrypted")
	}

	return r.encryption.Authenticate(password)
}

// GetEncryptionInfo returns the encryption information (for debugging/info purposes)
func (r *Reader) GetEncryptionInfo() *EncryptionInfo {
	return r.encryption
}

// isEncryptObject checks if the given object number is the Encrypt dictionary
func (r *Reader) isEncryptObject(objNum int) bool {
	if r.encryption == nil {
		return false
	}

	// Check if Encrypt entry in trailer points to this object
	if encryptRef, ok := r.trailer[core.Name("Encrypt")].(*core.Reference); ok {
		return encryptRef.ObjectNumber == objNum
	}

	return false
}

// decryptObject decrypts an object if necessary
func (r *Reader) decryptObject(obj core.Object, objectNumber, generationNumber int) core.Object {
	switch v := obj.(type) {
	case *core.Stream:
		// Decrypt stream data
		decryptedData := r.encryption.DecryptStream(v.Data, objectNumber, generationNumber)

		// Create new stream with decrypted data
		newDict := make(core.Dictionary)
		for k, val := range v.Dict {
			// Recursively decrypt dictionary values (except certain keys)
			if k != core.Name("Length") && k != core.Name("Filter") && k != core.Name("DecodeParms") {
				newDict[k] = r.decryptObject(val, objectNumber, generationNumber)
			} else {
				newDict[k] = val
			}
		}

		// Update Length to reflect decrypted data
		newDict[core.Name("Length")] = core.Integer(len(decryptedData))

		return &core.Stream{
			Dict: newDict,
			Data: decryptedData,
		}

	case core.String:
		// Decrypt string
		decrypted := r.encryption.DecryptString([]byte(v), objectNumber, generationNumber)
		return core.String(decrypted)

	case core.Dictionary:
		// Recursively decrypt dictionary values
		newDict := make(core.Dictionary)
		for k, val := range v {
			newDict[k] = r.decryptObject(val, objectNumber, generationNumber)
		}
		return newDict

	case core.Array:
		// Recursively decrypt array elements
		newArray := make(core.Array, len(v))
		for i, val := range v {
			newArray[i] = r.decryptObject(val, objectNumber, generationNumber)
		}
		return newArray

	default:
		// Other types don't need decryption
		return obj
	}
}
