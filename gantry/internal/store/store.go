package store

import (
	"context"
	"database/sql"
	"time"
)

type BucketStore interface {
	Create(ctx context.Context, id string, name string, createdAt time.Time) (BucketRecord, error)
	List(ctx context.Context) ([]BucketRecord, error)
	GetByName(ctx context.Context, name string) (BucketRecord, error)
}

type CradleServerStore interface {
	Upsert(ctx context.Context, id string, address string, createdAt time.Time) (CradleServerRecord, error)
	SelectForUpload(ctx context.Context) (CradleServerRecord, error)
}

type ObjectStore interface {
	CreatePending(ctx context.Context, id, bucketID, key string, sizeExpected int64, cradleServerID string, createdAt time.Time) (ObjectRecord, error)
}

type Store interface {
	Buckets() BucketStore
	CradleServers() CradleServerStore
	Objects() ObjectStore
}

type sqlStore struct {
	buckets       BucketStore
	cradleServers CradleServerStore
	objects       ObjectStore
}

func New(db *sql.DB) Store {
	return &sqlStore{
		buckets:       NewBucketStore(db),
		cradleServers: NewCradleServerStore(db),
		objects:       NewObjectStore(db),
	}
}

func (s *sqlStore) Buckets() BucketStore {
	return s.buckets
}

func (s *sqlStore) CradleServers() CradleServerStore {
	return s.cradleServers
}

func (s *sqlStore) Objects() ObjectStore {
	return s.objects
}
