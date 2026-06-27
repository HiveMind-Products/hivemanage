package api

import "time"

// CreateInviteRequest is the body for creating a Discord invite.
type CreateInviteRequest struct {
	Role            string            `json:"role"`
	Permissions     MemberPermissions `json:"permissions"`
	DiscordUsername string            `json:"discordUsername"`
	Email           string            `json:"email"`
}

// InviteResponse is the public shape of an invite, including its shareable link.
type InviteResponse struct {
	ID              string     `json:"id"`
	URL             string     `json:"url"`
	OrganizationID  string     `json:"organizationId"`
	Role            string     `json:"role"`
	DiscordUsername string     `json:"discordUsername"`
	Email           string     `json:"email"`
	ExpiresAt       *time.Time `json:"expiresAt"`
	AcceptedAt      *time.Time `json:"acceptedAt"`
	CreatedAt       time.Time  `json:"createdAt"`
}
