package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/ratdaddy/blockcloset/gantry/internal/grpcsvc"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	addr := ":8081"

	logger.Info("gantry starting", "addr", addr)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Error("listen failed", "err", err)
		os.Exit(1)
	}

	s := grpc.NewServer()

	grpcsvc.Register(s)

	enable_reflection, _ := strconv.ParseBool(os.Getenv("ENABLE_REFLECTION"))
	if enable_reflection {
		reflection.Register(s)
	}

	errCh := make(chan error, 1)
	go func() { errCh <- s.Serve(lis) }()

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
		// graceful stop with a short cap, then force
		done := make(chan struct{})
		go func() {
			s.GracefulStop()
			close(done)
		}()
		select {
		case <-done:
			logger.Info("grpc server stopped gracefully")
		case <-time.After(3 * time.Second):
			logger.Warn("graceful stop timed out; forcing")
			s.Stop()
		}
	case err := <-errCh:
		// Serve exited on its own (listener closed or fatal error)
		logger.Error("grpc serve exited", "err", err)
	}
}
