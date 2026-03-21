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

// TestCreateEncryptDictionaryValues は暗号化辞書の具体的な値をテストする
func TestCreateEncryptDictionaryValues(t *testing.T) {
	tests := []struct {
		name        string
		keyLength   int
		permissions security.Permissions
		wantV       int
		wantR       int
		wantLength  bool // whether Length key should be present
	}{
		{
			name:        "40-bit with default permissions",
			keyLength:   40,
			permissions: security.DefaultPermissions(),
			wantV:       1,
			wantR:       2,
			wantLength:  false,
		},
		{
			name:        "128-bit with default permissions",
			keyLength:   128,
			permissions: security.DefaultPermissions(),
			wantV:       2,
			wantR:       3,
			wantLength:  true,
		},
		{
			name:        "40-bit with restricted permissions",
			keyLength:   40,
			permissions: security.RestrictedPermissions(),
			wantV:       1,
			wantR:       2,
			wantLength:  false,
		},
		{
			name:        "128-bit with restricted permissions",
			keyLength:   128,
			permissions: security.RestrictedPermissions(),
			wantV:       2,
			wantR:       3,
			wantLength:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := SetupEncryption("user", "owner", tt.permissions, tt.keyLength)
			if err != nil {
				t.Fatalf("SetupEncryption failed: %v", err)
			}

			dict := info.CreateEncryptDictionary()

			// Check Filter value
			filter, ok := dict["Filter"].(core.Name)
			if !ok {
				t.Fatal("Filter should be core.Name")
			}
			if string(filter) != "Standard" {
				t.Errorf("Filter = %q, want %q", filter, "Standard")
			}

			// Check V value
			v, ok := dict["V"].(core.Integer)
			if !ok {
				t.Fatal("V should be core.Integer")
			}
			if int(v) != tt.wantV {
				t.Errorf("V = %d, want %d", v, tt.wantV)
			}

			// Check R value
			r, ok := dict["R"].(core.Integer)
			if !ok {
				t.Fatal("R should be core.Integer")
			}
			if int(r) != tt.wantR {
				t.Errorf("R = %d, want %d", r, tt.wantR)
			}

			// Check O value is a string with length 32
			o, ok := dict["O"].(core.String)
			if !ok {
				t.Fatal("O should be core.String")
			}
			if len(o) != 32 {
				t.Errorf("O length = %d, want 32", len(o))
			}

			// Check U value
			u, ok := dict["U"].(core.String)
			if !ok {
				t.Fatal("U should be core.String")
			}
			if len(u) < 16 {
				t.Errorf("U length = %d, want >= 16", len(u))
			}

			// Check P value
			p, ok := dict["P"].(core.Integer)
			if !ok {
				t.Fatal("P should be core.Integer")
			}
			expectedP := tt.permissions.ToInt32()
			if int32(p) != expectedP {
				t.Errorf("P = %d, want %d", p, expectedP)
			}

			// Check Length key presence
			_, hasLength := dict["Length"]
			if hasLength != tt.wantLength {
				t.Errorf("Length present = %v, want %v", hasLength, tt.wantLength)
			}

			if tt.wantLength {
				length, ok := dict["Length"].(core.Integer)
				if !ok {
					t.Fatal("Length should be core.Integer")
				}
				if int(length) != tt.keyLength {
					t.Errorf("Length = %d, want %d", length, tt.keyLength)
				}
			}
		})
	}
}

// TestCreateFileIDArrayContent はCreateFileIDArrayの具体的な内容をテストする
func TestCreateFileIDArrayContent(t *testing.T) {
	tests := []struct {
		name      string
		keyLength int
	}{
		{"40-bit", 40},
		{"128-bit", 128},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := SetupEncryption("user", "owner", security.DefaultPermissions(), tt.keyLength)
			if err != nil {
				t.Fatalf("SetupEncryption failed: %v", err)
			}

			idArray := info.CreateFileIDArray()

			if len(idArray) != 2 {
				t.Fatalf("FileID array length = %d, want 2", len(idArray))
			}

			// Both elements should be core.String
			id1, ok := idArray[0].(core.String)
			if !ok {
				t.Fatal("first element should be core.String")
			}

			id2, ok := idArray[1].(core.String)
			if !ok {
				t.Fatal("second element should be core.String")
			}

			// Both should be 16 bytes
			if len(id1) != 16 {
				t.Errorf("first ID length = %d, want 16", len(id1))
			}
			if len(id2) != 16 {
				t.Errorf("second ID length = %d, want 16", len(id2))
			}

			// Should match the FileID
			if string(id1) != string(info.FileID) {
				t.Error("first ID should match FileID")
			}
			if string(id2) != string(info.FileID) {
				t.Error("second ID should match FileID")
			}

			// Both IDs should be identical
			if string(id1) != string(id2) {
				t.Error("both IDs should be identical")
			}
		})
	}
}

// TestSetupEncryptionEmptyPasswords は空パスワードのセットアップをテストする
func TestSetupEncryptionEmptyPasswords(t *testing.T) {
	info, err := SetupEncryption("", "", security.DefaultPermissions(), 40)
	if err != nil {
		t.Fatalf("SetupEncryption with empty passwords failed: %v", err)
	}

	// Should still produce valid values
	if len(info.OValue) != 32 {
		t.Errorf("OValue length = %d, want 32", len(info.OValue))
	}
	if len(info.UValue) < 16 {
		t.Errorf("UValue length = %d, want >= 16", len(info.UValue))
	}
	if len(info.EncryptionKey) != 5 { // 40/8 = 5 bytes
		t.Errorf("EncryptionKey length = %d, want 5", len(info.EncryptionKey))
	}

	// CreateEncryptDictionary should still work
	dict := info.CreateEncryptDictionary()
	if dict["Filter"] == nil {
		t.Error("Encrypt dictionary should have Filter even with empty passwords")
	}
}

// TestEncryptionIntegration128Bit は128-bit暗号化の認証テストする
func TestEncryptionIntegration128Bit(t *testing.T) {
	userPass := "testuser"
	ownerPass := "testowner"
	permissions := security.Permissions{
		Print:  true,
		Modify: false,
		Copy:   true,
	}

	info, err := SetupEncryption(userPass, ownerPass, permissions, 128)
	if err != nil {
		t.Fatalf("SetupEncryption failed: %v", err)
	}

	// Verify user password authentication for revision 3
	authenticated := security.AuthenticateUserPassword(
		userPass,
		info.UValue,
		info.OValue,
		permissions.ToInt32(),
		info.FileID,
		3,  // Revision 3 for 128-bit
		16, // 128-bit = 16 bytes
	)

	if !authenticated {
		t.Error("User password authentication failed for 128-bit encryption")
	}

	// Verify encryption key length
	if len(info.EncryptionKey) != 16 {
		t.Errorf("EncryptionKey length = %d, want 16", len(info.EncryptionKey))
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
