package content

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ryomak/gopdf/internal/core"
)

func TestExtractGraphicsOperations(t *testing.T) {
	tests := []struct {
		name       string
		operations []Operation
		wantOps    []string // 出力に含まれるべき操作
		dontWant   []string // 出力に含まれてはいけない操作
	}{
		{
			name: "テキストブロックを除外しグラフィックスを保持",
			operations: []Operation{
				{Operator: "q"},
				{Operator: "w", Operands: []core.Object{core.Real(0.5)}},
				{Operator: "RG", Operands: []core.Object{core.Integer(0), core.Integer(0), core.Integer(0)}},
				{Operator: "re", Operands: []core.Object{core.Real(100), core.Real(100), core.Real(400), core.Real(300)}},
				{Operator: "S"},
				{Operator: "BT"},
				{Operator: "Tf", Operands: []core.Object{core.Name("F1"), core.Integer(12)}},
				{Operator: "Td", Operands: []core.Object{core.Real(110), core.Real(380)}},
				{Operator: "Tj", Operands: []core.Object{core.String("Hello")}},
				{Operator: "ET"},
				{Operator: "Q"},
			},
			wantOps:  []string{"q", "w", "RG", "re", "S", "Q"},
			dontWant: []string{"BT", "ET", "Tf", "Td", "Tj"},
		},
		{
			name: "Do（画像描画）を除外",
			operations: []Operation{
				{Operator: "q"},
				{Operator: "cm", Operands: []core.Object{core.Real(100), core.Integer(0), core.Integer(0), core.Real(100), core.Real(50), core.Real(50)}},
				{Operator: "Do", Operands: []core.Object{core.Name("Im1")}},
				{Operator: "Q"},
			},
			wantOps:  []string{"q", "cm", "Q"},
			dontWant: []string{"Do"},
		},
		{
			name: "空の操作リスト",
			operations: []Operation{},
			wantOps:  []string{},
			dontWant: []string{},
		},
		{
			name: "グラフィックスのみ（テキストなし）",
			operations: []Operation{
				{Operator: "q"},
				{Operator: "re", Operands: []core.Object{core.Real(10), core.Real(10), core.Real(100), core.Real(50)}},
				{Operator: "f"},
				{Operator: "m", Operands: []core.Object{core.Real(0), core.Real(0)}},
				{Operator: "l", Operands: []core.Object{core.Real(100), core.Real(100)}},
				{Operator: "S"},
				{Operator: "Q"},
			},
			wantOps:  []string{"q", "re", "f", "m", "l", "S", "Q"},
			dontWant: []string{},
		},
		{
			name: "テキストのみ（グラフィックスなし）",
			operations: []Operation{
				{Operator: "BT"},
				{Operator: "Tf", Operands: []core.Object{core.Name("F1"), core.Integer(12)}},
				{Operator: "Tj", Operands: []core.Object{core.String("Hello")}},
				{Operator: "ET"},
			},
			wantOps:  []string{},
			dontWant: []string{"BT", "ET", "Tf", "Tj"},
		},
		{
			name: "複数のBT/ETブロックとグラフィックスが混在",
			operations: []Operation{
				// テーブルの罫線
				{Operator: "q"},
				{Operator: "w", Operands: []core.Object{core.Real(1)}},
				{Operator: "re", Operands: []core.Object{core.Real(50), core.Real(700), core.Real(500), core.Real(20)}},
				{Operator: "S"},
				// セル1のテキスト
				{Operator: "BT"},
				{Operator: "Tj", Operands: []core.Object{core.String("Cell1")}},
				{Operator: "ET"},
				// 横線
				{Operator: "m", Operands: []core.Object{core.Real(50), core.Real(700)}},
				{Operator: "l", Operands: []core.Object{core.Real(550), core.Real(700)}},
				{Operator: "S"},
				// セル2のテキスト
				{Operator: "BT"},
				{Operator: "Tj", Operands: []core.Object{core.String("Cell2")}},
				{Operator: "ET"},
				{Operator: "Q"},
			},
			wantOps:  []string{"q", "w", "re", "S", "m", "l", "S", "Q"},
			dontWant: []string{"BT", "ET", "Tj"},
		},
		{
			name: "色設定オペレーターの保持",
			operations: []Operation{
				{Operator: "rg", Operands: []core.Object{core.Real(0.9), core.Real(0.9), core.Real(0.9)}},
				{Operator: "re", Operands: []core.Object{core.Real(50), core.Real(50), core.Real(200), core.Real(100)}},
				{Operator: "f"},
			},
			wantOps:  []string{"rg", "re", "f"},
			dontWant: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractGraphicsOperations(tt.operations)
			output := string(result)

			for _, want := range tt.wantOps {
				if !strings.Contains(output, want) {
					t.Errorf("出力に %q が含まれていない: %q", want, output)
				}
			}

			for _, dontWant := range tt.dontWant {
				// オペレーターとして含まれていないことを確認（行末のオペレーターを検出）
				lines := strings.Split(output, "\n")
				for _, line := range lines {
					trimmed := strings.TrimSpace(line)
					if trimmed == "" {
						continue
					}
					// 行の最後のワードがオペレーター
					parts := strings.Fields(trimmed)
					if len(parts) > 0 && parts[len(parts)-1] == dontWant {
						t.Errorf("出力に除外すべき操作 %q が含まれている: %q", dontWant, output)
					}
				}
			}
		})
	}
}

