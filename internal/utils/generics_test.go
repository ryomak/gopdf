package utils

import (
	"reflect"
	"testing"
)

func TestMap(t *testing.T) {
	tests := []struct {
		name     string
		items    []int
		fn       func(int) int
		expected []int
	}{
		{
			name:     "empty slice",
			items:    []int{},
			fn:       func(x int) int { return x * 2 },
			expected: []int{},
		},
		{
			name:     "single element",
			items:    []int{5},
			fn:       func(x int) int { return x * 2 },
			expected: []int{10},
		},
		{
			name:     "multiple elements",
			items:    []int{1, 2, 3, 4, 5},
			fn:       func(x int) int { return x * 2 },
			expected: []int{2, 4, 6, 8, 10},
		},
		{
			name:     "type conversion",
			items:    []int{1, 2, 3},
			fn:       func(x int) int { return x + 10 },
			expected: []int{11, 12, 13},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Map(tt.items, tt.fn)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Map() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMapStringToInt(t *testing.T) {
	tests := []struct {
		name     string
		items    []string
		fn       func(string) int
		expected []int
	}{
		{
			name:     "string length",
			items:    []string{"a", "ab", "abc"},
			fn:       func(s string) int { return len(s) },
			expected: []int{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Map(tt.items, tt.fn)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Map() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestExtractAs(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected any
		ok       bool
	}{
		{
			name:     "int to int",
			value:    42,
			expected: 42,
			ok:       true,
		},
		{
			name:     "string to string",
			value:    "hello",
			expected: "hello",
			ok:       true,
		},
		{
			name:     "int to string fails",
			value:    42,
			expected: "",
			ok:       false,
		},
		{
			name:     "nil value",
			value:    nil,
			expected: 0,
			ok:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.expected.(type) {
			case int:
				result, ok := ExtractAs[int](tt.value)
				if ok != tt.ok {
					t.Errorf("ExtractAs() ok = %v, want %v", ok, tt.ok)
				}
				if tt.ok && result != tt.expected {
					t.Errorf("ExtractAs() result = %v, want %v", result, tt.expected)
				}
			case string:
				result, ok := ExtractAs[string](tt.value)
				if ok != tt.ok {
					t.Errorf("ExtractAs() ok = %v, want %v", ok, tt.ok)
				}
				if tt.ok && result != tt.expected {
					t.Errorf("ExtractAs() result = %v, want %v", result, tt.expected)
				}
			}
		})
	}
}

func TestMustExtractAs(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		context   string
		expectErr bool
	}{
		{
			name:      "int success",
			value:     42,
			context:   "test",
			expectErr: false,
		},
		{
			name:      "string to int fails",
			value:     "hello",
			context:   "test context",
			expectErr: true,
		},
		{
			name:      "nil fails",
			value:     nil,
			context:   "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := MustExtractAs[int](tt.value, tt.context)
			if (err != nil) != tt.expectErr {
				t.Errorf("MustExtractAs() error = %v, expectErr %v", err, tt.expectErr)
			}
			if !tt.expectErr && result != tt.value.(int) {
				t.Errorf("MustExtractAs() result = %v, want %v", result, tt.value)
			}
		})
	}
}

// Benchmark tests
func BenchmarkMap(b *testing.B) {
	items := make([]int, 1000)
	for i := range items {
		items[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Map(items, func(x int) int { return x * 2 })
	}
}
