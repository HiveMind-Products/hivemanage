package tokenquery

import (
	"context"

	"github.com/fivemanage/lite/internal/database"
	"github.com/uptrace/bun"
)

func SelectByHash(ctx context.Context, db *bun.DB, tokenHash string) (*database.Token, error) {
	var token database.Token

	err := db.NewSelect().Model(&token).Where("token_hash = ?", tokenHash).Scan(ctx)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func Create(ctx context.Context, db *bun.DB, token *database.Token) error {
	_, err := db.NewInsert().Model(token).Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func List(ctx context.Context, db *bun.DB, organizationID string) ([]database.Token, error) {
	var tokens []database.Token

	err := db.NewSelect().
		Model(&tokens).
		Where("organization_id = ?", organizationID).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func Delete(ctx context.Context, db *bun.DB, organizationID string, tokenID int64) error {
	_, err := db.NewDelete().Model((*database.Token)(nil)).Where("organization_id = ?", organizationID).Where("id = ?", tokenID).Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}
