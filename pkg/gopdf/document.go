package gopdf

import (
	"io"

	"github.com/ryomak/gopdf/internal/core"
	"github.com/ryomak/gopdf/internal/writer"
)

// Document represents a PDF document.
type Document struct {
	pages []*Page
}

// New creates a new PDF document.
func New() *Document {
	return &Document{
		pages: make([]*Page, 0),
	}
}

// AddPage adds a new page to the document and returns it.
func (d *Document) AddPage(size PageSize, orientation Orientation) *Page {
	actualSize := orientation.Apply(size)
	page := &Page{
		width:  actualSize.Width,
		height: actualSize.Height,
	}
	d.pages = append(d.pages, page)
	return page
}

// WriteTo writes the PDF document to the given writer.
func (d *Document) WriteTo(w io.Writer) error {
	pdfWriter := writer.NewWriter(w)

	// ヘッダーを書く
	if err := pdfWriter.WriteHeader(); err != nil {
		return err
	}

	// まず、全ページで使用されているフォントを収集
	allFonts := make(map[string]*core.Reference)
	for _, page := range d.pages {
		for fontKey := range page.fonts {
			if _, exists := allFonts[fontKey]; !exists {
				// プレースホルダー（後で実際のオブジェクト番号を設定）
				allFonts[fontKey] = nil
			}
		}
	}

	// Pagesオブジェクトの番号を計算
	// Font(フォント数) + Content(1) + Page(1) のペアが len(d.pages) 個
	// 次のオブジェクト番号 = 1 + フォント数 + len(d.pages)*2
	pagesObjNum := 1 + len(allFonts) + len(d.pages)*2

	// フォントオブジェクトを作成
	for fontKey := range allFonts {
		// フォント名を取得
		var fontName string
		for _, page := range d.pages {
			if f, ok := page.fonts[fontKey]; ok {
				fontName = f.Name()
				break
			}
		}

		fontDict := core.Dictionary{
			core.Name("Type"):     core.Name("Font"),
			core.Name("Subtype"):  core.Name("Type1"),
			core.Name("BaseFont"): core.Name(fontName),
		}

		fontNum, err := pdfWriter.AddObject(fontDict)
		if err != nil {
			return err
		}

		allFonts[fontKey] = &core.Reference{
			ObjectNumber:     fontNum,
			GenerationNumber: 0,
		}
	}

	// 各ページのコンテンツストリームとPageオブジェクトを作成
	pageRefs := make([]*core.Reference, 0, len(d.pages))
	for _, page := range d.pages {
		// コンテンツストリームの作成
		contentData := page.content.Bytes()
		contentDict := core.Dictionary{
			core.Name("Length"): core.Integer(len(contentData)),
		}
		contentStream := &core.Stream{
			Dict: contentDict,
			Data: contentData,
		}

		// コンテンツストリームオブジェクトを追加
		contentNum, err := pdfWriter.AddObject(contentStream)
		if err != nil {
			return err
		}

		// Resourcesディクショナリを構築
		resourcesDict := core.Dictionary{}

		// このページで使用されているフォントをResourcesに追加
		if len(page.fonts) > 0 {
			fontResources := core.Dictionary{}
			for fontKey := range page.fonts {
				fontResources[core.Name(fontKey)] = allFonts[fontKey]
			}
			resourcesDict[core.Name("Font")] = fontResources
		}

		// Pageオブジェクトを作成（ParentにPagesへの参照を設定）
		pageDict := core.Dictionary{
			core.Name("Type"): core.Name("Page"),
			core.Name("Parent"): &core.Reference{
				ObjectNumber:     pagesObjNum,
				GenerationNumber: 0,
			},
			core.Name("MediaBox"): core.Array{
				core.Integer(0),
				core.Integer(0),
				core.Real(page.width),
				core.Real(page.height),
			},
			core.Name("Contents"): &core.Reference{
				ObjectNumber:     contentNum,
				GenerationNumber: 0,
			},
			core.Name("Resources"): resourcesDict,
		}

		// Pageオブジェクトを追加
		pageNum, err := pdfWriter.AddObject(pageDict)
		if err != nil {
			return err
		}

		pageRefs = append(pageRefs, &core.Reference{
			ObjectNumber:     pageNum,
			GenerationNumber: 0,
		})
	}

	// Pagesオブジェクトを作成
	kids := make(core.Array, len(pageRefs))
	for i, ref := range pageRefs {
		kids[i] = ref
	}

	pagesDict := core.Dictionary{
		core.Name("Type"):  core.Name("Pages"),
		core.Name("Kids"):  kids,
		core.Name("Count"): core.Integer(len(d.pages)),
	}

	pagesNum, err := pdfWriter.AddObject(pagesDict)
	if err != nil {
		return err
	}

	// Catalogオブジェクトを作成
	catalogDict := core.Dictionary{
		core.Name("Type"): core.Name("Catalog"),
		core.Name("Pages"): &core.Reference{
			ObjectNumber:     pagesNum,
			GenerationNumber: 0,
		},
	}

	catalogNum, err := pdfWriter.AddObject(catalogDict)
	if err != nil {
		return err
	}

	// Trailerを書く
	// ここで全オブジェクト数を計算: Catalog + Pages + (Content + Page) * ページ数 + 1(offset 0)
	totalObjects := 1 + 1 + len(d.pages)*2 + 1

	trailer := core.Dictionary{
		core.Name("Size"): core.Integer(totalObjects),
		core.Name("Root"): &core.Reference{
			ObjectNumber:     catalogNum,
			GenerationNumber: 0,
		},
	}

	return pdfWriter.WriteTrailer(trailer)
}

// PageCount returns the number of pages in the document.
func (d *Document) PageCount() int {
	return len(d.pages)
}
