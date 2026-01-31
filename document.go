package gopdf

import (
	"fmt"
	"io"

	"github.com/ryomak/gopdf/internal/core"
	"github.com/ryomak/gopdf/internal/writer"
)

// pdfBuildContext holds the state during PDF generation.
type pdfBuildContext struct {
	writer       *writer.Writer
	allFonts     map[string]*core.Reference
	allTTFFonts  map[string]*TTFFont
	ttfFontRefs  map[string]*core.Reference
	allImages    map[*Image]*core.Reference
	imageOrder   []*Image
	pagesObjNum  int
}

// Document represents a PDF document.
type Document struct {
	pages      []*Page
	encryption *EncryptionOptions
	metadata   *Metadata
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
	ctx := &pdfBuildContext{
		writer:      writer.NewWriter(w),
		allFonts:    make(map[string]*core.Reference),
		allTTFFonts: make(map[string]*TTFFont),
		ttfFontRefs: make(map[string]*core.Reference),
		allImages:   make(map[*Image]*core.Reference),
		imageOrder:  make([]*Image, 0),
	}

	if err := d.setupEncryption(ctx); err != nil {
		return err
	}

	if err := ctx.writer.WriteHeader(); err != nil {
		return err
	}

	d.collectResources(ctx)

	if err := d.embedTTFFonts(ctx); err != nil {
		return err
	}

	d.calculatePagesObjNum(ctx)

	if err := d.writeStandardFonts(ctx); err != nil {
		return err
	}

	if err := d.writeImages(ctx); err != nil {
		return err
	}

	pageRefs, err := d.writePages(ctx)
	if err != nil {
		return err
	}

	return d.writeCatalogAndTrailer(ctx, pageRefs)
}

// setupEncryption configures encryption if enabled.
func (d *Document) setupEncryption(ctx *pdfBuildContext) error {
	if d.encryption == nil {
		return nil
	}

	encryptionInfo, err := writer.SetupEncryption(
		d.encryption.UserPassword,
		d.encryption.OwnerPassword,
		d.encryption.Permissions.toInternal(),
		d.encryption.KeyLength,
	)
	if err != nil {
		return fmt.Errorf("failed to setup encryption: %w", err)
	}
	ctx.writer.SetEncryption(encryptionInfo)
	return nil
}

// collectResources gathers all fonts and images used across pages.
func (d *Document) collectResources(ctx *pdfBuildContext) {
	for _, page := range d.pages {
		// Collect standard fonts
		for fontKey := range page.fonts {
			if _, exists := ctx.allFonts[fontKey]; !exists {
				ctx.allFonts[fontKey] = nil
			}
		}

		// Collect TTF fonts
		for fontKey, ttfFont := range page.ttfFonts {
			if _, exists := ctx.allTTFFonts[fontKey]; !exists {
				ctx.allTTFFonts[fontKey] = ttfFont
				ctx.ttfFontRefs[fontKey] = nil
			}
		}

		// Collect images
		for _, img := range page.images {
			if _, exists := ctx.allImages[img]; !exists {
				ctx.allImages[img] = nil
				ctx.imageOrder = append(ctx.imageOrder, img)
			}
		}
	}
}

// embedTTFFonts embeds all TTF fonts into the PDF.
func (d *Document) embedTTFFonts(ctx *pdfBuildContext) error {
	ttfEmbedder := writer.NewTTFFontEmbedder(ctx.writer)
	for fontKey, ttfFont := range ctx.allTTFFonts {
		ttfFont.glyphsMutex.Lock()
		usedGlyphs := make(map[uint16]rune, len(ttfFont.usedGlyphs))
		for k, v := range ttfFont.usedGlyphs {
			usedGlyphs[k] = v
		}
		ttfFont.glyphsMutex.Unlock()

		fontRef, err := ttfEmbedder.EmbedTTFFont(ttfFont.internal, usedGlyphs)
		if err != nil {
			return fmt.Errorf("failed to embed TTF font %s: %w", fontKey, err)
		}
		ctx.ttfFontRefs[fontKey] = fontRef
	}
	return nil
}

