package security

// RC4Cipher implements the RC4 stream cipher
// RC4 is used for PDF encryption (40-bit and 128-bit)
type RC4Cipher struct {
	S    [256]byte
	i, j byte
}

// NewRC4 creates a new RC4 cipher with the given key
func NewRC4(key []byte) *RC4Cipher {
	cipher := &RC4Cipher{}
	cipher.init(key)
	return cipher
}

// init initializes the RC4 cipher with the key (KSA: Key-Scheduling Algorithm)
func (r *RC4Cipher) init(key []byte) {
	keyLen := len(key)
	if keyLen == 0 {
		return
	}

	// Initialize S-box
	for i := 0; i < 256; i++ {
		r.S[i] = byte(i)
	}

	// Key-scheduling algorithm
	j := byte(0)
	for i := 0; i < 256; i++ {
		j = j + r.S[i] + key[i%keyLen]
		r.S[i], r.S[j] = r.S[j], r.S[i]
	}

	r.i = 0
	r.j = 0
}

// XORKeyStream encrypts or decrypts data using RC4
// RC4 is symmetric: encryption and decryption use the same operation
func (r *RC4Cipher) XORKeyStream(dst, src []byte) {
	if len(dst) < len(src) {
		panic("rc4: dst buffer too small")
	}

	for k := 0; k < len(src); k++ {
		r.i++
		r.j += r.S[r.i]
		r.S[r.i], r.S[r.j] = r.S[r.j], r.S[r.i]
		dst[k] = src[k] ^ r.S[r.S[r.i]+r.S[r.j]]
	}
}

// Encrypt encrypts the data using RC4
func (r *RC4Cipher) Encrypt(data []byte) []byte {
	encrypted := make([]byte, len(data))
	r.XORKeyStream(encrypted, data)
	return encrypted
}

// Decrypt decrypts the data using RC4 (same as Encrypt due to XOR nature)
func (r *RC4Cipher) Decrypt(data []byte) []byte {
	decrypted := make([]byte, len(data))
	r.XORKeyStream(decrypted, data)
	return decrypted
}

// Reset resets the cipher state with a new key
func (r *RC4Cipher) Reset(key []byte) {
	r.init(key)
}
