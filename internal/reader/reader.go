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
)

// xrefEntry はクロスリファレンステーブルのエントリ
type xrefEntry struct {
	offset     int64 // ファイル内バイトオフセット
	generation int   // 世代番号
	inUse      bool  // 使用中かどうか
}

// Reader はPDFファイルを読み込み、解析する
type Reader struct {
	r        io.ReadSeeker       // ファイルのシーク可能なリーダー
	xref     map[int]xrefEntry   // オブジェクト番号 -> xrefエントリ
	trailer  core.Dictionary     // Trailer辞書
	objCache map[int]core.Object // オブジェクトキャッシュ
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

	return nil
}

// findStartXref はstartxrefの値を取得する
func (r *Reader) findStartXref() (int64, error) {
	// ファイルの末尾にシーク
	r.r.Seek(0, io.SeekEnd)

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
	r.r.Seek(startPos, io.SeekStart)
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
	r.r.Seek(offset, io.SeekStart)

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

	trailer, ok := trailerObj.(core.Dictionary)
	if !ok {
		return fmt.Errorf("trailer should be dictionary, got %T", trailerObj)
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
	r.r.Seek(entry.offset, io.SeekStart)

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
	rootRef, ok := r.trailer[core.Name("Root")].(*core.Reference)
	if !ok {
		return nil, fmt.Errorf("trailer /Root is not a reference")
	}

	// Catalogオブジェクトを取得
	catalogObj, err := r.GetObject(rootRef.ObjectNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get catalog: %w", err)
	}

	catalog, ok := catalogObj.(core.Dictionary)
	if !ok {
		return nil, fmt.Errorf("catalog is not a dictionary")
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
	pagesRef, ok := catalog[core.Name("Pages")].(*core.Reference)
	if !ok {
		return 0, fmt.Errorf("catalog /Pages is not a reference")
	}

	pagesObj, err := r.GetObject(pagesRef.ObjectNumber)
	if err != nil {
		return 0, fmt.Errorf("failed to get pages: %w", err)
	}

	pages, ok := pagesObj.(core.Dictionary)
	if !ok {
		return 0, fmt.Errorf("pages is not a dictionary")
	}

	// /Countを取得
	countObj, ok := pages[core.Name("Count")]
	if !ok {
		return 0, fmt.Errorf("pages dictionary has no /Count")
	}

	count, ok := countObj.(core.Integer)
	if !ok {
		return 0, fmt.Errorf("pages /Count is not an integer")
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
	pagesRef, ok := catalog[core.Name("Pages")].(*core.Reference)
	if !ok {
		return nil, fmt.Errorf("catalog /Pages is not a reference")
	}

	pagesObj, err := r.GetObject(pagesRef.ObjectNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get pages: %w", err)
	}

	pages, ok := pagesObj.(core.Dictionary)
	if !ok {
		return nil, fmt.Errorf("pages is not a dictionary")
	}

	// /Kidsから指定されたページを取得
	kidsObj, ok := pages[core.Name("Kids")]
	if !ok {
		return nil, fmt.Errorf("pages dictionary has no /Kids")
	}

	kids, ok := kidsObj.(core.Array)
	if !ok {
		return nil, fmt.Errorf("pages /Kids is not an array")
	}

	if pageNum < 0 || pageNum >= len(kids) {
		return nil, fmt.Errorf("page number %d out of range [0, %d)", pageNum, len(kids))
	}

	// ページ参照を取得
	pageRef, ok := kids[pageNum].(*core.Reference)
	if !ok {
		return nil, fmt.Errorf("page reference is not a reference")
	}

	pageObj, err := r.GetObject(pageRef.ObjectNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get page %d: %w", pageNum, err)
	}

	page, ok := pageObj.(core.Dictionary)
	if !ok {
		return nil, fmt.Errorf("page is not a dictionary")
	}

	return page, nil
}

// GetInfo はInfo辞書（メタデータ）を返す
func (r *Reader) GetInfo() (core.Dictionary, error) {
	// trailerから/Infoを取得
	infoRef, ok := r.trailer[core.Name("Info")].(*core.Reference)
	if !ok {
		// /Infoがない場合は空の辞書を返す
		return make(core.Dictionary), nil
	}

	// Infoオブジェクトを取得
	infoObj, err := r.GetObject(infoRef.ObjectNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get info: %w", err)
	}

	info, ok := infoObj.(core.Dictionary)
	if !ok {
		return nil, fmt.Errorf("info is not a dictionary")
	}

	return info, nil
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
	if ref, ok := contentsObj.(*core.Reference); ok {
		obj, err := r.GetObject(ref.ObjectNumber)
		if err != nil {
			return nil, fmt.Errorf("failed to get contents object: %w", err)
		}
		contentsObj = obj
	}

	// Streamの場合
	if stream, ok := contentsObj.(*core.Stream); ok {
		return r.decodeStream(stream)
	}

	// Arrayの場合（複数のストリーム）
	if array, ok := contentsObj.(core.Array); ok {
		var result []byte
		for _, item := range array {
			// 各要素を解決
			if ref, ok := item.(*core.Reference); ok {
				obj, err := r.GetObject(ref.ObjectNumber)
				if err != nil {
					return nil, fmt.Errorf("failed to get stream from array: %w", err)
				}
				item = obj
			}

			// Streamをデコード
			if stream, ok := item.(*core.Stream); ok {
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
	if ref, ok := filterObj.(*core.Reference); ok {
		obj, err := r.GetObject(ref.ObjectNumber)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve filter: %w", err)
		}
		filterObj = obj
	}

	// Filterが名前の場合
	if filterName, ok := filterObj.(core.Name); ok {
		return r.applyFilter(data, string(filterName))
	}

	// Filterが配列の場合（複数のフィルター）
	if filterArray, ok := filterObj.(core.Array); ok {
		for _, f := range filterArray {
			filterName, ok := f.(core.Name)
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
