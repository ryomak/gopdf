// Package utils provides generic utility functions for common operations.
package utils

import "fmt"

// Map applies a function to each element of a slice and returns a new slice with the results.
// Time complexity: O(n), Space complexity: O(n).
func Map[T, U any](items []T, fn func(T) U) []U {
	result := make([]U, len(items))
	for i, item := range items {
		result[i] = fn(item)
	}
	return result
}

// ExtractAs attempts to cast a value to a specified type.
// Returns the casted value and a boolean indicating success.
// Time complexity: O(1), Space complexity: O(1).
func ExtractAs[T any](value any) (T, bool) {
	v, ok := value.(T)
	return v, ok
}

// MustExtractAs attempts to cast a value to a specified type.
// Returns an error if the cast fails, including type information for debugging.
// Time complexity: O(1), Space complexity: O(1).
func MustExtractAs[T any](value any, context string) (T, error) {
	v, ok := value.(T)
	if !ok {
		var zero T
		if context != "" {
			return zero, fmt.Errorf("%s: expected type %T, got %T", context, zero, value)
		}
		return zero, fmt.Errorf("expected type %T, got %T", zero, value)
	}
	return v, nil
}
