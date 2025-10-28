package gopdf

import (
	"bytes"
	"image"
	image_color "image/color"
	image_jpeg "image/jpeg"
	"testing"
)

// createValidJPEG は有効なJPEGデータを生成する
func createValidJPEG(width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// 簡単なグラデーションパターンを描画
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetRGBA(x, y, image_color.RGBA{
				R: uint8(x * 255 / width),
				G: uint8(y * 255 / height),
				B: 128,
				A: 255,
			})
		}
	}

	var buf bytes.Buffer
	if err := image_jpeg.Encode(&buf, img, &image_jpeg.Options{Quality: 90}); err != nil {
		panic(err) // テスト用のヘルパー関数なのでpanicで問題ない
	}
	return buf.Bytes()
}

// TestImageInfo_ToImage_JPEG はJPEG画像のToImageメソッドをテストする
func TestImageInfo_ToImage_JPEG(t *testing.T) {
	// テスト用のJPEGデータを作成
	jpegData := createValidJPEG(100, 80)

	imgInfo := &ImageInfo{
		Name:        "TestImage",
		Width:       100,
		Height:      80,
		ColorSpace:  "DeviceRGB",
		BitsPerComp: 8,
		Filter:      "DCTDecode",
		Data:        jpegData,
		Format:      ImageFormatJPEG,
	}

	// ToImageでimage.Imageに変換
	img, err := imgInfo.ToImage()
	if err != nil {
		t.Fatalf("ToImage failed: %v", err)
	}

	if img == nil {
		t.Fatal("ToImage returned nil image")
	}

	bounds := img.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 80 {
		t.Errorf("Image size = %dx%d, want 100x80", bounds.Dx(), bounds.Dy())
	}
}

// TestImageInfo_ToImage_FlateDecode_RGB はRGB FlateDecode画像のToImageメソッドをテストする
func TestImageInfo_ToImage_FlateDecode_RGB(t *testing.T) {
	// 小さなRGB画像データを作成（2x2ピクセル）
	width, height := 2, 2
	rawData := []byte{
		255, 0, 0, // (0,0) 赤
		0, 255, 0, // (1,0) 緑
		0, 0, 255, // (0,1) 青
		255, 255, 0, // (1,1) 黄色
	}

	// Zlib圧縮
	compressedData, err := compressWithZlib(rawData)
	if err != nil {
		t.Fatalf("Failed to compress data: %v", err)
	}

	imgInfo := &ImageInfo{
		Name:        "TestRGB",
		Width:       width,
		Height:      height,
		ColorSpace:  "DeviceRGB",
		BitsPerComp: 8,
		Filter:      "FlateDecode",
		Data:        compressedData,
		Format:      ImageFormatPNG,
	}

	// ToImageでimage.Imageに変換
	img, err := imgInfo.ToImage()
	if err != nil {
		t.Fatalf("ToImage failed: %v", err)
	}

	if img == nil {
		t.Fatal("ToImage returned nil image")
	}

	bounds := img.Bounds()
	if bounds.Dx() != width || bounds.Dy() != height {
		t.Errorf("Image size = %dx%d, want %dx%d", bounds.Dx(), bounds.Dy(), width, height)
	}

	// ピクセルカラーを確認
	r, g, b, _ := img.At(0, 0).RGBA()
	if r>>8 != 255 || g>>8 != 0 || b>>8 != 0 {
		t.Errorf("Pixel (0,0) = RGB(%d,%d,%d), want RGB(255,0,0)", r>>8, g>>8, b>>8)
	}

	r, g, b, _ = img.At(1, 0).RGBA()
	if r>>8 != 0 || g>>8 != 255 || b>>8 != 0 {
		t.Errorf("Pixel (1,0) = RGB(%d,%d,%d), want RGB(0,255,0)", r>>8, g>>8, b>>8)
	}
}

// TestImageInfo_ToImage_FlateDecode_Gray はグレースケール FlateDecode画像のToImageメソッドをテストする
func TestImageInfo_ToImage_FlateDecode_Gray(t *testing.T) {
	// 小さなグレースケール画像データを作成（3x3ピクセル）
	width, height := 3, 3
	rawData := []byte{
		0, 85, 170, // 1行目
		85, 170, 255, // 2行目
		170, 255, 0, // 3行目
	}

	// Zlib圧縮
	compressedData, err := compressWithZlib(rawData)
	if err != nil {
		t.Fatalf("Failed to compress data: %v", err)
	}

	imgInfo := &ImageInfo{
		Name:        "TestGray",
		Width:       width,
		Height:      height,
		ColorSpace:  "DeviceGray",
		BitsPerComp: 8,
		Filter:      "FlateDecode",
		Data:        compressedData,
		Format:      ImageFormatPNG,
	}

	// ToImageでimage.Imageに変換
	img, err := imgInfo.ToImage()
	if err != nil {
		t.Fatalf("ToImage failed: %v", err)
	}

	if img == nil {
		t.Fatal("ToImage returned nil image")
	}

	bounds := img.Bounds()
	if bounds.Dx() != width || bounds.Dy() != height {
		t.Errorf("Image size = %dx%d, want %dx%d", bounds.Dx(), bounds.Dy(), width, height)
	}

	// グレー型にキャストして確認
	grayImg, ok := img.(*image.Gray)
	if !ok {
		t.Fatal("Expected *image.Gray")
	}

	if grayImg.GrayAt(0, 0).Y != 0 {
		t.Errorf("Pixel (0,0) gray = %d, want 0", grayImg.GrayAt(0, 0).Y)
	}

	if grayImg.GrayAt(1, 0).Y != 85 {
		t.Errorf("Pixel (1,0) gray = %d, want 85", grayImg.GrayAt(1, 0).Y)
	}

	if grayImg.GrayAt(2, 0).Y != 170 {
		t.Errorf("Pixel (2,0) gray = %d, want 170", grayImg.GrayAt(2, 0).Y)
	}
}

