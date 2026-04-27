package gantry

import (
	"context"
	"testing"
	"time"

	"github.com/ratdaddy/blockcloset/flatbed/internal/requestid"
)

func TestClientCommitObject(t *testing.T) {
	t.Parallel()

	client, svc := newTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	const (
		objectID         = "test-object-id"
		size      int64  = 1024
		lastModMs int64  = 1234567890000
	)

	if err := client.CommitObject(requestid.WithRequestID(ctx, "req-abc"), objectID, size, lastModMs); err != nil {
		t.Fatalf("CommitObject: %v", err)
	}

	call, ok := svc.LastCommitObjectCall()
	if !ok {
		t.Fatal("no CommitObject call recorded")
	}
	if call.Request.GetObjectId() != objectID {
		t.Fatalf("request ObjectId = %q, want %q", call.Request.GetObjectId(), objectID)
	}
	if call.Request.GetSize() != size {
		t.Fatalf("request Size = %d, want %d", call.Request.GetSize(), size)
	}
	if call.Request.GetLastModifiedMs() != lastModMs {
		t.Fatalf("request LastModifiedMs = %d, want %d", call.Request.GetLastModifiedMs(), lastModMs)
	}
	if meta := call.Metadata.Get("x-request-id"); len(meta) != 1 || meta[0] != "req-abc" {
		t.Fatalf("x-request-id = %v, want [req-abc]", meta)
	}
}
