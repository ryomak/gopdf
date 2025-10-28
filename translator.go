package gopdf

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/ryomak/gopdf/internal/font"
)

// Translator はテキスト翻訳のインターフェース
type Translator interface {
	// Translate はテキストを翻訳する
	Translate(text string) (string, error)
}

// TranslateFunc は関数型Translator
type TranslateFunc func(string) (string, error)

// Translate はTranslateFuncの実装
func (f TranslateFunc) Translate(text string) (string, error) {
	return f(text)
}

// PDFTranslatorOptions は翻訳オプション
type PDFTranslatorOptions struct {
	Translator     Translator    // 翻訳インターフェース
	TargetFont     interface{}   // ターゲット言語のフォント (font.StandardFont or *TTFFont)
	TargetFontName string        // フォント名（estimateTextWidth用）
	FittingOptions FitTextOptions // テキストフィッティングオプション
	KeepImages     bool          // 画像を保持（デフォルト: true）
	KeepLayout     bool          // レイアウトを保持（デフォルト: true）
}

// DefaultPDFTranslatorOptions はデフォルトのオプション
func DefaultPDFTranslatorOptions(targetFont interface{}, fontName string) PDFTranslatorOptions {
	return PDFTranslatorOptions{
		Translator:     nil, // ユーザーが設定する必要がある
		TargetFont:     targetFont,
		TargetFontName: fontName,
		FittingOptions: DefaultFitTextOptions(),
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
				translated, err := opts.Translator.Translate(layout.TextBlocks[j].Text)
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
				translated, err := opts.Translator.Translate(layout.TextBlocks[j].Text)
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

	// 画像を配置
	if opts.KeepImages {
		for _, img := range layout.Images {
			// 画像データからImageを作成
			pdfImage, err := loadImageFromImageInfo(img.ImageInfo)
			if err != nil {
				// 画像の読み込みに失敗しても続行
				continue
			}
			if err := page.DrawImage(pdfImage, img.X, img.Y, img.PlacedWidth, img.PlacedHeight); err != nil {
				// 画像の描画に失敗しても続行
				continue
			}
		}
	}

	// テキストを配置
	if opts.KeepLayout {
		if opts.TargetFont == nil {
			return nil, fmt.Errorf("target font is required")
		}

		for _, block := range layout.TextBlocks {
			// テキストをフィッティング
			fitted, err := FitText(block.Text, block.Bounds, opts.TargetFontName, opts.FittingOptions)
			if err != nil {
				// フィッティングできない場合は元のサイズを使用
				if err := setPageFont(page, opts.TargetFont, block.FontSize); err != nil {
					continue
				}
				// 適切な描画メソッドを使用
				_ = drawPageText(page, opts.TargetFont, block.Text, block.Bounds.X, block.Bounds.Y)
				continue
			}

			// 複数行を描画
			if err := setPageFont(page, opts.TargetFont, fitted.FontSize); err != nil {
				continue
			}
			// 上から下に描画（Y座標が大きい方から小さい方へ）
			y := block.Bounds.Y + block.Bounds.Height - fitted.LineHeight
			for _, line := range fitted.Lines {
				if line != "" {
					x := block.Bounds.X
					// アラインメントに応じてX座標を調整
					if opts.FittingOptions.Alignment == AlignCenter {
						lineWidth := estimateTextWidth(line, fitted.FontSize, opts.TargetFontName)
						x = block.Bounds.X + (block.Bounds.Width-lineWidth)/2
					} else if opts.FittingOptions.Alignment == AlignRight {
						lineWidth := estimateTextWidth(line, fitted.FontSize, opts.TargetFontName)
						x = block.Bounds.X + block.Bounds.Width - lineWidth
					}
					// 適切な描画メソッドを使用
					_ = drawPageText(page, opts.TargetFont, line, x, y)
				}
				y -= fitted.LineHeight
			}
		}
	}

	return page, nil
}

// setPageFont はページにフォントを設定する（型アサーション対応）
func setPageFont(page *Page, fontInterface interface{}, size float64) error {
	// font.StandardFontの場合
	if stdFont, ok := fontInterface.(font.StandardFont); ok {
		return page.SetFont(stdFont, size)
	}
	// *TTFFontの場合
	if ttfFont, ok := fontInterface.(*TTFFont); ok {
		return page.SetTTFFont(ttfFont, size)
	}
	return fmt.Errorf("unsupported font type")
}

// drawPageText はフォントタイプに応じて適切な描画メソッドを呼び出す
func drawPageText(page *Page, fontInterface interface{}, text string, x, y float64) error {
	// *TTFFontの場合
	if _, ok := fontInterface.(*TTFFont); ok {
		return page.DrawTextUTF8(text, x, y)
	}
	// font.StandardFontの場合
	return page.DrawText(text, x, y)
}

// loadImageFromImageInfo はImageInfoからImageを作成
func loadImageFromImageInfo(info ImageInfo) (*Image, error) {
	switch info.Format {
	case ImageFormatJPEG:
		return LoadJPEG(bytes.NewReader(info.Data))
	case ImageFormatPNG:
		// PNGの場合、FlateDecode済みのrawデータの可能性がある
		// とりあえずそのまま読み込みを試みる
		return LoadPNG(bytes.NewReader(info.Data))
	default:
		return nil, fmt.Errorf("unsupported image format: %s", info.Format)
	}
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
