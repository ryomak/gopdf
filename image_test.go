package gopdf

import (
	"bytes"
	"os"
	"testing"
)

// createMinimalJPEG creates a minimal valid JPEG for testing
func createMinimalJPEG(width, height, components int) []byte {
	buf := []byte{
		0xFF, 0xD8, // SOI
		0xFF, 0xC0, // SOF0
	}

	// Calculate SOF length
	sofLength := 8 + (components * 3)
	buf = append(buf, byte(sofLength>>8), byte(sofLength&0xFF))

	// SOF data
	buf = append(buf,
		0x08,                          // Bits per component
		byte(height>>8), byte(height), // Height
		byte(width>>8), byte(width), // Width
		byte(components), // Number of components
	)

	// Component info (simplified)
	for i := 0; i < components; i++ {
		buf = append(buf, byte(i+1), 0x11, 0x00)
	}

	// Add minimal scan data
	buf = append(buf,
		0xFF, 0xDA, // SOS
		0x00, 0x0C, // Length
		0x03, // Number of components in scan
		0x01, 0x00, // Component 1
		0x02, 0x11, // Component 2
		0x03, 0x11, // Component 3
		0x00, 0x3F, 0x00, // Spectral selection
	)

	// Add fake compressed data
	buf = append(buf, make([]byte, 100)...)

	// EOI
	buf = append(buf, 0xFF, 0xD9)

	return buf
}

// TestLoadJPEG はLoadJPEG関数をテストする
func TestLoadJPEG(t *testing.T) {
	tests := []struct {
		name           string
		jpegData       []byte
		expectedWidth  int
		expectedHeight int
		expectedColor  string
		expectError    bool
	}{
		{
			name:           "RGB JPEG 640x480",
			jpegData:       createMinimalJPEG(640, 480, 3),
			expectedWidth:  640,
			expectedHeight: 480,
			expectedColor:  "DeviceRGB",
			expectError:    false,
		},
		{
			name:           "Grayscale JPEG 100x100",
			jpegData:       createMinimalJPEG(100, 100, 1),
			expectedWidth:  100,
			expectedHeight: 100,
			expectedColor:  "DeviceGray",
			expectError:    false,
		},
		{
			name:        "Invalid JPEG",
			jpegData:    []byte{0x00, 0x01, 0x02},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader(tt.jpegData)
			img, err := LoadJPEG(reader)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if img.Width != tt.expectedWidth {
				t.Errorf("Width = %d, want %d", img.Width, tt.expectedWidth)
			}
			if img.Height != tt.expectedHeight {
				t.Errorf("Height = %d, want %d", img.Height, tt.expectedHeight)
			}
			if img.ColorSpace != tt.expectedColor {
				t.Errorf("ColorSpace = %s, want %s", img.ColorSpace, tt.expectedColor)
			}
			if img.BitsPerComponent != 8 {
				t.Errorf("BitsPerComponent = %d, want 8", img.BitsPerComponent)
			}
			if len(img.Data) == 0 {
				t.Error("Image data is empty")
			}
		})
	}
}

// TestLoadJPEGFile はLoadJPEGFile関数をテストする
func TestLoadJPEGFile(t *testing.T) {
	// Create a temporary JPEG file
	tmpfile, err := os.CreateTemp("", "test*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// Write minimal JPEG data
	jpegData := createMinimalJPEG(320, 240, 3)
	if _, err := tmpfile.Write(jpegData); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Test loading the file
	img, err := LoadJPEGFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to load JPEG file: %v", err)
	}

	if img.Width != 320 {
		t.Errorf("Width = %d, want 320", img.Width)
	}
	if img.Height != 240 {
		t.Errorf("Height = %d, want 240", img.Height)
	}
}

// TestLoadJPEGFile_NotFound はファイルが存在しない場合のエラーをテストする
func TestLoadJPEGFile_NotFound(t *testing.T) {
	_, err := LoadJPEGFile("/nonexistent/file.jpg")
	if err == nil {
		t.Error("Expected error for nonexistent file, but got none")
	}
}

