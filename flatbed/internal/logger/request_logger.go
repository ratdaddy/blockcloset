package logger

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/httplog/v3"

	"github.com/ratdaddy/blockcloset/flatbed/internal/config"
)

var RequestLogger = func(next http.Handler) http.Handler {
	logger := slog.Default()

	return httplog.RequestLogger(logger, &httplog.Options{
		Schema:        httplog.SchemaOTEL.Concise(config.LogVerbosity == config.LogConcise),
		RecoverPanics: true,
	})(next)
}
