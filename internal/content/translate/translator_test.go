package translate

import (
	"strings"
	"testing"
)

func TestFunc(t *testing.T) {
	// Create a simple translator that converts to uppercase
	upper := Func(func(text string) (string, error) {
		return strings.ToUpper(text), nil
	})

	// Test the translator
	result, err := upper.Translate("hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "HELLO" {
		t.Errorf("Translate() = %q, want %q", result, "HELLO")
	}
}

func TestTranslatorInterface(t *testing.T) {
	// Verify Func implements Translator interface
	var _ Translator = Func(func(text string) (string, error) {
		return text, nil
	})
}
