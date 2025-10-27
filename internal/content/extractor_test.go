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

	extractor := NewTextExtractor(operations)
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

	extractor := NewTextExtractor(operations)
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

	extractor := NewTextExtractor(operations)
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

	extractor := NewTextExtractor(operations)
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

	extractor := NewTextExtractor(operations)
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

	extractor := NewTextExtractor(operations)
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

	extractor := NewTextExtractor(operations)
	elements, err := extractor.Extract()

	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(elements) != 0 {
		t.Errorf("Expected 0 elements, got %d", len(elements))
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

	extractor := NewTextExtractor(operations)
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
