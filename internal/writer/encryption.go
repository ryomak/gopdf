package writer

import (
	"crypto/rand"
	"fmt"

	"github.com/ryomak/gopdf/internal/core"
	"github.com/ryomak/gopdf/internal/security"
)

// EncryptionInfo holds encryption-related information for PDF generation
type EncryptionInfo struct {
	UserPassword  string
	OwnerPassword string
	Permissions   security.Permissions
	KeyLength     int // 40 or 128 bits
	FileID        []byte
	EncryptionKey []byte
	OValue        []byte // Owner password string
	UValue        []byte // User password string
}

// GenerateFileID generates a random 16-byte file ID
func GenerateFileID() ([]byte, error) {
	fileID := make([]byte, 16)
	_, err := rand.Read(fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate file ID: %w", err)
	}
	return fileID, nil
}

// SetupEncryption initializes encryption parameters and computes O, U values
func SetupEncryption(userPassword, ownerPassword string, permissions security.Permissions, keyLength int) (*EncryptionInfo, error) {
	// Validate key length
	if keyLength != 40 && keyLength != 128 {
		return nil, fmt.Errorf("key length must be 40 or 128 bits, got %d", keyLength)
	}

	// Generate file ID
	fileID, err := GenerateFileID()
	if err != nil {
		return nil, err
	}

	// Determine revision based on key length
	revision := 2
	if keyLength == 128 {
		revision = 3
	}
	keyLengthBytes := keyLength / 8

	// Compute O value (owner password string)
	oValue := security.ComputeOwnerPassword(ownerPassword, userPassword, revision, keyLengthBytes)

	// Compute encryption key
	permInt := permissions.ToInt32()
	encryptionKey := security.ComputeEncryptionKey(userPassword, oValue, permInt, fileID, revision, keyLengthBytes)

	// Compute U value (user password string)
	uValue := security.ComputeUserPassword(encryptionKey, fileID, revision)

	return &EncryptionInfo{
		UserPassword:  userPassword,
		OwnerPassword: ownerPassword,
		Permissions:   permissions,
		KeyLength:     keyLength,
		FileID:        fileID,
		EncryptionKey: encryptionKey,
		OValue:        oValue,
		UValue:        uValue,
	}, nil
}

// CreateEncryptDictionary creates the Encrypt dictionary for the PDF
func (ei *EncryptionInfo) CreateEncryptDictionary() core.Dictionary {
	// Determine V and R based on key length
	v := 1
	r := 2
	if ei.KeyLength == 128 {
		v = 2
		r = 3
	}

	encryptDict := core.Dictionary{
		core.Name("Filter"): core.Name("Standard"),
		core.Name("V"):      core.Integer(v),
		core.Name("R"):      core.Integer(r),
		core.Name("O"):      core.String(ei.OValue),
		core.Name("U"):      core.String(ei.UValue),
		core.Name("P"):      core.Integer(ei.Permissions.ToInt32()),
	}

	// Length is only needed for V=2 or greater
	if v >= 2 {
		encryptDict[core.Name("Length")] = core.Integer(ei.KeyLength)
	}

	return encryptDict
}

// CreateFileIDArray creates the file ID array for the trailer
func (ei *EncryptionInfo) CreateFileIDArray() core.Array {
	// File ID array consists of two identical strings in a simple implementation
	return core.Array{
		core.String(ei.FileID),
		core.String(ei.FileID),
	}
}
