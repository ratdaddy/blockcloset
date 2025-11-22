package handlers

import (
	"net/http"
	"strconv"

	"github.com/ratdaddy/blockcloset/flatbed/internal/httpapi/respond"
)

const maxPutBytes = 5 * 1024 * 1024 * 1024 // 5 GiB

func (h *Handlers) PutObject(w http.ResponseWriter, r *http.Request) {
	bucket := r.PathValue("bucket")
	key := r.PathValue("key")

	// Reject chunked transfer encoding (check first, before Content-Length)
	// Go processes Transfer-Encoding and populates r.TransferEncoding slice
	if len(r.TransferEncoding) > 0 {
		respond.Error(w, r, "InvalidRequest", http.StatusBadRequest)
		return
	}

	// Validate Content-Length is present
	contentLengthStr := r.Header.Get("Content-Length")
	if contentLengthStr == "" {
		respond.Error(w, r, "MissingContentLength", http.StatusLengthRequired)
		return
	}

	// Validate Content-Length is greater than zero
	contentLength, err := strconv.ParseInt(contentLengthStr, 10, 64)
	if err != nil || contentLength <= 0 {
		respond.Error(w, r, "InvalidArgument", http.StatusBadRequest)
		return
	}

	// Validate Content-Length does not exceed maximum
	if contentLength > maxPutBytes {
		respond.Error(w, r, "EntityTooLarge", http.StatusBadRequest)
		return
	}

	// Validate bucket name
	if err := h.BucketValidator.ValidateBucketName(bucket); err != nil {
		respond.Error(w, r, "InvalidBucketName", http.StatusBadRequest)
		return
	}

	// Validate key
	if err := h.KeyValidator.ValidateKey(key); err != nil {
		respond.Error(w, r, "InvalidKeyName", http.StatusBadRequest)
		return
	}

	_, _, _ = h.Gantry.ResolveWrite(r.Context(), bucket, key, contentLength)

	w.WriteHeader(http.StatusOK)
}
