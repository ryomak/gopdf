package core

import (
	"testing"
)

// TestObjectTypes は各PDFオブジェクト型が正しく表現されることをテストする
func TestObjectTypes(t *testing.T) {
	tests := []struct {
		name     string
		obj      Object
		wantType string
	}{
		{
			name:     "Null object",
			obj:      Null{},
			wantType: "Null",
		},
		{
			name:     "Boolean true",
			obj:      Boolean(true),
			wantType: "Boolean",
		},
		{
			name:     "Boolean false",
			obj:      Boolean(false),
			wantType: "Boolean",
		},
		{
			name:     "Integer",
			obj:      Integer(42),
			wantType: "Integer",
		},
		{
			name:     "Real",
			obj:      Real(3.14),
			wantType: "Real",
		},
		{
			name:     "String literal",
			obj:      String("Hello, World!"),
			wantType: "String",
		},
		{
			name:     "Name",
			obj:      Name("Type"),
			wantType: "Name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.obj == nil {
				t.Errorf("Object is nil")
			}
		})
	}
}

// TestInteger はInteger型の振る舞いをテストする
func TestInteger(t *testing.T) {
	tests := []struct {
		name  string
		value int
		want  int
	}{
		{"positive", 42, 42},
		{"negative", -17, -17},
		{"zero", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := Integer(tt.value)
			if int(obj) != tt.want {
				t.Errorf("Integer value = %d, want %d", obj, tt.want)
			}
		})
	}
}

// TestReal はReal型の振る舞いをテストする
func TestReal(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		want  float64
	}{
		{"positive", 3.14, 3.14},
		{"negative", -0.001, -0.001},
		{"zero", 0.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := Real(tt.value)
			if float64(obj) != tt.want {
				t.Errorf("Real value = %f, want %f", obj, tt.want)
			}
		})
	}
}

// TestString はString型の振る舞いをテストする
func TestString(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{"simple", "Hello", "Hello"},
		{"empty", "", ""},
		{"with spaces", "Hello, World!", "Hello, World!"},
		{"unicode", "こんにちは", "こんにちは"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := String(tt.value)
			if string(obj) != tt.want {
				t.Errorf("String value = %s, want %s", obj, tt.want)
			}
		})
	}
}

// TestName はName型の振る舞いをテストする
func TestName(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{"simple", "Type", "Type"},
		{"with number", "F1", "F1"},
		{"camelCase", "MediaBox", "MediaBox"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := Name(tt.value)
			if string(obj) != tt.want {
				t.Errorf("Name value = %s, want %s", obj, tt.want)
			}
		})
	}
}

// TestBoolean はBoolean型の振る舞いをテストする
func TestBoolean(t *testing.T) {
	tests := []struct {
		name  string
		value bool
		want  bool
	}{
		{"true", true, true},
		{"false", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := Boolean(tt.value)
			if bool(obj) != tt.want {
				t.Errorf("Boolean value = %t, want %t", obj, tt.want)
			}
		})
	}
}
