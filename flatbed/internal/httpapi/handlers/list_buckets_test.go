package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ratdaddy/blockcloset/flatbed/internal/gantry"
	"github.com/ratdaddy/blockcloset/flatbed/internal/httpapi/handlers"
	"github.com/ratdaddy/blockcloset/flatbed/internal/testutil"
)

func TestHandlers_ListBuckets(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, time.January, 1, 12, 0, 0, 0, time.UTC)

	type tc struct {
		name             string
		listResp         []gantry.Bucket
		listErr          error
		wantStatus       int
		wantBodySubstr   string
		wantContentType  bool
		expectBucketData bool
	}

	cases := []tc{
		{
			name: "success returns JSON payload",
			listResp: []gantry.Bucket{
				{Name: "first-bucket", CreatedAt: now.Add(-2 * time.Hour)},
				{Name: "middle-bucket", CreatedAt: now},
				{Name: "third-bucket", CreatedAt: now.Add(5 * time.Hour)},
			},
			wantStatus:       http.StatusOK,
			wantContentType:  true,
			expectBucketData: true,
		},
		{
			name:           "gantry error propagates as internal error",
			listErr:        status.Error(codes.Internal, "gantry failure"),
			wantStatus:     http.StatusInternalServerError,
			wantBodySubstr: "InternalError",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			stub := testutil.NewGantryStub()
			stub.ListFn = func(ctx context.Context) ([]gantry.Bucket, error) {
				return c.listResp, c.listErr
			}

			h := &handlers.Handlers{Gantry: stub, Cradle: testutil.NewCradleStub()}

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()

			h.ListBuckets(rec, req)

			if stub.ListCalls != 1 {
				t.Fatalf("ListBuckets gantry calls: got %d want 1", stub.ListCalls)
			}

			if rec.Code != c.wantStatus {
				t.Fatalf("status code: got %d want %d", rec.Code, c.wantStatus)
			}

			if c.wantContentType {
				if got := rec.Header().Get("Content-Type"); got == "" || !strings.Contains(got, "application/json") {
					t.Fatalf("content type: got %q want application/json", got)
				}
			}

			if c.expectBucketData {
				type responseBucket struct {
					Name         string `json:"Name"`
					CreationDate string `json:"CreationDate"`
				}
				type responsePayload struct {
					Buckets []responseBucket `json:"Buckets"`
				}

				var payload responsePayload
				if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
					t.Fatalf("decode response: %v", err)
				}

				want := responsePayload{Buckets: make([]responseBucket, len(c.listResp))}
				for i, b := range c.listResp {
					want.Buckets[i] = responseBucket{
						Name:         b.Name,
						CreationDate: b.CreatedAt.UTC().Format(time.RFC3339),
					}
				}

				if diff := cmp.Diff(want, payload); diff != "" {
					t.Fatalf("Buckets mismatch (-want +got):\n%s", diff)
				}
			} else if c.wantBodySubstr != "" {
				body := rec.Body.String()
				if !strings.Contains(body, c.wantBodySubstr) {
					t.Fatalf("response body: want substring %q got %q", c.wantBodySubstr, body)
				}
			}
		})
	}
}
