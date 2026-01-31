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
