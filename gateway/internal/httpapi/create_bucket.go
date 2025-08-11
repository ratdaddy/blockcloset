package httpapi

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

type Handlers struct{}

func NewHandlers() *Handlers { return &Handlers{} }

func (h *Handlers) CreateBucket(w http.ResponseWriter, r *http.Request) {
	bucket := chi.URLParam(r, "bucket")
	w.Header().Set("Location", "/"+bucket)
	w.WriteHeader(http.StatusCreated)
}
