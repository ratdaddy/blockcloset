package grpcsvc

import (
	"context"
	"testing"

	servicev1 "github.com/ratdaddy/blockcloset/proto/gen/gantry/service/v1"
)

func TestService_ResolveWrite(t *testing.T) {
	t.Parallel()

	type tc struct {
		name              string
		bucket            string
		key               string
		size              int64
		wantErr           bool
		wantObjectID      bool
		wantCradleAddress bool
	}

	cases := []tc{
		{
			name:              "valid request returns object_id and cradle_address",
			bucket:            "my-bucket",
			key:               "my-key.txt",
			size:              1024,
			wantObjectID:      true,
			wantCradleAddress: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			logger := newDiscardLogger()
			svc := New(logger, nil)

			resp, err := svc.ResolveWrite(context.Background(), &servicev1.ResolveWriteRequest{
				Bucket: c.bucket,
				Key:    c.key,
				Size:   c.size,
			})

			if c.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			assertNoError(t, err)

			if c.wantObjectID && resp.GetObjectId() == "" {
				t.Fatal("expected non-empty object_id")
			}

			if c.wantCradleAddress && resp.GetCradleAddress() == "" {
				t.Fatal("expected non-empty cradle_address")
			}
		})
	}
}
