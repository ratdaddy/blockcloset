package main

import (
	"testing"

	"google.golang.org/protobuf/proto"

	servicev1 "github.com/ratdaddy/blockcloset/proto/gen/cradle/service/v1"
)

func TestWriteObjectResponseProtoRoundTrip(t *testing.T) {
	in := &servicev1.WriteObjectResponse{
		BytesWritten:  1024,
		CommittedAtMs: 1234567890,
	}

	b, err := proto.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out servicev1.WriteObjectResponse
	if err := proto.Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.GetBytesWritten() != in.GetBytesWritten() || out.GetCommittedAtMs() != in.GetCommittedAtMs() {
		t.Fatalf("mismatch: got=%+v want=%+v", &out, in)
	}
}
