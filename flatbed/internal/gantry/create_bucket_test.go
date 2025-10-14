package gantry

import (
	"context"
	"testing"
	"time"
)

func TestClientCreateBucket(t *testing.T) {
	t.Parallel()

	client, svc := newTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	const bucketName = "client-create-bucket"

	gotName, err := client.CreateBucket(ctx, bucketName)
	if err != nil {
		t.Fatalf("CreateBucket: %v", err)
	}

	if gotName != bucketName {
		t.Fatalf("CreateBucket returned %q, want %q", gotName, bucketName)
	}

	calls := svc.CreateBucketCalls()
	if len(calls) != 1 {
		t.Fatalf("CreateBucket call count = %d, want 1", len(calls))
	}

	if calls[0].Request.GetName() != bucketName {
		t.Fatalf("CreateBucket request Name = %q, want %q", calls[0].Request.GetName(), bucketName)
	}
}
