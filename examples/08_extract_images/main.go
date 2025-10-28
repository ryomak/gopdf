// Example: 08_extract_images
// This example demonstrates how to extract images from a PDF file.
package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"

	"github.com/ryomak/gopdf"
	"github.com/ryomak/gopdf/internal/font"
)

func main() {
	// まず、画像入りPDFを生成
	fmt.Println("Creating sample PDF with images...")
	createSamplePDF("sample.pdf")

	// PDFファイルを読み込む
	fmt.Println("\nReading PDF file...")
	reader, err := gopdf.Open("sample.pdf")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening PDF: %v\n", err)
		os.Exit(1)
	}
	defer reader.Close()

	// 全ページから画像を抽出
	fmt.Println("\nExtracting images from all pages:")
	allImages, err := reader.ExtractAllImages()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error extracting images: %v\n", err)
		os.Exit(1)
	}

	totalImages := 0
	for pageNum, images := range allImages {
		fmt.Printf("\nPage %d: %d image(s)\n", pageNum+1, len(images))
		totalImages += len(images)

		for i, img := range images {
			fmt.Printf("  Image %d:\n", i+1)
			fmt.Printf("    Name: %s\n", img.Name)
			fmt.Printf("    Size: %dx%d\n", img.Width, img.Height)
			fmt.Printf("    ColorSpace: %s\n", img.ColorSpace)
			fmt.Printf("    Format: %s\n", img.Format)
			fmt.Printf("    Filter: %s\n", img.Filter)
			fmt.Printf("    Data size: %d bytes\n", len(img.Data))

			// 画像をバイトデータとして保存
			filename := fmt.Sprintf("extracted_page%d_img%d.%s", pageNum+1, i+1, img.Format)
			if err := img.SaveImage(filename); err != nil {
				fmt.Fprintf(os.Stderr, "    Error saving image: %v\n", err)
				continue
			}
			fmt.Printf("    Saved (raw): %s\n", filename)

			// image.Image型に変換して保存
			stdImg, err := img.ToImage()
			if err != nil {
				fmt.Fprintf(os.Stderr, "    Error converting to image.Image: %v\n", err)
				continue
			}

			// image.Image型として処理（例: PNGとして再保存）
			pngFilename := fmt.Sprintf("extracted_page%d_img%d_converted.png", pageNum+1, i+1)
			if err := saveAsPNG(stdImg, pngFilename); err != nil {
				fmt.Fprintf(os.Stderr, "    Error saving as PNG: %v\n", err)
				continue
			}
			fmt.Printf("    Saved (as PNG): %s\n", pngFilename)
			fmt.Printf("    Image bounds: %v\n", stdImg.Bounds())
		}
	}

	fmt.Printf("\nTotal images extracted: %d\n", totalImages)
	fmt.Println("\nImage extraction completed successfully!")
}

// saveAsPNG はimage.Image型の画像をPNGファイルとして保存する
func saveAsPNG(img image.Image, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		return fmt.Errorf("failed to encode PNG: %w", err)
	}

	return nil
}

// createSamplePDF は画像入りPDFを作成する
func createSamplePDF(filename string) {
	doc := gopdf.New()

	// ページ1: JPEG画像
	page1 := doc.AddPage(gopdf.A4, gopdf.Portrait)
	page1.SetFont(font.HelveticaBold, 18)
	page1.DrawText("Image Extraction Example", 50, 800)

	page1.SetFont(font.Helvetica, 12)
	page1.DrawText("This page contains a JPEG image:", 50, 770)

	// JPEG画像を追加（テスト用に簡単な画像を生成）
	// Note: 実際のアプリケーションでは既存の画像ファイルを使用します
	jpegData := createTestJPEGData()
	jpegImage, err := gopdf.LoadJPEG(bytes.NewReader(jpegData))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading JPEG: %v\n", err)
		os.Exit(1)
	}

	page1.DrawImage(jpegImage, 50, 500, 200, 150)

	page1.SetFont(font.Courier, 10)
	page1.DrawText("JPEG image (200x150)", 50, 470)

	// ページ2: 複数の画像
	page2 := doc.AddPage(gopdf.A4, gopdf.Portrait)
	page2.SetFont(font.HelveticaBold, 18)
	page2.DrawText("Page 2 - Multiple Images", 50, 800)

	// 同じ画像を異なる位置に配置
	page2.DrawImage(jpegImage, 50, 600, 150, 100)
	page2.DrawImage(jpegImage, 250, 600, 150, 100)

	// ファイルに出力
	file, err := os.Create(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	if err := doc.WriteTo(file); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing PDF: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("  Created: %s\n", filename)
}

// createTestJPEGData はテスト用のJPEG画像データを生成する
func createTestJPEGData() []byte {
	// 100x100のグラデーション画像を作成
	width, height := 100, 100
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := color.RGBA{
				R: uint8(x * 255 / width),
				G: uint8(y * 255 / height),
				B: 128,
				A: 255,
			}
			img.Set(x, y, c)
		}
	}

	// JPEG形式でエンコード
	var buf bytes.Buffer
	jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
	return buf.Bytes()
}
