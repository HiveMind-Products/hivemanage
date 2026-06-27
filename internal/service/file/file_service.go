package file

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"mime/multipart"
	"strings"
	"time"

	"github.com/fivemanage/lite/api"
	"github.com/fivemanage/lite/internal/crypt"
	"github.com/fivemanage/lite/internal/database"
	filequery "github.com/fivemanage/lite/internal/database/query/file"
	"github.com/fivemanage/lite/internal/http/httputil"
	storages3 "github.com/fivemanage/lite/pkg/storage/s3"
	"github.com/uptrace/bun"
)

const (
	MaxUploadSize   int64         = 500 * 1024 * 1024
	SignedURLExpiry time.Duration = 15 * time.Minute
)

type Service struct {
	db           *bun.DB
	storage      *storages3.Storage
	bucketDomain string
}

func NewService(db *bun.DB, storageLayer *storages3.Storage, bucketDomain string) *Service {
	return &Service{db: db, storage: storageLayer, bucketDomain: bucketDomain}
}

func (s *Service) CreateFile(ctx context.Context, organizationID string, file multipart.File, fileHeader *multipart.FileHeader) (string, error) {
	asset, err := s.createAsset(ctx, organizationID, file, fileHeader, CreateFileV3Params{})
	if err != nil {
		return "", err
	}
	return asset.Key, nil
}
func (s *Service) CreateStorageFile(ctx context.Context, organizationID string, file multipart.File, fileHeader *multipart.FileHeader) error {
	_, err := s.createAsset(ctx, organizationID, file, fileHeader, CreateFileV3Params{})
	return err
}
func (s *Service) SignedURLByKey(ctx context.Context, key string) (string, error) {
	return s.storage.SignedURL(ctx, key, SignedURLExpiry)
}
func (s *Service) SignedURL(ctx context.Context, organizationID, fileID string) (string, error) {
	asset, err := filequery.FindFileByID(ctx, s.db, organizationID, fileID)
	if err != nil {
		return "", err
	}
	if asset == nil {
		return "", GetFileError{ErrorMsg: "file not found"}
	}
	return s.storage.SignedURL(ctx, asset.Key, SignedURLExpiry)
}
func (s *Service) DeleteStorageFile(ctx context.Context, organizationID, fileID string) error {
	asset, err := filequery.FindFileByID(ctx, s.db, organizationID, fileID)
	if err != nil {
		return err
	}
	if asset == nil {
		return GetFileError{ErrorMsg: "file not found"}
	}
	if err := s.storage.DeleteFile(ctx, asset.Key); err != nil {
		return err
	}
	return filequery.Delete(ctx, s.db, organizationID, fileID)
}

// CreateFileV3Params carries the optional V3 upload fields.
type CreateFileV3Params struct {
	Filename        string
	Path            string
	Metadata        map[string]any
	RetentionExempt bool
}

// CreateFileV3 stores a file with the optional V3 fields and returns the V3 view
// of the asset (id, filename, type, size, url, originalUrl, metadata).
func (s *Service) CreateFileV3(ctx context.Context, organizationID string, file multipart.File, fileHeader *multipart.FileHeader, params CreateFileV3Params) (*api.FileItemV3, error) {
	asset, err := s.createAsset(ctx, organizationID, file, fileHeader, params)
	if err != nil {
		return nil, err
	}
	return s.assetToFileItem(asset), nil
}

