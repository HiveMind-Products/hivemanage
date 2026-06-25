package api

type CreateMemberRequest struct {
	Username    string            `json:"username" validate:"required"`
	Email       string            `json:"email"`
	Role        string            `json:"role"`
	Permissions MemberPermissions `json:"permissions"`
}

// we should probably, probably probably use protobufs this point
type CreateMemberResponse struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UpdateMemberRequest struct {
	Role        string            `json:"role"`
	Permissions MemberPermissions `json:"permissions"`
}
