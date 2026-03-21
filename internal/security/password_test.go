package security

import (
	"bytes"
	"testing"
)

func TestPadOrTruncatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantLen  int
		check    func(t *testing.T, result []byte)
	}{
		{
			name:     "empty password returns padding string",
			password: "",
			wantLen:  32,
			check: func(t *testing.T, result []byte) {
				if !bytes.Equal(result, PDFPaddingString) {
					t.Error("empty password should return full padding string")
				}
			},
		},
		{
			name:     "short password padded with padding string",
			password: "abc",
			wantLen:  32,
			check: func(t *testing.T, result []byte) {
				if result[0] != 'a' || result[1] != 'b' || result[2] != 'c' {
					t.Error("first bytes should be password characters")
				}
				if !bytes.Equal(result[3:], PDFPaddingString[:29]) {
					t.Error("remaining bytes should be from padding string")
				}
			},
		},
		{
			name:     "exact 32 byte password",
			password: "12345678901234567890123456789012",
			wantLen:  32,
			check: func(t *testing.T, result []byte) {
				if string(result) != "12345678901234567890123456789012" {
					t.Error("exact 32-byte password should be returned as-is")
				}
			},
		},
		{
			name:     "longer than 32 bytes truncated",
			password: "123456789012345678901234567890123456",
			wantLen:  32,
			check: func(t *testing.T, result []byte) {
				if string(result) != "12345678901234567890123456789012" {
					t.Error("password longer than 32 should be truncated")
				}
			},
		},
		{
			name:     "single character password",
			password: "x",
			wantLen:  32,
			check: func(t *testing.T, result []byte) {
				if result[0] != 'x' {
					t.Error("first byte should be 'x'")
				}
				if !bytes.Equal(result[1:], PDFPaddingString[:31]) {
					t.Error("remaining bytes should be padding")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PadOrTruncatePassword(tt.password)
			if len(result) != tt.wantLen {
				t.Errorf("length = %d, want %d", len(result), tt.wantLen)
			}
			tt.check(t, result)
		})
	}
}

func TestComputeEncryptionKey(t *testing.T) {
	fileID := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10}
	oValue := make([]byte, 32)

	tests := []struct {
		name        string
		password    string
		o           []byte
		permissions int32
		fileID      []byte
		revision    int
		keyLength   int
		wantLen     int
	}{
		{
			name:        "revision 2 (40-bit) empty password",
			password:    "",
			o:           oValue,
			permissions: -4,
			fileID:      fileID,
			revision:    2,
			keyLength:   5,
			wantLen:     5,
		},
		{
			name:        "revision 2 (40-bit) with password",
			password:    "secret",
			o:           oValue,
			permissions: -4,
			fileID:      fileID,
			revision:    2,
			keyLength:   5,
			wantLen:     5,
		},
		{
			name:        "revision 3 (128-bit) empty password",
			password:    "",
			o:           oValue,
			permissions: -3904,
			fileID:      fileID,
			revision:    3,
			keyLength:   16,
			wantLen:     16,
		},
		{
			name:        "revision 3 (128-bit) with password",
			password:    "mypassword",
			o:           oValue,
			permissions: -3904,
			fileID:      fileID,
			revision:    3,
			keyLength:   16,
			wantLen:     16,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := ComputeEncryptionKey(tt.password, tt.o, tt.permissions, tt.fileID, tt.revision, tt.keyLength)
			if len(key) != tt.wantLen {
				t.Errorf("key length = %d, want %d", len(key), tt.wantLen)
			}
		})
	}

	// Verify determinism: same inputs produce same key
	t.Run("deterministic output", func(t *testing.T) {
		key1 := ComputeEncryptionKey("test", oValue, -4, fileID, 2, 5)
		key2 := ComputeEncryptionKey("test", oValue, -4, fileID, 2, 5)
		if !bytes.Equal(key1, key2) {
			t.Error("same inputs should produce same key")
		}
	})

	// Verify different passwords produce different keys
	t.Run("different passwords produce different keys", func(t *testing.T) {
		key1 := ComputeEncryptionKey("pass1", oValue, -4, fileID, 2, 5)
		key2 := ComputeEncryptionKey("pass2", oValue, -4, fileID, 2, 5)
		if bytes.Equal(key1, key2) {
			t.Error("different passwords should produce different keys")
		}
	})

	// Verify different revisions produce different keys
	t.Run("revision 2 vs 3 produce different keys", func(t *testing.T) {
		key2 := ComputeEncryptionKey("test", oValue, -4, fileID, 2, 5)
		key3 := ComputeEncryptionKey("test", oValue, -3904, fileID, 3, 5)
		if bytes.Equal(key2, key3) {
			t.Error("different revisions should produce different keys")
		}
	})

	// Verify different file IDs produce different keys
	t.Run("different file IDs produce different keys", func(t *testing.T) {
		fileID2 := []byte{0xFF, 0xFE, 0xFD, 0xFC, 0xFB, 0xFA, 0xF9, 0xF8,
			0xF7, 0xF6, 0xF5, 0xF4, 0xF3, 0xF2, 0xF1, 0xF0}
		key1 := ComputeEncryptionKey("test", oValue, -4, fileID, 2, 5)
		key2 := ComputeEncryptionKey("test", oValue, -4, fileID2, 2, 5)
		if bytes.Equal(key1, key2) {
			t.Error("different file IDs should produce different keys")
		}
	})
}

