package gantry

import (
	"context"
	"net"
	"sync"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/proto"

	bucketv1 "github.com/ratdaddy/blockcloset/proto/gen/gantry/bucket/v1"
	servicev1 "github.com/ratdaddy/blockcloset/proto/gen/gantry/service/v1"
	writeplanv1 "github.com/ratdaddy/blockcloset/proto/gen/gantry/write_plan/v1"
)

const testBufConnSize = 1024 * 1024

type createBucketCall struct {
	Metadata metadata.MD
	Request  *servicev1.CreateBucketRequest
}

type listBucketsCall struct {
	Metadata metadata.MD
	Request  *servicev1.ListBucketsRequest
}

type planWriteCall struct {
	Metadata metadata.MD
	Request  *servicev1.PlanWriteRequest
}

type captureGantryService struct {
	servicev1.UnimplementedGantryServiceServer

	mu                 sync.Mutex
	createBucketCalls  []createBucketCall
	createBucketHookFn func(context.Context, *servicev1.CreateBucketRequest) (*servicev1.CreateBucketResponse, error)
	listBucketCalls    []listBucketsCall
	listBucketHookFn   func(context.Context, *servicev1.ListBucketsRequest) (*servicev1.ListBucketsResponse, error)
	planWriteCalls     []planWriteCall
	planWriteHookFn    func(context.Context, *servicev1.PlanWriteRequest) (*servicev1.PlanWriteResponse, error)
}

func newCaptureGantryService() *captureGantryService {
	return &captureGantryService{}
}

func (s *captureGantryService) Reset() {
	s.mu.Lock()
	s.createBucketCalls = nil
	s.listBucketCalls = nil
	s.planWriteCalls = nil
	s.mu.Unlock()
}

func (s *captureGantryService) CreateBucketCalls() []createBucketCall {
	s.mu.Lock()
	defer s.mu.Unlock()
	calls := make([]createBucketCall, len(s.createBucketCalls))
	copy(calls, s.createBucketCalls)
	return calls
}

func (s *captureGantryService) LastCreateBucketCall() (createBucketCall, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.createBucketCalls) == 0 {
		return createBucketCall{}, false
	}
	return s.createBucketCalls[len(s.createBucketCalls)-1], true
}

func (s *captureGantryService) SetCreateBucketHook(fn func(context.Context, *servicev1.CreateBucketRequest) (*servicev1.CreateBucketResponse, error)) {
	s.mu.Lock()
	s.createBucketHookFn = fn
	s.mu.Unlock()
}

func (s *captureGantryService) CreateBucket(ctx context.Context, req *servicev1.CreateBucketRequest) (*servicev1.CreateBucketResponse, error) {
	call := createBucketCall{
		Request: proto.Clone(req).(*servicev1.CreateBucketRequest),
	}

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		call.Metadata = md.Copy()
	}

	s.mu.Lock()
	s.createBucketCalls = append(s.createBucketCalls, call)
	hook := s.createBucketHookFn
	s.mu.Unlock()

	if hook != nil {
		return hook(ctx, req)
	}

	return &servicev1.CreateBucketResponse{Bucket: &bucketv1.Bucket{Name: req.GetName()}}, nil
}

func (s *captureGantryService) ListBuckets(ctx context.Context, req *servicev1.ListBucketsRequest) (*servicev1.ListBucketsResponse, error) {
	call := listBucketsCall{
		Request: proto.Clone(req).(*servicev1.ListBucketsRequest),
	}

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		call.Metadata = md.Copy()
	}

	s.mu.Lock()
	s.listBucketCalls = append(s.listBucketCalls, call)
	hook := s.listBucketHookFn
	s.mu.Unlock()

	if hook != nil {
		return hook(ctx, req)
	}

	return &servicev1.ListBucketsResponse{}, nil
}

func (s *captureGantryService) ListBucketCalls() []listBucketsCall {
	s.mu.Lock()
	defer s.mu.Unlock()

	calls := make([]listBucketsCall, len(s.listBucketCalls))
	copy(calls, s.listBucketCalls)
	return calls
}

func (s *captureGantryService) LastListBucketsCall() (listBucketsCall, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.listBucketCalls) == 0 {
		return listBucketsCall{}, false
	}
	return s.listBucketCalls[len(s.listBucketCalls)-1], true
}

func (s *captureGantryService) SetListBucketsHook(fn func(context.Context, *servicev1.ListBucketsRequest) (*servicev1.ListBucketsResponse, error)) {
	s.mu.Lock()
	s.listBucketHookFn = fn
	s.mu.Unlock()
}

func (s *captureGantryService) PlanWrite(ctx context.Context, req *servicev1.PlanWriteRequest) (*servicev1.PlanWriteResponse, error) {
	call := planWriteCall{
		Request: proto.Clone(req).(*servicev1.PlanWriteRequest),
	}

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		call.Metadata = md.Copy()
	}

	s.mu.Lock()
	s.planWriteCalls = append(s.planWriteCalls, call)
	hook := s.planWriteHookFn
	s.mu.Unlock()

	if hook != nil {
		return hook(ctx, req)
	}

	return &servicev1.PlanWriteResponse{
		WritePlan: &writeplanv1.WritePlan{
			ObjectId:      "test-object-id",
			CradleAddress: "localhost:9002",
		},
	}, nil
}

func (s *captureGantryService) PlanWriteCalls() []planWriteCall {
	s.mu.Lock()
	defer s.mu.Unlock()

	calls := make([]planWriteCall, len(s.planWriteCalls))
	copy(calls, s.planWriteCalls)
	return calls
}

func (s *captureGantryService) LastPlanWriteCall() (planWriteCall, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.planWriteCalls) == 0 {
		return planWriteCall{}, false
	}
	return s.planWriteCalls[len(s.planWriteCalls)-1], true
}

func (s *captureGantryService) SetPlanWriteHook(fn func(context.Context, *servicev1.PlanWriteRequest) (*servicev1.PlanWriteResponse, error)) {
	s.mu.Lock()
	s.planWriteHookFn = fn
	s.mu.Unlock()
}

func newTestClient(t *testing.T) (*Client, *captureGantryService) {
	t.Helper()

	lis := bufconn.Listen(testBufConnSize)
	srv := grpc.NewServer()
	svc := newCaptureGantryService()
	servicev1.RegisterGantryServiceServer(srv, svc)

	go func() {
		if err := srv.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			panic(err)
		}
	}()

	t.Cleanup(func() {
		srv.Stop()
		_ = lis.Close()
	})

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

	return client, svc
}
