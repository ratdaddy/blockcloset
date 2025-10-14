package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ratdaddy/blockcloset/flatbed/internal/respond"
)

func (h *Handlers) ListBuckets(w http.ResponseWriter, r *http.Request) {
	buckets, err := h.Gantry.ListBuckets(r.Context())
	if err != nil {
		respond.Error(w, r, "InternalError", http.StatusInternalServerError)
		return
	}

	type respBucket struct {
		Name         string `json:"Name"`
		CreationDate string `json:"CreationDate"`
	}

	payload := struct {
		Buckets []respBucket `json:"Buckets"`
	}{}

	for _, b := range buckets {
		payload.Buckets = append(payload.Buckets, respBucket{
			Name:         b.Name,
			CreationDate: b.CreatedAt.UTC().Format(time.RFC3339),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		respond.Error(w, r, "InternalError", http.StatusInternalServerError)
		return
	}
}
