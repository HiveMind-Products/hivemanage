package organizationquery

import (
	"context"

	"github.com/fivemanage/lite/api"
	"github.com/fivemanage/lite/internal/database"
	"github.com/uptrace/bun"
)

func Create(ctx context.Context, db *bun.DB, organization *database.Organization) (bun.Tx, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return tx, err
	}

	_, err = tx.NewInsert().Model(organization).Exec(ctx)
	if err != nil {
		return tx, err
	}

	return tx, nil
}

func Find(ctx context.Context, db *bun.DB, ID string) (*database.Organization, error) {
	organization := new(database.Organization)
	err := db.NewSelect().Model(organization).Where("id = ?", ID).Scan(ctx)
	if err != nil {
		return nil, err
	}

	return organization, nil
}

func ListByUser(ctx context.Context, db *bun.DB, userID int64) ([]database.Organization, error) {
	var organizations []database.Organization
	err := db.NewSelect().
		Model(&organizations).
		Join("JOIN organization_member AS om ON om.organization_id = organization.id").
		Where("om.user_id = ?", userID).
		Order("organization.name ASC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return organizations, nil
}

func CreateMember(ctx context.Context, db *bun.DB, member *database.OrganizationMember) (bun.Tx, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return tx, err
	}

	_, err = tx.NewInsert().Model(member).Exec(ctx)
	if err != nil {
		return tx, err
	}

	return tx, nil
}

func FindMember(ctx context.Context, db *bun.DB, organizationID string, memberID int64) (*database.OrganizationMember, error) {
	member := new(database.OrganizationMember)
	err := db.NewSelect().
		Model(member).
		Relation("User").
		Where("organization_id = ?", organizationID).
		Where("organization_member.id = ?", memberID).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return member, nil
}

func FindMemberByUser(ctx context.Context, db *bun.DB, organizationID string, userID int64) (*database.OrganizationMember, error) {
	member := new(database.OrganizationMember)
	err := db.NewSelect().
		Model(member).
		Where("organization_id = ?", organizationID).
		Where("user_id = ?", userID).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return member, nil
}

func ListMembers(ctx context.Context, db *bun.DB, organizationID string) ([]database.OrganizationMember, error) {
	var members []database.OrganizationMember

	err := db.NewSelect().
		Model(&members).
		Relation("User").
		Where("organization_id = ?", organizationID).
		Order("organization_member.id ASC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return members, nil
}

func CountAdmins(ctx context.Context, db *bun.DB, organizationID string) (int, error) {
	return db.NewSelect().
		Model((*database.OrganizationMember)(nil)).
		Where("organization_id = ?", organizationID).
		Where("role = ?", "ADMIN").
		Count(ctx)
}

func UpdateMember(ctx context.Context, db *bun.DB, organizationID string, memberID int64, role string, permissions api.MemberPermissions) error {
	member := &database.OrganizationMember{ID: memberID, Role: role, Permissions: permissions}
	_, err := db.NewUpdate().
		Model(member).
		Column("role", "permissions").
		Where("id = ? AND organization_id = ?", memberID, organizationID).
		Exec(ctx)
	return err
}

func DeleteMember(ctx context.Context, db *bun.DB, memberID int64, organizationID string) error {
	_, err := db.NewDelete().
		Model((*database.OrganizationMember)(nil)).
		Where("id = ? AND organization_id = ?", memberID, organizationID).
		Exec(ctx)
	return err
}
