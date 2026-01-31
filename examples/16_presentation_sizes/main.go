package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ryomak/gopdf"
)

func main() {
	// Create 16:9 widescreen presentation
	if err := create16x9Presentation(); err != nil {
		log.Fatalf("Failed to create 16:9 presentation: %v", err)
	}
	fmt.Println("✓ Created 16x9_presentation.pdf")

	// Create 4:3 standard presentation
	if err := create4x3Presentation(); err != nil {
		log.Fatalf("Failed to create 4:3 presentation: %v", err)
	}
	fmt.Println("✓ Created 4x3_presentation.pdf")
}

func create16x9Presentation() error {
	doc := gopdf.New()

	// Slide 1: Title Slide
	page1 := doc.AddPage(gopdf.PageSizePresentation16x9, gopdf.Portrait)
	page1.SetFont(gopdf.FontHelvetica, 48)
	page1.DrawText("16:9 Widescreen", 50, 250)

	page1.SetFont(gopdf.FontHelvetica, 24)
	page1.DrawText("Modern Presentation Format", 50, 200)

	page1.SetFont(gopdf.FontHelvetica, 12)
	page1.DrawText("Size: 720 x 405 points (10\" x 5.625\")", 50, 150)

	// Slide 2: Content Slide
	page2 := doc.AddPage(gopdf.PageSizePresentation16x9, gopdf.Portrait)
	page2.SetFont(gopdf.FontHelvetica, 36)
	page2.DrawText("Features", 50, 350)

	page2.SetFont(gopdf.FontHelvetica, 18)
	page2.DrawText("• Perfect for modern displays", 70, 300)
	page2.DrawText("• 16:9 aspect ratio", 70, 270)
	page2.DrawText("• 720 x 405 points", 70, 240)
	page2.DrawText("• Widescreen format", 70, 210)

	// Draw a rectangle to show the slide boundary
	page2.SetStrokeColor(gopdf.Color{R: 0.8, G: 0.8, B: 0.8})
	page2.SetLineWidth(2)
	page2.DrawRectangle(10, 10, 700, 385)

	// Slide 3: Aspect Ratio Visualization
	page3 := doc.AddPage(gopdf.PageSizePresentation16x9, gopdf.Portrait)
	page3.SetFont(gopdf.FontHelvetica, 36)
	page3.DrawText("Aspect Ratio: 16:9", 50, 350)

	// Draw rectangles to visualize the ratio
	page3.SetFillColor(gopdf.Color{R: 0.2, G: 0.4, B: 0.8})
	for i := 0; i < 16; i++ {
		page3.FillRectangle(50+float64(i)*35, 200, 30, 30)
	}
	for i := 0; i < 9; i++ {
		page3.FillRectangle(50+float64(i)*35, 160, 30, 30)
	}

	page3.SetFont(gopdf.FontHelvetica, 14)
	page3.DrawText("16 units", 50, 110)
	page3.DrawText("9 units", 50, 90)

	f, err := os.Create("16x9_presentation.pdf")
	if err != nil {
		return err
	}
	defer f.Close()

	return doc.WriteTo(f)
}

func create4x3Presentation() error {
	doc := gopdf.New()

	// Slide 1: Title Slide
	page1 := doc.AddPage(gopdf.PageSizePresentation4x3, gopdf.Portrait)
	page1.SetFont(gopdf.FontHelvetica, 48)
	page1.DrawText("4:3 Standard", 50, 380)

	page1.SetFont(gopdf.FontHelvetica, 24)
	page1.DrawText("Classic Presentation Format", 50, 330)

	page1.SetFont(gopdf.FontHelvetica, 12)
	page1.DrawText("Size: 720 x 540 points (10\" x 7.5\")", 50, 280)

	// Slide 2: Content Slide
	page2 := doc.AddPage(gopdf.PageSizePresentation4x3, gopdf.Portrait)
	page2.SetFont(gopdf.FontHelvetica, 36)
	page2.DrawText("Features", 50, 480)

	page2.SetFont(gopdf.FontHelvetica, 18)
	page2.DrawText("• Classic presentation format", 70, 430)
	page2.DrawText("• 4:3 aspect ratio", 70, 400)
	page2.DrawText("• 720 x 540 points", 70, 370)
	page2.DrawText("• Compatible with older projectors", 70, 340)

	// Draw a rectangle to show the slide boundary
	page2.SetStrokeColor(gopdf.Color{R: 0.8, G: 0.8, B: 0.8})
	page2.SetLineWidth(2)
	page2.DrawRectangle(10, 10, 700, 520)

	// Slide 3: Aspect Ratio Visualization
	page3 := doc.AddPage(gopdf.PageSizePresentation4x3, gopdf.Portrait)
	page3.SetFont(gopdf.FontHelvetica, 36)
	page3.DrawText("Aspect Ratio: 4:3", 50, 480)

	// Draw rectangles to visualize the ratio
	page3.SetFillColor(gopdf.Color{R: 0.8, G: 0.4, B: 0.2})
	for i := 0; i < 4; i++ {
		page3.FillRectangle(50+float64(i)*80, 350, 70, 70)
	}
	for i := 0; i < 3; i++ {
		page3.FillRectangle(50+float64(i)*80, 270, 70, 70)
	}

	page3.SetFont(gopdf.FontHelvetica, 14)
	page3.DrawText("4 units", 50, 220)
	page3.DrawText("3 units", 50, 200)

	// Slide 4: Comparison
	page4 := doc.AddPage(gopdf.PageSizePresentation4x3, gopdf.Portrait)
	page4.SetFont(gopdf.FontHelvetica, 30)
	page4.DrawText("Format Comparison", 50, 480)

	page4.SetFont(gopdf.FontHelvetica, 14)
	page4.DrawText("16:9 Widescreen:", 50, 430)
	page4.DrawText("  • Modern displays", 70, 410)
	page4.DrawText("  • More horizontal space", 70, 390)
	page4.DrawText("  • 720 x 405 points", 70, 370)

	page4.DrawText("4:3 Standard:", 50, 330)
	page4.DrawText("  • Classic format", 70, 310)
	page4.DrawText("  • More vertical space", 70, 290)
	page4.DrawText("  • 720 x 540 points", 70, 270)

	f, err := os.Create("4x3_presentation.pdf")
	if err != nil {
		return err
	}
	defer f.Close()

	return doc.WriteTo(f)
}
