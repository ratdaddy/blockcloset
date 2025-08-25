package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var version = "0.0.1"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	logger.Info("gantryd starting", "version", version)
	<-ctx.Done()

	logger.Info("gantryd shutting down")
	time.Sleep(200 * time.Millisecond)
}
