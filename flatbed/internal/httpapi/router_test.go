package httpapi_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ratdaddy/blockcloset/flatbed/internal/httpapi"
)

// Table entries validate routing-only behavior. To cover a new endpoint,
// add the stubbed handler implementation plus a case with the HTTP method
// and target. Point callCount at the stub's counter so the test proves the
// router triggers the correct handler without involving Gantry stubs.

type stubBucketHandlers struct {
	createStatus    int
	listStatus      int
	putObjectStatus int
	createCalls     int
	listCalls       int
	putObjectCalls  int
	lastPutKey      string
}

func newStubBucketHandlers() *stubBucketHandlers {
	return &stubBucketHandlers{
		createStatus:    http.StatusCreated,
		listStatus:      http.StatusOK,
		putObjectStatus: http.StatusOK,
	}
}

func (s *stubBucketHandlers) CreateBucket(w http.ResponseWriter, r *http.Request) {
	s.createCalls++
	w.WriteHeader(s.createStatus)
}

func (s *stubBucketHandlers) ListBuckets(w http.ResponseWriter, r *http.Request) {
	s.listCalls++
	w.WriteHeader(s.listStatus)
}

func (s *stubBucketHandlers) PutObject(w http.ResponseWriter, r *http.Request) {
	s.putObjectCalls++
	s.lastPutKey = r.PathValue("key")
	w.WriteHeader(s.putObjectStatus)
}

func (s *stubBucketHandlers) CreateCount() int {
	return s.createCalls
}

func (s *stubBucketHandlers) ListCount() int {
	return s.listCalls
}

func (s *stubBucketHandlers) PutObjectCount() int {
	return s.putObjectCalls
}

func TestRouterChi_Routing(t *testing.T) {
	t.Parallel()

	type tc struct {
		name       string
		method     string
		target     string
		wantStatus int
		callName   string
		callCount  func(*stubBucketHandlers) int
		wantKey    string
	}
	cases := []tc{
		{
			name:       "PUT /{bucket} routes",
			method:     http.MethodPut,
			target:     "/alpha-bucket",
			wantStatus: http.StatusCreated,
			callName:   "create handler",
			callCount:  (*stubBucketHandlers).CreateCount,
		},
		{
			name:       "trailing slash still matches",
			method:     http.MethodPut,
			target:     "/bravo/",
			wantStatus: http.StatusCreated,
			callName:   "create handler",
			callCount:  (*stubBucketHandlers).CreateCount,
		},
		{
			name:       "PUT /{bucket}/{key} routes to PutObject",
			method:     http.MethodPut,
			target:     "/bucket/key",
			wantStatus: http.StatusOK,
			callName:   "put object handler",
			callCount:  (*stubBucketHandlers).PutObjectCount,
			wantKey:    "key",
		},
		{
			name:       "PUT /{bucket}/{key} with nested path",
			method:     http.MethodPut,
			target:     "/bucket/path/to/key",
			wantStatus: http.StatusOK,
			callName:   "put object handler",
			callCount:  (*stubBucketHandlers).PutObjectCount,
			wantKey:    "path/to/key",
		},
		{
			name:       "GET list buckets",
			method:     http.MethodGet,
			target:     "/",
			wantStatus: http.StatusOK,
			callName:   "list handler",
			callCount:  (*stubBucketHandlers).ListCount,
		},
		{
			name:       "GET wrong path => 404",
			method:     http.MethodGet,
			target:     "/alpha-bucket",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := newStubBucketHandlers()
			r := httpapi.NewRouter(h)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(c.method, c.target, nil)
			r.ServeHTTP(rec, req)

			if rec.Code != c.wantStatus {
				t.Fatalf("status: got %d, want %d", rec.Code, c.wantStatus)
			}

			if c.callCount != nil {
				if got := c.callCount(h); got != 1 {
					t.Fatalf("%s: %s count got %d, want 1", c.name, c.callName, got)
				}
			}

			if c.wantKey != "" {
				if h.lastPutKey != c.wantKey {
					t.Fatalf("%s: key got %q, want %q", c.name, h.lastPutKey, c.wantKey)
				}
			}
		})
	}
}
