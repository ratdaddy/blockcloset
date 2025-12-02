package grpcsvc

import (
	"log/slog"

	"google.golang.org/grpc"

	servicev1 "github.com/ratdaddy/blockcloset/proto/gen/cradle/service/v1"
)

type Service struct {
	servicev1.UnimplementedCradleServiceServer
	log *slog.Logger
}

func New(log *slog.Logger) *Service {
	return &Service{
		log: log,
	}
}

func Register(s *grpc.Server, svc *Service) {
	servicev1.RegisterCradleServiceServer(s, svc)
}
