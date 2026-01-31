package cmd

import (
	"encoding/json"
	"fmt"
	"os"

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

	reader, err := gopdf.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open PDF: %w", err)
	}
	defer reader.Close()

	if reader.IsEncrypted() && infoPassword != "" {
		if err := reader.AuthenticateWithPassword(infoPassword); err != nil {
			return fmt.Errorf("failed to authenticate: %w", err)
		}
	}

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

	if infoJSON {
		return outputJSON(info)
	}

	return outputText(info)
}

func outputJSON(info PDFInfo) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(info)
}

func outputText(info PDFInfo) error {
	fmt.Printf("File: %s\n", info.FileName)
	fmt.Printf("Pages: %d\n", info.PageCount)
	fmt.Printf("Encrypted: %v\n", info.Encrypted)
	fmt.Println()

	fmt.Println("Page Sizes:")
	for _, page := range info.Pages {
		fmt.Printf("  Page %d: %.2f x %.2f pt\n", page.Number, page.Width, page.Height)
	}

	if info.Metadata != nil {
		fmt.Println()
		fmt.Println("Metadata:")
		if info.Metadata.Title != "" {
			fmt.Printf("  Title: %s\n", info.Metadata.Title)
		}
		if info.Metadata.Author != "" {
			fmt.Printf("  Author: %s\n", info.Metadata.Author)
		}
		if info.Metadata.Subject != "" {
			fmt.Printf("  Subject: %s\n", info.Metadata.Subject)
		}
		if info.Metadata.Keywords != "" {
			fmt.Printf("  Keywords: %s\n", info.Metadata.Keywords)
		}
		if info.Metadata.Creator != "" {
			fmt.Printf("  Creator: %s\n", info.Metadata.Creator)
		}
		if info.Metadata.Producer != "" {
			fmt.Printf("  Producer: %s\n", info.Metadata.Producer)
		}
		if info.Metadata.CreationDate != "" {
			fmt.Printf("  Created: %s\n", info.Metadata.CreationDate)
		}
		if info.Metadata.ModDate != "" {
			fmt.Printf("  Modified: %s\n", info.Metadata.ModDate)
		}
	}

	return nil
}
