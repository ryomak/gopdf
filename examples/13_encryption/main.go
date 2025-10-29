package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ryomak/gopdf"
)

func main() {
	// Example 1: Basic password protection with default permissions
	if err := example1BasicPassword(); err != nil {
		log.Fatalf("Example 1 failed: %v", err)
	}

	// Example 2: Restricted permissions (view only)
	if err := example2RestrictedPermissions(); err != nil {
		log.Fatalf("Example 2 failed: %v", err)
	}

	// Example 3: Custom permissions (print and copy only)
	if err := example3CustomPermissions(); err != nil {
		log.Fatalf("Example 3 failed: %v", err)
	}

	// Example 4: 128-bit encryption for stronger security
	if err := example4StrongEncryption(); err != nil {
		log.Fatalf("Example 4 failed: %v", err)
	}

	fmt.Println("All encryption examples completed successfully!")
}

// Example 1: Basic password protection with default permissions (all allowed)
func example1BasicPassword() error {
	fmt.Println("Creating Example 1: Basic password protection...")

	// Create a new PDF document
	doc := gopdf.New()
	page := doc.AddPage(gopdf.A4, gopdf.Portrait)

	// Add some text
	if err := page.SetFont(gopdf.Helvetica, 24); err != nil {
		return err
	}
	page.DrawText("Password Protected PDF", 100, 700)

	page.SetFont(gopdf.Helvetica, 12)
	page.DrawText("This PDF requires a password to open.", 100, 670)
	page.DrawText("User password: user123", 100, 650)
	page.DrawText("Owner password: owner123", 100, 630)
	page.DrawText("All permissions are allowed.", 100, 610)

	// Set encryption with basic password protection
	// User can open with user password, owner can open with owner password
	err := doc.SetEncryption(gopdf.EncryptionOptions{
		UserPassword:  "user123",
		OwnerPassword: "owner123",
		Permissions:   gopdf.DefaultPermissions(), // All permissions allowed
		KeyLength:     40,                         // 40-bit RC4 encryption
	})
	if err != nil {
		return fmt.Errorf("failed to set encryption: %w", err)
	}

	// Write to file
	f, err := os.Create("example1_basic_password.pdf")
	if err != nil {
		return err
	}
	defer f.Close()

	if err := doc.WriteTo(f); err != nil {
		return err
	}

	fmt.Println("  Created: example1_basic_password.pdf")
	return nil
}

// Example 2: Restricted permissions (view only, no printing or copying)
func example2RestrictedPermissions() error {
	fmt.Println("Creating Example 2: Restricted permissions...")

	doc := gopdf.New()
	page := doc.AddPage(gopdf.A4, gopdf.Portrait)

	page.SetFont(gopdf.Helvetica, 24)
	page.DrawText("View-Only PDF", 100, 700)

	page.SetFont(gopdf.Helvetica, 12)
	page.DrawText("This PDF can be opened but cannot be printed or copied.", 100, 670)
	page.DrawText("User password: user123", 100, 650)
	page.DrawText("All operations are restricted.", 100, 630)

	// Set encryption with restricted permissions
	err := doc.SetEncryption(gopdf.EncryptionOptions{
		UserPassword:  "user123",
		OwnerPassword: "owner123",
		Permissions:   gopdf.RestrictedPermissions(), // No permissions allowed
		KeyLength:     40,
	})
	if err != nil {
		return fmt.Errorf("failed to set encryption: %w", err)
	}

	f, err := os.Create("example2_restricted.pdf")
	if err != nil {
		return err
	}
	defer f.Close()

	if err := doc.WriteTo(f); err != nil {
		return err
	}

	fmt.Println("  Created: example2_restricted.pdf")
	return nil
}

// Example 3: Custom permissions (allow printing and copying, but not modification)
func example3CustomPermissions() error {
	fmt.Println("Creating Example 3: Custom permissions...")

	doc := gopdf.New()
	page := doc.AddPage(gopdf.A4, gopdf.Portrait)

	page.SetFont(gopdf.Helvetica, 24)
	page.DrawText("Custom Permissions PDF", 100, 700)

	page.SetFont(gopdf.Helvetica, 12)
	page.DrawText("This PDF allows printing and copying only.", 100, 670)
	page.DrawText("User password: user123", 100, 650)
	page.DrawText("Allowed: Print, Copy", 100, 630)
	page.DrawText("Denied: Modify, Annotate, etc.", 100, 610)

	// Create custom permissions
	permissions := gopdf.Permissions{
		Print:            true,  // Allow printing
		Modify:           false, // Deny modification
		Copy:             true,  // Allow copying text/graphics
		Annotate:         false, // Deny annotations
		FillForms:        false,
		ExtractContent:   true, // Allow extraction for accessibility
		Assemble:         false,
		PrintHighQuality: true,
	}

	err := doc.SetEncryption(gopdf.EncryptionOptions{
		UserPassword:  "user123",
		OwnerPassword: "owner123",
		Permissions:   permissions,
		KeyLength:     40,
	})
	if err != nil {
		return fmt.Errorf("failed to set encryption: %w", err)
	}

	f, err := os.Create("example3_custom_permissions.pdf")
	if err != nil {
		return err
	}
	defer f.Close()

	if err := doc.WriteTo(f); err != nil {
		return err
	}

	fmt.Println("  Created: example3_custom_permissions.pdf")
	return nil
}

// Example 4: 128-bit encryption for stronger security
func example4StrongEncryption() error {
	fmt.Println("Creating Example 4: Strong encryption (128-bit)...")

	doc := gopdf.New()
	page := doc.AddPage(gopdf.A4, gopdf.Portrait)

	page.SetFont(gopdf.Helvetica, 24)
	page.DrawText("128-bit Encrypted PDF", 100, 700)

	page.SetFont(gopdf.Helvetica, 12)
	page.DrawText("This PDF uses 128-bit RC4 encryption.", 100, 670)
	page.DrawText("User password: strongpass", 100, 650)
	page.DrawText("Owner password: ownerpass", 100, 630)
	page.DrawText("Higher security than 40-bit encryption.", 100, 610)

	// Set 128-bit encryption
	err := doc.SetEncryption(gopdf.EncryptionOptions{
		UserPassword:  "strongpass",
		OwnerPassword: "ownerpass",
		Permissions:   gopdf.PrintOnlyPermissions(), // Allow printing only
		KeyLength:     128,                          // 128-bit RC4 encryption
	})
	if err != nil {
		return fmt.Errorf("failed to set encryption: %w", err)
	}

	f, err := os.Create("example4_strong_encryption.pdf")
	if err != nil {
		return err
	}
	defer f.Close()

	if err := doc.WriteTo(f); err != nil {
		return err
	}

	fmt.Println("  Created: example4_strong_encryption.pdf")
	return nil
}