// calculatePagesObjNum computes the object number for the Pages object.
func (d *Document) calculatePagesObjNum(ctx *pdfBuildContext) {
	ctx.pagesObjNum = 1 + len(ctx.allFonts) + len(ctx.allTTFFonts)*5 + len(ctx.allImages) + len(d.pages)*2
}

// writeStandardFonts writes standard font objects.
func (d *Document) writeStandardFonts(ctx *pdfBuildContext) error {
	for fontKey := range ctx.allFonts {
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

		fontNum, err := ctx.writer.AddObject(fontDict)
		if err != nil {
			return err
		}

		ctx.allFonts[fontKey] = &core.Reference{
			ObjectNumber:     fontNum,
			GenerationNumber: 0,
		}
	}
	return nil
}

// writeImages writes image XObjects.
func (d *Document) writeImages(ctx *pdfBuildContext) error {
	for _, img := range ctx.imageOrder {
		smaskRef, err := d.writeSMask(ctx, img)
		if err != nil {
			return err
		}

		imageDict := core.Dictionary{
			core.Name("Type"):             core.Name("XObject"),
			core.Name("Subtype"):          core.Name("Image"),
			core.Name("Width"):            core.Integer(img.Width),
			core.Name("Height"):           core.Integer(img.Height),
			core.Name("ColorSpace"):       core.Name(img.ColorSpace),
			core.Name("BitsPerComponent"): core.Integer(img.BitsPerComponent),
			core.Name("Filter"):           core.Name(img.Filter),
			core.Name("Length"):           core.Integer(len(img.Data)),
		}

		if smaskRef != nil {
			imageDict[core.Name("SMask")] = smaskRef
		}

		imageStream := &core.Stream{
			Dict: imageDict,
			Data: img.Data,
		}

		imgNum, err := ctx.writer.AddObject(imageStream)
		if err != nil {
			return err
		}

		ctx.allImages[img] = &core.Reference{
			ObjectNumber:     imgNum,
			GenerationNumber: 0,
		}
	}
	return nil
}

// writeSMask writes the SMask (alpha channel) for an image if present.
func (d *Document) writeSMask(ctx *pdfBuildContext, img *Image) (*core.Reference, error) {
	if img.SMask == nil {
		return nil, nil
	}

	smaskDict := core.Dictionary{
		core.Name("Type"):             core.Name("XObject"),
		core.Name("Subtype"):          core.Name("Image"),
		core.Name("Width"):            core.Integer(img.SMask.Width),
		core.Name("Height"):           core.Integer(img.SMask.Height),
		core.Name("ColorSpace"):       core.Name(img.SMask.ColorSpace),
		core.Name("BitsPerComponent"): core.Integer(img.SMask.BitsPerComponent),
		core.Name("Filter"):           core.Name(img.SMask.Filter),
		core.Name("Length"):           core.Integer(len(img.SMask.Data)),
	}

	smaskStream := &core.Stream{
		Dict: smaskDict,
		Data: img.SMask.Data,
	}

	smaskNum, err := ctx.writer.AddObject(smaskStream)
	if err != nil {
		return nil, err
	}

	return &core.Reference{
		ObjectNumber:     smaskNum,
		GenerationNumber: 0,
	}, nil
}

