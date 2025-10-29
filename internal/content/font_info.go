package content

import (
	"fmt"

	"github.com/ryomak/gopdf/internal/core"
	"github.com/ryomak/gopdf/internal/reader"
)

// FontInfo はフォント情報を保持する
type FontInfo struct {
	Name          string
	ToUnicodeCMap *ToUnicodeCMap // nilの場合は通常のエンコーディングを使用
}

// FontManager はページ内のフォント情報を管理する
type FontManager struct {
	reader *reader.Reader
	fonts  map[string]*FontInfo // フォント名 -> FontInfo のキャッシュ
}

// NewFontManager は新しいFontManagerを作成する
func NewFontManager(r *reader.Reader) *FontManager {
	return &FontManager{
		reader: r,
		fonts:  make(map[string]*FontInfo),
	}
}

// GetFont はフォント名から FontInfo を取得する
// pageResources はページの /Resources ディクショナリ
func (fm *FontManager) GetFont(fontName string, pageResources core.Dictionary) (*FontInfo, error) {
	// キャッシュをチェック
	if info, ok := fm.fonts[fontName]; ok {
		return info, nil
	}

	// フォント情報を読み込む
	info, err := fm.loadFontInfo(fontName, pageResources)
	if err != nil {
		return nil, err
	}

	// キャッシュに保存
	fm.fonts[fontName] = info
	return info, nil
}

// loadFontInfo はフォント情報を PDF から読み込む
func (fm *FontManager) loadFontInfo(fontName string, pageResources core.Dictionary) (*FontInfo, error) {
	info := &FontInfo{
		Name: fontName,
	}

	// /Resources/Font からフォント辞書を取得
	fontDict, err := fm.getFontDictionary(fontName, pageResources)
	if err != nil {
		// フォント辞書が見つからない場合は ToUnicode なしで返す
		return info, nil
	}

	// ToUnicode CMap を抽出
	toUnicodeCMap, err := fm.extractToUnicodeCMap(fontDict)
	if err != nil {
		// ToUnicode の抽出に失敗しても、フォント情報自体は返す
		// 従来のエンコーディングで処理される
		return info, nil
	}

	info.ToUnicodeCMap = toUnicodeCMap
	return info, nil
}

// getFontDictionary は /Resources/Font からフォント辞書を取得する
func (fm *FontManager) getFontDictionary(fontName string, pageResources core.Dictionary) (core.Dictionary, error) {
	if pageResources == nil {
		return nil, fmt.Errorf("page resources is nil")
	}

	// /Resources/Font を取得
	fontResources, ok := pageResources["Font"]
	if !ok {
		return nil, fmt.Errorf("no Font in Resources")
	}

	// 間接参照を解決
	if ref, ok := fontResources.(*core.Reference); ok {
		var err error
		fontResources, err = fm.reader.ResolveReference(ref)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve Font resources: %w", err)
		}
	}

	fontResourcesDict, ok := fontResources.(core.Dictionary)
	if !ok {
		return nil, fmt.Errorf("Font resources is not a Dictionary")
	}

	// フォント名でフォント辞書を取得（core.Name型に変換）
	fontObj, ok := fontResourcesDict[core.Name(fontName)]
	if !ok {
		return nil, fmt.Errorf("font %s not found in Font resources", fontName)
	}

	// 間接参照を解決
	if ref, ok := fontObj.(*core.Reference); ok {
		var err error
		fontObj, err = fm.reader.ResolveReference(ref)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve font object: %w", err)
		}
	}

	fontDict, ok := fontObj.(core.Dictionary)
	if !ok {
		return nil, fmt.Errorf("font object is not a Dictionary")
	}

	return fontDict, nil
}

// extractToUnicodeCMap はフォント辞書から ToUnicode CMap を抽出する
func (fm *FontManager) extractToUnicodeCMap(fontDict core.Dictionary) (*ToUnicodeCMap, error) {
	// /ToUnicode キーをチェック
	toUnicodeObj, ok := fontDict["ToUnicode"]
	if !ok {
		return nil, fmt.Errorf("no ToUnicode in font dictionary")
	}

	// 間接参照を解決
	if ref, ok := toUnicodeObj.(*core.Reference); ok {
		var err error
		toUnicodeObj, err = fm.reader.ResolveReference(ref)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve ToUnicode: %w", err)
		}
	}

	// ストリームオブジェクトであることを確認
	stream, ok := toUnicodeObj.(*core.Stream)
	if !ok {
		return nil, fmt.Errorf("ToUnicode is not a Stream")
	}

	// ストリームデータを取得
	data := stream.Data

	// CMap をパース
	cmap, err := ParseToUnicodeCMap(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ToUnicode CMap: %w", err)
	}

	return cmap, nil
}
