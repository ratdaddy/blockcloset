package testutil

import (
	"github.com/ratdaddy/blockcloset/gantry/internal/store"
)

// StoreFake implements store.Store for tests.
type StoreFake struct {
	BucketsStore store.BucketStore
	CradleStore  store.CradleServerStore
	ObjectsStore store.ObjectStore
}

var _ store.Store = (*StoreFake)(nil)

// StoreOption configures a StoreFake.
type StoreOption func(*StoreFake)

// WithBuckets sets a custom BucketStore implementation.
func WithBuckets(b store.BucketStore) StoreOption {
	return func(f *StoreFake) {
		f.BucketsStore = b
	}
}

// WithCradles sets a custom CradleServerStore implementation.
func WithCradles(c store.CradleServerStore) StoreOption {
	return func(f *StoreFake) {
		f.CradleStore = c
	}
}

// WithObjects sets a custom ObjectStore implementation.
func WithObjects(o store.ObjectStore) StoreOption {
	return func(f *StoreFake) {
		f.ObjectsStore = o
	}
}

// NewFakeStore creates a StoreFake with default fakes for all stores.
// Use options to override specific stores.
func NewFakeStore(opts ...StoreOption) *StoreFake {
	f := &StoreFake{
		BucketsStore: NewFakeBucketStore(),
		CradleStore:  NewFakeCradleStore(),
		ObjectsStore: NewFakeObjectStore(),
	}
	for _, opt := range opts {
		opt(f)
	}
	return f
}

func (f *StoreFake) Buckets() store.BucketStore {
	return f.BucketsStore
}

func (f *StoreFake) CradleServers() store.CradleServerStore {
	return f.CradleStore
}

func (f *StoreFake) Objects() store.ObjectStore {
	return f.ObjectsStore
}
