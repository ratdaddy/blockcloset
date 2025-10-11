package store_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	store "github.com/ratdaddy/blockcloset/gantry/internal/store"
)

func TestBucketStore_Create(t *testing.T) {
	t.Parallel()

	type tc struct {
		name    string
		setup   func(context.Context, *testing.T, store.BucketStore, time.Time)
		id      string
		bucket  string
		wantErr error
	}

	cases := []tc{
		{
			name:   "creates bucket",
			id:     "bucket-id-123",
			bucket: "integration-bucket",
		},
		{
			name:   "duplicate bucket name returns error",
			id:     "bucket-id-456",
			bucket: "existing-bucket",
			setup: func(ctx context.Context, t *testing.T, s store.BucketStore, createdAt time.Time) {
				t.Helper()
				if _, err := s.Create(ctx, "seed-bucket-id", "existing-bucket", createdAt); err != nil {
					t.Fatalf("seed create: %v", err)
				}
			},
			wantErr: store.ErrBucketAlreadyExists,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			db := openIsolatedDB(t)
			s := store.NewBucketStore(db)
			createdAt := time.Now().UTC()

			if c.setup != nil {
				c.setup(ctx, t, s, createdAt)
			}

			rec, err := s.Create(ctx, c.id, c.bucket, createdAt)

			if c.wantErr != nil {
				if err == nil {
					t.Fatalf("Create: expected error %v", c.wantErr)
				}

				if !errors.Is(err, c.wantErr) {
					t.Fatalf("Create error: got %v want %v", err, c.wantErr)
				}

				return
			}

			if err != nil {
				t.Fatalf("Create: unexpected error: %v", err)
			}

			assertBucketRecord(t, ctx, db, rec, c.id, createdAt)
		})
	}
}

func assertBucketRecord(t *testing.T, ctx context.Context, db *sql.DB, rec store.BucketRecord, expectedID string, expectedStamp time.Time) {
	t.Helper()

	if rec.ID != expectedID {
		t.Fatalf("Create: id mismatch: got %q want %q", rec.ID, expectedID)
	}

	if rec.Name == "" {
		t.Fatal("Create: name not populated")
	}

	if rec.CreatedAt.IsZero() {
		t.Fatal("Create: created_at not populated")
	}

	if rec.UpdatedAt.IsZero() {
		t.Fatal("Create: updated_at not populated")
	}

	if rec.UpdatedAt.Before(rec.CreatedAt) {
		t.Fatalf("Create: updated_at before created_at (%s < %s)", rec.UpdatedAt, rec.CreatedAt)
	}

	var (
		storedID        string
		storedName      string
		storedCreatedAt string
		storedUpdatedAt string
	)

	query := `SELECT id, name, created_at, updated_at FROM buckets WHERE name = ?`
	if err := db.QueryRowContext(ctx, query, rec.Name).Scan(&storedID, &storedName, &storedCreatedAt, &storedUpdatedAt); err != nil {
		t.Fatalf("fetch stored bucket: %v", err)
	}

	if storedID != rec.ID {
		t.Fatalf("stored id mismatch: got %q want %q", storedID, rec.ID)
	}

	if storedName != rec.Name {
		t.Fatalf("stored name mismatch: got %q want %q", storedName, rec.Name)
	}

	createdAt, err := time.Parse(time.RFC3339Nano, storedCreatedAt)
	if err != nil {
		t.Fatalf("parse created_at: %v", err)
	}

	updatedAt, err := time.Parse(time.RFC3339Nano, storedUpdatedAt)
	if err != nil {
		t.Fatalf("parse updated_at: %v", err)
	}

	if !rec.CreatedAt.Equal(expectedStamp) {
		t.Fatalf("returned CreatedAt mismatch: got %s want %s", rec.CreatedAt, createdAt)
	}

	if !rec.UpdatedAt.Equal(expectedStamp) {
		t.Fatalf("returned UpdatedAt mismatch: got %s want %s", rec.UpdatedAt, updatedAt)
	}

	wantTimestamp := expectedStamp.Truncate(time.Microsecond)
	if !createdAt.Equal(wantTimestamp) || !updatedAt.Equal(wantTimestamp) {
		t.Fatalf("stored timestamps mismatch: got created=%s updated=%s want %s", createdAt, updatedAt, wantTimestamp)
	}
}
