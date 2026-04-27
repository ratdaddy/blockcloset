package gantry

import (
	"context"
	"testing"
	"time"

	"github.com/ratdaddy/blockcloset/flatbed/internal/requestid"
	servicev1 "github.com/ratdaddy/blockcloset/proto/gen/gantry/service/v1"
	writeplanv1 "github.com/ratdaddy/blockcloset/proto/gen/gantry/write_plan/v1"
)

func TestClientPlanWrite(t *testing.T) {
	t.Parallel()

	client, svc := newTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	svc.SetPlanWriteHook(func(_ context.Context, _ *servicev1.PlanWriteRequest) (*servicev1.PlanWriteResponse, error) {
		return &servicev1.PlanWriteResponse{
			WritePlan: &writeplanv1.WritePlan{
				ObjectId:      "01JXXXXXXXXXXXXXXXXXXXXXXXXX",
				CradleAddress: "cradle.internal:9002",
			},
		}, nil
	})

	const (
		bucket = "photos"
		key    = "vacation/sunset.jpg"
		size   = int64(204800)
	)

	plan, err := client.PlanWrite(requestid.WithRequestID(ctx, "req-abc"), bucket, key, size)
	if err != nil {
		t.Fatalf("PlanWrite: %v", err)
	}
	if plan.GetObjectId() != "01JXXXXXXXXXXXXXXXXXXXXXXXXX" {
		t.Fatalf("ObjectId = %q, want %q", plan.GetObjectId(), "01JXXXXXXXXXXXXXXXXXXXXXXXXX")
	}
	if plan.GetCradleAddress() != "cradle.internal:9002" {
		t.Fatalf("CradleAddress = %q, want %q", plan.GetCradleAddress(), "cradle.internal:9002")
	}

	call, ok := svc.LastPlanWriteCall()
	if !ok {
		t.Fatal("no PlanWrite call recorded")
	}
	if call.Request.GetBucket() != bucket {
		t.Fatalf("request Bucket = %q, want %q", call.Request.GetBucket(), bucket)
	}
	if call.Request.GetKey() != key {
		t.Fatalf("request Key = %q, want %q", call.Request.GetKey(), key)
	}
	if call.Request.GetSize() != size {
		t.Fatalf("request Size = %d, want %d", call.Request.GetSize(), size)
	}
	if meta := call.Metadata.Get("x-request-id"); len(meta) != 1 || meta[0] != "req-abc" {
		t.Fatalf("x-request-id = %v, want [req-abc]", meta)
	}
}
