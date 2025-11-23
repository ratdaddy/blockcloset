package middleware

import (
	"net/http"

	"github.com/ratdaddy/blockcloset/flatbed/internal/httpapi/respond"
)

// ErrorHandler captures 404 and 405 responses and overrides them with
// custom error messages.
func ErrorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := &responseBuffer{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			headers:        make(http.Header),
		}

		next.ServeHTTP(buf, r)

		// Always convert 405 to 404 for S3 compatibility
		if buf.statusCode == http.StatusMethodNotAllowed {
			respond.Error(w, r, "page not found", http.StatusNotFound)
			return
		}

		// For 404: only override if handler didn't write a body (router-level 404)
		// If handler wrote a body (e.g., "NoSuchBucket"), let it pass through
		if buf.statusCode == http.StatusNotFound && len(buf.body) == 0 {
			respond.Error(w, r, "page not found", http.StatusNotFound)
			return
		}

		// Flush the buffered response
		buf.flush()
	})
}

type responseBuffer struct {
	http.ResponseWriter
	statusCode int
	headers    http.Header
	body       []byte
	wroteOnce  bool
}

func (rb *responseBuffer) WriteHeader(code int) {
	if !rb.wroteOnce {
		rb.statusCode = code
		rb.wroteOnce = true
		// Don't write to underlying writer yet - buffer it
	}
}

func (rb *responseBuffer) Write(b []byte) (int, error) {
	if !rb.wroteOnce {
		rb.WriteHeader(http.StatusOK)
	}
	// Buffer the response body instead of writing directly
	rb.body = append(rb.body, b...)
	return len(b), nil
}

func (rb *responseBuffer) Header() http.Header {
	return rb.headers
}

func (rb *responseBuffer) flush() {
	// Copy buffered headers to actual response
	for k, v := range rb.headers {
		for _, val := range v {
			rb.ResponseWriter.Header().Add(k, val)
		}
	}
	// Write status and body to actual response
	rb.ResponseWriter.WriteHeader(rb.statusCode)
	rb.ResponseWriter.Write(rb.body)
}
