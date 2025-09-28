package gantry

import (
	"context"
	"testing"

	"github.com/ratdaddy/blockcloset/gateway/internal/httpapi"
)

func TestClientRequestIDPropagation(t *testing.T) {
	t.Parallel()

	client, svc := newTestClient(t)

	const bucketName = "case-bucket"

	cases := []struct {
		name string
		ctx  func(*testing.T) context.Context
		want []string
	}{
		{
			name: "with request id",
			ctx:  func(*testing.T) context.Context { return httpapi.WithRequestID(context.Background(), "req-123") },
			want: []string{"req-123"},
		},
		{
			name: "without request id",
			ctx:  func(*testing.T) context.Context { return context.Background() },
			want: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc.Reset()

			if _, err := client.CreateBucket(tc.ctx(t), bucketName); err != nil {
				t.Fatalf("CreateBucket: %v", err)
			}

			call, ok := svc.LastCreateBucketCall()
			if !ok {
				t.Fatalf("expected CreateBucket call")
			}

			got := call.Metadata.Get("x-request-id")

			if len(got) != len(tc.want) {
				t.Fatalf("request id count: got %d want %d (values %v)", len(got), len(tc.want), got)
			}
			for i := range tc.want {
				if got[i] != tc.want[i] {
					t.Fatalf("request id[%d] = %q, want %q", i, got[i], tc.want[i])
				}
			}
		})
	}
}
