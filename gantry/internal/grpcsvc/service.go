package grpcsvc

import (
	"log/slog"

	"google.golang.org/grpc"

	"github.com/ratdaddy/blockcloset/pkg/storage/bucket"
	servicev1 "github.com/ratdaddy/blockcloset/proto/gen/gantry/service/v1"
)

type Service struct {
	servicev1.UnimplementedGantryServiceServer
	log       *slog.Logger
	validator bucket.BucketNameValidator
}

func New(log *slog.Logger) *Service {
	return &Service{
		log:       log,
		validator: bucket.DefaultBucketNameValidator{},
	}
}

func Register(s *grpc.Server, svc *Service) {
	servicev1.RegisterGantryServiceServer(s, svc)
}
