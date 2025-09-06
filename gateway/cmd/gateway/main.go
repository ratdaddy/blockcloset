package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/ratdaddy/blockcloset/gateway/internal/config"
	"github.com/ratdaddy/blockcloset/gateway/internal/httpapi"
	"github.com/ratdaddy/blockcloset/gateway/internal/logger"
)

var (
	buildHandler   = func() http.Handler { return httpapi.NewRouter(httpapi.NewHandlers()) }
	listenAndServe = http.ListenAndServe
)

func main() {
	config.Init()
	logger.Init()

	h := buildHandler()
	addr := ":8080"
	slog.Info("starting gateway", "addr", addr)
	if err := listenAndServe(addr, h); err != nil {
		slog.Error("http listen and serve exited", "err", err)
		os.Exit(1)
	}
}
