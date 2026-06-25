package permissions

import (
	"fmt"
	"strings"

	"github.com/fivemanage/lite/api"
)

var Modules = []string{
	api.PermissionModuleOverview,
	api.PermissionModuleLogs,
	api.PermissionModuleStorage,
	api.PermissionModuleTokens,
	api.PermissionModuleTeam,
}

func NormalizeRole(role string) (string, error) {
	switch strings.ToUpper(strings.TrimSpace(role)) {
	case "", api.MemberRoleViewer:
		return api.MemberRoleViewer, nil
	case api.MemberRoleAdmin:
		return api.MemberRoleAdmin, nil
	case api.MemberRoleEditor:
		return api.MemberRoleEditor, nil
	default:
		return "", fmt.Errorf("invalid member role %q", role)
	}
}

func Preset(role string) api.MemberPermissions {
	normalizedRole, err := NormalizeRole(role)
	if err != nil {
		normalizedRole = api.MemberRoleViewer
	}

	switch normalizedRole {
	case api.MemberRoleAdmin:
		return all(true, true)
	case api.MemberRoleEditor:
		return api.MemberPermissions{
			api.PermissionModuleOverview: {Read: true},
			api.PermissionModuleLogs:     {Read: true, Write: true},
			api.PermissionModuleStorage:  {Read: true, Write: true},
			api.PermissionModuleTokens:   {},
			api.PermissionModuleTeam:     {},
		}
	default:
		return api.MemberPermissions{
			api.PermissionModuleOverview: {Read: true},
			api.PermissionModuleLogs:     {Read: true},
			api.PermissionModuleStorage:  {Read: true},
			api.PermissionModuleTokens:   {},
			api.PermissionModuleTeam:     {},
		}
	}
}

func Normalize(role string, requested api.MemberPermissions) (string, api.MemberPermissions, error) {
	normalizedRole, err := NormalizeRole(role)
	if err != nil {
		return "", nil, err
	}

	normalized := Preset(normalizedRole)
	if requested == nil {
		return normalizedRole, normalized, nil
	}

	for module, access := range requested {
		if !isModule(module) {
			return "", nil, fmt.Errorf("invalid permission module %q", module)
		}
		if access.Write {
			access.Read = true
		}
		normalized[module] = access
	}

	if normalizedRole == api.MemberRoleAdmin {
		normalized = all(true, true)
	}

	return normalizedRole, normalized, nil
}

func Can(role string, memberPermissions api.MemberPermissions, module string, action string) bool {
	normalizedRole, _ := NormalizeRole(role)
	if normalizedRole == api.MemberRoleAdmin {
		return true
	}
	if !isModule(module) {
		return false
	}
	access := Preset(normalizedRole)[module]
	if memberPermissions != nil {
		if customAccess, ok := memberPermissions[module]; ok {
			access = customAccess
		}
	}
	if action == api.PermissionActionWrite {
		return access.Write
	}
	return access.Read
}

func all(read, write bool) api.MemberPermissions {
	perms := make(api.MemberPermissions, len(Modules))
	for _, module := range Modules {
		perms[module] = api.PermissionAccess{Read: read, Write: write}
	}
	return perms
}

func isModule(module string) bool {
	for _, allowed := range Modules {
		if module == allowed {
			return true
		}
	}
	return false
}
