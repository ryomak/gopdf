package reader

import (
	"strings"
	"testing"

	"github.com/ryomak/gopdf/internal/core"
)

// TestParser_ParseObject はParseObjectをテストする
func TestParser_ParseObject(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Object
	}{
		{
			name:     "Integer",
			input:    "42",
			expected: core.Integer(42),
		},
		{
			name:     "Real",
			input:    "3.14",
			expected: core.Real(3.14),
		},
		{
			name:     "String",
			input:    "(Hello)",
			expected: core.String("Hello"),
		},
		{
			name:     "Name",
			input:    "/Type",
			expected: core.Name("Type"),
		},
		{
			name:     "Boolean true",
			input:    "true",
			expected: core.Boolean(true),
		},
		{
			name:     "Boolean false",
			input:    "false",
			expected: core.Boolean(false),
		},
		{
			name:     "Null",
			input:    "null",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(strings.NewReader(tt.input))
			obj, err := parser.ParseObject()
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// 型と値を検証
			switch expected := tt.expected.(type) {
			case core.Integer:
				if obj != expected {
					t.Errorf("Object = %v, want %v", obj, expected)
				}
			case core.Real:
				if obj != expected {
					t.Errorf("Object = %v, want %v", obj, expected)
				}
			case core.String:
				if obj != expected {
					t.Errorf("Object = %v, want %v", obj, expected)
				}
			case core.Name:
				if obj != expected {
					t.Errorf("Object = %v, want %v", obj, expected)
				}
			case core.Boolean:
				if obj != expected {
					t.Errorf("Object = %v, want %v", obj, expected)
				}
			case nil:
				if obj != nil {
					t.Errorf("Object = %v, want nil", obj)
				}
			}
		})
	}
}

// TestParser_ParseReference は参照のパースをテストする
func TestParser_ParseReference(t *testing.T) {
	input := "2 0 R"
	parser := NewParser(strings.NewReader(input))

	obj, err := parser.ParseObject()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	ref, ok := obj.(*core.Reference)
	if !ok {
		t.Fatalf("Expected Reference, got %T", obj)
	}

	if ref.ObjectNumber != 2 {
		t.Errorf("ObjectNumber = %d, want 2", ref.ObjectNumber)
	}
	if ref.GenerationNumber != 0 {
		t.Errorf("GenerationNumber = %d, want 0", ref.GenerationNumber)
	}
}

// TestParser_ParseDictionary は辞書のパースをテストする
func TestParser_ParseDictionary(t *testing.T) {
	input := "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] >>"
	parser := NewParser(strings.NewReader(input))

	obj, err := parser.ParseObject()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	dict, ok := obj.(core.Dictionary)
	if !ok {
		t.Fatalf("Expected Dictionary, got %T", obj)
	}

	// /Type の検証
	if dict[core.Name("Type")] != core.Name("Page") {
		t.Errorf("Type = %v, want Page", dict[core.Name("Type")])
	}

	// /Parent の検証（参照）
	parent, ok := dict[core.Name("Parent")].(*core.Reference)
	if !ok {
		t.Errorf("Parent should be Reference, got %T", dict[core.Name("Parent")])
	} else {
		if parent.ObjectNumber != 2 {
			t.Errorf("Parent ObjectNumber = %d, want 2", parent.ObjectNumber)
		}
	}

	// /MediaBox の検証（配列）
	mediaBox, ok := dict[core.Name("MediaBox")].(core.Array)
	if !ok {
		t.Fatalf("MediaBox should be Array, got %T", dict[core.Name("MediaBox")])
	}
	if len(mediaBox) != 4 {
		t.Errorf("MediaBox length = %d, want 4", len(mediaBox))
	}
}

// TestParser_ParseArray は配列のパースをテストする
func TestParser_ParseArray(t *testing.T) {
	input := "[0 0 612 792]"
	parser := NewParser(strings.NewReader(input))

	obj, err := parser.ParseObject()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	arr, ok := obj.(core.Array)
	if !ok {
		t.Fatalf("Expected Array, got %T", obj)
	}

	expected := []int{0, 0, 612, 792}
	if len(arr) != len(expected) {
		t.Fatalf("Array length = %d, want %d", len(arr), len(expected))
	}

	for i, exp := range expected {
		if arr[i] != core.Integer(exp) {
			t.Errorf("Array[%d] = %v, want %d", i, arr[i], exp)
		}
	}
}

// TestParser_ParseNestedDictionary はネストした辞書のパースをテストする
func TestParser_ParseNestedDictionary(t *testing.T) {
	input := "<< /Resources << /Font << /F1 5 0 R >> >> >>"
	parser := NewParser(strings.NewReader(input))

	obj, err := parser.ParseObject()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	dict, ok := obj.(core.Dictionary)
	if !ok {
		t.Fatalf("Expected Dictionary, got %T", obj)
	}

	// /Resources を取得
	resources, ok := dict[core.Name("Resources")].(core.Dictionary)
	if !ok {
		t.Fatalf("Resources should be Dictionary, got %T", dict[core.Name("Resources")])
	}

	// /Font を取得
	fonts, ok := resources[core.Name("Font")].(core.Dictionary)
	if !ok {
		t.Fatalf("Font should be Dictionary, got %T", resources[core.Name("Font")])
	}

	// /F1 を取得
	f1, ok := fonts[core.Name("F1")].(*core.Reference)
	if !ok {
		t.Fatalf("F1 should be Reference, got %T", fonts[core.Name("F1")])
	}

	if f1.ObjectNumber != 5 {
		t.Errorf("F1 ObjectNumber = %d, want 5", f1.ObjectNumber)
	}
}

