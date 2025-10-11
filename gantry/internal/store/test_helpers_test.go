package store_test

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/ratdaddy/blockcloset/gantry/internal/config"
	"github.com/ratdaddy/blockcloset/gantry/internal/database"
)

var (
	migrationsOnce sync.Once
	migrationsErr  error
)

func ensureMigrationsCurrent(t *testing.T) {
	t.Helper()

	migrationsOnce.Do(func() {
		defaultPath := defaultDBPath()

		ctx := context.Background()

		db, err := database.OpenDatabase(ctx, defaultPath)
		if err != nil {
			migrationsErr = fmt.Errorf("open database: %w", err)
			return
		}
		defer db.Close()

		if err := database.VerifyMigrationsCurrent(db); err != nil {
			migrationsErr = fmt.Errorf("verify migrations current: %w", err)
			return
		}
	})

	if migrationsErr != nil {
		t.Fatalf("%v", migrationsErr)
	}
}

func defaultDBPath() string {
	if config.DatabasePath == "" {
		config.Init()
	}
	return config.DatabasePath
}

func openIsolatedDB(t *testing.T) *sql.DB {
	t.Helper()

	ensureMigrationsCurrent(t)

	base := defaultDBPath()
	copyPath := copyDatabaseFile(t, base)

	ctx := context.Background()

	db, err := database.OpenDatabase(ctx, copyPath)
	if err != nil {
		t.Fatalf("open isolated db: %v", err)
	}

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close isolated db: %v", err)
		}
	})

	return db
}

func copyDatabaseFile(t *testing.T, src string) string {
	t.Helper()

	outDir := t.TempDir()
	dst := filepath.Join(outDir, "test.db")

	srcFile, err := os.Open(src)
	if err != nil {
		t.Fatalf("open source db: %v", err)
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		t.Fatalf("create temp db: %v", err)
	}

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		dstFile.Close()
		t.Fatalf("copy db: %v", err)
	}

	if err := dstFile.Sync(); err != nil {
		dstFile.Close()
		t.Fatalf("sync temp db: %v", err)
	}

	if err := dstFile.Close(); err != nil {
		t.Fatalf("close temp db: %v", err)
	}

	return dst
}
