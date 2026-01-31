package gopdf

import (
	"testing"
)

// TestSortTextElements_SingleLine は同じ行のテキストのソートをテストする
func TestSortTextElements_SingleLine(t *testing.T) {
	elements := []TextElement{
		{Text: "World", X: 200, Y: 700, Size: 12},
		{Text: "Hello", X: 100, Y: 700, Size: 12},
		{Text: "!", X: 250, Y: 700, Size: 12},
	}

	sorted := SortTextElements(elements)

	// 期待される順序: Hello, World, !
	if len(sorted) != 3 {
		t.Fatalf("Expected 3 elements, got %d", len(sorted))
	}

	if sorted[0].Text != "Hello" {
		t.Errorf("First element = %q, want %q", sorted[0].Text, "Hello")
	}
	if sorted[1].Text != "World" {
		t.Errorf("Second element = %q, want %q", sorted[1].Text, "World")
	}
	if sorted[2].Text != "!" {
		t.Errorf("Third element = %q, want %q", sorted[2].Text, "!")
	}
}

// TestSortTextElements_MultipleLines は複数行のソートをテストする
func TestSortTextElements_MultipleLines(t *testing.T) {
	elements := []TextElement{
		{Text: "Bottom", X: 100, Y: 600, Size: 12},
		{Text: "World", X: 200, Y: 700, Size: 12},
		{Text: "Hello", X: 100, Y: 700, Size: 12},
	}

	sorted := SortTextElements(elements)

	// 期待される順序: Hello, World (Y=700の行), Bottom (Y=600の行)
	if len(sorted) != 3 {
		t.Fatalf("Expected 3 elements, got %d", len(sorted))
	}

	if sorted[0].Text != "Hello" {
		t.Errorf("First element = %q, want %q", sorted[0].Text, "Hello")
	}
	if sorted[1].Text != "World" {
		t.Errorf("Second element = %q, want %q", sorted[1].Text, "World")
	}
	if sorted[2].Text != "Bottom" {
		t.Errorf("Third element = %q, want %q", sorted[2].Text, "Bottom")
	}
}

// TestSortTextElements_Complex は複雑なレイアウトのソートをテストする
func TestSortTextElements_Complex(t *testing.T) {
	elements := []TextElement{
		// 2行目（Y=690）
		{Text: "line2-b", X: 200, Y: 690, Size: 12},
		{Text: "line2-a", X: 100, Y: 690, Size: 12},

		// 1行目（Y=700）
		{Text: "line1-c", X: 300, Y: 700, Size: 12},
		{Text: "line1-a", X: 100, Y: 700, Size: 12},
		{Text: "line1-b", X: 200, Y: 700, Size: 12},

		// 3行目（Y=600）
		{Text: "line3-a", X: 100, Y: 600, Size: 12},
	}

	sorted := SortTextElements(elements)

	// 期待される順序: line1-a, line1-b, line1-c, line2-a, line2-b, line3-a
	expected := []string{"line1-a", "line1-b", "line1-c", "line2-a", "line2-b", "line3-a"}

	if len(sorted) != len(expected) {
		t.Fatalf("Expected %d elements, got %d", len(expected), len(sorted))
	}

	for i, exp := range expected {
		if sorted[i].Text != exp {
			t.Errorf("Element[%d] = %q, want %q", i, sorted[i].Text, exp)
		}
	}
}

// TestSortTextElements_Empty は空の要素のソートをテストする
func TestSortTextElements_Empty(t *testing.T) {
	elements := []TextElement{}
	sorted := SortTextElements(elements)

	if len(sorted) != 0 {
		t.Errorf("Expected 0 elements, got %d", len(sorted))
	}
}

// TestGroupByLine は行のグループ化をテストする
func TestGroupByLine(t *testing.T) {
	elements := []TextElement{
		{Text: "a", X: 100, Y: 700, Size: 12},
		{Text: "b", X: 200, Y: 700, Size: 12},
		{Text: "c", X: 100, Y: 696, Size: 12}, // 同じ行（閾値内: 差=4 < 6）
		{Text: "d", X: 100, Y: 650, Size: 12}, // 別の行
	}

	lines := groupByLine(elements)

	// 2行に分かれるはず
	if len(lines) != 2 {
		t.Fatalf("Expected 2 lines, got %d", len(lines))
	}

	// 1行目: a, b, c
	if len(lines[0]) != 3 {
		t.Errorf("First line: expected 3 elements, got %d", len(lines[0]))
	}

	// 2行目: d
	if len(lines[1]) != 1 {
		t.Errorf("Second line: expected 1 element, got %d", len(lines[1]))
	}
}

// TestTextElementsToString はテキスト要素の文字列変換をテストする
func TestTextElementsToString(t *testing.T) {
	elements := []TextElement{
		{Text: "Hello", X: 100, Y: 700, Size: 12},
		{Text: "World", X: 200, Y: 700, Size: 12},
		{Text: "Bottom", X: 100, Y: 650, Size: 12},
	}

	// ソート済みと仮定
	sorted := SortTextElements(elements)
	text := TextElementsToString(sorted)

	// "Hello World\nBottom" のような形式
	if text != "Hello World\nBottom" {
		t.Errorf("Text = %q, want %q", text, "Hello World\nBottom")
	}
}

// TestTextElementsToString_Empty は空の要素の文字列変換をテストする
func TestTextElementsToString_Empty(t *testing.T) {
	elements := []TextElement{}
	text := TextElementsToString(elements)

	if text != "" {
		t.Errorf("Expected empty string, got %q", text)
	}
}

// TestEstimateTextWidth は幅の概算をテストする
func TestEstimateTextWidth(t *testing.T) {
	width := estimateTextWidth("Hello", 12, "Helvetica")

	// 5文字 * 12 * 0.6 = 36
	expected := 36.0
	if width != expected {
		t.Errorf("Width = %f, want %f", width, expected)
	}
}
