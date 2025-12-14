package handlers_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ratdaddy/blockcloset/flatbed/internal/httpapi/handlers"
	"github.com/ratdaddy/blockcloset/flatbed/internal/testutil"
	"github.com/ratdaddy/blockcloset/pkg/validation"
	servicev1 "github.com/ratdaddy/blockcloset/proto/gen/gantry/service/v1"
	writeplanv1 "github.com/ratdaddy/blockcloset/proto/gen/gantry/write_plan/v1"
)

func planWriteErr(code codes.Code, message string, reason servicev1.PlanWriteError_Reason, bucket string) error {
	st := status.New(code, message)
	detail := &servicev1.PlanWriteError{
		Reason: reason,
		Bucket: bucket,
	}
	st, err := st.WithDetails(detail)
	if err != nil {
		panic(err)
	}
	return st.Err()
}

func TestPutObject_ValidationGantryAndResponse(t *testing.T) {
	t.Parallel()

	type tc struct {
		name             string
		bucket           string
		key              string
		contentLength    string // empty string means omit header
		transferEncoding string // if set, adds Transfer-Encoding header
		gantryErr        error
		wantStatus       int
		wantResolves     int
		wantBucket       string
		wantKey          string
		wantSize         int64
		wantBodySubstr   string
	}

	cases := []tc{
		{
			name:          "valid bucket and key -> 200 and PlanWrite call",
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
		{
			name:          "gantry bucket not found -> 404",
			bucket:        "nonexistent-bucket",
			key:           "my-key",
			contentLength: "1024",
			gantryErr: planWriteErr(
				codes.NotFound,
				"bucket not found",
				servicev1.PlanWriteError_REASON_BUCKET_NOT_FOUND,
				"nonexistent-bucket",
			),
			wantStatus:     http.StatusNotFound,
			wantResolves:   1,
			wantBucket:     "nonexistent-bucket",
			wantKey:        "my-key",
			wantSize:       1024,
			wantBodySubstr: "NoSuchBucket",
		},
		{
			name:          "gantry bucket access denied -> 403",
			bucket:        "forbidden-bucket",
			key:           "my-key",
			contentLength: "1024",
			gantryErr: planWriteErr(
				codes.PermissionDenied,
				"access denied",
				servicev1.PlanWriteError_REASON_BUCKET_ACCESS_DENIED,
				"forbidden-bucket",
			),
			wantStatus:     http.StatusForbidden,
			wantResolves:   1,
			wantBucket:     "forbidden-bucket",
			wantKey:        "my-key",
			wantSize:       1024,
			wantBodySubstr: "AccessDenied",
		},
		{
			name:          "gantry no cradle servers -> 503",
			bucket:        "my-bucket",
			key:           "my-key",
			contentLength: "1024",
			gantryErr: planWriteErr(
				codes.FailedPrecondition,
				"no cradle servers available",
				servicev1.PlanWriteError_REASON_NO_CRADLE_SERVERS,
				"my-bucket",
			),
			wantStatus:     http.StatusServiceUnavailable,
			wantResolves:   1,
			wantBucket:     "my-bucket",
			wantKey:        "my-key",
			wantSize:       1024,
			wantBodySubstr: "ServiceUnavailable",
		},
		{
			name:          "gantry unexpected error -> 500",
			bucket:        "my-bucket",
			key:           "my-key",
			contentLength: "1024",
			gantryErr:     status.Error(codes.Internal, "unexpected database error"),
			wantStatus:    http.StatusInternalServerError,
			wantResolves:  1,
			wantBucket:    "my-bucket",
			wantKey:       "my-key",
			wantSize:      1024,
			wantBodySubstr: "InternalError",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			stub := testutil.NewGantryStub()
			if c.gantryErr != nil {
				stub.PlanWriteFn = func(context.Context, string, string, int64) (*writeplanv1.WritePlan, error) {
					return nil, c.gantryErr
				}
			}
			h := &handlers.Handlers{
				BucketValidator: validation.DefaultBucketNameValidator{},
				KeyValidator:    validation.DefaultKeyValidator{},
				Gantry:          stub,
				Cradle:          testutil.NewCradleStub(),
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

			if got := stub.PlanWriteCount(); got != c.wantResolves {
				t.Fatalf("PlanWrite calls: got %d, want %d", got, c.wantResolves)
			}

			if c.wantResolves > 0 {
				if len(stub.PlanWriteCalls) == 0 {
					t.Fatalf("expected PlanWrite call, got none")
				}
				call := stub.PlanWriteCalls[0]
				if call.Bucket != c.wantBucket {
					t.Fatalf("PlanWrite bucket: got %q, want %q", call.Bucket, c.wantBucket)
				}
				if call.Key != c.wantKey {
					t.Fatalf("PlanWrite key: got %q, want %q", call.Key, c.wantKey)
				}
				if call.Size != c.wantSize {
					t.Fatalf("PlanWrite size: got %d, want %d", call.Size, c.wantSize)
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

func TestPutObject_CradleIntegration(t *testing.T) {
	t.Parallel()

	type tc struct {
		name               string
		bucket             string
		key                string
		contentLength      string
		body               string
		planWriteResp      *writeplanv1.WritePlan
		cradleErr          error
		cradleBytesWritten int64 // if non-zero, stub returns this value
		wantStatus         int
		wantCradleCalls    int
		wantCradleAddress  string
		wantCradleObjID    string
		wantCradleBucket   string
		wantCradleSize     int64
		wantCradleBody     string
		wantBodySubstr     string
	}

	cases := []tc{
		{
			name:          "successful write streams to cradle",
			bucket:        "photos",
			key:           "vacation.jpg",
			contentLength: "17",
			body:          "test file content",
			planWriteResp: &writeplanv1.WritePlan{
				ObjectId:      "01ARZ3NDEKTSV4RRFFQ69G5FAV",
				CradleAddress: "localhost:9444",
			},
			wantStatus:        http.StatusOK,
			wantCradleCalls:   1,
			wantCradleAddress: "localhost:9444",
			wantCradleObjID:   "01ARZ3NDEKTSV4RRFFQ69G5FAV",
			wantCradleBucket:  "photos",
			wantCradleSize:    17,
			wantCradleBody:    "test file content",
		},
		{
			name:          "cradle write failure returns 500",
			bucket:        "photos",
			key:           "vacation.jpg",
			contentLength: "17",
			body:          "test file content",
			planWriteResp: &writeplanv1.WritePlan{
				ObjectId:      "01ARZ3NDEKTSV4RRFFQ69G5FAV",
				CradleAddress: "localhost:9444",
			},
			cradleErr:         errors.New("disk full"),
			wantStatus:        http.StatusInternalServerError,
			wantCradleCalls:   1,
			wantCradleAddress: "localhost:9444",
			wantCradleObjID:   "01ARZ3NDEKTSV4RRFFQ69G5FAV",
			wantCradleBucket:  "photos",
			wantCradleSize:    17,
			wantCradleBody:    "test file content",
			wantBodySubstr:    "InternalError",
		},
		{
			name:               "size mismatch returns 500",
			bucket:             "photos",
			key:                "vacation.jpg",
			contentLength:      "17",
			body:               "test file content",
			planWriteResp:      &writeplanv1.WritePlan{
				ObjectId:      "01ARZ3NDEKTSV4RRFFQ69G5FAV",
				CradleAddress: "localhost:9444",
			},
			cradleBytesWritten: 10, // mismatch: expected 17, got 10
			wantStatus:         http.StatusInternalServerError,
			wantCradleCalls:    1,
			wantCradleAddress:  "localhost:9444",
			wantCradleObjID:    "01ARZ3NDEKTSV4RRFFQ69G5FAV",
			wantCradleBucket:   "photos",
			wantCradleSize:     17,
			wantCradleBody:     "test file content",
			wantBodySubstr:     "InternalError",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			gantryStub := testutil.NewGantryStub()
			gantryStub.PlanWriteFn = func(ctx context.Context, bucket, key string, size int64) (*writeplanv1.WritePlan, error) {
				return c.planWriteResp, nil
			}

			cradleStub := testutil.NewCradleStub()
			if c.cradleErr != nil {
				cradleStub.WriteObjectFn = func(ctx context.Context, address, objectID, bucket string, size int64, body io.Reader) (int64, int64, error) {
					return 0, 0, c.cradleErr
				}
			}
			if c.cradleBytesWritten != 0 {
				cradleStub.WriteObjectFn = func(ctx context.Context, address, objectID, bucket string, size int64, body io.Reader) (int64, int64, error) {
					return c.cradleBytesWritten, 0, nil
				}
			}

			h := &handlers.Handlers{
				BucketValidator: validation.DefaultBucketNameValidator{},
				KeyValidator:    validation.DefaultKeyValidator{},
				Gantry:          gantryStub,
				Cradle:          cradleStub,
			}

			req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(c.body))
			req.SetPathValue("bucket", c.bucket)
			req.SetPathValue("key", c.key)
			req.Header.Set("Content-Length", c.contentLength)

			rec := httptest.NewRecorder()

			h.PutObject(rec, req)

			if rec.Code != c.wantStatus {
				t.Fatalf("status: got %d, want %d", rec.Code, c.wantStatus)
			}

			if cradleStub.WriteObjectCount() != c.wantCradleCalls {
				t.Fatalf("WriteObject call count: got %d, want %d", cradleStub.WriteObjectCount(), c.wantCradleCalls)
			}

			if c.wantCradleCalls > 0 {
				call := cradleStub.WriteObjectCalls[0]
				if call.Address != c.wantCradleAddress {
					t.Fatalf("WriteObject address: got %q, want %q", call.Address, c.wantCradleAddress)
				}
				if call.ObjectID != c.wantCradleObjID {
					t.Fatalf("WriteObject objectID: got %q, want %q", call.ObjectID, c.wantCradleObjID)
				}
				if call.Bucket != c.wantCradleBucket {
					t.Fatalf("WriteObject bucket: got %q, want %q", call.Bucket, c.wantCradleBucket)
				}
				if call.Size != c.wantCradleSize {
					t.Fatalf("WriteObject size: got %d, want %d", call.Size, c.wantCradleSize)
				}
				if string(call.BodyBytes) != c.wantCradleBody {
					t.Fatalf("WriteObject body: got %q, want %q", string(call.BodyBytes), c.wantCradleBody)
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
