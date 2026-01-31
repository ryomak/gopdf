package gopdf

import (
	"bytes"
	"testing"
	"time"
)

// TestPDFReader_Info_ReadWriteRoundtrip tests reading metadata written by gopdf
func TestPDFReader_Info_ReadWriteRoundtrip(t *testing.T) {
	// Create a PDF with metadata
	doc := New()
	doc.AddPage(PageSizeA4, Portrait)

	metadata := Metadata{
		Title:        "Test Document",
		Author:       "Test Author",
		Subject:      "Test Subject",
		Keywords:     "test, metadata, gopdf",
		Creator:      "gopdf Test",
		Producer:     "gopdf v1.0",
		CreationDate: time.Date(2025, 1, 29, 12, 30, 45, 0, time.UTC),
		ModDate:      time.Date(2025, 1, 29, 13, 45, 30, 0, time.UTC),
	}
	doc.SetMetadata(metadata)

	// Write to buffer
	var buf bytes.Buffer
	err := doc.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo() failed: %v", err)
	}

	// Read back
	reader, err := OpenReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("OpenReader() failed: %v", err)
	}
	defer reader.Close()

	// Get metadata
	readMetadata := reader.Info()

	// Verify standard fields
	if readMetadata.Title != metadata.Title {
		t.Errorf("Title = %q, want %q", readMetadata.Title, metadata.Title)
	}

	if readMetadata.Author != metadata.Author {
		t.Errorf("Author = %q, want %q", readMetadata.Author, metadata.Author)
	}

	if readMetadata.Subject != metadata.Subject {
		t.Errorf("Subject = %q, want %q", readMetadata.Subject, metadata.Subject)
	}

	if readMetadata.Keywords != metadata.Keywords {
		t.Errorf("Keywords = %q, want %q", readMetadata.Keywords, metadata.Keywords)
	}

	if readMetadata.Creator != metadata.Creator {
		t.Errorf("Creator = %q, want %q", readMetadata.Creator, metadata.Creator)
	}

	if readMetadata.Producer != metadata.Producer {
		t.Errorf("Producer = %q, want %q", readMetadata.Producer, metadata.Producer)
	}

	// Verify dates (within 1 second tolerance due to formatting)
	if !readMetadata.CreationDate.IsZero() {
		diff := readMetadata.CreationDate.Sub(metadata.CreationDate)
		if diff < 0 {
			diff = -diff
		}
		if diff > time.Second {
			t.Errorf("CreationDate = %v, want %v (diff: %v)", readMetadata.CreationDate, metadata.CreationDate, diff)
		}
	} else {
		t.Error("CreationDate is zero")
	}

	if !readMetadata.ModDate.IsZero() {
		diff := readMetadata.ModDate.Sub(metadata.ModDate)
		if diff < 0 {
			diff = -diff
		}
		if diff > time.Second {
			t.Errorf("ModDate = %v, want %v (diff: %v)", readMetadata.ModDate, metadata.ModDate, diff)
		}
	} else {
		t.Error("ModDate is zero")
	}
}

// TestPDFReader_Info_CustomFields tests reading custom metadata fields
func TestPDFReader_Info_CustomFields(t *testing.T) {
	// Create a PDF with custom metadata
	doc := New()
	doc.AddPage(PageSizeA4, Portrait)

	metadata := Metadata{
		Title: "Document with Custom Fields",
		Custom: map[string]string{
			"Department": "Engineering",
			"Project":    "gopdf",
			"Version":    "1.0.0",
		},
	}
	doc.SetMetadata(metadata)

	// Write to buffer
	var buf bytes.Buffer
	err := doc.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo() failed: %v", err)
	}

	// Read back
	reader, err := OpenReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("OpenReader() failed: %v", err)
	}
	defer reader.Close()

	// Get metadata
	readMetadata := reader.Info()

	// Verify title
	if readMetadata.Title != metadata.Title {
		t.Errorf("Title = %q, want %q", readMetadata.Title, metadata.Title)
	}

	// Verify custom fields
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

// TestPDFReader_Info_NonASCII tests reading non-ASCII metadata
func TestPDFReader_Info_NonASCII(t *testing.T) {
	// Create a PDF with Japanese metadata
	doc := New()
	doc.AddPage(PageSizeA4, Portrait)

	metadata := Metadata{
		Title:   "日本語タイトル",
		Author:  "田中太郎",
		Subject: "テスト文書",
	}
	doc.SetMetadata(metadata)

	// Write to buffer
	var buf bytes.Buffer
	err := doc.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo() failed: %v", err)
	}

	// Read back
	reader, err := OpenReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("OpenReader() failed: %v", err)
	}
	defer reader.Close()

	// Get metadata
	readMetadata := reader.Info()

	// Verify Japanese text
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

// TestPDFReader_Info_NoMetadata tests reading PDF without metadata
func TestPDFReader_Info_NoMetadata(t *testing.T) {
	// Create a PDF without metadata
	doc := New()
	doc.AddPage(PageSizeA4, Portrait)

	// Write to buffer
	var buf bytes.Buffer
	err := doc.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo() failed: %v", err)
	}

	// Read back
	reader, err := OpenReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("OpenReader() failed: %v", err)
	}
	defer reader.Close()

	// Get metadata
	metadata := reader.Info()

	// All fields should be empty or zero
	if metadata.Title != "" {
		t.Errorf("Title should be empty, got %q", metadata.Title)
	}

	if metadata.Author != "" {
		t.Errorf("Author should be empty, got %q", metadata.Author)
	}

	if !metadata.CreationDate.IsZero() {
		t.Errorf("CreationDate should be zero, got %v", metadata.CreationDate)
	}

	if len(metadata.Custom) > 0 {
		t.Errorf("Custom fields should be empty, got %v", metadata.Custom)
	}
}

// TestPDFReader_Info_SpecialCharacters tests reading metadata with special characters
func TestPDFReader_Info_SpecialCharacters(t *testing.T) {
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
			// Create PDF
			doc := New()
			doc.AddPage(PageSizeA4, Portrait)

			metadata := Metadata{
				Title: tt.input,
			}
			doc.SetMetadata(metadata)

			// Write to buffer
			var buf bytes.Buffer
			err := doc.WriteTo(&buf)
			if err != nil {
				t.Fatalf("WriteTo() failed: %v", err)
			}

			// Read back
			reader, err := OpenReader(bytes.NewReader(buf.Bytes()))
			if err != nil {
				t.Fatalf("OpenReader() failed: %v", err)
			}
			defer reader.Close()

			// Get metadata
			readMetadata := reader.Info()

			// Verify
			if readMetadata.Title != tt.input {
				t.Errorf("Title = %q, want %q", readMetadata.Title, tt.input)
			}
		})
	}
}
