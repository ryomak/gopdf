package security

import (
	"crypto/md5"
)

// PDFPaddingString is the standard padding string used in PDF encryption
var PDFPaddingString = []byte{
	0x28, 0xBF, 0x4E, 0x5E, 0x4E, 0x75, 0x8A, 0x41,
	0x64, 0x00, 0x4E, 0x56, 0xFF, 0xFA, 0x01, 0x08,
	0x2E, 0x2E, 0x00, 0xB6, 0xD0, 0x68, 0x3E, 0x80,
	0x2F, 0x0C, 0xA9, 0xFE, 0x64, 0x53, 0x69, 0x7A,
}

// PadOrTruncatePassword pads or truncates a password to 32 bytes
func PadOrTruncatePassword(password string) []byte {
	padded := make([]byte, 32)
	passwordBytes := []byte(password)

	if len(passwordBytes) >= 32 {
		copy(padded, passwordBytes[:32])
	} else {
		copy(padded, passwordBytes)
		copy(padded[len(passwordBytes):], PDFPaddingString)
	}

	return padded
}

// ComputeEncryptionKey computes the encryption key from password and parameters
// revision: 2 for 40-bit, 3 for 128-bit
// keyLength: 5 for 40-bit, 16 for 128-bit (in bytes)
func ComputeEncryptionKey(
	password string,
	o []byte,
	permissions int32,
	fileID []byte,
	revision int,
	keyLength int,
) []byte {
	// Step 1: Pad or truncate password
	paddedPassword := PadOrTruncatePassword(password)

	// Step 2: Initialize MD5 hash
	hash := md5.New()
	hash.Write(paddedPassword)
	hash.Write(o)

	// Step 3: Add permissions (low-order byte first)
	perm := make([]byte, 4)
	perm[0] = byte(permissions)
	perm[1] = byte(permissions >> 8)
	perm[2] = byte(permissions >> 16)
	perm[3] = byte(permissions >> 24)
	hash.Write(perm)

	// Step 4: Add file ID
	hash.Write(fileID)

	// Step 5: (Revision 4 or greater) If metadata is not encrypted, add 4 bytes 0xFFFFFFFF
	// For now, we always encrypt metadata, so skip this

	// Step 6: Compute hash
	digest := hash.Sum(nil)

	// Step 7: (Revision 3 or greater) Do 50 iterations of MD5
	if revision >= 3 {
		for i := 0; i < 50; i++ {
			hash = md5.New()
			hash.Write(digest[:keyLength])
			digest = hash.Sum(nil)
		}
	}

	// Step 8: Set encryption key to first keyLength bytes
	key := make([]byte, keyLength)
	copy(key, digest)

	return key
}

// ComputeOwnerPassword computes the O value (owner password string)
func ComputeOwnerPassword(
	ownerPassword, userPassword string,
	revision int,
	keyLength int,
) []byte {
	// Use owner password if provided, otherwise use user password
	password := ownerPassword
	if password == "" {
		password = userPassword
	}

	// Step 1: Pad or truncate password
	paddedPassword := PadOrTruncatePassword(password)

	// Step 2: Initialize MD5 hash
	hash := md5.New()
	hash.Write(paddedPassword)
	digest := hash.Sum(nil)

	// Step 3: (Revision 3 or greater) Do 50 iterations of MD5
	if revision >= 3 {
		for i := 0; i < 50; i++ {
			hash = md5.New()
			hash.Write(digest)
			digest = hash.Sum(nil)
		}
	}

	// Step 4: Create RC4 key from first keyLength bytes
	key := make([]byte, keyLength)
	copy(key, digest)

	// Step 5: Pad or truncate user password
	paddedUserPassword := PadOrTruncatePassword(userPassword)

	// Step 6: Encrypt padded user password with RC4
	cipher := NewRC4(key)
	o := cipher.Encrypt(paddedUserPassword)

	// Step 7: (Revision 3 or greater) Do 19 iterations with different keys
	if revision >= 3 {
		for i := 1; i <= 19; i++ {
			// XOR each byte of key with i
			newKey := make([]byte, keyLength)
			for j := 0; j < keyLength; j++ {
				newKey[j] = key[j] ^ byte(i)
			}
			cipher = NewRC4(newKey)
			o = cipher.Encrypt(o)
		}
	}

	return o
}

