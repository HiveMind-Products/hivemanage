package member

import (
	"context"
	"fmt"

	"github.com/fivemanage/lite/api"
	"github.com/fivemanage/lite/internal/crypt"
	"github.com/fivemanage/lite/internal/database"
	organizationquery "github.com/fivemanage/lite/internal/database/query/organization"
	"github.com/uptrace/bun"
)

type Service struct {
	db *bun.DB
}

func NewService(db *bun.DB) *Service {
	return &Service{
		db: db,
	}
}

func (s *Service) ListMembers(ctx context.Context, organizationID string) ([]*api.OrganizationMember, error) {
	dbMembers, err := organizationquery.ListMembers(ctx, s.db, organizationID)
	if err != nil {
		return nil, err
	}

	members := make([]*api.OrganizationMember, 0, len(dbMembers))
	for _, dbMember := range dbMembers {
		member := &api.OrganizationMember{
			ID:   dbMember.ID,
			Role: dbMember.Role,
		}

		if dbMember.User != nil {
			member.Email = dbMember.User.Email
			member.Name = dbMember.User.Name
		}

		members = append(members, member)
	}

	return members, nil
}

// AddMember creates a new local user account and adds them as a member of the organization.
// Returns the generated credentials so the admin can share them.
func (s *Service) AddMember(ctx context.Context, organizationID string, req *api.CreateMemberRequest) (*api.CreateMemberResponse, error) {
	// check username is not already taken
	existing := new(database.User)
	err := s.db.NewSelect().Model(existing).Where("username = ?", req.Username).Scan(ctx)
	if err == nil && existing.ID != 0 {
		return nil, fmt.Errorf("username '%s' is already taken", req.Username)
	}

	// generate a random password
	password, err := crypt.GeneratePassword()
	if err != nil {
		return nil, fmt.Errorf("failed to generate password: %w", err)
	}

	hash, err := crypt.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &database.User{
		Username: req.Username,
		Email:    req.Email,
		Name:     req.Username,
		PasswordHash: hash,
	}

	res, err := s.db.NewInsert().Model(user).Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	userID, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get user ID: %w", err)
	}

	member := &database.OrganizationMember{
		UserID:         userID,
		OrganizationID: organizationID,
		Role:           "member",
	}

	tx, err := organizationquery.CreateMember(ctx, s.db, member)
	if err != nil {
		return nil, fmt.Errorf("failed to add member to organization: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit member creation: %w", err)
	}

	return &api.CreateMemberResponse{
		Username: req.Username,
		Password: password,
	}, nil
}

// RemoveMember removes a member from the organization by their membership ID.
func (s *Service) RemoveMember(ctx context.Context, organizationID string, memberID int64) error {
	return organizationquery.DeleteMember(ctx, s.db, memberID, organizationID)
}
