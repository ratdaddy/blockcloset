package grpcsvc

import (
	"context"
	"errors"
	"testing"
	"time"

	"google.golang.org/grpc/codes"

	"github.com/ratdaddy/blockcloset/gantry/internal/store"
	servicev1 "github.com/ratdaddy/blockcloset/proto/gen/gantry/service/v1"
)

func TestService_ListBuckets(t *testing.T) {
	t.Parallel()

	base := time.Date(2025, time.January, 1, 12, 0, 0, 0, time.UTC)

	type tc struct {
		name            string
		records         []store.BucketRecord
		listErr         error
		wantErr         bool
		wantCode        codes.Code
		wantMessage     string
		wantNames       []string
		wantTimestamps  []string
		expectStoreCall bool
	}

	cases := []tc{
		{
			name:            "empty bucket list returns empty response",
			records:         nil,
			wantNames:       nil,
			wantTimestamps:  nil,
			expectStoreCall: true,
		},
		{
			name: "returns buckets preserving store order",
			records: []store.BucketRecord{
				newBucketRecord("first-bucket", base.Add(-1*time.Hour)),
				newBucketRecord("middle-bucket", base),
				newBucketRecord("second-bucket", base.Add(2*time.Hour)),
			},
			wantNames: []string{"first-bucket", "middle-bucket", "second-bucket"},
			wantTimestamps: []string{
				formatBucketTimestamp(base.Add(-1 * time.Hour)),
				formatBucketTimestamp(base),
				formatBucketTimestamp(base.Add(2 * time.Hour)),
			},
			expectStoreCall: true,
		},
		{
			name:            "store error surfaces as internal",
			records:         nil,
			listErr:         errors.New("list buckets failed"),
			wantErr:         true,
			wantCode:        codes.Internal,
			wantMessage:     "list buckets failed",
			expectStoreCall: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			logger := newDiscardLogger()
			svc := New(logger, nil)

			buckets := newFakeBucketStore()
			buckets.SetListRecords(c.records)
			buckets.SetListError(c.listErr)

			svc.store = newFakeStore(buckets)

			resp, err := svc.ListBuckets(context.Background(), &servicev1.ListBucketsRequest{})

			assertListInvocation(t, buckets, c.expectStoreCall)

			if c.wantErr {
				assertGRPCError(t, err, c.wantCode, c.wantMessage)
				if resp != nil {
					t.Fatalf("response: got %#v, want nil", resp)
				}
				return
			}

			assertNoError(t, err)
			assertListBucketsResponse(t, resp, c.wantNames, c.wantTimestamps)
		})
	}
}

func newBucketRecord(name string, createdAt time.Time) store.BucketRecord {
	return store.BucketRecord{
		ID:        name + "-id",
		Name:      name,
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	}
}

func assertListBucketsResponse(t *testing.T, resp *servicev1.ListBucketsResponse, wantNames, wantTimestamps []string) {
	t.Helper()

	if resp == nil {
		t.Fatal("response: got nil")
	}

	buckets := resp.GetBuckets()
	if len(buckets) != len(wantNames) {
		t.Fatalf("bucket count: got %d, want %d", len(buckets), len(wantNames))
	}

	for i, b := range buckets {
		if b.GetName() != wantNames[i] {
			t.Fatalf("bucket[%d] name: got %q, want %q", i, b.GetName(), wantNames[i])
		}

		if wantTimestamps != nil && b.GetCreatedAtRfc3339() != wantTimestamps[i] {
			t.Fatalf("bucket[%d] created_at_rfc3339: got %q, want %q", i, b.GetCreatedAtRfc3339(), wantTimestamps[i])
		}
	}
}

func formatBucketTimestamp(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05.000000Z")
}

func assertListInvocation(t *testing.T, s *bucketStoreFake, wantCall bool) {
	t.Helper()

	got := s.ListCallCount()
	if wantCall {
		if got != 1 {
			t.Fatalf("bucket store list calls: got %d, want 1", got)
		}
		return
	}
	if got != 0 {
		t.Fatalf("bucket store list calls: got %d, want 0", got)
	}
}
