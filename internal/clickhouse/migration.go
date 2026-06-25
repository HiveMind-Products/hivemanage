package clickhouse

import (
	"context"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/fivemanage/lite/internal/project"
	"github.com/golang-migrate/migrate/v4"
	clickhouseMigrate "github.com/golang-migrate/migrate/v4/database/clickhouse"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func AutoMigrate(ctx context.Context, config *Config) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	options := getClickhouseOptions(config)
	db := clickhouse.OpenDB(options)

	defer func() {
		err := db.Close()
		if err != nil {
			slog.Error("failed to close clickhouse connection", slog.Any("error", err))
		}
	}()

	if err := db.PingContext(ctx); err != nil {
		slog.Error("failed to ping clickhouse", slog.Any("error", err))
		return err
	}

	driver, err := clickhouseMigrate.WithInstance(db, &clickhouseMigrate.Config{
		MigrationsTableEngine: "MergeTree",
		MultiStatementEnabled: true,
	})
	if err != nil {
		slog.Error("failed to create clickhouse driver", slog.Any("error", err))
		return err
	}

	migrationsPath := filepath.Join(project.GetRoot(), "internal", "clickhouse", "migrations")
	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		slog.Error("failed to get absolute path", slog.Any("error", err))
		return err
	}

	sourcePath := filepath.ToSlash(absPath)
	sourceURL := "file://" + sourcePath

	m, err := migrate.NewWithDatabaseInstance(sourceURL, "default", driver)
	if err != nil {
		slog.Error("error creating migration instance", slog.Any("error", err))
		return err
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		slog.Error("failed to run migrations", slog.Any("error", err))
		return err
	}

	slog.Info("clickhouse migrations successfully completed")
	return nil
}
