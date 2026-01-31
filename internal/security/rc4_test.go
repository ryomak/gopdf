package security

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestRC4_NewRC4(t *testing.T) {
	key := []byte("test key")
	cipher := NewRC4(key)

	if cipher == nil {
		t.Fatal("NewRC4 returned nil")
	}

	// S-box should be initialized
	allZero := true
	for i := 0; i < 256; i++ {
		if cipher.S[i] != byte(i) {
			allZero = false
			break
		}
	}
	if allZero {
		t.Error("S-box not properly initialized")
	}
}

func TestRC4_EncryptDecrypt(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		plaintext string
	}{
		{
			name:      "Simple test",
			key:       "Key",
			plaintext: "Plaintext",
		},
		{
			name:      "Longer text",
			key:       "Wiki",
			plaintext: "pedia",
		},
		{
			name:      "Empty plaintext",
			key:       "test",
			plaintext: "",
		},
		{
			name:      "Single byte",
			key:       "k",
			plaintext: "x",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encrypt
			encCipher := NewRC4([]byte(tt.key))
			encrypted := encCipher.Encrypt([]byte(tt.plaintext))

			// Decrypt
			decCipher := NewRC4([]byte(tt.key))
			decrypted := decCipher.Decrypt(encrypted)

			// Verify
			if string(decrypted) != tt.plaintext {
				t.Errorf("Decrypted text = %q, want %q", string(decrypted), tt.plaintext)
			}
		})
	}
}

func TestRC4_KnownVector(t *testing.T) {
	// Known test vector from Wikipedia
	// Key: "Key", Plaintext: "Plaintext"
	// Expected ciphertext: BBF316E8D940AF0AD3
	key := []byte("Key")
	plaintext := []byte("Plaintext")
	expected, _ := hex.DecodeString("BBF316E8D940AF0AD3")

	cipher := NewRC4(key)
	encrypted := cipher.Encrypt(plaintext)

	if !bytes.Equal(encrypted, expected) {
		t.Errorf("Encrypted = %X, want %X", encrypted, expected)
	}

	// Verify decryption
	cipher2 := NewRC4(key)
	decrypted := cipher2.Decrypt(encrypted)

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("Decrypted = %q, want %q", string(decrypted), string(plaintext))
	}
}

func TestRC4_KnownVector2(t *testing.T) {
	// Another known test vector
	// Key: "Wiki", Plaintext: "pedia"
	// Expected ciphertext: 1021BF0420
	key := []byte("Wiki")
	plaintext := []byte("pedia")
	expected, _ := hex.DecodeString("1021BF0420")

	cipher := NewRC4(key)
	encrypted := cipher.Encrypt(plaintext)

	if !bytes.Equal(encrypted, expected) {
		t.Errorf("Encrypted = %X, want %X", encrypted, expected)
	}

	// Verify decryption
	cipher2 := NewRC4(key)
	decrypted := cipher2.Decrypt(encrypted)

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("Decrypted = %q, want %q", string(decrypted), string(plaintext))
	}
}

func TestRC4_XORKeyStream(t *testing.T) {
	key := []byte("test key")
	plaintext := []byte("Hello, World!")

	cipher1 := NewRC4(key)
	encrypted := make([]byte, len(plaintext))
	cipher1.XORKeyStream(encrypted, plaintext)

	cipher2 := NewRC4(key)
	decrypted := make([]byte, len(encrypted))
	cipher2.XORKeyStream(decrypted, encrypted)

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("Decrypted = %q, want %q", string(decrypted), string(plaintext))
	}
}

func TestRC4_Reset(t *testing.T) {
	key1 := []byte("key1")
	key2 := []byte("key2")
	plaintext := []byte("test data")

	cipher := NewRC4(key1)
	encrypted1 := cipher.Encrypt(plaintext)

	// Reset with new key
	cipher.Reset(key2)
	encrypted2 := cipher.Encrypt(plaintext)

	// Encrypted data should be different with different keys
	if bytes.Equal(encrypted1, encrypted2) {
		t.Error("Reset did not change cipher state")
	}

	// Verify decryption with key2
	cipher.Reset(key2)
	decrypted := cipher.Decrypt(encrypted2)

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("Decrypted after reset = %q, want %q", string(decrypted), string(plaintext))
	}
}

func TestRC4_DifferentKeys(t *testing.T) {
	plaintext := []byte("Secret message")

	// Encrypt with key1
	cipher1 := NewRC4([]byte("key1"))
	encrypted1 := cipher1.Encrypt(plaintext)

	// Encrypt with key2
	cipher2 := NewRC4([]byte("key2"))
	encrypted2 := cipher2.Encrypt(plaintext)

	// Ciphertexts should be different
	if bytes.Equal(encrypted1, encrypted2) {
		t.Error("Different keys produced same ciphertext")
	}

	// Decryption with wrong key should not work
	cipher3 := NewRC4([]byte("key1"))
	wrongDecryption := cipher3.Decrypt(encrypted2)

	if bytes.Equal(wrongDecryption, plaintext) {
		t.Error("Wrong key decrypted correctly (should not happen)")
	}
}

func TestRC4_LargeData(t *testing.T) {
	key := []byte("large data test key")
	plaintext := make([]byte, 10000)
	for i := range plaintext {
		plaintext[i] = byte(i % 256)
	}

	// Encrypt
	cipher1 := NewRC4(key)
	encrypted := cipher1.Encrypt(plaintext)

	// Decrypt
	cipher2 := NewRC4(key)
	decrypted := cipher2.Decrypt(encrypted)

	// Verify
	if !bytes.Equal(decrypted, plaintext) {
		t.Error("Large data encryption/decryption failed")
	}
}

func BenchmarkRC4_Encrypt(b *testing.B) {
	key := []byte("benchmark key")
	data := make([]byte, 1024) // 1KB

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cipher := NewRC4(key)
		cipher.Encrypt(data)
	}
}

func BenchmarkRC4_Decrypt(b *testing.B) {
	key := []byte("benchmark key")
	data := make([]byte, 1024) // 1KB
	cipher := NewRC4(key)
	encrypted := cipher.Encrypt(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cipher := NewRC4(key)
		cipher.Decrypt(encrypted)
	}
}
