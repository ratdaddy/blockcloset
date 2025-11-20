package handlers

import (
	"fmt"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ratdaddy/blockcloset/flatbed/internal/httpapi/respond"
	"github.com/ratdaddy/blockcloset/flatbed/internal/logger"
	servicev1 "github.com/ratdaddy/blockcloset/proto/gen/gantry/service/v1"
)

func (h *Handlers) CreateBucket(w http.ResponseWriter, r *http.Request) {
	bucket := r.PathValue("bucket")

	if err := h.Validator.ValidateBucketName(bucket); err != nil {
		respond.Error(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	if _, err := h.Gantry.CreateBucket(r.Context(), bucket); err != nil {
		st, ok := status.FromError(err)
		if !ok {
			respond.Error(w, r, "InternalError", http.StatusInternalServerError)
			return
		}

		switch st.Code() {
		case codes.InvalidArgument:
			respond.Error(w, r, st.Message(), http.StatusBadRequest)
		case codes.Internal:
			respond.Error(w, r, "InternalError", http.StatusInternalServerError)
		case codes.AlreadyExists:
			respond.Error(w, r, bucketConflictMessage(st), http.StatusConflict)
		default:
			respond.Error(w, r, "InternalError", http.StatusInternalServerError)
		}
		return
	}

	logger.LogResult(r, fmt.Sprintf("bucket <%s> created", bucket))
	w.Header().Set("Location", "/"+bucket)
	w.WriteHeader(http.StatusCreated)
}

func bucketConflictMessage(st *status.Status) string {
	const defaultMsg = "BucketAlreadyExists"

	for _, detail := range st.Details() {
		conflict, ok := detail.(*servicev1.BucketOwnershipConflict)
		if !ok {
			continue
		}

		switch conflict.GetReason() {
		case servicev1.BucketOwnershipConflict_REASON_BUCKET_ALREADY_OWNED_BY_YOU:
			return "BucketAlreadyOwnedByYou"
		case servicev1.BucketOwnershipConflict_REASON_BUCKET_ALREADY_EXISTS:
			return "BucketAlreadyExists"
		}
	}

	return defaultMsg
}
