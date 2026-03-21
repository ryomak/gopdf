package content

import (
	"testing"

	"github.com/ryomak/gopdf/internal/core"
)

// TestTextExtractor_Extract はTextExtractorの基本的な抽出をテストする
func TestTextExtractor_Extract(t *testing.T) {
	operations := []Operation{
		{Operator: "BT"},
		{Operator: "Tf", Operands: []core.Object{core.Name("F1"), core.Real(12)}},
		{Operator: "Td", Operands: []core.Object{core.Real(100), core.Real(700)}},
		{Operator: "Tj", Operands: []core.Object{core.String("Hello")}},
		{Operator: "ET"},
	}

	extractor := NewTextExtractor(operations, nil, nil)
	elements, err := extractor.Extract()

	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(elements) != 1 {
		t.Fatalf("Expected 1 element, got %d", len(elements))
	}

	elem := elements[0]
	if elem.Text != "Hello" {
		t.Errorf("Text = %q, want %q", elem.Text, "Hello")
	}
	if elem.X != 100 {
		t.Errorf("X = %f, want 100", elem.X)
	}
	if elem.Y != 700 {
		t.Errorf("Y = %f, want 700", elem.Y)
	}
	if elem.Font != "F1" {
		t.Errorf("Font = %q, want %q", elem.Font, "F1")
	}
	if elem.Size != 12 {
		t.Errorf("Size = %f, want 12", elem.Size)
	}
}

// TestTextExtractor_MultipleTexts は複数のテキストの抽出をテストする
func TestTextExtractor_MultipleTexts(t *testing.T) {
	operations := []Operation{
		{Operator: "BT"},
		{Operator: "Tf", Operands: []core.Object{core.Name("F1"), core.Real(12)}},
		{Operator: "Td", Operands: []core.Object{core.Real(100), core.Real(700)}},
		{Operator: "Tj", Operands: []core.Object{core.String("Hello")}},
		{Operator: "Td", Operands: []core.Object{core.Real(0), core.Real(-14)}},
		{Operator: "Tj", Operands: []core.Object{core.String("World")}},
		{Operator: "ET"},
	}

	extractor := NewTextExtractor(operations, nil, nil)
	elements, err := extractor.Extract()

	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(elements) != 2 {
		t.Fatalf("Expected 2 elements, got %d", len(elements))
	}

	if elements[0].Text != "Hello" {
		t.Errorf("First text = %q, want %q", elements[0].Text, "Hello")
	}
	if elements[1].Text != "World" {
		t.Errorf("Second text = %q, want %q", elements[1].Text, "World")
	}

	// Y座標が下がっていることを確認
	if elements[1].Y >= elements[0].Y {
		t.Errorf("Second text should be below first text")
	}
}

// TestTextExtractor_TJ はTJオペレーターをテストする
// TJ配列内のテキストは1つのTextElementとして結合される
func TestTextExtractor_TJ(t *testing.T) {
	operations := []Operation{
		{Operator: "BT"},
		{Operator: "Tf", Operands: []core.Object{core.Name("F1"), core.Real(12)}},
		{Operator: "Td", Operands: []core.Object{core.Real(100), core.Real(700)}},
		{Operator: "TJ", Operands: []core.Object{
			core.Array{core.String("Hello"), core.Integer(-50), core.String("World")},
		}},
		{Operator: "ET"},
	}

	extractor := NewTextExtractor(operations, nil, nil)
	elements, err := extractor.Extract()

	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	// TJ配列は1つのTextElementとして結合される
	if len(elements) != 1 {
		t.Fatalf("Expected 1 element, got %d", len(elements))
	}

	// -50は小さな位置調整なのでスペースは挿入されない（-250以下でスペース挿入）
	if elements[0].Text != "HelloWorld" {
		t.Errorf("Text = %q, want %q", elements[0].Text, "HelloWorld")
	}
}

