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

func TestService_PlanWrite(t *testing.T) {
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
		wantMessage           string
		wantErrorDetail       bool
		wantErrorReason       servicev1.PlanWriteError_Reason
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
			wantMessage:         "bucket not found",
			wantErrorDetail:     true,
			wantErrorReason:     servicev1.PlanWriteError_REASON_BUCKET_NOT_FOUND,
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
			wantMessage:           "no cradle servers available",
			wantErrorDetail:       true,
			wantErrorReason:       servicev1.PlanWriteError_REASON_NO_CRADLE_SERVERS,
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
			wantMessage:           "object store error",
			expectGetByNameCall:   true,
			expectSelectForUpload: true,
			expectObjectCreate:    true,
		},
		{
			name:        "invalid bucket name returns InvalidArgument",
			bucket:      "Bad!Name",
			key:         "my-key.txt",
			size:        1024,
			wantErr:     true,
			wantCode:    codes.InvalidArgument,
			wantMessage: "InvalidBucketName",
		},
		{
			name:        "invalid key returns InvalidArgument",
			bucket:      "my-bucket",
			key:         "",
			size:        1024,
			wantErr:     true,
			wantCode:    codes.InvalidArgument,
			wantMessage: "InvalidKeyName",
		},
		{
			name:        "zero size returns InvalidArgument",
			bucket:      "my-bucket",
			key:         "my-key.txt",
			size:        0,
			wantErr:     true,
			wantCode:    codes.InvalidArgument,
			wantMessage: "InvalidArgument",
		},
		{
			name:        "size exceeds max returns InvalidArgument",
			bucket:      "my-bucket",
			key:         "my-key.txt",
			size:        5*1024*1024*1024 + 1, // 5GB + 1 byte
			wantErr:     true,
			wantCode:    codes.InvalidArgument,
			wantMessage: "EntityTooLarge",
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

			resp, err := svc.PlanWrite(context.Background(), &servicev1.PlanWriteRequest{
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
				assertGRPCError(t, err, c.wantCode, c.wantMessage)
				if c.wantErrorDetail {
					assertPlanWriteErrorDetail(t, err, c.wantErrorReason, c.bucket)
				}
				return
			}

			assertNoError(t, err)

			if c.expectSelectForUpload {
				if cradles.SelectForUploadCallCount() != 1 {
					t.Fatalf("SelectForUpload calls: got %d, want 1", cradles.SelectForUploadCallCount())
				}
			}

			writePlan := resp.GetWritePlan()
			if writePlan == nil {
				t.Fatal("response.write_plan is nil")
			}

			if c.wantCradleAddress != "" && writePlan.GetCradleAddress() != c.wantCradleAddress {
				t.Fatalf("cradle_address: got %q, want %q", writePlan.GetCradleAddress(), c.wantCradleAddress)
			}

			if c.wantObjectID {
				if _, err := ulid.Parse(writePlan.GetObjectId()); err != nil {
					t.Fatalf("response object_id %q not a valid ULID: %v", writePlan.GetObjectId(), err)
				}
			}

			if c.expectObjectCreate {
				calls := objects.Calls()
				if len(calls) != 1 {
					t.Fatalf("Objects().CreatePending calls: got %d, want 1", len(calls))
				}

				call := calls[0]

				// Verify object_id was passed to CreatePending and matches response
				if c.wantObjectID && call.ID != writePlan.GetObjectId() {
					t.Fatalf("CreatePending object_id: got %q, want %q (from response)", call.ID, writePlan.GetObjectId())
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

func assertPlanWriteErrorDetail(t *testing.T, err error, wantReason servicev1.PlanWriteError_Reason, wantBucket string) {
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
		planErr, ok := detail.(*servicev1.PlanWriteError)
		if !ok {
			continue
		}

		if planErr.GetReason() != wantReason {
			t.Fatalf("PlanWriteError reason: got %v, want %v", planErr.GetReason(), wantReason)
		}

		if planErr.GetBucket() != wantBucket {
			t.Fatalf("PlanWriteError bucket: got %q, want %q", planErr.GetBucket(), wantBucket)
		}

		return
	}

	t.Fatalf("status missing PlanWriteError detail: %v", err)
}
