package httpapi

import (
	"net/http"

	"github.com/ratdaddy/blockcloset/flatbed/internal/httpapi/middleware"
	"github.com/ratdaddy/blockcloset/flatbed/internal/logger"
	"github.com/ratdaddy/blockcloset/flatbed/internal/requestid"
)

// Handler defines the interface for handling S3-compatible API operations.
type Handler interface {
	CreateBucket(http.ResponseWriter, *http.Request)
	ListBuckets(http.ResponseWriter, *http.Request)
	PutObject(http.ResponseWriter, *http.Request)
}

// NewRouter creates an HTTP router
func NewRouter(h Handler) http.Handler {
	mux := http.NewServeMux()

	// Register routes
	// Use /{$} to match exactly "/" and not act as a prefix matcher
	mux.HandleFunc("GET /{$}", h.ListBuckets)
	mux.HandleFunc("PUT /{bucket}/{key...}", h.PutObject)
	mux.HandleFunc("PUT /{bucket}", h.CreateBucket)

	mux.HandleFunc("GET /panic", func(w http.ResponseWriter, r *http.Request) {
		panic("intentional test panic")
	})

	// Apply middleware in reverse order (they wrap each other)
	var handler http.Handler = mux
	handler = middleware.StripTrailingSlashForBuckets(handler)
	handler = middleware.ErrorHandler(handler)
	handler = requestid.RequestID()(handler)
	handler = logger.RequestLogger(handler)

	return handler
}
