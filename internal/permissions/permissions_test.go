package permissions

import (
	"testing"

	"github.com/fivemanage/lite/api"
)

func TestPresetAdminCanReadAndWriteAllModules(t *testing.T) {
	preset := Preset(api.MemberRoleAdmin)
	for _, module := range Modules {
		access := preset[module]
		if !access.Read || !access.Write {
			t.Fatalf("expected admin read/write for %s", module)
		}
	}
}

func TestNormalizeDefaultsToViewer(t *testing.T) {
	role, perms, err := Normalize("", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if role != api.MemberRoleViewer {
		t.Fatalf("expected viewer role, got %s", role)
	}
	if !perms[api.PermissionModuleOverview].Read || perms[api.PermissionModuleOverview].Write {
		t.Fatalf("expected viewer overview read-only")
	}
	if perms[api.PermissionModuleTokens].Read || perms[api.PermissionModuleTokens].Write {
		t.Fatalf("expected viewer to have no token access")
	}
}

func TestNormalizeRejectsInvalidModule(t *testing.T) {
	_, _, err := Normalize(api.MemberRoleViewer, api.MemberPermissions{
		"billing": {Read: true},
	})
	if err == nil {
		t.Fatalf("expected invalid module error")
	}
}

func TestNormalizeWriteImpliesRead(t *testing.T) {
	_, perms, err := Normalize(api.MemberRoleViewer, api.MemberPermissions{
		api.PermissionModuleTokens: {Write: true},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !perms[api.PermissionModuleTokens].Read || !perms[api.PermissionModuleTokens].Write {
		t.Fatalf("expected write to imply read")
	}
}

func TestCanUsesPersistedPermissions(t *testing.T) {
	perms := api.MemberPermissions{
		api.PermissionModuleStorage: {Read: true, Write: false},
	}
	if !Can(api.MemberRoleViewer, perms, api.PermissionModuleStorage, api.PermissionActionRead) {
		t.Fatalf("expected storage read")
	}
	if Can(api.MemberRoleViewer, perms, api.PermissionModuleStorage, api.PermissionActionWrite) {
		t.Fatalf("did not expect storage write")
	}
}
