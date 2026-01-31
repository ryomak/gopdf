// Example: 05_png_images
// This example demonstrates how to embed PNG images (with and without transparency) in a PDF.
package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"

	"github.com/ryomak/gopdf"
)

// createSolidPNG creates a solid color PNG
func createSolidPNG(width, height int, col color.Color) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, col)
		}
	}

	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}

// createTransparentPNG creates a PNG with alpha gradient
func createTransparentPNG(width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Alpha gradient from left (transparent) to right (opaque)
			alpha := uint8(float64(x) / float64(width) * 255)
			img.SetRGBA(x, y, color.RGBA{R: 255, G: 100, B: 100, A: alpha})
		}
	}

	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}

// createGrayscalePNG creates a grayscale PNG
func createGrayscalePNG(width, height int) []byte {
	img := image.NewGray(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			g := uint8((x + y) * 255 / (width + height))
			img.SetGray(x, y, color.Gray{Y: g})
		}
	}

	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}

// createCirclePNG creates a PNG with a circle (demonstrating transparency)
func createCirclePNG(size int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	// Transparent background
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 0, G: 0, B: 0, A: 0})
		}
	}

	// Draw circle
	centerX, centerY := size/2, size/2
	radius := size / 2
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := x - centerX
			dy := y - centerY
			if dx*dx+dy*dy <= radius*radius {
				img.SetRGBA(x, y, color.RGBA{R: 100, G: 150, B: 255, A: 200})
			}
		}
	}

	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}

func main() {
	// 新規PDFドキュメントを作成
	doc := gopdf.New()

	// A4サイズの縦向きページを追加
	page := doc.AddPage(gopdf.PageSizeA4, gopdf.Portrait)

	// === タイトル ===
	page.SetFont(gopdf.FontHelveticaBold, 24)
	page.DrawText("PNG Image Embedding Demo", 50, 800)

	// === 説明 ===
	page.SetFont(gopdf.FontHelvetica, 12)
	page.DrawText("This PDF demonstrates PNG image embedding with transparency support.", 50, 775)

	// === 1. Solid Color PNG ===
	page.SetFont(gopdf.FontHelveticaBold, 14)
	page.DrawText("1. Solid Color PNG (No Alpha)", 50, 745)

	// Red PNG
	redPNG := createSolidPNG(100, 100, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	redImg, _ := gopdf.LoadPNG(bytes.NewReader(redPNG))
	page.DrawImage(redImg, 50, 640, 80, 80)
	page.SetFont(gopdf.FontHelvetica, 10)
	page.DrawText("Red", 70, 625)

	// Green PNG
	greenPNG := createSolidPNG(100, 100, color.RGBA{R: 0, G: 255, B: 0, A: 255})
	greenImg, _ := gopdf.LoadPNG(bytes.NewReader(greenPNG))
	page.DrawImage(greenImg, 150, 640, 80, 80)
	page.SetFont(gopdf.FontHelvetica, 10)
	page.DrawText("Green", 170, 625)

	// Blue PNG
	bluePNG := createSolidPNG(100, 100, color.RGBA{R: 0, G: 0, B: 255, A: 255})
	blueImg, _ := gopdf.LoadPNG(bytes.NewReader(bluePNG))
	page.DrawImage(blueImg, 250, 640, 80, 80)
	page.SetFont(gopdf.FontHelvetica, 10)
	page.DrawText("Blue", 275, 625)

	// === 2. Transparent PNG ===
	page.SetFont(gopdf.FontHelveticaBold, 14)
	page.DrawText("2. PNG with Alpha Gradient", 50, 595)

	// Create background rectangle to show transparency
	page.SetFillColor(gopdf.Color{R: 0.9, G: 0.9, B: 0.9})
	page.FillRectangle(50, 490, 200, 100)

	transparentPNG := createTransparentPNG(200, 100)
	transparentImg, _ := gopdf.LoadPNG(bytes.NewReader(transparentPNG))
	page.DrawImage(transparentImg, 50, 490, 200, 100)
	page.SetFont(gopdf.FontHelvetica, 10)
	page.DrawText("Alpha: 0% (left) -> 100% (right)", 50, 475)

	// === 3. Grayscale PNG ===
	page.SetFont(gopdf.FontHelveticaBold, 14)
	page.DrawText("3. Grayscale PNG", 50, 445)

	grayPNG := createGrayscalePNG(150, 100)
	grayImg, _ := gopdf.LoadPNG(bytes.NewReader(grayPNG))
	page.DrawImage(grayImg, 50, 340, 150, 100)
	page.SetFont(gopdf.FontHelvetica, 10)
	page.DrawText("Grayscale gradient", 50, 325)

	// === 4. Circle with Transparency ===
	page.SetFont(gopdf.FontHelveticaBold, 14)
	page.DrawText("4. Circle with Transparency", 50, 295)

	// Background to show transparency
	page.SetFillColor(gopdf.Color{R: 1.0, G: 1.0, B: 0.8})
	page.FillRectangle(50, 190, 120, 120)
	page.SetFillColor(gopdf.Color{R: 0.8, G: 1.0, B: 1.0})
	page.FillRectangle(110, 190, 120, 120)

	circlePNG := createCirclePNG(120)
	circleImg, _ := gopdf.LoadPNG(bytes.NewReader(circlePNG))
	page.DrawImage(circleImg, 80, 220, 120, 120)
	page.SetFont(gopdf.FontHelvetica, 10)
	page.DrawText("Semi-transparent circle", 90, 175)

	// === 5. Overlapping Transparent PNGs ===
	page.SetFont(gopdf.FontHelveticaBold, 14)
	page.DrawText("5. Overlapping Transparent Images", 300, 595)

	// Three overlapping circles
	circle1PNG := createCirclePNG(100)
	circle1Img, _ := gopdf.LoadPNG(bytes.NewReader(circle1PNG))

	circle2PNG := createSolidPNG(100, 100, color.RGBA{R: 255, G: 200, B: 0, A: 150})
	circle2Img, _ := gopdf.LoadPNG(bytes.NewReader(circle2PNG))

	circle3PNG := createSolidPNG(100, 100, color.RGBA{R: 0, G: 200, B: 150, A: 150})
	circle3Img, _ := gopdf.LoadPNG(bytes.NewReader(circle3PNG))

	page.DrawImage(circle1Img, 320, 490, 90, 90)
	page.DrawImage(circle2Img, 360, 520, 90, 90)
	page.DrawImage(circle3Img, 340, 540, 90, 90)
	page.SetFont(gopdf.FontHelvetica, 10)
	page.DrawText("Overlapping with alpha", 330, 470)

	// === フッター ===
	page.SetFont(gopdf.FontHelveticaOblique, 10)
	page.DrawText("Generated with gopdf - https://github.com/ryomak/gopdf", 50, 30)

	// ファイルに出力
	file, err := os.Create("output.pdf")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	if err := doc.WriteTo(file); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing PDF: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("PDF created successfully: output.pdf")
	fmt.Println("Open the file to see embedded PNG images with transparency!")
	fmt.Println()
	fmt.Println("PNG features demonstrated:")
	fmt.Println("  - Solid color PNG images")
	fmt.Println("  - PNG with alpha gradient")
	fmt.Println("  - Grayscale PNG")
	fmt.Println("  - PNG with transparency (circle)")
	fmt.Println("  - Overlapping transparent images")
}
