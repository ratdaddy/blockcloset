package httpapi

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

type Handlers struct {
	Validator BucketNameValidator
}

func NewHandlers() *Handlers { return &Handlers{Validator: DefaultBucketNameValidator{}} }

func (h *Handlers) CreateBucket(w http.ResponseWriter, r *http.Request) {
	bucket := chi.URLParam(r, "bucket")

	if err := h.Validator.ValidateBucketName(bucket); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", "/"+bucket)
	w.WriteHeader(http.StatusCreated)
}