func TestWriteOperand(t *testing.T) {
	tests := []struct {
		name string
		obj  core.Object
		want string
	}{
		{"Integer", core.Integer(42), "42"},
		{"Real", core.Real(3.14), "3.14"},
		{"String", core.String("hello"), "(hello)"},
		{"String with special chars", core.String("hello(world)"), "(hello\\(world\\))"},
		{"Name", core.Name("F1"), "/F1"},
		{"Boolean true", core.Boolean(true), "true"},
		{"Boolean false", core.Boolean(false), "false"},
		{"Array", core.Array{core.Integer(1), core.Integer(2)}, "[1 2]"},
		{"Nil", nil, "null"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			writeOperand(&buf, tt.obj)
			got := buf.String()
			if got != tt.want {
				t.Errorf("writeOperand() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExtractGraphicsOperations_RoundTrip(t *testing.T) {
	// テーブル罫線のような典型的なパターンをテスト
	operations := []Operation{
		{Operator: "q"},
		{Operator: "w", Operands: []core.Object{core.Real(0.5)}},
		{Operator: "RG", Operands: []core.Object{core.Integer(0), core.Integer(0), core.Integer(0)}},
		// 外枠
		{Operator: "re", Operands: []core.Object{core.Real(50), core.Real(500), core.Real(500), core.Real(200)}},
		{Operator: "S"},
		// 横線
		{Operator: "m", Operands: []core.Object{core.Real(50), core.Real(600)}},
		{Operator: "l", Operands: []core.Object{core.Real(550), core.Real(600)}},
		{Operator: "S"},
		// 縦線
		{Operator: "m", Operands: []core.Object{core.Real(300), core.Real(500)}},
		{Operator: "l", Operands: []core.Object{core.Real(300), core.Real(700)}},
		{Operator: "S"},
		{Operator: "Q"},
	}

	result := ExtractGraphicsOperations(operations)
	output := string(result)

	// 全てのグラフィックス操作が含まれることを確認
	expectedOps := []string{"q\n", "0.5 w\n", "0 0 0 RG\n", "re\n", "S\n", "m\n", "l\n", "Q\n"}
	for _, expected := range expectedOps {
		if !strings.Contains(output, expected) {
			t.Errorf("出力に %q が含まれていない。出力:\n%s", expected, output)
		}
	}
}
