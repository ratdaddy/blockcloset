package grpcsvc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ratdaddy/blockcloset/gantry/internal/store"
	"github.com/ratdaddy/blockcloset/loggrpc"
	"github.com/ratdaddy/blockcloset/pkg/validation"
	bucketv1 "github.com/ratdaddy/blockcloset/proto/gen/gantry/bucket/v1"
	servicev1 "github.com/ratdaddy/blockcloset/proto/gen/gantry/service/v1"
)

func (s *Service) CreateBucket(ctx context.Context, req *servicev1.CreateBucketRequest) (*servicev1.CreateBucketResponse, error) {
	name := req.GetName()

	validator := validation.DefaultBucketNameValidator{}

	if err := validator.ValidateBucketName(name); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if name == "bad" {
		return nil, status.Errorf(codes.InvalidArgument, "invalid bucket name")
	}

	if name == "panic" {
		panic(status.New(codes.Internal, "intentional test panic"))
	}

	bucketID := store.NewID()
	now := time.Now().UTC()
	buckets := s.store.Buckets()

	if _, err := buckets.Create(ctx, bucketID, name, now); err != nil {
		if errors.Is(err, store.ErrBucketAlreadyExists) {
			conflict := &servicev1.BucketOwnershipConflict{
				Reason: servicev1.BucketOwnershipConflict_REASON_BUCKET_ALREADY_OWNED_BY_YOU,
				Bucket: name,
			}
			st := status.New(codes.AlreadyExists, err.Error())
			withDetail, err := st.WithDetails(conflict)
			if err != nil {
				return nil, loggrpc.SetError(ctx, status.Error(codes.Internal, err.Error()))
			}
			return nil, loggrpc.SetError(ctx, withDetail.Err())
		}
		return nil, loggrpc.SetError(ctx, status.Error(codes.Internal, err.Error()))
	}

	loggrpc.SetAttrs(ctx, slog.String("result", fmt.Sprintf("bucket <%s> created", name)))

	return &servicev1.CreateBucketResponse{
		Bucket: &bucketv1.Bucket{
			Name: name,
		},
	}, nil
}