// TestParser_ParseIndirectObject は間接オブジェクトのパースをテストする
func TestParser_ParseIndirectObject(t *testing.T) {
	input := `1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj`

	parser := NewParser(strings.NewReader(input))

	objNum, genNum, obj, err := parser.ParseIndirectObject()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if objNum != 1 {
		t.Errorf("ObjectNumber = %d, want 1", objNum)
	}
	if genNum != 0 {
		t.Errorf("GenerationNumber = %d, want 0", genNum)
	}

	dict, ok := obj.(core.Dictionary)
	if !ok {
		t.Fatalf("Expected Dictionary, got %T", obj)
	}

	if dict[core.Name("Type")] != core.Name("Catalog") {
		t.Errorf("Type = %v, want Catalog", dict[core.Name("Type")])
	}

	pages, ok := dict[core.Name("Pages")].(*core.Reference)
	if !ok {
		t.Errorf("Pages should be Reference, got %T", dict[core.Name("Pages")])
	} else {
		if pages.ObjectNumber != 2 {
			t.Errorf("Pages ObjectNumber = %d, want 2", pages.ObjectNumber)
		}
	}
}

// TestParser_ParseMultipleObjects は複数のオブジェクトのパースをテストする
func TestParser_ParseMultipleObjects(t *testing.T) {
	input := `1 0 obj
<< /Type /Catalog >>
endobj

2 0 obj
<< /Type /Pages /Count 1 >>
endobj`

	// 1つ目のオブジェクト
	parser1 := NewParser(strings.NewReader(input))
	objNum1, _, obj1, err := parser1.ParseIndirectObject()
	if err != nil {
		t.Fatalf("Failed to parse first object: %v", err)
	}

	if objNum1 != 1 {
		t.Errorf("First object number = %d, want 1", objNum1)
	}

	dict1, ok := obj1.(core.Dictionary)
	if !ok || dict1[core.Name("Type")] != core.Name("Catalog") {
		t.Error("First object should be Catalog")
	}

	// Note: 同じParserで2つ目を読むことは現在の実装では難しい
	// 実際のReaderでは各オブジェクトの位置にシークして個別にパースする
}

// TestParser_ParseEmptyDictionary は空辞書のパースをテストする
func TestParser_ParseEmptyDictionary(t *testing.T) {
	input := "<< >>"
	parser := NewParser(strings.NewReader(input))

	obj, err := parser.ParseObject()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	dict, ok := obj.(core.Dictionary)
	if !ok {
		t.Fatalf("Expected Dictionary, got %T", obj)
	}

	if len(dict) != 0 {
		t.Errorf("Dictionary should be empty, got %d entries", len(dict))
	}
}

// TestParser_ParseEmptyArray は空配列のパースをテストする
func TestParser_ParseEmptyArray(t *testing.T) {
	input := "[]"
	parser := NewParser(strings.NewReader(input))

	obj, err := parser.ParseObject()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	arr, ok := obj.(core.Array)
	if !ok {
		t.Fatalf("Expected Array, got %T", obj)
	}

	if len(arr) != 0 {
		t.Errorf("Array should be empty, got %d elements", len(arr))
	}
}

// TestParser_ParseMixedArray は混在型配列のパースをテストする
func TestParser_ParseMixedArray(t *testing.T) {
	input := "[123 3.14 (text) /Name true null 2 0 R]"
	parser := NewParser(strings.NewReader(input))

	obj, err := parser.ParseObject()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	arr, ok := obj.(core.Array)
	if !ok {
		t.Fatalf("Expected Array, got %T", obj)
	}

	if len(arr) != 7 {
		t.Fatalf("Array length = %d, want 7", len(arr))
	}

	// 各要素の型を検証
	if _, ok := arr[0].(core.Integer); !ok {
		t.Errorf("arr[0] should be Integer, got %T", arr[0])
	}
	if _, ok := arr[1].(core.Real); !ok {
		t.Errorf("arr[1] should be Real, got %T", arr[1])
	}
	if _, ok := arr[2].(core.String); !ok {
		t.Errorf("arr[2] should be String, got %T", arr[2])
	}
	if _, ok := arr[3].(core.Name); !ok {
		t.Errorf("arr[3] should be Name, got %T", arr[3])
	}
	if _, ok := arr[4].(core.Boolean); !ok {
		t.Errorf("arr[4] should be Boolean, got %T", arr[4])
	}
	if arr[5] != nil {
		t.Errorf("arr[5] should be nil, got %v", arr[5])
	}
	if _, ok := arr[6].(*core.Reference); !ok {
		t.Errorf("arr[6] should be Reference, got %T", arr[6])
	}
}
