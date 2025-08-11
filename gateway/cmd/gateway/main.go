package main

import (
	"log"
	"net/http"

	"github.com/ratdaddy/blockcloset/gateway/internal/httpapi"
)

var (
	buildHandler   = func() http.Handler { return httpapi.NewRouter(httpapi.NewHandlers()) }
	listenAndServe = http.ListenAndServe
)

func main() {
	h := buildHandler()
	addr := ":8080"
	log.Printf("gateway listening on %s", addr)
	if err := listenAndServe(addr, h); err != nil {
		log.Fatal(err)
	}
}
