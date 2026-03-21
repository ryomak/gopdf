package reader

import (
	"fmt"
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

// TestParser_ParseStream はストリームのパースをテストする
func TestParser_ParseStream(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		dict         core.Dictionary
		expectedData string
		wantErr      bool
	}{
		{
			name:  "Simple stream with LF",
			input: "stream\nHello, World!\nendstream",
			dict: core.Dictionary{
				core.Name("Length"): core.Integer(14),
			},
			expectedData: "Hello, World!\n",
		},
		{
			name:  "Stream with CRLF",
			input: "stream\r\nTest Data\nendstream",
			dict: core.Dictionary{
				core.Name("Length"): core.Integer(10),
			},
			expectedData: "Test Data\n",
		},
		{
			name:  "Empty stream",
			input: "stream\n\nendstream",
			dict: core.Dictionary{
				core.Name("Length"): core.Integer(0),
			},
			expectedData: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(strings.NewReader(tt.input))
			stream, err := parser.ParseStream(tt.dict)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if string(stream.Data) != tt.expectedData {
				t.Errorf("Data = %q, want %q", string(stream.Data), tt.expectedData)
			}
		})
	}
}

// TestParser_ParseStream_MissingLength はLengthなしストリームのエラーテスト
func TestParser_ParseStream_MissingLength(t *testing.T) {
	input := "stream\ndata\nendstream"
	dict := core.Dictionary{} // no Length
	parser := NewParser(strings.NewReader(input))
	_, err := parser.ParseStream(dict)
	if err == nil {
		t.Error("Expected error for missing Length, got nil")
	}
}

// TestParser_ParseStream_InvalidLength は不正なLength型のエラーテスト
func TestParser_ParseStream_InvalidLength(t *testing.T) {
	input := "stream\ndata\nendstream"
	dict := core.Dictionary{
		core.Name("Length"): core.String("notanumber"),
	}
	parser := NewParser(strings.NewReader(input))
	_, err := parser.ParseStream(dict)
	if err == nil {
		t.Error("Expected error for invalid Length type, got nil")
	}
}

// TestParser_ParseStream_ReferenceLength は参照Lengthのエラーテスト
func TestParser_ParseStream_ReferenceLength(t *testing.T) {
	input := "stream\ndata\nendstream"
	dict := core.Dictionary{
		core.Name("Length"): &core.Reference{ObjectNumber: 10, GenerationNumber: 0},
	}
	parser := NewParser(strings.NewReader(input))
	_, err := parser.ParseStream(dict)
	if err == nil {
		t.Error("Expected error for reference Length, got nil")
	}
}

// TestParser_ParseStream_NotStreamKeyword は不正なキーワードのエラーテスト
func TestParser_ParseStream_NotStreamKeyword(t *testing.T) {
	input := "notstream\ndata\nendstream"
	dict := core.Dictionary{
		core.Name("Length"): core.Integer(4),
	}
	parser := NewParser(strings.NewReader(input))
	_, err := parser.ParseStream(dict)
	if err == nil {
		t.Error("Expected error for wrong keyword, got nil")
	}
}

// TestParser_ParseIndirectObject_Stream は間接オブジェクト内のストリームパースをテストする
func TestParser_ParseIndirectObject_Stream(t *testing.T) {
	streamContent := "BT /F1 12 Tf (Hello) Tj ET"
	input := fmt.Sprintf("4 0 obj\n<< /Length %d >>\nstream\n%s\nendstream\nendobj", len(streamContent)+1, streamContent)

	parser := NewParser(strings.NewReader(input))
	objNum, genNum, obj, err := parser.ParseIndirectObject()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if objNum != 4 {
		t.Errorf("ObjectNumber = %d, want 4", objNum)
	}
	if genNum != 0 {
		t.Errorf("GenerationNumber = %d, want 0", genNum)
	}

	stream, ok := obj.(*core.Stream)
	if !ok {
		t.Fatalf("Expected Stream, got %T", obj)
	}

	// ストリームデータの確認（末尾改行を含む）
	if len(stream.Data) == 0 {
		t.Error("Stream data is empty")
	}
}

