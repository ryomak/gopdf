package gopdf

import (
	"bytes"
	"os"
	"runtime"
	"strings"
	"testing"
)

// getJapaneseFontPath returns a path to a Japanese TTF font for testing
func getJapaneseFontPath() string {
	switch runtime.GOOS {
	case "linux":
		paths := []string{
			"/usr/share/fonts/opentype/ipafont-gothic/ipag.ttf",
			"/usr/share/fonts/truetype/fonts-japanese-gothic.ttf",
		}
		for _, p := range paths {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	case "darwin":
		paths := []string{
			"/System/Library/Fonts/ヒラギノ角ゴシック W3.ttc",
			"/System/Library/Fonts/Supplemental/Osaka.ttf",
		}
		for _, p := range paths {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	}
	return ""
}

// createEnglishTestPDF creates a simple English PDF for translation testing
func createEnglishTestPDF(t *testing.T) string {
	t.Helper()
	doc := New()
	page := doc.AddPage(PageSizeA4, Portrait)

	if err := page.SetFont(FontHelvetica, 18); err != nil {
		t.Fatalf("SetFont failed: %v", err)
	}
	_ = page.DrawText("Introduction to Go Programming", 50, 750)

	if err := page.SetFont(FontHelvetica, 12); err != nil {
		t.Fatalf("SetFont failed: %v", err)
	}
	_ = page.DrawText("Go is a statically typed programming language.", 50, 700)
	_ = page.DrawText("It was designed at Google.", 50, 680)
	_ = page.DrawText("Go is efficient and easy to learn.", 50, 660)

	tmpfile, err := os.CreateTemp("", "english_test*.pdf")
	if err != nil {
		t.Fatalf("CreateTemp failed: %v", err)
	}
	defer tmpfile.Close()

	if err := doc.WriteTo(tmpfile); err != nil {
		t.Fatalf("WriteTo failed: %v", err)
	}
	return tmpfile.Name()
}

// mockEnToJaTranslator is a dictionary-based English-to-Japanese translator for testing
type mockEnToJaTranslator struct{}

func (m *mockEnToJaTranslator) Translate(text string) (string, error) {
	dict := map[string]string{
		"Introduction to Go Programming":              "Goプログラミング入門",
		"Go is a statically typed programming language.": "Goは静的型付けプログラミング言語です。",
		"It was designed at Google.":                   "GoogleでDesignされました。",
		"Go is efficient and easy to learn.":           "Goは効率的で学びやすいです。",
	}
	if translated, ok := dict[strings.TrimSpace(text)]; ok {
		return translated, nil
	}
	// 辞書にない場合はそのまま返す
	return text, nil
}

func TestTranslatePDF_EnglishToJapanese(t *testing.T) {
	jpFontPath := getJapaneseFontPath()
	if jpFontPath == "" {
		t.Skip("No Japanese font available on this system")
	}

	jpFont, err := LoadTTF(jpFontPath)
	if err != nil {
		t.Fatalf("Failed to load Japanese font: %v", err)
	}

	// 英語PDFを生成
	inputPath := createEnglishTestPDF(t)
	defer os.Remove(inputPath)

	outputPath := inputPath + "_ja.pdf"
	defer os.Remove(outputPath)

	// 翻訳実行
	opts := PDFTranslatorOptions{
		Translator:    &mockEnToJaTranslator{},
		TargetFont:    jpFont,
		KeepImages:    true,
		KeepLayout:    true,
		TranslateUnit: TranslateUnitBlock,
		FittingOptions: DefaultFitOptions(),
	}

	err = TranslatePDF(inputPath, outputPath, opts)
	if err != nil {
		t.Fatalf("TranslatePDF failed: %v", err)
	}

	// 出力ファイルの検証
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("Output file not found: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("Output PDF is empty")
	}

	// 生成したPDFを読み込んで検証
	reader, err := Open(outputPath)
	if err != nil {
		t.Fatalf("Failed to open translated PDF: %v", err)
	}
	defer reader.Close()

	if reader.PageCount() != 1 {
		t.Errorf("PageCount = %d, want 1", reader.PageCount())
	}

	t.Logf("Translation succeeded: %s (%d bytes)", outputPath, info.Size())
}

func TestTranslatePDFToWriter_EnglishToJapanese(t *testing.T) {
	jpFontPath := getJapaneseFontPath()
	if jpFontPath == "" {
		t.Skip("No Japanese font available on this system")
	}

	jpFont, err := LoadTTF(jpFontPath)
	if err != nil {
		t.Fatalf("Failed to load Japanese font: %v", err)
	}

	// 英語PDFを生成してバッファに書き込み
	doc := New()
	page := doc.AddPage(PageSizeA4, Portrait)
	_ = page.SetFont(FontHelvetica, 14)
	_ = page.DrawText("Hello World", 50, 750)
	_ = page.DrawText("This is a test document.", 50, 720)

	var inputBuf bytes.Buffer
	if err := doc.WriteTo(&inputBuf); err != nil {
		t.Fatalf("WriteTo failed: %v", err)
	}

	// 翻訳実行（io.Writer版）
	var outputBuf bytes.Buffer
	opts := PDFTranslatorOptions{
		Translator: TranslateFunc(func(text string) (string, error) {
			dict := map[string]string{
				"Hello World":                "こんにちは世界",
				"This is a test document.":   "これはテストドキュメントです。",
			}
			if v, ok := dict[strings.TrimSpace(text)]; ok {
				return v, nil
			}
			return text, nil
		}),
		TargetFont:     jpFont,
		KeepImages:     true,
		KeepLayout:     true,
		FittingOptions: DefaultFitOptions(),
	}

	inputReader := bytes.NewReader(inputBuf.Bytes())
	err = TranslatePDFToWriter(inputReader, &outputBuf, opts)
	if err != nil {
		t.Fatalf("TranslatePDFToWriter failed: %v", err)
	}

	if outputBuf.Len() == 0 {
		t.Fatal("Output PDF is empty")
	}

	// 出力PDFを読み込んで検証
	reader, err := OpenReader(bytes.NewReader(outputBuf.Bytes()))
	if err != nil {
		t.Fatalf("Failed to open translated PDF: %v", err)
	}
	defer reader.Close()

	if reader.PageCount() != 1 {
		t.Errorf("PageCount = %d, want 1", reader.PageCount())
	}

	t.Logf("Writer translation succeeded: %d bytes", outputBuf.Len())
}

func TestTranslatePDF_TranslateUnits(t *testing.T) {
	jpFontPath := getJapaneseFontPath()
	if jpFontPath == "" {
		t.Skip("No Japanese font available on this system")
	}

	jpFont, err := LoadTTF(jpFontPath)
	if err != nil {
		t.Fatalf("Failed to load Japanese font: %v", err)
	}

	tests := []struct {
		name string
		unit TranslateUnit
	}{
		{"Block", TranslateUnitBlock},
		{"Line", TranslateUnitLine},
		{"Sentence", TranslateUnitSentence},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputPath := createEnglishTestPDF(t)
			defer os.Remove(inputPath)

			outputPath := inputPath + "_" + tt.name + ".pdf"
			defer os.Remove(outputPath)

			opts := PDFTranslatorOptions{
				Translator: TranslateFunc(func(text string) (string, error) {
					// 簡易的に日本語文字列を返す
					return "翻訳済み: " + text, nil
				}),
				TargetFont:     jpFont,
				KeepImages:     true,
				KeepLayout:     true,
				TranslateUnit:  tt.unit,
				FittingOptions: DefaultFitOptions(),
			}

			err := TranslatePDF(inputPath, outputPath, opts)
			if err != nil {
				t.Fatalf("TranslatePDF with unit %s failed: %v", tt.name, err)
			}

			info, err := os.Stat(outputPath)
			if err != nil {
				t.Fatalf("Output file not found: %v", err)
			}
			if info.Size() == 0 {
				t.Fatalf("Output PDF is empty for unit %s", tt.name)
			}

			t.Logf("Unit %s: %d bytes", tt.name, info.Size())
		})
	}
}

