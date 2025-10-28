package gopdf

// TextIterator はPDFページのテキスト要素をイテレートする
type TextIterator struct {
	elements []TextElement
	index    int
	current  *TextElement
}

// Next は次のテキスト要素に進む
// 次の要素が存在する場合はtrueを返す
func (it *TextIterator) Next() bool {
	if it.index >= len(it.elements) {
		it.current = nil
		return false
	}

	it.current = &it.elements[it.index]
	it.index++
	return true
}

// Element は現在のテキスト要素を返す
// Next()を呼んでいない場合や、要素がない場合はnilを返す
func (it *TextIterator) Element() *TextElement {
	return it.current
}

// HasNext は次の要素が存在するかを返す
func (it *TextIterator) HasNext() bool {
	return it.index < len(it.elements)
}

// Reset はイテレーターを最初の位置にリセットする
func (it *TextIterator) Reset() {
	it.index = 0
	it.current = nil
}

// Count は総要素数を返す
func (it *TextIterator) Count() int {
	return len(it.elements)
}

// TextIterator はページのテキスト要素をイテレートするためのイテレーターを返す
func (r *PDFReader) TextIterator(pageNum int) (*TextIterator, error) {
	elements, err := r.ExtractPageTextElements(pageNum)
	if err != nil {
		return nil, err
	}

	return &TextIterator{
		elements: elements,
		index:    0,
	}, nil
}

// AllPagesTextIterator は全ページのテキスト要素をイテレートする
type AllPagesTextIterator struct {
	reader      *PDFReader
	pageCount   int
	currentPage int
	pageIter    *TextIterator
}

// Next は次のテキスト要素に進む
func (it *AllPagesTextIterator) Next() bool {
	// 現在のページのイテレーターがある場合、次の要素を試す
	if it.pageIter != nil && it.pageIter.Next() {
		return true
	}

	// 次のページに進む
	it.currentPage++
	if it.currentPage >= it.pageCount {
		return false
	}

	// 新しいページのイテレーターを作成
	pageIter, err := it.reader.TextIterator(it.currentPage)
	if err != nil {
		return false
	}

	it.pageIter = pageIter

	// 新しいページの最初の要素に進む
	return it.pageIter.Next()
}

// Element は現在のテキスト要素を返す
func (it *AllPagesTextIterator) Element() *TextElement {
	if it.pageIter == nil {
		return nil
	}
	return it.pageIter.Element()
}

// CurrentPage は現在のページ番号を返す（0-indexed）
func (it *AllPagesTextIterator) CurrentPage() int {
	return it.currentPage
}

// Reset はイテレーターを最初の位置にリセットする
func (it *AllPagesTextIterator) Reset() {
	it.currentPage = -1
	it.pageIter = nil
}

// AllPagesTextIterator は全ページのテキスト要素をイテレートするイテレーターを返す
func (r *PDFReader) AllPagesTextIterator() *AllPagesTextIterator {
	return &AllPagesTextIterator{
		reader:      r,
		pageCount:   r.PageCount(),
		currentPage: -1,
	}
}
