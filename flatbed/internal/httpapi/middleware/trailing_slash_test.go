package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStripTrailingSlashForBuckets(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantPath string
	}{
		{
			name:     "root path unchanged",
			path:     "/",
			wantPath: "/",
		},
		{
			name:     "bucket without slash unchanged",
			path:     "/bucket",
			wantPath: "/bucket",
		},
		{
			name:     "bucket with slash stripped",
			path:     "/bucket/",
			wantPath: "/bucket",
		},
		{
			name:     "object key without slash unchanged",
			path:     "/bucket/key",
			wantPath: "/bucket/key",
		},
		{
			name:     "object key with slash preserved",
			path:     "/bucket/key/",
			wantPath: "/bucket/key/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedPath string

			// Create a handler that captures the path it receives
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				receivedPath = r.URL.Path
				w.WriteHeader(http.StatusOK)
			})

			// Wrap with StripTrailingSlashForBuckets middleware
			handler := StripTrailingSlashForBuckets(next)

			// Make request
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			// Check path
			if receivedPath != tt.wantPath {
				t.Errorf("path = %q, want %q", receivedPath, tt.wantPath)
			}
		})
	}
}
