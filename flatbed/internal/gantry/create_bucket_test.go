package gantry

import (
	"context"
	"testing"
	"time"

	"github.com/ratdaddy/blockcloset/flatbed/internal/requestid"
)

func TestClientCreateBucket(t *testing.T) {
	t.Parallel()

	client, svc := newTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	const name = "test-bucket"

	gotName, err := client.CreateBucket(requestid.WithRequestID(ctx, "req-abc"), name)
	if err != nil {
		t.Fatalf("CreateBucket: %v", err)
	}
	if gotName != name {
		t.Fatalf("CreateBucket returned %q, want %q", gotName, name)
	}

	call, ok := svc.LastCreateBucketCall()
	if !ok {
		t.Fatal("no CreateBucket call recorded")
	}
	if call.Request.GetName() != name {
		t.Fatalf("request Name = %q, want %q", call.Request.GetName(), name)
	}
	if meta := call.Metadata.Get("x-request-id"); len(meta) != 1 || meta[0] != "req-abc" {
		t.Fatalf("x-request-id = %v, want [req-abc]", meta)
	}

	svc.Reset()

	if _, err := client.CreateBucket(ctx, name); err != nil {
		t.Fatalf("CreateBucket (no request id): %v", err)
	}
	call, _ = svc.LastCreateBucketCall()
	if meta := call.Metadata.Get("x-request-id"); len(meta) != 0 {
		t.Fatalf("x-request-id without id = %v, want []", meta)
	}
}
