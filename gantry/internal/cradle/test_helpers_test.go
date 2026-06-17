package cradle

import (
	"context"
	"net"
	"sync"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"

	servicev1 "github.com/ratdaddy/blockcloset/proto/gen/cradle/service/v1"
)

const testBufConnSize = 1024 * 1024

type captureCradleService struct {
	servicev1.UnimplementedCradleServiceServer

	mu             sync.Mutex
	heartbeatCalls int
}

func (s *captureCradleService) Heartbeat(ctx context.Context, req *servicev1.HeartbeatRequest) (*servicev1.HeartbeatResponse, error) {
	s.mu.Lock()
	s.heartbeatCalls++
	s.mu.Unlock()
	return &servicev1.HeartbeatResponse{}, nil
}

func (s *captureCradleService) HeartbeatCalls() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.heartbeatCalls
}

func newTestClient(t *testing.T) (*Client, *captureCradleService) {
	t.Helper()

	lis := bufconn.Listen(testBufConnSize)
	srv := grpc.NewServer()
	svc := &captureCradleService{}
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

	client, err := New(context.Background(), "passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Fatalf("Close: %v", err)
		}
	})

	return client, svc
}
