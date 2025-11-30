package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type objectStore struct {
	db *sql.DB
}

func NewObjectStore(db *sql.DB) ObjectStore {
	return &objectStore{db: db}
}

type ObjectRecord struct {
	ID             string
	BucketID       string
	Key            string
	State          string
	SizeExpected   int64
	CradleServerID string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (s *objectStore) CreatePending(ctx context.Context, id, bucketID, key string, sizeExpected int64, cradleServerID string, createdAt time.Time) (ObjectRecord, error) {
	stamp := createdAt.UTC().Truncate(time.Microsecond)
	micros := stamp.UnixMicro()

	const insertObject = `
INSERT INTO objects (object_id, bucket_id, key, state, size_expected, cradle_server_id, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
`

	_, err := s.db.ExecContext(ctx, insertObject, id, bucketID, key, "PENDING", sizeExpected, cradleServerID, micros, micros)
	if err != nil {
		return ObjectRecord{}, fmt.Errorf("insert object: %w", err)
	}

	return ObjectRecord{
		ID:             id,
		BucketID:       bucketID,
		Key:            key,
		State:          "PENDING",
		SizeExpected:   sizeExpected,
		CradleServerID: cradleServerID,
		CreatedAt:      stamp,
		UpdatedAt:      stamp,
	}, nil
}
