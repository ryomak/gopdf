package gopdf

import (
	"fmt"

	"github.com/ryomak/gopdf/internal/security"
)

// EncryptionOptions はPDF暗号化のオプション
type EncryptionOptions struct {
	UserPassword  string      // ユーザーパスワード（PDFを開くために必要）
	OwnerPassword string      // オーナーパスワード（すべての権限）
	Permissions   Permissions // アクセス権限
	KeyLength     int         // 暗号鍵の長さ（40 or 128 bits）
}

// Permissions はPDFのアクセス権限
type Permissions struct {
	Print            bool // 印刷を許可
	Modify           bool // 内容の変更を許可
	Copy             bool // テキスト・グラフィックのコピーを許可
	Annotate         bool // 注釈の追加・変更を許可
	FillForms        bool // フォームフィールドの入力を許可
	ExtractContent   bool // テキスト・グラフィックの抽出を許可
	Assemble         bool // ページの挿入・削除・回転を許可
	PrintHighQuality bool // 高解像度での印刷を許可
}

// DefaultPermissions はデフォルトの権限（すべて許可）を返す
func DefaultPermissions() Permissions {
	return Permissions{
		Print:            true,
		Modify:           true,
		Copy:             true,
		Annotate:         true,
		FillForms:        true,
		ExtractContent:   true,
		Assemble:         true,
		PrintHighQuality: true,
	}
}

// RestrictedPermissions は制限された権限（閲覧のみ）を返す
func RestrictedPermissions() Permissions {
	return Permissions{
		Print:            false,
		Modify:           false,
		Copy:             false,
		Annotate:         false,
		FillForms:        false,
		ExtractContent:   false,
		Assemble:         false,
		PrintHighQuality: false,
	}
}

// PrintOnlyPermissions は印刷のみ許可する権限を返す
func PrintOnlyPermissions() Permissions {
	return Permissions{
		Print:            true,
		Modify:           false,
		Copy:             false,
		Annotate:         false,
		FillForms:        false,
		ExtractContent:   false,
		Assemble:         false,
		PrintHighQuality: true,
	}
}

// toInternal converts gopdf.Permissions to security.Permissions
func (p Permissions) toInternal() security.Permissions {
	return security.Permissions{
		Print:            p.Print,
		Modify:           p.Modify,
		Copy:             p.Copy,
		Annotate:         p.Annotate,
		FillForms:        p.FillForms,
		ExtractContent:   p.ExtractContent,
		Assemble:         p.Assemble,
		PrintHighQuality: p.PrintHighQuality,
	}
}

// Validate validates the encryption options
func (opts EncryptionOptions) Validate() error {
	// At least one password must be set
	if opts.UserPassword == "" && opts.OwnerPassword == "" {
		return fmt.Errorf("at least one password must be set")
	}

	// Key length must be 40 or 128
	if opts.KeyLength != 40 && opts.KeyLength != 128 {
		return fmt.Errorf("key length must be 40 or 128 bits, got %d", opts.KeyLength)
	}

	return nil
}

// GetRevision returns the PDF encryption revision number based on key length
func (opts EncryptionOptions) GetRevision() int {
	if opts.KeyLength == 40 {
		return 2 // Revision 2 for 40-bit
	}
	return 3 // Revision 3 for 128-bit
}

// GetKeyLengthBytes returns the key length in bytes
func (opts EncryptionOptions) GetKeyLengthBytes() int {
	return opts.KeyLength / 8
}