func TestComputeOwnerPassword(t *testing.T) {
	tests := []struct {
		name          string
		ownerPassword string
		userPassword  string
		revision      int
		keyLength     int
		wantLen       int
	}{
		{
			name:          "revision 2 with owner password",
			ownerPassword: "owner",
			userPassword:  "user",
			revision:      2,
			keyLength:     5,
			wantLen:       32,
		},
		{
			name:          "revision 2 without owner password (uses user password)",
			ownerPassword: "",
			userPassword:  "user",
			revision:      2,
			keyLength:     5,
			wantLen:       32,
		},
		{
			name:          "revision 3 with owner password",
			ownerPassword: "owner",
			userPassword:  "user",
			revision:      3,
			keyLength:     16,
			wantLen:       32,
		},
		{
			name:          "revision 3 without owner password (uses user password)",
			ownerPassword: "",
			userPassword:  "user",
			revision:      3,
			keyLength:     16,
			wantLen:       32,
		},
		{
			name:          "revision 3 both passwords empty",
			ownerPassword: "",
			userPassword:  "",
			revision:      3,
			keyLength:     16,
			wantLen:       32,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := ComputeOwnerPassword(tt.ownerPassword, tt.userPassword, tt.revision, tt.keyLength)
			if len(o) != tt.wantLen {
				t.Errorf("O value length = %d, want %d", len(o), tt.wantLen)
			}
		})
	}

	// Verify determinism
	t.Run("deterministic output", func(t *testing.T) {
		o1 := ComputeOwnerPassword("owner", "user", 2, 5)
		o2 := ComputeOwnerPassword("owner", "user", 2, 5)
		if !bytes.Equal(o1, o2) {
			t.Error("same inputs should produce same O value")
		}
	})

	// Verify different owner passwords produce different O values
	t.Run("different owner passwords produce different O values", func(t *testing.T) {
		o1 := ComputeOwnerPassword("owner1", "user", 3, 16)
		o2 := ComputeOwnerPassword("owner2", "user", 3, 16)
		if bytes.Equal(o1, o2) {
			t.Error("different owner passwords should produce different O values")
		}
	})

	// Verify different user passwords produce different O values
	t.Run("different user passwords produce different O values", func(t *testing.T) {
		o1 := ComputeOwnerPassword("owner", "user1", 3, 16)
		o2 := ComputeOwnerPassword("owner", "user2", 3, 16)
		if bytes.Equal(o1, o2) {
			t.Error("different user passwords should produce different O values")
		}
	})

	// Verify revision 2 vs 3 produce different O values
	t.Run("revision 2 vs 3 produce different O values", func(t *testing.T) {
		o2 := ComputeOwnerPassword("owner", "user", 2, 5)
		o3 := ComputeOwnerPassword("owner", "user", 3, 5)
		if bytes.Equal(o2, o3) {
			t.Error("different revisions should produce different O values")
		}
	})
}

func TestComputeUserPassword(t *testing.T) {
	fileID := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10}

	tests := []struct {
		name          string
		encryptionKey []byte
		fileID        []byte
		revision      int
		wantLen       int
	}{
		{
			name:          "revision 2 (40-bit key)",
			encryptionKey: []byte{0x01, 0x02, 0x03, 0x04, 0x05},
			fileID:        fileID,
			revision:      2,
			wantLen:       32,
		},
		{
			name:          "revision 3 (128-bit key)",
			encryptionKey: make([]byte, 16),
			fileID:        fileID,
			revision:      3,
			wantLen:       32,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := ComputeUserPassword(tt.encryptionKey, tt.fileID, tt.revision)
			if len(u) != tt.wantLen {
				t.Errorf("U value length = %d, want %d", len(u), tt.wantLen)
			}
		})
	}

	// Verify determinism
	t.Run("deterministic output", func(t *testing.T) {
		key := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
		u1 := ComputeUserPassword(key, fileID, 2)
		u2 := ComputeUserPassword(key, fileID, 2)
		if !bytes.Equal(u1, u2) {
			t.Error("same inputs should produce same U value")
		}
	})

	// Verify different keys produce different U values
	t.Run("different keys produce different U values", func(t *testing.T) {
		key1 := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
		key2 := []byte{0x05, 0x04, 0x03, 0x02, 0x01}
		u1 := ComputeUserPassword(key1, fileID, 2)
		u2 := ComputeUserPassword(key2, fileID, 2)
		if bytes.Equal(u1, u2) {
			t.Error("different keys should produce different U values")
		}
	})

	// Verify revision 2 vs 3 produce different results
	t.Run("revision 2 vs 3 produce different U values", func(t *testing.T) {
		key := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
		u2 := ComputeUserPassword(key, fileID, 2)
		u3 := ComputeUserPassword(key, fileID, 3)
		if bytes.Equal(u2, u3) {
			t.Error("different revisions should produce different U values")
		}
	})
}

