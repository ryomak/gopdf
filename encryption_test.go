package gopdf

import (
	"bytes"
	"strings"
	"testing"
)

func TestEncryptionOptionsValidate(t *testing.T) {
	tests := []struct {
		name    string
		opts    EncryptionOptions
		wantErr bool
	}{
		{
			name: "Valid: both passwords set",
			opts: EncryptionOptions{
				UserPassword:  "user",
				OwnerPassword: "owner",
				Permissions:   DefaultPermissions(),
				KeyLength:     40,
			},
			wantErr: false,
		},
		{
			name: "Valid: only user password",
			opts: EncryptionOptions{
				UserPassword: "user",
				Permissions:  DefaultPermissions(),
				KeyLength:    40,
			},
			wantErr: false,
		},
		{
			name: "Valid: only owner password",
			opts: EncryptionOptions{
				OwnerPassword: "owner",
				Permissions:   DefaultPermissions(),
				KeyLength:     40,
			},
			wantErr: false,
		},
		{
			name: "Invalid: no passwords",
			opts: EncryptionOptions{
				Permissions: DefaultPermissions(),
				KeyLength:   40,
			},
			wantErr: true,
		},
		{
			name: "Invalid: invalid key length",
			opts: EncryptionOptions{
				UserPassword: "user",
				Permissions:  DefaultPermissions(),
				KeyLength:    64, // Not 40 or 128
			},
			wantErr: true,
		},
		{
			name: "Valid: 128-bit encryption",
			opts: EncryptionOptions{
				UserPassword: "user",
				Permissions:  DefaultPermissions(),
				KeyLength:    128,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEncryptionOptionsGetRevision(t *testing.T) {
	tests := []struct {
		name      string
		keyLength int
		want      int
	}{
		{"40-bit", 40, 2},
		{"128-bit", 128, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := EncryptionOptions{KeyLength: tt.keyLength}
			if got := opts.GetRevision(); got != tt.want {
				t.Errorf("GetRevision() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestEncryptionOptionsGetKeyLengthBytes(t *testing.T) {
	tests := []struct {
		name      string
		keyLength int
		want      int
	}{
		{"40-bit", 40, 5},
		{"128-bit", 128, 16},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := EncryptionOptions{KeyLength: tt.keyLength}
			if got := opts.GetKeyLengthBytes(); got != tt.want {
				t.Errorf("GetKeyLengthBytes() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestDocumentSetEncryption(t *testing.T) {
	tests := []struct {
		name    string
		opts    EncryptionOptions
		wantErr bool
	}{
		{
			name: "Valid encryption",
			opts: EncryptionOptions{
				UserPassword:  "user123",
				OwnerPassword: "owner123",
				Permissions:   DefaultPermissions(),
				KeyLength:     40,
			},
			wantErr: false,
		},
		{
			name: "Invalid encryption (no password)",
			opts: EncryptionOptions{
				Permissions: DefaultPermissions(),
				KeyLength:   40,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := New()
			err := doc.SetEncryption(tt.opts)

			if (err != nil) != tt.wantErr {
				t.Errorf("SetEncryption() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil && !doc.HasEncryption() {
				t.Error("HasEncryption() = false, want true after SetEncryption()")
			}
		})
	}
}

func TestDocumentWithEncryption(t *testing.T) {
	// Create a document with encryption
	doc := New()
	page := doc.AddPage(A4, Portrait)

	// Add some content
	page.SetLineWidth(1)
	page.DrawRectangle(100, 100, 200, 100)

	// Set encryption
	err := doc.SetEncryption(EncryptionOptions{
		UserPassword:  "user123",
		OwnerPassword: "owner123",
		Permissions:   DefaultPermissions(),
		KeyLength:     40,
	})
	if err != nil {
		t.Fatalf("SetEncryption failed: %v", err)
	}

	// Write to buffer
	var buf bytes.Buffer
	if err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo failed: %v", err)
	}

	// Verify PDF contains encryption markers
	pdfContent := buf.String()

	// Check for Encrypt dictionary reference
	if !strings.Contains(pdfContent, "/Encrypt") {
		t.Error("PDF should contain /Encrypt reference")
	}

	// Check for file ID
	if !strings.Contains(pdfContent, "/ID") {
		t.Error("PDF should contain /ID array")
	}

	// Check for encryption dictionary components
	if !strings.Contains(pdfContent, "/Filter /Standard") {
		t.Error("PDF should contain /Filter /Standard")
	}

	// Check for O and U values
	if !strings.Contains(pdfContent, "/O") {
		t.Error("PDF should contain /O (owner password)")
	}

	if !strings.Contains(pdfContent, "/U") {
		t.Error("PDF should contain /U (user password)")
	}

	// Check for permissions
	if !strings.Contains(pdfContent, "/P") {
		t.Error("PDF should contain /P (permissions)")
	}

	// Check for revision
	if !strings.Contains(pdfContent, "/R") {
		t.Error("PDF should contain /R (revision)")
	}

	// Check for version
	if !strings.Contains(pdfContent, "/V") {
		t.Error("PDF should contain /V (version)")
	}
}

func TestDocumentWith128BitEncryption(t *testing.T) {
	doc := New()
	page := doc.AddPage(A4, Portrait)
	page.DrawRectangle(100, 100, 200, 100)

	// Set 128-bit encryption
	err := doc.SetEncryption(EncryptionOptions{
		UserPassword:  "strongpass",
		OwnerPassword: "ownerpass",
		Permissions:   RestrictedPermissions(),
		KeyLength:     128,
	})
	if err != nil {
		t.Fatalf("SetEncryption failed: %v", err)
	}

	// Write to buffer
	var buf bytes.Buffer
	if err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo failed: %v", err)
	}

	pdfContent := buf.String()

	// Check for 128-bit specific markers
	// V should be 2 for 128-bit
	if !strings.Contains(pdfContent, "/V 2") {
		t.Error("128-bit encryption should have /V 2")
	}

	// R should be 3 for 128-bit
	if !strings.Contains(pdfContent, "/R 3") {
		t.Error("128-bit encryption should have /R 3")
	}

	// Length should be specified for V >= 2
	if !strings.Contains(pdfContent, "/Length 128") {
		t.Error("128-bit encryption should have /Length 128")
	}
}

func TestPermissionsPresets(t *testing.T) {
	// Test DefaultPermissions
	defaultPerms := DefaultPermissions()
	if !defaultPerms.Print || !defaultPerms.Modify || !defaultPerms.Copy {
		t.Error("DefaultPermissions should allow all operations")
	}

	// Test RestrictedPermissions
	restrictedPerms := RestrictedPermissions()
	if restrictedPerms.Print || restrictedPerms.Modify || restrictedPerms.Copy {
		t.Error("RestrictedPermissions should deny all operations")
	}

	// Test PrintOnlyPermissions
	printOnly := PrintOnlyPermissions()
	if !printOnly.Print || printOnly.Modify || printOnly.Copy {
		t.Error("PrintOnlyPermissions should only allow printing")
	}

	if !printOnly.PrintHighQuality {
		t.Error("PrintOnlyPermissions should allow high quality printing")
	}
}

func TestDocumentWithoutEncryption(t *testing.T) {
	// Create a document without encryption
	doc := New()
	page := doc.AddPage(A4, Portrait)
	page.DrawRectangle(100, 100, 200, 100)

	// Should not have encryption
	if doc.HasEncryption() {
		t.Error("New document should not have encryption by default")
	}

	// Write to buffer
	var buf bytes.Buffer
	if err := doc.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo failed: %v", err)
	}

	pdfContent := buf.String()

	// Should not contain encryption markers
	if strings.Contains(pdfContent, "/Encrypt") {
		t.Error("Non-encrypted PDF should not contain /Encrypt")
	}
}
