package gopdf

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

// TestMetadata_SetAndGet tests setting and getting metadata
func TestMetadata_SetAndGet(t *testing.T) {
	doc := New()

	metadata := Metadata{
		Title:   "Test Title",
		Author:  "Test Author",
		Subject: "Test Subject",
	}

	doc.SetMetadata(metadata)

	got := doc.GetMetadata()
	if got == nil {
		t.Fatal("GetMetadata() returned nil")
	}

	if got.Title != metadata.Title {
		t.Errorf("Title = %q, want %q", got.Title, metadata.Title)
	}

	if got.Author != metadata.Author {
		t.Errorf("Author = %q, want %q", got.Author, metadata.Author)
	}

	if got.Subject != metadata.Subject {
		t.Errorf("Subject = %q, want %q", got.Subject, metadata.Subject)
	}
}

// TestMetadata_GetNil tests getting metadata when none is set
func TestMetadata_GetNil(t *testing.T) {
	doc := New()

	got := doc.GetMetadata()
	if got != nil {
		t.Errorf("GetMetadata() = %v, want nil", got)
	}
}

// TestMetadata_StandardFields tests all standard metadata fields
func TestMetadata_StandardFields(t *testing.T) {
	tests := []struct {
		name     string
		metadata Metadata
		checks   map[string]string
	}{
		{
			name: "title only",
			metadata: Metadata{
				Title: "Sample Document",
			},
			checks: map[string]string{
				"/Title": "(Sample Document)",
			},
		},
		{
			name: "author and subject",
			metadata: Metadata{
				Author:  "John Doe",
				Subject: "PDF Metadata Test",
			},
			checks: map[string]string{
				"/Author":  "(John Doe)",
				"/Subject": "(PDF Metadata Test)",
			},
		},
		{
			name: "all standard fields",
			metadata: Metadata{
				Title:    "Complete Metadata",
				Author:   "Jane Smith",
				Subject:  "Testing",
				Keywords: "PDF, metadata, gopdf",
				Creator:  "My App",
				Producer: "gopdf v1.0",
			},
			checks: map[string]string{
				"/Title":    "(Complete Metadata)",
				"/Author":   "(Jane Smith)",
				"/Subject":  "(Testing)",
				"/Keywords": "(PDF, metadata, gopdf)",
				"/Creator":  "(My App)",
				"/Producer": "(gopdf v1.0)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := New()
			doc.AddPage(A4, Portrait)
			doc.SetMetadata(tt.metadata)

			var buf bytes.Buffer
			err := doc.WriteTo(&buf)
			if err != nil {
				t.Fatalf("WriteTo() failed: %v", err)
			}

			output := buf.String()

			for key, expectedValue := range tt.checks {
				if !strings.Contains(output, key) {
					t.Errorf("Output should contain %q", key)
				}
				if !strings.Contains(output, expectedValue) {
					t.Errorf("Output should contain %q", expectedValue)
				}
			}

			// Trailer should contain /Info reference
			if !strings.Contains(output, "/Info") {
				t.Error("Trailer should contain /Info reference")
			}
		})
	}
}

// TestMetadata_CustomFields tests custom metadata fields
func TestMetadata_CustomFields(t *testing.T) {
	doc := New()
	doc.AddPage(A4, Portrait)

	metadata := Metadata{
		Title: "Document with Custom Fields",
		Custom: map[string]string{
			"Department": "Engineering",
			"Project":    "gopdf",
			"Version":    "1.0.0",
		},
	}
	doc.SetMetadata(metadata)

	var buf bytes.Buffer
	err := doc.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo() failed: %v", err)
	}

	output := buf.String()

	// Check standard field
	if !strings.Contains(output, "/Title") {
		t.Error("Output should contain /Title")
	}

	// Check custom fields
	customChecks := map[string]string{
		"/Department": "(Engineering)",
		"/Project":    "(gopdf)",
		"/Version":    "(1.0.0)",
	}

	for key, expectedValue := range customChecks {
		if !strings.Contains(output, key) {
			t.Errorf("Output should contain custom field %q", key)
		}
		if !strings.Contains(output, expectedValue) {
			t.Errorf("Output should contain custom value %q", expectedValue)
		}
	}
}

// TestMetadata_DateFields tests date field formatting
func TestMetadata_DateFields(t *testing.T) {
	doc := New()
	doc.AddPage(A4, Portrait)

	// Use specific time for testing
	creationTime := time.Date(2025, 1, 29, 12, 30, 45, 0, time.FixedZone("JST", 9*3600))
	modTime := time.Date(2025, 1, 29, 13, 45, 30, 0, time.FixedZone("JST", 9*3600))

	metadata := Metadata{
		Title:        "Date Test",
		CreationDate: creationTime,
		ModDate:      modTime,
	}
	doc.SetMetadata(metadata)

	var buf bytes.Buffer
	err := doc.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo() failed: %v", err)
	}

	output := buf.String()

	// Check date format: D:YYYYMMDDHHmmSSOHH'mm'
	expectedCreation := "D:20250129123045+09'00'"
	expectedMod := "D:20250129134530+09'00'"

	if !strings.Contains(output, expectedCreation) {
		t.Errorf("Output should contain creation date %q", expectedCreation)
	}

	if !strings.Contains(output, expectedMod) {
		t.Errorf("Output should contain modification date %q", expectedMod)
	}
}

