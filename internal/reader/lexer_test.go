package reader

import (
	"strings"
	"testing"
)

// TestLexer_NextToken はLexerの基本的なトークン化をテストする
func TestLexer_NextToken(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Token
	}{
		{
			name:  "Integer",
			input: "42",
			expected: []Token{
				{Type: TokenInteger, Value: 42},
			},
		},
		{
			name:  "Negative integer",
			input: "-17",
			expected: []Token{
				{Type: TokenInteger, Value: -17},
			},
		},
		{
			name:  "Real number",
			input: "3.14",
			expected: []Token{
				{Type: TokenReal, Value: 3.14},
			},
		},
		{
			name:  "Negative real",
			input: "-0.001",
			expected: []Token{
				{Type: TokenReal, Value: -0.001},
			},
		},
		{
			name:  "Literal string",
			input: "(Hello, World!)",
			expected: []Token{
				{Type: TokenString, Value: "Hello, World!"},
			},
		},
		{
			name:  "String with escape",
			input: `(Test\nNew Line)`,
			expected: []Token{
				{Type: TokenString, Value: "Test\nNew Line"},
			},
		},
		{
			name:  "Hex string",
			input: "<48656C6C6F>",
			expected: []Token{
				{Type: TokenString, Value: "Hello"},
			},
		},
		{
			name:  "Name",
			input: "/Type",
			expected: []Token{
				{Type: TokenName, Value: "Type"},
			},
		},
		{
			name:  "Name with special chars",
			input: "/Helvetica-Bold",
			expected: []Token{
				{Type: TokenName, Value: "Helvetica-Bold"},
			},
		},
		{
			name:  "Boolean true",
			input: "true",
			expected: []Token{
				{Type: TokenBoolean, Value: true},
			},
		},
		{
			name:  "Boolean false",
			input: "false",
			expected: []Token{
				{Type: TokenBoolean, Value: false},
			},
		},
		{
			name:  "Null",
			input: "null",
			expected: []Token{
				{Type: TokenNull, Value: nil},
			},
		},
		{
			name:  "Reference R",
			input: "R",
			expected: []Token{
				{Type: TokenRef},
			},
		},
		{
			name:  "Keyword obj",
			input: "obj",
			expected: []Token{
				{Type: TokenKeyword, Value: "obj"},
			},
		},
		{
			name:  "Keyword endobj",
			input: "endobj",
			expected: []Token{
				{Type: TokenKeyword, Value: "endobj"},
			},
		},
		{
			name:  "Dictionary start",
			input: "<<",
			expected: []Token{
				{Type: TokenDictStart},
			},
		},
		{
			name:  "Dictionary end",
			input: ">>",
			expected: []Token{
				{Type: TokenDictEnd},
			},
		},
		{
			name:  "Array start",
			input: "[",
			expected: []Token{
				{Type: TokenArrayStart},
			},
		},
		{
			name:  "Array end",
			input: "]",
			expected: []Token{
				{Type: TokenArrayEnd},
			},
		},
		{
			name:  "Simple dictionary",
			input: "<< /Type /Page >>",
			expected: []Token{
				{Type: TokenDictStart},
				{Type: TokenName, Value: "Type"},
				{Type: TokenName, Value: "Page"},
				{Type: TokenDictEnd},
			},
		},
		{
			name:  "Array with numbers",
			input: "[0 0 612 792]",
			expected: []Token{
				{Type: TokenArrayStart},
				{Type: TokenInteger, Value: 0},
				{Type: TokenInteger, Value: 0},
				{Type: TokenInteger, Value: 612},
				{Type: TokenInteger, Value: 792},
				{Type: TokenArrayEnd},
			},
		},
		{
			name:  "Reference",
			input: "2 0 R",
			expected: []Token{
				{Type: TokenInteger, Value: 2},
				{Type: TokenInteger, Value: 0},
				{Type: TokenRef},
			},
		},
		{
			name:  "Comment",
			input: "% This is a comment\n/Type",
			expected: []Token{
				{Type: TokenName, Value: "Type"},
			},
		},
		{
			name:  "Multiple tokens with whitespace",
			input: "1 2   3\n4\t5",
			expected: []Token{
				{Type: TokenInteger, Value: 1},
				{Type: TokenInteger, Value: 2},
				{Type: TokenInteger, Value: 3},
				{Type: TokenInteger, Value: 4},
				{Type: TokenInteger, Value: 5},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(strings.NewReader(tt.input))

			for i, expected := range tt.expected {
				token, err := lexer.NextToken()
				if err != nil {
					t.Fatalf("Token %d: unexpected error: %v", i, err)
				}

				if token.Type != expected.Type {
					t.Errorf("Token %d: Type = %v, want %v", i, token.Type, expected.Type)
				}

				// Value検証（型によって異なる）
				switch expected.Type {
				case TokenInteger:
					if token.Value != expected.Value {
						t.Errorf("Token %d: Value = %v, want %v", i, token.Value, expected.Value)
					}
				case TokenReal:
					if token.Value != expected.Value {
						t.Errorf("Token %d: Value = %v, want %v", i, token.Value, expected.Value)
					}
				case TokenString, TokenName, TokenKeyword:
					if token.Value != expected.Value {
						t.Errorf("Token %d: Value = %v, want %v", i, token.Value, expected.Value)
					}
				case TokenBoolean:
					if token.Value != expected.Value {
						t.Errorf("Token %d: Value = %v, want %v", i, token.Value, expected.Value)
					}
				case TokenNull:
					if token.Value != nil {
						t.Errorf("Token %d: Value = %v, want nil", i, token.Value)
					}
				}
			}

			// 最後のトークンの後はEOF
			token, err := lexer.NextToken()
			if err != nil {
				t.Fatalf("Expected EOF, got error: %v", err)
			}
			if token.Type != TokenEOF {
				t.Errorf("Expected EOF, got %v", token.Type)
			}
		})
	}
}

