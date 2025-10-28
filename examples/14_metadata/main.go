package main

import (
	"fmt"
	"os"
	"time"

	"github.com/ryomak/gopdf"
	"github.com/ryomak/gopdf/internal/font"
)

func main() {
	// Create a new PDF document
	doc := gopdf.New()

	// Add a page
	page := doc.AddPage(gopdf.A4, gopdf.Portrait)

	// Set font and draw text
	page.SetFont(font.HelveticaBold, 24)
	page.DrawText("PDF Metadata Example", 100, 750)

	page.SetFont(font.Helvetica, 12)
	page.DrawText("This PDF has custom metadata.", 100, 700)
	page.DrawText("You can view it in your PDF viewer's properties.", 100, 680)

	// Set metadata
	metadata := gopdf.Metadata{
		Title:        "PDF Metadata Example",
		Author:       "gopdf Library",
		Subject:      "Demonstration of PDF Metadata",
		Keywords:     "PDF, metadata, gopdf, example",
		Creator:      "gopdf Example Program",
		Producer:     "gopdf v1.0",
		CreationDate: time.Date(2025, 1, 29, 12, 0, 0, 0, time.UTC),
		ModDate:      time.Now(),
	}
	doc.SetMetadata(metadata)

	// Write to file
	file, err := os.Create("metadata_example.pdf")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	if err := doc.WriteTo(file); err != nil {
		panic(err)
	}

	fmt.Println("PDF created successfully: metadata_example.pdf")
	fmt.Println("Metadata:")
	fmt.Printf("  Title: %s\n", metadata.Title)
	fmt.Printf("  Author: %s\n", metadata.Author)
	fmt.Printf("  Subject: %s\n", metadata.Subject)
	fmt.Printf("  Keywords: %s\n", metadata.Keywords)
}
