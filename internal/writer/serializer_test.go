package writer

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ryomak/gopdf/internal/core"
)

// TestSerializeNull はNull型のシリアライズをテストする
func TestSerializeNull(t *testing.T) {
	var buf bytes.Buffer
	s := NewSerializer(&buf)

	err := s.Serialize(core.Null{})
	if err != nil {
		t.Fatalf("Serialize(Null) failed: %v", err)
	}

	got := buf.String()
	want := "null"
	if got != want {
		t.Errorf("Serialize(Null) = %q, want %q", got, want)
	}
}

// TestSerializeBoolean はBoolean型のシリアライズをテストする
func TestSerializeBoolean(t *testing.T) {
	tests := []struct {
		name  string
		value core.Boolean
		want  string
	}{
		{"true", core.Boolean(true), "true"},
		{"false", core.Boolean(false), "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			s := NewSerializer(&buf)

			err := s.Serialize(tt.value)
			if err != nil {
				t.Fatalf("Serialize(Boolean) failed: %v", err)
			}

			got := buf.String()
			if got != tt.want {
				t.Errorf("Serialize(Boolean(%t)) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

// TestSerializeInteger はInteger型のシリアライズをテストする
func TestSerializeInteger(t *testing.T) {
	tests := []struct {
		name  string
		value core.Integer
		want  string
	}{
		{"positive", core.Integer(42), "42"},
		{"negative", core.Integer(-17), "-17"},
		{"zero", core.Integer(0), "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			s := NewSerializer(&buf)

			err := s.Serialize(tt.value)
			if err != nil {
				t.Fatalf("Serialize(Integer) failed: %v", err)
			}

			got := buf.String()
			if got != tt.want {
				t.Errorf("Serialize(Integer(%d)) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

// TestSerializeReal はReal型のシリアライズをテストする
func TestSerializeReal(t *testing.T) {
	tests := []struct {
		name  string
		value core.Real
		want  string
	}{
		{"positive", core.Real(3.14), "3.14"},
		{"negative", core.Real(-0.001), "-0.001"},
		{"zero", core.Real(0.0), "0"},
		{"integer-like", core.Real(42.0), "42"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			s := NewSerializer(&buf)

			err := s.Serialize(tt.value)
			if err != nil {
				t.Fatalf("Serialize(Real) failed: %v", err)
			}

			got := buf.String()
			if got != tt.want {
				t.Errorf("Serialize(Real(%f)) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

// TestSerializeString はString型のシリアライズをテストする
func TestSerializeString(t *testing.T) {
	tests := []struct {
		name  string
		value core.String
		want  string
	}{
		{"simple", core.String("Hello"), "(Hello)"},
		{"empty", core.String(""), "()"},
		{"with spaces", core.String("Hello, World!"), "(Hello, World!)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			s := NewSerializer(&buf)

			err := s.Serialize(tt.value)
			if err != nil {
				t.Fatalf("Serialize(String) failed: %v", err)
			}

			got := buf.String()
			if got != tt.want {
				t.Errorf("Serialize(String(%q)) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

// TestSerializeName はName型のシリアライズをテストする
func TestSerializeName(t *testing.T) {
	tests := []struct {
		name  string
		value core.Name
		want  string
	}{
		{"simple", core.Name("Type"), "/Type"},
		{"with number", core.Name("F1"), "/F1"},
		{"camelCase", core.Name("MediaBox"), "/MediaBox"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			s := NewSerializer(&buf)

			err := s.Serialize(tt.value)
			if err != nil {
				t.Fatalf("Serialize(Name) failed: %v", err)
			}

			got := buf.String()
			if got != tt.want {
				t.Errorf("Serialize(Name(%q)) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

// TestSerializeArray はArray型のシリアライズをテストする
func TestSerializeArray(t *testing.T) {
	tests := []struct {
		name  string
		value core.Array
		want  string
	}{
		{
			"empty",
			core.Array{},
			"[]",
		},
		{
			"integers",
			core.Array{core.Integer(1), core.Integer(2), core.Integer(3)},
			"[1 2 3]",
		},
		{
			"mixed types",
			core.Array{core.Integer(42), core.String("hello"), core.Boolean(true)},
			"[42 (hello) true]",
		},
		{
			"MediaBox",
			core.Array{core.Integer(0), core.Integer(0), core.Integer(612), core.Integer(792)},
			"[0 0 612 792]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			s := NewSerializer(&buf)

			err := s.Serialize(tt.value)
			if err != nil {
				t.Fatalf("Serialize(Array) failed: %v", err)
			}

			got := buf.String()
			if got != tt.want {
				t.Errorf("Serialize(Array) = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestSerializeDictionary はDictionary型のシリアライズをテストする
func TestSerializeDictionary(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		dict := core.Dictionary{}
		var buf bytes.Buffer
		s := NewSerializer(&buf)

		err := s.Serialize(dict)
		if err != nil {
			t.Fatalf("Serialize(Dictionary) failed: %v", err)
		}

		got := buf.String()
		want := "<<>>"
		if got != want {
			t.Errorf("Serialize(empty Dictionary) = %q, want %q", got, want)
		}
	})

	t.Run("simple", func(t *testing.T) {
		dict := core.Dictionary{
			core.Name("Type"): core.Name("Catalog"),
		}
		var buf bytes.Buffer
		s := NewSerializer(&buf)

		err := s.Serialize(dict)
		if err != nil {
			t.Fatalf("Serialize(Dictionary) failed: %v", err)
		}

		got := buf.String()
		// Dictionaryのキーは順不同なので、期待する要素が含まれているか確認
		if !strings.Contains(got, "/Type") || !strings.Contains(got, "/Catalog") {
			t.Errorf("Serialize(Dictionary) = %q, should contain /Type and /Catalog", got)
		}
		if !strings.HasPrefix(got, "<<") || !strings.HasSuffix(got, ">>") {
			t.Errorf("Serialize(Dictionary) = %q, should be wrapped in << >>", got)
		}
	})
}

// TestSerializeReference はReference型のシリアライズをテストする
func TestSerializeReference(t *testing.T) {
	tests := []struct {
		name  string
		value *core.Reference
		want  string
	}{
		{
			"simple",
			&core.Reference{ObjectNumber: 1, GenerationNumber: 0},
			"1 0 R",
		},
		{
			"with generation",
			&core.Reference{ObjectNumber: 5, GenerationNumber: 2},
			"5 2 R",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			s := NewSerializer(&buf)

			err := s.Serialize(tt.value)
			if err != nil {
				t.Fatalf("Serialize(Reference) failed: %v", err)
			}

			got := buf.String()
			if got != tt.want {
				t.Errorf("Serialize(Reference) = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestSerializeStream はStream型のシリアライズをテストする
func TestSerializeStream(t *testing.T) {
	data := []byte("test data")
	stream := &core.Stream{
		Dict: core.Dictionary{
			core.Name("Length"): core.Integer(len(data)),
		},
		Data: data,
	}

	var buf bytes.Buffer
	s := NewSerializer(&buf)

	err := s.Serialize(stream)
	if err != nil {
		t.Fatalf("Serialize(Stream) failed: %v", err)
	}

	got := buf.String()
	// ストリームの形式を確認
	if !strings.Contains(got, "<<") || !strings.Contains(got, ">>") {
		t.Errorf("Stream should contain dictionary")
	}
	if !strings.Contains(got, "stream\n") {
		t.Errorf("Stream should contain 'stream' keyword")
	}
	if !strings.Contains(got, "\nendstream") {
		t.Errorf("Stream should contain 'endstream' keyword")
	}
	if !strings.Contains(got, "test data") {
		t.Errorf("Stream should contain data")
	}
}
