package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/ratdaddy/blockcloset/gantry/internal/config"
	"github.com/ratdaddy/blockcloset/gantry/internal/grpcsvc"
	"github.com/ratdaddy/blockcloset/gantry/internal/logger"
	"github.com/ratdaddy/blockcloset/loggrpc"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	config.Init()
	logger.Init()

	addr := ":8081"

	slog.Info("starting gantry", "addr", addr)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		slog.Error("listen failed", "err", err)
		os.Exit(1)
	}

	logger := slog.Default()
	s := grpc.NewServer(grpc.UnaryInterceptor(loggrpc.UnaryServerInterceptor(logger, &loggrpc.Options{
		Schema: loggrpc.SchemaOTEL.Concise(config.LogVerbosity == config.LogConcise),
	})))

	grpcsvc.Register(s, grpcsvc.New(logger))
	if config.EnableReflection {
		slog.Info("grpc reflection enabled")
		reflection.Register(s)
	}

	errCh := make(chan error, 1)
	go func() { errCh <- s.Serve(lis) }()

	select {
	case <-ctx.Done():
		slog.Info("shutdown signal received")
		done := make(chan struct{})
		go func() {
			s.GracefulStop()
			close(done)
		}()
		select {
		case <-done:
			slog.Info("grpc server stopped gracefully")
		case <-time.After(3 * time.Second):
			slog.Warn("graceful stop timed out; forcing")
			s.Stop()
		}
	case err := <-errCh:
		slog.Error("grpc serve exited", "err", err)
	}
}
