package clickhouse

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/fivemanage/lite/api"
	"github.com/fivemanage/lite/pkg/querybuilder"
)

func (c *Client) QueryLog(ctx context.Context, organizationID, datasetID, logID string) (*api.DatasetLog, error) {
	if c == nil || !c.Enabled || c.conn == nil {
		return nil, ErrUnavailable
	}
	chCtx := clickhouse.Context(ctx, clickhouse.WithParameters(clickhouse.Parameters{
		"TeamId":    organizationID,
		"DatasetId": datasetID,
		"TraceId":   logID,
	}))

	query, err := c.conn.Query(chCtx, "SELECT * FROM logs WHERE TeamId = {TeamId:String} AND DatasetId = {DatasetId:String} AND TraceId = {TraceId:String} LIMIT 1")
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := query.Close(); err != nil {
			slog.Error("error closing clickhouse rows", "err", err)
		}
	}()

	for query.Next() {
		var (
			Timestamp     time.Time
			DatasetId     string
			TraceId       string
			TeamId        string
			Body          string
			Attributes    map[string]string
			RetentionDays uint32
		)
		if err := query.Scan(&Timestamp, &DatasetId, &TraceId, &TeamId, &Body, &Attributes, &RetentionDays); err != nil {
			return nil, err
		}

		log := api.DatasetLog{
			Timestamp:  Timestamp,
			DatasetId:  DatasetId,
			TraceId:    TraceId,
			TeamId:     TeamId,
			Body:       Body,
			Attributes: Attributes,
		}

		return &log, nil
	}

	if err := query.Err(); err != nil {
		return nil, err
	}

	return nil, nil
}

func (c *Client) QueryLogs(ctx context.Context, organizationID, datasetID string, startTime, endTime time.Time, filter api.DatasetFilter, cursor int) ([]api.DatasetLog, error) {
	if c == nil || !c.Enabled || c.conn == nil {
		return nil, ErrUnavailable
	}
	qb := querybuilder.New().Filter(filter).WithDateRange(startTime, endTime)
	query, args, err := qb.Build(organizationID, datasetID)
	if err != nil {
		return nil, err
	}

	rows, err := c.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := rows.Close(); err != nil {
			slog.Error("error closing clickhouse rows", "err", err)
		}
	}()

	var logs []api.DatasetLog
	for rows.Next() {
		var (
			Timestamp     time.Time
			DatasetId     string
			TraceId       string
			TeamId        string
			Body          string
			Attributes    map[string]string
			RetentionDays uint32
		)
		if err := rows.Scan(&Timestamp, &DatasetId, &TraceId, &TeamId, &Body, &Attributes, &RetentionDays); err != nil {
			return nil, err
		}
		logs = append(logs, api.DatasetLog{
			Timestamp:  Timestamp,
			DatasetId:  DatasetId,
			TraceId:    TraceId,
			TeamId:     TeamId,
			Body:       Body,
			Attributes: Attributes,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return logs, nil
}

func (r *Client) QueryLogFields(ctx context.Context, organizationID, datasetID string) ([]LogField, error) {
	if r == nil || !r.Enabled || r.conn == nil {
		return nil, ErrUnavailable
	}
	chCtx := clickhouse.Context(ctx, clickhouse.WithParameters(clickhouse.Parameters{
		"TeamId":    organizationID,
		"DatasetId": datasetID,
	}))

	query, err := r.conn.Query(chCtx, "SELECT DISTINCT Key, Type FROM log_keys WHERE TeamId = {TeamId:String} AND DatasetId = {DatasetId:String}")
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := query.Close(); err != nil {
			slog.Error("error closing clickhouse rows", "err", err)
		}
	}()

	var fields []LogField

	for query.Next() {
		var (
			field     string
			fieldType string
		)

		err := query.Scan(&field, &fieldType)
		if err != nil {
			return nil, err
		}

		fields = append(fields, LogField{Field: field, Type: fieldType})
	}

	if err := query.Err(); err != nil {
		return nil, err
	}

	return fields, nil
}

func (r *Client) QueryTotalLogs(ctx context.Context, organizationID, datasetID string) (int, error) {
	if r == nil || !r.Enabled || r.conn == nil {
		return 0, ErrUnavailable
	}
	chCtx := clickhouse.Context(ctx, clickhouse.WithParameters(clickhouse.Parameters{
		"TeamId":    organizationID,
		"DatasetId": datasetID,
	}))

	var count uint64
	err := r.conn.QueryRow(chCtx, "SELECT count() FROM logs WHERE TeamId = {TeamId:String} AND DatasetId = {DatasetId:String}").Scan(&count)
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

func (r *Client) QueryTotalLogsByOrg(ctx context.Context, organizationID string) (int, error) {
	if r == nil || !r.Enabled || r.conn == nil {
		return 0, ErrUnavailable
	}
	chCtx := clickhouse.Context(ctx, clickhouse.WithParameters(clickhouse.Parameters{
		"TeamId": organizationID,
	}))

	var count uint64
	err := r.conn.QueryRow(chCtx, "SELECT count() FROM logs WHERE TeamId = {TeamId:String}").Scan(&count)
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

// QueryLogCountsByDataset returns the log volume per dataset for an organization,
// keyed by dataset id.
func (r *Client) QueryLogCountsByDataset(ctx context.Context, organizationID string) (map[string]int, error) {
	if r == nil || !r.Enabled || r.conn == nil {
		return nil, ErrUnavailable
	}
	chCtx := clickhouse.Context(ctx, clickhouse.WithParameters(clickhouse.Parameters{
		"TeamId": organizationID,
	}))

	rows, err := r.conn.Query(chCtx, "SELECT DatasetId, count() FROM logs WHERE TeamId = {TeamId:String} GROUP BY DatasetId")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var datasetID string
		var count uint64
		if err := rows.Scan(&datasetID, &count); err != nil {
			return nil, err
		}
		counts[datasetID] = int(count)
	}
	return counts, rows.Err()
}

// QueryLogTimeseriesByOrg returns the daily log volume for an organization over
// the trailing `days` days, oldest day first. Days with no logs are omitted.
func (r *Client) QueryLogTimeseriesByOrg(ctx context.Context, organizationID string, days int) ([]api.LogDayCount, error) {
	if r == nil || !r.Enabled || r.conn == nil {
		return nil, ErrUnavailable
	}
	chCtx := clickhouse.Context(ctx, clickhouse.WithParameters(clickhouse.Parameters{
		"TeamId": organizationID,
		"Days":   fmt.Sprintf("%d", days),
	}))

	rows, err := r.conn.Query(chCtx, "SELECT toDate(Timestamp) AS day, count() FROM logs WHERE TeamId = {TeamId:String} AND Timestamp >= now() - INTERVAL {Days:UInt32} DAY GROUP BY day ORDER BY day")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var series []api.LogDayCount
	for rows.Next() {
		var day time.Time
		var count uint64
		if err := rows.Scan(&day, &count); err != nil {
			return nil, err
		}
		series = append(series, api.LogDayCount{Date: day.Format("2006-01-02"), Count: int(count)})
	}
	return series, rows.Err()
}
