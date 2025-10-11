package requestid

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"maps"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/go-chi/httplog/v3"

	"github.com/ratdaddy/blockcloset/gateway/internal/config"
)

type captureHandler struct {
	mu      sync.Mutex
	records []map[string]any
	level   slog.Leveler
}

func (h *captureHandler) Records() []map[string]any {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.records) == 0 {
		return nil
	}
	prefix := h.records[0]
	out := make([]map[string]any, 0, len(h.records))
	for i, rec := range h.records {
		if i == 0 && prefix != nil {
			continue
		}
		m := make(map[string]any, len(prefix)+len(rec))
		maps.Copy(m, prefix)
		maps.Copy(m, rec)
		out = append(out, m)
	}
	return out
}

func parseJSONStream(t *testing.T, r io.Reader) []map[string]any {
	t.Helper()
	dec := json.NewDecoder(r)

	var out []map[string]any
	for {
		var m map[string]any
		if err := dec.Decode(&m); err == io.EOF {
			break
		} else if err != nil {
			t.Fatalf("bad json log: %v", err)
		}
		out = append(out, m)
	}
	return out
}

func exercise(t *testing.T, req *http.Request) (*httptest.ResponseRecorder, string, []map[string]any) {
	t.Helper()
	config.LogVerbosity = config.LogVerbose

	var observedID string

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		observedID = RequestIDFromContext(r.Context())
		if observedID == "" {
			t.Fatalf("expected request id in context")
		}
		if got := r.Header.Get("X-Request-ID"); got != observedID {
			t.Fatalf("expected r.Header[X-Request-ID]=%q, got %q", observedID, got)
		}
		_, _ = io.WriteString(w, observedID)
	})

	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	chain := RequestID()(h)
	chain = httplog.RequestLogger(logger, nil)(chain)

	rr := httptest.NewRecorder()
	chain.ServeHTTP(rr, req)

	return rr, observedID, parseJSONStream(t, &buf)

}

func TestMiddleware_Defaults_Table(t *testing.T) {
	safeRe := regexp.MustCompile(`^[A-Za-z0-9._-]{1,128}$`)
	ulidRe := regexp.MustCompile(`^[0-9A-HJKMNP-TV-Z]{26}$`)

	tests := []struct {
		name           string
		inHeader       string
		wantUseInbound bool
		wantGenerated  bool
	}{
		{
			name:           "uses_well_formed_inbound_header",
			inHeader:       "req_abc-123.DEF",
			wantUseInbound: true,
		},
		{
			name:          "generates_when_missing",
			inHeader:      "",
			wantGenerated: true,
		},
		{
			name:          "rejects_unsafe_inbound_and_generates",
			inHeader:      "bad\nheader",
			wantGenerated: true,
		},
		{
			name:           "trims_and_uses_inbound_if_safe_after_trim",
			inHeader:       "   abc_DEF-123   ",
			wantUseInbound: true,
		},
		{
			name:          "rejects_too_long_inbound",
			inHeader:      strings.Repeat("a", 129),
			wantGenerated: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "http://example.test/ok", nil)
			if tt.inHeader != "" {
				req.Header.Set("X-Request-ID", tt.inHeader)
			}

			rr, observed, logs := exercise(t, req)

			gotHeader := rr.Header().Get("X-Request-ID")
			if gotHeader == "" {
				t.Fatalf("expected response X-Request-ID header to be set")
			}

			body := strings.TrimSpace(rr.Body.String())
			if body == "" {
				t.Fatalf("expected body to contain observed id")
			}

			if observed != gotHeader || observed != body {
				t.Fatalf("id mismatch: observed=%q respHeader=%q body=%q", observed, gotHeader, body)
			}

			if !safeRe.MatchString(observed) {
				t.Fatalf("resulting id %q does not match safe charset", observed)
			}

			switch {
			case tt.wantUseInbound:
				want := strings.TrimSpace(tt.inHeader)
				if observed != want {
					t.Fatalf("expected to use inbound id %q, got %q", want, observed)
				}
			case tt.wantGenerated:
				if !ulidRe.MatchString(observed) {
					t.Fatalf("expected generated ksuid id (26 chars, Crockford Base32), got %q", observed)
				}
				if tt.inHeader != "" && strings.TrimSpace(tt.inHeader) == observed {
					t.Fatalf("expected generated id to differ from unsafe inbound, got same %q", observed)
				}
			default:
				t.Fatalf("test case must set wantUseInbound or wantGenerated")
			}

			found := false
			for _, rec := range logs {
				if v, ok := rec[otelKey]; ok {
					if s, ok := v.(string); ok && s == observed {
						found = true
						break
					}
				}
			}
			if !found {
				t.Fatalf("expected at least one log record with %s=%q; got %+v", otelKey, observed, logs)
			}
		})
	}
}

func TestMiddleware_GenerationProducesDifferentIDs(t *testing.T) {
	req1 := httptest.NewRequest(http.MethodGet, "http://example.test/a", nil)
	req2 := httptest.NewRequest(http.MethodGet, "http://example.test/b", nil)

	rr1, id1, _ := exercise(t, req1)
	_, id2, _ := exercise(t, req2)

	if id1 == id2 {
		t.Fatalf("expected different generated ids; got same %q", id1)
	}
	if !(id1 < id2) {
		t.Fatalf("expected ULIDs to sort by time: %s !< %s", id1, id2)
	}
	if rr1.Header().Get("X-Request-ID") != id1 {
		t.Fatalf("response header should echo id1")
	}
}

func TestWithRequestIDAndRequestIDFromContext(t *testing.T) {
	t.Parallel()

	base := context.Background()

	if got := RequestIDFromContext(base); got != "" {
		t.Fatalf("RequestIDFromContext empty base = %q, want \"\"", got)
	}

	tests := []struct {
		name     string
		ctx      context.Context
		id       string
		want     string
		wantSame bool
	}{
		{
			name: "sets id on empty context",
			ctx:  base,
			id:   "req-123",
			want: "req-123",
		},
		{
			name: "overrides existing id",
			ctx:  WithRequestID(base, "old"),
			id:   "new",
			want: "new",
		},
		{
			name:     "empty id leaves context unchanged",
			ctx:      WithRequestID(base, "keep"),
			id:       "",
			want:     "keep",
			wantSame: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotCtx := WithRequestID(tc.ctx, tc.id)

			if tc.wantSame && gotCtx != tc.ctx {
				t.Fatalf("expected context to be unchanged")
			}

			if got := RequestIDFromContext(gotCtx); got != tc.want {
				t.Fatalf("RequestIDFromContext = %q, want %q", got, tc.want)
			}
		})
	}
}
