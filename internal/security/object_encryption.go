package security

import (
	"crypto/md5"
)

// ComputeObjectKey computes the encryption key for a specific object
// This key is derived from the document encryption key, object number, and generation number
func ComputeObjectKey(encryptionKey []byte, objectNumber, generationNumber int, keyLength int) []byte {
	// Step 1: Start with the encryption key
	data := make([]byte, len(encryptionKey)+5)
	copy(data, encryptionKey)

	// Step 2: Append object number (3 bytes, low-order byte first)
	data[len(encryptionKey)] = byte(objectNumber)
	data[len(encryptionKey)+1] = byte(objectNumber >> 8)
	data[len(encryptionKey)+2] = byte(objectNumber >> 16)

	// Step 3: Append generation number (2 bytes, low-order byte first)
	data[len(encryptionKey)+3] = byte(generationNumber)
	data[len(encryptionKey)+4] = byte(generationNumber >> 8)

	// Step 4: Compute MD5 hash
	hash := md5.Sum(data)

	// Step 5: Use first (n+5) bytes, with maximum of 16
	resultLength := keyLength + 5
	if resultLength > 16 {
		resultLength = 16
	}

	result := make([]byte, resultLength)
	copy(result, hash[:resultLength])

	return result
}

// EncryptStream encrypts stream data using RC4
func EncryptStream(data []byte, encryptionKey []byte, objectNumber, generationNumber int, keyLength int) []byte {
	// Compute object-specific key
	objectKey := ComputeObjectKey(encryptionKey, objectNumber, generationNumber, keyLength)

	// Encrypt with RC4
	cipher := NewRC4(objectKey)
	return cipher.Encrypt(data)
}

// DecryptStream decrypts stream data using RC4
func DecryptStream(data []byte, encryptionKey []byte, objectNumber, generationNumber int, keyLength int) []byte {
	// Compute object-specific key
	objectKey := ComputeObjectKey(encryptionKey, objectNumber, generationNumber, keyLength)

	// Decrypt with RC4
	cipher := NewRC4(objectKey)
	return cipher.Decrypt(data)
}

// EncryptString encrypts a string value using RC4
func EncryptString(str string, encryptionKey []byte, objectNumber, generationNumber int, keyLength int) []byte {
	return EncryptStream([]byte(str), encryptionKey, objectNumber, generationNumber, keyLength)
}

// DecryptString decrypts a string value using RC4
func DecryptString(data []byte, encryptionKey []byte, objectNumber, generationNumber int, keyLength int) string {
	decrypted := DecryptStream(data, encryptionKey, objectNumber, generationNumber, keyLength)
	return string(decrypted)
}
