package translate

import (
	"regexp"
	"strings"
)

// TranslateUnit represents the unit of translation.
type TranslateUnit int

const (
	// TranslateUnitBlock translates the entire block at once (default).
	TranslateUnitBlock TranslateUnit = iota
	// TranslateUnitLine translates line by line.
	TranslateUnitLine
	// TranslateUnitSentence translates sentence by sentence.
	TranslateUnitSentence
)

// TranslateText translates text using the given translator and unit.
func TranslateText(text string, translator Translator, unit TranslateUnit) (string, error) {
	switch unit {
	case TranslateUnitBlock:
		// Translate entire block
		return translator.Translate(text)

	case TranslateUnitLine:
		// Translate line by line
		return TranslateByDelimiter(text, translator, "\n", "\n")

	case TranslateUnitSentence:
		// Translate sentence by sentence
		// Temporarily convert newlines to spaces to join sentences
		normalized := strings.ReplaceAll(text, "\n", " ")
		// Split by sentence-ending punctuation and translate
		translated, err := TranslateBySentence(normalized, translator)
		if err != nil {
			return "", err
		}
		return translated, nil

	default:
		return translator.Translate(text)
	}
}

// TranslateByDelimiter splits text by a delimiter and translates each part.
func TranslateByDelimiter(text string, translator Translator, splitDelim, joinDelim string) (string, error) {
	parts := strings.Split(text, splitDelim)
	translatedParts := make([]string, len(parts))

	for i, part := range parts {
		if part == "" {
			translatedParts[i] = ""
			continue
		}
		translated, err := translator.Translate(part)
		if err != nil {
			return "", err
		}
		translatedParts[i] = translated
	}

	return strings.Join(translatedParts, joinDelim), nil
}

// sentenceEndPattern detects sentence-ending punctuation.
var sentenceEndPattern = regexp.MustCompile(`([.!?。！？]+)\s*`)

// TranslateBySentence translates text sentence by sentence.
func TranslateBySentence(text string, translator Translator) (string, error) {
	// Split by sentence-ending punctuation (preserving delimiters)
	parts := sentenceEndPattern.Split(text, -1)
	delimiters := sentenceEndPattern.FindAllString(text, -1)

	if len(parts) == 0 {
		return translator.Translate(text)
	}

	var result strings.Builder
	for i, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			if i < len(delimiters) {
				result.WriteString(delimiters[i])
			}
			continue
		}

		// Translate the sentence
		translated, err := translator.Translate(part)
		if err != nil {
			return "", err
		}
		result.WriteString(translated)

		// Add delimiter
		if i < len(delimiters) {
			result.WriteString(delimiters[i])
		}
	}

	return result.String(), nil
}

// IsASCIIOnly checks if a string contains only ASCII characters.
func IsASCIIOnly(s string) bool {
	for _, r := range s {
		if r > 127 {
			return false
		}
	}
	return true
}
