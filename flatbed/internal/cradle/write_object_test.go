package cradle

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/ratdaddy/blockcloset/flatbed/internal/requestid"
)

func TestClientWriteObject(t *testing.T) {
	t.Parallel()

	const address = "localhost:9444"

	type tc struct {
		name             string
		objectID         string
		bucket           string
		size             int64
		body             string
		wantErr          bool
		wantBytesWritten int64
		wantChunkCount   int
	}

	cases := []tc{
		{
			name:             "successful write",
			objectID:         "01JXXXXXXXXXXXXXXXXXXXXXXXXX",
			bucket:           "photos",
			size:             11,
			body:             "hello world",
			wantBytesWritten: 11,
			wantChunkCount:   1,
		},
		{
			name:     "size mismatch returns error",
			objectID: "01JYYYYYYYYYYYYYYYYYYYYYYYYY",
			bucket:   "photos",
			size:     100,
			body:     "hello world",
			wantErr:  true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			client, svc := newTestClient(t)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			t.Cleanup(cancel)

			bytesWritten, committedAtMs, err := client.WriteObject(
				requestid.WithRequestID(ctx, "req-abc"),
				address, c.objectID, c.bucket, c.size, strings.NewReader(c.body),
			)

			if c.wantErr {
				if err == nil {
					t.Fatal("WriteObject returned nil error, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("WriteObject: %v", err)
			}

			if bytesWritten != c.wantBytesWritten {
				t.Fatalf("bytesWritten: got %d, want %d", bytesWritten, c.wantBytesWritten)
			}
			if committedAtMs != 1234567890 {
				t.Fatalf("committedAtMs: got %d, want 1234567890", committedAtMs)
			}

			call, ok := svc.LastWriteObjectCall()
			if !ok {
				t.Fatal("no WriteObject call recorded")
			}
			if call.ObjectID != c.objectID {
				t.Fatalf("ObjectID: got %q, want %q", call.ObjectID, c.objectID)
			}
			if call.Bucket != c.bucket {
				t.Fatalf("Bucket: got %q, want %q", call.Bucket, c.bucket)
			}
			if call.Size != c.size {
				t.Fatalf("Size: got %d, want %d", call.Size, c.size)
			}
			if len(call.Chunks) != c.wantChunkCount {
				t.Fatalf("chunk count: got %d, want %d", len(call.Chunks), c.wantChunkCount)
			}

			var receivedBody bytes.Buffer
			for _, chunk := range call.Chunks {
				receivedBody.Write(chunk)
			}
			if receivedBody.String() != c.body {
				t.Fatalf("received body: got %q, want %q", receivedBody.String(), c.body)
			}

			if meta := call.Metadata.Get("x-request-id"); len(meta) != 1 || meta[0] != "req-abc" {
				t.Fatalf("x-request-id = %v, want [req-abc]", meta)
			}

			svc.Reset()

			_, _, err = client.WriteObject(ctx, address, c.objectID, c.bucket, c.size, strings.NewReader(c.body))
			if err != nil {
				t.Fatalf("WriteObject (no request id): %v", err)
			}
			call, _ = svc.LastWriteObjectCall()
			if meta := call.Metadata.Get("x-request-id"); len(meta) != 0 {
				t.Fatalf("x-request-id without id = %v, want []", meta)
			}
		})
	}
}
