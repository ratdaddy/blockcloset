package grpcsvc

import (
	"context"
	"fmt"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ratdaddy/blockcloset/loggrpc"
	"github.com/ratdaddy/blockcloset/pkg/storage/bucket"
	bucketv1 "github.com/ratdaddy/blockcloset/proto/gen/gantry/bucket/v1"
	servicev1 "github.com/ratdaddy/blockcloset/proto/gen/gantry/service/v1"
)

func (s *Service) CreateBucket(ctx context.Context, req *servicev1.CreateBucketRequest) (*servicev1.CreateBucketResponse, error) {
	name := req.GetName()

	validator := s.validator
	if validator == nil {
		validator = bucket.DefaultBucketNameValidator{}
	}

	if err := validator.ValidateBucketName(name); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if name == "bad" {
		return nil, status.Errorf(codes.InvalidArgument, "invalid bucket name")
	}

	if name == "panic" {
		panic(status.New(codes.Internal, "intentional test panic"))
	}

	loggrpc.SetAttrs(ctx, slog.String("result", fmt.Sprintf("bucket <%s> created", name)))

	return &servicev1.CreateBucketResponse{
		Bucket: &bucketv1.Bucket{
			Name: name,
		},
	}, nil
}
