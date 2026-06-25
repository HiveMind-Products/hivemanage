package file

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"mime/multipart"
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
	db      *bun.DB
	storage *storages3.Storage
}

func NewService(db *bun.DB, storageLayer *storages3.Storage) *Service {
	return &Service{db: db, storage: storageLayer}
}

func (s *Service) CreateFile(ctx context.Context, organizationID string, file multipart.File, fileHeader *multipart.FileHeader) (string, error) {
	asset, err := s.createAsset(ctx, organizationID, file, fileHeader)
	if err != nil {
		return "", err
	}
	return asset.Key, nil
}
func (s *Service) CreateStorageFile(ctx context.Context, organizationID string, file multipart.File, fileHeader *multipart.FileHeader) error {
	_, err := s.createAsset(ctx, organizationID, file, fileHeader)
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

func (s *Service) createAsset(ctx context.Context, organizationID string, file multipart.File, fileHeader *multipart.FileHeader) (*database.Asset, error) {
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
	key, err := generateFileKey(organizationID, ext)
	if err != nil {
		return nil, UploadStorageError{ErrorMsg: errors.New("failed to generate file key").Error()}
	}
	asset := &database.Asset{ID: primaryKey, Type: fileType, Size: fileHeader.Size, OrganizationID: organizationID, Key: key, OriginalName: fileHeader.Filename}
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