// TestTextExtractor_TJ_WithSpace はTJオペレーターでスペース挿入をテストする
func TestTextExtractor_TJ_WithSpace(t *testing.T) {
	operations := []Operation{
		{Operator: "BT"},
		{Operator: "Tf", Operands: []core.Object{core.Name("F1"), core.Real(12)}},
		{Operator: "Td", Operands: []core.Object{core.Real(100), core.Real(700)}},
		{Operator: "TJ", Operands: []core.Object{
			// -300は十分大きいのでスペースとして扱われる（-250以下でスペース挿入）
			core.Array{core.String("Hello"), core.Integer(-300), core.String("World")},
		}},
		{Operator: "ET"},
	}

	extractor := NewTextExtractor(operations, nil, nil)
	elements, err := extractor.Extract()

	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(elements) != 1 {
		t.Fatalf("Expected 1 element, got %d", len(elements))
	}

	// -300は大きな位置調整なのでスペースが挿入される
	if elements[0].Text != "Hello World" {
		t.Errorf("Text = %q, want %q", elements[0].Text, "Hello World")
	}
}

// TestTextExtractor_Tm はTmオペレーターをテストする
func TestTextExtractor_Tm(t *testing.T) {
	operations := []Operation{
		{Operator: "BT"},
		{Operator: "Tf", Operands: []core.Object{core.Name("F1"), core.Real(12)}},
		{Operator: "Tm", Operands: []core.Object{
			core.Real(1), core.Real(0), core.Real(0), core.Real(1),
			core.Real(150), core.Real(750),
		}},
		{Operator: "Tj", Operands: []core.Object{core.String("Test")}},
		{Operator: "ET"},
	}

	extractor := NewTextExtractor(operations, nil, nil)
	elements, err := extractor.Extract()

	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(elements) != 1 {
		t.Fatalf("Expected 1 element, got %d", len(elements))
	}

	elem := elements[0]
	if elem.X != 150 {
		t.Errorf("X = %f, want 150", elem.X)
	}
	if elem.Y != 750 {
		t.Errorf("Y = %f, want 750", elem.Y)
	}
}

// TestTextExtractor_TStar はT*オペレーターをテストする
func TestTextExtractor_TStar(t *testing.T) {
	operations := []Operation{
		{Operator: "BT"},
		{Operator: "Tf", Operands: []core.Object{core.Name("F1"), core.Real(12)}},
		{Operator: "TL", Operands: []core.Object{core.Real(14)}}, // Set leading
		{Operator: "Td", Operands: []core.Object{core.Real(100), core.Real(700)}},
		{Operator: "Tj", Operands: []core.Object{core.String("Line 1")}},
		{Operator: "T*"}, // Move to next line
		{Operator: "Tj", Operands: []core.Object{core.String("Line 2")}},
		{Operator: "ET"},
	}

	extractor := NewTextExtractor(operations, nil, nil)
	elements, err := extractor.Extract()

	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(elements) != 2 {
		t.Fatalf("Expected 2 elements, got %d", len(elements))
	}

	// Y座標が下がっていることを確認
	if elements[1].Y >= elements[0].Y {
		t.Errorf("Second line should be below first line")
	}
}

// TestTextExtractor_Quote は'オペレーターをテストする
func TestTextExtractor_Quote(t *testing.T) {
	operations := []Operation{
		{Operator: "BT"},
		{Operator: "Tf", Operands: []core.Object{core.Name("F1"), core.Real(12)}},
		{Operator: "TL", Operands: []core.Object{core.Real(14)}},
		{Operator: "Td", Operands: []core.Object{core.Real(100), core.Real(700)}},
		{Operator: "'", Operands: []core.Object{core.String("Next line")}},
		{Operator: "ET"},
	}

	extractor := NewTextExtractor(operations, nil, nil)
	elements, err := extractor.Extract()

	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(elements) != 1 {
		t.Fatalf("Expected 1 element, got %d", len(elements))
	}

	if elements[0].Text != "Next line" {
		t.Errorf("Text = %q, want %q", elements[0].Text, "Next line")
	}
}

