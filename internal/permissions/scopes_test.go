package permissions

import (
	"errors"
	"testing"

	"github.com/fivemanage/lite/api"
)

func TestNormalizeTokenScopesEmptyMeansFullAccess(t *testing.T) {
	got, err := NormalizeTokenScopes(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil (full access) for empty request, got %v", got)
	}
}

func TestNormalizeTokenScopesDedupesAndValidates(t *testing.T) {
	got, err := NormalizeTokenScopes([]string{ScopeStorageWrite, ScopeStorageWrite, ScopeLogsWrite})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 de-duped scopes, got %v", got)
	}
}

func TestNormalizeTokenScopesRejectsUnknown(t *testing.T) {
	_, err := NormalizeTokenScopes([]string{"storage:write", "team:write"})
	var invalid *InvalidScopeError
	if !errors.As(err, &invalid) {
		t.Fatalf("expected InvalidScopeError, got %v", err)
	}
	if invalid.Scope != "team:write" {
		t.Fatalf("expected offending scope team:write, got %q", invalid.Scope)
	}
}

func TestTokenAllowsEmptyScopesIsFullAccess(t *testing.T) {
	if !TokenAllows(nil, api.PermissionModuleStorage, api.PermissionActionWrite) {
		t.Fatal("legacy (empty-scope) token should have full access")
	}
}

func TestTokenAllowsRestrictsScopedToken(t *testing.T) {
	scopes := []string{ScopeLogsWrite}
	if !TokenAllows(scopes, api.PermissionModuleLogs, api.PermissionActionWrite) {
		t.Fatal("logs:write token should be allowed to write logs")
	}
	if TokenAllows(scopes, api.PermissionModuleStorage, api.PermissionActionWrite) {
		t.Fatal("logs:write token must NOT be allowed to write storage")
	}
	if TokenAllows(scopes, api.PermissionModuleStorage, api.PermissionActionRead) {
		t.Fatal("logs:write token must NOT be allowed to read storage")
	}
}
