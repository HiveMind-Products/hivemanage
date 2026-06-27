package organization

import (
	"context"

	"github.com/fivemanage/lite/api"
	"github.com/fivemanage/lite/internal/clickhouse"
	"github.com/fivemanage/lite/internal/crypt"
	"github.com/fivemanage/lite/internal/database"
	datasetquery "github.com/fivemanage/lite/internal/database/query/dataset"
	filequery "github.com/fivemanage/lite/internal/database/query/file"
	organizationquery "github.com/fivemanage/lite/internal/database/query/organization"
	tokenquery "github.com/fivemanage/lite/internal/database/query/token"
	"github.com/fivemanage/lite/internal/permissions"
	"github.com/uptrace/bun"
)

type Service struct {
	db *bun.DB
	ch *clickhouse.Client
}

func NewService(db *bun.DB, ch *clickhouse.Client) *Service {
	return &Service{
		db: db,
		ch: ch,
	}
}

func (r *Service) GetStats(ctx context.Context, organizationID string) (*api.OrganizationStats, error) {
	totalFiles, err := filequery.FindTotalStorageCount(ctx, r.db, organizationID)
	if err != nil {
		return nil, err
	}

	totalSize, err := filequery.FindTotalStorageSize(ctx, r.db, organizationID)
	if err != nil {
		return nil, err
	}

	totalLogs, err := r.ch.QueryTotalLogsByOrg(ctx, organizationID)
	if err != nil {
		return nil, err
	}

	tokens, err := tokenquery.List(ctx, r.db, organizationID)
	if err != nil {
		return nil, err
	}

	datasets, err := datasetquery.List(ctx, r.db, organizationID)
	if err != nil {
		return nil, err
	}

	return &api.OrganizationStats{
		TotalFiles:   totalFiles,
		TotalSize:    totalSize,
		TotalLogs:    totalLogs,
		TotalTokens:  len(tokens),
		TotalDataset: len(datasets),
	}, nil
}

func (r *Service) CreateOrganization(ctx context.Context, data *api.CreateOrganizationRequest, userID int64) (*api.Organization, error) {
	orgId, err := crypt.GeneratePrimaryKey()
	if err != nil {
		return nil, err
	}

	dbOrganization := &database.Organization{
		Name: data.Name,
		ID:   orgId,
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.NewInsert().Model(dbOrganization).Exec(ctx); err != nil {
		return nil, err
	}

	if _, err := tx.NewInsert().Model(&database.OrganizationMember{
		Role:           api.MemberRoleAdmin,
		Permissions:    permissions.Preset(api.MemberRoleAdmin),
		OrganizationID: dbOrganization.ID,
		UserID:         userID,
	}).Exec(ctx); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	organization := &api.Organization{
		ID:   dbOrganization.ID,
		Name: dbOrganization.Name,
	}

	return organization, nil
}

func (r *Service) UpdateOrganization(ctx context.Context, ID, name string) (*api.Organization, error) {
	if _, err := r.db.NewUpdate().
		Model((*database.Organization)(nil)).
		Set("name = ?", name).
		Where("id = ?", ID).
		Exec(ctx); err != nil {
		return nil, err
	}

	return &api.Organization{ID: ID, Name: name}, nil
}

// DeleteOrganization removes an organization and all of its relational data
// (members, tokens, datasets, invites, asset records) in a single transaction.
// Note: it does not delete the underlying S3 objects or ClickHouse log rows;
// ClickHouse data ages out via per-dataset TTL.
func (r *Service) DeleteOrganization(ctx context.Context, ID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	models := []interface{}{
		(*database.Invite)(nil),
		(*database.Dataset)(nil),
		(*database.Token)(nil),
		(*database.Asset)(nil),
		(*database.OrganizationMember)(nil),
	}
	for _, model := range models {
		if _, err := tx.NewDelete().Model(model).Where("organization_id = ?", ID).Exec(ctx); err != nil {
			return err
		}
	}

	if _, err := tx.NewDelete().Model((*database.Organization)(nil)).Where("id = ?", ID).Exec(ctx); err != nil {
		return err
	}

	return tx.Commit()
}

// GetUsage returns a detailed storage and logging breakdown for an organization.
// Logging figures are best-effort: if ClickHouse is unavailable, the log fields
// are left empty and LoggingAvailable is false.
func (r *Service) GetUsage(ctx context.Context, organizationID string) (*api.OrganizationUsage, error) {
	totalFiles, err := filequery.FindTotalStorageCount(ctx, r.db, organizationID)
	if err != nil {
		return nil, err
	}

	totalSize, err := filequery.FindTotalStorageSize(ctx, r.db, organizationID)
	if err != nil {
		return nil, err
	}

	categories, err := filequery.FindStorageByCategory(ctx, r.db, organizationID)
	if err != nil {
		return nil, err
	}

	largest, err := filequery.FindLargestFiles(ctx, r.db, organizationID, 5)
	if err != nil {
		return nil, err
	}
	largestFiles := make([]api.LargestFile, 0, len(largest))
	for _, asset := range largest {
		largestFiles = append(largestFiles, api.LargestFile{
			ID:   asset.ID,
			Name: asset.OriginalName,
			Type: asset.Type,
			Size: asset.Size,
		})
	}

	usage := &api.OrganizationUsage{
		TotalFiles:        totalFiles,
		TotalSize:         totalSize,
		StorageByCategory: categories,
		LargestFiles:      largestFiles,
		LoggingAvailable:  true,
	}

	datasets, err := datasetquery.List(ctx, r.db, organizationID)
	if err != nil {
		return nil, err
	}
	datasetNames := make(map[string]string, len(datasets))
	for _, ds := range datasets {
		datasetNames[ds.ID] = ds.Name
	}

	counts, err := r.ch.QueryLogCountsByDataset(ctx, organizationID)
	if err != nil {
		// Logging is optional; degrade gracefully when ClickHouse is down.
		usage.LoggingAvailable = false
		return usage, nil
	}

	total := 0
	logsByDataset := make([]api.DatasetLogCount, 0, len(counts))
	for datasetID, count := range counts {
		name := datasetNames[datasetID]
		if name == "" {
			name = datasetID
		}
		logsByDataset = append(logsByDataset, api.DatasetLogCount{
			DatasetID: datasetID,
			Name:      name,
			Count:     count,
		})
		total += count
	}
	usage.TotalLogs = total
	usage.LogsByDataset = logsByDataset

	if series, err := r.ch.QueryLogTimeseriesByOrg(ctx, organizationID, 14); err == nil {
		usage.LogsTimeseries = series
	}

	return usage, nil
}

func (r *Service) FindOrganizationByID(ctx context.Context, ID string) (*api.Organization, error) {
	dbOrganization, err := organizationquery.Find(ctx, r.db, ID)
	if err != nil {
		return nil, err
	}

	organization := &api.Organization{
		ID:   dbOrganization.ID,
		Name: dbOrganization.Name,
	}

	return organization, nil
}

func (r *Service) ListOrganizations(ctx context.Context, userID int64) ([]*api.Organization, error) {
	dbOrganizations, err := organizationquery.ListByUser(ctx, r.db, userID)
	if err != nil {
		return nil, err
	}

	organizations := make([]*api.Organization, 0, len(dbOrganizations))
	for _, dbOrganization := range dbOrganizations {
		organization := &api.Organization{
			ID:   dbOrganization.ID,
			Name: dbOrganization.Name,
		}

		organizations = append(organizations, organization)
	}

	return organizations, nil
}
