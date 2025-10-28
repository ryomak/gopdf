package security

import "testing"

func TestDefaultPermissions(t *testing.T) {
	perm := DefaultPermissions()

	if !perm.Print || !perm.Modify || !perm.Copy || !perm.Annotate {
		t.Error("DefaultPermissions should grant all permissions")
	}
}

func TestRestrictedPermissions(t *testing.T) {
	perm := RestrictedPermissions()

	if perm.Print || perm.Modify || perm.Copy || perm.Annotate {
		t.Error("RestrictedPermissions should deny all permissions")
	}
}

func TestPermissionsToInt32(t *testing.T) {
	perm := DefaultPermissions()
	flags := perm.ToInt32()

	// Check that all permission bits are set
	if (flags & PermPrint) == 0 {
		t.Error("Print permission not set")
	}
	if (flags & PermModify) == 0 {
		t.Error("Modify permission not set")
	}
}

func TestFromInt32(t *testing.T) {
	flags := int32(PermPrint | PermCopy)
	perm := FromInt32(flags)

	if !perm.Print {
		t.Error("Print permission should be true")
	}
	if !perm.Copy {
		t.Error("Copy permission should be true")
	}
	if perm.Modify {
		t.Error("Modify permission should be false")
	}
}
