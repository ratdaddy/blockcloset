package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/ratdaddy/blockcloset/flatbed/internal/config"
	"github.com/ratdaddy/blockcloset/flatbed/internal/gantry"
	"github.com/ratdaddy/blockcloset/flatbed/internal/httpapi"
	"github.com/ratdaddy/blockcloset/flatbed/internal/httpapi/handlers"
	"github.com/ratdaddy/blockcloset/flatbed/internal/logger"
)

var (
	buildHandler = func(g handlers.GantryClient) http.Handler {
		return httpapi.NewRouter(handlers.NewHandlers(g))
	}
	gantryClient = func(addr string) (handlers.GantryClient, error) {
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

	addr := fmt.Sprintf(":%d", config.FlatbedPort)
	slog.Info("starting flatbed", "addr", addr)
	if err := listenAndServe(addr, h); err != nil {
		slog.Error("http listen and serve exited", "err", err)
		os.Exit(1)
	}
}
