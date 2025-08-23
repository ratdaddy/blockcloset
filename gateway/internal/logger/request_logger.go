package logger

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/httplog/v3"

	"github.com/ratdaddy/blockcloset/gateway/internal/config"
)

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
