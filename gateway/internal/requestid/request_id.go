package requestid

import (
	"context"
	"crypto/rand"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/httplog/v3"
	"github.com/oklog/ulid/v2"

	"github.com/ratdaddy/blockcloset/gateway/internal/config"
)

type ctxKey int

const (
	ctxKeyID        ctxKey = iota
	headerRequestID        = "X-Request-ID"
	otelKey                = "http.request.header.x-request-id"
)

var (
	ulidMu      sync.Mutex
	ulidEntropy = ulid.Monotonic(rand.Reader, 0)
)

func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(ctxKeyID).(string)
	return id
}

func WithRequestID(ctx context.Context, id string) context.Context {
	if id == "" {
		return ctx
	}
	return context.WithValue(ctx, ctxKeyID, id)
}

func defaultID() string {
	ulidMu.Lock()
	id := ulid.MustNew(ulid.Timestamp(time.Now()), ulidEntropy)
	ulidMu.Unlock()
	return id.String()
}

func RequestID() func(http.Handler) http.Handler {
	safe := regexp.MustCompile(`^[A-Za-z0-9._-]{1,128}$`)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var id string
			if v := strings.TrimSpace(r.Header.Get(headerRequestID)); safe.MatchString(v) {
				id = v
			}
			if id == "" {
				id = defaultID()
			}

			ctx := WithRequestID(r.Context(), id)
			r.Header.Set(headerRequestID, id)
			w.Header().Set(headerRequestID, id)

			if config.LogVerbosity == config.LogVerbose {
				httplog.SetAttrs(ctx, slog.String(otelKey, id))
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
