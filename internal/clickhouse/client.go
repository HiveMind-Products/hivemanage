package clickhouse

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type Config struct {
	Host     string
	Username string
	Password string
	Database string
}

var (
	ErrUnavailable = errors.New("clickhouse unavailable")
	// ErrWriteFailed wraps infrastructure failures from the batch write path so
	// callers can distinguish them from client-side validation errors.
	ErrWriteFailed = errors.New("clickhouse write failed")
)

type Client struct {
	conn    driver.Conn
	Enabled bool
}

type Log struct {
	TraceID       string
	Timestamp     time.Time
	TeamID        string
	DatasetID     string
	Body          string
	Attributes    map[string]string
	RetentionDays int
}

type LogField struct {
	Field string
	Type  string
}

func NewClient(config *Config, enabled bool) *Client {
	if !enabled {
		slog.Warn("clickhosue is disabled, logging is not available")
		return &Client{
			Enabled: false,
		}
	}

	c := &Client{
		Enabled: true,
	}

	options := getClickhouseOptions(config)
	conn, err := connect(options)
	if err != nil {
		slog.Error("error connecting to ClickHouse", slog.Any("error", err))
		c.Enabled = false
		return c
	}

	c.conn = conn
	return c
}

func getClickhouseOptions(config *Config) *clickhouse.Options {
	return &clickhouse.Options{
		Addr: []string{config.Host},
		Auth: clickhouse.Auth{
			Database: config.Database,
			Username: config.Username,
			Password: config.Password,
		},
		// MaxOpenConns: 1000,
		ClientInfo: clickhouse.ClientInfo{
			Products: []struct {
				Name    string
				Version string
			}{
				{Name: "lite-server", Version: "0.1"},
			},
		},
		Debugf: func(format string, v ...any) {
			slog.Debug(fmt.Sprintf(format, v...))
		},
	}
}

func connect(options *clickhouse.Options) (driver.Conn, error) {
	ctx := context.Background()
	conn, err := clickhouse.Open(options)
	if err != nil {
		return nil, err
	}

	if err := conn.Ping(ctx); err != nil {
		var exception *clickhouse.Exception
		if errors.As(err, &exception) {
			slog.Error("clickhouse exception on ping",
				"code", exception.Code,
				"message", exception.Message,
				"stacktrace", exception.StackTrace,
			)
		}
		return nil, err
	}

	return conn, nil
}

func (c *Client) BatchWriteLogRows(ctx context.Context, logs []*Log) error {
	if c == nil || !c.Enabled || c.conn == nil {
		return ErrUnavailable
	}

	if len(logs) == 0 {
		return nil
	}

	batch, err := c.conn.PrepareBatch(ctx, "INSERT INTO logs (Timestamp, DatasetId, TraceId, TeamId, Body, Attributes, RetentionDays)")
	if err != nil {
		slog.Error("failed to prepare clickhouse batch", "err", err)
		return fmt.Errorf("%w: failed to prepare batch: %v", ErrWriteFailed, err)
	}

	for _, log := range logs {
		err := batch.Append(
			log.Timestamp,
			log.DatasetID,
			log.TraceID,
			log.TeamID,
			log.Body,
			log.Attributes,
			log.RetentionDays,
		)
		if err != nil {
			slog.Error("failed to append log to clickhouse batch", "err", err)
			return fmt.Errorf("%w: failed to append log to batch: %v", ErrWriteFailed, err)
		}
	}

	if err := batch.Send(); err != nil {
		slog.Error("failed to send clickhouse batch", "err", err)
		return fmt.Errorf("%w: %v", ErrWriteFailed, err)
	}
	return nil
}
