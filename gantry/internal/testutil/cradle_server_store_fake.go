package testutil

import (
	"context"
	"sync"
	"time"

	"github.com/ratdaddy/blockcloset/gantry/internal/store"
)

// CradleUpsertCall captures the parameters for Upsert invocations.
type CradleUpsertCall struct {
	ID      string
	Address string
	Stamp   time.Time
}

// CradleStoreFake implements store.CradleServerStore for tests.
type CradleStoreFake struct {
	mu                       sync.Mutex
	upsertErr                error
	upsertCalls              []CradleUpsertCall
	allResponse              []store.CradleServerRecord
	allErr                   error
	allCallCount             int
	selectForUploadResponse  store.CradleServerRecord
	selectForUploadErr       error
	selectForUploadCallCount int
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

func (f *CradleStoreFake) UpsertCalls() []CradleUpsertCall {
	f.mu.Lock()
	defer f.mu.Unlock()

	calls := make([]CradleUpsertCall, len(f.upsertCalls))
	copy(calls, f.upsertCalls)
	return calls
}

func (f *CradleStoreFake) UpsertCallCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.upsertCalls)
}

func (f *CradleStoreFake) SetAllResponse(recs []store.CradleServerRecord) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.allResponse = append([]store.CradleServerRecord(nil), recs...)
}

func (f *CradleStoreFake) SetAllError(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.allErr = err
}

func (f *CradleStoreFake) All(ctx context.Context) ([]store.CradleServerRecord, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.allCallCount++

	if f.allErr != nil {
		return nil, f.allErr
	}

	recs := make([]store.CradleServerRecord, len(f.allResponse))
	copy(recs, f.allResponse)
	return recs, nil
}

func (f *CradleStoreFake) AllCallCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.allCallCount
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
