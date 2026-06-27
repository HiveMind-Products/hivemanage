package filequery

import (
	"context"
	"database/sql"
	"errors"

	"github.com/fivemanage/lite/api"
	"github.com/fivemanage/lite/internal/database"
	"github.com/uptrace/bun"
)

// todo: rename to Insert
func Create(ctx context.Context, db *bun.DB, file *database.Asset) (bun.Tx, error) {
	var err error

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return tx, err
	}

	_, err = tx.NewInsert().Model(file).Exec(ctx)
	if err != nil {
		_ = tx.Rollback()
		return tx, err
	}

	return tx, nil
}

func FindFileByID(ctx context.Context, db *bun.DB, organizationID, id string) (*database.Asset, error) {
	var file database.Asset
	err := db.NewSelect().
		Model(&file).
		Where("organization_id = ?", organizationID).
		Where("id = ?", id).
		Limit(1).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &file, nil
}

func FindStorageFiles(ctx context.Context, db *bun.DB, organizationID, search, fileType string, page, pageSize int) ([]*database.Asset, error) {
	var files []*database.Asset

	// no, this does not scale at all, but...soonTM
	// we are also missing indexes I need to create migrations for
	sb := db.NewSelect().
		Model(&files).
		Where("organization_id = ?", organizationID).
		Order("created_at DESC")

	if search != "" {
		// not my proudest moment
		sb = sb.Where("key LIKE ?", "%"+search+"%")
	}

	if fileType != "" && fileType != "all" {
		sb = sb.Where("type = ?", fileType)
	}

	if pageSize > 0 {
		sb = sb.Limit(pageSize).Offset(page * pageSize)
	}

	err := sb.Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return files, nil
}

func FindTotalStorageCount(ctx context.Context, db *bun.DB, organizationID string) (int, error) {
	count, err := db.NewSelect().
		Model((*database.Asset)(nil)).
		Where("organization_id = ?", organizationID).
		Count(ctx)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func FindTotalStorageSize(ctx context.Context, db *bun.DB, organizationID string) (int64, error) {
	var size int64
	err := db.NewSelect().
		Model((*database.Asset)(nil)).
		ColumnExpr("COALESCE(SUM(size), 0)").
		Where("organization_id = ?", organizationID).
		Scan(ctx, &size)
	if err != nil {
		return 0, err
	}

	return size, nil
}

// FindStorageByCategory groups an organization's assets by their top-level MIME
// category (the part before the "/"), returning a count and total size per group.
func FindStorageByCategory(ctx context.Context, db *bun.DB, organizationID string) ([]api.StorageCategory, error) {
	var rows []api.StorageCategory
	err := db.NewSelect().
		Model((*database.Asset)(nil)).
		ColumnExpr("COALESCE(NULLIF(split_part(type, '/', 1), ''), 'other') AS category").
		ColumnExpr("COUNT(*) AS count").
		ColumnExpr("COALESCE(SUM(size), 0) AS size").
		Where("organization_id = ?", organizationID).
		GroupExpr("COALESCE(NULLIF(split_part(type, '/', 1), ''), 'other')").
		OrderExpr("size DESC").
		Scan(ctx, &rows)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// FindLargestFiles returns the largest assets in an organization, newest size first.
func FindLargestFiles(ctx context.Context, db *bun.DB, organizationID string, limit int) ([]*database.Asset, error) {
	var assets []*database.Asset
	err := db.NewSelect().
		Model(&assets).
		Where("organization_id = ?", organizationID).
		Order("size DESC").
		Limit(limit).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return assets, nil
}

func Delete(ctx context.Context, db *bun.DB, organizationID, id string) error {
	_, err := db.NewDelete().Model((*database.Asset)(nil)).Where("organization_id = ?", organizationID).Where("id = ?", id).Exec(ctx)
	return err
}
