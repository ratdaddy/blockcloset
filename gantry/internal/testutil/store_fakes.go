package testutil

import (
	"context"
	"sync"
	"time"

	"github.com/ratdaddy/blockcloset/gantry/internal/store"
)

// BucketCreateCall captures the parameters for Create invocations.
type BucketCreateCall struct {
	ID        string
	Name      string
	CreatedAt time.Time
}

// BucketStoreFake implements store.BucketStore for tests.
type BucketStoreFake struct {
	mu                sync.Mutex
	createErr         error
	createResponse    store.BucketRecord
	hasCreateResponse bool
	createCalls       []BucketCreateCall

	listErr     error
	listRecords []store.BucketRecord
	listCalls   int

	getByNameErr      error
	getByNameResponse store.BucketRecord
	getByNameCalls    []string
}

var _ store.BucketStore = (*BucketStoreFake)(nil)

func NewFakeBucketStore() *BucketStoreFake {
	return &BucketStoreFake{}
}

func (f *BucketStoreFake) SetCreateError(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.createErr = err
}

func (f *BucketStoreFake) SetCreateResponse(rec store.BucketRecord) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.createResponse = rec
	f.hasCreateResponse = true
}

func (f *BucketStoreFake) SetListError(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.listErr = err
}

func (f *BucketStoreFake) SetListRecords(records []store.BucketRecord) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.listRecords = append([]store.BucketRecord(nil), records...)
}

func (f *BucketStoreFake) SetGetByNameError(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.getByNameErr = err
}

func (f *BucketStoreFake) Create(ctx context.Context, id, name string, createdAt time.Time) (store.BucketRecord, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.createCalls = append(f.createCalls, BucketCreateCall{ID: id, Name: name, CreatedAt: createdAt})

	if f.createErr != nil {
		return store.BucketRecord{}, f.createErr
	}

	if f.hasCreateResponse {
		return f.createResponse, nil
	}

	return store.BucketRecord{ID: id, Name: name, CreatedAt: createdAt, UpdatedAt: createdAt}, nil
}

func (f *BucketStoreFake) List(ctx context.Context) ([]store.BucketRecord, error) {
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

func (f *BucketStoreFake) GetByName(ctx context.Context, name string) (store.BucketRecord, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.getByNameCalls = append(f.getByNameCalls, name)

	if f.getByNameErr != nil {
		return store.BucketRecord{}, f.getByNameErr
	}

	return f.getByNameResponse, nil
}

func (f *BucketStoreFake) GetByNameCalls() []string {
	f.mu.Lock()
	defer f.mu.Unlock()

	calls := make([]string, len(f.getByNameCalls))
	copy(calls, f.getByNameCalls)
	return calls
}

func (f *BucketStoreFake) Calls() []BucketCreateCall {
	f.mu.Lock()
	defer f.mu.Unlock()

	calls := make([]BucketCreateCall, len(f.createCalls))
	copy(calls, f.createCalls)
	return calls
}

func (f *BucketStoreFake) ListCallCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.listCalls
}

// CradleUpsertCall captures the parameters for Upsert invocations.
type CradleUpsertCall struct {
	ID      string
	Address string
	Stamp   time.Time
}

// CradleStoreFake implements store.CradleServerStore for tests.
type CradleStoreFake struct {
	mu                        sync.Mutex
	upsertErr                 error
	upsertCalls               []CradleUpsertCall
	selectForUploadResponse   store.CradleServerRecord
	selectForUploadErr        error
	selectForUploadCallCount  int
}

var _ store.CradleServerStore = (*CradleStoreFake)(nil)

func NewFakeCradleStore() *CradleStoreFake {
	return &CradleStoreFake{}
}

func (f *CradleStoreFake) SetUpsertError(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.upsertErr = err
}

func (f *CradleStoreFake) SetSelectForUploadResponse(rec store.CradleServerRecord) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.selectForUploadResponse = rec
}

func (f *CradleStoreFake) SetSelectForUploadError(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.selectForUploadErr = err
}

func (f *CradleStoreFake) Upsert(ctx context.Context, id string, address string, stamp time.Time) (store.CradleServerRecord, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.upsertCalls = append(f.upsertCalls, CradleUpsertCall{ID: id, Address: address, Stamp: stamp})

	if f.upsertErr != nil {
		return store.CradleServerRecord{}, f.upsertErr
	}

	return store.CradleServerRecord{ID: id, Address: address, CreatedAt: stamp, UpdatedAt: stamp}, nil
}

func (f *CradleStoreFake) Calls() []CradleUpsertCall {
	f.mu.Lock()
	defer f.mu.Unlock()

	calls := make([]CradleUpsertCall, len(f.upsertCalls))
	copy(calls, f.upsertCalls)
	return calls
}

func (f *CradleStoreFake) CallCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.upsertCalls)
}

func (f *CradleStoreFake) SelectForUpload(ctx context.Context) (store.CradleServerRecord, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.selectForUploadCallCount++

	if f.selectForUploadErr != nil {
		return store.CradleServerRecord{}, f.selectForUploadErr
	}

	return f.selectForUploadResponse, nil
}

func (f *CradleStoreFake) SelectForUploadCallCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.selectForUploadCallCount
}

// StoreFake implements store.Store for tests.
type StoreFake struct {
	BucketsStore store.BucketStore
	CradleStore  store.CradleServerStore
}

var _ store.Store = (*StoreFake)(nil)

func NewFakeStore(b store.BucketStore, c store.CradleServerStore) *StoreFake {
	return &StoreFake{BucketsStore: b, CradleStore: c}
}

func (f *StoreFake) Buckets() store.BucketStore {
	return f.BucketsStore
}

func (f *StoreFake) CradleServers() store.CradleServerStore {
	return f.CradleStore
}
