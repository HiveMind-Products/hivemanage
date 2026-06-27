package permissions

import (
	"testing"

	"github.com/fivemanage/lite/api"
)

// These invariants underpin the member-service privilege-escalation guard
// (member.ErrEscalation): the admin role must always carry the full permission
// set, and role/module normalization must reject anything it does not
// explicitly recognize.

func TestNormalizeRoleAcceptsValidRoles(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"empty-defaults-viewer", "", api.MemberRoleViewer},
		{"viewer", "VIEWER", api.MemberRoleViewer},
		{"lowercase-viewer", "viewer", api.MemberRoleViewer},
		{"admin", "ADMIN", api.MemberRoleAdmin},
		{"padded-admin", "  admin  ", api.MemberRoleAdmin},
		{"editor", "EDITOR", api.MemberRoleEditor},
		{"mixed-case-editor", "EdItOr", api.MemberRoleEditor},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NormalizeRole(tc.in)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("NormalizeRole(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestNormalizeRoleRejectsInvalidRoles(t *testing.T) {
	for _, role := range []string{"superadmin", "owner", "guest", "ADMINISTRATOR", "root"} {
		if _, err := NormalizeRole(role); err == nil {
			t.Fatalf("NormalizeRole(%q) accepted an invalid role", role)
		}
	}
}

func TestNormalizeRejectsInvalidRole(t *testing.T) {
	if _, _, err := Normalize("owner", nil); err == nil {
		t.Fatalf("expected Normalize to reject invalid role")
	}
}

func TestAdminRoleAlwaysGetsAllPermissions(t *testing.T) {
	// Even if a caller requests a reduced permission set for an admin, Normalize
	// must force the full all() set — admin is all-or-nothing.
	requested := api.MemberPermissions{
		api.PermissionModuleTeam:   {},
		api.PermissionModuleTokens: {Read: false, Write: false},
	}
	role, perms, err := Normalize(api.MemberRoleAdmin, requested)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if role != api.MemberRoleAdmin {
		t.Fatalf("role = %q, want admin", role)
	}
	for _, module := range Modules {
		access := perms[module]
		if !access.Read || !access.Write {
			t.Fatalf("admin must have read+write on %s, got %+v", module, access)
		}
	}
}

func TestAllProducesEveryModuleReadWrite(t *testing.T) {
	perms := all()
	if len(perms) != len(Modules) {
		t.Fatalf("all() has %d modules, want %d", len(perms), len(Modules))
	}
	for _, module := range Modules {
		access, ok := perms[module]
		if !ok {
			t.Fatalf("all() missing module %s", module)
		}
		if !access.Read || !access.Write {
			t.Fatalf("all() module %s should be read+write, got %+v", module, access)
		}
	}
}

func TestWriteImpliesReadAcrossModules(t *testing.T) {
	for _, module := range Modules {
		_, perms, err := Normalize(api.MemberRoleViewer, api.MemberPermissions{
			module: {Write: true},
		})
		if err != nil {
			t.Fatalf("module %s: unexpected error: %v", module, err)
		}
		if !perms[module].Read || !perms[module].Write {
			t.Fatalf("module %s: write must imply read, got %+v", module, perms[module])
		}
	}
}

func TestCanAdminAlwaysAllowed(t *testing.T) {
	for _, module := range Modules {
		for _, action := range []string{api.PermissionActionRead, api.PermissionActionWrite} {
			if !Can(api.MemberRoleAdmin, nil, module, action) {
				t.Fatalf("admin should be allowed %s on %s", action, module)
			}
		}
	}
	// Admin is allowed even for an unknown module, since admin short-circuits.
	if !Can(api.MemberRoleAdmin, nil, "billing", api.PermissionActionWrite) {
		t.Fatalf("admin should be allowed regardless of module")
	}
}

func TestCanRejectsUnknownModuleForNonAdmin(t *testing.T) {
	if Can(api.MemberRoleEditor, nil, "billing", api.PermissionActionRead) {
		t.Fatalf("non-admin must not be granted access to an unknown module")
	}
}
