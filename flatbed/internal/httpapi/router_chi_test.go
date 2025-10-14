package httpapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ratdaddy/blockcloset/flatbed/internal/gantry"
	"github.com/ratdaddy/blockcloset/flatbed/internal/httpapi"
	"github.com/ratdaddy/blockcloset/flatbed/internal/testutil"
)

func TestRouterChi_Routing(t *testing.T) {
	t.Parallel()

	type tc struct {
		name       string
		method     string
		target     string
		wantStatus int
		wantCalls  int
		wantBucket string
	}
	cases := []tc{
		{"PUT /{bucket} routes", http.MethodPut, "/alpha-bucket", http.StatusCreated, 1, "alpha-bucket"},
		{"trailing slash still matches", http.MethodPut, "/bravo/", http.StatusCreated, 1, "bravo"},
		{"GET list buckets", http.MethodGet, "/", http.StatusOK, 1, ""},
		{"GET wrong path => 404", http.MethodGet, "/alpha-bucket", http.StatusNotFound, 0, ""},
		{"subpath does not match", http.MethodPut, "/alpha-bucket/obj", http.StatusNotFound, 0, ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			g := testutil.NewGantryStub()
			if c.wantBucket != "" {
				g.CreateFn = func(_ context.Context, name string) (string, error) {
					return name, nil
				}
			}
			if c.method == http.MethodGet && c.target == "/" {
				g.ListFn = func(context.Context) ([]gantry.Bucket, error) {
					return nil, nil
				}
			}

			h := httpapi.NewHandlers(g)
			r := httpapi.NewRouter(h)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(c.method, c.target, nil)
			r.ServeHTTP(rec, req)

			if rec.Code != c.wantStatus {
				t.Fatalf("status: got %d, want %d", rec.Code, c.wantStatus)
			}

			switch c.method {
			case http.MethodPut:
				if got := len(g.CreateCalls); got != c.wantCalls {
					t.Fatalf("create calls: got %d want %d (calls=%v)", got, c.wantCalls, g.CreateCalls)
				}
				if c.wantBucket != "" && g.CreateCalls[0] != c.wantBucket {
					t.Fatalf("bucket: got %q, want %q", g.CreateCalls[0], c.wantBucket)
				}
			case http.MethodGet:
				if got := g.ListCalls; got != c.wantCalls {
					t.Fatalf("list calls: got %d want %d", got, c.wantCalls)
				}
			}
		})
	}
}
