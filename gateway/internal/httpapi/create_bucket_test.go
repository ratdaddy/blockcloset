package httpapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/ratdaddy/blockcloset/gateway/internal/httpapi"
)

func reqWithBucket(t *testing.T, method, name string) *http.Request {
	t.Helper()
	r := httptest.NewRequest(method, "/", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("bucket", name)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestCreateBucket_SetsLocationAnd201(t *testing.T) {
	t.Parallel()

	h := httpapi.NewHandlers()
	req := reqWithBucket(t, http.MethodPut, "my-bucket-123")
	rec := httptest.NewRecorder()

	h.CreateBucket(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusCreated)
	}
	if got := rec.Header().Get("Location"); got != "/my-bucket-123" {
		t.Fatalf("Location: got %q, want %q", got, "/my-bucket-123")
	}
}
