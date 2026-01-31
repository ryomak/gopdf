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
	TranslateByLine bool          // 行単位で翻訳（デフォルト: false）- 非推奨、TranslateUnitを使用
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

// getTranslateUnit は翻訳単位を取得（後方互換性のためTranslateByLineも考慮）
func (opts PDFTranslatorOptions) getTranslateUnit() TranslateUnit {
	// 新しいTranslateUnitが設定されていればそれを使用
	if opts.TranslateUnit != TranslateUnitBlock {
		return opts.TranslateUnit
	}
	// 後方互換性: TranslateByLineがtrueならTranslateUnitLineを返す
	if opts.TranslateByLine {
		return TranslateUnitLine
	}
	return TranslateUnitBlock
}

// DefaultPDFTranslatorOptions はデフォルトのオプション
// fontNameは省略可（空文字の場合、targetFontから自動取得）
func DefaultPDFTranslatorOptions(targetFont Font) PDFTranslatorOptions {
	return PDFTranslatorOptions{
		Translator:     nil, // ユーザーが設定する必要がある
		TargetFont:     targetFont,
		TargetFontName: "", // targetFontから自動取得
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
	// 1. 元PDFを読み込み
	reader, err := Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input PDF: %w", err)
	}
	defer reader.Close()

	// 2. 新しいPDFドキュメントを作成
	doc := New()

	// 3. 各ページを処理
	pageCount := reader.PageCount()
	for i := 0; i < pageCount; i++ {
		layout, err := reader.ExtractPageLayout(i)
		if err != nil {
			return fmt.Errorf("failed to extract layout from page %d: %w", i, err)
		}

		// 4. テキストを翻訳
		if opts.Translator != nil {
			for j := range layout.TextBlocks {
				translated, err := translateText(layout.TextBlocks[j].Text, opts.Translator, opts.getTranslateUnit())
				if err != nil {
					return fmt.Errorf("translation failed on page %d, block %d: %w", i, j, err)
				}
				layout.TextBlocks[j].Text = translated
			}
		}

		// 5. ページを生成
		_, err = RenderLayout(doc, layout, opts)
		if err != nil {
			return fmt.Errorf("failed to render page %d: %w", i, err)
		}
	}

	// 6. 出力
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	return doc.WriteTo(file)
}

// TranslatePDFToWriter はPDFを翻訳してWriterに出力
func TranslatePDFToWriter(input io.ReadSeeker, output io.Writer, opts PDFTranslatorOptions) error {
	// 1. 元PDFを読み込み
	reader, err := OpenReader(input)
	if err != nil {
		return fmt.Errorf("failed to open input PDF: %w", err)
	}
	defer reader.Close()

	// 2. 新しいPDFドキュメントを作成
	doc := New()

	// 3. 各ページを処理
	pageCount := reader.PageCount()
	for i := 0; i < pageCount; i++ {
		layout, err := reader.ExtractPageLayout(i)
		if err != nil {
			return fmt.Errorf("failed to extract layout from page %d: %w", i, err)
		}

		// 4. テキストを翻訳
		if opts.Translator != nil {
			for j := range layout.TextBlocks {
				translated, err := translateText(layout.TextBlocks[j].Text, opts.Translator, opts.getTranslateUnit())
				if err != nil {
					return fmt.Errorf("translation failed on page %d, block %d: %w", i, j, err)
				}
				layout.TextBlocks[j].Text = translated
			}
		}

		// 5. ページを生成
		_, err = RenderLayout(doc, layout, opts)
		if err != nil {
			return fmt.Errorf("failed to render page %d: %w", i, err)
		}
	}

	// 6. 出力
	return doc.WriteTo(output)
}

