package grpcsvc

import (
	"context"
	"fmt"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ratdaddy/blockcloset/loggrpc"
	bucketv1 "github.com/ratdaddy/blockcloset/proto/gen/gantry/bucket/v1"
	servicev1 "github.com/ratdaddy/blockcloset/proto/gen/gantry/service/v1"
)

type Service struct {
	servicev1.UnimplementedGantryServiceServer
	log *slog.Logger
}

func New(log *slog.Logger) *Service {
	return &Service{log: log}
}

func Register(s *grpc.Server, svc *Service) {
	servicev1.RegisterGantryServiceServer(s, svc)
}

func (s *Service) CreateBucket(ctx context.Context, req *servicev1.CreateBucketRequest) (*servicev1.CreateBucketResponse, error) {
	bucket := req.GetName()

	if bucket == "bad" {
		return nil, status.Errorf(codes.InvalidArgument, "bucket name %q is not allowed", bucket)
	}

	if bucket == "panic" {
		panic(status.New(codes.Internal, "intentional test panic"))
	}

	loggrpc.SetAttrs(ctx, slog.String("result", fmt.Sprintf("bucket <%s> created", bucket)))

	return &servicev1.CreateBucketResponse{
		Bucket: &bucketv1.Bucket{
			Name: bucket,
		},
	}, nil
}
