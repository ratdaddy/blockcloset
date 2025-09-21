package httpapi_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/ratdaddy/blockcloset/gateway/internal/httpapi"
	_ "github.com/ratdaddy/blockcloset/gateway/internal/testutil"
)

func reqWithBucket(t *testing.T, method, name string) *http.Request {
	t.Helper()
	r := httptest.NewRequest(method, "/", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("bucket", name)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

type stubValidator struct {
	err   error
	calls []string
}

func (s *stubValidator) ValidateBucketName(name string) error {
	s.calls = append(s.calls, name)
	return s.err
}

type stubGantryClient struct {
	calls []string
}

func (s *stubGantryClient) CreateBucket(ctx context.Context, name string) (string, error) {
	s.calls = append(s.calls, name)
	return "", nil
}

func TestCreateBucket_ValidationGantryAndResponse(t *testing.T) {
	t.Parallel()

	type tc struct {
		name         string
		bucket       string
		validatorErr error
		wantStatus   int
		wantLoc      string
		wantBodySub  string
	}

	cases := []tc{
		{
			name:         "valid bucket -> 201 and Location",
			bucket:       "my-bucket-123",
			validatorErr: nil,
			wantStatus:   http.StatusCreated,
			wantLoc:      "/my-bucket-123",
			wantBodySub:  "",
		},
		{
			name:         "invalid bucket -> 400",
			bucket:       "Bad!Name",
			validatorErr: httpapi.ErrInvalidBucketName,
			wantStatus:   http.StatusBadRequest,
			wantLoc:      "",
			wantBodySub:  httpapi.ErrInvalidBucketName.Error(),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			v := &stubValidator{err: c.validatorErr}
			g := &stubGantryClient{}
			h := &httpapi.Handlers{Validator: v, Gantry: g}

			req := reqWithBucket(t, http.MethodPut, c.bucket)
			rec := httptest.NewRecorder()

			h.CreateBucket(rec, req)

			if len(v.calls) != 1 || v.calls[0] != c.bucket {
				t.Fatalf("validator calls = %#v; want exactly [%q]", v.calls, c.bucket)
			}

			if c.validatorErr == nil {
				if len(g.calls) != 1 || g.calls[0] != c.bucket {
					t.Fatalf("gantry create_bucket calls = %#v; want exactly [%q]", g.calls, c.bucket)
				}
			} else {
				if len(g.calls) != 0 {
					t.Fatalf("expected no gantry calls, got %#v", g.calls)
				}
			}

			if rec.Code != c.wantStatus {
				t.Fatalf("status: got %d, want %d", rec.Code, c.wantStatus)
			}

			if c.wantLoc != "" {
				if got := rec.Header().Get("Location"); got != c.wantLoc {
					t.Fatalf("Location: got %q, want %q", got, c.wantLoc)
				}
			} else if got := rec.Header().Get("Location"); got != "" {
				t.Fatalf("unexpected Location header on error: %q", got)
			}

			if c.wantBodySub != "" {
				body, _ := io.ReadAll(rec.Body)
				if !strings.Contains(string(body), c.wantBodySub) {
					t.Fatalf("body: expected substring %q, got %q", c.wantBodySub, string(body))
				}
			}
		})
	}
}
