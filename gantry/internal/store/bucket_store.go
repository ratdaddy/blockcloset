package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

var ErrBucketAlreadyExists = errors.New("bucket already exists")

type bucketStore struct {
	db *sql.DB
}

func NewBucketStore(db *sql.DB) BucketStore {
	return &bucketStore{db: db}
}

type BucketRecord struct {
	ID        string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (s *bucketStore) Create(ctx context.Context, id string, name string, createdAt time.Time) (BucketRecord, error) {
	stamp := createdAt.UTC().Truncate(time.Microsecond)
	micros := stamp.UnixMicro()

	const insertBucket = `INSERT INTO buckets (id, name, created_at, updated_at) VALUES (?, ?, ?, ?)`

	if _, err := s.db.ExecContext(ctx, insertBucket, id, name, micros, micros); err != nil {
		var sqliteErr *sqlite.Error
		if errors.As(err, &sqliteErr) && sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE {
			return BucketRecord{}, fmt.Errorf("insert bucket: %w", ErrBucketAlreadyExists)
		}
		return BucketRecord{}, fmt.Errorf("insert bucket: %w", err)
	}

	return BucketRecord{ID: id, Name: name, CreatedAt: stamp, UpdatedAt: stamp}, nil
}

func (s *bucketStore) List(ctx context.Context) ([]BucketRecord, error) {
	const selectBuckets = `SELECT id, name, created_at, updated_at FROM buckets ORDER BY created_at ASC`

	rows, err := s.db.QueryContext(ctx, selectBuckets)
	if err != nil {
		return nil, fmt.Errorf("list buckets: %w", err)
	}
	defer rows.Close()

	var records []BucketRecord

	for rows.Next() {
		var (
			rec       BucketRecord
			createdAt int64
			updatedAt int64
		)

		if err = rows.Scan(&rec.ID, &rec.Name, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("scan bucket: %w", err)
		}

		rec.CreatedAt = time.UnixMicro(createdAt).UTC()
		rec.UpdatedAt = time.UnixMicro(updatedAt).UTC()

		records = append(records, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate buckets: %w", err)
	}

	return records, nil
}
