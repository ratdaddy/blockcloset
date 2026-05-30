package grpcsvc

import (
	"log/slog"
	"syscall"

	"google.golang.org/grpc"

	"github.com/ratdaddy/blockcloset/cradle/internal/config"
	"github.com/ratdaddy/blockcloset/cradle/internal/storage"
	servicev1 "github.com/ratdaddy/blockcloset/proto/gen/cradle/service/v1"
)

type Service struct {
	servicev1.UnimplementedCradleServiceServer
	log         *slog.Logger
	objectsRoot string
	newWriter   func(objectsRoot, bucket, objectID string) (*storage.Writer, error)
	availableBytes func(path string) (uint64, error)
}

func New(log *slog.Logger) *Service {
	return &Service{
		log:         log,
		objectsRoot: config.ObjectsRoot,
		newWriter:   storage.NewWriter,
		availableBytes: availableBytes,
	}
}

func availableBytes(path string) (uint64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, err
	}
	return stat.Bavail * uint64(stat.Bsize), nil
}

func Register(s *grpc.Server, svc *Service) {
	servicev1.RegisterCradleServiceServer(s, svc)
}
