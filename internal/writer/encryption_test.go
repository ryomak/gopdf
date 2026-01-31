package writer

import (
	"testing"

	"github.com/ryomak/gopdf/internal/security"
)

func TestGenerateFileID(t *testing.T) {
	// Test that FileID is generated with correct length
	fileID, err := GenerateFileID()
	if err != nil {
		t.Fatalf("GenerateFileID failed: %v", err)
	}

	if len(fileID) != 16 {
		t.Errorf("FileID length = %d, want 16", len(fileID))
	}

	// Test that multiple calls generate different IDs
	fileID2, err := GenerateFileID()
	if err != nil {
		t.Fatalf("GenerateFileID failed: %v", err)
	}

	if string(fileID) == string(fileID2) {
		t.Error("FileIDs should be different")
	}
}

func TestSetupEncryption(t *testing.T) {
	tests := []struct {
		name          string
		userPassword  string
		ownerPassword string
		permissions   security.Permissions
		keyLength     int
		wantErr       bool
	}{
		{
			name:          "40-bit encryption with both passwords",
			userPassword:  "user123",
			ownerPassword: "owner123",
			permissions:   security.DefaultPermissions(),
			keyLength:     40,
			wantErr:       false,
		},
		{
			name:          "128-bit encryption",
			userPassword:  "user123",
			ownerPassword: "owner123",
			permissions:   security.RestrictedPermissions(),
			keyLength:     128,
			wantErr:       false,
		},
		{
			name:          "Only user password",
			userPassword:  "user123",
			ownerPassword: "",
			permissions:   security.DefaultPermissions(),
			keyLength:     40,
			wantErr:       false,
		},
		{
			name:          "Only owner password",
			userPassword:  "",
			ownerPassword: "owner123",
			permissions:   security.DefaultPermissions(),
			keyLength:     40,
			wantErr:       false,
		},
		{
			name:          "Invalid key length",
			userPassword:  "user123",
			ownerPassword: "owner123",
			permissions:   security.DefaultPermissions(),
			keyLength:     64, // Invalid
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := SetupEncryption(tt.userPassword, tt.ownerPassword, tt.permissions, tt.keyLength)

			if (err != nil) != tt.wantErr {
				t.Errorf("SetupEncryption() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return // Expected error
			}

			// Verify encryption info
			if len(info.FileID) != 16 {
				t.Errorf("FileID length = %d, want 16", len(info.FileID))
			}

			if info.KeyLength != tt.keyLength {
				t.Errorf("KeyLength = %d, want %d", info.KeyLength, tt.keyLength)
			}

			// Verify O value
			expectedKeyLengthBytes := tt.keyLength / 8
			if len(info.OValue) != 32 {
				t.Errorf("OValue length = %d, want 32", len(info.OValue))
			}

			// Verify U value
			if len(info.UValue) < 16 {
				t.Errorf("UValue length = %d, want >= 16", len(info.UValue))
			}

			// Verify encryption key
			if len(info.EncryptionKey) != expectedKeyLengthBytes {
				t.Errorf("EncryptionKey length = %d, want %d", len(info.EncryptionKey), expectedKeyLengthBytes)
			}
		})
	}
}

func TestCreateEncryptDictionary(t *testing.T) {
	tests := []struct {
		name      string
		keyLength int
		wantV     int
		wantR     int
	}{
		{
			name:      "40-bit encryption",
			keyLength: 40,
			wantV:     1,
			wantR:     2,
		},
		{
			name:      "128-bit encryption",
			keyLength: 128,
			wantV:     2,
			wantR:     3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := SetupEncryption("user", "owner", security.DefaultPermissions(), tt.keyLength)
			if err != nil {
				t.Fatalf("SetupEncryption failed: %v", err)
			}

			dict := info.CreateEncryptDictionary()

			// Check required keys
			if dict["Filter"] == nil {
				t.Error("Filter not set in Encrypt dictionary")
			}

			if dict["V"] == nil {
				t.Error("V not set in Encrypt dictionary")
			}

			if dict["R"] == nil {
				t.Error("R not set in Encrypt dictionary")
			}

			if dict["O"] == nil {
				t.Error("O not set in Encrypt dictionary")
			}

			if dict["U"] == nil {
				t.Error("U not set in Encrypt dictionary")
			}

			if dict["P"] == nil {
				t.Error("P not set in Encrypt dictionary")
			}

			// For 128-bit, Length should be set
			if tt.keyLength == 128 && dict["Length"] == nil {
				t.Error("Length not set for 128-bit encryption")
			}
		})
	}
}

func TestCreateFileIDArray(t *testing.T) {
	info, err := SetupEncryption("user", "owner", security.DefaultPermissions(), 40)
	if err != nil {
		t.Fatalf("SetupEncryption failed: %v", err)
	}

	idArray := info.CreateFileIDArray()

	if len(idArray) != 2 {
		t.Errorf("FileID array length = %d, want 2", len(idArray))
	}

	// Both IDs should be the same in simple implementation
	if idArray[0] != idArray[1] {
		t.Error("FileID array elements should be identical")
	}
}

func TestEncryptionIntegration(t *testing.T) {
	// Test complete encryption setup with password authentication
	userPass := "testuser"
	ownerPass := "testowner"
	permissions := security.Permissions{
		Print:  true,
		Modify: false,
		Copy:   true,
	}

	info, err := SetupEncryption(userPass, ownerPass, permissions, 40)
	if err != nil {
		t.Fatalf("SetupEncryption failed: %v", err)
	}

	// Verify user password authentication
	authenticated := security.AuthenticateUserPassword(
		userPass,
		info.UValue,
		info.OValue,
		permissions.ToInt32(),
		info.FileID,
		2, // Revision 2
		5, // 40-bit = 5 bytes
	)

	if !authenticated {
		t.Error("User password authentication failed")
	}

	// Verify owner password authentication
	recoveredUserPass, ok := security.AuthenticateOwnerPassword(
		ownerPass,
		info.OValue,
		2, // Revision 2
		5, // 40-bit = 5 bytes
	)

	if !ok {
		t.Error("Owner password authentication failed")
	}

	// The recovered password might have padding, so just check it's not empty
	if recoveredUserPass == "" {
		t.Error("Recovered user password is empty")
	}
}
