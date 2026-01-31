package gopdf

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/ryomak/gopdf/internal/content/text"
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
	TargetFontName  string        // フォント名（estimateTextWidth用）
	FittingOptions  FitOptions    // テキストフィッティングオプション（FitOptions）
	KeepImages      bool          // 画像を保持（デフォルト: true）
	KeepLayout      bool          // レイアウトを保持（デフォルト: true）
	TranslateByLine bool          // 行単位で翻訳（デフォルト: false）- 非推奨、TranslateUnitを使用
	TranslateUnit   TranslateUnit // 翻訳単位（デフォルト: TranslateUnitBlock）
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
func DefaultPDFTranslatorOptions(targetFont Font, fontName string) PDFTranslatorOptions {
	return PDFTranslatorOptions{
		Translator:     nil, // ユーザーが設定する必要がある
		TargetFont:     targetFont,
		TargetFontName: fontName,
		FittingOptions: DefaultFitOptions(),
		KeepImages:     true,
		KeepLayout:     true,
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
				drawX, drawY := validateImagePosition(
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

				if isASCIIOnly(textBlock.Text) {
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
					targetFont = opts.TargetFont
					fontName = opts.TargetFontName
				}

				// テキストをフィッティング
				fitted, err := text.Fit(textBlock.Text, textBlock.Rect, fontName, opts.FittingOptions, text.DefaultWidthEstimator)
				if err != nil {
					// フィッティングできない場合は元のサイズを使用
					if err := setPageFont(page, targetFont, textBlock.FontSize); err != nil {
						continue
					}
					// 適切な描画メソッドを使用
					_ = drawPageText(page, targetFont, textBlock.Text, textBlock.Rect.X, textBlock.Rect.Y)
					continue
				}

				// 複数行を描画
				if err := setPageFont(page, targetFont, fitted.FontSize); err != nil {
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

// validateImagePosition は画像座標を検証し、異常な場合はフォールバック位置を返す
func validateImagePosition(x, y, width, height, pageWidth, pageHeight float64) (newX, newY float64) {
	const maxOffset = 10000.0

	// 異常値の検出
	if x < -maxOffset || x > pageWidth+maxOffset ||
		y < -maxOffset || y > pageHeight+maxOffset {
		// フォールバック: ページ中央上部に配置
		newX = (pageWidth - width) / 2
		if newX < 0 {
			newX = 0
		}
		newY = pageHeight - height - 50
		if newY < 0 {
			newY = 0
		}
		return newX, newY
	}

	return x, y
}

// isASCIIOnly はテキストがASCII文字のみかどうかを判定
func isASCIIOnly(s string) bool {
	for _, r := range s {
		if r > 127 {
			return false
		}
	}
	return true
}

// standardFontMap はPDFフォント名からStandardFontへのマッピング
var standardFontMap = map[string]StandardFont{
	// Helvetica variants
	"Helvetica":            FontHelvetica,
	"Helvetica-Bold":       FontHelveticaBold,
	"Helvetica-Oblique":    FontHelveticaOblique,
	"Helvetica-BoldOblique": FontHelveticaBoldOblique,
	// Times variants
	"Times-Roman":      FontTimesRoman,
	"Times-Bold":       FontTimesBold,
	"Times-Italic":     FontTimesItalic,
	"Times-BoldItalic": FontTimesBoldItalic,
	// Courier variants
	"Courier":            FontCourier,
	"Courier-Bold":       FontCourierBold,
	"Courier-Oblique":    FontCourierOblique,
	"Courier-BoldOblique": FontCourierBoldOblique,
	// Symbol fonts
	"Symbol":       FontSymbol,
	"ZapfDingbats": FontZapfDingbats,
}

// mapToStandardFont はPDFフォント名をStandardFontにマッピング
// マッピングできない場合はnilを返す
func mapToStandardFont(fontName string, isBold bool) (StandardFont, bool) {
	// 直接マッチ
	if stdFont, ok := standardFontMap[fontName]; ok {
		return stdFont, true
	}

	// 部分マッチでフォントファミリーを推測
	// 例: "BCDEEE+Helvetica-Bold" -> "Helvetica-Bold"
	for name, stdFont := range standardFontMap {
		if len(fontName) >= len(name) && fontName[len(fontName)-len(name):] == name {
			return stdFont, true
		}
	}

	// フォント名からファミリーを推測
	switch {
	case containsFont(fontName, "Helvetica"):
		if isBold {
			return FontHelveticaBold, true
		}
		return FontHelvetica, true
	case containsFont(fontName, "Times"):
		if isBold {
			return FontTimesBold, true
		}
		return FontTimesRoman, true
	case containsFont(fontName, "Courier"):
		if isBold {
			return FontCourierBold, true
		}
		return FontCourier, true
	case containsFont(fontName, "Symbol"):
		return FontSymbol, true
	case containsFont(fontName, "ZapfDingbats"), containsFont(fontName, "Dingbats"):
		return FontZapfDingbats, true
	}

	return "", false
}

// containsFont はフォント名に指定された名前が含まれているかチェック
func containsFont(fontName, target string) bool {
	// 大文字小文字を無視して部分一致を確認
	for i := 0; i <= len(fontName)-len(target); i++ {
		match := true
		for j := 0; j < len(target); j++ {
			c1 := fontName[i+j]
			c2 := target[j]
			// 大文字小文字を無視
			if c1 >= 'A' && c1 <= 'Z' {
				c1 += 'a' - 'A'
			}
			if c2 >= 'A' && c2 <= 'Z' {
				c2 += 'a' - 'A'
			}
			if c1 != c2 {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// setPageFont はページにフォントを設定する
func setPageFont(page *Page, f Font, size float64) error {
	switch font := f.(type) {
	case StandardFont:
		return page.SetFont(font, size)
	case *TTFFont:
		return page.SetTTFFont(font, size)
	default:
		return fmt.Errorf("unsupported font type: %T", f)
	}
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
	switch unit {
	case TranslateUnitBlock:
		// ブロック全体を翻訳
		return translator.Translate(text)

	case TranslateUnitLine:
		// 行単位で翻訳
		return translateByDelimiter(text, translator, "\n", "\n")

	case TranslateUnitSentence:
		// 文単位で翻訳
		// 改行を一時的にスペースに変換して文を連結
		normalized := strings.ReplaceAll(text, "\n", " ")
		// 文末記号で分割して翻訳
		translated, err := translateBySentence(normalized, translator)
		if err != nil {
			return "", err
		}
		return translated, nil

	default:
		return translator.Translate(text)
	}
}

// translateByDelimiter は指定した区切り文字で分割して翻訳
func translateByDelimiter(text string, translator Translator, splitDelim, joinDelim string) (string, error) {
	parts := strings.Split(text, splitDelim)
	translatedParts := make([]string, len(parts))

	for i, part := range parts {
		if part == "" {
			translatedParts[i] = ""
			continue
		}
		translated, err := translator.Translate(part)
		if err != nil {
			return "", err
		}
		translatedParts[i] = translated
	}

	return strings.Join(translatedParts, joinDelim), nil
}

// sentenceEndPattern は文末を検出する正規表現
var sentenceEndPattern = regexp.MustCompile(`([.!?。！？]+)\s*`)

// translateBySentence は文単位で翻訳
func translateBySentence(text string, translator Translator) (string, error) {
	// 文末記号で分割（記号も保持）
	// "Hello. World!" -> ["Hello", ".", " ", "World", "!"]
	parts := sentenceEndPattern.Split(text, -1)
	delimiters := sentenceEndPattern.FindAllString(text, -1)

	if len(parts) == 0 {
		return translator.Translate(text)
	}

	var result strings.Builder
	for i, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			if i < len(delimiters) {
				result.WriteString(delimiters[i])
			}
			continue
		}

		// 文を翻訳
		translated, err := translator.Translate(part)
		if err != nil {
			return "", err
		}
		result.WriteString(translated)

		// 区切り文字を追加
		if i < len(delimiters) {
			result.WriteString(delimiters[i])
		}
	}

	return result.String(), nil
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
