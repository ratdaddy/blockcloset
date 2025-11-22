package grpcsvc

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ratdaddy/blockcloset/loggrpc"
	servicev1 "github.com/ratdaddy/blockcloset/proto/gen/gantry/service/v1"
)

func (s *Service) ResolveWrite(ctx context.Context, req *servicev1.ResolveWriteRequest) (*servicev1.ResolveWriteResponse, error) {
	bucket := req.GetBucket()
	key := req.GetKey()
	size := req.GetSize()

	resp := &servicev1.ResolveWriteResponse{
		ObjectId:      "stub-object-id",
		CradleAddress: "localhost:9002",
	}

	loggrpc.SetAttrs(ctx,
		slog.String("result", fmt.Sprintf("object %s/%s (%d bytes) created, write to %s",
			bucket, key, size, resp.CradleAddress)))

	return resp, nil
}
