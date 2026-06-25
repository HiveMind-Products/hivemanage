package api

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type User struct {
	ID                        int64                        `json:"id"`
	Name                      *string                      `json:"name"`
	Username                  string                       `json:"username"`
	Email                     *string                      `json:"email"`
	Avatar                    *string                      `json:"avatar"`
	IsAdmin                   bool                         `json:"isAdmin"`
	PermissionsByOrganization map[string]MemberPermissions `json:"permissionsByOrganization"`
}
