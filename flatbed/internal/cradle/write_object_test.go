package cradle

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
)

func TestClient_WriteObject(t *testing.T) {
	t.Parallel()

	type tc struct {
		name             string
		address          string
		objectID         string
		bucket           string
		size             int64
		body             string
		wantErr          bool
		wantBytesWritten int64
		wantMetadata     bool
		wantChunkCount   int
	}

	cases := []tc{
		{
			name:             "successful write sends metadata and chunks",
			address:          "localhost:9444",
			objectID:         "obj-123",
			bucket:           "photos",
			size:             11,
			body:             "hello world",
			wantBytesWritten: 11,
			wantMetadata:     true,
			wantChunkCount:   1,
		},
		{
			name:     "size mismatch returns error",
			address:  "localhost:9444",
			objectID: "obj-456",
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

			bodyReader := strings.NewReader(c.body)

			bytesWritten, committedAtMs, err := client.WriteObject(ctx, c.address, c.objectID, c.bucket, c.size, bodyReader)

			if c.wantErr && err == nil {
				t.Fatal("WriteObject returned nil error, want error")
			}
			if !c.wantErr && err != nil {
				t.Fatalf("WriteObject returned error: %v", err)
			}

			if c.wantErr {
				return
			}

			if bytesWritten != c.wantBytesWritten {
				t.Fatalf("bytesWritten: got %d, want %d", bytesWritten, c.wantBytesWritten)
			}

			if committedAtMs == 0 {
				t.Fatal("committedAtMs is 0, want non-zero timestamp")
			}

			calls := svc.WriteObjectCalls()
			if len(calls) != 1 {
				t.Fatalf("WriteObject call count: got %d, want 1", len(calls))
			}

			call := calls[0]

			if c.wantMetadata {
				if call.ObjectID != c.objectID {
					t.Fatalf("call.ObjectID: got %q, want %q", call.ObjectID, c.objectID)
				}
				if call.Bucket != c.bucket {
					t.Fatalf("call.Bucket: got %q, want %q", call.Bucket, c.bucket)
				}
				if call.Size != c.size {
					t.Fatalf("call.Size: got %d, want %d", call.Size, c.size)
				}
			}

			if len(call.Chunks) != c.wantChunkCount {
				t.Fatalf("chunk count: got %d, want %d", len(call.Chunks), c.wantChunkCount)
			}

			// Verify all chunks concatenated match the body
			var receivedBody bytes.Buffer
			for _, chunk := range call.Chunks {
				receivedBody.Write(chunk)
			}

			if receivedBody.String() != c.body {
				t.Fatalf("received body: got %q, want %q", receivedBody.String(), c.body)
			}
		})
	}
}