func TestAuthenticateUserPassword(t *testing.T) {
	fileID := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10}

	tests := []struct {
		name        string
		userPass    string
		permissions int32
		revision    int
		keyLength   int
		testPass    string
		want        bool
	}{
		{
			name:        "revision 2 correct password",
			userPass:    "user",
			permissions: -4,
			revision:    2,
			keyLength:   5,
			testPass:    "user",
			want:        true,
		},
		{
			name:        "revision 2 wrong password",
			userPass:    "user",
			permissions: -4,
			revision:    2,
			keyLength:   5,
			testPass:    "wrong",
			want:        false,
		},
		{
			name:        "revision 2 empty password correct",
			userPass:    "",
			permissions: -4,
			revision:    2,
			keyLength:   5,
			testPass:    "",
			want:        true,
		},
		{
			name:        "revision 3 correct password",
			userPass:    "user",
			permissions: -3904,
			revision:    3,
			keyLength:   16,
			testPass:    "user",
			want:        true,
		},
		{
			name:        "revision 3 wrong password",
			userPass:    "user",
			permissions: -3904,
			revision:    3,
			keyLength:   16,
			testPass:    "wrong",
			want:        false,
		},
		{
			name:        "revision 3 empty password correct",
			userPass:    "",
			permissions: -3904,
			revision:    3,
			keyLength:   16,
			testPass:    "",
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// First compute O and U values using the user password
			o := ComputeOwnerPassword("", tt.userPass, tt.revision, tt.keyLength)
			key := ComputeEncryptionKey(tt.userPass, o, tt.permissions, fileID, tt.revision, tt.keyLength)
			u := ComputeUserPassword(key, fileID, tt.revision)

			// Now authenticate with the test password
			got := AuthenticateUserPassword(tt.testPass, u, o, tt.permissions, fileID, tt.revision, tt.keyLength)
			if got != tt.want {
				t.Errorf("AuthenticateUserPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthenticateOwnerPassword(t *testing.T) {
	tests := []struct {
		name          string
		ownerPassword string
		userPassword  string
		revision      int
		keyLength     int
		testPass      string
		wantUser      string
		wantOK        bool
	}{
		{
			name:          "revision 2 correct owner password",
			ownerPassword: "owner",
			userPassword:  "user",
			revision:      2,
			keyLength:     5,
			testPass:      "owner",
			wantUser:      "user",
			wantOK:        true,
		},
		{
			name:          "revision 3 correct owner password",
			ownerPassword: "owner",
			userPassword:  "user",
			revision:      3,
			keyLength:     16,
			testPass:      "owner",
			wantUser:      "user",
			wantOK:        true,
		},
		{
			name:          "revision 2 empty user password",
			ownerPassword: "owner",
			userPassword:  "",
			revision:      2,
			keyLength:     5,
			testPass:      "owner",
			wantUser:      "",
			wantOK:        true,
		},
		{
			name:          "revision 3 empty user password",
			ownerPassword: "owner",
			userPassword:  "",
			revision:      3,
			keyLength:     16,
			testPass:      "owner",
			wantUser:      "",
			wantOK:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Compute O value
			o := ComputeOwnerPassword(tt.ownerPassword, tt.userPassword, tt.revision, tt.keyLength)

			// Authenticate with the test password
			gotUser, gotOK := AuthenticateOwnerPassword(tt.testPass, o, tt.revision, tt.keyLength)
			if gotOK != tt.wantOK {
				t.Errorf("AuthenticateOwnerPassword() ok = %v, want %v", gotOK, tt.wantOK)
			}
			if gotUser != tt.wantUser {
				t.Errorf("AuthenticateOwnerPassword() user = %q, want %q", gotUser, tt.wantUser)
			}
		})
	}

	// Test wrong owner password produces incorrect user password for rev 2
	t.Run("revision 2 wrong owner password gives wrong user", func(t *testing.T) {
		o := ComputeOwnerPassword("owner", "user", 2, 5)
		gotUser, _ := AuthenticateOwnerPassword("wrongowner", o, 2, 5)
		if gotUser == "user" {
			t.Error("wrong owner password should not recover correct user password")
		}
	})

	// Test wrong owner password produces incorrect user password for rev 3
	t.Run("revision 3 wrong owner password gives wrong user", func(t *testing.T) {
		o := ComputeOwnerPassword("owner", "user", 3, 16)
		gotUser, _ := AuthenticateOwnerPassword("wrongowner", o, 3, 16)
		if gotUser == "user" {
			t.Error("wrong owner password should not recover correct user password")
		}
	})
}

// TestFullEncryptionRoundTrip tests the complete encryption workflow:
// setting up encryption parameters, computing O and U values, and authenticating.
func TestFullEncryptionRoundTrip(t *testing.T) {
	tests := []struct {
		name          string
		ownerPassword string
		userPassword  string
		revision      int
		keyLength     int
		permissions   int32
	}{
		{
			name:          "revision 2 40-bit with both passwords",
			ownerPassword: "ownerpass",
			userPassword:  "userpass",
			revision:      2,
			keyLength:     5,
			permissions:   -4,
		},
		{
			name:          "revision 3 128-bit with both passwords",
			ownerPassword: "ownerpass",
			userPassword:  "userpass",
			revision:      3,
			keyLength:     16,
			permissions:   -3904,
		},
		{
			name:          "revision 2 40-bit with empty user password",
			ownerPassword: "ownerpass",
			userPassword:  "",
			revision:      2,
			keyLength:     5,
			permissions:   -4,
		},
		{
			name:          "revision 3 128-bit with empty user password",
			ownerPassword: "ownerpass",
			userPassword:  "",
			revision:      3,
			keyLength:     16,
			permissions:   -3904,
		},
	}

	fileID := []byte{0xDE, 0xAD, 0xBE, 0xEF, 0x01, 0x02, 0x03, 0x04,
		0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Step 1: Compute O value
			o := ComputeOwnerPassword(tt.ownerPassword, tt.userPassword, tt.revision, tt.keyLength)
			if len(o) != 32 {
				t.Fatalf("O value length = %d, want 32", len(o))
			}

			// Step 2: Compute encryption key from user password
			encKey := ComputeEncryptionKey(tt.userPassword, o, tt.permissions, fileID, tt.revision, tt.keyLength)
			if len(encKey) != tt.keyLength {
				t.Fatalf("encryption key length = %d, want %d", len(encKey), tt.keyLength)
			}

			// Step 3: Compute U value
			u := ComputeUserPassword(encKey, fileID, tt.revision)
			if len(u) != 32 {
				t.Fatalf("U value length = %d, want 32", len(u))
			}

			// Step 4: Authenticate user password (should succeed)
			if !AuthenticateUserPassword(tt.userPassword, u, o, tt.permissions, fileID, tt.revision, tt.keyLength) {
				t.Error("user password authentication should succeed")
			}

			// Step 5: Wrong user password should fail
			if AuthenticateUserPassword("wrongpassword", u, o, tt.permissions, fileID, tt.revision, tt.keyLength) {
				t.Error("wrong user password should fail authentication")
			}

			// Step 6: Authenticate owner password and recover user password
			recoveredUser, ok := AuthenticateOwnerPassword(tt.ownerPassword, o, tt.revision, tt.keyLength)
			if !ok {
				t.Error("owner password authentication should succeed")
			}
			if recoveredUser != tt.userPassword {
				t.Errorf("recovered user password = %q, want %q", recoveredUser, tt.userPassword)
			}
		})
	}
}

// TestPermissionsIntegrationWithEncryption verifies permissions work correctly
// when used with encryption key computation.
func TestPermissionsIntegrationWithEncryption(t *testing.T) {
	fileID := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10}
	oValue := make([]byte, 32)

	tests := []struct {
		name  string
		perms Permissions
	}{
		{"default permissions", DefaultPermissions()},
		{"restricted permissions", RestrictedPermissions()},
		{"print only permissions", PrintOnlyPermissions()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			permFlags := tt.perms.ToInt32()
			key := ComputeEncryptionKey("test", oValue, permFlags, fileID, 3, 16)
			if len(key) != 16 {
				t.Errorf("encryption key length = %d, want 16", len(key))
			}
		})
	}

	// Different permissions should produce different encryption keys
	t.Run("different permissions produce different keys", func(t *testing.T) {
		key1 := ComputeEncryptionKey("test", oValue, DefaultPermissions().ToInt32(), fileID, 3, 16)
		key2 := ComputeEncryptionKey("test", oValue, RestrictedPermissions().ToInt32(), fileID, 3, 16)
		if bytes.Equal(key1, key2) {
			t.Error("different permissions should produce different encryption keys")
		}
	})
}
