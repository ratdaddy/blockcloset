package grpcsvc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	servicev1 "github.com/ratdaddy/blockcloset/proto/gen/cradle/service/v1"
)

func (svc *Service) Heartbeat(_ context.Context, _ *servicev1.HeartbeatRequest) (*servicev1.HeartbeatResponse, error) {
	avail, err := svc.availableBytes(svc.objectsRoot)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "statfs: %v", err)
	}
	return &servicev1.HeartbeatResponse{AvailableBytes: int64(avail)}, nil
}
