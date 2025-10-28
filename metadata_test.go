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
	}{
		{
			name: "title only",
			metadata: Metadata{
				Title: "Sample Document",
			},
		},
		{
			name: "author and subject",
			metadata: Metadata{
				Author:  "John Doe",
				Subject: "PDF Metadata Test",
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

			// Trailer should contain /Info reference
			if !strings.Contains(output, "/Info") {
				t.Error("Trailer should contain /Info reference")
			}

			// Verify by round-trip reading
			reader, err := OpenReader(bytes.NewReader(buf.Bytes()))
			if err != nil {
				t.Fatalf("OpenReader() failed: %v", err)
			}
			defer reader.Close()

			readMetadata := reader.Info()

			if readMetadata.Title != tt.metadata.Title {
				t.Errorf("Title = %q, want %q", readMetadata.Title, tt.metadata.Title)
			}
			if readMetadata.Author != tt.metadata.Author {
				t.Errorf("Author = %q, want %q", readMetadata.Author, tt.metadata.Author)
			}
			if readMetadata.Subject != tt.metadata.Subject {
				t.Errorf("Subject = %q, want %q", readMetadata.Subject, tt.metadata.Subject)
			}
			if readMetadata.Keywords != tt.metadata.Keywords {
				t.Errorf("Keywords = %q, want %q", readMetadata.Keywords, tt.metadata.Keywords)
			}
			if readMetadata.Creator != tt.metadata.Creator {
				t.Errorf("Creator = %q, want %q", readMetadata.Creator, tt.metadata.Creator)
			}
			if tt.metadata.Producer != "" && readMetadata.Producer != tt.metadata.Producer {
				t.Errorf("Producer = %q, want %q", readMetadata.Producer, tt.metadata.Producer)
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

	// Check that /Info exists
	if !strings.Contains(output, "/Info") {
		t.Error("Output should contain /Info")
	}

	// Verify by round-trip reading
	reader, err := OpenReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("OpenReader() failed: %v", err)
	}
	defer reader.Close()

	readMetadata := reader.Info()

	if readMetadata.Title != metadata.Title {
		t.Errorf("Title = %q, want %q", readMetadata.Title, metadata.Title)
	}

	if readMetadata.Custom == nil {
		t.Fatal("Custom fields are nil")
	}

	for key, expectedValue := range metadata.Custom {
		actualValue, ok := readMetadata.Custom[key]
		if !ok {
			t.Errorf("Custom field %q not found", key)
			continue
		}
		if actualValue != expectedValue {
			t.Errorf("Custom field %q = %q, want %q", key, actualValue, expectedValue)
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

	// Verify by round-trip reading
	reader, err := OpenReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("OpenReader() failed: %v", err)
	}
	defer reader.Close()

	readMetadata := reader.Info()

	// Verify dates (within 1 second tolerance due to formatting)
	if !readMetadata.CreationDate.IsZero() {
		diff := readMetadata.CreationDate.Sub(creationTime)
		if diff < 0 {
			diff = -diff
		}
		if diff > time.Second {
			t.Errorf("CreationDate = %v, want %v (diff: %v)", readMetadata.CreationDate, creationTime, diff)
		}
	} else {
		t.Error("CreationDate is zero")
	}

	if !readMetadata.ModDate.IsZero() {
		diff := readMetadata.ModDate.Sub(modTime)
		if diff < 0 {
			diff = -diff
		}
		if diff > time.Second {
			t.Errorf("ModDate = %v, want %v (diff: %v)", readMetadata.ModDate, modTime, diff)
		}
	} else {
		t.Error("ModDate is zero")
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

	// Verify by round-trip reading
	reader, err := OpenReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("OpenReader() failed: %v", err)
	}
	defer reader.Close()

	readMetadata := reader.Info()

	if readMetadata.Producer != "gopdf" {
		t.Errorf("Producer = %q, want %q", readMetadata.Producer, "gopdf")
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

	// Verify by round-trip reading
	reader, err := OpenReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("OpenReader() failed: %v", err)
	}
	defer reader.Close()

	readMetadata := reader.Info()

	// Should have a creation date (automatically set to current time)
	if readMetadata.CreationDate.IsZero() {
		t.Error("CreationDate should be set automatically")
	}
}

// TestMetadata_SpecialCharacters tests escaping of special characters
func TestMetadata_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "parentheses",
			input: "Title (with parentheses)",
		},
		{
			name:  "backslash",
			input: "Path\\to\\file",
		},
		{
			name:  "mixed",
			input: "Text with (parens) and \\ backslash",
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

			// Verify by round-trip reading
			reader, err := OpenReader(bytes.NewReader(buf.Bytes()))
			if err != nil {
				t.Fatalf("OpenReader() failed: %v", err)
			}
			defer reader.Close()

			readMetadata := reader.Info()

			if readMetadata.Title != tt.input {
				t.Errorf("Title = %q, want %q", readMetadata.Title, tt.input)
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

	// Verify by round-trip reading
	reader, err := OpenReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("OpenReader() failed: %v", err)
	}
	defer reader.Close()

	readMetadata := reader.Info()

	if readMetadata.Title != metadata.Title {
		t.Errorf("Title = %q, want %q", readMetadata.Title, metadata.Title)
	}

	if readMetadata.Author != metadata.Author {
		t.Errorf("Author = %q, want %q", readMetadata.Author, metadata.Author)
	}

	if readMetadata.Subject != metadata.Subject {
		t.Errorf("Subject = %q, want %q", readMetadata.Subject, metadata.Subject)
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
