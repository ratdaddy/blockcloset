package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/ratdaddy/blockcloset/cradle/internal/config"
	"github.com/ratdaddy/blockcloset/cradle/internal/grpcsvc"
	"github.com/ratdaddy/blockcloset/cradle/internal/logger"
	"github.com/ratdaddy/blockcloset/loggrpc"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	config.Init()
	logger.Init()

	addr := fmt.Sprintf(":%d", config.CradlePort)

	slog.Info("starting cradle", "addr", addr)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		slog.Error("listen failed", "err", err)
		os.Exit(1)
	}

	slogger := slog.Default()
	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			loggrpc.UnaryServerInterceptor(slogger, &loggrpc.Options{
				Schema: loggrpc.SchemaOTEL.Concise(config.LogVerbosity == config.LogConcise),
			}),
			recovery.UnaryServerInterceptor(
				recovery.WithRecoveryHandlerContext(loggrpc.RecoverToStatus),
			),
		),
		grpc.ChainStreamInterceptor(
			loggrpc.StreamServerInterceptor(slogger, &loggrpc.Options{
				Schema: loggrpc.SchemaOTEL.Concise(config.LogVerbosity == config.LogConcise),
			}),
			recovery.StreamServerInterceptor(
				recovery.WithRecoveryHandlerContext(loggrpc.RecoverToStatus),
			),
		),
	)

	grpcsvc.Register(s, grpcsvc.New(slogger))
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
