package gopdf

import (
	"io"
	"os"

	"github.com/ryomak/gopdf/internal/core"
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

	return Metadata{
		Title:    getString(infoDict, "Title"),
		Author:   getString(infoDict, "Author"),
		Subject:  getString(infoDict, "Subject"),
		Keywords: getString(infoDict, "Keywords"),
		Creator:  getString(infoDict, "Creator"),
		Producer: getString(infoDict, "Producer"),
	}
}

// Metadata はPDFメタデータ
type Metadata struct {
	Title    string
	Author   string
	Subject  string
	Keywords string
	Creator  string
	Producer string
}

// getString は辞書から文字列値を取得する
func getString(dict core.Dictionary, key string) string {
	obj, ok := dict[core.Name(key)]
	if !ok {
		return ""
	}

	switch v := obj.(type) {
	case core.String:
		return string(v)
	default:
		return ""
	}
}
