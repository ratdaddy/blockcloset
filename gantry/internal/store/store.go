package store

import (
	"context"
	"database/sql"
	"time"
)

type BucketStore interface {
	Create(ctx context.Context, id string, name string, createdAt time.Time) (BucketRecord, error)
	List(ctx context.Context) ([]BucketRecord, error)
}

type Store interface {
	Buckets() BucketStore
}

type sqlStore struct {
	buckets BucketStore
}

func New(db *sql.DB) Store {
	return &sqlStore{
		buckets: NewBucketStore(db),
	}
}

func (s *sqlStore) Buckets() BucketStore {
	return s.buckets
}