// TestTextExtractor_NoText はテキストがない場合をテストする
func TestTextExtractor_NoText(t *testing.T) {
	operations := []Operation{
		{Operator: "q"},
		{Operator: "Q"},
	}

	extractor := NewTextExtractor(operations, nil, nil)
	elements, err := extractor.Extract()

	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(elements) != 0 {
		t.Errorf("Expected 0 elements, got %d", len(elements))
	}
}

// TestTextElement_Getters はTextElementのGetter関数をテストする
func TestTextElement_Getters(t *testing.T) {
	elem := TextElement{
		Text: "Hello",
		X:    100,
		Y:    200,
		Font: "F1",
		Size: 12,
	}

	if elem.GetX() != 100 {
		t.Errorf("GetX() = %v, want 100", elem.GetX())
	}
	if elem.GetY() != 200 {
		t.Errorf("GetY() = %v, want 200", elem.GetY())
	}
	if elem.GetSize() != 12 {
		t.Errorf("GetSize() = %v, want 12", elem.GetSize())
	}
}

// TestTextExtractor_TD はTDオペレーターをテストする
func TestTextExtractor_TD(t *testing.T) {
	operations := []Operation{
		{Operator: "BT"},
		{Operator: "Tf", Operands: []core.Object{core.Name("F1"), core.Real(12)}},
		{Operator: "Td", Operands: []core.Object{core.Real(100), core.Real(700)}},
		{Operator: "Tj", Operands: []core.Object{core.String("First")}},
		{Operator: "TD", Operands: []core.Object{core.Real(0), core.Real(-20)}},
		{Operator: "Tj", Operands: []core.Object{core.String("Second")}},
		{Operator: "ET"},
	}

	extractor := NewTextExtractor(operations, nil, nil)
	elements, err := extractor.Extract()
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(elements) != 2 {
		t.Fatalf("Expected 2 elements, got %d", len(elements))
	}

	if elements[0].Text != "First" {
		t.Errorf("First text = %q, want %q", elements[0].Text, "First")
	}
	if elements[1].Text != "Second" {
		t.Errorf("Second text = %q, want %q", elements[1].Text, "Second")
	}
	// TD sets leading to -ty, so leading = 20
	// Y should decrease by 20
	if elements[1].Y >= elements[0].Y {
		t.Errorf("Second text Y (%v) should be below first (%v)", elements[1].Y, elements[0].Y)
	}
}

// TestTextExtractor_DoubleQuote は"オペレーターをテストする
func TestTextExtractor_DoubleQuote(t *testing.T) {
	operations := []Operation{
		{Operator: "BT"},
		{Operator: "Tf", Operands: []core.Object{core.Name("F1"), core.Real(12)}},
		{Operator: "TL", Operands: []core.Object{core.Real(14)}},
		{Operator: "Td", Operands: []core.Object{core.Real(100), core.Real(700)}},
		{Operator: "\"", Operands: []core.Object{core.Real(1.5), core.Real(0.5), core.String("Quoted text")}},
		{Operator: "ET"},
	}

	extractor := NewTextExtractor(operations, nil, nil)
	elements, err := extractor.Extract()
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(elements) != 1 {
		t.Fatalf("Expected 1 element, got %d", len(elements))
	}

	if elements[0].Text != "Quoted text" {
		t.Errorf("Text = %q, want %q", elements[0].Text, "Quoted text")
	}
}

// TestTextExtractor_TcTw はTc/Twオペレーターをテストする
func TestTextExtractor_TcTw(t *testing.T) {
	operations := []Operation{
		{Operator: "BT"},
		{Operator: "Tf", Operands: []core.Object{core.Name("F1"), core.Real(12)}},
		{Operator: "Tc", Operands: []core.Object{core.Real(0.5)}},
		{Operator: "Tw", Operands: []core.Object{core.Real(1.0)}},
		{Operator: "Td", Operands: []core.Object{core.Real(100), core.Real(700)}},
		{Operator: "Tj", Operands: []core.Object{core.String("Test")}},
		{Operator: "ET"},
	}

	extractor := NewTextExtractor(operations, nil, nil)
	elements, err := extractor.Extract()
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(elements) != 1 {
		t.Fatalf("Expected 1 element, got %d", len(elements))
	}

	if elements[0].Text != "Test" {
		t.Errorf("Text = %q, want %q", elements[0].Text, "Test")
	}
}

