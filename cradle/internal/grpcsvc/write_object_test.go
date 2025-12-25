package grpcsvc

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"

	"github.com/ratdaddy/blockcloset/cradle/internal/storage"
	servicev1 "github.com/ratdaddy/blockcloset/proto/gen/cradle/service/v1"
)

func TestService_WriteObject(t *testing.T) {
	t.Parallel()

	type tc struct {
		name                  string
		requests              []*servicev1.WriteObjectRequest
		recvErr               error
		newWriterErr          error // if set, inject a newWriter that returns this error
		closeWriterFile       bool  // if true, inject a newWriter that closes the file before returning (causes Write to fail)
		invalidFinalPath      bool  // if true, inject a newWriter with invalid finalPath (causes Commit to fail)
		wantErr               bool
		wantCode              codes.Code
		wantMessage           string // check message contains this substring (works for exact matches too)
		wantBytes             int64
		wantTempFileInBucket  string // if set, verify a temp file (starts with .) exists in this bucket
		wantTempFileContent   string // if set, verify temp file contains this data
		wantFinalFileInBucket string // if set, verify a non-temp file exists in this bucket
		wantFinalContent      string // if set, verify final file contains this data
		wantNoTempFiles       bool   // if true, verify no temp files exist in bucket
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
			name: "stream recv error",
			requests: []*servicev1.WriteObjectRequest{
				newMetadataRequest("obj-123", "documents", 12),
				newChunkRequest([]byte("test ")),
				newChunkRequest([]byte("content")),
			},
			recvErr:         errors.New("stream interrupted"),
			wantErr:         true,
			wantCode:        codes.Internal,
			wantMessage:     "stream interrupted",
			wantNoTempFiles: true,
		},
		{
			name: "writes to storage",
			requests: []*servicev1.WriteObjectRequest{
				newMetadataRequest("obj-789", "photos", 11),
				newChunkRequest([]byte("hello ")),
				newChunkRequest([]byte("world")),
			},
			wantBytes:             11,
			wantFinalFileInBucket: "photos",
			wantFinalContent:      "hello world",
			wantNoTempFiles:       true,
		},
		{
			name: "NewWriter error",
			requests: []*servicev1.WriteObjectRequest{
				newMetadataRequest("obj-999", "invalid", 0),
			},
			newWriterErr: errors.New("mkdir /data/invalid: permission denied"),
			wantErr:      true,
			wantCode:     codes.Internal,
			wantMessage:  "mkdir /data/invalid: permission denied",
		},
		{
			name: "Write error",
			requests: []*servicev1.WriteObjectRequest{
				newMetadataRequest("obj-555", "videos", 10),
				newChunkRequest([]byte("test data")),
			},
			closeWriterFile: true,
			wantErr:         true,
			wantCode:        codes.Internal,
			wantMessage:     "file already closed",
			wantNoTempFiles: true,
		},
		{
			name: "Commit error",
			requests: []*servicev1.WriteObjectRequest{
				newMetadataRequest("obj-777", "music", 7),
				newChunkRequest([]byte("test123")),
			},
			invalidFinalPath: true,
			wantErr:          true,
			wantCode:         codes.Internal,
			wantMessage:      "no such file or directory",
			wantNoTempFiles:  true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			objectsRoot := t.TempDir()

			stream := newWriteObjectStreamFake(c.requests...)
			stream.recvErr = c.recvErr

			svc := New(newDiscardLogger())
			svc.objectsRoot = objectsRoot

			// Inject mock newWriter if error should be returned
			if c.newWriterErr != nil {
				svc.newWriter = func(objRoot, bucket, objectID string) (*storage.Writer, error) {
					return nil, c.newWriterErr
				}
			}

			// Inject mock newWriter that closes file (causes Write to fail)
			if c.closeWriterFile {
				svc.newWriter = func(objRoot, bucket, objectID string) (*storage.Writer, error) {
					w, err := storage.NewWriter(objRoot, bucket, objectID)
					if err != nil {
						return nil, err
					}
					// Close the file so Write will fail
					w.File.Close()
					return w, nil
				}
			}

			// Inject mock newWriter with invalid finalPath (causes Commit to fail)
			if c.invalidFinalPath {
				svc.newWriter = func(objRoot, bucket, objectID string) (*storage.Writer, error) {
					w, err := storage.NewWriter(objRoot, bucket, objectID)
					if err != nil {
						return nil, err
					}
					// Set FinalPath to non-existent directory so rename will fail
					w.FinalPath = filepath.Join(objRoot, "nonexistent", objectID)
					return w, nil
				}
			}

			err := svc.WriteObject(stream)

			// Verify no temp files exist in bucket if expected (check before error handling early return)
			if c.wantNoTempFiles {
				bucket := c.wantFinalFileInBucket
				if bucket == "" {
					bucket = c.wantTempFileInBucket
				}
				// Extract bucket from first request metadata if not otherwise specified
				if bucket == "" && len(c.requests) > 0 {
					if meta := c.requests[0].GetMetadata(); meta != nil {
						bucket = meta.GetBucket()
					}
				}
				if bucket != "" {
					entries, err := os.ReadDir(filepath.Join(objectsRoot, bucket))
					if err != nil && !os.IsNotExist(err) {
						t.Fatalf("failed to read bucket %s: %v", bucket, err)
					}
					if err == nil {
						for _, entry := range entries {
							if strings.HasPrefix(entry.Name(), ".") && !entry.IsDir() {
								t.Fatalf("found temp file in bucket %s: %s", bucket, entry.Name())
							}
						}
					}
				}
			}

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

			// Verify temp file exists in bucket if expected
			if c.wantTempFileInBucket != "" {
				entries, err := os.ReadDir(filepath.Join(objectsRoot, c.wantTempFileInBucket))
				if err != nil {
					t.Fatalf("failed to read bucket %s: %v", c.wantTempFileInBucket, err)
				}

				var tempFileContent []byte
				for _, entry := range entries {
					if strings.HasPrefix(entry.Name(), ".") && !entry.IsDir() {
						tempFileContent, err = os.ReadFile(filepath.Join(objectsRoot, c.wantTempFileInBucket, entry.Name()))
						if err != nil {
							t.Fatalf("failed to read temp file: %v", err)
						}
						break
					}
				}

				if tempFileContent == nil {
					t.Fatalf("no temp file found in bucket %s", c.wantTempFileInBucket)
				}

				// Verify temp file contents if expected
				if c.wantTempFileContent != "" && string(tempFileContent) != c.wantTempFileContent {
					t.Fatalf("temp file content: got %q, want %q", string(tempFileContent), c.wantTempFileContent)
				}
			}

			// Verify final file exists in bucket if expected
			if c.wantFinalFileInBucket != "" {
				entries, err := os.ReadDir(filepath.Join(objectsRoot, c.wantFinalFileInBucket))
				if err != nil {
					t.Fatalf("failed to read bucket %s: %v", c.wantFinalFileInBucket, err)
				}

				var finalFileContent []byte
				for _, entry := range entries {
					if !strings.HasPrefix(entry.Name(), ".") && !entry.IsDir() {
						finalFileContent, err = os.ReadFile(filepath.Join(objectsRoot, c.wantFinalFileInBucket, entry.Name()))
						if err != nil {
							t.Fatalf("failed to read final file: %v", err)
						}
						break
					}
				}

				if finalFileContent == nil {
					t.Fatalf("no final file found in bucket %s", c.wantFinalFileInBucket)
				}

				// Verify final file contents if expected
				if c.wantFinalContent != "" && string(finalFileContent) != c.wantFinalContent {
					t.Fatalf("final file content: got %q, want %q", string(finalFileContent), c.wantFinalContent)
				}
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
