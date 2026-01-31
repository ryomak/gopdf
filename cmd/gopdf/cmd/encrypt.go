package cmd

import (
	"fmt"
	"os"

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

	if encryptKeyLength != 40 && encryptKeyLength != 128 {
		return fmt.Errorf("key length must be 40 or 128")
	}

	if encryptOwnerPassword == "" {
		encryptOwnerPassword = encryptUserPassword
	}

	reader, err := gopdf.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input PDF: %w", err)
	}
	defer reader.Close()

	doc := gopdf.New()

	for i := 0; i < reader.PageCount(); i++ {
		layout, err := reader.ExtractPageLayout(i)
		if err != nil {
			return fmt.Errorf("failed to extract layout from page %d: %w", i+1, err)
		}

		pageSize := gopdf.PageSize{Width: layout.Width, Height: layout.Height}
		_, err = gopdf.RenderLayout(doc, layout, gopdf.PDFTranslatorOptions{
			TargetFont: gopdf.FontHelvetica,
			KeepImages: true,
			KeepLayout: true,
			FittingOptions: gopdf.FitOptions{
				MaxFontSize: 72,
				MinFontSize: 6,
				AllowShrink: true,
				AllowGrow:   false,
			},
		})
		if err != nil {
			_ = pageSize
			return fmt.Errorf("failed to render page %d: %w", i+1, err)
		}
	}

	permissions := gopdf.DefaultPermissions()
	if encryptNoPrint {
		permissions.Print = false
		permissions.PrintHighQuality = false
	}
	if encryptNoCopy {
		permissions.Copy = false
		permissions.ExtractContent = false
	}
	if encryptNoModify {
		permissions.Modify = false
		permissions.Annotate = false
		permissions.FillForms = false
		permissions.Assemble = false
	}

	err = doc.SetEncryption(gopdf.EncryptionOptions{
		UserPassword:  encryptUserPassword,
		OwnerPassword: encryptOwnerPassword,
		Permissions:   permissions,
		KeyLength:     encryptKeyLength,
	})
	if err != nil {
		return fmt.Errorf("failed to set encryption: %w", err)
	}

	output, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer output.Close()

	if err := doc.WriteTo(output); err != nil {
		return fmt.Errorf("failed to write encrypted PDF: %w", err)
	}

	if !quiet {
		fmt.Printf("Encrypted PDF created: %s\n", outputPath)
		fmt.Printf("  Key length: %d-bit\n", encryptKeyLength)
		fmt.Printf("  User password: %s\n", encryptUserPassword)
		if encryptOwnerPassword != encryptUserPassword {
			fmt.Printf("  Owner password: %s\n", encryptOwnerPassword)
		}
	}

	return nil
}
