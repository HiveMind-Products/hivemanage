package organizationquery

import (
	"context"

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

func ListMembers(ctx context.Context, db *bun.DB, organizationID string) ([]database.OrganizationMember, error) {
	var members []database.OrganizationMember

	err := db.NewSelect().
		Model(&members).
		Relation("User").
		Where("organization_id = ?", organizationID).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return members, nil
}
