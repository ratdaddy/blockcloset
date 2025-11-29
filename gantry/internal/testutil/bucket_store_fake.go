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
