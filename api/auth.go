package api

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type UpdateProfileRequest struct {
	Name   *string `json:"name" validate:"omitempty,max=128"`
	Avatar *string `json:"avatar" validate:"omitempty,max=512"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword" validate:"required,min=8,max=128"`
}

type User struct {
	ID                        int64                        `json:"id"`
	Name                      *string                      `json:"name"`
	Username                  string                       `json:"username"`
	Email                     *string                      `json:"email"`
	Avatar                    *string                      `json:"avatar"`
	IsAdmin                   bool                         `json:"isAdmin"`
	DiscordLinked             bool                         `json:"discordLinked"`
	PermissionsByOrganization map[string]MemberPermissions `json:"permissionsByOrganization"`
}
