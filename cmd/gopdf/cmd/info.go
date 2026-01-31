package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ryomak/gopdf"
	"github.com/spf13/cobra"
)

var (
	infoJSON     bool
	infoPassword string
)

type PDFInfo struct {
	FileName   string     `json:"file_name"`
	PageCount  int        `json:"page_count"`
	Pages      []PageInfo `json:"pages"`
	Metadata   *Metadata  `json:"metadata,omitempty"`
	Encrypted  bool       `json:"encrypted"`
}

type PageInfo struct {
	Number int     `json:"number"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type Metadata struct {
	Title        string `json:"title,omitempty"`
	Author       string `json:"author,omitempty"`
	Subject      string `json:"subject,omitempty"`
	Keywords     string `json:"keywords,omitempty"`
	Creator      string `json:"creator,omitempty"`
	Producer     string `json:"producer,omitempty"`
	CreationDate string `json:"creation_date,omitempty"`
	ModDate      string `json:"mod_date,omitempty"`
}

const dateFormat = "2006-01-02 15:04:05"

var infoCmd = &cobra.Command{
	Use:   "info <file.pdf>",
	Short: "Display PDF file information",
	Long: `Display information about a PDF file including:
  - Number of pages
  - Page sizes
  - Metadata (title, author, etc.)
  - Encryption status`,
	Args: cobra.ExactArgs(1),
	RunE: runInfo,
}

func init() {
	rootCmd.AddCommand(infoCmd)
	infoCmd.Flags().BoolVar(&infoJSON, "json", false, "output in JSON format")
	infoCmd.Flags().StringVarP(&infoPassword, "password", "p", "", "password for encrypted PDF")
}

func runInfo(cmd *cobra.Command, args []string) error {
	filePath := args[0]
	log := NewLogger()

	log.Header(iconFile, "PDF Information")
	log.Step("Opening %s", filepath.Base(filePath))

	reader, err := gopdf.Open(filePath)
	if err != nil {
		log.Error("Failed to open PDF: %v", err)
		return fmt.Errorf("failed to open PDF: %w", err)
	}
	defer reader.Close()

	if reader.IsEncrypted() {
		if infoPassword != "" {
			log.Step("Authenticating with password...")
			if err := reader.AuthenticateWithPassword(infoPassword); err != nil {
				log.Error("Authentication failed")
				return fmt.Errorf("failed to authenticate: %w", err)
			}
			log.Success("Authentication successful")
		} else {
			log.Warning("PDF is encrypted (use --password to authenticate)")
		}
	}

	log.Step("Analyzing PDF structure...")

	info := PDFInfo{
		FileName:  filePath,
		PageCount: reader.PageCount(),
		Pages:     make([]PageInfo, 0, reader.PageCount()),
		Encrypted: reader.IsEncrypted(),
	}

	for i := 0; i < reader.PageCount(); i++ {
		layout, err := reader.ExtractPageLayout(i)
		if err != nil {
			return fmt.Errorf("failed to get page %d layout: %w", i+1, err)
		}
		info.Pages = append(info.Pages, PageInfo{
			Number: i + 1,
			Width:  layout.Width,
			Height: layout.Height,
		})
	}

	meta := reader.Info()
	info.Metadata = &Metadata{
		Title:    meta.Title,
		Author:   meta.Author,
		Subject:  meta.Subject,
		Keywords: meta.Keywords,
		Creator:  meta.Creator,
		Producer: meta.Producer,
	}
	if !meta.CreationDate.IsZero() {
		info.Metadata.CreationDate = meta.CreationDate.Format(dateFormat)
	}
	if !meta.ModDate.IsZero() {
		info.Metadata.ModDate = meta.ModDate.Format(dateFormat)
	}

	log.Success("Analysis complete")

	if infoJSON {
		log.Println()
		return outputJSON(info)
	}

	return outputText(info, log)
}

func outputJSON(info PDFInfo) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(info)
}

func outputText(info PDFInfo, log *Logger) error {
	log.Section("📊 Summary")
	log.Table("File", filepath.Base(info.FileName))
	log.Table("Pages", info.PageCount)
	if info.Encrypted {
		log.Table("Encrypted", fmt.Sprintf("%s Yes", iconLock))
	} else {
		log.Table("Encrypted", "No")
	}

	log.Section("📐 Page Sizes")
	for _, page := range info.Pages {
		log.Table(fmt.Sprintf("Page %d", page.Number),
			fmt.Sprintf("%.0f × %.0f pt", page.Width, page.Height))
	}

	if info.Metadata != nil && hasMetadata(info.Metadata) {
		log.Section("📝 Metadata")
		if info.Metadata.Title != "" {
			log.Table("Title", info.Metadata.Title)
		}
		if info.Metadata.Author != "" {
			log.Table("Author", info.Metadata.Author)
		}
		if info.Metadata.Subject != "" {
			log.Table("Subject", info.Metadata.Subject)
		}
		if info.Metadata.Keywords != "" {
			log.Table("Keywords", info.Metadata.Keywords)
		}
		if info.Metadata.Creator != "" {
			log.Table("Creator", info.Metadata.Creator)
		}
		if info.Metadata.Producer != "" {
			log.Table("Producer", info.Metadata.Producer)
		}
		if info.Metadata.CreationDate != "" {
			log.Table("Created", info.Metadata.CreationDate)
		}
		if info.Metadata.ModDate != "" {
			log.Table("Modified", info.Metadata.ModDate)
		}
	}

	log.Println()
	return nil
}

func hasMetadata(m *Metadata) bool {
	return m.Title != "" || m.Author != "" || m.Subject != "" ||
		m.Keywords != "" || m.Creator != "" || m.Producer != "" ||
		m.CreationDate != "" || m.ModDate != ""
}
