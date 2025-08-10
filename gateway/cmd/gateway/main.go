package main

import (
	"log"
	"net/http"
)

func newMux() *http.ServeMux {
	mux := http.NewServeMux()
	return mux
}

func main() {
	mux := newMux()
	addr := ":8080"
	log.Printf("gateway listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}