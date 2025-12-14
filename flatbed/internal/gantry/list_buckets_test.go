package gantry

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	bucketv1 "github.com/ratdaddy/blockcloset/proto/gen/gantry/bucket/v1"
	servicev1 "github.com/ratdaddy/blockcloset/proto/gen/gantry/service/v1"
)

func TestClientListBuckets(t *testing.T) {
	t.Parallel()

	type tc struct {
		name        string
		ctx         context.Context
		setup       func(*captureGantryService)
		wantBuckets []Bucket
		wantErr     bool
	}

	cases := []tc{
		{
			name: "hydrates buckets from response",
			ctx:  context.Background(),
			setup: func(svc *captureGantryService) {
				resp := &servicev1.ListBucketsResponse{
					Buckets: []*bucketv1.Bucket{
						{Name: "first", CreatedAtRfc3339: "2025-01-01T01:02:03Z"},
						{Name: "second", CreatedAtRfc3339: "2025-01-02T04:05:06Z"},
					},
				}
				svc.SetListBucketsHook(func(context.Context, *servicev1.ListBucketsRequest) (*servicev1.ListBucketsResponse, error) {
					return resp, nil
				})
			},
			wantBuckets: []Bucket{
				{Name: "first", CreatedAt: parseTime(t, "2025-01-01T01:02:03Z")},
				{Name: "second", CreatedAt: parseTime(t, "2025-01-02T04:05:06Z")},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client, svc := newTestClient(t)
			if tc.setup != nil {
				tc.setup(svc)
			}

			out, err := client.ListBuckets(tc.ctx)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("ListBuckets: expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("ListBuckets: %v", err)
			}

			if diff := cmp.Diff(tc.wantBuckets, out); diff != "" {
				t.Fatalf("ListBuckets diff (-want +got):\n%s", diff)
			}
		})
	}
}

func parseTime(t *testing.T, v string) time.Time {
	t.Helper()
	ts, err := time.Parse(time.RFC3339, v)
	if err != nil {
		t.Fatalf("parse time %q: %v", v, err)
	}
	return ts
}