// TestTextExtractor_cm はcmオペレーターをテストする
func TestTextExtractor_cm(t *testing.T) {
	operations := []Operation{
		{Operator: "cm", Operands: []core.Object{
			core.Real(2), core.Real(0), core.Real(0), core.Real(2), core.Real(50), core.Real(100),
		}},
		{Operator: "BT"},
		{Operator: "Tf", Operands: []core.Object{core.Name("F1"), core.Real(12)}},
		{Operator: "Td", Operands: []core.Object{core.Real(100), core.Real(700)}},
		{Operator: "Tj", Operands: []core.Object{core.String("Scaled")}},
		{Operator: "ET"},
	}

	extractor := NewTextExtractor(operations, nil, nil)
	elements, err := extractor.Extract()
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(elements) != 1 {
		t.Fatalf("Expected 1 element, got %d", len(elements))
	}

	// Verify pageLevelCTM was set
	ctm := extractor.GetPageLevelCTM()
	if ctm == nil {
		t.Fatal("GetPageLevelCTM() should not be nil after cm operator")
	}
	if ctm.A != 2 || ctm.E != 50 {
		t.Errorf("PageLevelCTM = %+v, expected A=2, E=50", *ctm)
	}
}

// TestTextExtractor_qQ はq/Q（グラフィックス状態の保存/復元）をテストする
func TestTextExtractor_qQ(t *testing.T) {
	operations := []Operation{
		{Operator: "q"},
		{Operator: "cm", Operands: []core.Object{
			core.Real(2), core.Real(0), core.Real(0), core.Real(2), core.Real(0), core.Real(0),
		}},
		{Operator: "Q"},
		{Operator: "BT"},
		{Operator: "Tf", Operands: []core.Object{core.Name("F1"), core.Real(10)}},
		{Operator: "Td", Operands: []core.Object{core.Real(50), core.Real(500)}},
		{Operator: "Tj", Operands: []core.Object{core.String("After restore")}},
		{Operator: "ET"},
	}

	extractor := NewTextExtractor(operations, nil, nil)
	elements, err := extractor.Extract()
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(elements) != 1 {
		t.Fatalf("Expected 1 element, got %d", len(elements))
	}

	if elements[0].Text != "After restore" {
		t.Errorf("Text = %q, want %q", elements[0].Text, "After restore")
	}
}

// TestTextExtractor_Q_EmptyStack はスタックが空の場合のQをテストする
func TestTextExtractor_Q_EmptyStack(t *testing.T) {
	operations := []Operation{
		{Operator: "Q"}, // No matching q - should not panic
		{Operator: "BT"},
		{Operator: "Tf", Operands: []core.Object{core.Name("F1"), core.Real(10)}},
		{Operator: "Td", Operands: []core.Object{core.Real(50), core.Real(500)}},
		{Operator: "Tj", Operands: []core.Object{core.String("Text")}},
		{Operator: "ET"},
	}

	extractor := NewTextExtractor(operations, nil, nil)
	elements, err := extractor.Extract()
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(elements) != 1 {
		t.Fatalf("Expected 1 element, got %d", len(elements))
	}
}

// TestTextExtractor_GetPageLevelCTM_Nil はcmがない場合をテストする
func TestTextExtractor_GetPageLevelCTM_Nil(t *testing.T) {
	operations := []Operation{
		{Operator: "BT"},
		{Operator: "ET"},
	}

	extractor := NewTextExtractor(operations, nil, nil)
	_, _ = extractor.Extract()

	if extractor.GetPageLevelCTM() != nil {
		t.Error("GetPageLevelCTM() should be nil when no cm operator used")
	}
}

