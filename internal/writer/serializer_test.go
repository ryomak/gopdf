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
		// 特殊文字のエスケープテスト
		{"with space", core.Name("Name With Space"), "/Name#20With#20Space"},
		{"with hash", core.Name("Name#1"), "/Name#231"},
		{"with parentheses", core.Name("A(B)C"), "/A#28B#29C"},
		{"with slash", core.Name("A/B"), "/A#2FB"},
		{"with percent", core.Name("100%"), "/100#25"},
		{"with brackets", core.Name("[array]"), "/#5Barray#5D"},
		{"with angle brackets", core.Name("<tag>"), "/#3Ctag#3E"},
		{"with braces", core.Name("{dict}"), "/#7Bdict#7D"},
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

// TestEscapeString はescapeStringメソッドの各特殊文字分岐を直接テストする
func TestEscapeString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"plain text", "Hello", "Hello"},
		{"empty", "", ""},
		{"backslash", `a\b`, `a\\b`},
		{"open paren", "a(b", `a\(b`},
		{"close paren", "a)b", `a\)b`},
		{"newline", "a\nb", `a\nb`},
		{"carriage return", "a\rb", `a\rb`},
		{"tab", "a\tb", `a\tb`},
		{"multiple specials", "(\n)", `\(\n\)`},
		{"backslash and paren", `\(`, `\\\(`},
		{"all special chars", "\\\n\r\t()", `\\\n\r\t\(\)`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			s := NewSerializer(&buf)
			got := s.escapeString(tt.input)
			if got != tt.want {
				t.Errorf("escapeString(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestToHexString はtoHexStringメソッドをテストする
func TestToHexString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", "<>"},
		{"simple ASCII", "AB", "<4142>"},
		{"single byte", "A", "<41>"},
		{"null byte", "\x00", "<00>"},
		{"binary data", "\x00\xFF\x0A", "<00FF0A>"},
		{"hello", "Hello", "<48656C6C6F>"},
		{"special chars", "()", "<2829>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			s := NewSerializer(&buf)
			got := s.toHexString(tt.input)
			if got != tt.want {
				t.Errorf("toHexString(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestNeedsHexEncoding はneedsHexEncodingメソッドをテストする
func TestNeedsHexEncoding(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"plain ASCII", "Hello World", false},
		{"empty string", "", false},
		{"digits", "12345", false},
		{"printable symbols", "!@#$^&*-+=", false},
		{"with newline", "a\nb", true},
		{"with carriage return", "a\rb", true},
		{"with tab", "a\tb", true},
		{"with null byte", "a\x00b", true},
		{"with open paren", "a(b", true},
		{"with close paren", "a)b", true},
		{"with backslash", `a\b`, true},
		{"high byte", "a\x80b", true},
		{"DEL character", "a\x7Fb", true},
		{"control char", "\x01", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			s := NewSerializer(&buf)
			got := s.needsHexEncoding(tt.input)
			if got != tt.want {
				t.Errorf("needsHexEncoding(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestSerializeStringWithSpecialChars はString型で特殊文字を含む場合のシリアライズをテストする
func TestSerializeStringWithSpecialChars(t *testing.T) {
	tests := []struct {
		name  string
		value core.String
		want  string
	}{
		{
			"with newline uses hex",
			core.String("Hello\nWorld"),
			"<48656C6C6F0A576F726C64>",
		},
		{
			"with parentheses uses hex",
			core.String("a(b)c"),
			"<6128622963>",
		},
		{
			"with backslash uses hex",
			core.String(`a\b`),
			"<615C62>",
		},
		{
			"with tab uses hex",
			core.String("a\tb"),
			"<610962>",
		},
		{
			"with carriage return uses hex",
			core.String("a\rb"),
			"<610D62>",
		},
		{
			"binary data uses hex",
			core.String("\x00\x01\x02"),
			"<000102>",
		},
		{
			"high bytes use hex",
			core.String("\x80\xFF"),
			"<80FF>",
		},
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
				t.Errorf("Serialize(String) = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestSerializeStreamExactOutput はストリームの正確な出力フォーマットをテストする
func TestSerializeStreamExactOutput(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		dict core.Dictionary
	}{
		{
			"empty data",
			[]byte{},
			core.Dictionary{
				core.Name("Length"): core.Integer(0),
			},
		},
		{
			"binary data",
			[]byte{0x00, 0xFF, 0x0A, 0x0D},
			core.Dictionary{
				core.Name("Length"): core.Integer(4),
			},
		},
		{
			"multiple dict entries",
			[]byte("BT /F1 12 Tf ET"),
			core.Dictionary{
				core.Name("Length"): core.Integer(15),
				core.Name("Filter"): core.Name("FlateDecode"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream := &core.Stream{
				Dict: tt.dict,
				Data: tt.data,
			}

			var buf bytes.Buffer
			s := NewSerializer(&buf)

			err := s.Serialize(stream)
			if err != nil {
				t.Fatalf("Serialize(Stream) failed: %v", err)
			}

			got := buf.String()

			// Must start with dictionary
			if !strings.HasPrefix(got, "<<") {
				t.Error("Stream output must start with <<")
			}

			// Must contain stream/endstream keywords with proper newlines
			if !strings.Contains(got, "\nstream\n") {
				t.Error("Stream must contain newline-delimited 'stream' keyword")
			}
			if !strings.HasSuffix(got, "\nendstream") {
				t.Error("Stream must end with '\\nendstream'")
			}

			// Data must be present between stream and endstream
			streamIdx := strings.Index(got, "stream\n")
			endstreamIdx := strings.Index(got, "\nendstream")
			if streamIdx == -1 || endstreamIdx == -1 {
				t.Fatal("stream/endstream not found")
			}
			dataSection := got[streamIdx+len("stream\n") : endstreamIdx]
			if dataSection != string(tt.data) {
				t.Errorf("stream data = %q, want %q", dataSection, string(tt.data))
			}
		})
	}
}

// TestSerializeIndirectObject はSerializeIndirectObjectをテストする
func TestSerializeIndirectObject(t *testing.T) {
	tests := []struct {
		name    string
		obj     *core.IndirectObject
		wantPre string // prefix "N G obj\n"
		wantSuf string // suffix "\nendobj\n"
		wantCon string // content between prefix and suffix
	}{
		{
			"integer object",
			&core.IndirectObject{
				ObjectNumber:     1,
				GenerationNumber: 0,
				Object:           core.Integer(42),
			},
			"1 0 obj\n",
			"\nendobj\n",
			"42",
		},
		{
			"boolean object",
			&core.IndirectObject{
				ObjectNumber:     5,
				GenerationNumber: 2,
				Object:           core.Boolean(true),
			},
			"5 2 obj\n",
			"\nendobj\n",
			"true",
		},
		{
			"null object",
			&core.IndirectObject{
				ObjectNumber:     3,
				GenerationNumber: 0,
				Object:           core.Null{},
			},
			"3 0 obj\n",
			"\nendobj\n",
			"null",
		},
		{
			"dictionary object",
			&core.IndirectObject{
				ObjectNumber:     1,
				GenerationNumber: 0,
				Object: core.Dictionary{
					core.Name("Type"): core.Name("Catalog"),
				},
			},
			"1 0 obj\n",
			"\nendobj\n",
			"",
		},
		{
			"stream object",
			&core.IndirectObject{
				ObjectNumber:     10,
				GenerationNumber: 0,
				Object: &core.Stream{
					Dict: core.Dictionary{
						core.Name("Length"): core.Integer(5),
					},
					Data: []byte("hello"),
				},
			},
			"10 0 obj\n",
			"\nendobj\n",
			"",
		},
		{
			"array object",
			&core.IndirectObject{
				ObjectNumber:     2,
				GenerationNumber: 0,
				Object:           core.Array{core.Integer(1), core.Integer(2)},
			},
			"2 0 obj\n",
			"\nendobj\n",
			"[1 2]",
		},
		{
			"reference as content",
			&core.IndirectObject{
				ObjectNumber:     4,
				GenerationNumber: 0,
				Object:           &core.Reference{ObjectNumber: 1, GenerationNumber: 0},
			},
			"4 0 obj\n",
			"\nendobj\n",
			"1 0 R",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			s := NewSerializer(&buf)

			err := s.SerializeIndirectObject(tt.obj)
			if err != nil {
				t.Fatalf("SerializeIndirectObject() failed: %v", err)
			}

			got := buf.String()

			if !strings.HasPrefix(got, tt.wantPre) {
				t.Errorf("output prefix = %q, want prefix %q", got, tt.wantPre)
			}
			if !strings.HasSuffix(got, tt.wantSuf) {
				t.Errorf("output suffix = %q, want suffix %q", got, tt.wantSuf)
			}

			// Check inner content for simple types
			if tt.wantCon != "" {
				inner := got[len(tt.wantPre) : len(got)-len(tt.wantSuf)]
				if inner != tt.wantCon {
					t.Errorf("inner content = %q, want %q", inner, tt.wantCon)
				}
			}
		})
	}
}

// TestSerializeUnsupportedType は未サポート型のシリアライズエラーをテストする
func TestSerializeUnsupportedType(t *testing.T) {
	var buf bytes.Buffer
	s := NewSerializer(&buf)

	// IndirectObject is not handled by Serialize (only SerializeIndirectObject)
	err := s.Serialize(&core.IndirectObject{
		ObjectNumber: 1,
		Object:       core.Integer(42),
	})

	if err == nil {
		t.Error("Serialize(IndirectObject) should return error for unsupported type")
	}
	if !strings.Contains(err.Error(), "unsupported object type") {
		t.Errorf("error message should mention 'unsupported object type', got: %v", err)
	}
}

// TestSerializeDictionaryMultipleKeys は複数キーのDictionaryの出力順序をテストする
func TestSerializeDictionaryMultipleKeys(t *testing.T) {
	dict := core.Dictionary{
		core.Name("Type"):     core.Name("Page"),
		core.Name("Parent"):   &core.Reference{ObjectNumber: 2, GenerationNumber: 0},
		core.Name("MediaBox"): core.Array{core.Integer(0), core.Integer(0), core.Integer(612), core.Integer(792)},
	}

	var buf bytes.Buffer
	s := NewSerializer(&buf)

	err := s.Serialize(dict)
	if err != nil {
		t.Fatalf("Serialize(Dictionary) failed: %v", err)
	}

	got := buf.String()

	// Keys should be sorted alphabetically: MediaBox, Parent, Type
	mediaBoxIdx := strings.Index(got, "/MediaBox")
	parentIdx := strings.Index(got, "/Parent")
	typeIdx := strings.Index(got, "/Type")

	if mediaBoxIdx == -1 || parentIdx == -1 || typeIdx == -1 {
		t.Fatalf("missing keys in output: %q", got)
	}

	if !(mediaBoxIdx < parentIdx && parentIdx < typeIdx) {
		t.Errorf("keys not sorted: MediaBox@%d, Parent@%d, Type@%d", mediaBoxIdx, parentIdx, typeIdx)
	}
}

// TestSerializeArrayWithNestedArray はネストされた配列のシリアライズをテストする
func TestSerializeArrayWithNestedArray(t *testing.T) {
	arr := core.Array{
		core.Integer(1),
		core.Array{core.Integer(2), core.Integer(3)},
		core.Integer(4),
	}

	var buf bytes.Buffer
	s := NewSerializer(&buf)

	err := s.Serialize(arr)
	if err != nil {
		t.Fatalf("Serialize(Array) failed: %v", err)
	}

	got := buf.String()
	want := "[1 [2 3] 4]"
	if got != want {
		t.Errorf("Serialize(nested Array) = %q, want %q", got, want)
	}
}

// TestSerializeDictionaryWithNestedDict はネストされた辞書のシリアライズをテストする
func TestSerializeDictionaryWithNestedDict(t *testing.T) {
	dict := core.Dictionary{
		core.Name("Font"): core.Dictionary{
			core.Name("F1"): &core.Reference{ObjectNumber: 3, GenerationNumber: 0},
		},
	}

	var buf bytes.Buffer
	s := NewSerializer(&buf)

	err := s.Serialize(dict)
	if err != nil {
		t.Fatalf("Serialize(Dictionary) failed: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "<</F1 3 0 R>>") {
		t.Errorf("nested dictionary not serialized properly: %q", got)
	}
}
