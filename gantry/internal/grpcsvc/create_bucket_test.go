package grpcsvc_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ratdaddy/blockcloset/gantry/internal/grpcsvc"
	"github.com/ratdaddy/blockcloset/pkg/storage/bucket"
	servicev1 "github.com/ratdaddy/blockcloset/proto/gen/gantry/service/v1"
)

func TestService_CreateBucket(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	svc := grpcsvc.New(logger)

	type tc struct {
		name    string
		bucket  string
		wantErr bool
		code    codes.Code
		message string
	}

	cases := []tc{
		{
			name:   "valid bucket",
			bucket: "my-bucket-123",
		},
		{
			name:    "invalid bucket",
			bucket:  "Bad!Name",
			wantErr: true,
			code:    codes.InvalidArgument,
			message: bucket.ErrInvalidBucketName.Error(),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			resp, err := svc.CreateBucket(context.Background(), &servicev1.CreateBucketRequest{Name: c.bucket})

			if c.wantErr {
				if err == nil {
					t.Fatalf("want error, got nil")
				}

				st, ok := status.FromError(err)
				if !ok {
					t.Fatalf("want gRPC status error, got %v", err)
				}

				if st.Code() != c.code {
					t.Fatalf("status code: got %v, want %v", st.Code(), c.code)
				}

				if st.Message() != c.message {
					t.Fatalf("status message: got %q, want %q", st.Message(), c.message)
				}

				return
			}

			if err != nil {
				t.Fatalf("want nil error, got %v", err)
			}

			if resp == nil || resp.GetBucket() == nil {
				t.Fatalf("response bucket missing: %#v", resp)
			}

			if resp.GetBucket().GetName() != c.bucket {
				t.Fatalf("bucket name: got %q, want %q", resp.GetBucket().GetName(), c.bucket)
			}
		})
	}
}
