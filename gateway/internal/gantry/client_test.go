package gantry

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"

	"github.com/ratdaddy/blockcloset/gateway/internal/httpapi"
	bucketv1 "github.com/ratdaddy/blockcloset/proto/gen/gantry/bucket/v1"
	servicev1 "github.com/ratdaddy/blockcloset/proto/gen/gantry/service/v1"
)

const bufSize = 1024 * 1024

type captureService struct {
	servicev1.UnimplementedGantryServiceServer
	lastIDs []string
}

func (s *captureService) CreateBucket(ctx context.Context, req *servicev1.CreateBucketRequest) (*servicev1.CreateBucketResponse, error) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		s.lastIDs = append([]string{}, md.Get("x-request-id")...)
	} else {
		s.lastIDs = nil
	}
	return &servicev1.CreateBucketResponse{Bucket: &bucketv1.Bucket{Name: req.GetName()}}, nil
}

func TestClientRequestIDPropagation(t *testing.T) {
	t.Parallel()

	lis := bufconn.Listen(bufSize)
	srv := grpc.NewServer()
	svc := &captureService{}
	servicev1.RegisterGantryServiceServer(srv, svc)

	go func() {
		if err := srv.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			panic(err)
		}
	}()
	t.Cleanup(func() { srv.Stop() })

	client, err := New(context.Background(), "passthrough:///bufnet", grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
		return lis.DialContext(ctx)
	}))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Fatalf("Close: %v", err)
		}
	})

	cases := []struct {
		name string
		ctx  func(*testing.T) context.Context
		want []string
	}{
		{
			name: "with request id",
			ctx:  func(*testing.T) context.Context { return httpapi.WithRequestID(context.Background(), "req-123") },
			want: []string{"req-123"},
		},
		{
			name: "without request id",
			ctx:  func(*testing.T) context.Context { return context.Background() },
			want: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc.lastIDs = nil

			if _, err := client.CreateBucket(tc.ctx(t), "case-bucket"); err != nil {
				t.Fatalf("CreateBucket: %v", err)
			}

			got := svc.lastIDs

			if len(got) != len(tc.want) {
				t.Fatalf("request id count: got %d want %d (values %v)", len(got), len(tc.want), got)
			}
			for i := range tc.want {
				if got[i] != tc.want[i] {
					t.Fatalf("request id[%d] = %q, want %q", i, got[i], tc.want[i])
				}
			}
		})
	}
}
