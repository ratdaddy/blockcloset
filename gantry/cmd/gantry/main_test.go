package main

import (
	"testing"

	"google.golang.org/protobuf/proto"

	bucketv1 "github.com/ratdaddy/blockcloset/proto/gen/gantry/bucket/v1"
)

func TestBucketProtoRoundTrip(t *testing.T) {
	in := &bucketv1.Bucket{
		Name:             "alpha",
		CreatedAtRfc3339: "2025-08-26T00:00:00Z",
	}

	b, err := proto.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out bucketv1.Bucket
	if err := proto.Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.GetName() != in.GetName() || out.GetCreatedAtRfc3339() != in.GetCreatedAtRfc3339() {
		t.Fatalf("mismatch: got=%+v want=%+v", &out, in)
	}
}
