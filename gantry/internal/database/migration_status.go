package database

import (
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type migrationStatus struct {
	currentVersion    int
	dirty             bool
	latestFileVersion int
}

func VerifyMigrationsCurrent(db *sql.DB) error {
	current, err := migrationsCurrent(db)
	if err != nil {
		return fmt.Errorf("check migrations: %w", err)
	}

	if !current {
		return fmt.Errorf("database migrations are not up to date; run make migrate")
	}

	return nil
}

func migrationsCurrent(db *sql.DB) (bool, error) {
	status, err := getMigrationStatus(db)
	if err != nil {
		return false, err
	}

	if status.dirty {
		return false, nil
	}

	return status.currentVersion == status.latestFileVersion, nil
}

func getMigrationStatus(db *sql.DB) (migrationStatus, error) {
	status := migrationStatus{}

	latest, err := latestMigrationVersion()
	if err != nil {
		return status, err
	}
	status.latestFileVersion = latest

	var version sql.NullInt64
	var dirty sql.NullBool

	row := db.QueryRow(`SELECT version, dirty FROM schema_migrations LIMIT 1`)
	scanErr := row.Scan(&version, &dirty)
	if scanErr != nil {
		switch {
		case errors.Is(scanErr, sql.ErrNoRows):
			return status, nil
		case isNoSuchTableError(scanErr):
			return status, nil
		default:
			return status, fmt.Errorf("scan schema_migrations: %w", scanErr)
		}
	}

	if version.Valid {
		status.currentVersion = int(version.Int64)
	}

	if dirty.Valid {
		status.dirty = dirty.Bool
	}

	return status, nil
}

func latestMigrationVersion() (int, error) {
	var migrationsDirPath = locateMigrationsDir()
	if migrationsDirPath == "" {
		return 0, fmt.Errorf("migrations directory not found; ensure migrations are available in the runtime path")
	}

	entries, err := os.ReadDir(migrationsDirPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return 0, nil
		}
		return 0, fmt.Errorf("read migrations dir: %w", err)
	}

	var versions []int

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		dot := strings.Index(name, "_")
		if dot <= 0 {
			continue
		}

		numStr := name[:dot]
		if v, err := strconv.Atoi(numStr); err == nil {
			versions = append(versions, v)
		}
	}

	if len(versions) == 0 {
		return 0, nil
	}

	sort.Ints(versions)
	return versions[len(versions)-1], nil
}

func locateMigrationsDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	for {
		candidate := filepath.Join(dir, "migrations")
		if info, statErr := os.Stat(candidate); statErr == nil && info.IsDir() {
			return filepath.Clean(candidate)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func isNoSuchTableError(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "no such table")
}