func TestTranslatePDF_WithFunctionalOptions(t *testing.T) {
	jpFontPath := getJapaneseFontPath()
	if jpFontPath == "" {
		t.Skip("No Japanese font available on this system")
	}

	jpFont, err := LoadTTF(jpFontPath)
	if err != nil {
		t.Fatalf("Failed to load Japanese font: %v", err)
	}

	inputPath := createEnglishTestPDF(t)
	defer os.Remove(inputPath)

	outputPath := inputPath + "_functional.pdf"
	defer os.Remove(outputPath)

	opts := NewTranslatorOptions(
		WithTranslatorFunc(TranslateFunc(func(text string) (string, error) {
			return "日本語テスト", nil
		})),
		WithTranslatorTargetFont(jpFont),
		WithTranslatorKeepImages(true),
		WithTranslatorKeepLayout(true),
		WithTranslatorUnit(TranslateUnitSentence),
		WithTranslatorFittingOptions(NewFitOptions(
			WithFitMaxFontSize(16),
			WithFitMinFontSize(6),
			WithFitAllowShrink(true),
		)),
	)

	err = TranslatePDF(inputPath, outputPath, opts)
	if err != nil {
		t.Fatalf("TranslatePDF with functional options failed: %v", err)
	}

	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("Output file not found: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("Output PDF is empty")
	}

	t.Logf("Functional options translation succeeded: %d bytes", info.Size())
}
