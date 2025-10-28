package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ryomak/gopdf/internal/reader"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go <pdf-file> <password>")
		fmt.Println("\nExample: go run main.go ../../examples/13_encryption/example1_basic_password.pdf user123")
		os.Exit(1)
	}

	pdfPath := os.Args[1]
	password := os.Args[2]

	// Open the PDF file
	file, err := os.Open(pdfPath)
	if err != nil {
		log.Fatalf("Failed to open PDF: %v", err)
	}
	defer file.Close()

	// Create a new reader
	r, err := reader.NewReader(file)
	if err != nil {
		log.Fatalf("Failed to create reader: %v", err)
	}

	// Check if the PDF is encrypted
	if !r.IsEncrypted() {
		fmt.Println("✓ PDF is not encrypted")
		fmt.Println("  No password required to read this PDF")
		displayPDFInfo(r)
		return
	}

	fmt.Println("🔒 PDF is encrypted")
	fmt.Printf("   Attempting to authenticate with password: %s\n", password)

	// Authenticate with password
	err = r.AuthenticateWithPassword(password)
	if err != nil {
		log.Fatalf("❌ Authentication failed: %v", err)
	}

	fmt.Println("✓ Authentication successful!")

	// Get encryption info
	encInfo := r.GetEncryptionInfo()
	if encInfo != nil {
		fmt.Println("\n📋 Encryption Details:")
		fmt.Printf("   Filter: %s\n", encInfo.Filter)
		fmt.Printf("   Algorithm Version (V): %d\n", encInfo.V)
		fmt.Printf("   Revision (R): %d\n", encInfo.R)
		fmt.Printf("   Key Length: %d bits\n", encInfo.Length)
		fmt.Printf("   Authenticated as: %s\n", map[bool]string{
			true:  "Owner (full access)",
			false: "User (restricted by permissions)",
		}[encInfo.IsOwner])

		// Display permissions
		fmt.Println("\n🔐 Permissions:")
		perms := encInfo.P
		displayPermission("Print", (perms & (1 << 2)) != 0)
		displayPermission("Modify", (perms & (1 << 3)) != 0)
		displayPermission("Copy/Extract", (perms & (1 << 4)) != 0)
		displayPermission("Annotate", (perms & (1 << 5)) != 0)
		displayPermission("Fill Forms", (perms & (1 << 8)) != 0)
		displayPermission("Extract for Accessibility", (perms & (1 << 9)) != 0)
		displayPermission("Assemble Document", (perms & (1 << 10)) != 0)
		displayPermission("Print High Quality", (perms & (1 << 11)) != 0)
	}

	// Display PDF information
	displayPDFInfo(r)
}

func displayPermission(name string, allowed bool) {
	status := "❌ Denied"
	if allowed {
		status = "✓ Allowed"
	}
	fmt.Printf("   %s: %s\n", name, status)
}

func displayPDFInfo(r *reader.Reader) {
	fmt.Println("\n📄 PDF Information:")

	// Get catalog (just to verify it's readable)
	_, err := r.GetCatalog()
	if err != nil {
		fmt.Printf("   Error reading catalog: %v\n", err)
		return
	}

	// Get page count
	pageCount, err := r.GetPageCount()
	if err != nil {
		fmt.Printf("   Error getting page count: %v\n", err)
		return
	}
	fmt.Printf("   Page Count: %d\n", pageCount)

	// Get info dictionary
	info, err := r.GetInfo()
	if err != nil {
		fmt.Printf("   Error reading info: %v\n", err)
		return
	}

	// Display metadata
	if len(info) > 0 {
		fmt.Println("\n📝 Metadata:")
		for key, value := range info {
			// Try to convert to string
			fmt.Printf("   %s: %v\n", key, value)
		}
	}

	// Try to read first page content
	fmt.Println("\n📖 Reading First Page:")
	page, err := r.GetPage(0)
	if err != nil {
		fmt.Printf("   Error reading page: %v\n", err)
		return
	}

	fmt.Printf("   ✓ Successfully read page 1\n")
	fmt.Printf("   Page object has %d keys\n", len(page))

	// Check if page has content
	if contentRef, ok := page["Contents"]; ok {
		fmt.Printf("   ✓ Page has content stream\n")
		_ = contentRef // Just verify it exists
	}

	fmt.Println("\n✅ Successfully decrypted and read encrypted PDF!")
}
