package httpapi

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/ratdaddy/blockcloset/gateway/internal/logger"
	"github.com/ratdaddy/blockcloset/gateway/internal/respond"
)

type Handlers struct {
	Validator BucketNameValidator
}

func NewHandlers() *Handlers { return &Handlers{Validator: DefaultBucketNameValidator{}} }

func (h *Handlers) CreateBucket(w http.ResponseWriter, r *http.Request) {
	bucket := chi.URLParam(r, "bucket")

	if err := h.Validator.ValidateBucketName(bucket); err != nil {
		respond.Error(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	logger.LogResult(r, fmt.Sprintf("bucket <%s> created", bucket))
	w.Header().Set("Location", "/"+bucket)
	w.WriteHeader(http.StatusCreated)
}
