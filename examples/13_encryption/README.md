# PDF Encryption Examples

This directory contains examples demonstrating PDF encryption and password protection features.

## Overview

The gopdf library supports PDF encryption using the Standard Security Handler with RC4 encryption algorithm. You can protect PDFs with passwords and control access permissions.

## Features

- **Password Protection**: Protect PDFs with user and owner passwords
- **Access Control**: Control what users can do with the PDF (print, copy, modify, etc.)
- **Encryption Strength**: Support for both 40-bit and 128-bit RC4 encryption

## Examples

### Example 1: Basic Password Protection

Creates a PDF with basic password protection. All permissions are allowed.

```go
doc.SetEncryption(gopdf.EncryptionOptions{
    UserPassword:  "user123",
    OwnerPassword: "owner123",
    Permissions:   gopdf.DefaultPermissions(),
    KeyLength:     40,
})
```

- **User Password**: `user123` - Required to open the PDF
- **Owner Password**: `owner123` - Full access to modify settings
- **Permissions**: All operations allowed
- **Encryption**: 40-bit RC4

### Example 2: Restricted Permissions

Creates a view-only PDF. Users can open and view but cannot print, copy, or modify.

```go
doc.SetEncryption(gopdf.EncryptionOptions{
    UserPassword:  "user123",
    OwnerPassword: "owner123",
    Permissions:   gopdf.RestrictedPermissions(),
    KeyLength:     40,
})
```

- **Permissions**: All operations denied (view only)

### Example 3: Custom Permissions

Creates a PDF with custom permissions - allows printing and copying but not modification.

```go
permissions := gopdf.Permissions{
    Print:          true,
    Modify:         false,
    Copy:           true,
    Annotate:       false,
    // ... other permissions
}

doc.SetEncryption(gopdf.EncryptionOptions{
    UserPassword:  "user123",
    OwnerPassword: "owner123",
    Permissions:   permissions,
    KeyLength:     40,
})
```

### Example 4: Strong Encryption (128-bit)

Creates a PDF with 128-bit RC4 encryption for enhanced security.

```go
doc.SetEncryption(gopdf.EncryptionOptions{
    UserPassword:  "strongpass",
    OwnerPassword: "ownerpass",
    Permissions:   gopdf.PrintOnlyPermissions(),
    KeyLength:     128,
})
```

- **Encryption**: 128-bit RC4 (stronger than 40-bit)
- **Permissions**: Print only

## Permission Types

The `Permissions` struct allows fine-grained control over PDF access:

| Permission | Description |
|-----------|-------------|
| `Print` | Allow printing the document |
| `Modify` | Allow modifying the contents |
| `Copy` | Allow copying text and graphics |
| `Annotate` | Allow adding or modifying annotations |
| `FillForms` | Allow filling in form fields |
| `ExtractContent` | Allow extracting text and graphics (for accessibility) |
| `Assemble` | Allow inserting, deleting, and rotating pages |
| `PrintHighQuality` | Allow high-resolution printing |

### Pre-defined Permission Sets

- `DefaultPermissions()` - All permissions allowed
- `RestrictedPermissions()` - All permissions denied (view only)
- `PrintOnlyPermissions()` - Only printing allowed

## Encryption Strengths

### 40-bit RC4 (Revision 2)
- Compatible with older PDF readers
- `KeyLength: 40`

### 128-bit RC4 (Revision 3)
- Enhanced security
- `KeyLength: 128`
- Recommended for sensitive documents

## Password Types

### User Password
- Required to open the PDF
- Users opening with the user password have restricted permissions
- Can be empty (PDF opens without password but has restrictions)

### Owner Password
- Provides full access to the PDF
- Can change security settings and permissions
- If only owner password is set, PDF can be opened without password

**Note**: At least one password (user or owner) must be set.

## Running the Examples

```bash
cd examples/13_encryption
go run main.go
```

This will create four PDF files:
- `example1_basic_password.pdf` - Basic password protection
- `example2_restricted.pdf` - View-only permissions
- `example3_custom_permissions.pdf` - Custom permissions
- `example4_strong_encryption.pdf` - 128-bit encryption

## Testing the PDFs

Open the generated PDFs with a PDF reader to test:

1. Try opening with the user password
2. Try printing, copying, or modifying (based on permissions)
3. Open with the owner password for full access

## Security Notes

- Encryption is applied to the PDF metadata and structure
- RC4 40-bit and 128-bit encryption are supported by most PDF readers
- For maximum security, use 128-bit encryption with strong passwords
- The encryption follows the PDF Reference 1.7 specification

## API Reference

```go
type EncryptionOptions struct {
    UserPassword  string      // User password (required to open)
    OwnerPassword string      // Owner password (full access)
    Permissions   Permissions // Access control settings
    KeyLength     int         // 40 or 128 bits
}

// Set encryption on a document
err := doc.SetEncryption(opts)
```
