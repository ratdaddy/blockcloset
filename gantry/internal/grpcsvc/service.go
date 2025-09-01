package grpcsvc

import (
	servicev1 "github.com/ratdaddy/blockcloset/proto/gen/gantry/service/v1"
	"google.golang.org/grpc"
)

type Service struct {
	servicev1.UnimplementedGantryServiceServer
}

func Register(s *grpc.Server) {
	servicev1.RegisterGantryServiceServer(s, &Service{})
}
