package gopdf

import (
	"fmt"
	"os"
)

// SaveImage は画像をファイルに保存する
func (img *ImageInfo) SaveImage(filename string) error {
	// ファイルを作成
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// データを書き込み
	_, err = file.Write(img.Data)
	if err != nil {
		return fmt.Errorf("failed to write image data: %w", err)
	}

	return nil
}