// TestPageDrawImage はDrawImageメソッドをテストする
func TestPageDrawImage(t *testing.T) {
	doc := New()
	page := doc.AddPage(A4, Portrait)

	// Create a minimal JPEG image
	jpegData := createMinimalJPEG(100, 100, 3)
	img, err := LoadJPEG(bytes.NewReader(jpegData))
	if err != nil {
		t.Fatalf("Failed to load JPEG: %v", err)
	}

	// Draw image at (50, 500) with size 200x150
	err = page.DrawImage(img, 50, 500, 200, 150)
	if err != nil {
		t.Fatalf("Failed to draw image: %v", err)
	}

	// Check content stream contains correct operators
	content := page.content.String()

	// Should contain: q (save state), cm (transform matrix), Do (draw), Q (restore state)
	if !containsSubstring(content, "q\n") {
		t.Error("Content should contain 'q' (save graphics state)")
	}
	if !containsSubstring(content, " cm\n") {
		t.Error("Content should contain 'cm' (transform matrix)")
	}
	if !containsSubstring(content, " Do\n") {
		t.Error("Content should contain 'Do' (draw XObject)")
	}
	if !containsSubstring(content, "Q\n") {
		t.Error("Content should contain 'Q' (restore graphics state)")
	}

	// Check for transform matrix with correct values
	// Format: width 0 0 height x y cm
	expectedMatrix := "200.00 0.00 0.00 150.00 50.00 500.00 cm"
	if !containsSubstring(content, expectedMatrix) {
		t.Errorf("Content should contain transform matrix: %s", expectedMatrix)
	}
}

// TestDocumentWithImage はDocument全体で画像が正しく処理されることをテストする
func TestDocumentWithImage(t *testing.T) {
	doc := New()
	page := doc.AddPage(A4, Portrait)

	// Create and load a test image
	jpegData := createMinimalJPEG(320, 240, 3)
	img, err := LoadJPEG(bytes.NewReader(jpegData))
	if err != nil {
		t.Fatalf("Failed to load JPEG: %v", err)
	}

	// Draw the image
	err = page.DrawImage(img, 100, 600, 300, 225)
	if err != nil {
		t.Fatalf("Failed to draw image: %v", err)
	}

	// Write to buffer to ensure no errors
	var buf bytes.Buffer
	err = doc.WriteTo(&buf)
	if err != nil {
		t.Fatalf("Failed to write PDF: %v", err)
	}

	// Check that PDF contains expected markers
	pdfContent := buf.String()
	if !containsSubstring(pdfContent, "%PDF-1.7") {
		t.Error("PDF should contain version header")
	}
	if !containsSubstring(pdfContent, "/XObject") {
		t.Error("PDF should contain XObject reference for image")
	}
	if !containsSubstring(pdfContent, "/Image") {
		t.Error("PDF should contain Image subtype")
	}
	if !containsSubstring(pdfContent, "/DCTDecode") {
		t.Error("PDF should contain DCTDecode filter for JPEG")
	}
}

// TestMultipleImages は複数の画像を同じページに配置できることをテストする
func TestMultipleImages(t *testing.T) {
	doc := New()
	page := doc.AddPage(A4, Portrait)

	// Create two different images
	img1, _ := LoadJPEG(bytes.NewReader(createMinimalJPEG(100, 100, 3)))
	img2, _ := LoadJPEG(bytes.NewReader(createMinimalJPEG(200, 150, 1)))

	// Draw both images
	_ = page.DrawImage(img1, 50, 700, 100, 100)
	_ = page.DrawImage(img2, 200, 700, 200, 150)

	// Check that content contains both Do operations
	content := page.content.String()
	// Count occurrences of "Do" operator (should be at least 2)
	count := 0
	for i := 0; i < len(content); i++ {
		if i+3 <= len(content) && content[i:i+3] == "Do\n" {
			count++
		}
	}
	if count < 2 {
		t.Errorf("Expected at least 2 'Do' operators, found %d", count)
	}
}
