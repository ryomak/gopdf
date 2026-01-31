package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/ryomak/gopdf"
	"github.com/spf13/cobra"
)

var (
	encryptUserPassword  string
	encryptOwnerPassword string
	encryptKeyLength     int
	encryptNoPrint       bool
	encryptNoCopy        bool
	encryptNoModify      bool
)

var encryptCmd = &cobra.Command{
	Use:   "encrypt <input.pdf> <output.pdf>",
	Short: "Encrypt a PDF file",
	Long: `Encrypt a PDF file with password protection.

This command reads an existing PDF, extracts its content, and creates
a new encrypted PDF with the specified password and permissions.

Note: Complex layouts may not be perfectly preserved due to PDF reconstruction.`,
	Args: cobra.ExactArgs(2),
	RunE: runEncrypt,
}

func init() {
	rootCmd.AddCommand(encryptCmd)

	encryptCmd.Flags().StringVar(&encryptUserPassword, "user-password", "", "password required to open the PDF")
	encryptCmd.Flags().StringVar(&encryptOwnerPassword, "owner-password", "", "password for full access (defaults to user password)")
	encryptCmd.Flags().IntVar(&encryptKeyLength, "key-length", 128, "encryption key length (40 or 128)")
	encryptCmd.Flags().BoolVar(&encryptNoPrint, "no-print", false, "disable printing")
	encryptCmd.Flags().BoolVar(&encryptNoCopy, "no-copy", false, "disable copying text/images")
	encryptCmd.Flags().BoolVar(&encryptNoModify, "no-modify", false, "disable modification")

	_ = encryptCmd.MarkFlagRequired("user-password")
}

func runEncrypt(cmd *cobra.Command, args []string) error {
	inputPath := args[0]
	outputPath := args[1]
	log := NewLogger()

	log.Header(iconLock, "Encrypt PDF")

	if encryptKeyLength != 40 && encryptKeyLength != 128 {
		log.Error("Key length must be 40 or 128")
		return fmt.Errorf("key length must be 40 or 128")
	}

	if encryptOwnerPassword == "" {
		encryptOwnerPassword = encryptUserPassword
	}

	log.Step("Opening %s", filepath.Base(inputPath))

	// Build permissions
	permissions := gopdf.DefaultPermissions()
	var restrictions []string

	if encryptNoPrint {
		permissions.Print = false
		permissions.PrintHighQuality = false
		restrictions = append(restrictions, "print")
	}
	if encryptNoCopy {
		permissions.Copy = false
		permissions.ExtractContent = false
		restrictions = append(restrictions, "copy")
	}
	if encryptNoModify {
		permissions.Modify = false
		permissions.Annotate = false
		permissions.FillForms = false
		permissions.Assemble = false
		restrictions = append(restrictions, "modify")
	}

	log.Step("Encrypting PDF (preserving original content)...")

	// Use the new EncryptExistingPDF function that preserves the original PDF
	err := gopdf.EncryptExistingPDF(inputPath, outputPath, gopdf.EncryptionOptions{
		UserPassword:  encryptUserPassword,
		OwnerPassword: encryptOwnerPassword,
		Permissions:   permissions,
		KeyLength:     encryptKeyLength,
	})
	if err != nil {
		log.Error("Failed to encrypt PDF: %v", err)
		return fmt.Errorf("failed to encrypt PDF: %w", err)
	}

	log.Divider()
	log.Success("Encrypted PDF created: %s", outputPath)
	log.Section("🔐 Encryption Details")
	log.Table("Algorithm", fmt.Sprintf("RC4 %d-bit", encryptKeyLength))
	log.Table("User Password", maskPassword(encryptUserPassword))
	if encryptOwnerPassword != encryptUserPassword {
		log.Table("Owner Password", maskPassword(encryptOwnerPassword))
	}
	if len(restrictions) > 0 {
		log.Table("Restrictions", fmt.Sprintf("%v", restrictions))
	} else {
		log.Table("Permissions", "All allowed")
	}
	log.Println()

	return nil
}

func maskPassword(pwd string) string {
	if len(pwd) <= 2 {
		return "***"
	}
	return pwd[:1] + "***" + pwd[len(pwd)-1:]
}
