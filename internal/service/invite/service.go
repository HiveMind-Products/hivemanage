package invite

import (
	"context"
	"errors"
	"time"

	"github.com/fivemanage/lite/api"
	"github.com/fivemanage/lite/internal/crypt"
	"github.com/fivemanage/lite/internal/database"
	"github.com/fivemanage/lite/internal/permissions"
	"github.com/uptrace/bun"
)

// DefaultInviteTTL is how long an invite link stays valid if no explicit expiry is given.
const DefaultInviteTTL = 7 * 24 * time.Hour

var (
	// ErrInviteInvalid is returned when an invite is missing, expired, or already used.
	ErrInviteInvalid = errors.New("invite is invalid or expired")
	// ErrInviteWrongDiscord is returned when the logged-in Discord account does not
	// match the account the invite was issued for.
	ErrInviteWrongDiscord = errors.New("this invite is for a different Discord account")
)

type Service struct {
	db *bun.DB
}

func NewService(db *bun.DB) *Service {
	return &Service{db: db}
}

// CreateParams describes a new invite.
type CreateParams struct {
	OrganizationID  string
	Role            string
	Permissions     api.MemberPermissions
	DiscordUsername string
	DiscordID       string
	Email           string
	CreatedBy       int64
	TTL             time.Duration
}

// Create issues a new invite and returns it. The invite ID is an opaque,
// unguessable token used in the invite link.
func (s *Service) Create(ctx context.Context, p CreateParams) (*database.Invite, error) {
	role, perms, err := permissions.Normalize(p.Role, p.Permissions)
	if err != nil {
		return nil, err
	}

	token, err := crypt.GenerateSessionID()
	if err != nil {
		return nil, err
	}

	ttl := p.TTL
	if ttl <= 0 {
		ttl = DefaultInviteTTL
	}
	expiresAt := time.Now().Add(ttl)

	invite := &database.Invite{
		ID:              token,
		OrganizationID:  p.OrganizationID,
		Role:            role,
		Permissions:     perms,
		DiscordID:       p.DiscordID,
		DiscordUsername: p.DiscordUsername,
		Email:           p.Email,
		CreatedBy:       p.CreatedBy,
		ExpiresAt:       &expiresAt,
		CreatedAt:       time.Now(),
	}

	if _, err := s.db.NewInsert().Model(invite).Exec(ctx); err != nil {
		return nil, err
	}

	return invite, nil
}

// ListPending returns unaccepted invites for an organization, newest first.
func (s *Service) ListPending(ctx context.Context, organizationID string) ([]*database.Invite, error) {
	var invites []*database.Invite
	err := s.db.NewSelect().
		Model(&invites).
		Where("organization_id = ?", organizationID).
		Where("accepted_at IS NULL").
		Order("created_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return invites, nil
}

// Accept validates the invite token for the given logged-in user/Discord account,
// creates a scoped organization membership, and marks the invite accepted. It is a
// no-op on membership if the user already belongs to the organization.
func (s *Service) Accept(ctx context.Context, token string, userID int64, discordID string) error {
	invite := new(database.Invite)
	err := s.db.NewSelect().Model(invite).Where("id = ?", token).Scan(ctx)
	if err != nil {
		return ErrInviteInvalid
	}
	if invite.AcceptedAt != nil {
		return ErrInviteInvalid
	}
	if invite.ExpiresAt != nil && invite.ExpiresAt.Before(time.Now()) {
		return ErrInviteInvalid
	}
	if invite.DiscordID != "" && invite.DiscordID != discordID {
		return ErrInviteWrongDiscord
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	exists, err := tx.NewSelect().
		Model((*database.OrganizationMember)(nil)).
		Where("organization_id = ?", invite.OrganizationID).
		Where("user_id = ?", userID).
		Exists(ctx)
	if err != nil {
		return err
	}

	if !exists {
		member := &database.OrganizationMember{
			UserID:         userID,
			OrganizationID: invite.OrganizationID,
			Role:           invite.Role,
			Permissions:    invite.Permissions,
		}
		if _, err := tx.NewInsert().Model(member).Exec(ctx); err != nil {
			return err
		}
	}

	now := time.Now()
	invite.AcceptedAt = &now
	invite.AcceptedBy = userID
	invite.DiscordID = discordID
	if _, err := tx.NewUpdate().
		Model(invite).
		Column("accepted_at", "accepted_by", "discord_id").
		WherePK().
		Exec(ctx); err != nil {
		return err
	}

	return tx.Commit()
}
