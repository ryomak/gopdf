package gopdf

import (
	"bytes"
	"testing"
)

func TestEncryptExistingPDF(t *testing.T) {
	// Create a test PDF in memory
	doc := New()
	page := doc.AddPage(PageSizeA4, Portrait)
	_ = page.SetFont(FontHelvetica, 12)
	_ = page.DrawText("Hello World", 50, 800)
	_ = page.DrawText("Test Document", 50, 780)

	var originalBuf bytes.Buffer
	if err := doc.WriteTo(&originalBuf); err != nil {
		t.Fatalf("Failed to create original PDF: %v", err)
	}

	// Encrypt the PDF
	opts := EncryptionOptions{
		UserPassword:  "user123",
		OwnerPassword: "owner456",
		Permissions:   DefaultPermissions(),
		KeyLength:     128,
	}

	var encryptedBuf bytes.Buffer
	if err := EncryptExistingPDFReader(bytes.NewReader(originalBuf.Bytes()), &encryptedBuf, opts); err != nil {
		t.Fatalf("Failed to encrypt PDF: %v", err)
	}

	// Verify the encrypted PDF can be opened
	reader, err := OpenReader(bytes.NewReader(encryptedBuf.Bytes()))
	if err != nil {
		t.Fatalf("Failed to open encrypted PDF: %v", err)
	}
	defer reader.Close()

	// Verify it's encrypted
	if !reader.IsEncrypted() {
		t.Error("PDF should be encrypted")
	}

	// Authenticate with password
	if err := reader.AuthenticateWithPassword("user123"); err != nil {
		t.Fatalf("Failed to authenticate: %v", err)
	}

	// Verify page count
	if reader.PageCount() != 1 {
		t.Errorf("Expected 1 page, got %d", reader.PageCount())
	}
}

func TestEncryptExistingPDF_40bit(t *testing.T) {
	// Create a test PDF in memory
	doc := New()
	page := doc.AddPage(PageSizeA4, Portrait)
	_ = page.SetFont(FontHelvetica, 12)
	_ = page.DrawText("Test 40-bit encryption", 50, 800)

	var originalBuf bytes.Buffer
	if err := doc.WriteTo(&originalBuf); err != nil {
		t.Fatalf("Failed to create original PDF: %v", err)
	}

	// Encrypt with 40-bit
	opts := EncryptionOptions{
		UserPassword:  "password",
		OwnerPassword: "password",
		Permissions:   DefaultPermissions(),
		KeyLength:     40,
	}

	var encryptedBuf bytes.Buffer
	if err := EncryptExistingPDFReader(bytes.NewReader(originalBuf.Bytes()), &encryptedBuf, opts); err != nil {
		t.Fatalf("Failed to encrypt PDF: %v", err)
	}

	// Verify the encrypted PDF
	reader, err := OpenReader(bytes.NewReader(encryptedBuf.Bytes()))
	if err != nil {
		t.Fatalf("Failed to open encrypted PDF: %v", err)
	}
	defer reader.Close()

	if !reader.IsEncrypted() {
		t.Error("PDF should be encrypted")
	}

	info := reader.GetEncryptionInfo()
	if info.R != 2 {
		t.Errorf("Expected revision 2 for 40-bit, got %d", info.R)
	}
}

func TestEncryptExistingPDF_AlreadyEncrypted(t *testing.T) {
	// Create and encrypt a PDF
	doc := New()
	page := doc.AddPage(PageSizeA4, Portrait)
	_ = page.SetFont(FontHelvetica, 12)
	_ = page.DrawText("Already encrypted", 50, 800)

	_ = doc.SetEncryption(EncryptionOptions{
		UserPassword:  "test",
		OwnerPassword: "test",
		Permissions:   DefaultPermissions(),
		KeyLength:     128,
	})

	var encryptedBuf bytes.Buffer
	if err := doc.WriteTo(&encryptedBuf); err != nil {
		t.Fatalf("Failed to create encrypted PDF: %v", err)
	}

	// Try to encrypt again - should fail
	opts := EncryptionOptions{
		UserPassword:  "new",
		OwnerPassword: "new",
		Permissions:   DefaultPermissions(),
		KeyLength:     128,
	}

	var output bytes.Buffer
	err := EncryptExistingPDFReader(bytes.NewReader(encryptedBuf.Bytes()), &output, opts)
	if err == nil {
		t.Error("Expected error when encrypting already encrypted PDF")
	}
}

func TestEncryptExistingPDF_RestrictedPermissions(t *testing.T) {
	// Create a test PDF
	doc := New()
	page := doc.AddPage(PageSizeA4, Portrait)
	_ = page.SetFont(FontHelvetica, 12)
	_ = page.DrawText("Restricted permissions", 50, 800)

	var originalBuf bytes.Buffer
	if err := doc.WriteTo(&originalBuf); err != nil {
		t.Fatalf("Failed to create original PDF: %v", err)
	}

	// Encrypt with restricted permissions
	opts := EncryptionOptions{
		UserPassword:  "user",
		OwnerPassword: "owner",
		Permissions:   RestrictedPermissions(),
		KeyLength:     128,
	}

	var encryptedBuf bytes.Buffer
	if err := EncryptExistingPDFReader(bytes.NewReader(originalBuf.Bytes()), &encryptedBuf, opts); err != nil {
		t.Fatalf("Failed to encrypt PDF: %v", err)
	}

	// Verify the encrypted PDF
	reader, err := OpenReader(bytes.NewReader(encryptedBuf.Bytes()))
	if err != nil {
		t.Fatalf("Failed to open encrypted PDF: %v", err)
	}
	defer reader.Close()

	if !reader.IsEncrypted() {
		t.Error("PDF should be encrypted")
	}
}
