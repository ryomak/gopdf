package gopdf

import (
	"fmt"
	"io"
	"os"

	"github.com/ryomak/gopdf/internal/content/text"
	"github.com/ryomak/gopdf/internal/content/translate"
	"github.com/ryomak/gopdf/internal/font"
)

// TranslateUnit は翻訳の単位を指定する
type TranslateUnit int

const (
	// TranslateUnitBlock はブロック全体を1回で翻訳（デフォルト）
	TranslateUnitBlock TranslateUnit = iota
	// TranslateUnitLine は行単位で翻訳
	TranslateUnitLine
	// TranslateUnitSentence は文単位で翻訳（. 。 ! ? で区切る）
	TranslateUnitSentence
)

// PDFTranslatorOptions は翻訳オプション
type PDFTranslatorOptions struct {
	Translator      Translator    // 翻訳インターフェース（translate.Translator）
	TargetFont      Font          // ターゲット言語のフォント (StandardFont or *TTFFont)
	TargetBoldFont  Font          // ターゲット言語の太字フォント（省略可）
	TargetFontName  string        // フォント名（省略可、TargetFontから自動取得）
	FittingOptions  FitOptions    // テキストフィッティングオプション（FitOptions）
	KeepImages      bool          // 画像を保持（デフォルト: true）
	KeepLayout      bool          // レイアウトを保持（デフォルト: true）
	TranslateUnit   TranslateUnit // 翻訳単位（デフォルト: TranslateUnitBlock）
}

// getTargetFontName はフォント名を取得（自動取得対応）
func (opts PDFTranslatorOptions) getTargetFontName() string {
	if opts.TargetFontName != "" {
		return opts.TargetFontName
	}
	if opts.TargetFont != nil {
		return opts.TargetFont.Name()
	}
	return ""
}

// getTranslateUnit は翻訳単位を取得
func (opts PDFTranslatorOptions) getTranslateUnit() TranslateUnit {
	return opts.TranslateUnit
}

// DefaultPDFTranslatorOptions はデフォルトのオプション
func DefaultPDFTranslatorOptions(targetFont Font) PDFTranslatorOptions {
	return PDFTranslatorOptions{
		Translator:     nil,
		TargetFont:     targetFont,
		TargetFontName: "",
		FittingOptions: DefaultFitOptions(),
		KeepImages:     true,
		KeepLayout:     true,
	}
}

// TranslatorOption is a function that modifies PDFTranslatorOptions.
type TranslatorOption func(*PDFTranslatorOptions)

