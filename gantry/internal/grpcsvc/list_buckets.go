package grpcsvc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ratdaddy/blockcloset/loggrpc"
	bucketv1 "github.com/ratdaddy/blockcloset/proto/gen/gantry/bucket/v1"
	servicev1 "github.com/ratdaddy/blockcloset/proto/gen/gantry/service/v1"
)

func (s *Service) ListBuckets(ctx context.Context, _ *servicev1.ListBucketsRequest) (*servicev1.ListBucketsResponse, error) {
	buckets := s.store.Buckets()

	records, err := buckets.List(ctx)
	if err != nil {
		return nil, loggrpc.SetError(ctx, status.Error(codes.Internal, err.Error()))
	}

	resp := &servicev1.ListBucketsResponse{}
	for _, rec := range records {
		resp.Buckets = append(resp.Buckets, &bucketv1.Bucket{
			Name:             rec.Name,
			CreatedAtRfc3339: rec.CreatedAt.UTC().Format("2006-01-02T15:04:05.000000Z"),
		})
	}

	return resp, nil
}
