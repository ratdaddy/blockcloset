package gantry

import (
	"context"
	"testing"
	"time"
)

func TestClientResolveWrite(t *testing.T) {
	t.Parallel()

	client, svc := newTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	const (
		bucket = "test-bucket"
		key    = "test-key.txt"
		size   = int64(1024)
	)

	objectID, cradleAddress, err := client.ResolveWrite(ctx, bucket, key, size)
	if err != nil {
		t.Fatalf("ResolveWrite: %v", err)
	}

	if objectID == "" {
		t.Fatal("ResolveWrite returned empty objectID")
	}

	if cradleAddress == "" {
		t.Fatal("ResolveWrite returned empty cradleAddress")
	}

	calls := svc.ResolveWriteCalls()
	if len(calls) != 1 {
		t.Fatalf("ResolveWrite call count = %d, want 1", len(calls))
	}

	if calls[0].Request.GetBucket() != bucket {
		t.Fatalf("ResolveWrite request Bucket = %q, want %q", calls[0].Request.GetBucket(), bucket)
	}

	if calls[0].Request.GetKey() != key {
		t.Fatalf("ResolveWrite request Key = %q, want %q", calls[0].Request.GetKey(), key)
	}

	if calls[0].Request.GetSize() != size {
		t.Fatalf("ResolveWrite request Size = %d, want %d", calls[0].Request.GetSize(), size)
	}
}