func (s *Service) createAsset(ctx context.Context, organizationID string, file multipart.File, fileHeader *multipart.FileHeader, params CreateFileV3Params) (*database.Asset, error) {
	if fileHeader.Size <= 0 || fileHeader.Size > MaxUploadSize {
		return nil, UploadStorageError{ErrorMsg: fmt.Sprintf("file size must be between 1 and %d bytes", MaxUploadSize)}
	}
	primaryKey, err := crypt.GeneratePrimaryKey()
	if err != nil {
		return nil, UploadStorageError{ErrorMsg: err.Error()}
	}
	mimeType, ext, fileType, err := httputil.GetMimeDetails(fileHeader, file)
	if err != nil {
		return nil, UploadStorageError{ErrorMsg: errors.New("failed to get mime type").Error()}
	}
	key, err := generateFileKeyV3(organizationID, params.Path, params.Filename, ext)
	if err != nil {
		return nil, UploadStorageError{ErrorMsg: errors.New("failed to generate file key").Error()}
	}
	originalName := fileHeader.Filename
	if params.Filename != "" {
		originalName = params.Filename
	}
	asset := &database.Asset{
		ID:              primaryKey,
		Type:            fileType,
		Size:            fileHeader.Size,
		OrganizationID:  organizationID,
		Key:             key,
		OriginalName:    originalName,
		Metadata:        params.Metadata,
		RetentionExempt: params.RetentionExempt,
	}
	tx, err := filequery.Create(ctx, s.db, asset)
	if err != nil {
		return nil, err
	}
	if _, err := file.Seek(0, 0); err != nil {
		_ = tx.Rollback()
		return nil, UploadStorageError{ErrorMsg: err.Error()}
	}
	if err := s.storage.UploadFile(ctx, file, key, mimeType); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			slog.Error("FileService.createAsset rollback", "organization_id", organizationID, "err", rbErr)
		}
		return nil, UploadStorageError{ErrorMsg: err.Error()}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return asset, nil
}

// buildFileURLs returns (url, originalUrl) for a stored object. originalUrl is the
// default storage URL; url uses the custom CDN domain (BUCKET_DOMAIN) when set.
// Both fall back to a signed URL if no public base is configured.
func (s *Service) buildFileURLs(ctx context.Context, key string) (string, string) {
	originalURL := s.storage.PublicURL(key)
	var url string
	if s.bucketDomain != "" {
		url = strings.TrimRight(s.bucketDomain, "/") + "/" + key
	}
	if url == "" {
		url = originalURL
	}
	if originalURL == "" {
		originalURL = url
	}
	if url == "" {
		// Last resort so the response is still usable for private buckets with no
		// configured public base.
		if signed, err := s.storage.SignedURL(ctx, key, SignedURLExpiry); err == nil {
			url, originalURL = signed, signed
		}
	}
	return url, originalURL
}

func (s *Service) assetToFileItem(asset *database.Asset) *api.FileItemV3 {
	url, originalURL := s.buildFileURLs(context.Background(), asset.Key)
	return &api.FileItemV3{
		ID:          asset.ID,
		Filename:    asset.OriginalName,
		Type:        asset.Type,
		Size:        asset.Size,
		URL:         url,
		OriginalURL: originalURL,
		Metadata:    asset.Metadata,
	}
}

func (s *Service) ListStorageFiles(ctx context.Context, organizationID string, search string, fileType string, page int, pageSize int) (*api.AssetResponse, error) {
	files, err := filequery.FindStorageFiles(ctx, s.db, organizationID, search, fileType, page, pageSize)
	if err != nil {
		storageError := &ListStorageError{ErrorMsg: err.Error()}
		slog.Error("FileService.ListStorageFiles", "organization_id", organizationID, "err", storageError)
		return nil, storageError
	}
	assets := make([]*api.Asset, len(files))
	for i, file := range files {
		assets[i] = &api.Asset{ID: file.ID, Type: file.Type, Key: file.Key, OriginalName: file.OriginalName, Size: file.Size, CreatedAt: file.CreatedAt}
	}
	totalCount, err := filequery.FindTotalStorageCount(ctx, s.db, organizationID)
	if err != nil {
		storageError := &ListStorageError{ErrorMsg: err.Error()}
		slog.Error("FileService.ListStorageFiles", "organization_id", organizationID, "err", storageError)
		return nil, storageError
	}
	return &api.AssetResponse{StorageFiles: assets, TotalCount: totalCount}, nil
}

func (s *Service) GetStorageFile(ctx context.Context, organizationID string, fileID string) (*api.Asset, error) {
	file, err := filequery.FindFileByID(ctx, s.db, organizationID, fileID)
	if err != nil {
		storageError := &GetFileError{ErrorMsg: err.Error()}
		slog.Error("FileService.GetStorageFile", "organization_id", organizationID, "err", storageError)
		return nil, storageError
	}
	if file == nil {
		return nil, &GetFileError{ErrorMsg: "file not found"}
	}
	return &api.Asset{ID: file.ID, Type: file.Type, Key: file.Key, OriginalName: file.OriginalName, Size: file.Size, CreatedAt: file.CreatedAt}, nil
}