// RenderLayout はPageLayoutからPageを生成
func RenderLayout(doc *Document, layout *PageLayout, opts PDFTranslatorOptions) (*Page, error) {
	// カスタムサイズでページを追加
	customSize := PageSize{Width: layout.Width, Height: layout.Height}
	page := doc.AddPage(customSize, Portrait)

	// ContentBlocksを使用して、画像とテキストを正しい順序で描画
	// 設計書: docs/render_layout_order_issue.md
	// 注: 座標はExtractPageLayoutで既に標準座標系に変換済み
	contentBlocks := layout.SortedContentBlocks()

	for _, block := range contentBlocks {
		switch block.Type() {
		case ContentBlockTypeImage:
			if opts.KeepImages {
				// 画像を描画
				img, ok := block.(ImageBlock)
				if !ok {
					continue
				}
				pdfImage, err := loadImageFromImageInfo(img.ImageInfo)
				if err != nil {
					// 画像の読み込みに失敗しても続行
					continue
				}

				// 画像座標を検証し、異常な場合はフォールバック
				drawX, drawY := translate.ValidateImagePosition(
					img.X, img.Y,
					img.PlacedWidth, img.PlacedHeight,
					layout.Width, layout.Height,
				)

				if err := page.DrawImage(pdfImage, drawX, drawY, img.PlacedWidth, img.PlacedHeight); err != nil {
					// 画像の描画に失敗しても続行
					continue
				}
			}

		case ContentBlockTypeText:
			if opts.KeepLayout {
				textBlock, ok := block.(TextBlock)
				if !ok {
					continue
				}

				// フォントを選択
				// 1. ASCIIのみのテキストは元のフォントを保持
				// 2. 非ASCIIテキストはTargetFont（TTFフォント等）を使用
				var targetFont Font
				var fontName string

				if translate.IsASCIIOnly(textBlock.Text) {
					// 元のフォントをStandardFontにマッピング
					if stdFont, ok := mapToStandardFont(textBlock.Font, textBlock.IsBold); ok {
						targetFont = stdFont
						fontName = string(stdFont)
					} else {
						// マッピングできない場合はデフォルトのHelvetica系を使用
						if textBlock.IsBold {
							targetFont = FontHelveticaBold
							fontName = "Helvetica-Bold"
						} else {
							targetFont = FontHelvetica
							fontName = "Helvetica"
						}
					}
				} else {
					// 非ASCIIテキストはTargetFontを使用
					if opts.TargetFont == nil {
						return nil, fmt.Errorf("target font is required for non-ASCII text")
					}
					// Boldの場合はTargetBoldFontを使用（設定されていれば）
					if textBlock.IsBold && opts.TargetBoldFont != nil {
						targetFont = opts.TargetBoldFont
						fontName = opts.TargetBoldFont.Name()
					} else {
						targetFont = opts.TargetFont
						fontName = opts.getTargetFontName()
					}
				}

				// テキストをフィッティング
				fitted, err := text.Fit(textBlock.Text, textBlock.Rect, fontName, opts.FittingOptions, text.DefaultWidthEstimator)
				if err != nil {
					// フィッティングできない場合は元のサイズを使用
					if err := page.SetFont(targetFont, textBlock.FontSize); err != nil {
						continue
					}
					// 適切な描画メソッドを使用
					_ = drawPageText(page, targetFont, textBlock.Text, textBlock.Rect.X, textBlock.Rect.Y)
					continue
				}

				// 複数行を描画
				if err := page.SetFont(targetFont, fitted.FontSize); err != nil {
					continue
				}
				// 上から下に描画（Y座標が大きい方から小さい方へ）
				y := textBlock.Rect.Y + textBlock.Rect.Height - fitted.LineHeight
				for _, line := range fitted.Lines {
					if line != "" {
						x := textBlock.Rect.X
						// アラインメントに応じてX座標を調整
						if opts.FittingOptions.Alignment == text.AlignCenter {
							lineWidth := estimateTextWidth(line, fitted.FontSize, fontName)
							x = textBlock.Rect.X + (textBlock.Rect.Width-lineWidth)/2
						} else if opts.FittingOptions.Alignment == text.AlignRight {
							lineWidth := estimateTextWidth(line, fitted.FontSize, fontName)
							x = textBlock.Rect.X + textBlock.Rect.Width - lineWidth
						}
						// 適切な描画メソッドを使用
						_ = drawPageText(page, targetFont, line, x, y)
					}
					y -= fitted.LineHeight
				}
			}
		}
	}

	return page, nil
}

// mapToStandardFont はPDFフォント名をStandardFontにマッピング
// マッピングできない場合は空文字列とfalseを返す
func mapToStandardFont(fontName string, isBold bool) (StandardFont, bool) {
	stdFont, ok := font.MapToStandardFont(fontName, isBold)
	if ok {
		return StandardFont(stdFont), true
	}
	return "", false
}

// drawPageText はページにテキストを描画する
// DrawTextが自動的にフォントタイプを判定するため、常にDrawTextを使用
func drawPageText(page *Page, f Font, s string, x, y float64) error {
	return page.DrawText(s, x, y)
}

// loadImageFromImageInfo はImageInfoからImageを作成
// PDFから抽出されたImageInfoは既に必要な情報を持っているため、
// 直接Imageオブジェクトを構築する
func loadImageFromImageInfo(info ImageInfo) (*Image, error) {
	if len(info.Data) == 0 {
		return nil, fmt.Errorf("image data is empty")
	}

	// ImageInfoからImageを直接作成
	img := &Image{
		Width:            info.Width,
		Height:           info.Height,
		Data:             info.Data,
		ColorSpace:       info.ColorSpace,
		BitsPerComponent: info.BitsPerComp,
		Filter:           info.Filter,
	}

	// デフォルト値の設定
	if img.ColorSpace == "" {
		img.ColorSpace = "DeviceRGB"
	}
	if img.BitsPerComponent == 0 {
		img.BitsPerComponent = 8
	}
	if img.Filter == "" {
		// データがzlib圧縮されているか確認（0x78で始まる場合）
		if len(info.Data) >= 2 && info.Data[0] == 0x78 {
			img.Filter = "FlateDecode"
		}
	}

	return img, nil
}

// translateText はテキストを翻訳する
// translateByLine が true の場合、テキストを行単位で分割して翻訳し、再結合する
func translateText(text string, translator Translator, unit TranslateUnit) (string, error) {
	// TranslateUnitをinternal packageの型に変換
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