// NewTranslatorOptions creates PDFTranslatorOptions with the given options.
func NewTranslatorOptions(opts ...TranslatorOption) PDFTranslatorOptions {
	options := PDFTranslatorOptions{
		FittingOptions: DefaultFitOptions(),
		KeepImages:     true,
		KeepLayout:     true,
	}
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

// WithTranslatorFunc sets the translator function.
func WithTranslatorFunc(t Translator) TranslatorOption {
	return func(o *PDFTranslatorOptions) {
		o.Translator = t
	}
}

// WithTranslatorTargetFont sets the target font for non-ASCII text.
func WithTranslatorTargetFont(f Font) TranslatorOption {
	return func(o *PDFTranslatorOptions) {
		o.TargetFont = f
	}
}

// WithTranslatorTargetBoldFont sets the target bold font.
func WithTranslatorTargetBoldFont(f Font) TranslatorOption {
	return func(o *PDFTranslatorOptions) {
		o.TargetBoldFont = f
	}
}

// WithTranslatorTargetFontName sets the target font name explicitly.
func WithTranslatorTargetFontName(name string) TranslatorOption {
	return func(o *PDFTranslatorOptions) {
		o.TargetFontName = name
	}
}

// WithTranslatorFittingOptions sets the text fitting options.
func WithTranslatorFittingOptions(fo FitOptions) TranslatorOption {
	return func(o *PDFTranslatorOptions) {
		o.FittingOptions = fo
	}
}

// WithTranslatorKeepImages sets whether to preserve images.
func WithTranslatorKeepImages(keep bool) TranslatorOption {
	return func(o *PDFTranslatorOptions) {
		o.KeepImages = keep
	}
}

// WithTranslatorKeepLayout sets whether to preserve layout.
func WithTranslatorKeepLayout(keep bool) TranslatorOption {
	return func(o *PDFTranslatorOptions) {
		o.KeepLayout = keep
	}
}

// WithTranslatorUnit sets the translation unit (block, line, or sentence).
func WithTranslatorUnit(unit TranslateUnit) TranslatorOption {
	return func(o *PDFTranslatorOptions) {
		o.TranslateUnit = unit
	}
}

// TranslatePDF はPDFを翻訳して新しいPDFを生成
func TranslatePDF(inputPath string, outputPath string, opts PDFTranslatorOptions) error {
	reader, err := Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input PDF: %w", err)
	}
	defer reader.Close()

	doc, err := translatePages(reader, opts)
	if err != nil {
		return err
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	return doc.WriteTo(file)
}

// TranslatePDFToWriter はPDFを翻訳してWriterに出力
func TranslatePDFToWriter(input io.ReadSeeker, output io.Writer, opts PDFTranslatorOptions) error {
	reader, err := OpenReader(input)
	if err != nil {
		return fmt.Errorf("failed to open input PDF: %w", err)
	}
	defer reader.Close()

	doc, err := translatePages(reader, opts)
	if err != nil {
		return err
	}

	return doc.WriteTo(output)
}

// translatePages は全ページを翻訳してDocumentを生成する共通処理
func translatePages(reader *PDFReader, opts PDFTranslatorOptions) (*Document, error) {
	doc := New()
	pageCount := reader.PageCount()

	for i := 0; i < pageCount; i++ {
		layout, err := reader.ExtractPageLayout(i)
		if err != nil {
			return nil, fmt.Errorf("failed to extract layout from page %d: %w", i, err)
		}

		if opts.Translator != nil {
			if err := translateLayoutBlocks(layout, opts); err != nil {
				return nil, fmt.Errorf("translation failed on page %d: %w", i, err)
			}
		}

		if _, err := RenderLayout(doc, layout, opts); err != nil {
			return nil, fmt.Errorf("failed to render page %d: %w", i, err)
		}
	}

	return doc, nil
}

// translateLayoutBlocks はレイアウト内のテキストブロックを翻訳
func translateLayoutBlocks(layout *PageLayout, opts PDFTranslatorOptions) error {
	unit := opts.getTranslateUnit()
	for j := range layout.TextBlocks {
		translated, err := translateText(layout.TextBlocks[j].Text, opts.Translator, unit)
		if err != nil {
			return fmt.Errorf("block %d: %w", j, err)
		}
		layout.TextBlocks[j].Text = translated
	}
	return nil
}

// RenderLayout はPageLayoutからPageを生成
func RenderLayout(doc *Document, layout *PageLayout, opts PDFTranslatorOptions) (*Page, error) {
	customSize := PageSize{Width: layout.Width, Height: layout.Height}
	page := doc.AddPage(customSize, Portrait)

	// まずグラフィックス操作（罫線、矩形、パス等）を描画
	if len(layout.GraphicsOperations) > 0 {
		page.WriteRawContent(layout.GraphicsOperations)
	}

	contentBlocks := layout.SortedContentBlocks()

	for _, block := range contentBlocks {
		switch block.Type() {
		case ContentBlockTypeImage:
			if opts.KeepImages {
				renderImageBlock(page, block, layout)
			}
		case ContentBlockTypeText:
			if opts.KeepLayout {
				if err := renderTextBlock(page, block, layout, opts); err != nil {
					return nil, err
				}
			}
		}
	}

	return page, nil
}

// renderImageBlock は画像ブロックをページに描画する
func renderImageBlock(page *Page, block ContentBlock, layout *PageLayout) {
	img, ok := block.(ImageBlock)
	if !ok {
		return
	}

	pdfImage, err := loadImageFromImageInfo(img.ImageInfo)
	if err != nil {
		return
	}

	drawX, drawY := translate.ValidateImagePosition(
		img.X, img.Y,
		img.PlacedWidth, img.PlacedHeight,
		layout.Width, layout.Height,
	)

	_ = page.DrawImage(pdfImage, drawX, drawY, img.PlacedWidth, img.PlacedHeight)
}

// renderTextBlock はテキストブロックをページに描画する
func renderTextBlock(page *Page, block ContentBlock, _ *PageLayout, opts PDFTranslatorOptions) error {
	textBlock, ok := block.(TextBlock)
	if !ok {
		return nil
	}

	targetFont, fontName, err := selectFont(textBlock, opts)
	if err != nil {
		return err
	}

	fitted, err := text.Fit(textBlock.Text, textBlock.Rect, fontName, opts.FittingOptions, text.DefaultWidthEstimator)
	if err != nil {
		// フィッティングできない場合は最小フォントサイズでクリップ描画
		minSize := opts.FittingOptions.MinFontSize
		if minSize <= 0 {
			minSize = 6.0
		}
		if err := page.SetFont(targetFont, minSize); err != nil {
			return nil
		}
		// 矩形の上端から描画開始し、はみ出さないようにする
		y := textBlock.Rect.Y + textBlock.Rect.Height - minSize
		_ = page.DrawText(textBlock.Text, textBlock.Rect.X, y)
		return nil
	}

	if err := page.SetFont(targetFont, fitted.FontSize); err != nil {
		return nil
	}

	drawFittedLines(page, fitted, textBlock.Rect, fontName, opts.FittingOptions.Alignment)
	return nil
}

// selectFont はテキストブロックに適切なフォントを選択する
func selectFont(textBlock TextBlock, opts PDFTranslatorOptions) (Font, string, error) {
	if translate.IsASCIIOnly(textBlock.Text) {
		return selectASCIIFont(textBlock)
	}
	return selectNonASCIIFont(textBlock, opts)
}

// selectASCIIFont はASCIIテキスト用のフォントを選択する
func selectASCIIFont(textBlock TextBlock) (Font, string, error) {
	if stdFont, ok := mapToStandardFont(textBlock.Font, textBlock.IsBold); ok {
		return stdFont, string(stdFont), nil
	}
	if textBlock.IsBold {
		return FontHelveticaBold, "Helvetica-Bold", nil
	}
	return FontHelvetica, "Helvetica", nil
}

// selectNonASCIIFont は非ASCIIテキスト用のフォントを選択する
func selectNonASCIIFont(textBlock TextBlock, opts PDFTranslatorOptions) (Font, string, error) {
	if opts.TargetFont == nil {
		return nil, "", fmt.Errorf("target font is required for non-ASCII text")
	}
	if textBlock.IsBold && opts.TargetBoldFont != nil {
		return opts.TargetBoldFont, opts.TargetBoldFont.Name(), nil
	}
	return opts.TargetFont, opts.getTargetFontName(), nil
}

// drawFittedLines はフィッティング済みテキストの各行を描画する
func drawFittedLines(page *Page, fitted *FittedText, rect Rectangle, fontName string, alignment Align) {
	y := rect.Y + rect.Height - fitted.LineHeight
	for _, line := range fitted.Lines {
		if line != "" {
			x := calculateLineX(line, fitted.FontSize, fontName, rect, alignment)
			_ = page.DrawText(line, x, y)
		}
		y -= fitted.LineHeight
	}
}

// calculateLineX はアラインメントに応じてテキストのX座標を計算する
func calculateLineX(line string, fontSize float64, fontName string, rect Rectangle, alignment Align) float64 {
	switch alignment {
	case text.AlignCenter:
		lineWidth := estimateTextWidth(line, fontSize, fontName)
		return rect.X + (rect.Width-lineWidth)/2
	case text.AlignRight:
		lineWidth := estimateTextWidth(line, fontSize, fontName)
		return rect.X + rect.Width - lineWidth
	default:
		return rect.X
	}
}

// mapToStandardFont はPDFフォント名をStandardFontにマッピング
func mapToStandardFont(fontName string, isBold bool) (StandardFont, bool) {
	stdFont, ok := font.MapToStandardFont(fontName, isBold)
	if ok {
		return StandardFont(stdFont), true
	}
	return "", false
}

// loadImageFromImageInfo はImageInfoからImageを作成
func loadImageFromImageInfo(info ImageInfo) (*Image, error) {
	if len(info.Data) == 0 {
		return nil, fmt.Errorf("image data is empty")
	}

	img := &Image{
		Width:            info.Width,
		Height:           info.Height,
		Data:             info.Data,
		ColorSpace:       info.ColorSpace,
		BitsPerComponent: info.BitsPerComp,
		Filter:           info.Filter,
	}

	if img.ColorSpace == "" {
		img.ColorSpace = "DeviceRGB"
	}
	if img.BitsPerComponent == 0 {
		img.BitsPerComponent = 8
	}
	if img.Filter == "" {
		if len(info.Data) >= 2 && info.Data[0] == 0x78 {
			img.Filter = "FlateDecode"
		}
	}

	return img, nil
}

// translateText はテキストを翻訳する
func translateText(text string, translator Translator, unit TranslateUnit) (string, error) {
	internalUnit := translate.TranslateUnit(unit)
	return translate.TranslateText(text, translator, internalUnit)
}

// TranslateTextBlocks はTextBlocksのテキストを翻訳
func TranslateTextBlocks(blocks []TextBlock, translator Translator) error {
	if translator == nil {
		return fmt.Errorf("translator is nil")
	}

	for i := range blocks {
		translated, err := translator.Translate(blocks[i].Text)
		if err != nil {
			return fmt.Errorf("translation failed for block %d: %w", i, err)
		}
		blocks[i].Text = translated
	}

	return nil
}
