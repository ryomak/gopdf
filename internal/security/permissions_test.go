package security

import "testing"

func TestDefaultPermissions(t *testing.T) {
	perm := DefaultPermissions()

	if !perm.Print || !perm.Modify || !perm.Copy || !perm.Annotate {
		t.Error("DefaultPermissions should grant all permissions")
	}
	if !perm.FillForms || !perm.ExtractContent || !perm.Assemble || !perm.PrintHighQuality {
		t.Error("DefaultPermissions should grant all extended permissions")
	}
}

func TestRestrictedPermissions(t *testing.T) {
	perm := RestrictedPermissions()

	if perm.Print || perm.Modify || perm.Copy || perm.Annotate {
		t.Error("RestrictedPermissions should deny all permissions")
	}
	if perm.FillForms || perm.ExtractContent || perm.Assemble || perm.PrintHighQuality {
		t.Error("RestrictedPermissions should deny all extended permissions")
	}
}

func TestPrintOnlyPermissions(t *testing.T) {
	perm := PrintOnlyPermissions()

	if !perm.Print {
		t.Error("PrintOnlyPermissions should allow Print")
	}
	if !perm.PrintHighQuality {
		t.Error("PrintOnlyPermissions should allow PrintHighQuality")
	}
	if perm.Modify {
		t.Error("PrintOnlyPermissions should deny Modify")
	}
	if perm.Copy {
		t.Error("PrintOnlyPermissions should deny Copy")
	}
	if perm.Annotate {
		t.Error("PrintOnlyPermissions should deny Annotate")
	}
	if perm.FillForms {
		t.Error("PrintOnlyPermissions should deny FillForms")
	}
	if perm.ExtractContent {
		t.Error("PrintOnlyPermissions should deny ExtractContent")
	}
	if perm.Assemble {
		t.Error("PrintOnlyPermissions should deny Assemble")
	}
}

func TestPermissionsToInt32(t *testing.T) {
	tests := []struct {
		name  string
		perms Permissions
		check func(t *testing.T, flags int32)
	}{
		{
			name:  "default permissions has all bits set",
			perms: DefaultPermissions(),
			check: func(t *testing.T, flags int32) {
				for _, bit := range []int32{PermPrint, PermModify, PermCopy, PermAnnotate,
					PermFillForms, PermExtract, PermAssemble, PermPrintHighQual} {
					if (flags & bit) == 0 {
						t.Errorf("permission bit 0x%X should be set", bit)
					}
				}
			},
		},
		{
			name:  "restricted permissions has no permission bits set",
			perms: RestrictedPermissions(),
			check: func(t *testing.T, flags int32) {
				for _, bit := range []int32{PermPrint, PermModify, PermCopy, PermAnnotate,
					PermFillForms, PermExtract, PermAssemble, PermPrintHighQual} {
					if (flags & bit) != 0 {
						t.Errorf("permission bit 0x%X should not be set", bit)
					}
				}
			},
		},
		{
			name:  "print only has print bits set",
			perms: PrintOnlyPermissions(),
			check: func(t *testing.T, flags int32) {
				if (flags & PermPrint) == 0 {
					t.Error("Print permission should be set")
				}
				if (flags & PermPrintHighQual) == 0 {
					t.Error("PrintHighQuality permission should be set")
				}
				if (flags & PermModify) != 0 {
					t.Error("Modify permission should not be set")
				}
				if (flags & PermCopy) != 0 {
					t.Error("Copy permission should not be set")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := tt.perms.ToInt32()
			tt.check(t, flags)
		})
	}
}

func TestFromInt32(t *testing.T) {
	tests := []struct {
		name  string
		flags int32
		want  Permissions
	}{
		{
			name:  "print and copy only",
			flags: int32(PermPrint | PermCopy),
			want: Permissions{
				Print: true,
				Copy:  true,
			},
		},
		{
			name:  "all permissions",
			flags: int32(PermPrint | PermModify | PermCopy | PermAnnotate | PermFillForms | PermExtract | PermAssemble | PermPrintHighQual),
			want: Permissions{
				Print:            true,
				Modify:           true,
				Copy:             true,
				Annotate:         true,
				FillForms:        true,
				ExtractContent:   true,
				Assemble:         true,
				PrintHighQuality: true,
			},
		},
		{
			name:  "no permissions",
			flags: 0,
			want:  Permissions{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			perm := FromInt32(tt.flags)
			if perm != tt.want {
				t.Errorf("FromInt32(%d) = %+v, want %+v", tt.flags, perm, tt.want)
			}
		})
	}
}

func TestPermissionsRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		perms Permissions
	}{
		{"default", DefaultPermissions()},
		{"restricted", RestrictedPermissions()},
		{"print only", PrintOnlyPermissions()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := tt.perms.ToInt32()
			roundTripped := FromInt32(flags)
			if roundTripped != tt.perms {
				t.Errorf("round trip failed: got %+v, want %+v", roundTripped, tt.perms)
			}
		})
	}
}