// TestMetadata_DefaultProducer tests default Producer value
func TestMetadata_DefaultProducer(t *testing.T) {
	doc := New()
	doc.AddPage(A4, Portrait)

	metadata := Metadata{
		Title: "Test Default Producer",
		// Producer is not set
	}
	doc.SetMetadata(metadata)

	var buf bytes.Buffer
	err := doc.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo() failed: %v", err)
	}

	output := buf.String()

	// Should contain default Producer
	if !strings.Contains(output, "/Producer") {
		t.Error("Output should contain /Producer")
	}

	if !strings.Contains(output, "(gopdf)") {
		t.Error("Output should contain default Producer value 'gopdf'")
	}
}

// TestMetadata_DefaultCreationDate tests default CreationDate value
func TestMetadata_DefaultCreationDate(t *testing.T) {
	doc := New()
	doc.AddPage(A4, Portrait)

	metadata := Metadata{
		Title: "Test Default CreationDate",
		// CreationDate is not set (zero value)
	}
	doc.SetMetadata(metadata)

	var buf bytes.Buffer
	err := doc.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo() failed: %v", err)
	}

	output := buf.String()

	// Should contain CreationDate
	if !strings.Contains(output, "/CreationDate") {
		t.Error("Output should contain /CreationDate")
	}

	// Should contain date in PDF format (D:YYYY...)
	if !strings.Contains(output, "D:") {
		t.Error("Output should contain date in PDF format")
	}
}

// TestMetadata_SpecialCharacters tests escaping of special characters
func TestMetadata_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "parentheses",
			input:    "Title (with parentheses)",
			expected: "Title \\(with parentheses\\)",
		},
		{
			name:     "backslash",
			input:    "Path\\to\\file",
			expected: "Path\\\\to\\\\file",
		},
		{
			name:     "mixed",
			input:    "Text with (parens) and \\ backslash",
			expected: "Text with \\(parens\\) and \\\\ backslash",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := New()
			doc.AddPage(A4, Portrait)

			metadata := Metadata{
				Title: tt.input,
			}
			doc.SetMetadata(metadata)

			var buf bytes.Buffer
			err := doc.WriteTo(&buf)
			if err != nil {
				t.Fatalf("WriteTo() failed: %v", err)
			}

			output := buf.String()

			if !strings.Contains(output, tt.expected) {
				t.Errorf("Output should contain escaped string %q, got output:\n%s",
					tt.expected, output)
			}
		})
	}
}

// TestMetadata_NonASCII tests encoding of non-ASCII characters
func TestMetadata_NonASCII(t *testing.T) {
	doc := New()
	doc.AddPage(A4, Portrait)

	metadata := Metadata{
		Title:   "日本語タイトル",
		Author:  "田中太郎",
		Subject: "テスト文書",
	}
	doc.SetMetadata(metadata)

	var buf bytes.Buffer
	err := doc.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo() failed: %v", err)
	}

	output := buf.String()

	// Non-ASCII strings should be encoded as UTF-16BE with BOM
	// The output should contain the hex encoded string starting with FEFF (BOM)
	if !strings.Contains(output, "<FEFF") {
		t.Error("Output should contain UTF-16BE BOM (FEFF) for non-ASCII text")
	}
}

// TestMetadata_NoMetadata tests that PDF works without metadata
func TestMetadata_NoMetadata(t *testing.T) {
	doc := New()
	doc.AddPage(A4, Portrait)
	// No metadata set

	var buf bytes.Buffer
	err := doc.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo() failed: %v", err)
	}

	output := buf.String()

	// Should not contain /Info in trailer
	trailerIndex := strings.Index(output, "trailer")
	if trailerIndex == -1 {
		t.Fatal("Output should contain trailer")
	}

	trailerSection := output[trailerIndex:]
	if strings.Contains(trailerSection, "/Info") {
		t.Error("Trailer should not contain /Info when no metadata is set")
	}
}

// TestMetadata_WithEncryption tests metadata with encryption
func TestMetadata_WithEncryption(t *testing.T) {
	doc := New()
	doc.AddPage(A4, Portrait)

	// Set metadata
	metadata := Metadata{
		Title:  "Encrypted Document",
		Author: "Test User",
	}
	doc.SetMetadata(metadata)

	// Set encryption
	encryptOpts := EncryptionOptions{
		UserPassword:  "user123",
		OwnerPassword: "owner456",
		Permissions: Permissions{
			Print:  true,
			Modify: false,
			Copy:   false,
		},
		KeyLength: 128,
	}
	err := doc.SetEncryption(encryptOpts)
	if err != nil {
		t.Fatalf("SetEncryption() failed: %v", err)
	}

	var buf bytes.Buffer
	err = doc.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo() failed: %v", err)
	}

	output := buf.String()

	// Should contain both /Info and /Encrypt
	if !strings.Contains(output, "/Info") {
		t.Error("Output should contain /Info")
	}

	if !strings.Contains(output, "/Encrypt") {
		t.Error("Output should contain /Encrypt")
	}
}
