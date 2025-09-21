package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ratdaddy/blockcloset/gateway/internal/httpapi"
)

type fakeGantry struct{ calls []string }

func (f *fakeGantry) CreateBucket(ctx context.Context, name string) (string, error) {
	f.calls = append(f.calls, name)
	return "", nil
}

func TestMain_WiresAddressAndHandler(t *testing.T) {
	t.Parallel()

	origBuild, origListen, origGantry := buildHandler, listenAndServe, gantryClient
	defer func() { buildHandler, listenAndServe, gantryClient = origBuild, origListen, origGantry }()

	fg := &fakeGantry{}
	gantryClient = func(addr string) (httpapi.GantryClient, error) {
		return fg, nil
	}

	var gotAddr string
	var served bool

	listenAndServe = func(addr string, h http.Handler) error {
		gotAddr = addr

		req := httptest.NewRequest(http.MethodPut, "/smoke-bucket", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("smoke: status got %d, want %d", rec.Code, http.StatusCreated)
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
