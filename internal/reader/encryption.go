package reader

import (
	"fmt"

	"github.com/ryomak/gopdf/internal/core"
	"github.com/ryomak/gopdf/internal/security"
)

// EncryptionInfo holds decryption information for reading encrypted PDFs
type EncryptionInfo struct {
	Filter         string   // Should be "Standard"
	V              int      // Version (1 or 2)
	R              int      // Revision (2 or 3)
	O              []byte   // Owner password string
	U              []byte   // User password string
	P              int32    // Permission flags
	Length         int      // Key length in bits (40 or 128)
	FileID         []byte   // File ID from trailer
	EncryptionKey  []byte   // Computed encryption key
	KeyLengthBytes int      // Key length in bytes
	Authenticated  bool     // Whether password was successfully authenticated
	IsOwner        bool     // Whether authenticated as owner
}

// parseEncryptDict parses the Encrypt dictionary from the PDF
func parseEncryptDict(encryptDict core.Dictionary, fileID []byte) (*EncryptionInfo, error) {
	info := &EncryptionInfo{
		FileID: fileID,
	}

	// Filter (required)
	if filter, ok := encryptDict[core.Name("Filter")].(core.Name); ok {
		info.Filter = string(filter)
		if info.Filter != "Standard" {
			return nil, fmt.Errorf("unsupported encryption filter: %s", info.Filter)
		}
	} else {
		return nil, fmt.Errorf("missing Filter in Encrypt dictionary")
	}

	// V (required) - algorithm version
	if v, ok := encryptDict[core.Name("V")].(core.Integer); ok {
		info.V = int(v)
	} else {
		return nil, fmt.Errorf("missing V in Encrypt dictionary")
	}

	// R (required) - revision
	if r, ok := encryptDict[core.Name("R")].(core.Integer); ok {
		info.R = int(r)
	} else {
		return nil, fmt.Errorf("missing R in Encrypt dictionary")
	}

	// O (required) - owner password string
	if o, ok := encryptDict[core.Name("O")].(core.String); ok {
		info.O = []byte(o)
	} else {
		return nil, fmt.Errorf("missing O in Encrypt dictionary")
	}

	// U (required) - user password string
	if u, ok := encryptDict[core.Name("U")].(core.String); ok {
		info.U = []byte(u)
	} else {
		return nil, fmt.Errorf("missing U in Encrypt dictionary")
	}

	// P (required) - permissions
	if p, ok := encryptDict[core.Name("P")].(core.Integer); ok {
		info.P = int32(p)
	} else {
		return nil, fmt.Errorf("missing P in Encrypt dictionary")
	}

	// Length (optional, default is 40)
	if length, ok := encryptDict[core.Name("Length")].(core.Integer); ok {
		info.Length = int(length)
	} else {
		// Default length for V=1 is 40 bits
		info.Length = 40
	}

	info.KeyLengthBytes = info.Length / 8

	return info, nil
}

// Authenticate attempts to authenticate with the given password
func (ei *EncryptionInfo) Authenticate(password string) error {
	// Try as user password first
	if security.AuthenticateUserPassword(
		password,
		ei.U,
		ei.O,
		ei.P,
		ei.FileID,
		ei.R,
		ei.KeyLengthBytes,
	) {
		// Compute encryption key
		ei.EncryptionKey = security.ComputeEncryptionKey(
			password,
			ei.O,
			ei.P,
			ei.FileID,
			ei.R,
			ei.KeyLengthBytes,
		)
		ei.Authenticated = true
		ei.IsOwner = false
		return nil
	}

	// Try as owner password
	userPassword, ok := security.AuthenticateOwnerPassword(
		password,
		ei.O,
		ei.R,
		ei.KeyLengthBytes,
	)
	if ok {
		// Authenticate with recovered user password
		if security.AuthenticateUserPassword(
			userPassword,
			ei.U,
			ei.O,
			ei.P,
			ei.FileID,
			ei.R,
			ei.KeyLengthBytes,
		) {
			// Compute encryption key using user password
			ei.EncryptionKey = security.ComputeEncryptionKey(
				userPassword,
				ei.O,
				ei.P,
				ei.FileID,
				ei.R,
				ei.KeyLengthBytes,
			)
			ei.Authenticated = true
			ei.IsOwner = true
			return nil
		}
	}

	return fmt.Errorf("password authentication failed")
}

// DecryptStream decrypts a stream object
func (ei *EncryptionInfo) DecryptStream(data []byte, objectNumber, generationNumber int) []byte {
	if !ei.Authenticated {
		return data // Return as-is if not authenticated
	}

	return security.DecryptStream(data, ei.EncryptionKey, objectNumber, generationNumber, ei.KeyLengthBytes)
}

// DecryptString decrypts a string object
func (ei *EncryptionInfo) DecryptString(data []byte, objectNumber, generationNumber int) string {
	if !ei.Authenticated {
		return string(data) // Return as-is if not authenticated
	}

	return security.DecryptString(data, ei.EncryptionKey, objectNumber, generationNumber, ei.KeyLengthBytes)
}
