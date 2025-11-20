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
			name:           "404 gets custom error",
			handlerStatus:  http.StatusNotFound,
			handlerBody:    "original not found",
			wantStatus:     http.StatusNotFound,
			wantBodyPrefix: "page not found",
		},
		{
			name:           "405 gets custom error as 404",
			handlerStatus:  http.StatusMethodNotAllowed,
			handlerBody:    "original method not allowed",
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
