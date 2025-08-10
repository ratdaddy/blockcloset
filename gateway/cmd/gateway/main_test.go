package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateBucket(t *testing.T) {
	t.Parallel()

	mux := newMux()

	req := httptest.NewRequest(http.MethodPut, "/my-bucket-123", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusCreated)
	}
	if loc := rec.Header().Get("Location"); loc != "/my-bucket-123" {
		t.Fatalf("Location: got %q, want %q", loc, "/my-bucket-123")
	}
}