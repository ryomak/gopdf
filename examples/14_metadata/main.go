package main

import (
	"fmt"
	"os"
	"time"

	"github.com/ryomak/gopdf"
)

func main() {
	// Create a new PDF document
	doc := gopdf.New()

	// Add a page
	page := doc.AddPage(gopdf.A4, gopdf.Portrait)

	// Set font and draw text
	page.SetFont(gopdf.HelveticaBold, 24)
	page.DrawText("PDF Metadata Example", 100, 750)

	page.SetFont(gopdf.Helvetica, 12)
	page.DrawText("This PDF has custom metadata.", 100, 700)
	page.DrawText("You can view it in your PDF viewer's properties.", 100, 680)

	// Set metadata (including custom fields)
	metadata := gopdf.Metadata{
		Title:        "PDF Metadata Example",
		Author:       "gopdf Library",
		Subject:      "Demonstration of PDF Metadata",
		Keywords:     "PDF, metadata, gopdf, example",
		Creator:      "gopdf Example Program",
		Producer:     "gopdf v1.0",
		CreationDate: time.Date(2025, 1, 29, 12, 0, 0, 0, time.UTC),
		ModDate:      time.Now(),
		Custom: map[string]string{
			"Department": "Engineering",
			"Project":    "gopdf",
		},
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

	fmt.Println("=== PDF created successfully: metadata_example.pdf ===")

	// Now read the metadata back
	fmt.Println("=== Reading metadata back ===")
	reader, err := gopdf.Open("metadata_example.pdf")
	if err != nil {
		panic(err)
	}
	defer reader.Close()

	readMetadata := reader.Info()

	fmt.Println("\nStandard fields:")
	fmt.Printf("  Title:        %s\n", readMetadata.Title)
	fmt.Printf("  Author:       %s\n", readMetadata.Author)
	fmt.Printf("  Subject:      %s\n", readMetadata.Subject)
	fmt.Printf("  Keywords:     %s\n", readMetadata.Keywords)
	fmt.Printf("  Creator:      %s\n", readMetadata.Creator)
	fmt.Printf("  Producer:     %s\n", readMetadata.Producer)
	fmt.Printf("  CreationDate: %s\n", readMetadata.CreationDate.Format(time.RFC3339))
	fmt.Printf("  ModDate:      %s\n", readMetadata.ModDate.Format(time.RFC3339))

	if len(readMetadata.Custom) > 0 {
		fmt.Println("\nCustom fields:")
		for key, value := range readMetadata.Custom {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}
}