// TestTextExtractor_TJ_WithRealSpacing はTJ配列でReal型の数値をテストする
func TestTextExtractor_TJ_WithRealSpacing(t *testing.T) {
	operations := []Operation{
		{Operator: "BT"},
		{Operator: "Tf", Operands: []core.Object{core.Name("F1"), core.Real(12)}},
		{Operator: "Td", Operands: []core.Object{core.Real(100), core.Real(700)}},
		{Operator: "TJ", Operands: []core.Object{
			core.Array{core.String("Hello"), core.Real(-300), core.String("World")},
		}},
		{Operator: "ET"},
	}

	extractor := NewTextExtractor(operations, nil, nil)
	elements, err := extractor.Extract()
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(elements) != 1 {
		t.Fatalf("Expected 1 element, got %d", len(elements))
	}

	if elements[0].Text != "Hello World" {
		t.Errorf("Text = %q, want %q", elements[0].Text, "Hello World")
	}
}

// TestGetNumber はgetNumber関数をテストする
func TestGetNumber(t *testing.T) {
	tests := []struct {
		name   string
		input  core.Object
		expect float64
	}{
		{"integer", core.Integer(42), 42},
		{"real", core.Real(3.14), 3.14},
		{"string returns 0", core.String("hello"), 0},
		{"nil returns 0", nil, 0},
		{"name returns 0", core.Name("test"), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getNumber(tt.input)
			if got != tt.expect {
				t.Errorf("getNumber(%v) = %v, want %v", tt.input, got, tt.expect)
			}
		})
	}
}

// TestGetString はgetString関数をテストする
func TestGetString(t *testing.T) {
	tests := []struct {
		name   string
		input  core.Object
		expect string
	}{
		{"string", core.String("hello"), "hello"},
		{"name", core.Name("F1"), "F1"},
		{"integer returns empty", core.Integer(42), ""},
		{"nil returns empty", nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getString(tt.input)
			if got != tt.expect {
				t.Errorf("getString(%v) = %q, want %q", tt.input, got, tt.expect)
			}
		})
	}
}

// TestGetTextString はgetTextString関数をテストする
func TestGetTextString(t *testing.T) {
	extractor := NewTextExtractor(nil, nil, nil)

	tests := []struct {
		name   string
		input  core.Object
		expect string
	}{
		{"string", core.String("hello"), "hello"},
		{"name", core.Name("F1"), "F1"},
		{"integer returns empty", core.Integer(42), ""},
		{"nil returns empty", nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractor.getTextString(tt.input)
			if got != tt.expect {
				t.Errorf("getTextString(%v) = %q, want %q", tt.input, got, tt.expect)
			}
		})
	}
}

// TestDetectImageFormat はdetectImageFormat関数をテストする
func TestDetectImageFormat(t *testing.T) {
	tests := []struct {
		name   string
		filter string
		expect ImageFormat
	}{
		{"DCTDecode is JPEG", "DCTDecode", ImageFormatJPEG},
		{"FlateDecode is PNG", "FlateDecode", ImageFormatPNG},
		{"unknown filter", "LZWDecode", ImageFormatUnknown},
		{"empty filter", "", ImageFormatUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectImageFormat(tt.filter, nil)
			if got != tt.expect {
				t.Errorf("detectImageFormat(%q) = %v, want %v", tt.filter, got, tt.expect)
			}
		})
	}
}

// TestToFloat64 はtoFloat64関数をテストする
func TestToFloat64(t *testing.T) {
	tests := []struct {
		name   string
		input  core.Object
		expect float64
	}{
		{"integer", core.Integer(42), 42},
		{"real", core.Real(3.14), 3.14},
		{"string returns 0", core.String("hello"), 0},
		{"nil returns 0", nil, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toFloat64(tt.input)
			if got != tt.expect {
				t.Errorf("toFloat64(%v) = %v, want %v", tt.input, got, tt.expect)
			}
		})
	}
}

// TestIsCJKRune はisCJKRune関数をテストする
func TestIsCJKRune(t *testing.T) {
	tests := []struct {
		name   string
		r      rune
		expect bool
	}{
		{"ASCII letter", 'A', false},
		{"Kanji", '漢', true},
		{"Hiragana", 'あ', true},
		{"Katakana", 'ア', true},
		{"Hangul", '한', true},
		{"CJK punctuation", '〇', true},
		{"Fullwidth letter", 'Ａ', true},
		{"Space", ' ', false},
		{"Digit", '1', false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isCJKRune(tt.r)
			if got != tt.expect {
				t.Errorf("isCJKRune(%q) = %v, want %v", tt.r, got, tt.expect)
			}
		})
	}
}

