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
