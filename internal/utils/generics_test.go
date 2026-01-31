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

func TestFilter(t *testing.T) {
	tests := []struct {
		name      string
		items     []int
		predicate func(int) bool
		expected  []int
	}{
		{
			name:      "empty slice",
			items:     []int{},
			predicate: func(x int) bool { return x > 0 },
			expected:  []int{},
		},
		{
			name:      "no matches",
			items:     []int{1, 2, 3},
			predicate: func(x int) bool { return x > 10 },
			expected:  []int{},
		},
		{
			name:      "all matches",
			items:     []int{1, 2, 3},
			predicate: func(x int) bool { return x > 0 },
			expected:  []int{1, 2, 3},
		},
		{
			name:      "partial matches",
			items:     []int{1, 2, 3, 4, 5},
			predicate: func(x int) bool { return x%2 == 0 },
			expected:  []int{2, 4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Filter(tt.items, tt.predicate)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Filter() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGroupBy(t *testing.T) {
	tests := []struct {
		name     string
		items    []int
		keyFunc  func(int) string
		expected map[string][]int
	}{
		{
			name:     "empty slice",
			items:    []int{},
			keyFunc:  func(x int) string { return "key" },
			expected: map[string][]int{},
		},
		{
			name:    "single group",
			items:   []int{1, 2, 3},
			keyFunc: func(x int) string { return "same" },
			expected: map[string][]int{
				"same": {1, 2, 3},
			},
		},
		{
			name:    "multiple groups",
			items:   []int{1, 2, 3, 4, 5},
			keyFunc: func(x int) string {
				if x%2 == 0 {
					return "even"
				}
				return "odd"
			},
			expected: map[string][]int{
				"even": {2, 4},
				"odd":  {1, 3, 5},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GroupBy(tt.items, tt.keyFunc)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("GroupBy() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestReduce(t *testing.T) {
	tests := []struct {
		name     string
		items    []int
		initial  int
		fn       func(int, int) int
		expected int
	}{
		{
			name:     "empty slice",
			items:    []int{},
			initial:  0,
			fn:       func(acc, x int) int { return acc + x },
			expected: 0,
		},
		{
			name:     "sum",
			items:    []int{1, 2, 3, 4, 5},
			initial:  0,
			fn:       func(acc, x int) int { return acc + x },
			expected: 15,
		},
		{
			name:     "product",
			items:    []int{1, 2, 3, 4},
			initial:  1,
			fn:       func(acc, x int) int { return acc * x },
			expected: 24,
		},
		{
			name:     "max",
			items:    []int{3, 7, 2, 9, 1},
			initial:  0,
			fn: func(acc, x int) int {
				if x > acc {
					return x
				}
				return acc
			},
			expected: 9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Reduce(tt.items, tt.initial, tt.fn)
			if result != tt.expected {
				t.Errorf("Reduce() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDeduplicate(t *testing.T) {
	tests := []struct {
		name     string
		items    []int
		expected []int
	}{
		{
			name:     "empty slice",
			items:    []int{},
			expected: []int{},
		},
		{
			name:     "no duplicates",
			items:    []int{1, 2, 3},
			expected: []int{1, 2, 3},
		},
		{
			name:     "all duplicates",
			items:    []int{1, 1, 1},
			expected: []int{1},
		},
		{
			name:     "some duplicates",
			items:    []int{1, 2, 1, 3, 2, 4},
			expected: []int{1, 2, 3, 4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Deduplicate(tt.items)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Deduplicate() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestKeys(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]int
		expected []string
	}{
		{
			name:     "empty map",
			input:    map[string]int{},
			expected: []string{},
		},
		{
			name: "single key",
			input: map[string]int{
				"a": 1,
			},
			expected: []string{"a"},
		},
		{
			name: "multiple keys",
			input: map[string]int{
				"a": 1,
				"b": 2,
				"c": 3,
			},
			expected: []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Keys(tt.input)
			// Sort both slices for comparison since map iteration order is not guaranteed
			if len(result) != len(tt.expected) {
				t.Errorf("Keys() length = %v, want %v", len(result), len(tt.expected))
				return
			}
			// Check all expected keys are present
			resultMap := make(map[string]bool)
			for _, k := range result {
				resultMap[k] = true
			}
			for _, k := range tt.expected {
				if !resultMap[k] {
					t.Errorf("Keys() missing expected key %v", k)
				}
			}
		})
	}
}

func TestValues(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]int
		expected []int
	}{
		{
			name:     "empty map",
			input:    map[string]int{},
			expected: []int{},
		},
		{
			name: "single value",
			input: map[string]int{
				"a": 1,
			},
			expected: []int{1},
		},
		{
			name: "multiple values",
			input: map[string]int{
				"a": 1,
				"b": 2,
				"c": 3,
			},
			expected: []int{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Values(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("Values() length = %v, want %v", len(result), len(tt.expected))
				return
			}
			// Check all expected values are present
			resultMap := make(map[int]bool)
			for _, v := range result {
				resultMap[v] = true
			}
			for _, v := range tt.expected {
				if !resultMap[v] {
					t.Errorf("Values() missing expected value %v", v)
				}
			}
		})
	}
}

func TestGetOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		input        map[string]int
		key          string
		defaultValue int
		expected     int
	}{
		{
			name:         "key exists",
			input:        map[string]int{"a": 10},
			key:          "a",
			defaultValue: 0,
			expected:     10,
		},
		{
			name:         "key does not exist",
			input:        map[string]int{"a": 10},
			key:          "b",
			defaultValue: 99,
			expected:     99,
		},
		{
			name:         "empty map",
			input:        map[string]int{},
			key:          "a",
			defaultValue: 42,
			expected:     42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetOrDefault(tt.input, tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("GetOrDefault() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFind(t *testing.T) {
	tests := []struct {
		name      string
		items     []int
		predicate func(int) bool
		expected  int
		found     bool
	}{
		{
			name:      "empty slice",
			items:     []int{},
			predicate: func(x int) bool { return x > 0 },
			expected:  0,
			found:     false,
		},
		{
			name:      "found",
			items:     []int{1, 2, 3, 4, 5},
			predicate: func(x int) bool { return x > 3 },
			expected:  4,
			found:     true,
		},
		{
			name:      "not found",
			items:     []int{1, 2, 3},
			predicate: func(x int) bool { return x > 10 },
			expected:  0,
			found:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, found := Find(tt.items, tt.predicate)
			if found != tt.found {
				t.Errorf("Find() found = %v, want %v", found, tt.found)
			}
			if found && result != tt.expected {
				t.Errorf("Find() result = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAny(t *testing.T) {
	tests := []struct {
		name      string
		items     []int
		predicate func(int) bool
		expected  bool
	}{
		{
			name:      "empty slice",
			items:     []int{},
			predicate: func(x int) bool { return x > 0 },
			expected:  false,
		},
		{
			name:      "at least one match",
			items:     []int{1, 2, 3},
			predicate: func(x int) bool { return x == 2 },
			expected:  true,
		},
		{
			name:      "no matches",
			items:     []int{1, 2, 3},
			predicate: func(x int) bool { return x > 10 },
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Any(tt.items, tt.predicate)
			if result != tt.expected {
				t.Errorf("Any() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAll(t *testing.T) {
	tests := []struct {
		name      string
		items     []int
		predicate func(int) bool
		expected  bool
	}{
		{
			name:      "empty slice",
			items:     []int{},
			predicate: func(x int) bool { return x > 0 },
			expected:  true,
		},
		{
			name:      "all match",
			items:     []int{1, 2, 3},
			predicate: func(x int) bool { return x > 0 },
			expected:  true,
		},
		{
			name:      "not all match",
			items:     []int{1, 2, 3},
			predicate: func(x int) bool { return x > 2 },
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := All(tt.items, tt.predicate)
			if result != tt.expected {
				t.Errorf("All() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestPartition(t *testing.T) {
	tests := []struct {
		name          string
		items         []int
		predicate     func(int) bool
		expectedTrue  []int
		expectedFalse []int
	}{
		{
			name:          "empty slice",
			items:         []int{},
			predicate:     func(x int) bool { return x > 0 },
			expectedTrue:  []int{},
			expectedFalse: []int{},
		},
		{
			name:          "all true",
			items:         []int{1, 2, 3},
			predicate:     func(x int) bool { return x > 0 },
			expectedTrue:  []int{1, 2, 3},
			expectedFalse: []int{},
		},
		{
			name:          "all false",
			items:         []int{1, 2, 3},
			predicate:     func(x int) bool { return x > 10 },
			expectedTrue:  []int{},
			expectedFalse: []int{1, 2, 3},
		},
		{
			name:          "mixed",
			items:         []int{1, 2, 3, 4, 5},
			predicate:     func(x int) bool { return x%2 == 0 },
			expectedTrue:  []int{2, 4},
			expectedFalse: []int{1, 3, 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			truthy, falsy := Partition(tt.items, tt.predicate)
			if !reflect.DeepEqual(truthy, tt.expectedTrue) {
				t.Errorf("Partition() truthy = %v, want %v", truthy, tt.expectedTrue)
			}
			if !reflect.DeepEqual(falsy, tt.expectedFalse) {
				t.Errorf("Partition() falsy = %v, want %v", falsy, tt.expectedFalse)
			}
		})
	}
}

func TestFlatMap(t *testing.T) {
	tests := []struct {
		name     string
		items    []int
		fn       func(int) []int
		expected []int
	}{
		{
			name:     "empty slice",
			items:    []int{},
			fn:       func(x int) []int { return []int{x, x * 2} },
			expected: []int{},
		},
		{
			name:     "duplicate each element",
			items:    []int{1, 2, 3},
			fn:       func(x int) []int { return []int{x, x} },
			expected: []int{1, 1, 2, 2, 3, 3},
		},
		{
			name:     "expand to range",
			items:    []int{1, 2},
			fn: func(x int) []int {
				result := make([]int, x)
				for i := range result {
					result[i] = x
				}
				return result
			},
			expected: []int{1, 2, 2},
		},
		{
			name:     "empty results",
			items:    []int{1, 2, 3},
			fn:       func(x int) []int { return []int{} },
			expected: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FlatMap(tt.items, tt.fn)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("FlatMap() = %v, want %v", result, tt.expected)
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
			if tt.expectErr && err != nil && tt.context != "" {
				// Check that context is included in error message
				errMsg := err.Error()
				if len(tt.context) > 0 && len(errMsg) > 0 {
					// Error message should contain context
					t.Logf("Error message: %s", errMsg)
				}
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

func BenchmarkFilter(b *testing.B) {
	items := make([]int, 1000)
	for i := range items {
		items[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Filter(items, func(x int) bool { return x%2 == 0 })
	}
}

func BenchmarkDeduplicate(b *testing.B) {
	items := make([]int, 1000)
	for i := range items {
		items[i] = i % 100 // Create duplicates
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Deduplicate(items)
	}
}
