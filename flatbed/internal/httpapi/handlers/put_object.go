package handlers

import (
	"net/http"
)

func (h *Handlers) PutObject(w http.ResponseWriter, r *http.Request) {
	bucket := r.PathValue("bucket")
	key := r.PathValue("key")

	_ = h.Gantry.ResolveWrite(r.Context(), bucket, key)

	w.WriteHeader(http.StatusOK)
}
