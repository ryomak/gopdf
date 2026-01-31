package core

import (
	"testing"
)

// TestReference はReference型の振る舞いをテストする
func TestReference(t *testing.T) {
	t.Run("simple reference", func(t *testing.T) {
		ref := &Reference{
			ObjectNumber:     1,
			GenerationNumber: 0,
		}
		if ref.ObjectNumber != 1 {
			t.Errorf("Reference ObjectNumber = %d, want 1", ref.ObjectNumber)
		}
		if ref.GenerationNumber != 0 {
			t.Errorf("Reference GenerationNumber = %d, want 0", ref.GenerationNumber)
		}
	})

	t.Run("reference with generation", func(t *testing.T) {
		ref := &Reference{
			ObjectNumber:     5,
			GenerationNumber: 2,
		}
		if ref.ObjectNumber != 5 {
			t.Errorf("Reference ObjectNumber = %d, want 5", ref.ObjectNumber)
		}
		if ref.GenerationNumber != 2 {
			t.Errorf("Reference GenerationNumber = %d, want 2", ref.GenerationNumber)
		}
	})

	t.Run("reference in dictionary", func(t *testing.T) {
		// Catalogでの使用例
		dict := Dictionary{
			Name("Type"):  Name("Catalog"),
			Name("Pages"): &Reference{ObjectNumber: 2, GenerationNumber: 0},
		}
		ref, ok := dict[Name("Pages")].(*Reference)
		if !ok {
			t.Fatal("Expected Pages to be a Reference")
		}
		if ref.ObjectNumber != 2 {
			t.Errorf("Pages reference ObjectNumber = %d, want 2", ref.ObjectNumber)
		}
	})

	t.Run("reference in array", func(t *testing.T) {
		// Pagesツリーの子ノード参照
		kids := Array{
			&Reference{ObjectNumber: 3, GenerationNumber: 0},
			&Reference{ObjectNumber: 4, GenerationNumber: 0},
			&Reference{ObjectNumber: 5, GenerationNumber: 0},
		}
		if len(kids) != 3 {
			t.Errorf("Kids array length = %d, want 3", len(kids))
		}
		ref, ok := kids[0].(*Reference)
		if !ok {
			t.Fatal("Expected first kid to be a Reference")
		}
		if ref.ObjectNumber != 3 {
			t.Errorf("First kid ObjectNumber = %d, want 3", ref.ObjectNumber)
		}
	})
}

// TestIndirectObject はIndirectObject型の振る舞いをテストする
func TestIndirectObject(t *testing.T) {
	t.Run("indirect object with dictionary", func(t *testing.T) {
		obj := &IndirectObject{
			ObjectNumber:     1,
			GenerationNumber: 0,
			Object: Dictionary{
				Name("Type"):  Name("Catalog"),
				Name("Pages"): &Reference{ObjectNumber: 2, GenerationNumber: 0},
			},
		}
		if obj.ObjectNumber != 1 {
			t.Errorf("IndirectObject ObjectNumber = %d, want 1", obj.ObjectNumber)
		}
		dict, ok := obj.Object.(Dictionary)
		if !ok {
			t.Fatal("Expected Object to be a Dictionary")
		}
		if dict[Name("Type")] != Name("Catalog") {
			t.Errorf("Catalog Type = %v, want Name(Catalog)", dict[Name("Type")])
		}
	})

	t.Run("indirect object with stream", func(t *testing.T) {
		stream := &Stream{
			Dict: Dictionary{
				Name("Length"): Integer(10),
			},
			Data: []byte("test data!"),
		}
		obj := &IndirectObject{
			ObjectNumber:     4,
			GenerationNumber: 0,
			Object:           stream,
		}
		if obj.ObjectNumber != 4 {
			t.Errorf("IndirectObject ObjectNumber = %d, want 4", obj.ObjectNumber)
		}
		s, ok := obj.Object.(*Stream)
		if !ok {
			t.Fatal("Expected Object to be a Stream")
		}
		if len(s.Data) != 10 {
			t.Errorf("Stream data length = %d, want 10", len(s.Data))
		}
	})

	t.Run("indirect object with simple type", func(t *testing.T) {
		obj := &IndirectObject{
			ObjectNumber:     10,
			GenerationNumber: 0,
			Object:           Integer(42),
		}
		val, ok := obj.Object.(Integer)
		if !ok {
			t.Fatal("Expected Object to be an Integer")
		}
		if val != Integer(42) {
			t.Errorf("Integer value = %d, want 42", val)
		}
	})
}

// TestReferenceEquality は参照の比較をテストする
func TestReferenceEquality(t *testing.T) {
	ref1 := &Reference{ObjectNumber: 1, GenerationNumber: 0}
	ref2 := &Reference{ObjectNumber: 1, GenerationNumber: 0}
	ref3 := &Reference{ObjectNumber: 2, GenerationNumber: 0}

	// ポインタが異なるので、同じ内容でも等しくない
	if ref1 == ref2 {
		t.Error("Different Reference pointers should not be equal")
	}

	// 内容が異なる
	if ref1.ObjectNumber == ref3.ObjectNumber {
		t.Error("References with different ObjectNumbers should not be equal")
	}

	// 内容を比較
	if ref1.ObjectNumber != ref2.ObjectNumber || ref1.GenerationNumber != ref2.GenerationNumber {
		t.Error("References with same values should have equal fields")
	}
}
