package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ratdaddy/blockcloset/flatbed/internal/httpapi/handlers"
	"github.com/ratdaddy/blockcloset/flatbed/internal/testutil"
)

// TestMain_WiresAddressAndHandler verifies that main wires each public endpoint
// through to the expected Gantry call. When adding a new endpoint, add a table
// entry that issues the HTTP request, sets up any stub behavior, and points
// callCount at the Gantry method the handler should invoke so we keep coverage
// across the full main wiring.
func TestMain_WiresAddressAndHandler(t *testing.T) {
	t.Parallel()

	origBuild, origListen, origGantry := buildHandler, listenAndServe, gantryClient
	defer func() { buildHandler, listenAndServe, gantryClient = origBuild, origListen, origGantry }()

	fg := testutil.NewGantryStub()
	gantryClient = func(addr string) (handlers.GantryClient, error) {
		return fg, nil
	}

	var gotAddr string
	var served bool

	type testCase struct {
		name       string
		method     string
		target     string
		headers    map[string]string
		wantStatus int
		callName   string
		callCount  func(*testutil.GantryStub) int
	}

	tests := []testCase{
		{
			name:       "E2E - PutBucket",
			method:     http.MethodPut,
			target:     "/e2e-bucket",
			wantStatus: http.StatusCreated,
			callName:   "gantry create",
			callCount:  (*testutil.GantryStub).CreateCount,
		},
		{
			name:       "E2E - ListBuckets",
			method:     http.MethodGet,
			target:     "/",
			wantStatus: http.StatusOK,
			callName:   "gantry list",
			callCount:  (*testutil.GantryStub).ListCount,
		},
		{
			name:       "E2E - PutObject",
			method:     http.MethodPut,
			target:     "/demo-bucket/demo-key",
			headers:    map[string]string{"Content-Length": "1024"},
			wantStatus: http.StatusOK,
			callName:   "gantry resolve write",
			callCount:  (*testutil.GantryStub).ResolveWriteCount,
		},
	}

	listenAndServe = func(addr string, h http.Handler) error {
		gotAddr = addr

		for _, tt := range tests {
			req := httptest.NewRequest(tt.method, tt.target, nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("%s: status got %d, want %d", tt.name, rec.Code, tt.wantStatus)
			}
			if tt.callCount != nil {
				if got := tt.callCount(fg); got != 1 {
					t.Fatalf("%s: %s count got %d, want 1", tt.name, tt.callName, got)
				}
			}

			// reset for next iteration
			fg.CreateCalls = nil
			fg.ListCalls = 0
			fg.ResolveWriteCalls = nil
			fg.CreateFn = nil
			fg.ListFn = nil
		}
		served = true
		return nil
	}

	main()

	if gotAddr != ":8080" {
		t.Fatalf("addr: got %q, want %q", gotAddr, ":8080")
	}
	if !served {
		t.Fatalf("expected handler to serve a request")
	}
}
