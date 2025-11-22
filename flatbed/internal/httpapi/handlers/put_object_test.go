package handlers_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ratdaddy/blockcloset/flatbed/internal/httpapi/handlers"
	"github.com/ratdaddy/blockcloset/flatbed/internal/testutil"
	"github.com/ratdaddy/blockcloset/pkg/validation"
)

func TestPutObject_ValidationGantryAndResponse(t *testing.T) {
	t.Parallel()

	type tc struct {
		name              string
		bucket            string
		key               string
		contentLength     string // empty string means omit header
		transferEncoding  string // if set, adds Transfer-Encoding header
		wantStatus        int
		wantResolves      int
		wantBucket        string
		wantKey           string
		wantSize          int64
		wantBodySubstr    string
	}

	cases := []tc{
		{
			name:          "valid bucket and key -> 200 and ResolveWrite call",
			bucket:        "my-bucket",
			key:           "my-key",
			contentLength: "1024",
			wantStatus:    http.StatusOK,
			wantResolves:  1,
			wantBucket:    "my-bucket",
			wantKey:       "my-key",
			wantSize:      1024,
		},
		{
			name:           "missing Content-Length -> 411",
			bucket:         "my-bucket",
			key:            "my-key",
			contentLength:  "", // omit header
			wantStatus:     http.StatusLengthRequired,
			wantResolves:   0,
			wantBodySubstr: "MissingContentLength",
		},
		{
			name:           "zero Content-Length -> 400",
			bucket:         "my-bucket",
			key:            "my-key",
			contentLength:  "0",
			wantStatus:     http.StatusBadRequest,
			wantResolves:   0,
			wantBodySubstr: "InvalidArgument",
		},
		{
			name:           "oversized Content-Length -> 400",
			bucket:         "my-bucket",
			key:            "my-key",
			contentLength:  "5368709121", // 5 GiB + 1 byte
			wantStatus:     http.StatusBadRequest,
			wantResolves:   0,
			wantBodySubstr: "EntityTooLarge",
		},
		{
			name:             "chunked transfer encoding -> 400",
			bucket:           "my-bucket",
			key:              "my-key",
			contentLength:    "1024",
			transferEncoding: "chunked",
			wantStatus:       http.StatusBadRequest,
			wantResolves:     0,
			wantBodySubstr:   "InvalidRequest",
		},
		{
			name:           "invalid bucket name -> 400",
			bucket:         "INVALID-BUCKET",
			key:            "my-key",
			contentLength:  "1024",
			wantStatus:     http.StatusBadRequest,
			wantResolves:   0,
			wantBodySubstr: "InvalidBucketName",
		},
		{
			name:           "invalid key (null byte) -> 400",
			bucket:         "my-bucket",
			key:            "file\x00name",
			contentLength:  "1024",
			wantStatus:     http.StatusBadRequest,
			wantResolves:   0,
			wantBodySubstr: "InvalidKeyName",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			stub := testutil.NewGantryStub()
			h := &handlers.Handlers{
				BucketValidator: validation.DefaultBucketNameValidator{},
				KeyValidator:    validation.DefaultKeyValidator{},
				Gantry:          stub,
			}

			req := httptest.NewRequest(http.MethodPut, "/", nil)
			req.SetPathValue("bucket", c.bucket)
			req.SetPathValue("key", c.key)
			if c.contentLength != "" {
				req.Header.Set("Content-Length", c.contentLength)
			}
			if c.transferEncoding != "" {
				// Set TransferEncoding field directly (httptest doesn't process headers like real server)
				req.TransferEncoding = []string{c.transferEncoding}
			}
			rec := httptest.NewRecorder()

			h.PutObject(rec, req)

			if rec.Code != c.wantStatus {
				t.Fatalf("status: got %d, want %d", rec.Code, c.wantStatus)
			}

			if got := stub.ResolveWriteCount(); got != c.wantResolves {
				t.Fatalf("ResolveWrite calls: got %d, want %d", got, c.wantResolves)
			}

			if c.wantResolves > 0 {
				if len(stub.ResolveWriteCalls) == 0 {
					t.Fatalf("expected ResolveWrite call, got none")
				}
				call := stub.ResolveWriteCalls[0]
				if call.Bucket != c.wantBucket {
					t.Fatalf("ResolveWrite bucket: got %q, want %q", call.Bucket, c.wantBucket)
				}
				if call.Key != c.wantKey {
					t.Fatalf("ResolveWrite key: got %q, want %q", call.Key, c.wantKey)
				}
				if call.Size != c.wantSize {
					t.Fatalf("ResolveWrite size: got %d, want %d", call.Size, c.wantSize)
				}
			}

			if c.wantBodySubstr != "" {
				body, _ := io.ReadAll(rec.Body)
				if !strings.Contains(string(body), c.wantBodySubstr) {
					t.Fatalf("body: expected substring %q, got %q", c.wantBodySubstr, string(body))
				}
			}
		})
	}
}
