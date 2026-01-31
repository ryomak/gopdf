package content

import (
	"testing"

	"github.com/ryomak/gopdf/internal/core"
)

// TestStreamParser_ParseOperations はStreamParserの基本的なパースをテストする
func TestStreamParser_ParseOperations(t *testing.T) {
	tests := []struct {
		name          string
		stream        string
		expectedCount int
		checkFirst    func(*testing.T, Operation)
	}{
		{
			name:          "Simple text operation",
			stream:        "BT\n/F1 12 Tf\n100 700 Td\n(Hello) Tj\nET",
			expectedCount: 5,
			checkFirst: func(t *testing.T, op Operation) {
				if op.Operator != "BT" {
					t.Errorf("First operator = %s, want BT", op.Operator)
				}
				if len(op.Operands) != 0 {
					t.Errorf("BT should have no operands, got %d", len(op.Operands))
				}
			},
		},
		{
			name:          "Tf operator",
			stream:        "/F1 12 Tf",
			expectedCount: 1,
			checkFirst: func(t *testing.T, op Operation) {
				if op.Operator != "Tf" {
					t.Errorf("Operator = %s, want Tf", op.Operator)
				}
				if len(op.Operands) != 2 {
					t.Fatalf("Tf should have 2 operands, got %d", len(op.Operands))
				}
				if op.Operands[0] != core.Name("F1") {
					t.Errorf("First operand = %v, want F1", op.Operands[0])
				}
			},
		},
		{
			name:          "Td operator",
			stream:        "100 700 Td",
			expectedCount: 1,
			checkFirst: func(t *testing.T, op Operation) {
				if op.Operator != "Td" {
					t.Errorf("Operator = %s, want Td", op.Operator)
				}
				if len(op.Operands) != 2 {
					t.Fatalf("Td should have 2 operands, got %d", len(op.Operands))
				}
			},
		},
		{
			name:          "Tj operator",
			stream:        "(Hello, World!) Tj",
			expectedCount: 1,
			checkFirst: func(t *testing.T, op Operation) {
				if op.Operator != "Tj" {
					t.Errorf("Operator = %s, want Tj", op.Operator)
				}
				if len(op.Operands) != 1 {
					t.Fatalf("Tj should have 1 operand, got %d", len(op.Operands))
				}
				str, ok := op.Operands[0].(core.String)
				if !ok {
					t.Fatalf("Operand should be String, got %T", op.Operands[0])
				}
				if string(str) != "Hello, World!" {
					t.Errorf("String = %q, want %q", string(str), "Hello, World!")
				}
			},
		},
		{
			name:          "TJ operator with array",
			stream:        "[(Hello) -50 (World)] TJ",
			expectedCount: 1,
			checkFirst: func(t *testing.T, op Operation) {
				if op.Operator != "TJ" {
					t.Errorf("Operator = %s, want TJ", op.Operator)
				}
				if len(op.Operands) != 1 {
					t.Fatalf("TJ should have 1 operand, got %d", len(op.Operands))
				}
				arr, ok := op.Operands[0].(core.Array)
				if !ok {
					t.Fatalf("Operand should be Array, got %T", op.Operands[0])
				}
				if len(arr) != 3 {
					t.Errorf("Array length = %d, want 3", len(arr))
				}
			},
		},
		{
			name:          "Multiple operations",
			stream:        "q\n1 0 0 1 50 50 cm\nQ",
			expectedCount: 3,
			checkFirst: func(t *testing.T, op Operation) {
				if op.Operator != "q" {
					t.Errorf("First operator = %s, want q", op.Operator)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewStreamParser([]byte(tt.stream))
			operations, err := parser.ParseOperations()
			if err != nil {
				t.Fatalf("ParseOperations failed: %v", err)
			}

			if len(operations) != tt.expectedCount {
				t.Errorf("Operation count = %d, want %d", len(operations), tt.expectedCount)
			}

			if len(operations) > 0 && tt.checkFirst != nil {
				tt.checkFirst(t, operations[0])
			}
		})
	}
}

// TestStreamParser_ComplexStream は複雑なストリームのパースをテストする
func TestStreamParser_ComplexStream(t *testing.T) {
	stream := `BT
/F1 12 Tf
100 700 Td
(Hello, World!) Tj
0 -14 Td
(Second line) Tj
ET
q
1 0 0 1 50 50 cm
100 100 m
200 200 l
S
Q`

	parser := NewStreamParser([]byte(stream))
	operations, err := parser.ParseOperations()
	if err != nil {
		t.Fatalf("ParseOperations failed: %v", err)
	}

	// 最低でも10個以上のオペレーションがあることを確認
	if len(operations) < 10 {
		t.Errorf("Expected at least 10 operations, got %d", len(operations))
	}

	// BTとETが含まれることを確認
	hasBT := false
	hasET := false
	for _, op := range operations {
		if op.Operator == "BT" {
			hasBT = true
		}
		if op.Operator == "ET" {
			hasET = true
		}
	}

	if !hasBT {
		t.Error("Expected BT operator")
	}
	if !hasET {
		t.Error("Expected ET operator")
	}
}

// TestStreamParser_RealNumbers は実数のパースをテストする
func TestStreamParser_RealNumbers(t *testing.T) {
	stream := "100.5 200.75 Td"

	parser := NewStreamParser([]byte(stream))
	operations, err := parser.ParseOperations()
	if err != nil {
		t.Fatalf("ParseOperations failed: %v", err)
	}

	if len(operations) != 1 {
		t.Fatalf("Expected 1 operation, got %d", len(operations))
	}

	op := operations[0]
	if op.Operator != "Td" {
		t.Errorf("Operator = %s, want Td", op.Operator)
	}

	if len(op.Operands) != 2 {
		t.Fatalf("Expected 2 operands, got %d", len(op.Operands))
	}

	// 実数として取得
	val1, ok := op.Operands[0].(core.Real)
	if !ok {
		t.Errorf("First operand should be Real, got %T", op.Operands[0])
	} else if val1 != 100.5 {
		t.Errorf("First operand = %v, want 100.5", val1)
	}

	val2, ok := op.Operands[1].(core.Real)
	if !ok {
		t.Errorf("Second operand should be Real, got %T", op.Operands[1])
	} else if val2 != 200.75 {
		t.Errorf("Second operand = %v, want 200.75", val2)
	}
}

// TestStreamParser_EmptyStream は空のストリームをテストする
func TestStreamParser_EmptyStream(t *testing.T) {
	stream := ""

	parser := NewStreamParser([]byte(stream))
	operations, err := parser.ParseOperations()
	if err != nil {
		t.Fatalf("ParseOperations failed: %v", err)
	}

	if len(operations) != 0 {
		t.Errorf("Expected 0 operations, got %d", len(operations))
	}
}
