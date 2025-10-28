package security

import (
	"bytes"
	"testing"
)

func TestComputeObjectKey(t *testing.T) {
	encryptionKey := []byte{0x01, 0x02, 0x03, 0x04, 0x05} // 40-bit key
	objectNumber := 10
	generationNumber := 0

	key := ComputeObjectKey(encryptionKey, objectNumber, generationNumber, 5)

	// Key should be (5 + 5) = 10 bytes for 40-bit encryption
	expectedLength := 10
	if len(key) != expectedLength {
		t.Errorf("Object key length = %d, want %d", len(key), expectedLength)
	}

	// Verify that different object numbers produce different keys
	key2 := ComputeObjectKey(encryptionKey, 11, generationNumber, 5)
	if bytes.Equal(key, key2) {
		t.Error("Different object numbers should produce different keys")
	}

	// Verify that different generation numbers produce different keys
	key3 := ComputeObjectKey(encryptionKey, objectNumber, 1, 5)
	if bytes.Equal(key, key3) {
		t.Error("Different generation numbers should produce different keys")
	}
}

func TestComputeObjectKey128Bit(t *testing.T) {
	// 128-bit encryption key
	encryptionKey := make([]byte, 16)
	for i := range encryptionKey {
		encryptionKey[i] = byte(i)
	}

	objectNumber := 20
	generationNumber := 0

	key := ComputeObjectKey(encryptionKey, objectNumber, generationNumber, 16)

	// Key should be maximum 16 bytes for 128-bit encryption
	// (16 + 5 = 21, but capped at 16)
	expectedLength := 16
	if len(key) != expectedLength {
		t.Errorf("Object key length = %d, want %d", len(key), expectedLength)
	}
}

func TestEncryptDecryptStream(t *testing.T) {
	encryptionKey := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	originalData := []byte("This is a test stream content")
	objectNumber := 5
	generationNumber := 0

	// Encrypt the data
	encrypted := EncryptStream(originalData, encryptionKey, objectNumber, generationNumber, 5)

	// Encrypted data should be different from original
	if bytes.Equal(originalData, encrypted) {
		t.Error("Encrypted data should differ from original")
	}

	// Encrypted data should have the same length (RC4 stream cipher)
	if len(encrypted) != len(originalData) {
		t.Errorf("Encrypted length = %d, want %d", len(encrypted), len(originalData))
	}

	// Decrypt the data
	decrypted := DecryptStream(encrypted, encryptionKey, objectNumber, generationNumber, 5)

	// Decrypted data should match original
	if !bytes.Equal(originalData, decrypted) {
		t.Errorf("Decrypted data does not match original.\nOriginal: %s\nDecrypted: %s",
			string(originalData), string(decrypted))
	}
}

func TestEncryptDecryptString(t *testing.T) {
	encryptionKey := []byte{0xAA, 0xBB, 0xCC, 0xDD, 0xEE}
	originalString := "Hello, PDF encryption!"
	objectNumber := 15
	generationNumber := 0

	// Encrypt the string
	encrypted := EncryptString(originalString, encryptionKey, objectNumber, generationNumber, 5)

	// Encrypted data should be different
	if string(encrypted) == originalString {
		t.Error("Encrypted string should differ from original")
	}

	// Decrypt the string
	decrypted := DecryptString(encrypted, encryptionKey, objectNumber, generationNumber, 5)

	// Should match original
	if decrypted != originalString {
		t.Errorf("Decrypted string = %q, want %q", decrypted, originalString)
	}
}

func TestStreamEncryptionWithDifferentData(t *testing.T) {
	encryptionKey := []byte{0x12, 0x34, 0x56, 0x78, 0x9A}
	objectNumber := 3
	generationNumber := 0

	testCases := []struct {
		name string
		data []byte
	}{
		{"Empty data", []byte{}},
		{"Single byte", []byte{0xFF}},
		{"Small data", []byte("Test")},
		{"Medium data", []byte("This is a longer test string with more content to encrypt and decrypt.")},
		{"Binary data", []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD, 0xFC}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encrypted := EncryptStream(tc.data, encryptionKey, objectNumber, generationNumber, 5)
			decrypted := DecryptStream(encrypted, encryptionKey, objectNumber, generationNumber, 5)

			if !bytes.Equal(tc.data, decrypted) {
				t.Errorf("Round-trip encryption failed for %s", tc.name)
			}
		})
	}
}

func TestEncryptionWithDifferentKeys(t *testing.T) {
	data := []byte("Sensitive data")
	objectNumber := 1
	generationNumber := 0

	key1 := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	key2 := []byte{0x05, 0x04, 0x03, 0x02, 0x01}

	encrypted1 := EncryptStream(data, key1, objectNumber, generationNumber, 5)
	encrypted2 := EncryptStream(data, key2, objectNumber, generationNumber, 5)

	// Different keys should produce different ciphertext
	if bytes.Equal(encrypted1, encrypted2) {
		t.Error("Different encryption keys should produce different ciphertext")
	}

	// Each should decrypt correctly with its own key
	decrypted1 := DecryptStream(encrypted1, key1, objectNumber, generationNumber, 5)
	if !bytes.Equal(data, decrypted1) {
		t.Error("Failed to decrypt with key1")
	}

	decrypted2 := DecryptStream(encrypted2, key2, objectNumber, generationNumber, 5)
	if !bytes.Equal(data, decrypted2) {
		t.Error("Failed to decrypt with key2")
	}
}
