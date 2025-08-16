package logger

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/httplog/v3"

	"github.com/ratdaddy/blockcloset/gateway/internal/config"
)

type ctxWriter struct {
	http.ResponseWriter
	ctx context.Context
}

func (w *ctxWriter) Context() context.Context { return w.ctx }

var RequestLogger = func(next http.Handler) http.Handler {
	var sc *httplog.Schema

	if config.LogVerbosity == config.LogConcise {
		sc = httplog.SchemaOTEL.Concise(true)
	} else {
		sc = httplog.SchemaOTEL
	}

	logger := slog.Default()

	return httplog.RequestLogger(logger, &httplog.Options{
		Schema:        sc,
		RecoverPanics: true,
	})(next)
}