// ComputeUserPassword computes the U value (user password string)
func ComputeUserPassword(
	encryptionKey, fileID []byte,
	revision int,
) []byte {
	if revision == 2 {
		// Revision 2: Simply encrypt padding string
		cipher := NewRC4(encryptionKey)
		return cipher.Encrypt(PDFPaddingString)
	}

	// Revision 3 or greater
	// Step 1: Create MD5 hash of padding string and file ID
	hash := md5.New()
	hash.Write(PDFPaddingString)
	hash.Write(fileID)
	digest := hash.Sum(nil)

	// Step 2: Encrypt hash with RC4
	cipher := NewRC4(encryptionKey)
	u := cipher.Encrypt(digest)

	// Step 3: Do 19 iterations with different keys
	for i := 1; i <= 19; i++ {
		// XOR each byte of key with i
		newKey := make([]byte, len(encryptionKey))
		for j := 0; j < len(encryptionKey); j++ {
			newKey[j] = encryptionKey[j] ^ byte(i)
		}
		cipher = NewRC4(newKey)
		u = cipher.Encrypt(u)
	}

	// Step 4: Pad to 32 bytes with arbitrary data (we use zeros)
	if len(u) < 32 {
		padded := make([]byte, 32)
		copy(padded, u)
		u = padded
	}

	return u
}

// AuthenticateUserPassword authenticates a user password
// Returns true if password is correct
func AuthenticateUserPassword(
	password string,
	u, o []byte,
	permissions int32,
	fileID []byte,
	revision int,
	keyLength int,
) bool {
	// Compute encryption key from provided password
	key := ComputeEncryptionKey(password, o, permissions, fileID, revision, keyLength)

	// Compute expected U value
	expectedU := ComputeUserPassword(key, fileID, revision)

	// Compare first 16 bytes (for revision 3+) or all 32 bytes (for revision 2)
	compareLength := 32
	if revision >= 3 {
		compareLength = 16
	}

	for i := 0; i < compareLength && i < len(u) && i < len(expectedU); i++ {
		if u[i] != expectedU[i] {
			return false
		}
	}

	return true
}

// AuthenticateOwnerPassword authenticates an owner password
// Returns (userPassword, true) if successful, ("", false) if failed
func AuthenticateOwnerPassword(
	password string,
	o []byte,
	revision int,
	keyLength int,
) (string, bool) {
	// Pad owner password
	paddedPassword := PadOrTruncatePassword(password)

	// Hash password
	hash := md5.New()
	hash.Write(paddedPassword)
	digest := hash.Sum(nil)

	// (Revision 3+) Do 50 iterations
	if revision >= 3 {
		for i := 0; i < 50; i++ {
			hash = md5.New()
			hash.Write(digest)
			digest = hash.Sum(nil)
		}
	}

	// Create RC4 key
	key := make([]byte, keyLength)
	copy(key, digest)

	// Decrypt O value
	userPassword := make([]byte, len(o))
	copy(userPassword, o)

	if revision >= 3 {
		// Do 19 iterations in reverse
		for i := 19; i >= 0; i-- {
			newKey := make([]byte, keyLength)
			for j := 0; j < keyLength; j++ {
				newKey[j] = key[j] ^ byte(i)
			}
			cipher := NewRC4(newKey)
			userPassword = cipher.Decrypt(userPassword)
		}
	} else {
		// Revision 2: Single decryption
		cipher := NewRC4(key)
		userPassword = cipher.Decrypt(userPassword)
	}

	// Remove padding from user password
	for i := 0; i < len(userPassword); i++ {
		// Find start of padding string
		match := true
		for j := 0; i+j < len(userPassword) && j < len(PDFPaddingString); j++ {
			if userPassword[i+j] != PDFPaddingString[j] {
				match = false
				break
			}
		}
		if match {
			return string(userPassword[:i]), true
		}
	}

	return string(userPassword), true
}
