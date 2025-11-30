package gantry

import (
	"context"
	"testing"
	"time"
)

func TestClientPlanWrite(t *testing.T) {
	t.Parallel()

	client, svc := newTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	const (
		bucket = "test-bucket"
		key    = "test-key.txt"
		size   = int64(1024)
	)

	writePlan, err := client.PlanWrite(ctx, bucket, key, size)
	if err != nil {
		t.Fatalf("PlanWrite: %v", err)
	}

	if writePlan == nil {
		t.Fatal("PlanWrite returned nil WritePlan")
	}

	if writePlan.GetObjectId() == "" {
		t.Fatal("PlanWrite returned empty objectID")
	}

	if writePlan.GetCradleAddress() == "" {
		t.Fatal("PlanWrite returned empty cradleAddress")
	}

	calls := svc.PlanWriteCalls()
	if len(calls) != 1 {
		t.Fatalf("PlanWrite call count = %d, want 1", len(calls))
	}

	if calls[0].Request.GetBucket() != bucket {
		t.Fatalf("PlanWrite request Bucket = %q, want %q", calls[0].Request.GetBucket(), bucket)
	}

	if calls[0].Request.GetKey() != key {
		t.Fatalf("PlanWrite request Key = %q, want %q", calls[0].Request.GetKey(), key)
	}

	if calls[0].Request.GetSize() != size {
		t.Fatalf("PlanWrite request Size = %d, want %d", calls[0].Request.GetSize(), size)
	}
}
