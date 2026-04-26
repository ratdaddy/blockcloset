package grpcsvc

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ratdaddy/blockcloset/loggrpc"
	servicev1 "github.com/ratdaddy/blockcloset/proto/gen/gantry/service/v1"
)

func (s *Service) CommitObject(ctx context.Context, req *servicev1.CommitObjectRequest) (*servicev1.CommitObjectResponse, error) {
	sizeActual := req.GetSize()
	objectID := req.GetObjectId()
	lastModifiedMs := req.GetLastModifiedMs()

	if sizeActual <= 0 {
		return nil, status.Error(codes.InvalidArgument, "InvalidSize")
	}

	if len(objectID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "InvalidObjectID")
	}

	if lastModifiedMs <= 0 {
		return nil, status.Error(codes.InvalidArgument, "InvalidLastModifiedMs")
	}

	now := time.Now().UTC()

	objects := s.store.Objects()
	if err := objects.CommitWithReplace(ctx, objectID, sizeActual, lastModifiedMs, now); err != nil {
		return nil, loggrpc.SetError(ctx, status.Error(codes.Internal, err.Error()))
	}

	return &servicev1.CommitObjectResponse{}, nil
}
