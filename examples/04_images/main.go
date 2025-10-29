// Example: 04_images
// This example demonstrates how to embed images (JPEG) in a PDF.
package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"os"

	"github.com/ryomak/gopdf"
)

// createSampleImage creates a simple colored rectangle image
func createSampleImage(width, height int, col color.Color) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, col)
		}
	}
	return img
}

// createGradientImage creates a gradient image
func createGradientImage(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r := uint8(float64(x) / float64(width) * 255)
			g := uint8(float64(y) / float64(height) * 255)
			b := uint8(128)
			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}
	return img
}

// encodeJPEG encodes an image to JPEG format
func encodeJPEG(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func main() {
	// 新規PDFドキュメントを作成
	doc := gopdf.New()

	// A4サイズの縦向きページを追加
	page := doc.AddPage(gopdf.PageSizeA4, gopdf.Portrait)

	// === タイトル ===
	page.SetFont(gopdf.FontHelveticaBold, 24)
	page.DrawText("Image Embedding Demo", 50, 800)

	// === 説明 ===
	page.SetFont(gopdf.FontHelvetica, 12)
	page.DrawText("This PDF demonstrates JPEG image embedding capabilities.", 50, 775)

	// === 色付き矩形画像 ===
	page.SetFont(gopdf.FontHelveticaBold, 14)
	page.DrawText("1. Colored Rectangles", 50, 745)

	// 赤い矩形
	redImg := createSampleImage(200, 100, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	redJPEG, _ := encodeJPEG(redImg)
	redImage, err := gopdf.LoadJPEG(bytes.NewReader(redJPEG))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading red image: %v\n", err)
		os.Exit(1)
	}
	page.DrawImage(redImage, 50, 630, 150, 75)
	page.SetFont(gopdf.FontHelvetica, 10)
	page.DrawText("Red", 105, 615)

	// 緑の矩形
	greenImg := createSampleImage(200, 100, color.RGBA{R: 0, G: 255, B: 0, A: 255})
	greenJPEG, _ := encodeJPEG(greenImg)
	greenImage, _ := gopdf.LoadJPEG(bytes.NewReader(greenJPEG))
	page.DrawImage(greenImage, 220, 630, 150, 75)
	page.SetFont(gopdf.FontHelvetica, 10)
	page.DrawText("Green", 275, 615)

	// 青い矩形
	blueImg := createSampleImage(200, 100, color.RGBA{R: 0, G: 0, B: 255, A: 255})
	blueJPEG, _ := encodeJPEG(blueImg)
	blueImage, _ := gopdf.LoadJPEG(bytes.NewReader(blueJPEG))
	page.DrawImage(blueImage, 390, 630, 150, 75)
	page.SetFont(gopdf.FontHelvetica, 10)
	page.DrawText("Blue", 450, 615)

	// === グラデーション画像 ===
	page.SetFont(gopdf.FontHelveticaBold, 14)
	page.DrawText("2. Gradient Image", 50, 580)

	gradientImg := createGradientImage(400, 200)
	gradientJPEG, _ := encodeJPEG(gradientImg)
	gradientImage, _ := gopdf.LoadJPEG(bytes.NewReader(gradientJPEG))
	page.DrawImage(gradientImage, 50, 380, 300, 150)

	page.SetFont(gopdf.FontHelvetica, 10)
	page.DrawText("RGB gradient: X-axis (red), Y-axis (green)", 50, 365)

	// === 異なるサイズの画像 ===
	page.SetFont(gopdf.FontHelveticaBold, 14)
	page.DrawText("3. Different Sizes", 50, 335)

	// 小サイズ
	smallImg := createSampleImage(100, 100, color.RGBA{R: 255, G: 165, B: 0, A: 255})
	smallJPEG, _ := encodeJPEG(smallImg)
	smallImage, _ := gopdf.LoadJPEG(bytes.NewReader(smallJPEG))
	page.DrawImage(smallImage, 50, 260, 50, 50)
	page.SetFont(gopdf.FontHelvetica, 10)
	page.DrawText("50x50", 55, 245)

	// 中サイズ
	page.DrawImage(smallImage, 120, 250, 75, 75)
	page.SetFont(gopdf.FontHelvetica, 10)
	page.DrawText("75x75", 135, 235)

	// 大サイズ
	page.DrawImage(smallImage, 215, 235, 100, 100)
	page.SetFont(gopdf.FontHelvetica, 10)
	page.DrawText("100x100", 240, 220)

	// === パターン ===
	page.SetFont(gopdf.FontHelveticaBold, 14)
	page.DrawText("4. Pattern", 50, 190)

	// チェッカーボードパターン
	checkerImg := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			if (x/10+y/10)%2 == 0 {
				checkerImg.Set(x, y, color.Black)
			} else {
				checkerImg.Set(x, y, color.White)
			}
		}
	}
	checkerJPEG, _ := encodeJPEG(checkerImg)
	checkerImage, _ := gopdf.LoadJPEG(bytes.NewReader(checkerJPEG))
	page.DrawImage(checkerImage, 50, 80, 120, 120)

	page.SetFont(gopdf.FontHelvetica, 10)
	page.DrawText("Checkerboard pattern", 50, 65)

	// === フッター ===
	page.SetFont(gopdf.FontHelveticaOblique, 10)
	page.SetStrokeColor(gopdf.ColorBlack)
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
	fmt.Println("Open the file to see embedded JPEG images!")
	fmt.Println()
	fmt.Println("Images generated:")
	fmt.Println("  - Colored rectangles (red, green, blue)")
	fmt.Println("  - Gradient image (RGB)")
	fmt.Println("  - Different sized images")
	fmt.Println("  - Checkerboard pattern")
}
