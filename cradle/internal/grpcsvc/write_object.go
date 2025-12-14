package grpcsvc

import (
	"errors"
	"io"
	"log/slog"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ratdaddy/blockcloset/loggrpc"
	servicev1 "github.com/ratdaddy/blockcloset/proto/gen/cradle/service/v1"
)

func (s *Service) WriteObject(stream servicev1.CradleService_WriteObjectServer) error {
	ctx := stream.Context()

	req, err := stream.Recv()
	if err != nil {
		return loggrpc.SetError(ctx, status.Error(codes.Internal, err.Error()))
	}

	meta := req.GetMetadata()
	if meta == nil {
		return loggrpc.SetError(ctx, status.Error(codes.InvalidArgument, "metadata must be the first message"))
	}

	bucket := meta.GetBucket()
	objectID := meta.GetObjectId()
	size := meta.GetSize()

	s.log.InfoContext(ctx, "write metadata received",
		"bucket", bucket,
		"object_id", objectID,
		"size", size,
	)

	loggrpc.SetAttrs(ctx,
		slog.String("bucket", bucket),
		slog.String("object_id", objectID),
		slog.Int64("size", size),
	)

	var total int64

	for {
		req, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			s.log.InfoContext(ctx, "write stream complete", "bytes_received", total)

			// Check for special test buckets that trigger specific behaviors
			bytesToReport := checkTestBucket(bucket, total)

			return stream.SendAndClose(&servicev1.WriteObjectResponse{
				BytesWritten:  bytesToReport,
				CommittedAtMs: time.Now().UnixMilli(),
			})
		}
		if err != nil {
			return loggrpc.SetError(ctx, status.Error(codes.Internal, err.Error()))
		}

		chunk := req.GetChunk()
		if chunk == nil {
			return loggrpc.SetError(ctx, status.Error(codes.InvalidArgument, "chunk payload missing"))
		}

		chunkBytes := int64(len(chunk))
		total += chunkBytes
		s.log.InfoContext(ctx, "write chunk received",
			"chunk_contents", string(chunk),
			"chunk_bytes", chunkBytes,
			"accumulated_bytes", total,
		)
	}
}
