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

// Filter returns a new slice containing only the elements that satisfy the predicate.
// Time complexity: O(n), Space complexity: O(n) worst case.
func Filter[T any](items []T, predicate func(T) bool) []T {
	result := make([]T, 0, len(items))
	for _, item := range items {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}

// GroupBy groups elements by a key function and returns a map of keys to slices of elements.
// Time complexity: O(n), Space complexity: O(n).
func GroupBy[T any, K comparable](items []T, keyFunc func(T) K) map[K][]T {
	result := make(map[K][]T)
	for _, item := range items {
		key := keyFunc(item)
		result[key] = append(result[key], item)
	}
	return result
}

// Reduce aggregates all elements of a slice into a single value using an accumulator function.
// Time complexity: O(n), Space complexity: O(1).
func Reduce[T, U any](items []T, initial U, fn func(U, T) U) U {
	result := initial
	for _, item := range items {
		result = fn(result, item)
	}
	return result
}

// Deduplicate removes duplicate elements from a slice while preserving order.
// Time complexity: O(n), Space complexity: O(n).
func Deduplicate[T comparable](items []T) []T {
	seen := make(map[T]struct{}, len(items))
	result := make([]T, 0, len(items))
	for _, item := range items {
		if _, exists := seen[item]; !exists {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

// Keys returns a slice containing all keys from a map.
// Order is not guaranteed.
// Time complexity: O(n), Space complexity: O(n).
func Keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Values returns a slice containing all values from a map.
// Order is not guaranteed.
// Time complexity: O(n), Space complexity: O(n).
func Values[K comparable, V any](m map[K]V) []V {
	values := make([]V, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

// GetOrDefault retrieves a value from a map, returning the default value if the key doesn't exist.
// Time complexity: O(1), Space complexity: O(1).
func GetOrDefault[K comparable, V any](m map[K]V, key K, defaultValue V) V {
	if v, exists := m[key]; exists {
		return v
	}
	return defaultValue
}

// Find returns the first element that satisfies the predicate, along with a boolean indicating if found.
// Time complexity: O(n), Space complexity: O(1).
func Find[T any](items []T, predicate func(T) bool) (T, bool) {
	for _, item := range items {
		if predicate(item) {
			return item, true
		}
	}
	var zero T
	return zero, false
}

// Any returns true if at least one element satisfies the predicate.
// Time complexity: O(n), Space complexity: O(1).
func Any[T any](items []T, predicate func(T) bool) bool {
	for _, item := range items {
		if predicate(item) {
			return true
		}
	}
	return false
}

// All returns true if all elements satisfy the predicate.
// Time complexity: O(n), Space complexity: O(1).
func All[T any](items []T, predicate func(T) bool) bool {
	for _, item := range items {
		if !predicate(item) {
			return false
		}
	}
	return true
}

// Partition splits a slice into two slices based on a predicate.
// The first slice contains elements that satisfy the predicate, the second contains the rest.
// Time complexity: O(n), Space complexity: O(n).
func Partition[T any](items []T, predicate func(T) bool) ([]T, []T) {
	truthy := make([]T, 0, len(items))
	falsy := make([]T, 0, len(items))
	for _, item := range items {
		if predicate(item) {
			truthy = append(truthy, item)
		} else {
			falsy = append(falsy, item)
		}
	}
	return truthy, falsy
}

// FlatMap applies a function that returns a slice to each element and flattens the results.
// Time complexity: O(n*m) where m is average result size, Space complexity: O(n*m).
func FlatMap[T, U any](items []T, fn func(T) []U) []U {
	result := make([]U, 0, len(items))
	for _, item := range items {
		result = append(result, fn(item)...)
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
