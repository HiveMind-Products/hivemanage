package api

type OrganizationMember struct {
	ID          int64             `json:"id"`
	Role        string            `json:"role"`
	Email       string            `json:"email"`
	Name        string            `json:"name"`
	Permissions MemberPermissions `json:"permissions"`
}
