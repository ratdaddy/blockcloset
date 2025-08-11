package httpapi_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ratdaddy/blockcloset/gateway/internal/httpapi"
)

func TestCreateBucket(t *testing.T) {
	t.Parallel()

	r := httpapi.NewStdlibRouter()

	req := httptest.NewRequest(http.MethodPut, "/my-bucket-123", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusCreated)
	}
	if loc := rec.Header().Get("Location"); loc != "/my-bucket-123" {
		t.Fatalf("Location: got %q, want %q", loc, "/my-bucket-123")
	}
}
