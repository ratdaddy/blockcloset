package logger

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/httplog/v3"
)

func LogResult(r *http.Request, result string) {
	ctx := r.Context()
	httplog.SetAttrs(ctx, slog.String("result", result))
}

func LogError(w http.ResponseWriter, r *http.Request, err string) {
	httplog.SetAttrs(r.Context(), slog.String("error.message", err))
}

func LogGantryError(r *http.Request, err error) {
	httplog.SetAttrs(r.Context(),
		slog.String("error.type", "gantry"),
		slog.String("error.detail", err.Error()),
	)
}

// LogWritePlan logs the write plan information returned from Gantry.
func LogWritePlan(r *http.Request, objectID, cradleAddress string, size int64) {
	httplog.SetAttrs(r.Context(),
		slog.String("write_plan.object_id", objectID),
		slog.String("write_plan.cradle_address", cradleAddress),
		slog.Int64("write_plan.size", size),
	)
}
