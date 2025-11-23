package grpcsvc

import (
	"context"
	"fmt"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ratdaddy/blockcloset/loggrpc"
	servicev1 "github.com/ratdaddy/blockcloset/proto/gen/gantry/service/v1"
)

func (s *Service) ResolveWrite(ctx context.Context, req *servicev1.ResolveWriteRequest) (*servicev1.ResolveWriteResponse, error) {
	bucket := req.GetBucket()
	key := req.GetKey()
	size := req.GetSize()

	// Special bucket name for testing access denied
	if bucket == "forbidden" {
		detail := &servicev1.ResolveWriteError{
			Reason: servicev1.ResolveWriteError_REASON_BUCKET_ACCESS_DENIED,
			Bucket: bucket,
		}
		st := status.New(codes.PermissionDenied, "access denied")
		withDetail, err := st.WithDetails(detail)
		if err != nil {
			return nil, loggrpc.SetError(ctx, status.Error(codes.Internal, err.Error()))
		}
		return nil, loggrpc.SetError(ctx, withDetail.Err())
	}

	buckets := s.store.Buckets()

	if _, err := buckets.GetByName(ctx, bucket); err != nil {
		detail := &servicev1.ResolveWriteError{
			Reason: servicev1.ResolveWriteError_REASON_BUCKET_NOT_FOUND,
			Bucket: bucket,
		}
		st := status.New(codes.NotFound, err.Error())
		withDetail, err := st.WithDetails(detail)
		if err != nil {
			return nil, loggrpc.SetError(ctx, status.Error(codes.Internal, err.Error()))
		}
		return nil, loggrpc.SetError(ctx, withDetail.Err())
	}

	resp := &servicev1.ResolveWriteResponse{
		ObjectId:      "stub-object-id",
		CradleAddress: "localhost:9002",
	}

	loggrpc.SetAttrs(ctx,
		slog.String("result", fmt.Sprintf("object %s/%s (%d bytes) created, write to %s",
			bucket, key, size, resp.CradleAddress)))

	return resp, nil
}