// TestTextExtractor_TJ_CJK_NoSpace はTJ配列でCJK文字後のスペース抑制をテストする
func TestTextExtractor_TJ_CJK_NoSpace(t *testing.T) {
	operations := []Operation{
		{Operator: "BT"},
		{Operator: "Tf", Operands: []core.Object{core.Name("F1"), core.Real(12)}},
		{Operator: "Td", Operands: []core.Object{core.Real(100), core.Real(700)}},
		{Operator: "TJ", Operands: []core.Object{
			// After CJK character, large spacing should NOT insert space
			core.Array{core.String("漢字"), core.Integer(-300), core.String("テスト")},
		}},
		{Operator: "ET"},
	}

	extractor := NewTextExtractor(operations, nil, nil)
	elements, err := extractor.Extract()
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(elements) != 1 {
		t.Fatalf("Expected 1 element, got %d", len(elements))
	}

	// CJK characters should NOT get space inserted
	if elements[0].Text != "漢字テスト" {
		t.Errorf("Text = %q, want %q", elements[0].Text, "漢字テスト")
	}
}

// TestTextExtractor_TJ_CJK_Real はTJ配列でReal型の大きなスペーシングとCJKをテストする
func TestTextExtractor_TJ_CJK_Real(t *testing.T) {
	operations := []Operation{
		{Operator: "BT"},
		{Operator: "Tf", Operands: []core.Object{core.Name("F1"), core.Real(12)}},
		{Operator: "Td", Operands: []core.Object{core.Real(100), core.Real(700)}},
		{Operator: "TJ", Operands: []core.Object{
			core.Array{core.String("漢字"), core.Real(-300), core.String("テスト")},
		}},
		{Operator: "ET"},
	}

	extractor := NewTextExtractor(operations, nil, nil)
	elements, err := extractor.Extract()
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(elements) != 1 {
		t.Fatalf("Expected 1 element, got %d", len(elements))
	}

	// CJK after Real spacing should NOT get space
	if elements[0].Text != "漢字テスト" {
		t.Errorf("Text = %q, want %q", elements[0].Text, "漢字テスト")
	}
}

// TestTextExtractor_cm_InsideQ はq内のcmがpageLevelCTMを設定しないことをテストする
func TestTextExtractor_cm_InsideQ(t *testing.T) {
	operations := []Operation{
		{Operator: "q"},
		{Operator: "cm", Operands: []core.Object{
			core.Real(2), core.Real(0), core.Real(0), core.Real(2), core.Real(50), core.Real(100),
		}},
		{Operator: "Q"},
		{Operator: "BT"},
		{Operator: "ET"},
	}

	extractor := NewTextExtractor(operations, nil, nil)
	_, _ = extractor.Extract()

	// cm inside q should NOT set pageLevelCTM (stack is non-empty)
	if extractor.GetPageLevelCTM() != nil {
		t.Error("GetPageLevelCTM() should be nil when cm is inside q/Q")
	}
}

// TestTextExtractor_TJ_EmptyArray はTJ空配列をテストする
func TestTextExtractor_TJ_EmptyArray(t *testing.T) {
	operations := []Operation{
		{Operator: "BT"},
		{Operator: "Tf", Operands: []core.Object{core.Name("F1"), core.Real(12)}},
		{Operator: "Td", Operands: []core.Object{core.Real(100), core.Real(700)}},
		{Operator: "TJ", Operands: []core.Object{core.Array{}}},
		{Operator: "ET"},
	}

	extractor := NewTextExtractor(operations, nil, nil)
	elements, err := extractor.Extract()
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	// Empty TJ array should produce no elements
	if len(elements) != 0 {
		t.Errorf("Expected 0 elements, got %d", len(elements))
	}
}

