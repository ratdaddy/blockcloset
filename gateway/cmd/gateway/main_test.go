package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMain_WiresAddressAndHandler(t *testing.T) {
	t.Parallel()

	origBuild := buildHandler
	origListen := listenAndServe
	defer func() {
		buildHandler = origBuild
		listenAndServe = origListen
	}()

	buildHandler = origBuild

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
