package permissions

import "github.com/fivemanage/lite/api"

// API token scopes are "module:action" capability strings. They intentionally
// reuse the permission module/action vocabulary so the model stays consistent
// with member permissions.
const (
	ScopeStorageRead  = api.PermissionModuleStorage + ":" + api.PermissionActionRead
	ScopeStorageWrite = api.PermissionModuleStorage + ":" + api.PermissionActionWrite
	ScopeLogsWrite    = api.PermissionModuleLogs + ":" + api.PermissionActionWrite
)

// validTokenScopes is the set of scopes a token may be granted. It is limited to
// the actions the public API actually exposes.
var validTokenScopes = map[string]bool{
	ScopeStorageRead:  true,
	ScopeStorageWrite: true,
	ScopeLogsWrite:    true,
}

// ScopeString builds the "module:action" scope key.
func ScopeString(module, action string) string {
	return module + ":" + action
}

// IsValidTokenScope reports whether s is a scope a token may be granted.
func IsValidTokenScope(s string) bool {
	return validTokenScopes[s]
}

// NormalizeTokenScopes validates a requested scope list, returning a de-duped
// copy. A nil/empty request is allowed and yields nil (full access). An unknown
// scope is rejected.
func NormalizeTokenScopes(requested []string) ([]string, error) {
	if len(requested) == 0 {
		return nil, nil
	}
	seen := make(map[string]bool, len(requested))
	out := make([]string, 0, len(requested))
	for _, s := range requested {
		if !IsValidTokenScope(s) {
			return nil, &InvalidScopeError{Scope: s}
		}
		if seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out, nil
}

// TokenAllows reports whether a token with the given scopes may perform the
// module:action. An empty scope list means full access (legacy tokens created
// before scoping existed).
func TokenAllows(scopes []string, module, action string) bool {
	if len(scopes) == 0 {
		return true
	}
	want := ScopeString(module, action)
	for _, s := range scopes {
		if s == want {
			return true
		}
	}
	return false
}

// InvalidScopeError is returned when a requested token scope is not recognized.
type InvalidScopeError struct{ Scope string }

func (e *InvalidScopeError) Error() string {
	return "invalid token scope: " + e.Scope
}
