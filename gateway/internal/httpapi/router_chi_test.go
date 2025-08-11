package httpapi_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/ratdaddy/blockcloset/gateway/internal/httpapi"
)

type stubHandlers struct {
	called     bool
	lastBucket string
	status     int
}

func (s *stubHandlers) CreateBucket(w http.ResponseWriter, r *http.Request) {
	s.called = true
	s.lastBucket = chi.URLParam(r, "bucket")
	if s.status == 0 {
		s.status = http.StatusCreated
	}
	w.WriteHeader(s.status)
}

func TestRouterChi_Routing(t *testing.T) {
	t.Parallel()

	type tc struct {
		name       string
		method     string
		target     string
		wantStatus int
		wantCalled bool
		wantBucket string
	}
	cases := []tc{
		{"PUT /{bucket} routes", http.MethodPut, "/alpha-bucket", 0, true, "alpha-bucket"},
		{"trailing slash still matches", http.MethodPut, "/bravo/", 0, true, "bravo"},
		{"GET wrong method => 404", http.MethodGet, "/alpha-bucket", http.StatusNotFound, false, ""},
		{"subpath does not match", http.MethodPut, "/alpha-bucket/obj", http.StatusNotFound, false, ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			stub := &stubHandlers{}
			r := httpapi.NewRouter(stub)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(c.method, c.target, nil)
			r.ServeHTTP(rec, req)

			if c.wantStatus != 0 && rec.Code != c.wantStatus {
				t.Fatalf("status: got %d, want %d", rec.Code, c.wantStatus)
			}
			if stub.called != c.wantCalled {
				t.Fatalf("called: got %v, want %v", stub.called, c.wantCalled)
			}
			if c.wantCalled && stub.lastBucket != c.wantBucket {
				t.Fatalf("bucket: got %q, want %q", stub.lastBucket, c.wantBucket)
			}
		})
	}
}
