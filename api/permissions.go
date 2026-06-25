package api

type PermissionAccess struct {
	Read  bool `json:"read"`
	Write bool `json:"write"`
}

type MemberPermissions map[string]PermissionAccess

const (
	PermissionModuleOverview = "overview"
	PermissionModuleLogs     = "logs"
	PermissionModuleStorage  = "storage"
	PermissionModuleTokens   = "tokens"
	PermissionModuleTeam     = "team"

	PermissionActionRead  = "read"
	PermissionActionWrite = "write"

	MemberRoleAdmin  = "ADMIN"
	MemberRoleEditor = "EDITOR"
	MemberRoleViewer = "VIEWER"
)