// TestLexer_IndirectObject は間接オブジェクトのトークン化をテストする
func TestLexer_IndirectObject(t *testing.T) {
	input := `1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj`

	expected := []Token{
		{Type: TokenInteger, Value: 1},
		{Type: TokenInteger, Value: 0},
		{Type: TokenKeyword, Value: "obj"},
		{Type: TokenDictStart},
		{Type: TokenName, Value: "Type"},
		{Type: TokenName, Value: "Catalog"},
		{Type: TokenName, Value: "Pages"},
		{Type: TokenInteger, Value: 2},
		{Type: TokenInteger, Value: 0},
		{Type: TokenRef},
		{Type: TokenDictEnd},
		{Type: TokenKeyword, Value: "endobj"},
	}

	lexer := NewLexer(strings.NewReader(input))

	for i, exp := range expected {
		token, err := lexer.NextToken()
		if err != nil {
			t.Fatalf("Token %d: unexpected error: %v", i, err)
		}

		if token.Type != exp.Type {
			t.Errorf("Token %d: Type = %v, want %v", i, token.Type, exp.Type)
		}

		// 値の検証
		if exp.Value != nil && token.Value != exp.Value {
			t.Errorf("Token %d: Value = %v, want %v", i, token.Value, exp.Value)
		}
	}
}

// TestLexer_ComplexDictionary は複雑な辞書のトークン化をテストする
func TestLexer_ComplexDictionary(t *testing.T) {
	input := `<< /Type /Page
   /Parent 2 0 R
   /MediaBox [0 0 612 792]
   /Contents 4 0 R
   /Resources << /Font << /F1 5 0 R >> >>
>>`

	lexer := NewLexer(strings.NewReader(input))

	// 最初のトークンがDictStartであることを確認
	token, err := lexer.NextToken()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if token.Type != TokenDictStart {
		t.Errorf("First token should be DictStart, got %v", token.Type)
	}

	// すべてのトークンを読む
	count := 1
	for {
		token, err := lexer.NextToken()
		if err != nil {
			t.Fatalf("Unexpected error at token %d: %v", count, err)
		}
		if token.Type == TokenEOF {
			break
		}
		count++
	}

	// トークン数が適切であることを確認
	if count < 20 {
		t.Errorf("Expected more than 20 tokens, got %d", count)
	}
}

// TestLexer_ReadBytes はReadBytesをテストする
func TestLexer_ReadBytes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		n        int
		expected string
		wantErr  bool
	}{
		{
			name:     "Read all bytes",
			input:    "Hello",
			n:        5,
			expected: "Hello",
		},
		{
			name:     "Read partial bytes",
			input:    "Hello, World!",
			n:        5,
			expected: "Hello",
		},
		{
			name:     "Read zero bytes",
			input:    "Hello",
			n:        0,
			expected: "",
		},
		{
			name:    "Read more than available",
			input:   "Hi",
			n:       10,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(strings.NewReader(tt.input))
			result, err := lexer.ReadBytes(tt.n)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if string(result) != tt.expected {
				t.Errorf("Result = %q, want %q", string(result), tt.expected)
			}
		})
	}
}

// TestLexer_ReadBytes_WithPeekedBuffer はpeekedバッファありのReadBytesをテストする
func TestLexer_ReadBytes_WithPeekedBuffer(t *testing.T) {
	lexer := NewLexer(strings.NewReader("ABCDEF"))

	// peekByteで先読みバッファにデータを入れる
	b, err := lexer.peekByte()
	if err != nil {
		t.Fatalf("peekByte error: %v", err)
	}
	if b != 'A' {
		t.Fatalf("peeked = %c, want A", b)
	}

	// ReadBytesで先読みバッファ含めて読む
	result, err := lexer.ReadBytes(4)
	if err != nil {
		t.Fatalf("ReadBytes error: %v", err)
	}
	if string(result) != "ABCD" {
		t.Errorf("Result = %q, want %q", string(result), "ABCD")
	}
}

