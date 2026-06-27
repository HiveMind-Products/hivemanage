package file

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"mime/multipart"
	"strings"

	"github.com/fivemanage/lite/api"
	"github.com/fivemanage/lite/internal/database"
	filequery "github.com/fivemanage/lite/internal/database/query/file"
)

// DefaultListLimit and MaxListLimit bound the V3 list endpoint page size.
const (
	DefaultListLimit = 50
	MaxListLimit     = 100
)

// memoryFile adapts an in-memory byte slice to the multipart.File interface so
// the base64 upload path can reuse the shared create logic.
type memoryFile struct {
	*bytes.Reader
}

func (memoryFile) Close() error { return nil }

// CreateFileBase64V3 decodes a (optionally data-URI prefixed) base64 payload and
// stores it, returning the V3 file item.
func (s *Service) CreateFileBase64V3(ctx context.Context, organizationID string, req api.UploadFileBase64Request) (*api.FileItemV3, error) {
	raw, err := DecodeBase64Payload(req.Base64)
	if err != nil {
		return nil, UploadStorageError{ErrorMsg: err.Error()}
	}
	metadata, err := ParseMetadataJSON(req.Metadata)
	if err != nil {
		return nil, UploadStorageError{ErrorMsg: "invalid metadata: " + err.Error()}
	}

	mf := memoryFile{bytes.NewReader(raw)}
	header := &multipart.FileHeader{Filename: req.Filename, Size: int64(len(raw))}
	asset, err := s.createAsset(ctx, organizationID, mf, header, CreateFileV3Params{
		Filename:        req.Filename,
		Path:            req.Path,
		Metadata:        metadata,
		RetentionExempt: req.RetentionExempt,
	})
	if err != nil {
		return nil, err
	}
	return s.assetToFileItem(ctx, asset), nil
}

// NormalizeListParams clamps the V3 list pagination inputs and returns the
// effective page (>=1), limit (1..MaxListLimit) and zero-based offset.
func NormalizeListParams(page, limit int) (int, int, int) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = DefaultListLimit
	}
	if limit > MaxListLimit {
		limit = MaxListLimit
	}
	return page, limit, (page - 1) * limit
}

// ListFilesV3 returns a page of files with the V3 type/folder filters plus the
// total count for pagination. page is 1-based.
func (s *Service) ListFilesV3(ctx context.Context, organizationID, fileType, folder string, page, limit int) ([]*api.FileItemV3, int, error) {
	page, limit, offset := NormalizeListParams(page, limit)
	folder = sanitizeFolder(folder)

	files, err := filequery.FindFilesV3(ctx, s.db, organizationID, fileType, folder, offset, limit)
	if err != nil {
		return nil, 0, &ListStorageError{ErrorMsg: err.Error()}
	}
	total, err := filequery.CountFilesV3(ctx, s.db, organizationID, fileType, folder)
	if err != nil {
		return nil, 0, &ListStorageError{ErrorMsg: err.Error()}
	}

	items := make([]*api.FileItemV3, 0, len(files))
	for _, f := range files {
		items = append(items, s.assetToFileItem(ctx, f))
	}
	return items, total, nil
}

// GetFileV3 resolves a file by its id or its storage key.
func (s *Service) GetFileV3(ctx context.Context, organizationID, idOrKey string) (*api.FileItemV3, error) {
	asset, err := s.findByIDOrKey(ctx, organizationID, idOrKey)
	if err != nil {
		return nil, err
	}
	return s.assetToFileItem(ctx, asset), nil
}

// DeleteFileV3 deletes a file (storage object + DB record) by id or storage key.
func (s *Service) DeleteFileV3(ctx context.Context, organizationID, idOrKey string) error {
	asset, err := s.findByIDOrKey(ctx, organizationID, idOrKey)
	if err != nil {
		return err
	}
	if err := s.storage.DeleteFile(ctx, asset.Key); err != nil {
		return err
	}
	return filequery.Delete(ctx, s.db, organizationID, asset.ID)
}

func (s *Service) findByIDOrKey(ctx context.Context, organizationID, idOrKey string) (*database.Asset, error) {
	asset, err := filequery.FindFileByID(ctx, s.db, organizationID, idOrKey)
	if err != nil {
		return nil, &GetFileError{ErrorMsg: err.Error()}
	}
	if asset == nil {
		asset, err = filequery.FindFileByKey(ctx, s.db, organizationID, idOrKey)
		if err != nil {
			return nil, &GetFileError{ErrorMsg: err.Error()}
		}
	}
	if asset == nil {
		return nil, &GetFileError{ErrorMsg: "file not found"}
	}
	return asset, nil
}

// DecodeBase64Payload decodes a base64 string, transparently stripping a data-URI
// prefix (e.g. "data:image/png;base64,") when present. It accepts both standard
// and URL-safe alphabets, with or without padding.
func DecodeBase64Payload(input string) ([]byte, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, errors.New("base64 payload is empty")
	}
	if strings.HasPrefix(input, "data:") {
		comma := strings.IndexByte(input, ',')
		if comma < 0 {
			return nil, errors.New("invalid data URI: missing comma")
		}
		header := input[:comma]
		if !strings.Contains(header, ";base64") {
			return nil, errors.New("invalid data URI: not base64 encoded")
		}
		input = input[comma+1:]
	}
	input = strings.TrimSpace(input)

	for _, enc := range []*base64.Encoding{base64.StdEncoding, base64.RawStdEncoding, base64.URLEncoding, base64.RawURLEncoding} {
		if data, err := enc.DecodeString(input); err == nil {
			if len(data) == 0 {
				return nil, errors.New("decoded payload is empty")
			}
			return data, nil
		}
	}
	return nil, errors.New("invalid base64 payload")
}

// ParseMetadataJSON parses an optional JSON-object metadata string. An empty
// string yields a nil map (no metadata).
func ParseMetadataJSON(raw string) (map[string]any, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, err
	}
	return out, nil
}
