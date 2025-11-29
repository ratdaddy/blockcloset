package store_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	store "github.com/ratdaddy/blockcloset/gantry/internal/store"
)

func TestObjectStore_Create(t *testing.T) {
	t.Parallel()

	type tc struct {
		name             string
		id               string
		bucketID         string
		key              string
		sizeExpected     int64
		cradleServerID   string
		skipBucket       bool
		skipCradleServer bool
		setup            func(context.Context, *testing.T, *store.ObjectStore, time.Time, *sql.DB, string, string)
		wantErr          bool
	}

	base := time.Date(2025, time.January, 1, 12, 0, 0, 0, time.UTC)

	cases := []tc{
		{
			name:           "creates object with PENDING state",
			id:             "object-id-1",
			bucketID:       "bucket-id-1",
			key:            "test-key.txt",
			sizeExpected:   1024,
			cradleServerID: "cradle-id-1",
		},
		{
			name:             "invalid foreign keys return error",
			id:               "object-id-2",
			bucketID:         "nonexistent-bucket-id",
			key:              "test-key.txt",
			sizeExpected:     1024,
			cradleServerID:   "nonexistent-cradle-id",
			skipBucket:       true,
			skipCradleServer: true,
			wantErr:          true,
		},
		{
			name:           "duplicate bucket+key allowed (multiple PENDING records)",
			id:             "object-id-3",
			bucketID:       "bucket-id-1",
			key:            "duplicate-key.txt",
			sizeExpected:   2048,
			cradleServerID: "cradle-id-1",
			setup: func(ctx context.Context, t *testing.T, s *store.ObjectStore, createdAt time.Time, db *sql.DB, bucketID, cradleServerID string) {
				// Create first object with same bucket+key
				_, err := (*s).Create(ctx, "object-id-first", bucketID, "duplicate-key.txt", 1024, cradleServerID, createdAt)
				if err != nil {
					t.Fatalf("setup: create first object: %v", err)
				}
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			db := openIsolatedDB(t)
			s := store.NewObjectStore(db)

			createdAt := base

			// Setup prerequisites (bucket and cradle server)
			setupPrerequisites(ctx, t, db, c.bucketID, c.cradleServerID, createdAt, c.skipBucket, c.skipCradleServer)

			// Additional test-specific setup
			if c.setup != nil {
				c.setup(ctx, t, &s, createdAt, db, c.bucketID, c.cradleServerID)
			}

			rec, err := s.Create(ctx, c.id, c.bucketID, c.key, c.sizeExpected, c.cradleServerID, createdAt)

			if c.wantErr {
				if err == nil {
					t.Fatal("Create: expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Create: unexpected error: %v", err)
			}

			assertObjectRecord(t, ctx, db, rec, c.id, c.bucketID, c.key, c.sizeExpected, c.cradleServerID, createdAt)
		})
	}
}

func setupPrerequisites(ctx context.Context, t *testing.T, db *sql.DB, bucketID, cradleServerID string, createdAt time.Time, skipBucket, skipCradleServer bool) {
	t.Helper()

	if !skipBucket {
		buckets := store.NewBucketStore(db)
		_, err := buckets.Create(ctx, bucketID, "test-bucket", createdAt)
		if err != nil {
			t.Fatalf("setup: create bucket: %v", err)
		}
	}

	if !skipCradleServer {
		cradles := store.NewCradleServerStore(db)
		_, err := cradles.Upsert(ctx, cradleServerID, "127.0.0.1:9444", createdAt)
		if err != nil {
			t.Fatalf("setup: upsert cradle server: %v", err)
		}
	}
}

func assertObjectRecord(t *testing.T, ctx context.Context, db *sql.DB, rec store.ObjectRecord, wantID, wantBucketID, wantKey string, wantSize int64, wantCradleServerID string, expectedStamp time.Time) {
	t.Helper()

	wantStamp := expectedStamp.UTC().Truncate(time.Microsecond)

	// Verify returned record fields
	if rec.ID != wantID {
		t.Fatalf("returned ID: got %q, want %q", rec.ID, wantID)
	}
	if rec.BucketID != wantBucketID {
		t.Fatalf("returned BucketID: got %q, want %q", rec.BucketID, wantBucketID)
	}
	if rec.Key != wantKey {
		t.Fatalf("returned Key: got %q, want %q", rec.Key, wantKey)
	}
	if rec.State != "PENDING" {
		t.Fatalf("returned State: got %q, want %q", rec.State, "PENDING")
	}
	if rec.SizeExpected != wantSize {
		t.Fatalf("returned SizeExpected: got %d, want %d", rec.SizeExpected, wantSize)
	}
	if rec.CradleServerID != wantCradleServerID {
		t.Fatalf("returned CradleServerID: got %q, want %q", rec.CradleServerID, wantCradleServerID)
	}
	if !rec.CreatedAt.Equal(wantStamp) {
		t.Fatalf("returned CreatedAt: got %s, want %s", rec.CreatedAt, wantStamp)
	}
	if !rec.UpdatedAt.Equal(wantStamp) {
		t.Fatalf("returned UpdatedAt: got %s, want %s", rec.UpdatedAt, wantStamp)
	}

	// Query database to verify record was actually written
	var (
		storedID             string
		storedBucketID       string
		storedKey            string
		storedState          string
		storedSizeExpected   int64
		storedCradleServerID string
		storedCreatedAt      int64
		storedUpdatedAt      int64
	)

	query := `SELECT object_id, bucket_id, key, state, size_expected, cradle_server_id, created_at, updated_at FROM objects WHERE object_id = ?`
	if err := db.QueryRowContext(ctx, query, rec.ID).Scan(&storedID, &storedBucketID, &storedKey, &storedState, &storedSizeExpected, &storedCradleServerID, &storedCreatedAt, &storedUpdatedAt); err != nil {
		t.Fatalf("fetch stored object: %v", err)
	}

	if storedID != rec.ID {
		t.Fatalf("stored object_id: got %q, want %q", storedID, rec.ID)
	}
	if storedBucketID != rec.BucketID {
		t.Fatalf("stored bucket_id: got %q, want %q", storedBucketID, rec.BucketID)
	}
	if storedKey != rec.Key {
		t.Fatalf("stored key: got %q, want %q", storedKey, rec.Key)
	}
	if storedState != "PENDING" {
		t.Fatalf("stored state: got %q, want %q", storedState, "PENDING")
	}
	if storedSizeExpected != rec.SizeExpected {
		t.Fatalf("stored size_expected: got %d, want %d", storedSizeExpected, rec.SizeExpected)
	}
	if storedCradleServerID != rec.CradleServerID {
		t.Fatalf("stored cradle_server_id: got %q, want %q", storedCradleServerID, rec.CradleServerID)
	}

	createdAt := time.UnixMicro(storedCreatedAt).UTC()
	updatedAt := time.UnixMicro(storedUpdatedAt).UTC()

	if !createdAt.Equal(wantStamp) {
		t.Fatalf("stored created_at: got %s, want %s", createdAt, wantStamp)
	}
	if !updatedAt.Equal(wantStamp) {
		t.Fatalf("stored updated_at: got %s, want %s", updatedAt, wantStamp)
	}
}
