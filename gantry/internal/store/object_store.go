package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var ErrObjectNotPending = errors.New("object not found or not in PENDING state")

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

func (s *objectStore) CommitWithReplace(ctx context.Context, objectID string, sizeActual int64, lastModifiedMs int64, updatedAt time.Time) error {
	stamp := updatedAt.UTC().Truncate(time.Microsecond)
	micros := stamp.UnixMicro()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("commit object, begin tx: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		WITH pending AS (
			SELECT bucket_id, key FROM objects WHERE object_id = ? AND state = 'PENDING'
		)
		UPDATE objects
		SET state = 'REPLACED',
		    updated_at = ?
		WHERE state = 'COMMITTED'
		  AND bucket_id = (SELECT bucket_id FROM pending)
		  AND key = (SELECT key FROM pending)
	`, objectID, micros)
	if err != nil {
		return fmt.Errorf("commit object, replace previous: %w", err)
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE objects
		SET state = 'COMMITTED',
		    size_actual = ?,
		    last_modified = ?,
		    updated_at = ?
		WHERE object_id = ?
		  AND state = 'PENDING'
	`, sizeActual, lastModifiedMs, micros, objectID)
	if err != nil {
		return fmt.Errorf("commit object: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("commit object, rows affected: %w", err)
	}

	if rows != 1 {
		return fmt.Errorf("commit object: %w", ErrObjectNotPending)
	}

	return tx.Commit()
}
