package database

import (
	"database/sql"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

type PostgreSQL struct{}

func (r *PostgreSQL) Connect(dsn string) *bun.DB {
	var sqldb *sql.DB
	var err error

	maxRetries := 3
	retryWaitTime := 5 * time.Second

	connected := false
	for range maxRetries {
		sqldb = sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))

		err = sqldb.Ping()
		if err == nil {
			slog.Info("successfully connected to the PostgreSQL database")
			connected = true
			break
		}

		slog.Warn("could not connect to the database. Retrying in 5s...", slog.Any("error", err))

		time.Sleep(retryWaitTime)
	}

	if !connected {
		// Returning a broken DB would let the app start in an unusable state and
		// hang on the first query. Fail fast instead.
		slog.Error("failed to connect to the database after multiple attempts", slog.Any("error", err))
		os.Exit(1)
	}

	// Bound the pool so a burst of slow requests (e.g. concurrent uploads) cannot
	// open unlimited connections and exhaust Postgres.
	sqldb.SetMaxOpenConns(maxOpenConns())
	sqldb.SetMaxIdleConns(10)
	sqldb.SetConnMaxLifetime(30 * time.Minute)
	sqldb.SetConnMaxIdleTime(5 * time.Minute)

	return bun.NewDB(sqldb, pgdialect.New())
}

// maxOpenConns reads DB_MAX_OPEN_CONNS, defaulting to a conservative 25.
func maxOpenConns() int {
	if v := os.Getenv("DB_MAX_OPEN_CONNS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return 25
}
