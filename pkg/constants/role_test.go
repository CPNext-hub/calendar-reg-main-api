package constants

import "testing"

func TestValidRoles_ContainsAllRoles(t *testing.T) {
	expected := []string{RoleSuperAdmin, RoleAdmin, RoleStudent}
	for _, role := range expected {
		if !ValidRoles[role] {
			t.Errorf("expected ValidRoles to contain %q", role)
		}
	}
}

func TestValidRoles_RejectsInvalid(t *testing.T) {
	invalid := []string{"", "guest", "moderator", "ADMIN"}
	for _, role := range invalid {
		if ValidRoles[role] {
			t.Errorf("expected ValidRoles to NOT contain %q", role)
		}
	}
}

func TestPrivilegedRoles_ContainsSuperAdminAndAdmin(t *testing.T) {
	if !PrivilegedRoles[RoleSuperAdmin] {
		t.Error("expected PrivilegedRoles to contain superadmin")
	}
	if !PrivilegedRoles[RoleAdmin] {
		t.Error("expected PrivilegedRoles to contain admin")
	}
}

func TestPrivilegedRoles_ExcludesStudent(t *testing.T) {
	if PrivilegedRoles[RoleStudent] {
		t.Error("expected PrivilegedRoles to NOT contain student")
	}
}

func TestRoleConstants_Values(t *testing.T) {
	if RoleSuperAdmin != "superadmin" {
		t.Errorf("expected RoleSuperAdmin='superadmin', got %q", RoleSuperAdmin)
	}
	if RoleAdmin != "admin" {
		t.Errorf("expected RoleAdmin='admin', got %q", RoleAdmin)
	}
	if RoleStudent != "student" {
		t.Errorf("expected RoleStudent='student', got %q", RoleStudent)
	}
}
