package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/ratdaddy/blockcloset/gantry/internal/config"

	_ "modernc.org/sqlite"
)

func Init(ctx context.Context) (*sql.DB, func(), error) {
	dbPath := config.DatabasePath

	db, err := OpenDatabase(ctx, dbPath)
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		if err := db.Close(); err != nil {
			slog.Warn("database close failed", "err", err)
		}
	}

	if err := VerifyMigrationsCurrent(db); err != nil {
		_ = db.Close()
		return nil, nil, err
	}

	return db, cleanup, nil
}

func OpenDatabase(ctx context.Context, dbPath string) (*sql.DB, error) {
	dsn := "file:" + filepath.ToSlash(dbPath)

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	return db, nil
}
