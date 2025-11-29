package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var ErrNoCradleServersAvailable = errors.New("no cradle servers available")

type cradleServerStore struct {
	db *sql.DB
}

func NewCradleServerStore(db *sql.DB) CradleServerStore {
	return &cradleServerStore{db: db}
}

type CradleServerRecord struct {
	ID        string
	Address   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (s *cradleServerStore) Upsert(ctx context.Context, id string, address string, updatedAt time.Time) (CradleServerRecord, error) {
	stamp := updatedAt.UTC().Truncate(time.Microsecond)
	micros := stamp.UnixMicro()

	const upsertObject = `
INSERT INTO cradle_servers (id, address, created_at, updated_at)
VALUES ($1, $2, $3, $3)
ON CONFLICT (address)
DO UPDATE SET
	updated_at = EXCLUDED.updated_at
RETURNING id, address, created_at, updated_at;
`

	row := s.db.QueryRowContext(ctx, upsertObject, id, address, micros, micros)

	var (
		rec       CradleServerRecord
		createdAt int64
	)

	if err := row.Scan(&rec.ID, &rec.Address, &createdAt, &micros); err != nil {
		return CradleServerRecord{}, fmt.Errorf("upsert cradle server: %w", err)
	}
	rec.CreatedAt = time.UnixMicro(createdAt).UTC()
	rec.UpdatedAt = stamp

	return rec, nil
}

func (s *cradleServerStore) SelectForUpload(ctx context.Context) (CradleServerRecord, error) {
	const selectCradleServer = `SELECT id, address, created_at, updated_at FROM cradle_servers LIMIT 1`

	row := s.db.QueryRowContext(ctx, selectCradleServer)

	var (
		rec       CradleServerRecord
		createdAt int64
		updatedAt int64
	)

	if err := row.Scan(&rec.ID, &rec.Address, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return CradleServerRecord{}, ErrNoCradleServersAvailable
		}
		return CradleServerRecord{}, fmt.Errorf("select cradle server: %w", err)
	}

	rec.CreatedAt = time.UnixMicro(createdAt).UTC()
	rec.UpdatedAt = time.UnixMicro(updatedAt).UTC()

	return rec, nil
}
