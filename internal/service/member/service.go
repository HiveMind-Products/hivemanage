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
	tokenquery "github.com/fivemanage/lite/internal/database/query/token"
	"github.com/fivemanage/lite/internal/permissions"
	"github.com/uptrace/bun"
)

var (
	ErrLastAdmin = errors.New("organization must keep at least one admin")
	// ErrEscalation is returned when a non-admin actor attempts to grant the admin
	// role (which carries all permissions) — preventing privilege escalation via
	// the team:write permission.
	ErrEscalation = errors.New("only an organization admin can assign the admin role")
)

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
		members = append(members, memberResponse(&dbMember))
	}

	return members, nil
}

func (s *Service) AddMember(ctx context.Context, organizationID string, req *api.CreateMemberRequest, actorIsAdmin bool) (*api.CreateMemberResponse, error) {
	role, memberPermissions, err := permissions.Normalize(req.Role, req.Permissions)
	if err != nil {
		return nil, err
	}

	if role == api.MemberRoleAdmin && !actorIsAdmin {
		return nil, ErrEscalation
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

func (s *Service) UpdateMember(ctx context.Context, organizationID string, memberID int64, req *api.UpdateMemberRequest, actorIsAdmin bool) (*api.OrganizationMember, error) {
	current, err := organizationquery.FindMember(ctx, s.db, organizationID, memberID)
	if err != nil {
		return nil, err
	}

	role, memberPermissions, err := permissions.Normalize(req.Role, req.Permissions)
	if err != nil {
		return nil, err
	}

	currentRole := permissions.RoleOrViewer(current.Role)
	// A non-admin actor may not promote anyone (including themselves) to admin,
	// nor strip the admin role off an existing admin to take over.
	if !actorIsAdmin && (role == api.MemberRoleAdmin || currentRole == api.MemberRoleAdmin) {
		return nil, ErrEscalation
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
	memberRole := permissions.RoleOrViewer(member.Role)
	if memberRole == api.MemberRoleAdmin {
		if err := s.ensureNotLastAdmin(ctx, organizationID); err != nil {
			return err
		}
	}
	// Revoke the member's API tokens BEFORE removing the membership so the
	// operation is safely retryable: token deletion is idempotent and, if the
	// membership delete later fails, FindMember still resolves on retry. This
	// stops a removed member from retaining org data access (upload/list/read/
	// delete, log ingest) through a still-valid token after offboarding.
	if err := tokenquery.DeleteByUser(ctx, s.db, organizationID, member.UserID); err != nil {
		return err
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
		role = permissions.RoleOrViewer(dbMember.Role)
		memberPermissions = permissions.Preset(role)
	}
	member := &api.OrganizationMember{ID: dbMember.ID, Role: role, Permissions: memberPermissions}
	if dbMember.User != nil {
		member.Email = dbMember.User.Email
		member.Name = dbMember.User.Name
	}
	return member
}