// writePages writes content streams and page objects.
func (d *Document) writePages(ctx *pdfBuildContext) ([]*core.Reference, error) {
	pageRefs := make([]*core.Reference, 0, len(d.pages))

	for _, page := range d.pages {
		contentNum, err := d.writePageContent(ctx, page)
		if err != nil {
			return nil, err
		}

		resourcesDict := d.buildPageResources(ctx, page)

		pageDict := core.Dictionary{
			core.Name("Type"): core.Name("Page"),
			core.Name("Parent"): &core.Reference{
				ObjectNumber:     ctx.pagesObjNum,
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

		pageNum, err := ctx.writer.AddObject(pageDict)
		if err != nil {
			return nil, err
		}

		pageRefs = append(pageRefs, &core.Reference{
			ObjectNumber:     pageNum,
			GenerationNumber: 0,
		})
	}

	return pageRefs, nil
}

// writePageContent writes the content stream for a page.
func (d *Document) writePageContent(ctx *pdfBuildContext, page *Page) (int, error) {
	contentData := page.content.Bytes()
	contentDict := core.Dictionary{
		core.Name("Length"): core.Integer(len(contentData)),
	}
	contentStream := &core.Stream{
		Dict: contentDict,
		Data: contentData,
	}
	return ctx.writer.AddObject(contentStream)
}

// buildPageResources builds the Resources dictionary for a page.
func (d *Document) buildPageResources(ctx *pdfBuildContext, page *Page) core.Dictionary {
	resourcesDict := core.Dictionary{}

	// Font resources
	if len(page.fonts) > 0 || len(page.ttfFonts) > 0 {
		fontResources := core.Dictionary{}
		for fontKey := range page.fonts {
			fontResources[core.Name(fontKey)] = ctx.allFonts[fontKey]
		}
		for fontKey := range page.ttfFonts {
			fontResources[core.Name(fontKey)] = ctx.ttfFontRefs[fontKey]
		}
		resourcesDict[core.Name("Font")] = fontResources
	}

	// Image resources
	if len(page.images) > 0 {
		xobjectResources := core.Dictionary{}
		for i, img := range page.images {
			imageKey := fmt.Sprintf("Im%d", i+1)
			xobjectResources[core.Name(imageKey)] = ctx.allImages[img]
		}
		resourcesDict[core.Name("XObject")] = xobjectResources
	}

	// Graphics state resources
	if len(page.graphicsStates) > 0 {
		extGStateResources := core.Dictionary{}
		for gsName, opacity := range page.graphicsStates {
			gsDict := core.Dictionary{
				core.Name("Type"): core.Name("ExtGState"),
				core.Name("ca"):   core.Real(opacity),
				core.Name("CA"):   core.Real(opacity),
			}
			extGStateResources[core.Name(gsName)] = gsDict
		}
		resourcesDict[core.Name("ExtGState")] = extGStateResources
	}

	return resourcesDict
}

// writeCatalogAndTrailer writes Pages, Catalog, Info, and Trailer.
func (d *Document) writeCatalogAndTrailer(ctx *pdfBuildContext, pageRefs []*core.Reference) error {
	// Pages object
	kids := make(core.Array, len(pageRefs))
	for i, ref := range pageRefs {
		kids[i] = ref
	}

	pagesDict := core.Dictionary{
		core.Name("Type"):  core.Name("Pages"),
		core.Name("Kids"):  kids,
		core.Name("Count"): core.Integer(len(d.pages)),
	}

	pagesNum, err := ctx.writer.AddObject(pagesDict)
	if err != nil {
		return err
	}

	// Catalog object
	catalogDict := core.Dictionary{
		core.Name("Type"): core.Name("Catalog"),
		core.Name("Pages"): &core.Reference{
			ObjectNumber:     pagesNum,
			GenerationNumber: 0,
		},
	}

	catalogNum, err := ctx.writer.AddObject(catalogDict)
	if err != nil {
		return err
	}

	// Info dictionary
	var infoNum int
	if d.metadata != nil {
		infoDict := createInfoDict(d.metadata)
		if len(infoDict) > 0 {
			infoNum, err = ctx.writer.AddObject(infoDict)
			if err != nil {
				return err
			}
		}
	}

	// Trailer
	totalObjects := 1 + 1 + len(d.pages)*2 + 1
	if infoNum > 0 {
		totalObjects++
	}

	trailer := core.Dictionary{
		core.Name("Size"): core.Integer(totalObjects),
		core.Name("Root"): &core.Reference{
			ObjectNumber:     catalogNum,
			GenerationNumber: 0,
		},
	}

	if infoNum > 0 {
		trailer[core.Name("Info")] = &core.Reference{
			ObjectNumber:     infoNum,
			GenerationNumber: 0,
		}
	}

	return ctx.writer.WriteTrailer(trailer)
}

// PageCount returns the number of pages in the document.
func (d *Document) PageCount() int {
	return len(d.pages)
}

// SetEncryption sets encryption options for the PDF
// Must be called before WriteTo()
func (d *Document) SetEncryption(opts EncryptionOptions) error {
	// Validate options
	if err := opts.Validate(); err != nil {
		return fmt.Errorf("invalid encryption options: %w", err)
	}

	d.encryption = &opts
	return nil
}

// HasEncryption returns true if encryption is enabled
func (d *Document) HasEncryption() bool {
	return d.encryption != nil
}
