package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/ratdaddy/blockcloset/gateway/internal/config"
	"github.com/ratdaddy/blockcloset/gateway/internal/gantry"
	"github.com/ratdaddy/blockcloset/gateway/internal/httpapi"
	"github.com/ratdaddy/blockcloset/gateway/internal/logger"
)

var (
	buildHandler = func(g httpapi.GantryClient) http.Handler {
		return httpapi.NewRouter(httpapi.NewHandlers(g))
	}
	gantryClient = func(addr string) (httpapi.GantryClient, error) {
		return gantry.New(context.Background(), addr)
	}
	listenAndServe = http.ListenAndServe
)

func main() {
	config.Init()
	logger.Init()

	g, err := gantryClient(config.GantryAddr)
	if err != nil {
		slog.Error("gantry dial", "err", err)
		os.Exit(1)
	}
	if closer, ok := g.(interface{ Close() error }); ok {
		defer func() {
			if err := closer.Close(); err != nil {
				slog.Error("gantry client close", "err", err)
			}
		}()
	}

	h := buildHandler(g)

	addr := fmt.Sprintf(":%d", config.GatewayPort)
	slog.Info("starting gateway", "addr", addr)
	if err := listenAndServe(addr, h); err != nil {
		slog.Error("http listen and serve exited", "err", err)
		os.Exit(1)
	}
}
