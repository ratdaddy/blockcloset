package grpcsvc

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ratdaddy/blockcloset/gantry/internal/store"
	"github.com/ratdaddy/blockcloset/loggrpc"
	"github.com/ratdaddy/blockcloset/pkg/validation"
	servicev1 "github.com/ratdaddy/blockcloset/proto/gen/gantry/service/v1"
)

const maxPutBytes = 5 * 1024 * 1024 * 1024 // 5 GiB

func (s *Service) ResolveWrite(ctx context.Context, req *servicev1.ResolveWriteRequest) (*servicev1.ResolveWriteResponse, error) {
	bucketName := req.GetBucket()
	key := req.GetKey()
	size := req.GetSize()

	bucketValidator := validation.DefaultBucketNameValidator{}
	keyValidator := validation.DefaultKeyValidator{}

	//Validate size is greater than zero
	if size <= 0 {
		return nil, status.Error(codes.InvalidArgument, "InvalidArgument")
	}

	if size > maxPutBytes {
		return nil, status.Error(codes.InvalidArgument, "EntityTooLarge")
	}

	// Validate bucket name
	if err := bucketValidator.ValidateBucketName(bucketName); err != nil {
		return nil, status.Error(codes.InvalidArgument, "InvalidBucketName")
	}

	// Validate key
	if err := keyValidator.ValidateKey(key); err != nil {
		return nil, status.Error(codes.InvalidArgument, "InvalidKeyName")
	}

	if err := checkTestBucket(bucketName); err != nil {
		return nil, loggrpc.SetError(ctx, err)
	}

	buckets := s.store.Buckets()

	bucket, err := buckets.GetByName(ctx, bucketName)
	if err != nil {
		detail := &servicev1.ResolveWriteError{
			Reason: servicev1.ResolveWriteError_REASON_BUCKET_NOT_FOUND,
			Bucket: bucketName,
		}
		st := status.New(codes.NotFound, err.Error())
		withDetail, err := st.WithDetails(detail)
		if err != nil {
			return nil, loggrpc.SetError(ctx, status.Error(codes.Internal, err.Error()))
		}
		return nil, loggrpc.SetError(ctx, withDetail.Err())
	}

	cradle_servers := s.store.CradleServers()
	server, err := cradle_servers.SelectForUpload(ctx)
	if err != nil {
		detail := &servicev1.ResolveWriteError{
			Reason: servicev1.ResolveWriteError_REASON_NO_CRADLE_SERVERS,
			Bucket: bucketName,
		}
		st := status.New(codes.FailedPrecondition, err.Error())
		withDetail, err := st.WithDetails(detail)
		if err != nil {
			return nil, loggrpc.SetError(ctx, status.Error(codes.Internal, err.Error()))
		}
		return nil, loggrpc.SetError(ctx, withDetail.Err())
	}

	objectID := store.NewID()
	now := time.Now().UTC()
	objects := s.store.Objects()
	if _, err = objects.CreatePending(ctx, objectID, bucket.ID, key, size, server.ID, now); err != nil {
		return nil, loggrpc.SetError(ctx, status.Error(codes.Internal, err.Error()))
	}

	resp := &servicev1.ResolveWriteResponse{
		ObjectId:      objectID,
		CradleAddress: server.Address,
	}

	loggrpc.SetAttrs(ctx,
		slog.String("result", fmt.Sprintf("object %s/%s (%d bytes) created, write to %s",
			bucketName, key, size, resp.CradleAddress)))

	return resp, nil
}
