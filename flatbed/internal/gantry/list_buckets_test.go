package gantry

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/ratdaddy/blockcloset/flatbed/internal/requestid"
	bucketv1 "github.com/ratdaddy/blockcloset/proto/gen/gantry/bucket/v1"
	servicev1 "github.com/ratdaddy/blockcloset/proto/gen/gantry/service/v1"
)

func TestClientListBuckets(t *testing.T) {
	t.Parallel()

	client, svc := newTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	svc.SetListBucketsHook(func(_ context.Context, _ *servicev1.ListBucketsRequest) (*servicev1.ListBucketsResponse, error) {
		return &servicev1.ListBucketsResponse{
			Buckets: []*bucketv1.Bucket{
				{Name: "alpha", CreatedAtRfc3339: "2025-01-01T01:02:03Z"},
				{Name: "beta", CreatedAtRfc3339: "2025-06-01T12:00:00Z"},
			},
		}, nil
	})

	want := []Bucket{
		{Name: "alpha", CreatedAt: parseTime(t, "2025-01-01T01:02:03Z")},
		{Name: "beta", CreatedAt: parseTime(t, "2025-06-01T12:00:00Z")},
	}

	got, err := client.ListBuckets(requestid.WithRequestID(ctx, "req-abc"))
	if err != nil {
		t.Fatalf("ListBuckets: %v", err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("ListBuckets diff (-want +got):\n%s", diff)
	}

	call, ok := svc.LastListBucketsCall()
	if !ok {
		t.Fatal("no ListBuckets call recorded")
	}
	if meta := call.Metadata.Get("x-request-id"); len(meta) != 1 || meta[0] != "req-abc" {
		t.Fatalf("x-request-id = %v, want [req-abc]", meta)
	}
}
