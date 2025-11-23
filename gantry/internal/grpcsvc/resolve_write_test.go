package grpcsvc

import (
	"context"
	"testing"

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
		getByNameErr          error
		wantErr               bool
		wantCode              codes.Code
		wantErrorDetail       bool
		wantErrorReason       servicev1.ResolveWriteError_Reason
		wantObjectID          bool
		wantCradleAddress     bool
		expectGetByNameCall   bool
	}

	cases := []tc{
		{
			name:                  "valid request returns object_id and cradle_address",
			bucket:                "my-bucket",
			key:                   "my-key.txt",
			size:                  1024,
			wantObjectID:          true,
			wantCradleAddress:     true,
			expectGetByNameCall:   true,
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
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			logger := newDiscardLogger()
			svc := New(logger, nil)

			buckets := testutil.NewFakeBucketStore()
			if c.getByNameErr != nil {
				buckets.SetGetByNameError(c.getByNameErr)
			}
			svc.store = testutil.NewFakeStore(buckets, nil)

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

			if c.wantObjectID && resp.GetObjectId() == "" {
				t.Fatal("expected non-empty object_id")
			}

			if c.wantCradleAddress && resp.GetCradleAddress() == "" {
				t.Fatal("expected non-empty cradle_address")
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
