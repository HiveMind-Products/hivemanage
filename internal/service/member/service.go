package member

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/fivemanage/lite/api"
	"github.com/fivemanage/lite/internal/crypt"
	"github.com/fivemanage/lite/internal/database"
	organizationquery "github.com/fivemanage/lite/internal/database/query/organization"
	"github.com/fivemanage/lite/internal/permissions"
	"github.com/uptrace/bun"
)

var ErrLastAdmin = errors.New("organization must keep at least one admin")

type Service struct {
	db *bun.DB
}

func NewService(db *bun.DB) *Service {
	return &Service{db: db}
}

func (s *Service) ListMembers(ctx context.Context, organizationID string) ([]*api.OrganizationMember, error) {
	dbMembers, err := organizationquery.ListMembers(ctx, s.db, organizationID)
	if err != nil {
		return nil, err
	}

	members := make([]*api.OrganizationMember, 0, len(dbMembers))
	for _, dbMember := range dbMembers {
		role, memberPermissions, err := permissions.Normalize(dbMember.Role, dbMember.Permissions)
		if err != nil {
			role = api.MemberRoleViewer
			memberPermissions = permissions.Preset(role)
		}
		member := &api.OrganizationMember{
			ID:          dbMember.ID,
			Role:        role,
			Permissions: memberPermissions,
		}

		if dbMember.User != nil {
			member.Email = dbMember.User.Email
			member.Name = dbMember.User.Name
		}

		members = append(members, member)
	}

	return members, nil
}

func (s *Service) AddMember(ctx context.Context, organizationID string, req *api.CreateMemberRequest) (*api.CreateMemberResponse, error) {
	role, memberPermissions, err := permissions.Normalize(req.Role, req.Permissions)
	if err != nil {
		return nil, err
	}

	existing := new(database.User)
	err = s.db.NewSelect().Model(existing).Where("username = ?", req.Username).Scan(ctx)
	if err == nil && existing.ID != 0 {
		return nil, fmt.Errorf("username %q is already taken", req.Username)
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	password, err := crypt.GeneratePassword()
	if err != nil {
		return nil, fmt.Errorf("failed to generate password: %w", err)
	}

	hash, err := crypt.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &database.User{
		Username:     req.Username,
		Email:        req.Email,
		Name:         req.Username,
		PasswordHash: hash,
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.NewInsert().Model(user).Returning("id").Exec(ctx); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	member := &database.OrganizationMember{
		UserID:         user.ID,
		OrganizationID: organizationID,
		Role:           role,
		Permissions:    memberPermissions,
	}
	if _, err := tx.NewInsert().Model(member).Exec(ctx); err != nil {
		return nil, fmt.Errorf("failed to add member to organization: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit member creation: %w", err)
	}

	return &api.CreateMemberResponse{Username: req.Username, Password: password}, nil
}

func (s *Service) UpdateMember(ctx context.Context, organizationID string, memberID int64, req *api.UpdateMemberRequest) (*api.OrganizationMember, error) {
	current, err := organizationquery.FindMember(ctx, s.db, organizationID, memberID)
	if err != nil {
		return nil, err
	}

	role, memberPermissions, err := permissions.Normalize(req.Role, req.Permissions)
	if err != nil {
		return nil, err
	}

	currentRole, err := permissions.NormalizeRole(current.Role)
	if err != nil {
		currentRole = api.MemberRoleViewer
	}
	if currentRole == api.MemberRoleAdmin && role != api.MemberRoleAdmin {
		if err := s.ensureNotLastAdmin(ctx, organizationID); err != nil {
			return nil, err
		}
	}

	if err := organizationquery.UpdateMember(ctx, s.db, organizationID, memberID, role, memberPermissions); err != nil {
		return nil, err
	}

	updated, err := organizationquery.FindMember(ctx, s.db, organizationID, memberID)
	if err != nil {
		return nil, err
	}
	return memberResponse(updated), nil
}

func (s *Service) RemoveMember(ctx context.Context, organizationID string, memberID int64) error {
	member, err := organizationquery.FindMember(ctx, s.db, organizationID, memberID)
	if err != nil {
		return err
	}
	memberRole, err := permissions.NormalizeRole(member.Role)
	if err != nil {
		memberRole = api.MemberRoleViewer
	}
	if memberRole == api.MemberRoleAdmin {
		if err := s.ensureNotLastAdmin(ctx, organizationID); err != nil {
			return err
		}
	}
	return organizationquery.DeleteMember(ctx, s.db, memberID, organizationID)
}

func (s *Service) ensureNotLastAdmin(ctx context.Context, organizationID string) error {
	adminCount, err := organizationquery.CountAdmins(ctx, s.db, organizationID)
	if err != nil {
		return err
	}
	if adminCount <= 1 {
		return ErrLastAdmin
	}
	return nil
}

func memberResponse(dbMember *database.OrganizationMember) *api.OrganizationMember {
	role, memberPermissions, err := permissions.Normalize(dbMember.Role, dbMember.Permissions)
	if err != nil {
		role = api.MemberRoleViewer
		memberPermissions = permissions.Preset(role)
	}
	member := &api.OrganizationMember{ID: dbMember.ID, Role: role, Permissions: memberPermissions}
	if dbMember.User != nil {
		member.Email = dbMember.User.Email
		member.Name = dbMember.User.Name
	}
	return member
}
