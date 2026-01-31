package security

// Permission flags for PDF access control (bit positions)
const (
	PermPrint         = 1 << 2  // bit 3: Print
	PermModify        = 1 << 3  // bit 4: Modify contents
	PermCopy          = 1 << 4  // bit 5: Copy/extract text and graphics
	PermAnnotate      = 1 << 5  // bit 6: Add or modify annotations
	PermFillForms     = 1 << 8  // bit 9: Fill in form fields
	PermExtract       = 1 << 9  // bit 10: Extract text and graphics
	PermAssemble      = 1 << 10 // bit 11: Assemble document (insert, rotate, delete pages)
	PermPrintHighQual = 1 << 11 // bit 12: Print in high quality
)

// Permissions represents PDF access permissions
type Permissions struct {
	Print            bool // Allow printing
	Modify           bool // Allow modifying contents
	Copy             bool // Allow copying text/graphics
	Annotate         bool // Allow adding annotations
	FillForms        bool // Allow filling form fields
	ExtractContent   bool // Allow extracting content
	Assemble         bool // Allow page operations
	PrintHighQuality bool // Allow high-quality printing
}

// DefaultPermissions returns permissions with all rights granted
func DefaultPermissions() Permissions {
	return Permissions{
		Print:            true,
		Modify:           true,
		Copy:             true,
		Annotate:         true,
		FillForms:        true,
		ExtractContent:   true,
		Assemble:         true,
		PrintHighQuality: true,
	}
}

// RestrictedPermissions returns permissions with minimal rights (view only)
func RestrictedPermissions() Permissions {
	return Permissions{
		Print:            false,
		Modify:           false,
		Copy:             false,
		Annotate:         false,
		FillForms:        false,
		ExtractContent:   false,
		Assemble:         false,
		PrintHighQuality: false,
	}
}

// PrintOnlyPermissions returns permissions allowing only printing
func PrintOnlyPermissions() Permissions {
	return Permissions{
		Print:            true,
		Modify:           false,
		Copy:             false,
		Annotate:         false,
		FillForms:        false,
		ExtractContent:   false,
		Assemble:         false,
		PrintHighQuality: true,
	}
}

// ToInt32 converts Permissions to a 32-bit integer flag
// According to PDF specification, bits 1-2, 7-8, and 13-32 must be set to 1
func (p Permissions) ToInt32() int32 {
	var flags int32 = -3904 // 0xFFFFF0C0 as signed int32

	if p.Print {
		flags |= PermPrint
	}
	if p.Modify {
		flags |= PermModify
	}
	if p.Copy {
		flags |= PermCopy
	}
	if p.Annotate {
		flags |= PermAnnotate
	}
	if p.FillForms {
		flags |= PermFillForms
	}
	if p.ExtractContent {
		flags |= PermExtract
	}
	if p.Assemble {
		flags |= PermAssemble
	}
	if p.PrintHighQuality {
		flags |= PermPrintHighQual
	}

	return flags
}

// FromInt32 creates Permissions from a 32-bit integer flag
func FromInt32(flags int32) Permissions {
	return Permissions{
		Print:            (flags & PermPrint) != 0,
		Modify:           (flags & PermModify) != 0,
		Copy:             (flags & PermCopy) != 0,
		Annotate:         (flags & PermAnnotate) != 0,
		FillForms:        (flags & PermFillForms) != 0,
		ExtractContent:   (flags & PermExtract) != 0,
		Assemble:         (flags & PermAssemble) != 0,
		PrintHighQuality: (flags & PermPrintHighQual) != 0,
	}
}
