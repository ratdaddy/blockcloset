package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ratdaddy/blockcloset/gateway/internal/logger"
	"github.com/ratdaddy/blockcloset/gateway/internal/respond"
)

type BucketHandlers interface {
	CreateBucket(http.ResponseWriter, *http.Request)
}

func NewRouter(h BucketHandlers) http.Handler {
	mux := chi.NewRouter()
	mux.Use(logger.RequestLogger)

	mux.Use(middleware.StripSlashes)

	mux.Put("/{bucket}", h.CreateBucket)

	mux.Get("/panic", func(w http.ResponseWriter, r *http.Request) {
		panic("intentional test panic")
	})

	mux.MethodNotAllowed(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respond.Error(w, r, "page not found", http.StatusNotFound)
	}))

	mux.NotFound(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respond.Error(w, r, "page not found", http.StatusNotFound)
	}))

	return mux
}