// TestImageInfo_ToImage_FlateDecode_CMYK はCMYK FlateDecode画像のToImageメソッドをテストする
func TestImageInfo_ToImage_FlateDecode_CMYK(t *testing.T) {
	// 小さなCMYK画像データを作成（2x2ピクセル）
	width, height := 2, 2
	rawData := []byte{
		0, 0, 0, 0, // (0,0) White (C=0, M=0, Y=0, K=0)
		255, 0, 0, 0, // (1,0) Cyan (C=255, M=0, Y=0, K=0)
		0, 255, 0, 0, // (0,1) Magenta (C=0, M=255, Y=0, K=0)
		0, 0, 255, 0, // (1,1) Yellow (C=0, M=0, Y=255, K=0)
	}

	// Zlib圧縮
	compressedData, err := compressWithZlib(rawData)
	if err != nil {
		t.Fatalf("Failed to compress data: %v", err)
	}

	imgInfo := &ImageInfo{
		Name:        "TestCMYK",
		Width:       width,
		Height:      height,
		ColorSpace:  "DeviceCMYK",
		BitsPerComp: 8,
		Filter:      "FlateDecode",
		Data:        compressedData,
		Format:      ImageFormatPNG,
	}

	// ToImageでimage.Imageに変換
	img, err := imgInfo.ToImage()
	if err != nil {
		t.Fatalf("ToImage failed: %v", err)
	}

	if img == nil {
		t.Fatal("ToImage returned nil image")
	}

	bounds := img.Bounds()
	if bounds.Dx() != width || bounds.Dy() != height {
		t.Errorf("Image size = %dx%d, want %dx%d", bounds.Dx(), bounds.Dy(), width, height)
	}

	// RGBA型にキャストして確認
	rgbaImg, ok := img.(*image.RGBA)
	if !ok {
		t.Fatal("Expected *image.RGBA for CMYK image")
	}

	// White (C=0, M=0, Y=0, K=0) should convert to RGB(255, 255, 255)
	r, g, b, a := rgbaImg.At(0, 0).RGBA()
	if r>>8 != 255 || g>>8 != 255 || b>>8 != 255 || a>>8 != 255 {
		t.Errorf("Pixel (0,0) RGBA = (%d,%d,%d,%d), want (255,255,255,255)", r>>8, g>>8, b>>8, a>>8)
	}
}

// TestImageInfo_ToImage_UnsupportedFormat は未対応フォーマットでエラーを返すことをテストする
func TestImageInfo_ToImage_UnsupportedFormat(t *testing.T) {
	imgInfo := &ImageInfo{
		Name:        "TestUnsupported",
		Width:       100,
		Height:      100,
		ColorSpace:  "DeviceRGB",
		BitsPerComp: 8,
		Filter:      "Unknown",
		Data:        []byte{},
		Format:      ImageFormatUnknown,
	}

	_, err := imgInfo.ToImage()
	if err == nil {
		t.Error("Expected error for unsupported format, but got nil")
	}
}

// TestExtractImagesWithToImage は画像抽出とToImageを組み合わせたテストする
func TestExtractImagesWithToImage(t *testing.T) {
	// 画像入りPDFを生成
	doc := New()
	page := doc.AddPage(A4, Portrait)

	// JPEG画像を追加（有効なJPEGを生成）
	jpegData := createValidJPEG(100, 100)
	jpegImage, err := LoadJPEG(bytes.NewReader(jpegData))
	if err != nil {
		t.Fatalf("Failed to load JPEG: %v", err)
	}

	if err := page.DrawImage(jpegImage, 100, 700, 200, 150); err != nil {
		t.Fatalf("Failed to draw image: %v", err)
	}

	var buf bytes.Buffer
	if err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("Failed to write PDF: %v", err)
	}

	// PDFを読み込み
	reader, err := OpenReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("Failed to open PDF: %v", err)
	}

	// 画像を抽出
	images, err := reader.ExtractImages(0)
	if err != nil {
		t.Fatalf("ExtractImages failed: %v", err)
	}

	if len(images) != 1 {
		t.Fatalf("Expected 1 image, got %d", len(images))
	}

	// ToImageでimage.Imageに変換
	img, err := images[0].ToImage()
	if err != nil {
		t.Fatalf("ToImage failed: %v", err)
	}

	if img == nil {
		t.Fatal("ToImage returned nil image")
	}

	// サイズを確認
	bounds := img.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 100 {
		t.Errorf("Image size = %dx%d, want 100x100", bounds.Dx(), bounds.Dy())
	}
}