// TestLexer_ReadBytes_PeekedExact はpeekedバッファだけで足りる場合のテスト
func TestLexer_ReadBytes_PeekedExact(t *testing.T) {
	lexer := NewLexer(strings.NewReader("XY"))

	// 2バイト先読み
	_, _ = lexer.peekByte()
	_, _ = lexer.readByte()
	_, _ = lexer.peekByte()

	// peekedから1バイト読む
	result, err := lexer.ReadBytes(1)
	if err != nil {
		t.Fatalf("ReadBytes error: %v", err)
	}
	if string(result) != "Y" {
		t.Errorf("Result = %q, want %q", string(result), "Y")
	}
}

// TestLexer_LiteralString_NestedParens はネストした括弧のテスト
func TestLexer_LiteralString_NestedParens(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple nested parens",
			input:    "(Hello (World))",
			expected: "Hello (World)",
		},
		{
			name:     "Double nested parens",
			input:    "(a (b (c)))",
			expected: "a (b (c))",
		},
		{
			name:     "Escaped parens",
			input:    `(Hello \(World\))`,
			expected: "Hello (World)",
		},
		{
			name:     "Escaped backslash",
			input:    `(Path\\Name)`,
			expected: "Path\\Name",
		},
		{
			name:     "Tab escape",
			input:    `(Col1\tCol2)`,
			expected: "Col1\tCol2",
		},
		{
			name:     "Carriage return escape",
			input:    `(Line\r)`,
			expected: "Line\r",
		},
		{
			name:     "Unknown escape passes through",
			input:    `(Test\x)`,
			expected: "Testx",
		},
		{
			name:     "Empty string",
			input:    "()",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(strings.NewReader(tt.input))
			token, err := lexer.NextToken()
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if token.Type != TokenString {
				t.Fatalf("Type = %v, want TokenString", token.Type)
			}
			if token.Value != tt.expected {
				t.Errorf("Value = %q, want %q", token.Value, tt.expected)
			}
		})
	}
}

// TestLexer_Name_SpecialChars は名前の特殊文字テスト
func TestLexer_Name_SpecialChars(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple name",
			input:    "/Type",
			expected: "Type",
		},
		{
			name:     "Name with hyphen",
			input:    "/Helvetica-Bold",
			expected: "Helvetica-Bold",
		},
		{
			name:     "Name with hex escape",
			input:    "/A#20B",
			expected: "A B",
		},
		{
			name:     "Name at EOF",
			input:    "/LastName",
			expected: "LastName",
		},
		{
			name:     "Name before delimiter",
			input:    "/Key>>",
			expected: "Key",
		},
		{
			name:     "Name with dots",
			input:    "/Adobe.PPKLite",
			expected: "Adobe.PPKLite",
		},
		{
			name:     "Name with underscore",
			input:    "/My_Name",
			expected: "My_Name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(strings.NewReader(tt.input))
			token, err := lexer.NextToken()
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if token.Type != TokenName {
				t.Fatalf("Type = %v, want TokenName", token.Type)
			}
			if token.Value != tt.expected {
				t.Errorf("Value = %q, want %q", token.Value, tt.expected)
			}
		})
	}
}

// TestLexer_StreamKeyword はstreamキーワードのトークン化テスト
func TestLexer_StreamKeyword(t *testing.T) {
	input := "stream"
	lexer := NewLexer(strings.NewReader(input))
	token, err := lexer.NextToken()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if token.Type != TokenKeyword {
		t.Fatalf("Type = %v, want TokenKeyword", token.Type)
	}
	if token.Value != "stream" {
		t.Errorf("Value = %v, want stream", token.Value)
	}
}

// TestLexer_Number_EdgeCases は数値の境界ケーステスト
func TestLexer_Number_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		tokenType TokenType
		value     interface{}
	}{
		{
			name:      "Plus sign prefix",
			input:     "+5",
			tokenType: TokenInteger,
			value:     5,
		},
		{
			name:      "Leading dot real",
			input:     ".5",
			tokenType: TokenReal,
			value:     0.5,
		},
		{
			name:      "Zero",
			input:     "0",
			tokenType: TokenInteger,
			value:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(strings.NewReader(tt.input))
			token, err := lexer.NextToken()
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if token.Type != tt.tokenType {
				t.Errorf("Type = %v, want %v", token.Type, tt.tokenType)
			}
			if token.Value != tt.value {
				t.Errorf("Value = %v, want %v", token.Value, tt.value)
			}
		})
	}
}

// TestDecodeHexString はHex文字列のデコードをテストする
func TestDecodeHexString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Hello",
			input:    "48656C6C6F",
			expected: "Hello",
		},
		{
			name:     "World",
			input:    "576F726C64",
			expected: "World",
		},
		{
			name:     "Empty",
			input:    "",
			expected: "",
		},
		{
			name:     "Odd length",
			input:    "48656C6C6F0",
			expected: "Hello\x00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := decodeHexString(tt.input)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Result = %q, want %q", result, tt.expected)
			}
		})
	}
}
