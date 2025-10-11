package gantry

import (
	"context"
	"testing"

	"google.golang.org/grpc/metadata"

	"github.com/ratdaddy/blockcloset/gateway/internal/requestid"
)

func TestClientRequestIDPropagation(t *testing.T) {
	t.Parallel()

	type call struct {
		name   string
		invoke func(*Client, context.Context) error
		last   func(*captureGantryService) (metadata.MD, bool)
	}

	cases := []call{
		{
			name: "CreateBucket",
			invoke: func(client *Client, ctx context.Context) error {
				_, err := client.CreateBucket(ctx, "bucket-one")
				return err
			},
			last: func(svc *captureGantryService) (metadata.MD, bool) {
				call, ok := svc.LastCreateBucketCall()
				if !ok {
					return nil, false
				}
				return call.Metadata, true
			},
		},
		{
			name: "ListBuckets",
			invoke: func(client *Client, ctx context.Context) error {
				_, err := client.ListBuckets(ctx)
				return err
			},
			last: func(svc *captureGantryService) (metadata.MD, bool) {
				call, ok := svc.LastListBucketsCall()
				if !ok {
					return nil, false
				}
				return call.Metadata, true
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client, svc := newTestClient(t)

			ctxWithID := requestid.WithRequestID(context.Background(), "req-123")
			if err := tc.invoke(client, ctxWithID); err != nil {
				t.Fatalf("%s with request id: %v", tc.name, err)
			}

			md, ok := tc.last(svc)
			if !ok {
				t.Fatalf("%s: expected call to be recorded", tc.name)
			}
			meta := md.Get("x-request-id")
			if len(meta) != 1 || meta[0] != "req-123" {
				t.Fatalf("%s metadata = %v, want [req-123]", tc.name, meta)
			}

			if err := tc.invoke(client, context.Background()); err != nil {
				t.Fatalf("%s without request id: %v", tc.name, err)
			}

			md, ok = tc.last(svc)
			if !ok {
				t.Fatalf("%s: expected call to be recorded", tc.name)
			}
			if meta := md.Get("x-request-id"); len(meta) != 0 {
				t.Fatalf("%s metadata without id = %v, want []", tc.name, meta)
			}
		})
	}
}
