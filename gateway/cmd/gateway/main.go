package main

import (
	"log"
	"net/http"

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
	log.Printf("gateway listening on %s", addr)
	if err := listenAndServe(addr, h); err != nil {
		log.Fatal(err)
	}
}
