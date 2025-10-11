package grpcsvc

import (
	"context"
	"sync"
	"time"

	"github.com/ratdaddy/blockcloset/gantry/internal/store"
)

type bucketCreateCall struct {
	ID        string
	Name      string
	CreatedAt time.Time
}

type bucketStoreFake struct {
	mu                sync.Mutex
	createErr         error
	createResponse    store.BucketRecord
	hasCreateResponse bool
	createCalls       []bucketCreateCall

	listErr     error
	listRecords []store.BucketRecord
	listCalls   int
}

var _ store.BucketStore = (*bucketStoreFake)(nil)

func newFakeBucketStore() *bucketStoreFake {
	return &bucketStoreFake{}
}

func (f *bucketStoreFake) SetCreateError(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.createErr = err
}

func (f *bucketStoreFake) SetCreateResponse(rec store.BucketRecord) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.createResponse = rec
	f.hasCreateResponse = true
}

func (f *bucketStoreFake) SetListError(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.listErr = err
}

func (f *bucketStoreFake) SetListRecords(records []store.BucketRecord) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.listRecords = append([]store.BucketRecord(nil), records...)
}

func (f *bucketStoreFake) Create(ctx context.Context, id, name string, createdAt time.Time) (store.BucketRecord, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.createCalls = append(f.createCalls, bucketCreateCall{
		ID:        id,
		Name:      name,
		CreatedAt: createdAt,
	})

	if f.createErr != nil {
		return store.BucketRecord{}, f.createErr
	}

	if f.hasCreateResponse {
		return f.createResponse, nil
	}

	return store.BucketRecord{
		ID:        id,
		Name:      name,
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	}, nil
}

func (f *bucketStoreFake) List(ctx context.Context) ([]store.BucketRecord, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.listCalls++

	if f.listErr != nil {
		return nil, f.listErr
	}

	records := make([]store.BucketRecord, len(f.listRecords))
	copy(records, f.listRecords)
	return records, nil
}

func (f *bucketStoreFake) Calls() []bucketCreateCall {
	f.mu.Lock()
	defer f.mu.Unlock()

	calls := make([]bucketCreateCall, len(f.createCalls))
	copy(calls, f.createCalls)
	return calls
}

func (f *bucketStoreFake) ListCallCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.listCalls
}

type fakeStore struct {
	buckets store.BucketStore
}

func newFakeStore(buckets store.BucketStore) *fakeStore {
	return &fakeStore{buckets: buckets}
}

func (f *fakeStore) Buckets() store.BucketStore {
	return f.buckets
}

var _ store.Store = (*fakeStore)(nil)
