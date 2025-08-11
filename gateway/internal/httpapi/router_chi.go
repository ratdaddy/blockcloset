package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type BucketHandlers interface {
	CreateBucket(http.ResponseWriter, *http.Request)
}

func NewRouter(h BucketHandlers) http.Handler {
	mux := chi.NewRouter()

	mux.Use(middleware.StripSlashes)

	mux.MethodNotAllowed(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		http.NotFound(w, req)
	}))

	mux.Put("/{bucket}", h.CreateBucket)

	return mux
}