// TestParser_ParseIndirectObject_VariousTypes は様々な型の間接オブジェクトテスト
func TestParser_ParseIndirectObject_VariousTypes(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		objNum    int
		genNum    int
		checkType func(t *testing.T, obj core.Object)
	}{
		{
			name:   "Integer object",
			input:  "10 0 obj\n42\nendobj",
			objNum: 10,
			genNum: 0,
			checkType: func(t *testing.T, obj core.Object) {
				t.Helper()
				v, ok := obj.(core.Integer)
				if !ok {
					t.Fatalf("Expected Integer, got %T", obj)
				}
				if int(v) != 42 {
					t.Errorf("Value = %d, want 42", v)
				}
			},
		},
		{
			name:   "String object",
			input:  "5 0 obj\n(Hello PDF)\nendobj",
			objNum: 5,
			genNum: 0,
			checkType: func(t *testing.T, obj core.Object) {
				t.Helper()
				v, ok := obj.(core.String)
				if !ok {
					t.Fatalf("Expected String, got %T", obj)
				}
				if string(v) != "Hello PDF" {
					t.Errorf("Value = %q, want %q", string(v), "Hello PDF")
				}
			},
		},
		{
			name:   "Array object",
			input:  "7 0 obj\n[1 2 3]\nendobj",
			objNum: 7,
			genNum: 0,
			checkType: func(t *testing.T, obj core.Object) {
				t.Helper()
				arr, ok := obj.(core.Array)
				if !ok {
					t.Fatalf("Expected Array, got %T", obj)
				}
				if len(arr) != 3 {
					t.Errorf("Array length = %d, want 3", len(arr))
				}
			},
		},
		{
			name:   "Boolean object",
			input:  "8 0 obj\ntrue\nendobj",
			objNum: 8,
			genNum: 0,
			checkType: func(t *testing.T, obj core.Object) {
				t.Helper()
				v, ok := obj.(core.Boolean)
				if !ok {
					t.Fatalf("Expected Boolean, got %T", obj)
				}
				if bool(v) != true {
					t.Errorf("Value = %v, want true", v)
				}
			},
		},
		{
			name:   "Name object",
			input:  "9 0 obj\n/Helvetica\nendobj",
			objNum: 9,
			genNum: 0,
			checkType: func(t *testing.T, obj core.Object) {
				t.Helper()
				v, ok := obj.(core.Name)
				if !ok {
					t.Fatalf("Expected Name, got %T", obj)
				}
				if string(v) != "Helvetica" {
					t.Errorf("Value = %q, want %q", string(v), "Helvetica")
				}
			},
		},
		{
			name:   "Non-zero generation number",
			input:  "3 2 obj\n(test)\nendobj",
			objNum: 3,
			genNum: 2,
			checkType: func(t *testing.T, obj core.Object) {
				t.Helper()
				_, ok := obj.(core.String)
				if !ok {
					t.Fatalf("Expected String, got %T", obj)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(strings.NewReader(tt.input))
			objNum, genNum, obj, err := parser.ParseIndirectObject()
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if objNum != tt.objNum {
				t.Errorf("ObjectNumber = %d, want %d", objNum, tt.objNum)
			}
			if genNum != tt.genNum {
				t.Errorf("GenerationNumber = %d, want %d", genNum, tt.genNum)
			}
			tt.checkType(t, obj)
		})
	}
}

// TestParser_ParseDictionary_ErrorCases は辞書パースのエラーケーステスト
func TestParser_ParseDictionary_ErrorCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "Non-name key",
			input: "<< 42 /Value >>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(strings.NewReader(tt.input))
			_, err := parser.ParseObject()
			if err == nil {
				t.Error("Expected error, got nil")
			}
		})
	}
}

// TestParser_ParseObject_UnexpectedToken は予期しないトークンのエラーテスト
func TestParser_ParseObject_UnexpectedToken(t *testing.T) {
	// ">>" は辞書の終了で、オブジェクトの開始としては不正
	input := ">>"
	parser := NewParser(strings.NewReader(input))
	_, err := parser.ParseObject()
	if err == nil {
		t.Error("Expected error for unexpected token, got nil")
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
