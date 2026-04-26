package testutil

import (
	"context"
	"sync"
	"time"

	"github.com/ratdaddy/blockcloset/gantry/internal/store"
)

// ObjectCreateCall captures the parameters for CreatePending invocations.
type ObjectCreateCall struct {
	ID             string
	BucketID       string
	Key            string
	SizeExpected   int64
	CradleServerID string
	CreatedAt      time.Time
}

// ObjectCommitCall captures the parameters for CommitWithReplace invocations.
type ObjectCommitCall struct {
	ObjectID       string
	SizeActual     int64
	LastModifiedMs int64
	UpdatedAt      time.Time
}

// ObjectStoreFake implements store.ObjectStore for tests.
type ObjectStoreFake struct {
	mu                sync.Mutex
	createErr         error
	createCalls       []ObjectCreateCall
	createResponse    store.ObjectRecord
	hasCreateResponse bool
	commitErr         error
	commitCalls       []ObjectCommitCall
}

var _ store.ObjectStore = (*ObjectStoreFake)(nil)

func NewFakeObjectStore() *ObjectStoreFake {
	return &ObjectStoreFake{}
}

func (f *ObjectStoreFake) SetCreateError(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.createErr = err
}

func (f *ObjectStoreFake) SetCreateResponse(rec store.ObjectRecord) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.createResponse = rec
	f.hasCreateResponse = true
}

func (f *ObjectStoreFake) CreatePending(ctx context.Context, id, bucketID, key string, sizeExpected int64, cradleServerID string, createdAt time.Time) (store.ObjectRecord, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.createCalls = append(f.createCalls, ObjectCreateCall{
		ID:             id,
		BucketID:       bucketID,
		Key:            key,
		SizeExpected:   sizeExpected,
		CradleServerID: cradleServerID,
		CreatedAt:      createdAt,
	})

	if f.createErr != nil {
		return store.ObjectRecord{}, f.createErr
	}

	if f.hasCreateResponse {
		return f.createResponse, nil
	}

	return store.ObjectRecord{
		ID:             id,
		BucketID:       bucketID,
		Key:            key,
		State:          "PENDING",
		SizeExpected:   sizeExpected,
		CradleServerID: cradleServerID,
		CreatedAt:      createdAt,
		UpdatedAt:      createdAt,
	}, nil
}

func (f *ObjectStoreFake) SetCommitError(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.commitErr = err
}

func (f *ObjectStoreFake) CommitWithReplace(ctx context.Context, objectID string, sizeActual int64, lastModifiedMs int64, updatedAt time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.commitCalls = append(f.commitCalls, ObjectCommitCall{
		ObjectID:       objectID,
		SizeActual:     sizeActual,
		LastModifiedMs: lastModifiedMs,
		UpdatedAt:      updatedAt,
	})
	return f.commitErr
}

func (f *ObjectStoreFake) CommitCalls() []ObjectCommitCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	calls := make([]ObjectCommitCall, len(f.commitCalls))
	copy(calls, f.commitCalls)
	return calls
}

func (f *ObjectStoreFake) Calls() []ObjectCreateCall {
	f.mu.Lock()
	defer f.mu.Unlock()

	calls := make([]ObjectCreateCall, len(f.createCalls))
	copy(calls, f.createCalls)
	return calls
}
