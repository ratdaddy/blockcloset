package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

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

	const query = `
INSERT INTO cradle_servers (id, address, created_at, updated_at)
VALUES ($1, $2, $3, $3)
ON CONFLICT (address)
DO UPDATE SET
	id = EXCLUDED.id,
	updated_at = EXCLUDED.updated_at
RETURNING id, address, created_at, updated_at;
`

	row := s.db.QueryRowContext(ctx, query, id, address, micros, micros)

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
