package log

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"strings"
	"time"

	"github.com/fivemanage/lite/api"
	"github.com/fivemanage/lite/internal/clickhouse"
	"github.com/fivemanage/lite/internal/crypt"
	"github.com/fivemanage/lite/internal/service/dataset"
	"github.com/uptrace/bun"
)

const (
	MaxLogBatchSize        = 1000
	MaxLogMetadataEntries  = 100
	MaxLogMetadataKeyLen   = 128
	MaxLogMetadataValueLen = 4096
)

type Service struct {
	db               *bun.DB
	datasetService   *dataset.Service
	clickhouseClient *clickhouse.Client
}

func NewService(db *bun.DB, clickhouseClient *clickhouse.Client, datasetService *dataset.Service) *Service {
	return &Service{
		db:               db,
		datasetService:   datasetService,
		clickhouseClient: clickhouseClient,
	}
}

func (r *Service) SubmitLogs(ctx context.Context, organizationId string, datasetName string, logs []api.Log) error {
	if len(logs) == 0 {
		return errors.New("log batch is empty")
	}
	if len(logs) > MaxLogBatchSize {
		return fmt.Errorf("log batch exceeds maximum of %d", MaxLogBatchSize)
	}
	clickhouseLogs := make([]*clickhouse.Log, 0, len(logs))

	dataset, err := r.datasetService.FindByName(ctx, organizationId, datasetName)
	if err != nil {
		slog.Error("failed to find dataset", "err", err)
		return err
	}

	for _, log := range logs {
		if err := validateLog(log); err != nil {
			return err
		}

		traceID, err := crypt.GeneratePrimaryKey()
		if err != nil {
			slog.Error("failed to generate trace id", "err", err)
			return err
		}

		timestamp := time.Now().UTC()
		metadata := make(map[string]string)

		if len(log.Metadata) > MaxLogMetadataEntries {
			return fmt.Errorf("metadata exceeds maximum of %d entries", MaxLogMetadataEntries)
		}
		for k, v := range log.Metadata {
			if metadataDepth(v, 1) > MaxLogAttributesDepth {
				return fmt.Errorf("metadata key %q exceeds maximum depth of %d", k, MaxLogAttributesDepth)
			}
			maps.Copy(metadata, buildLogAttributes(k, v, 0))
		}
		if err := validateLogAttributes(metadata); err != nil {
			return err
		}

		// fix log message because it can be silly sometimes
		logMessage := formatMessage(log.Message)

		// these should override any metadata that matches
		metadata["message"] = logMessage
		metadata["severity"] = log.Level
		metadata["_resource"] = log.Resource

		clickhouseLogs = append(clickhouseLogs, &clickhouse.Log{
			TraceID:       traceID,
			Timestamp:     timestamp,
			TeamID:        organizationId,
			DatasetID:     dataset.ID,
			Body:          logMessage,
			Attributes:    metadata,
			RetentionDays: dataset.RetentionDays,
		})
	}

	err = r.clickhouseClient.BatchWriteLogRows(ctx, clickhouseLogs)
	if err != nil {
		slog.Error("failed to submit logs to clickhouse", "err", err)
		return err
	}
	return nil
}

func validateLog(log api.Log) error {
	if strings.TrimSpace(log.Level) == "" {
		return errors.New("log level is required")
	}
	if len(log.Level) > 32 {
		return errors.New("log level exceeds maximum length")
	}
	if strings.TrimSpace(log.Message) == "" {
		return errors.New("log message is required")
	}
	if len(log.Message) > 8192 {
		return errors.New("log message exceeds maximum length")
	}
	if len(log.Resource) > 255 {
		return errors.New("log resource exceeds maximum length")
	}
	return nil
}

func metadataDepth(value any, depth int) int {
	switch typed := value.(type) {
	case map[string]any:
		maxDepth := depth
		for _, nested := range typed {
			if childDepth := metadataDepth(nested, depth+1); childDepth > maxDepth {
				maxDepth = childDepth
			}
		}
		return maxDepth
	case []any:
		maxDepth := depth
		for _, nested := range typed {
			if childDepth := metadataDepth(nested, depth+1); childDepth > maxDepth {
				maxDepth = childDepth
			}
		}
		return maxDepth
	default:
		return depth
	}
}
func validateLogAttributes(attrs map[string]string) error {
	if len(attrs) > MaxLogMetadataEntries+3 {
		return fmt.Errorf("flattened metadata exceeds maximum of %d entries", MaxLogMetadataEntries)
	}
	for k, v := range attrs {
		if len(k) > MaxLogMetadataKeyLen {
			return fmt.Errorf("metadata key %q exceeds maximum length", k)
		}
		if len(v) > MaxLogMetadataValueLen {
			return fmt.Errorf("metadata value for %q exceeds maximum length", k)
		}
	}
	return nil
}
