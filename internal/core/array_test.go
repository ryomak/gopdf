package core

import (
	"testing"
)

// TestArray はArray型の振る舞いをテストする
func TestArray(t *testing.T) {
	t.Run("empty array", func(t *testing.T) {
		arr := Array{}
		if len(arr) != 0 {
			t.Errorf("Empty array length = %d, want 0", len(arr))
		}
	})

	t.Run("array with integers", func(t *testing.T) {
		arr := Array{
			Integer(1),
			Integer(2),
			Integer(3),
		}
		if len(arr) != 3 {
			t.Errorf("Array length = %d, want 3", len(arr))
		}
		if arr[0] != Integer(1) {
			t.Errorf("Array[0] = %v, want Integer(1)", arr[0])
		}
	})

	t.Run("array with mixed types", func(t *testing.T) {
		arr := Array{
			Integer(42),
			String("hello"),
			Boolean(true),
			Real(3.14),
		}
		if len(arr) != 4 {
			t.Errorf("Array length = %d, want 4", len(arr))
		}
	})

	t.Run("nested array", func(t *testing.T) {
		inner := Array{Integer(1), Integer(2)}
		outer := Array{
			String("outer"),
			inner,
		}
		if len(outer) != 2 {
			t.Errorf("Outer array length = %d, want 2", len(outer))
		}
		if nested, ok := outer[1].(Array); !ok {
			t.Error("Expected nested array")
		} else if len(nested) != 2 {
			t.Errorf("Nested array length = %d, want 2", len(nested))
		}
	})

	t.Run("MediaBox array", func(t *testing.T) {
		// PDFのMediaBox: [0 0 612 792] (Letter size)
		mediaBox := Array{
			Integer(0),
			Integer(0),
			Integer(612),
			Integer(792),
		}
		if len(mediaBox) != 4 {
			t.Errorf("MediaBox length = %d, want 4", len(mediaBox))
		}
		if mediaBox[2] != Integer(612) {
			t.Errorf("MediaBox width = %v, want Integer(612)", mediaBox[2])
		}
		if mediaBox[3] != Integer(792) {
			t.Errorf("MediaBox height = %v, want Integer(792)", mediaBox[3])
		}
	})
}

// TestArrayAppend はArrayへの要素追加をテストする
func TestArrayAppend(t *testing.T) {
	arr := Array{}
	arr = append(arr, Integer(1))
	arr = append(arr, Integer(2))
	arr = append(arr, Integer(3))

	if len(arr) != 3 {
		t.Errorf("Array length after append = %d, want 3", len(arr))
	}
	if arr[2] != Integer(3) {
		t.Errorf("arr[2] = %v, want Integer(3)", arr[2])
	}
}
