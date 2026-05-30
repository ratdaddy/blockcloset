package grpcsvc

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc/codes"

	servicev1 "github.com/ratdaddy/blockcloset/proto/gen/cradle/service/v1"
)

func TestService_Heartbeat(t *testing.T) {
	t.Parallel()

	const oneGiB = 1073741824

	cases := []struct {
		name           string
		availBytes     uint64
		availErr       error
		wantAvailBytes int64
		wantErr        bool
		wantCode       codes.Code
		wantMessage    string
	}{
		{
			name:           "returns available bytes",
			availBytes:     oneGiB,
			wantAvailBytes: oneGiB,
		},
		{
			name:        "statfs error",
			availErr:    errors.New("no such file or directory"),
			wantErr:     true,
			wantCode:    codes.Internal,
			wantMessage: "no such file or directory",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			svc := New(newDiscardLogger())
			svc.availableBytes = func(_ string) (uint64, error) { return c.availBytes, c.availErr }
			resp, err := svc.Heartbeat(context.Background(), &servicev1.HeartbeatRequest{})

			if c.wantErr {
				assertGRPCError(t, err, c.wantCode, c.wantMessage)
				return
			}

			assertNoError(t, err)
			if resp.GetAvailableBytes() != c.wantAvailBytes {
				t.Fatalf("available_bytes: got %d, want %d", resp.GetAvailableBytes(), c.wantAvailBytes)
			}
		})
	}
}
