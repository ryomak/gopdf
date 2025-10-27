package core

import (
	"testing"
)

// TestDictionary はDictionary型の振る舞いをテストする
func TestDictionary(t *testing.T) {
	t.Run("empty dictionary", func(t *testing.T) {
		dict := Dictionary{}
		if len(dict) != 0 {
			t.Errorf("Empty dictionary length = %d, want 0", len(dict))
		}
	})

	t.Run("dictionary with simple values", func(t *testing.T) {
		dict := Dictionary{
			Name("Type"):   Name("Page"),
			Name("MediaBox"): Array{Integer(0), Integer(0), Integer(612), Integer(792)},
		}
		if len(dict) != 2 {
			t.Errorf("Dictionary length = %d, want 2", len(dict))
		}
		if dict[Name("Type")] != Name("Page") {
			t.Errorf("dict[Type] = %v, want Name(Page)", dict[Name("Type")])
		}
	})

	t.Run("nested dictionary", func(t *testing.T) {
		fontDict := Dictionary{
			Name("Type"):    Name("Font"),
			Name("Subtype"): Name("Type1"),
			Name("BaseFont"): Name("Helvetica"),
		}
		resourceDict := Dictionary{
			Name("Font"): Dictionary{
				Name("F1"): fontDict,
			},
		}
		if len(resourceDict) != 1 {
			t.Errorf("Resource dictionary length = %d, want 1", len(resourceDict))
		}
		fonts, ok := resourceDict[Name("Font")].(Dictionary)
		if !ok {
			t.Fatal("Expected Font to be a Dictionary")
		}
		f1, ok := fonts[Name("F1")].(Dictionary)
		if !ok {
			t.Fatal("Expected F1 to be a Dictionary")
		}
		if f1[Name("BaseFont")] != Name("Helvetica") {
			t.Errorf("F1 BaseFont = %v, want Name(Helvetica)", f1[Name("BaseFont")])
		}
	})

	t.Run("catalog dictionary", func(t *testing.T) {
		// PDFのCatalog辞書
		catalog := Dictionary{
			Name("Type"):  Name("Catalog"),
			Name("Pages"): &Reference{ObjectNumber: 2, GenerationNumber: 0},
		}
		if catalog[Name("Type")] != Name("Catalog") {
			t.Errorf("Catalog Type = %v, want Name(Catalog)", catalog[Name("Type")])
		}
		ref, ok := catalog[Name("Pages")].(*Reference)
		if !ok {
			t.Fatal("Expected Pages to be a Reference")
		}
		if ref.ObjectNumber != 2 {
			t.Errorf("Pages reference object number = %d, want 2", ref.ObjectNumber)
		}
	})
}

// TestDictionaryGetters はDictionaryの取得メソッドをテストする
func TestDictionaryGetters(t *testing.T) {
	dict := Dictionary{
		Name("Type"):    Name("Page"),
		Name("Count"):   Integer(3),
		Name("Visible"): Boolean(true),
		Name("Width"):   Real(612.0),
		Name("Title"):   String("Test"),
	}

	t.Run("get existing key", func(t *testing.T) {
		val, exists := dict[Name("Type")]
		if !exists {
			t.Error("Expected Type key to exist")
		}
		if val != Name("Page") {
			t.Errorf("Type value = %v, want Name(Page)", val)
		}
	})

	t.Run("get non-existing key", func(t *testing.T) {
		val, exists := dict[Name("NonExistent")]
		if exists {
			t.Error("Expected NonExistent key to not exist")
		}
		if val != nil {
			t.Errorf("NonExistent value = %v, want nil", val)
		}
	})

	t.Run("iterate over dictionary", func(t *testing.T) {
		count := 0
		for key, val := range dict {
			if key == "" {
				t.Error("Key should not be empty")
			}
			if val == nil {
				t.Error("Value should not be nil")
			}
			count++
		}
		if count != 5 {
			t.Errorf("Iterated %d items, want 5", count)
		}
	})
}

// TestDictionaryModification はDictionaryの変更をテストする
func TestDictionaryModification(t *testing.T) {
	dict := Dictionary{}

	// 追加
	dict[Name("Type")] = Name("Page")
	if len(dict) != 1 {
		t.Errorf("Dictionary length after add = %d, want 1", len(dict))
	}

	// 更新
	dict[Name("Type")] = Name("Catalog")
	if dict[Name("Type")] != Name("Catalog") {
		t.Errorf("Type after update = %v, want Name(Catalog)", dict[Name("Type")])
	}
	if len(dict) != 1 {
		t.Errorf("Dictionary length after update = %d, want 1", len(dict))
	}

	// 削除
	delete(dict, Name("Type"))
	if len(dict) != 0 {
		t.Errorf("Dictionary length after delete = %d, want 0", len(dict))
	}
}