// TestTextExtractor_Tf_InsufficientOperands はTfオペランド不足をテストする
func TestTextExtractor_Tf_InsufficientOperands(t *testing.T) {
	operations := []Operation{
		{Operator: "BT"},
		{Operator: "Tf", Operands: []core.Object{core.Name("F1")}}, // Only 1 operand
		{Operator: "Td", Operands: []core.Object{core.Real(100), core.Real(700)}},
		{Operator: "Tj", Operands: []core.Object{core.String("Test")}},
		{Operator: "ET"},
	}

	extractor := NewTextExtractor(operations, nil, nil)
	elements, err := extractor.Extract()
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(elements) != 1 {
		t.Fatalf("Expected 1 element, got %d", len(elements))
	}
}

// TestTextExtractor_Tm_InsufficientOperands はTmオペランド不足をテストする
func TestTextExtractor_Tm_InsufficientOperands(t *testing.T) {
	operations := []Operation{
		{Operator: "BT"},
		{Operator: "Tf", Operands: []core.Object{core.Name("F1"), core.Real(12)}},
		{Operator: "Tm", Operands: []core.Object{core.Real(1), core.Real(0)}}, // Only 2 operands
		{Operator: "Td", Operands: []core.Object{core.Real(100), core.Real(700)}},
		{Operator: "Tj", Operands: []core.Object{core.String("Test")}},
		{Operator: "ET"},
	}

	extractor := NewTextExtractor(operations, nil, nil)
	elements, err := extractor.Extract()
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	// Should still work, Tm was skipped due to insufficient operands
	if len(elements) != 1 {
		t.Fatalf("Expected 1 element, got %d", len(elements))
	}
}

// TestTextExtractor_cm_InsufficientOperands はcmオペランド不足をテストする
func TestTextExtractor_cm_InsufficientOperands(t *testing.T) {
	operations := []Operation{
		{Operator: "cm", Operands: []core.Object{core.Real(1)}}, // Insufficient
		{Operator: "BT"},
		{Operator: "ET"},
	}

	extractor := NewTextExtractor(operations, nil, nil)
	_, err := extractor.Extract()
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	// pageLevelCTM should be nil since cm had insufficient operands
	if extractor.GetPageLevelCTM() != nil {
		t.Error("GetPageLevelCTM() should be nil for insufficient cm operands")
	}
}

// TestGetTextString_WithToUnicode はToUnicode CMapがある場合のgetTextStringをテストする
func TestGetTextString_WithToUnicode(t *testing.T) {
	extractor := NewTextExtractor(nil, nil, nil)
	// Manually set fontInfo with ToUnicodeCMap
	extractor.currentFontInfo = &FontInfo{
		Name:     "F1",
		BaseFont: "TestFont",
		ToUnicodeCMap: &ToUnicodeCMap{
			charMap: map[uint16]rune{
				0x0026: 'H',
				0x004f: 'i',
			},
		},
	}

	// Test with data that maps via ToUnicode
	result := extractor.getTextString(core.String(string([]byte{0x00, 0x26, 0x00, 0x4f})))
	if result != "Hi" {
		t.Errorf("getTextString with ToUnicode = %q, want %q", result, "Hi")
	}
}

// TestGetTextString_WithToUnicode_Fallback はToUnicodeが空を返す場合のフォールバックをテストする
func TestGetTextString_WithToUnicode_Fallback(t *testing.T) {
	extractor := NewTextExtractor(nil, nil, nil)
	extractor.currentFontInfo = &FontInfo{
		Name:     "F1",
		BaseFont: "TestFont",
		ToUnicodeCMap: &ToUnicodeCMap{
			charMap: map[uint16]rune{}, // Empty map
		},
	}

	// Empty data triggers LookupString to return "" -> falls back to decodePDFString
	result := extractor.getTextString(core.String(""))
	if result != "" {
		t.Errorf("getTextString fallback for empty = %q, want empty", result)
	}

	// Non-empty data with CMap: LookupString processes the CIDs
	// "AB" = CID 0x4142, no mapping -> returns rune(0x4142)
	result2 := extractor.getTextString(core.String("AB"))
	if result2 == "" {
		t.Error("getTextString should return non-empty for non-empty input with CMap")
	}
}

