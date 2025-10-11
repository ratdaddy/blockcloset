package grpcsvc

import (
	"database/sql"
	"log/slog"

	"google.golang.org/grpc"

	"github.com/ratdaddy/blockcloset/gantry/internal/store"
	servicev1 "github.com/ratdaddy/blockcloset/proto/gen/gantry/service/v1"
)

type Service struct {
	servicev1.UnimplementedGantryServiceServer
	log   *slog.Logger
	db    *sql.DB
	store store.Store
}

func New(log *slog.Logger, db *sql.DB) *Service {
	svc := &Service{
		log: log,
		db:  db,
	}

	if db != nil {
		svc.store = store.New(db)
	}

	return svc
}

func Register(s *grpc.Server, svc *Service) {
	servicev1.RegisterGantryServiceServer(s, svc)
}
