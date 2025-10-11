package grpcsvc

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/google/uuid"
	"github.com/ratdaddy/blockcloset/gantry/internal/store"
	"github.com/ratdaddy/blockcloset/pkg/storage/bucket"
	servicev1 "github.com/ratdaddy/blockcloset/proto/gen/gantry/service/v1"
)

func TestService_CreateBucket(t *testing.T) {
	t.Parallel()

	type tc struct {
		name             string
		bucket           string
		wantErr          bool
		wantResponse     bool
		code             codes.Code
		message          string
		storeErr         error
		expectStoreCall  bool
		wantConflictInfo bool
		conflictReason   servicev1.BucketOwnershipConflict_Reason
	}

	cases := []tc{
		{
			name:            "valid bucket",
			bucket:          "my-bucket-123",
			wantResponse:    true,
			expectStoreCall: true,
		},
		{
			name:         "invalid bucket",
			bucket:       "Bad!Name",
			wantErr:      true,
			code:         codes.InvalidArgument,
			message:      bucket.ErrInvalidBucketName.Error(),
			wantResponse: false,
		},
		{
			name:            "bucket store error surfaces as internal",
			bucket:          "store-error-bucket",
			wantErr:         true,
			code:            codes.Internal,
			message:         "bucket store error",
			storeErr:        errors.New("bucket store error"),
			expectStoreCall: true,
		},
		{
			name:             "bucket already exists surfaces as already exists",
			bucket:           "duplicate-bucket",
			wantErr:          true,
			code:             codes.AlreadyExists,
			message:          store.ErrBucketAlreadyExists.Error(),
			storeErr:         store.ErrBucketAlreadyExists,
			expectStoreCall:  true,
			wantConflictInfo: true,
			conflictReason:   servicev1.BucketOwnershipConflict_REASON_BUCKET_ALREADY_OWNED_BY_YOU,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			logger := newDiscardLogger()
			svc := New(logger, nil)
			buckets := newFakeBucketStore()
			if c.storeErr != nil {
				buckets.err = c.storeErr
			}
			svc.store = newFakeStore(buckets)

			resp, err := svc.CreateBucket(context.Background(), &servicev1.CreateBucketRequest{Name: c.bucket})

			if c.wantErr {
				assertGRPCError(t, err, c.code, c.message)
				if c.wantConflictInfo {
					assertConflictDetail(t, err, c.conflictReason, c.bucket)
				}
			} else {
				assertNoError(t, err)
			}

			if c.wantResponse {
				assertBucketResponse(t, resp, c.bucket)
			} else {
				assertNoResponse(t, resp)
			}

			if c.expectStoreCall {
				assertStoreCreateCalled(t, buckets, c.bucket)
			} else {
				assertStoreNotCalled(t, buckets)
			}
		})
	}
}

func assertGRPCError(t *testing.T, err error, code codes.Code, message string) {
	t.Helper()

	if err == nil {
		t.Fatalf("want error, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("want gRPC status error, got %v", err)
	}

	if st.Code() != code {
		t.Fatalf("status code: got %v, want %v", st.Code(), code)
	}

	if st.Message() != message {
		t.Fatalf("status message: got %q, want %q", st.Message(), message)
	}
}

func assertConflictDetail(t *testing.T, err error, reason servicev1.BucketOwnershipConflict_Reason, bucket string) {
	t.Helper()

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("want gRPC status error, got %v", err)
	}

	details := st.Details()
	if len(details) == 0 {
		t.Fatalf("status missing details: %v", err)
	}

	for _, detail := range details {
		conflict, ok := detail.(*servicev1.BucketOwnershipConflict)
		if !ok {
			continue
		}

		if conflict.GetReason() != reason {
			t.Fatalf("conflict reason: got %v, want %v", conflict.GetReason(), reason)
		}

		if conflict.GetBucket() != bucket {
			t.Fatalf("conflict bucket: got %q, want %q", conflict.GetBucket(), bucket)
		}

		return
	}

	t.Fatalf("status missing BucketOwnershipConflict detail: %v", err)
}

func assertNoError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("want nil error, got %v", err)
	}
}

func assertBucketResponse(t *testing.T, resp *servicev1.CreateBucketResponse, wantName string) {
	t.Helper()

	if resp == nil || resp.GetBucket() == nil {
		t.Fatalf("response bucket missing: %#v", resp)
	}

	if resp.GetBucket().GetName() != wantName {
		t.Fatalf("bucket name: got %q, want %q", resp.GetBucket().GetName(), wantName)
	}
}

func assertNoResponse(t *testing.T, resp *servicev1.CreateBucketResponse) {
	t.Helper()

	if resp != nil {
		t.Fatalf("response: got %#v, want nil", resp)
	}
}

func assertStoreCreateCalled(t *testing.T, buckets *fakeBucketStore, wantName string) {
	t.Helper()

	if buckets == nil {
		t.Fatal("bucket store fake not provided")
	}

	calls := buckets.Calls()
	if len(calls) != 1 {
		t.Fatalf("bucket store calls: got %d, want 1 (calls=%v)", len(calls), calls)
	}

	call := calls[0]
	if call.ID == "" {
		t.Fatal("bucket store id not populated")
	}
	if _, err := uuid.Parse(call.ID); err != nil {
		t.Fatalf("bucket store id: %q not a valid UUID: %v", call.ID, err)
	}
	if call.Name != wantName {
		t.Fatalf("bucket store name: got %q, want %q", call.Name, wantName)
	}

	if call.CreatedAt.IsZero() {
		t.Fatal("bucket store createdAt timestamp not populated")
	}
}

func assertStoreNotCalled(t *testing.T, buckets *fakeBucketStore) {
	t.Helper()

	if buckets == nil {
		t.Fatal("bucket store fake not provided")
	}

	if calls := buckets.Calls(); len(calls) != 0 {
		t.Fatalf("bucket store calls: got %d, want 0 (calls=%v)", len(calls), calls)
	}
}

func newDiscardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
}

type bucketCreateCall struct {
	ID        string
	Name      string
	CreatedAt time.Time
}

type fakeBucketStore struct {
	mu     sync.Mutex
	err    error
	record store.BucketRecord
	calls  []bucketCreateCall
}

var _ store.BucketStore = (*fakeBucketStore)(nil)

func newFakeBucketStore() *fakeBucketStore {
	return &fakeBucketStore{}
}

func (f *fakeBucketStore) Create(ctx context.Context, id string, name string, createdAt time.Time) (store.BucketRecord, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.calls = append(f.calls, bucketCreateCall{
		ID:        id,
		Name:      name,
		CreatedAt: createdAt,
	})

	if f.err != nil {
		return store.BucketRecord{}, f.err
	}

	if f.record.Name == "" && f.record.CreatedAt.IsZero() && f.record.UpdatedAt.IsZero() {
		f.record = store.BucketRecord{
			ID:        id,
			Name:      name,
			CreatedAt: createdAt,
			UpdatedAt: createdAt,
		}
	}

	return f.record, nil
}

func (f *fakeBucketStore) Calls() []bucketCreateCall {
	f.mu.Lock()
	defer f.mu.Unlock()

	calls := make([]bucketCreateCall, len(f.calls))
	copy(calls, f.calls)

	return calls
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
