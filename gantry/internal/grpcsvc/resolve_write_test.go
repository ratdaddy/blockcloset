package grpcsvc

import (
	"context"
	"errors"
	"testing"

	"github.com/oklog/ulid/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ratdaddy/blockcloset/gantry/internal/store"
	"github.com/ratdaddy/blockcloset/gantry/internal/testutil"
	servicev1 "github.com/ratdaddy/blockcloset/proto/gen/gantry/service/v1"
)

func TestService_ResolveWrite(t *testing.T) {
	t.Parallel()

	type tc struct {
		name                  string
		bucket                string
		key                   string
		size                  int64
		bucketID              string
		getByNameErr          error
		cradleID              string
		cradleAddress         string
		selectForUploadErr    error
		objectCreateErr       error
		wantErr               bool
		wantCode              codes.Code
		wantErrorDetail       bool
		wantErrorReason       servicev1.ResolveWriteError_Reason
		wantObjectID          bool
		wantCradleAddress     string
		expectGetByNameCall   bool
		expectSelectForUpload bool
		expectObjectCreate    bool
	}

	cases := []tc{
		{
			name:                  "valid request returns object_id and cradle_address",
			bucket:                "my-bucket",
			key:                   "my-key.txt",
			size:                  1024,
			bucketID:              "bucket-id-123",
			cradleID:              "cradle-id-456",
			cradleAddress:         "127.0.0.1:9444",
			wantObjectID:          true,
			wantCradleAddress:     "127.0.0.1:9444",
			expectGetByNameCall:   true,
			expectSelectForUpload: true,
			expectObjectCreate:    true,
		},
		{
			name:                "bucket not found returns NotFound",
			bucket:              "nonexistent-bucket",
			key:                 "my-key.txt",
			size:                1024,
			getByNameErr:        store.ErrBucketNotFound,
			wantErr:             true,
			wantCode:            codes.NotFound,
			wantErrorDetail:     true,
			wantErrorReason:     servicev1.ResolveWriteError_REASON_BUCKET_NOT_FOUND,
			expectGetByNameCall: true,
		},
		{
			name:                  "no cradle servers returns FailedPrecondition",
			bucket:                "my-bucket",
			key:                   "my-key.txt",
			size:                  1024,
			selectForUploadErr:    store.ErrNoCradleServersAvailable,
			wantErr:               true,
			wantCode:              codes.FailedPrecondition,
			wantErrorDetail:       true,
			wantErrorReason:       servicev1.ResolveWriteError_REASON_NO_CRADLE_SERVERS,
			expectGetByNameCall:   true,
			expectSelectForUpload: true,
		},
		{
			name:                  "object store error returns Internal",
			bucket:                "my-bucket",
			key:                   "my-key.txt",
			size:                  1024,
			bucketID:              "bucket-id-123",
			cradleID:              "cradle-id-456",
			cradleAddress:         "127.0.0.1:9444",
			objectCreateErr:       errors.New("object store error"),
			wantErr:               true,
			wantCode:              codes.Internal,
			expectGetByNameCall:   true,
			expectSelectForUpload: true,
			expectObjectCreate:    true,
		},
		{
			name:     "invalid bucket name returns InvalidArgument",
			bucket:   "Bad!Name",
			key:      "my-key.txt",
			size:     1024,
			wantErr:  true,
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "invalid key returns InvalidArgument",
			bucket:   "my-bucket",
			key:      "",
			size:     1024,
			wantErr:  true,
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "zero size returns InvalidArgument",
			bucket:   "my-bucket",
			key:      "my-key.txt",
			size:     0,
			wantErr:  true,
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "size exceeds max returns InvalidArgument",
			bucket:   "my-bucket",
			key:      "my-key.txt",
			size:     5*1024*1024*1024 + 1, // 5GB + 1 byte
			wantErr:  true,
			wantCode: codes.InvalidArgument,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			logger := newDiscardLogger()
			svc := New(logger, nil)

			buckets := testutil.NewFakeBucketStore()
			if c.getByNameErr != nil {
				buckets.SetGetByNameError(c.getByNameErr)
			} else if c.bucketID != "" {
				buckets.SetGetByNameResponse(store.BucketRecord{
					ID:   c.bucketID,
					Name: c.bucket,
				})
			}

			cradles := testutil.NewFakeCradleStore()
			if c.cradleAddress != "" {
				cradles.SetSelectForUploadResponse(store.CradleServerRecord{
					ID:      c.cradleID,
					Address: c.cradleAddress,
				})
			}
			if c.selectForUploadErr != nil {
				cradles.SetSelectForUploadError(c.selectForUploadErr)
			}

			objects := testutil.NewFakeObjectStore()
			if c.objectCreateErr != nil {
				objects.SetCreateError(c.objectCreateErr)
			}

			svc.store = testutil.NewFakeStore(
				testutil.WithBuckets(buckets),
				testutil.WithCradles(cradles),
				testutil.WithObjects(objects),
			)

			resp, err := svc.ResolveWrite(context.Background(), &servicev1.ResolveWriteRequest{
				Bucket: c.bucket,
				Key:    c.key,
				Size:   c.size,
			})

			if c.expectGetByNameCall {
				calls := buckets.GetByNameCalls()
				if len(calls) != 1 {
					t.Fatalf("GetByName calls: got %d, want 1", len(calls))
				}
				if calls[0] != c.bucket {
					t.Fatalf("GetByName bucket: got %q, want %q", calls[0], c.bucket)
				}
			}

			if c.wantErr {
				assertResolveWriteError(t, err, c.wantCode)
				if c.wantErrorDetail {
					assertResolveWriteErrorDetail(t, err, c.wantErrorReason, c.bucket)
				}
				return
			}

			assertNoError(t, err)

			if c.expectSelectForUpload {
				if cradles.SelectForUploadCallCount() != 1 {
					t.Fatalf("SelectForUpload calls: got %d, want 1", cradles.SelectForUploadCallCount())
				}
			}

			if c.wantCradleAddress != "" && resp.GetCradleAddress() != c.wantCradleAddress {
				t.Fatalf("cradle_address: got %q, want %q", resp.GetCradleAddress(), c.wantCradleAddress)
			}

			if c.wantObjectID {
				if _, err := ulid.Parse(resp.GetObjectId()); err != nil {
					t.Fatalf("response object_id %q not a valid ULID: %v", resp.GetObjectId(), err)
				}
			}

			if c.expectObjectCreate {
				calls := objects.Calls()
				if len(calls) != 1 {
					t.Fatalf("Objects().CreatePending calls: got %d, want 1", len(calls))
				}

				call := calls[0]

				// Verify object_id was passed to CreatePending and matches response
				if c.wantObjectID && call.ID != resp.GetObjectId() {
					t.Fatalf("CreatePending object_id: got %q, want %q (from response)", call.ID, resp.GetObjectId())
				}

				if call.BucketID != c.bucketID {
					t.Fatalf("CreatePending bucket_id: got %q, want %q", call.BucketID, c.bucketID)
				}
				if call.Key != c.key {
					t.Fatalf("CreatePending key: got %q, want %q", call.Key, c.key)
				}
				if call.SizeExpected != c.size {
					t.Fatalf("CreatePending size: got %d, want %d", call.SizeExpected, c.size)
				}
				if call.CradleServerID != c.cradleID {
					t.Fatalf("CreatePending cradle_server_id: got %q, want %q", call.CradleServerID, c.cradleID)
				}

				if call.CreatedAt.IsZero() {
					t.Fatal("CreatePending createdAt timestamp not populated")
				}
			}
		})
	}
}

func assertResolveWriteError(t *testing.T, err error, wantCode codes.Code) {
	t.Helper()

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %v", err)
	}

	if st.Code() != wantCode {
		t.Fatalf("status code: got %v, want %v", st.Code(), wantCode)
	}
}

func assertResolveWriteErrorDetail(t *testing.T, err error, wantReason servicev1.ResolveWriteError_Reason, wantBucket string) {
	t.Helper()

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %v", err)
	}

	details := st.Details()
	if len(details) == 0 {
		t.Fatalf("status missing details: %v", err)
	}

	for _, detail := range details {
		resolveErr, ok := detail.(*servicev1.ResolveWriteError)
		if !ok {
			continue
		}

		if resolveErr.GetReason() != wantReason {
			t.Fatalf("ResolveWriteError reason: got %v, want %v", resolveErr.GetReason(), wantReason)
		}

		if resolveErr.GetBucket() != wantBucket {
			t.Fatalf("ResolveWriteError bucket: got %q, want %q", resolveErr.GetBucket(), wantBucket)
		}

		return
	}

	t.Fatalf("status missing ResolveWriteError detail: %v", err)
}
