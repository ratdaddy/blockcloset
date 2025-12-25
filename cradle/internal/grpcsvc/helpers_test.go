package grpcsvc

import (
	"io"
	"log/slog"
	"strings"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func newDiscardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
}

func assertNoError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("want nil error, got %v", err)
	}
}

func assertGRPCError(t *testing.T, err error, code codes.Code, substring string) {
	t.Helper()

	if err == nil {
		t.Fatalf("want error, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("want gRPC status error, got %v", err)
	}

	if st.Code() != code {
		t.Fatalf("status code: got %v, want %v", st.Code(), code)
	}

	if !strings.Contains(st.Message(), substring) {
		t.Fatalf("status message: got %q, want to contain %q", st.Message(), substring)
	}
}
