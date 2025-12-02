package grpcsvc

import (
	"context"
	"errors"
	"io"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"

	servicev1 "github.com/ratdaddy/blockcloset/proto/gen/cradle/service/v1"
)

func TestService_WriteObject(t *testing.T) {
	t.Parallel()

	type tc struct {
		name        string
		requests    []*servicev1.WriteObjectRequest
		recvErr     error
		wantErr     bool
		wantCode    codes.Code
		wantMessage string
		wantBytes   int64
	}

	cases := []tc{
		{
			name: "receives metadata and chunks",
			requests: []*servicev1.WriteObjectRequest{
				newMetadataRequest("obj-123", "photos", 11),
				newChunkRequest([]byte("hello ")),
				newChunkRequest([]byte("world")),
			},
			wantBytes: 11,
		},
		{
			name: "first message must be metadata",
			requests: []*servicev1.WriteObjectRequest{
				newChunkRequest([]byte("hello")),
			},
			wantErr:     true,
			wantCode:    codes.InvalidArgument,
			wantMessage: "metadata must be the first message",
		},
		{
			name: "stream recv error surfaces",
			requests: []*servicev1.WriteObjectRequest{
				newMetadataRequest("obj-456", "docs", 5),
			},
			recvErr:     errors.New("stream recv failed"),
			wantErr:     true,
			wantCode:    codes.Internal,
			wantMessage: "stream recv failed",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			stream := newWriteObjectStreamFake(c.requests...)
			stream.recvErr = c.recvErr

			svc := New(newDiscardLogger())

			err := svc.WriteObject(stream)

			if c.wantErr {
				assertGRPCError(t, err, c.wantCode, c.wantMessage)
				if stream.response != nil {
					t.Fatal("SendAndClose called unexpectedly")
				}
				return
			}

			assertNoError(t, err)

			if stream.response == nil {
				t.Fatal("response not sent")
			}

			if got := stream.response.GetBytesWritten(); got != c.wantBytes {
				t.Fatalf("bytes_written: got %d, want %d", got, c.wantBytes)
			}

			if stream.response.GetCommittedAtMs() == 0 {
				t.Fatal("committed_at_ms not populated")
			}
		})
	}
}

func newMetadataRequest(objectID, bucket string, size int64) *servicev1.WriteObjectRequest {
	return &servicev1.WriteObjectRequest{
		Payload: &servicev1.WriteObjectRequest_Metadata{
			Metadata: &servicev1.WriteObjectMetadata{
				ObjectId: objectID,
				Bucket:   bucket,
				Size:     size,
			},
		},
	}
}

func newChunkRequest(chunk []byte) *servicev1.WriteObjectRequest {
	return &servicev1.WriteObjectRequest{
		Payload: &servicev1.WriteObjectRequest_Chunk{
			Chunk: chunk,
		},
	}
}

type writeObjectStreamFake struct {
	ctx       context.Context
	requests  []*servicev1.WriteObjectRequest
	response  *servicev1.WriteObjectResponse
	recvIndex int
	recvErr   error
}

func newWriteObjectStreamFake(reqs ...*servicev1.WriteObjectRequest) *writeObjectStreamFake {
	return &writeObjectStreamFake{
		ctx:      context.Background(),
		requests: reqs,
	}
}

func (f *writeObjectStreamFake) Recv() (*servicev1.WriteObjectRequest, error) {
	if f.recvIndex >= len(f.requests) {
		if f.recvErr != nil {
			err := f.recvErr
			f.recvErr = nil
			return nil, err
		}
		return nil, io.EOF
	}

	req := f.requests[f.recvIndex]
	f.recvIndex++
	return req, nil
}

func (f *writeObjectStreamFake) SendAndClose(resp *servicev1.WriteObjectResponse) error {
	f.response = resp
	return nil
}

func (f *writeObjectStreamFake) SetHeader(metadata.MD) error  { return nil }
func (f *writeObjectStreamFake) SendHeader(metadata.MD) error { return nil }
func (f *writeObjectStreamFake) SetTrailer(metadata.MD)       {}

func (f *writeObjectStreamFake) Context() context.Context {
	if f.ctx == nil {
		return context.Background()
	}
	return f.ctx
}

func (f *writeObjectStreamFake) SendMsg(any) error { return nil }
func (f *writeObjectStreamFake) RecvMsg(any) error { return nil }