// TestTextExtractor_WithFontInfo_BaseFont はBaseFontが設定されている場合のテストする
func TestTextExtractor_WithFontInfo_BaseFont(t *testing.T) {
	operations := []Operation{
		{Operator: "BT"},
		{Operator: "Tf", Operands: []core.Object{core.Name("F1"), core.Real(12)}},
		{Operator: "Td", Operands: []core.Object{core.Real(100), core.Real(700)}},
		{Operator: "Tj", Operands: []core.Object{core.String("Test")}},
		{Operator: "ET"},
	}

	extractor := NewTextExtractor(operations, nil, nil)
	// Simulate having font info loaded
	extractor.currentFontInfo = &FontInfo{
		Name:     "F1",
		BaseFont: "Helvetica-Bold",
	}

	elements, err := extractor.Extract()
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(elements) != 1 {
		t.Fatalf("Expected 1 element, got %d", len(elements))
	}

	// Font should use BaseFont name
	if elements[0].Font != "Helvetica-Bold" {
		t.Errorf("Font = %q, want %q", elements[0].Font, "Helvetica-Bold")
	}
}

// TestTextExtractor_TJ_WithFontInfo はTJでBaseFontが使われるケースをテストする
func TestTextExtractor_TJ_WithFontInfo(t *testing.T) {
	operations := []Operation{
		{Operator: "BT"},
		{Operator: "Tf", Operands: []core.Object{core.Name("F1"), core.Real(12)}},
		{Operator: "Td", Operands: []core.Object{core.Real(100), core.Real(700)}},
		{Operator: "TJ", Operands: []core.Object{
			core.Array{core.String("Hello")},
		}},
		{Operator: "ET"},
	}

	extractor := NewTextExtractor(operations, nil, nil)
	extractor.currentFontInfo = &FontInfo{
		Name:     "F1",
		BaseFont: "Times-Roman",
	}

	elements, err := extractor.Extract()
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(elements) != 1 {
		t.Fatalf("Expected 1 element, got %d", len(elements))
	}

	if elements[0].Font != "Times-Roman" {
		t.Errorf("Font = %q, want %q", elements[0].Font, "Times-Roman")
	}
}

// TestDecodePDFString_InvalidUTF8 はUTF-8として無効なバイト列をテストする
func TestDecodePDFString_InvalidUTF8(t *testing.T) {
	// Invalid UTF-8 sequence with Latin-1 high bytes (0xA0-0xFF range, not 0x80-0x9F)
	data := []byte{0xC0, 0xC1} // Invalid UTF-8 lead bytes
	result := decodePDFString(data)
	// Should be handled as PDFDocEncoding since it's not valid UTF-8
	if result == "" {
		t.Error("Expected non-empty result for invalid UTF-8")
	}
}

// TestTextExtractor_ComplexStream は複雑なストリームをテストする
func TestTextExtractor_ComplexStream(t *testing.T) {
	// 実際のPDFに近いストリームをシミュレート
	stream := `BT
/F1 12 Tf
100 750 Td
(Title) Tj
0 -20 Td
(Subtitle) Tj
ET
BT
/F2 10 Tf
100 700 Td
(Body text line 1) Tj
T*
(Body text line 2) Tj
ET`

	parser := NewStreamParser([]byte(stream))
	operations, err := parser.ParseOperations()
	if err != nil {
		t.Fatalf("ParseOperations failed: %v", err)
	}

	extractor := NewTextExtractor(operations, nil, nil)
	elements, err := extractor.Extract()

	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	// 少なくとも4つのテキスト要素があるはず
	if len(elements) < 4 {
		t.Errorf("Expected at least 4 elements, got %d", len(elements))
	}

	// 最初のテキストが"Title"であることを確認
	if elements[0].Text != "Title" {
		t.Errorf("First text = %q, want %q", elements[0].Text, "Title")
	}
}
