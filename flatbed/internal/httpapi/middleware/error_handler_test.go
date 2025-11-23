package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestErrorHandler(t *testing.T) {
	tests := []struct {
		name           string
		handlerStatus  int
		handlerBody    string
		wantStatus     int
		wantBodyPrefix string
	}{
		{
			name:           "404 without body gets custom error",
			handlerStatus:  http.StatusNotFound,
			handlerBody:    "",
			wantStatus:     http.StatusNotFound,
			wantBodyPrefix: "page not found",
		},
		{
			name:           "405 without body gets custom error as 404",
			handlerStatus:  http.StatusMethodNotAllowed,
			handlerBody:    "",
			wantStatus:     http.StatusNotFound,
			wantBodyPrefix: "page not found",
		},
		{
			name:           "404 with body passes through unchanged",
			handlerStatus:  http.StatusNotFound,
			handlerBody:    "NoSuchBucket",
			wantStatus:     http.StatusNotFound,
			wantBodyPrefix: "NoSuchBucket",
		},
		{
			name:           "405 with body still gets converted to 404",
			handlerStatus:  http.StatusMethodNotAllowed,
			handlerBody:    "MethodNotAllowed",
			wantStatus:     http.StatusNotFound,
			wantBodyPrefix: "page not found",
		},
		{
			name:           "400 passes through unchanged",
			handlerStatus:  http.StatusBadRequest,
			handlerBody:    "bad request body",
			wantStatus:     http.StatusBadRequest,
			wantBodyPrefix: "bad request body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a handler that returns the specified status and body
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.handlerStatus)
				w.Write([]byte(tt.handlerBody))
			})

			// Wrap with ErrorHandler middleware
			handler := ErrorHandler(next)

			// Make request
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			// Check status
			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			// Check body
			body := rec.Body.String()
			if !strings.Contains(body, tt.wantBodyPrefix) {
				t.Errorf("body = %q, want to contain %q", body, tt.wantBodyPrefix)
			}
		})
	}
}
