package grpcsvc

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc/codes"

	"github.com/ratdaddy/blockcloset/gantry/internal/testutil"
	servicev1 "github.com/ratdaddy/blockcloset/proto/gen/gantry/service/v1"
)

func TestService_CommitObject(t *testing.T) {
	t.Parallel()

	type tc struct {
		name             string
		objectID         string
		bucket           string
		key              string
		size             int64
		lastModifiedMs   int64
		commitErr        error
		wantErr          bool
		wantCode         codes.Code
		wantMessage      string
		expectCommitCall bool
	}

	cases := []tc{
		{
			name:             "valid request commits object and returns empty response",
			objectID:         "01JEBF2KR8JXZB3Q4V5TW6Y7Z8",
			bucket:           "my-bucket",
			key:              "photos/sunset.jpg",
			size:             4096,
			lastModifiedMs:   1735689600000,
			expectCommitCall: true,
		},
		{
			name:           "empty object_id returns InvalidArgument",
			objectID:       "",
			bucket:         "my-bucket",
			key:            "photos/sunset.jpg",
			size:           4096,
			lastModifiedMs: 1735689600000,
			wantErr:        true,
			wantCode:       codes.InvalidArgument,
			wantMessage:    "InvalidObjectID",
		},
		{
			name:           "zero size returns InvalidArgument",
			objectID:       "01JEBF2KR8JXZB3Q4V5TW6Y7Z8",
			bucket:         "my-bucket",
			key:            "photos/sunset.jpg",
			size:           0,
			lastModifiedMs: 1735689600000,
			wantErr:        true,
			wantCode:       codes.InvalidArgument,
			wantMessage:    "InvalidSize",
		},
		{
			name:           "zero last_modified_ms returns InvalidArgument",
			objectID:       "01JEBF2KR8JXZB3Q4V5TW6Y7Z8",
			bucket:         "my-bucket",
			key:            "photos/sunset.jpg",
			size:           4096,
			lastModifiedMs: 0,
			wantErr:        true,
			wantCode:       codes.InvalidArgument,
			wantMessage:    "InvalidLastModifiedMs",
		},
		{
			name:           "commit store error returns Internal",
			objectID:       "01JEBF2KR8JXZB3Q4V5TW6Y7Z8",
			bucket:         "my-bucket",
			key:            "photos/sunset.jpg",
			size:           4096,
			lastModifiedMs: 1735689600000,
			commitErr:      errors.New("commit error"),
			wantErr:        true,
			wantCode:       codes.Internal,
			wantMessage:    "commit error",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			logger := newDiscardLogger()
			svc := New(logger, nil)

			objects := testutil.NewFakeObjectStore()
			if c.commitErr != nil {
				objects.SetCommitError(c.commitErr)
			}

			svc.store = testutil.NewFakeStore(
				testutil.WithObjects(objects),
			)

			resp, err := svc.CommitObject(context.Background(), &servicev1.CommitObjectRequest{
				ObjectId:       c.objectID,
				Size:           c.size,
				LastModifiedMs: c.lastModifiedMs,
			})

			if c.wantErr {
				assertGRPCError(t, err, c.wantCode, c.wantMessage)
				return
			}

			assertNoError(t, err)

			if resp == nil {
				t.Fatal("response is nil")
			}

			if c.expectCommitCall {
				calls := objects.CommitCalls()
				if len(calls) != 1 {
					t.Fatalf("CommitWithReplace calls: got %d, want 1", len(calls))
				}
				call := calls[0]
				if call.ObjectID != c.objectID {
					t.Fatalf("CommitWithReplace object_id: got %q, want %q", call.ObjectID, c.objectID)
				}
				if call.SizeActual != c.size {
					t.Fatalf("CommitWithReplace size_actual: got %d, want %d", call.SizeActual, c.size)
				}
				if call.LastModifiedMs != c.lastModifiedMs {
					t.Fatalf("CommitWithReplace last_modified_ms: got %d, want %d", call.LastModifiedMs, c.lastModifiedMs)
				}
				if call.UpdatedAt.IsZero() {
					t.Fatal("CommitWithReplace updatedAt is zero")
				}
			}
		})
	}
}
