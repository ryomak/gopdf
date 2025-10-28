package gopdf

// RubyText はルビ（ふりがな）付きテキスト
type RubyText struct {
	Base string // 親文字（漢字など）
	Ruby string // ルビテキスト（ひらがななど）
}

// RubyAlignment はルビの配置方法
type RubyAlignment int

const (
	RubyAlignCenter RubyAlignment = iota // 中央揃え（デフォルト）
	RubyAlignLeft                        // 左揃え
	RubyAlignRight                       // 右揃え
)

// RubyStyle はルビのスタイル設定
type RubyStyle struct {
	Alignment   RubyAlignment // 配置方法
	Offset      float64       // 親文字との間隔（pt）
	SizeRatio   float64       // 親文字に対するサイズ比率（0.0-1.0）
	CopyMode    RubyCopyMode  // コピー時の動作
}

// RubyCopyMode はPDFからテキストをコピーする時の動作
type RubyCopyMode int

const (
	RubyCopyBase RubyCopyMode = iota // 親文字のみコピー（デフォルト）
	RubyCopyRuby                     // ルビのみコピー
	RubyCopyBoth                     // 両方コピー（親文字(ルビ)形式）
)

// DefaultRubyStyle はデフォルトのルビスタイル
func DefaultRubyStyle() RubyStyle {
	return RubyStyle{
		Alignment: RubyAlignCenter,
		Offset:    1.0,
		SizeRatio: 0.5, // 親文字の50%サイズ
		CopyMode:  RubyCopyBase,
	}
}

// NewRubyText はRubyTextを作成する
func NewRubyText(base, ruby string) RubyText {
	return RubyText{
		Base: base,
		Ruby: ruby,
	}
}

// NewRubyTextPairs は複数のルビテキストを作成する
func NewRubyTextPairs(pairs ...string) []RubyText {
	if len(pairs)%2 != 0 {
		// 奇数の場合は最後の要素を無視
		pairs = pairs[:len(pairs)-1]
	}

	result := make([]RubyText, 0, len(pairs)/2)
	for i := 0; i < len(pairs); i += 2 {
		result = append(result, RubyText{
			Base: pairs[i],
			Ruby: pairs[i+1],
		})
	}
	return result
}
