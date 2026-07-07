package api

import "time"

type CreateTokenRequest struct {
	Identifier     string `json:"identifier" validate:"required,min=1,max=128"`
	Type           string `json:"type"`
	OrganizationID string `json:"omitzero"`
	// Scopes optionally restricts the token to a set of "module:action"
	// capabilities (e.g. ["storage:write","logs:write"]). Empty means full access.
	Scopes []string `json:"scopes" validate:"omitempty,max=16,dive,min=1,max=64"`
	// ExpiresAt optionally sets when the token stops working. Nil means it never
	// expires.
	ExpiresAt *time.Time `json:"expiresAt"`
}

type CreateTokenResponse struct {
	Token string `json:"token"`
}

type ListTokensResponse struct {
	ID         int64      `json:"id"`
	Identifier string     `json:"identifier"`
	Scopes     []string   `json:"scopes"`
	ExpiresAt  *time.Time `json:"expiresAt"`
}
