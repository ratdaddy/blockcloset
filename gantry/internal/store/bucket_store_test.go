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

func TestBucketStore_List(t *testing.T) {
	t.Parallel()

	type seed struct {
		id   string
		name string
		at   time.Time
	}

	base := time.Date(2025, time.January, 1, 12, 0, 0, 0, time.UTC)

	type tc struct {
		name        string
		seeds       []seed
		setupDB     func(*testing.T, *sql.DB)
		wantOrder   []seed
		expectError bool
	}

	cases := []tc{
		{
			name: "returns buckets oldest first",
			seeds: []seed{
				{id: "bucket-oldest", name: "first-bucket", at: base.Add(-2 * time.Hour)},
				{id: "bucket-newest", name: "second-bucket", at: base.Add(3 * time.Hour)},
				{id: "bucket-middle", name: "middle-bucket", at: base},
			},
			wantOrder: []seed{
				{id: "bucket-oldest", name: "first-bucket", at: base.Add(-2 * time.Hour)},
				{id: "bucket-middle", name: "middle-bucket", at: base},
				{id: "bucket-newest", name: "second-bucket", at: base.Add(3 * time.Hour)},
			},
		},
		{
			name: "propagates query errors",
			setupDB: func(t *testing.T, db *sql.DB) {
				t.Helper()
				if err := db.Close(); err != nil {
					t.Fatalf("close db: %v", err)
				}
			},
			expectError: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			db := openIsolatedDB(t)
			s := store.NewBucketStore(db)

			if c.setupDB != nil {
				c.setupDB(t, db)
			} else {
				for _, seed := range c.seeds {
					ts := seed.at.UTC().Truncate(time.Microsecond)
					if _, err := s.Create(ctx, seed.id, seed.name, ts); err != nil {
						t.Fatalf("seed create %q: %v", seed.name, err)
					}
				}
			}

			records, err := s.List(ctx)

			if c.expectError {
				if err == nil {
					t.Fatal("List: expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("List: unexpected error: %v", err)
			}

			if len(records) != len(c.wantOrder) {
				t.Fatalf("List: got %d records, want %d", len(records), len(c.wantOrder))
			}

			for i, want := range c.wantOrder {
				wantTime := want.at.UTC().Truncate(time.Microsecond)
				got := records[i]

				if got.ID != want.id {
					t.Fatalf("List[%d]: id mismatch: got %q want %q", i, got.ID, want.id)
				}
				if got.Name != want.name {
					t.Fatalf("List[%d]: name mismatch: got %q want %q", i, got.Name, want.name)
				}
				if !got.CreatedAt.Equal(wantTime) {
					t.Fatalf("List[%d]: created_at mismatch: got %s want %s", i, got.CreatedAt, wantTime)
				}
				if !got.UpdatedAt.Equal(wantTime) {
					t.Fatalf("List[%d]: updated_at mismatch: got %s want %s", i, got.UpdatedAt, wantTime)
				}
			}
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
