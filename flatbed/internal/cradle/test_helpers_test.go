package cradle

import (
	"context"
	"io"
	"net"
	"sync"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"

	servicev1 "github.com/ratdaddy/blockcloset/proto/gen/cradle/service/v1"
)

const testBufConnSize = 1024 * 1024

type writeObjectCall struct {
	Metadata metadata.MD
	ObjectID string
	Bucket   string
	Size     int64
	Chunks   [][]byte
}

type captureCradleService struct {
	servicev1.UnimplementedCradleServiceServer

	mu               sync.Mutex
	writeObjectCalls []writeObjectCall
	writeObjectHook  func(servicev1.CradleService_WriteObjectServer) error
}

func newCaptureCradleService() *captureCradleService {
	return &captureCradleService{}
}

func (s *captureCradleService) Reset() {
	s.mu.Lock()
	s.writeObjectCalls = nil
	s.mu.Unlock()
}

func (s *captureCradleService) WriteObjectCalls() []writeObjectCall {
	s.mu.Lock()
	defer s.mu.Unlock()
	calls := make([]writeObjectCall, len(s.writeObjectCalls))
	copy(calls, s.writeObjectCalls)
	return calls
}

func (s *captureCradleService) LastWriteObjectCall() (writeObjectCall, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.writeObjectCalls) == 0 {
		return writeObjectCall{}, false
	}
	return s.writeObjectCalls[len(s.writeObjectCalls)-1], true
}

func (s *captureCradleService) SetWriteObjectHook(fn func(servicev1.CradleService_WriteObjectServer) error) {
	s.mu.Lock()
	s.writeObjectHook = fn
	s.mu.Unlock()
}

func (s *captureCradleService) WriteObject(stream servicev1.CradleService_WriteObjectServer) error {
	call := writeObjectCall{}

	if md, ok := metadata.FromIncomingContext(stream.Context()); ok {
		call.Metadata = md.Copy()
	}

	s.mu.Lock()
	hook := s.writeObjectHook
	s.mu.Unlock()

	if hook != nil {
		return hook(stream)
	}

	// Read metadata
	req, err := stream.Recv()
	if err != nil {
		return err
	}

	meta := req.GetMetadata()
	if meta != nil {
		call.ObjectID = meta.GetObjectId()
		call.Bucket = meta.GetBucket()
		call.Size = meta.GetSize()
	}

	// Read all chunks
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if chunk := req.GetChunk(); chunk != nil {
			call.Chunks = append(call.Chunks, chunk)
		}
	}

	s.mu.Lock()
	s.writeObjectCalls = append(s.writeObjectCalls, call)
	s.mu.Unlock()

	// Send response
	var totalBytes int64
	for _, chunk := range call.Chunks {
		totalBytes += int64(len(chunk))
	}

	return stream.SendAndClose(&servicev1.WriteObjectResponse{
		BytesWritten:  totalBytes,
		CommittedAtMs: 1234567890,
	})
}

func newTestClient(t *testing.T) (*Client, *captureCradleService) {
	t.Helper()

	lis := bufconn.Listen(testBufConnSize)
	srv := grpc.NewServer()
	svc := newCaptureCradleService()
	servicev1.RegisterCradleServiceServer(srv, svc)

	go func() {
		if err := srv.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			panic(err)
		}
	}()

	t.Cleanup(func() {
		srv.Stop()
		_ = lis.Close()
	})

	// Create a pool that uses bufconn for testing
	pool := NewPoolWithDialer(func(ctx context.Context, address string) (*grpc.ClientConn, error) {
		return grpc.NewClient(address,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
				return lis.DialContext(ctx)
			}))
	})

	t.Cleanup(func() {
		if err := pool.Close(); err != nil {
			t.Fatalf("Close: %v", err)
		}
	})

	client := New(pool)
	return client, svc
}
